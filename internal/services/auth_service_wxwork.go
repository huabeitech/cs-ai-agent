package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/config"
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/errorsx"
	"cs-agent/internal/wxwork"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/mlogclub/simple/sqls"
)

const (
	wxWorkProviderName         = "wxwork"
	wxWorkProviderDisplayName  = "企业微信"
	wxWorkStateTTL             = 5 * time.Minute
	wxWorkLoginTicketTTL       = 1 * time.Minute
	defaultWxWorkLoginNextPath = "/dashboard"
)

var (
	errWxWorkStateInvalid  = errors.New("企业微信登录状态无效")
	wxWorkLoginTicketStore sync.Map
)

type wxWorkStatePayload struct {
	Next      string `json:"next"`
	Nonce     string `json:"nonce"`
	ExpiredAt int64  `json:"expiredAt"`
}

type wxWorkLoginTicket struct {
	Response  *response.LoginResponse
	ExpiredAt time.Time
}

func (s *authService) BuildWxWorkLoginURL(next string) (string, error) {
	if !wxwork.Enabled() {
		return "", errorsx.BusinessError(1, "企业微信登录未启用")
	}
	state, err := s.createWxWorkState(next)
	if err != nil {
		return "", err
	}
	return wxwork.BuildLoginURL(state)
}

func (s *authService) BuildWxWorkQRCodeLoginURL(next string) (string, error) {
	if !wxwork.Enabled() {
		return "", errorsx.BusinessError(1, "企业微信登录未启用")
	}
	state, err := s.createWxWorkState(next)
	if err != nil {
		return "", err
	}
	return wxwork.BuildQRCodeLoginURL(state)
}

func (s *authService) LoginByWxWork(code, state string, authCfg config.AuthConfig, clientIP, userAgent string) (string, string, error) {
	next, err := s.parseWxWorkState(state)
	if err != nil {
		return "", "", errorsx.Unauthorized("企业微信登录状态无效或已过期")
	}
	profile, err := wxwork.GetLoginUser(code)
	if err != nil {
		return "", "", err
	}
	loginResp, err := s.loginWithWxWorkProfile(profile, authCfg, clientIP, userAgent)
	if err != nil {
		return "", "", err
	}
	ticket, err := s.issueWxWorkLoginTicket(loginResp)
	if err != nil {
		return "", "", err
	}
	return ticket, next, nil
}

func (s *authService) ExchangeWxWorkLoginTicket(ticket string) (*response.LoginResponse, error) {
	return s.consumeWxWorkLoginTicket(ticket)
}

func (s *authService) loginWithWxWorkProfile(profile *wxwork.LoginUser, authCfg config.AuthConfig, clientIP, userAgent string) (*response.LoginResponse, error) {
	if profile == nil || strings.TrimSpace(profile.UserID) == "" {
		return nil, errorsx.BusinessError(2, "企业微信用户信息不存在")
	}
	now := time.Now()
	identity := UserIdentityService.FindOne(sqls.NewCnd().
		Eq("provider", wxWorkProviderName).
		Eq("provider_corp_id", strings.TrimSpace(profile.CorpID)).
		Eq("provider_user_id", strings.TrimSpace(profile.UserID)))

	var user *models.User
	var err error
	if identity == nil {
		user, identity, err = s.createWxWorkUser(profile, now)
		if err != nil {
			return nil, err
		}
	} else {
		if identity.Status != enums.StatusOk {
			return nil, errorsx.BusinessError(3, "当前企业微信绑定已停用")
		}
		user = UserService.Get(identity.UserID)
		if user == nil {
			return nil, errorsx.BusinessError(4, "企业微信账号绑定的系统用户不存在")
		}
	}

	if user.Status != enums.StatusOk {
		return nil, errorsx.Unauthorized("当前系统账号已被禁用")
	}

	ret, err := s.issueTokens(user, clientIP, userAgent, authCfg)
	if err != nil {
		return nil, err
	}

	rawProfile, _ := json.Marshal(profile)
	if err = sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		if err := ctx.Tx.Model(&models.User{}).Where("id = ?", user.ID).Updates(map[string]any{
			"nickname":         s.resolveWxWorkNickname(user.Nickname, profile),
			"avatar":           s.resolveWxWorkAvatar(user.Avatar, profile),
			"last_login_at":    now,
			"last_login_ip":    clientIP,
			"update_user_id":   user.ID,
			"update_user_name": user.Username,
			"updated_at":       now,
		}).Error; err != nil {
			return err
		}
		return ctx.Tx.Model(&models.UserIdentity{}).Where("id = ?", identity.ID).Updates(map[string]any{
			"raw_profile":      string(rawProfile),
			"last_auth_at":     now,
			"status":           enums.StatusOk,
			"update_user_id":   user.ID,
			"update_user_name": user.Username,
			"updated_at":       now,
		}).Error
	}); err != nil {
		return nil, err
	}
	return ret, nil
}

