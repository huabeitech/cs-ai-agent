package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/errorsx"
	"cs-agent/internal/pkg/utils"
	"cs-agent/internal/repositories"
	"slices"
	"strings"
	"time"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"golang.org/x/crypto/bcrypt"
)

var UserService = newUserService()

func newUserService() *userService {
	return &userService{}
}

type userService struct {
}

func (s *userService) Get(id int64) *models.User {
	return repositories.UserRepository.Get(sqls.DB(), id)
}

func (s *userService) Take(where ...interface{}) *models.User {
	return repositories.UserRepository.Take(sqls.DB(), where...)
}

func (s *userService) Find(cnd *sqls.Cnd) []models.User {
	return repositories.UserRepository.Find(sqls.DB(), cnd)
}

func (s *userService) FindOne(cnd *sqls.Cnd) *models.User {
	return repositories.UserRepository.FindOne(sqls.DB(), cnd)
}

func (s *userService) FindPageByParams(params *params.QueryParams) (list []models.User, paging *sqls.Paging) {
	return repositories.UserRepository.FindPageByParams(sqls.DB(), params)
}

func (s *userService) FindPageByCnd(cnd *sqls.Cnd) (list []models.User, paging *sqls.Paging) {
	return repositories.UserRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *userService) Count(cnd *sqls.Cnd) int64 {
	return repositories.UserRepository.Count(sqls.DB(), cnd)
}

func (s *userService) FindByIds(ids []int64) []models.User {
	return repositories.UserRepository.FindByIds(sqls.DB(), ids)
}

func (s *userService) Create(t *models.User) error {
	return repositories.UserRepository.Create(sqls.DB(), t)
}

func (s *userService) Update(t *models.User) error {
	return repositories.UserRepository.Update(sqls.DB(), t)
}

func (s *userService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.UserRepository.Updates(sqls.DB(), id, columns)
}

func (s *userService) UpdateColumn(id int64, name string, value interface{}) error {
	return repositories.UserRepository.UpdateColumn(sqls.DB(), id, name, value)
}

func (s *userService) CreateUser(req request.CreateUserRequest, operator *dto.AuthPrincipal) (*models.User, error) {
	username := strings.TrimSpace(req.Username)
	if username == "" {
		return nil, errorsx.InvalidParam("用户名不能为空")
	}
	if strings.TrimSpace(req.Password) == "" {
		return nil, errorsx.InvalidParam("密码不能为空")
	}
	if s.Take("username = ? AND status != ?", username, enums.StatusDisabled) != nil {
		return nil, errorsx.InvalidParam("用户名已存在")
	}

	mobile := utils.NormalizeNullableString(req.Mobile)
	email := utils.NormalizeNullableString(req.Email)
	if mobile != nil && s.Take("mobile = ? AND status != ?", *mobile, enums.StatusDisabled) != nil {
		return nil, errorsx.InvalidParam("手机号已存在")
	}
	if email != nil && s.Take("email = ? AND status != ?", *email, enums.StatusDisabled) != nil {
		return nil, errorsx.InvalidParam("邮箱已存在")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Username:     username,
		Nickname:     strings.TrimSpace(req.Nickname),
		Password:     string(passwordHash),
		Avatar:       strings.TrimSpace(req.Avatar),
		Mobile:       mobile,
		Email:        email,
		Status:       enums.StatusOk,
		Remark:       strings.TrimSpace(req.Remark),
		PasswordSalt: "",
		AuditFields:  utils.BuildAuditFields(operator),
	}
	if user.Nickname == "" {
		user.Nickname = username
	}

	if err = s.replaceUserRoles(user.ID, req.RoleIDs, operator); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userService) UpdateUser(req request.UpdateUserRequest, operator *dto.AuthPrincipal) error {
	user := s.Get(req.ID)
	if user == nil || user.DeletedAt != nil {
		return errorsx.InvalidParam("用户不存在")
	}

	mobile := utils.NormalizeNullableString(req.Mobile)
	email := utils.NormalizeNullableString(req.Email)
	if mobile != nil {
		if existed := s.Take("mobile = ? AND id <> ? AND status != ?", *mobile, req.ID, enums.StatusDisabled); existed != nil {
			return errorsx.InvalidParam("手机号已存在")
		}
	}
	if email != nil {
		if existed := s.Take("email = ? AND id <> ? AND status != ?", *email, req.ID, enums.StatusDisabled); existed != nil {
			return errorsx.InvalidParam("邮箱已存在")
		}
	}

	return s.Updates(req.ID, map[string]any{
		"nickname":         strings.TrimSpace(req.Nickname),
		"avatar":           strings.TrimSpace(req.Avatar),
		"mobile":           mobile,
		"email":            email,
		"remark":           strings.TrimSpace(req.Remark),
		"update_user_id":   operator.UserID,
		"update_user_name": operator.Username,
		"updated_at":       time.Now(),
	})
}

