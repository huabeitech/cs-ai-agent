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
)

var RoleService = newRoleService()

func newRoleService() *roleService {
	return &roleService{}
}

type roleService struct {
}

func (s *roleService) Get(id int64) *models.Role {
	return repositories.RoleRepository.Get(sqls.DB(), id)
}

func (s *roleService) Take(where ...interface{}) *models.Role {
	return repositories.RoleRepository.Take(sqls.DB(), where...)
}

func (s *roleService) Find(cnd *sqls.Cnd) []models.Role {
	return repositories.RoleRepository.Find(sqls.DB(), cnd)
}

func (s *roleService) FindOne(cnd *sqls.Cnd) *models.Role {
	return repositories.RoleRepository.FindOne(sqls.DB(), cnd)
}

func (s *roleService) FindPageByParams(params *params.QueryParams) (list []models.Role, paging *sqls.Paging) {
	return repositories.RoleRepository.FindPageByParams(sqls.DB(), params)
}

func (s *roleService) FindPageByCnd(cnd *sqls.Cnd) (list []models.Role, paging *sqls.Paging) {
	return repositories.RoleRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *roleService) Count(cnd *sqls.Cnd) int64 {
	return repositories.RoleRepository.Count(sqls.DB(), cnd)
}

func (s *roleService) Create(t *models.Role) error {
	return repositories.RoleRepository.Create(sqls.DB(), t)
}

func (s *roleService) Update(t *models.Role) error {
	return repositories.RoleRepository.Update(sqls.DB(), t)
}

func (s *roleService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.RoleRepository.Updates(sqls.DB(), id, columns)
}

func (s *roleService) UpdateColumn(id int64, name string, value interface{}) error {
	return repositories.RoleRepository.UpdateColumn(sqls.DB(), id, name, value)
}

func (s *roleService) Delete(id int64) {
	repositories.RoleRepository.Delete(sqls.DB(), id)
}

func (s *roleService) CreateRole(req request.CreateRoleRequest, operator *dto.AuthPrincipal) (*models.Role, error) {
	name := strings.TrimSpace(req.Name)
	code := strings.TrimSpace(req.Code)
	if name == "" || code == "" {
		return nil, errorsx.InvalidParam("角色名称和编码不能为空")
	}
	if s.Take("code = ?", code) != nil {
		return nil, errorsx.InvalidParam("角色编码已存在")
	}

	role := &models.Role{
		Name:        name,
		Code:        code,
		Status:      enums.StatusOk,
		IsSystem:    false,
		SortNo:      req.SortNo,
		Remark:      strings.TrimSpace(req.Remark),
		AuditFields: utils.BuildAuditFields(operator),
	}
	if err := s.Create(role); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *roleService) UpdateRole(req request.UpdateRoleRequest, operator *dto.AuthPrincipal) error {
	role := s.Get(req.ID)
	if role == nil {
		return errorsx.InvalidParam("角色不存在")
	}
	now := time.Now()
	return s.Updates(req.ID, map[string]any{
		"name":             strings.TrimSpace(req.Name),
		"sort_no":          req.SortNo,
		"remark":           strings.TrimSpace(req.Remark),
		"update_user_id":   operator.UserID,
		"update_user_name": operator.Username,
		"updated_at":       now,
	})
}

func (s *roleService) NextSortNo() int {
	if latest := s.FindOne(sqls.NewCnd().Asc("sort_no").Desc("id")); latest != nil {
		return latest.SortNo + 1
	}
	return 0
}

func (s *roleService) UpdateSort(ids []int64) error {
	return sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		for i, id := range ids {
			if err := repositories.RoleRepository.UpdateColumn(ctx.Tx, id, "sort_no", i); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *roleService) DeleteRole(id int64) error {
	role := s.Get(id)
	if role == nil {
		return errorsx.InvalidParam("角色不存在")
	}
	if role.IsSystem {
		return errorsx.Forbidden("系统内置角色不允许删除")
	}
	if UserRoleService.Take("role_id = ?", id) != nil {
		return errorsx.Forbidden("角色已被用户使用，无法删除")
	}
	s.Delete(id)
	return nil
}

func (s *roleService) UpdateStatus(id int64, status enums.Status, operator *dto.AuthPrincipal) error {
	role := s.Get(id)
	if role == nil {
		return errorsx.InvalidParam("角色不存在")
	}
	if !slices.Contains(enums.StatusValues, status) {
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
	return s.revokeRoleUsersSessions(id, operator)
}

func (s *roleService) AssignPermissions(roleID int64, permissionIDs []int64, operator *dto.AuthPrincipal) error {
	role := s.Get(roleID)
	if role == nil {
		return errorsx.InvalidParam("角色不存在")
	}

	if err := s.replaceRolePermissions(roleID, permissionIDs, operator); err != nil {
		return err
	}
	return s.revokeRoleUsersSessions(roleID, operator)
}

func (s *roleService) replaceRolePermissions(roleID int64, permissionIDs []int64, operator *dto.AuthPrincipal) error {
	return sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		if err := ctx.Tx.Where("role_id = ?", roleID).Delete(&models.RolePermission{}).Error; err != nil {
			return err
		}
		for _, permissionID := range permissionIDs {
			permission := PermissionService.Get(permissionID)
			if permission == nil {
				return errorsx.InvalidParam("权限不存在")
			}
			relation := &models.RolePermission{
				RoleID:       roleID,
				PermissionID: permissionID,
				AuditFields:  utils.BuildAuditFields(operator),
			}
			if err := ctx.Tx.Create(relation).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *roleService) revokeRoleUsersSessions(roleID int64, operator *dto.AuthPrincipal) error {
	userRoles := UserRoleService.Find(sqls.NewCnd().Eq("role_id", roleID))
	seen := make(map[int64]struct{}, len(userRoles))
	for _, item := range userRoles {
		if _, ok := seen[item.UserID]; ok {
			continue
		}
		seen[item.UserID] = struct{}{}
		if err := LoginSessionService.RevokeByUser(item.UserID, operator.UserID, operator.Username); err != nil {
			return err
		}
	}
	return nil
}