func (s *authService) createWxWorkUser(profile *wxwork.LoginUser, now time.Time) (*models.User, *models.UserIdentity, error) {
	rawProfile, _ := json.Marshal(profile)
	user := &models.User{
		Username:     s.buildWxWorkUsername(profile.CorpID, profile.UserID),
		Nickname:     s.resolveWxWorkNickname("", profile),
		Avatar:       s.resolveWxWorkAvatar("", profile),
		Mobile:       s.normalizeAvailableContact(profile.Mobile, "mobile"),
		Email:        s.normalizeAvailableContact(s.firstNonEmpty(profile.Email, profile.BizMail), "email"),
		Password:     "",
		PasswordSalt: "",
		Status:       enums.StatusOk,
		AuditFields: models.AuditFields{
			CreatedAt:      now,
			CreateUserID:   0,
			CreateUserName: wxWorkProviderDisplayName,
			UpdatedAt:      now,
			UpdateUserID:   0,
			UpdateUserName: wxWorkProviderDisplayName,
		},
	}

	var createdIdentity *models.UserIdentity
	err := sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		if err := ctx.Tx.Create(user).Error; err != nil {
			return err
		}

		identity := &models.UserIdentity{
			UserID:         user.ID,
			Provider:       wxWorkProviderName,
			ProviderUserID: strings.TrimSpace(profile.UserID),
			ProviderCorpID: strings.TrimSpace(profile.CorpID),
			ProviderName:   wxWorkProviderDisplayName,
			RawProfile:     string(rawProfile),
			Status:         enums.StatusOk,
			LastAuthAt:     &now,
			AuditFields: models.AuditFields{
				CreatedAt:      now,
				CreateUserID:   user.ID,
				CreateUserName: user.Username,
				UpdatedAt:      now,
				UpdateUserID:   user.ID,
				UpdateUserName: user.Username,
			},
		}
		if unionID := strings.TrimSpace(profile.OpenID); unionID != "" {
			identity.ProviderUnionID = &unionID
		}
		if err := ctx.Tx.Create(identity).Error; err != nil {
			return err
		}
		createdIdentity = identity
		return nil
	})
	if err != nil {
		if existing := UserIdentityService.FindOne(sqls.NewCnd().
			Eq("provider", wxWorkProviderName).
			Eq("provider_corp_id", strings.TrimSpace(profile.CorpID)).
			Eq("provider_user_id", strings.TrimSpace(profile.UserID))); existing != nil {
			existingUser := UserService.Get(existing.UserID)
			if existingUser != nil {
				return existingUser, existing, nil
			}
		}
		return nil, nil, err
	}
	return user, createdIdentity, nil
}

func (s *authService) buildWxWorkUsername(corpID, userID string) string {
	raw := strings.ToLower(strings.TrimSpace(corpID) + "_" + strings.TrimSpace(userID))
	raw = strings.NewReplacer("@", "_", ".", "_", "-", "_", " ", "_", "/", "_", "\\", "_", ":", "_").Replace(raw)
	raw = strings.Trim(raw, "_")
	base := "wxwork_" + raw
	if len(base) <= 100 && UserService.Take("username = ?", base) == nil {
		return base
	}
	sum := sha256.Sum256([]byte(strings.TrimSpace(corpID) + ":" + strings.TrimSpace(userID)))
	return "wxwork_" + hex.EncodeToString(sum[:16])
}

func (s *authService) resolveWxWorkNickname(current string, profile *wxwork.LoginUser) string {
	if profile != nil {
		if name := strings.TrimSpace(profile.Name); name != "" {
			return name
		}
	}
	if current = strings.TrimSpace(current); current != "" {
		return current
	}
	if profile != nil {
		return strings.TrimSpace(profile.UserID)
	}
	return ""
}

