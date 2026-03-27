package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/config"
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/errorsx"
	"cs-agent/internal/repositories"
	"cs-agent/internal/wxwork"
	"strings"
	"time"

	"github.com/mlogclub/simple/common/jsons"
	"github.com/mlogclub/simple/sqls"
	"gorm.io/gorm"
)

var WxWorkLoginService = &wxWorkLoginService{}

type wxWorkLoginService struct {
}

func (s *wxWorkLoginService) BuildWxWorkLoginURL(next string) (string, error) {
	if !wxwork.Enabled() {
		return "", errorsx.BusinessError(1, "企业微信登录未启用")
	}
	state, err := wxwork.CreateState(next)
	if err != nil {
		return "", err
	}
	return wxwork.BuildLoginURL(state)
}

func (s *wxWorkLoginService) BuildWxWorkQRCodeLoginURL(next string) (string, error) {
	if !wxwork.Enabled() {
		return "", errorsx.BusinessError(1, "企业微信登录未启用")
	}
	state, err := wxwork.CreateState(next)
	if err != nil {
		return "", err
	}
	return wxwork.BuildQRCodeLoginURL(state)
}

func (s *wxWorkLoginService) LoginByWxWork(code, state string, authCfg config.AuthConfig, clientIP, userAgent string) (string, string, error) {
	next, err := wxwork.ParseState(state)
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
	ticket, err := wxwork.IssueLoginTicket(loginResp)
	if err != nil {
		return "", "", err
	}
	return ticket, next, nil
}

func (s *wxWorkLoginService) ExchangeWxWorkLoginTicket(ticket string) (*response.LoginResponse, error) {
	return wxwork.ConsumeLoginTicket(ticket)
}

func (s *wxWorkLoginService) loginWithWxWorkProfile(profile *wxwork.LoginUser, authCfg config.AuthConfig, clientIP, userAgent string) (*response.LoginResponse, error) {
	if profile == nil || strings.TrimSpace(profile.UserID) == "" {
		return nil, errorsx.BusinessError(2, "企业微信用户信息不存在")
	}

	var (
		identity = UserIdentityService.GetBy(enums.ThirdProviderWxWork, profile.CorpID, profile.UserID)
		user     *models.User
		err      error
	)
	if identity == nil {
		user, identity, err = s.createWxWorkUser(profile)
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

	ret, err := AuthService.issueTokens(user, clientIP, userAgent, authCfg)
	if err != nil {
		return nil, err
	}

	if err = sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		if err := repositories.UserRepository.Updates(ctx.Tx, user.ID, map[string]any{
			"nickname":         s.resolveWxWorkNickname(user.Nickname, profile),
			"avatar":           s.resolveWxWorkAvatar(user.Avatar, profile),
			"last_login_at":    time.Now(),
			"last_login_ip":    clientIP,
			"update_user_id":   user.ID,
			"update_user_name": user.Username,
			"updated_at":       time.Now(),
		}); err != nil {
			return err
		}
		return repositories.UserIdentityRepository.Updates(ctx.Tx, identity.ID, map[string]any{
			"raw_profile":      jsons.ToJsonStr(profile),
			"last_auth_at":     time.Now(),
			"status":           enums.StatusOk,
			"update_user_id":   user.ID,
			"update_user_name": user.Username,
			"updated_at":       time.Now(),
		})
	}); err != nil {
		return nil, err
	}
	return ret, nil
}

func (s *wxWorkLoginService) createWxWorkUser(profile *wxwork.LoginUser) (*models.User, *models.UserIdentity, error) {
	username := s.buildWxWorkUsername(profile.UserID)
	mobileValue := strings.TrimSpace(profile.Mobile)
	emailValue := strings.TrimSpace(s.firstNonEmpty(profile.Email, profile.BizMail))

	user := &models.User{
		Username:     username,
		Nickname:     s.resolveWxWorkNickname("", profile),
		Avatar:       s.resolveWxWorkAvatar("", profile),
		Password:     "",
		PasswordSalt: "",
		Status:       enums.StatusOk,
		AuditFields: models.AuditFields{
			CreatedAt:      time.Now(),
			CreateUserID:   0,
			CreateUserName: enums.GetThirdProviderLabel(enums.ThirdProviderWxWork),
			UpdatedAt:      time.Now(),
			UpdateUserID:   0,
			UpdateUserName: enums.GetThirdProviderLabel(enums.ThirdProviderWxWork),
		},
	}

	var createdIdentity *models.UserIdentity
	err := sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		if existing := repositories.UserRepository.Take(ctx.Tx, "username = ?", username); existing != nil {
			return errorsx.BusinessError(5, "企业微信用户ID已被系统用户名占用")
		}
		var err error
		user.Mobile, err = s.normalizeAvailableContactTx(ctx.Tx, mobileValue, "mobile")
		if err != nil {
			return err
		}
		user.Email, err = s.normalizeAvailableContactTx(ctx.Tx, emailValue, "email")
		if err != nil {
			return err
		}
		if err := ctx.Tx.Create(user).Error; err != nil {
			return err
		}

		identity := &models.UserIdentity{
			UserID:         user.ID,
			Provider:       enums.ThirdProviderWxWork,
			ProviderUserID: strings.TrimSpace(profile.UserID),
			ProviderCorpID: strings.TrimSpace(profile.CorpID),
			ProviderName:   enums.GetThirdProviderLabel(enums.ThirdProviderWxWork),
			RawProfile:     jsons.ToJsonStr(profile),
			Status:         enums.StatusOk,
			LastAuthAt:     new(time.Now()),
			AuditFields: models.AuditFields{
				CreatedAt:      time.Now(),
				CreateUserID:   user.ID,
				CreateUserName: user.Username,
				UpdatedAt:      time.Now(),
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
			Eq("provider", enums.ThirdProviderWxWork).
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

func (s *wxWorkLoginService) buildWxWorkUsername(userID string) string {
	return strings.TrimSpace(userID)
}

func (s *wxWorkLoginService) resolveWxWorkNickname(current string, profile *wxwork.LoginUser) string {
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

func (s *wxWorkLoginService) resolveWxWorkAvatar(current string, profile *wxwork.LoginUser) string {
	if profile != nil {
		if avatar := strings.TrimSpace(profile.Avatar); avatar != "" {
			return avatar
		}
	}
	return strings.TrimSpace(current)
}

func (s *wxWorkLoginService) normalizeAvailableContactTx(tx *gorm.DB, value string, field string) (*string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	switch field {
	case "mobile":
		if repositories.UserRepository.Take(tx, "mobile = ? AND status != ?", value, enums.StatusDisabled) != nil {
			return nil, errorsx.BusinessError(6, "企业微信手机号已被系统用户占用")
		}
	case "email":
		if repositories.UserRepository.Take(tx, "email = ? AND status != ?", value, enums.StatusDisabled) != nil {
			return nil, errorsx.BusinessError(7, "企业微信邮箱已被系统用户占用")
		}
	}
	return &value, nil
}

func (s *wxWorkLoginService) firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}