func (s *userService) DeleteUser(id int64, operator *dto.AuthPrincipal) error {
	user := s.Get(id)
	if user == nil {
		return errorsx.InvalidParam("用户不存在")
	}

	if err := s.Updates(id, map[string]any{
		"status":           enums.StatusDisabled,
		"deleted_at":       time.Now(),
		"update_user_id":   operator.UserID,
		"update_user_name": operator.Username,
		"updated_at":       time.Now(),
	}); err != nil {
		return err
	}
	return LoginSessionService.RevokeByUser(id, operator.UserID, operator.Username)
}

func (s *userService) UpdateStatus(id int64, status int, operator *dto.AuthPrincipal) error {
	user := s.Get(id)
	if user == nil {
		return errorsx.InvalidParam("用户不存在")
	}
	if !slices.Contains(enums.StatusValues, enums.Status(status)) {
		return errorsx.InvalidParam("状态值不合法")
	}
	if err := s.Updates(id, map[string]any{
		"status":           status,
		"update_user_id":   operator.UserID,
		"update_user_name": operator.Username,
		"updated_at":       time.Now(),
	}); err != nil {
		return err
	}
	if status == int(enums.StatusDisabled) || status == int(enums.StatusDeleted) {
		return LoginSessionService.RevokeByUser(id, operator.UserID, operator.Username)
	}
	return nil
}

func (s *userService) ResetPassword(userID int64, operator *dto.AuthPrincipal) (string, error) {
	password, err := utils.GenerateRandomPassword(12)
	if err != nil {
		return "", err
	}
	if err = s.changePassword(userID, password, operator); err != nil {
		return "", err
	}
	return password, nil
}

func (s *userService) ChangeOwnPassword(password string, operator *dto.AuthPrincipal) error {
	if operator == nil || operator.UserID <= 0 {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	return s.changePassword(operator.UserID, password, operator)
}

func (s *userService) AssignRoles(userID int64, roleIDs []int64, operator *dto.AuthPrincipal) error {
	user := s.Get(userID)
	if user == nil || user.DeletedAt != nil {
		return errorsx.InvalidParam("用户不存在")
	}
	if err := s.replaceUserRoles(userID, roleIDs, operator); err != nil {
		return err
	}
	return LoginSessionService.RevokeByUser(userID, operator.UserID, operator.Username)
}

func (s *userService) replaceUserRoles(userID int64, roleIDs []int64, operator *dto.AuthPrincipal) error {
	return sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		if err := ctx.Tx.Where("user_id = ?", userID).Delete(&models.UserRole{}).Error; err != nil {
			return err
		}
		for _, roleID := range roleIDs {
			role := RoleService.Get(roleID)
			if role == nil {
				return errorsx.InvalidParam("角色不存在")
			}
			if role.Status != enums.StatusOk {
				return errorsx.InvalidParam("禁用角色不允许分配")
			}
			relation := &models.UserRole{
				UserID:      userID,
				RoleID:      roleID,
				AuditFields: utils.BuildAuditFields(operator),
			}
			if err := ctx.Tx.Create(relation).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *userService) changePassword(userID int64, password string, operator *dto.AuthPrincipal) error {
	user := s.Get(userID)
	if user == nil || user.DeletedAt != nil {
		return errorsx.InvalidParam("用户不存在")
	}
	if strings.TrimSpace(password) == "" {
		return errorsx.InvalidParam("新密码不能为空")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	now := time.Now()
	if err = s.Updates(userID, map[string]any{
		"password":         string(passwordHash),
		"update_user_id":   operator.UserID,
		"update_user_name": operator.Username,
		"updated_at":       now,
	}); err != nil {
		return err
	}
	return LoginSessionService.RevokeByUser(userID, operator.UserID, operator.Username)
}
