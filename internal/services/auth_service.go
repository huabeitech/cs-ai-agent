package services

import (
	"crypto/rand"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/config"
	"cs-agent/internal/pkg/constants"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/errorsx"
	"encoding/hex"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/sqls"
	"golang.org/x/crypto/bcrypt"
)

const (
	authPrincipalContextKey = "authPrincipal"
)

var AuthService = newAuthService()

func newAuthService() *authService {
	return &authService{}
}

type authService struct {
}

func (s *authService) GetAuthPrincipal(ctx iris.Context) *dto.AuthPrincipal {
	if ctx == nil {
		return nil
	}
	v := ctx.Values().Get(authPrincipalContextKey)
	if principal, ok := v.(*dto.AuthPrincipal); ok {
		return principal
	}
	return nil
}

func (s *authService) GetImPrincipal(ctx iris.Context) (*dto.AuthPrincipal, error) {
	if principal := s.GetAuthPrincipal(ctx); principal != nil {
		return principal, nil
	}
	if token := s.extractBearerToken(ctx.GetHeader("Authorization")); token != "" {
		return s.Authenticate(ctx)
	}
	visitorID := strings.TrimSpace(ctx.GetHeader("X-Visitor-Id"))
	if visitorID == "" {
		visitorID = strings.TrimSpace(ctx.URLParam("visitorId"))
	}
	if visitorID == "" {
		return nil, errorsx.Unauthorized("访客标识不能为空")
	}
	principal := &dto.AuthPrincipal{
		Username:  "访客",
		Nickname:  "访客",
		Status:    enums.StatusOk,
		IsVisitor: true,
		VisitorID: visitorID,
	}
	ctx.Values().Set(authPrincipalContextKey, principal)
	return principal, nil
}

func (s *authService) setAuthPrincipal(ctx iris.Context, user *models.User, roles, permissions []string) *dto.AuthPrincipal {
	principal := &dto.AuthPrincipal{
		UserID:      user.ID,
		Username:    user.Username,
		Nickname:    user.Nickname,
		Avatar:      user.Avatar,
		Status:      user.Status,
		Roles:       roles,
		Permissions: permissions,
	}
	ctx.Values().Set(authPrincipalContextKey, principal)
	return principal
}

func (s *authService) RequirePermission(ctx iris.Context, permission constants.Permission) error {
	if s.GetAuthPrincipal(ctx) == nil {
		if _, err := s.Authenticate(ctx); err != nil {
			return err
		}
	}
	if !s.HasPermission(ctx, permission.Code) {
		return errorsx.Forbidden("无权限执行该操作")
	}
	return nil
}

func (s *authService) Login(req request.LoginRequest, authCfg config.AuthConfig, clientIP, userAgent string) (*response.LoginResponse, error) {
	username := strings.TrimSpace(req.Username)
	password := req.Password
	if username == "" || strings.TrimSpace(password) == "" {
		return nil, errorsx.InvalidParam("用户名和密码不能为空")
	}

	user := UserService.Take("username = ? AND status = ?", username, enums.StatusOk)
	if user == nil {
		_ = s.createLoginCredentialLog(username, 0, false, clientIP, userAgent, "user not found")
		return nil, errorsx.InvalidAccount("用户名或密码错误")
	}
	if user.Password == "" || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		_ = s.createLoginCredentialLog(username, user.ID, false, clientIP, userAgent, "password mismatch")
		return nil, errorsx.InvalidAccount("用户名或密码错误")
	}

	ret, err := s.issueTokens(user, clientIP, userAgent, authCfg)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	if err = UserService.Updates(user.ID, map[string]any{
		"last_login_at":    now,
		"last_login_ip":    clientIP,
		"update_user_id":   user.ID,
		"update_user_name": user.Username,
		"updated_at":       now,
	}); err != nil {
		return nil, err
	}
	_ = s.createLoginCredentialLog(username, user.ID, true, clientIP, userAgent, "")
	return ret, nil
}