func (s *authService) resolveWxWorkAvatar(current string, profile *wxwork.LoginUser) string {
	if profile != nil {
		if avatar := strings.TrimSpace(profile.Avatar); avatar != "" {
			return avatar
		}
	}
	return strings.TrimSpace(current)
}

func (s *authService) normalizeAvailableContact(value string, field string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	switch field {
	case "mobile":
		if UserService.Take("mobile = ? AND status != ?", value, enums.StatusDisabled) != nil {
			return nil
		}
	case "email":
		if UserService.Take("email = ? AND status != ?", value, enums.StatusDisabled) != nil {
			return nil
		}
	}
	return &value
}

func (s *authService) createWxWorkState(next string) (string, error) {
	secret := strings.TrimSpace(wxwork.StateSecret())
	if secret == "" {
		return "", errorsx.BusinessError(1, "企业微信登录密钥未配置")
	}
	payload := wxWorkStatePayload{
		Next:      sanitizeNextPath(next),
		ExpiredAt: time.Now().Add(wxWorkStateTTL).Unix(),
	}
	nonce, err := randomToken("ws_")
	if err != nil {
		return "", err
	}
	payload.Nonce = nonce

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	encoded := base64.RawURLEncoding.EncodeToString(body)
	return encoded + "." + s.signWxWorkState(encoded, secret), nil
}

func (s *authService) parseWxWorkState(state string) (string, error) {
	secret := strings.TrimSpace(wxwork.StateSecret())
	if secret == "" {
		return "", errWxWorkStateInvalid
	}
	parts := strings.Split(strings.TrimSpace(state), ".")
	if len(parts) != 2 {
		return "", errWxWorkStateInvalid
	}
	if !hmac.Equal([]byte(parts[1]), []byte(s.signWxWorkState(parts[0], secret))) {
		return "", errWxWorkStateInvalid
	}
	body, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", errWxWorkStateInvalid
	}
	payload := wxWorkStatePayload{}
	if err = json.Unmarshal(body, &payload); err != nil {
		return "", errWxWorkStateInvalid
	}
	if payload.ExpiredAt <= time.Now().Unix() {
		return "", errWxWorkStateInvalid
	}
	return sanitizeNextPath(payload.Next), nil
}

func (s *authService) signWxWorkState(content, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(content))
	return hex.EncodeToString(mac.Sum(nil))
}

func (s *authService) issueWxWorkLoginTicket(loginResp *response.LoginResponse) (string, error) {
	if loginResp == nil {
		return "", fmt.Errorf("登录结果不能为空")
	}
	ticket, err := randomToken("wlt_")
	if err != nil {
		return "", err
	}
	s.cleanupExpiredWxWorkLoginTickets()
	wxWorkLoginTicketStore.Store(ticket, wxWorkLoginTicket{
		Response:  loginResp,
		ExpiredAt: time.Now().Add(wxWorkLoginTicketTTL),
	})
	return ticket, nil
}

func (s *authService) consumeWxWorkLoginTicket(ticket string) (*response.LoginResponse, error) {
	ticket = strings.TrimSpace(ticket)
	if ticket == "" {
		return nil, errorsx.InvalidParam("ticket 不能为空")
	}
	value, ok := wxWorkLoginTicketStore.LoadAndDelete(ticket)
	if !ok {
		return nil, errorsx.Unauthorized("登录票据无效或已过期")
	}
	record, ok := value.(wxWorkLoginTicket)
	if !ok || record.Response == nil || time.Now().After(record.ExpiredAt) {
		return nil, errorsx.Unauthorized("登录票据无效或已过期")
	}
	return record.Response, nil
}

func (s *authService) cleanupExpiredWxWorkLoginTickets() {
	now := time.Now()
	wxWorkLoginTicketStore.Range(func(key, value any) bool {
		record, ok := value.(wxWorkLoginTicket)
		if !ok || now.After(record.ExpiredAt) {
			wxWorkLoginTicketStore.Delete(key)
		}
		return true
	})
}

func (s *authService) firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}

func sanitizeNextPath(next string) string {
	next = strings.TrimSpace(next)
	if next == "" || !strings.HasPrefix(next, "/") || strings.HasPrefix(next, "//") {
		return defaultWxWorkLoginNextPath
	}
	return next
}