func (s *authService) RefreshToken(refreshToken string, authCfg config.AuthConfig, clientIP, userAgent string) (*response.LoginResponse, error) {
	session, err := s.validateSessionToken(refreshToken, constants.TokenTypeRefresh)
	if err != nil {
		return nil, err
	}
	if session.RevokedAt != nil {
		return nil, errorsx.InvalidToken("refresh token 已失效")
	}

	user := UserService.Get(session.UserID)
	if user == nil || user.Status != enums.StatusOk {
		return nil, errorsx.Unauthorized("用户不存在或已被禁用")
	}

	now := time.Now()
	if err = LoginSessionService.Updates(session.ID, map[string]any{
		"revoked_at":       now,
		"update_user_id":   user.ID,
		"update_user_name": user.Username,
		"updated_at":       now,
	}); err != nil {
		return nil, err
	}
	return s.issueTokens(user, clientIP, userAgent, authCfg)
}

func (s *authService) Logout(accessToken, refreshToken string) error {
	accessToken = s.extractBearerToken(accessToken)
	now := time.Now()
	if accessToken != "" {
		if session := LoginSessionService.FindOne(sqls.NewCnd().Eq("token_id", accessToken).Eq("token_type", constants.TokenTypeAccess)); session != nil && session.RevokedAt == nil {
			if err := LoginSessionService.Updates(session.ID, map[string]any{
				"revoked_at": now,
				"updated_at": now,
			}); err != nil {
				return err
			}
		}
	}
	if refreshToken != "" {
		if session := LoginSessionService.FindOne(sqls.NewCnd().Eq("token_id", refreshToken).Eq("token_type", constants.TokenTypeRefresh)); session != nil && session.RevokedAt == nil {
			if err := LoginSessionService.Updates(session.ID, map[string]any{
				"revoked_at": now,
				"updated_at": now,
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *authService) Authenticate(ctx iris.Context) (*dto.AuthPrincipal, error) {
	if principal := s.GetAuthPrincipal(ctx); principal != nil {
		return principal, nil
	}

	token := s.extractBearerToken(ctx.GetHeader("Authorization"))
	if token == "" {
		token = strings.TrimSpace(ctx.URLParam("accessToken"))
	}
	if token == "" {
		return nil, errorsx.Unauthorized("未登录或登录已过期")
	}

	session, err := s.validateSessionToken(token, constants.TokenTypeAccess)
	if err != nil {
		return nil, err
	}

	user := UserService.Get(session.UserID)
	if user == nil || user.Status != enums.StatusOk {
		return nil, errorsx.Unauthorized("用户不存在或已被禁用")
	}

	roles, permissions, err := s.loadUserAuthScope(user.ID)
	if err != nil {
		return nil, err
	}
	principal := s.setAuthPrincipal(ctx, user, roles, permissions)

	now := time.Now()
	_ = LoginSessionService.Updates(session.ID, map[string]any{
		"last_seen_at": now,
		"updated_at":   now,
	})

	return principal, nil
}

func (s *authService) HasPermission(ctx iris.Context, permissionCode string) bool {
	principal := s.GetAuthPrincipal(ctx)
	if principal == nil {
		return false
	}
	if slices.Contains(principal.Roles, constants.RoleCodeSuperAdmin) {
		return true
	}
	return slices.Contains(principal.Permissions, permissionCode)
}

func (s *authService) CurrentProfile(ctx iris.Context) (*response.LoginResponse, error) {
	principal, err := s.Authenticate(ctx)
	if err != nil {
		return nil, err
	}

	return &response.LoginResponse{
		User: &response.AuthUserResponse{
			ID:       principal.UserID,
			Username: principal.Username,
			Nickname: principal.Nickname,
			Avatar:   principal.Avatar,
			Status:   principal.Status,
			Roles:    principal.Roles,
		},
		Permissions: principal.Permissions,
		Roles:       principal.Roles,
	}, nil
}

func (s *authService) GetUserRoles(userID int64) ([]string, error) {
	return s.loadUserRoleCodes(userID)
}

func (s *authService) GetUserPermissions(userID int64) ([]string, error) {
	return s.loadUserPermissionCodes(userID)
}

func (s *authService) issueTokens(user *models.User, clientIP, userAgent string, authCfg config.AuthConfig) (*response.LoginResponse, error) {
	roles, permissions, err := s.loadUserAuthScope(user.ID)
	if err != nil {
		return nil, err
	}

	accessTTL, refreshTTL := s.resolveTokenTTL(authCfg)
	now := time.Now()
	accessToken, err := randomToken(constants.AccessTokenPrefix)
	if err != nil {
		return nil, err
	}
	refreshToken, err := randomToken(constants.RefreshTokenPrefix)
	if err != nil {
		return nil, err
	}

	if err = sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		accessSession := &models.LoginSession{
			UserID:     user.ID,
			TokenID:    accessToken,
			TokenType:  constants.TokenTypeAccess,
			ClientType: constants.ClientTypeAdminWeb,
			ClientIP:   clientIP,
			UserAgent:  userAgent,
			ExpiredAt:  now.Add(accessTTL),
			LastSeenAt: &now,
			AuditFields: models.AuditFields{
				CreatedAt:      now,
				CreateUserID:   user.ID,
				CreateUserName: user.Username,
				UpdatedAt:      now,
				UpdateUserID:   user.ID,
				UpdateUserName: user.Username,
			},
		}
		if err := ctx.Tx.Create(accessSession).Error; err != nil {
			return err
		}

		refreshSession := &models.LoginSession{
			UserID:     user.ID,
			TokenID:    refreshToken,
			TokenType:  constants.TokenTypeRefresh,
			ClientType: constants.ClientTypeAdminWeb,
			ClientIP:   clientIP,
			UserAgent:  userAgent,
			ExpiredAt:  now.Add(refreshTTL),
			LastSeenAt: &now,
			AuditFields: models.AuditFields{
				CreatedAt:      now,
				CreateUserID:   user.ID,
				CreateUserName: user.Username,
				UpdatedAt:      now,
				UpdateUserID:   user.ID,
				UpdateUserName: user.Username,
			},
		}
		return ctx.Tx.Create(refreshSession).Error
	}); err != nil {
		return nil, err
	}

	return &response.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    now.Add(accessTTL).Format(time.DateTime),
		User: &response.AuthUserResponse{
			ID:       user.ID,
			Username: user.Username,
			Nickname: user.Nickname,
			Avatar:   user.Avatar,
			Status:   user.Status,
			Roles:    roles,
		},
		Permissions: permissions,
		Roles:       roles,
	}, nil
}

func (s *authService) resolveTokenTTL(authCfg config.AuthConfig) (time.Duration, time.Duration) {
	accessTTL := 12 * time.Hour
	refreshTTL := 7 * 24 * time.Hour
	if authCfg.AccessTokenTTLHours > 0 {
		accessTTL = time.Duration(authCfg.AccessTokenTTLHours) * time.Hour
	}
	if authCfg.RefreshTokenTTLDays > 0 {
		refreshTTL = time.Duration(authCfg.RefreshTokenTTLDays) * 24 * time.Hour
	}
	return accessTTL, refreshTTL
}

func (s *authService) validateSessionToken(token, tokenType string) (*models.LoginSession, error) {
	if strings.TrimSpace(token) == "" {
		return nil, errorsx.InvalidToken("token 不能为空")
	}
	session := LoginSessionService.FindOne(sqls.NewCnd().
		Eq("token_id", token).
		Eq("token_type", tokenType))
	if session == nil {
		return nil, errorsx.InvalidToken("token 无效")
	}
	if session.RevokedAt != nil {
		return nil, errorsx.InvalidToken("token 已失效")
	}
	if time.Now().After(session.ExpiredAt) {
		return nil, errorsx.InvalidToken("token 已过期")
	}
	return session, nil
}

func (s *authService) loadUserAuthScope(userID int64) ([]string, []string, error) {
	roleCodes, err := s.loadUserRoleCodes(userID)
	if err != nil {
		return nil, nil, err
	}
	permissionCodes, err := s.loadUserPermissionCodes(userID)
	if err != nil {
		return nil, nil, err
	}
	return roleCodes, permissionCodes, nil
}

func (s *authService) loadUserRoleCodes(userID int64) ([]string, error) {
	roleRows := make([]struct {
		Code string
	}, 0)
	if err := sqls.DB().
		Table("t_role AS r").
		Select("r.code").
		Joins("JOIN t_user_role AS ur ON ur.role_id = r.id").
		Where("ur.user_id = ? AND r.status = ?", userID, enums.StatusOk).
		Order("r.sort_no ASC, r.id ASC").
		Scan(&roleRows).Error; err != nil {
		return nil, err
	}

	roleCodes := make([]string, 0, len(roleRows))
	for _, role := range roleRows {
		roleCodes = append(roleCodes, role.Code)
	}
	return roleCodes, nil
}

func (s *authService) loadUserPermissionCodes(userID int64) ([]string, error) {
	isSuperAdmin, err := s.hasSuperAdminRole(userID)
	if err != nil {
		return nil, err
	}

	permissionRows := make([]struct {
		Code string
	}, 0)
	db := sqls.DB().Table("t_permission AS p").Select("DISTINCT p.code").Where("p.status = ?", enums.StatusOk)
	if !isSuperAdmin {
		db = db.Joins("JOIN t_role_permission AS rp ON rp.permission_id = p.id").
			Joins("JOIN t_user_role AS ur ON ur.role_id = rp.role_id").
			Where("ur.user_id = ?", userID)
	}
	if err := db.Order("p.sort_no ASC, p.id ASC").Scan(&permissionRows).Error; err != nil {
		return nil, err
	}

	permissionCodes := make([]string, 0, len(permissionRows))
	for _, permission := range permissionRows {
		permissionCodes = append(permissionCodes, permission.Code)
	}

	overrideRows := make([]struct {
		Code   string
		Effect int
	}, 0)
	if err := sqls.DB().
		Table("t_user_permission AS up").
		Select("p.code, up.effect").
		Joins("JOIN t_permission AS p ON p.id = up.permission_id").
		Where("up.user_id = ? AND (up.expired_at IS NULL OR up.expired_at > ?)", userID, time.Now()).
		Scan(&overrideRows).Error; err != nil {
		return nil, err
	}

	permissionSet := make(map[string]bool, len(permissionCodes))
	for _, code := range permissionCodes {
		permissionSet[code] = true
	}
	for _, override := range overrideRows {
		if override.Effect < 0 {
			delete(permissionSet, override.Code)
			continue
		}
		permissionSet[override.Code] = true
	}

	permissionCodes = permissionCodes[:0]
	for code := range permissionSet {
		permissionCodes = append(permissionCodes, code)
	}
	sort.Strings(permissionCodes)
	return permissionCodes, nil
}

func (s *authService) hasSuperAdminRole(userID int64) (bool, error) {
	var count int64
	if err := sqls.DB().
		Table("t_role AS r").
		Joins("JOIN t_user_role AS ur ON ur.role_id = r.id").
		Where("ur.user_id = ? AND r.status = ? AND r.code = ?", userID, enums.StatusOk, constants.RoleCodeSuperAdmin).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *authService) extractBearerToken(header string) string {
	header = strings.TrimSpace(header)
	if header == "" {
		return ""
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func (s *authService) createLoginCredentialLog(principal string, userID int64, success bool, clientIP, userAgent, reason string) error {
	return LoginCredentialLogService.Create(&models.LoginCredentialLog{
		Principal: principal,
		UserID:    userID,
		Success:   success,
		ClientIP:  clientIP,
		UserAgent: userAgent,
		Reason:    reason,
		CreatedAt: time.Now(),
	})
}

func randomToken(prefix string) (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return prefix + hex.EncodeToString(buf), nil
}
