package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/repositories"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
)

var RolePermissionService = newRolePermissionService()

func newRolePermissionService() *rolePermissionService {
	return &rolePermissionService{}
}

type rolePermissionService struct {
}

func (s *rolePermissionService) Get(id int64) *models.RolePermission {
	return repositories.RolePermissionRepository.Get(sqls.DB(), id)
}

func (s *rolePermissionService) Take(where ...interface{}) *models.RolePermission {
	return repositories.RolePermissionRepository.Take(sqls.DB(), where...)
}

func (s *rolePermissionService) Find(cnd *sqls.Cnd) []models.RolePermission {
	return repositories.RolePermissionRepository.Find(sqls.DB(), cnd)
}

func (s *rolePermissionService) FindOne(cnd *sqls.Cnd) *models.RolePermission {
	return repositories.RolePermissionRepository.FindOne(sqls.DB(), cnd)
}

func (s *rolePermissionService) FindPageByParams(params *params.QueryParams) (list []models.RolePermission, paging *sqls.Paging) {
	return repositories.RolePermissionRepository.FindPageByParams(sqls.DB(), params)
}

func (s *rolePermissionService) FindPageByCnd(cnd *sqls.Cnd) (list []models.RolePermission, paging *sqls.Paging) {
	return repositories.RolePermissionRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *rolePermissionService) Count(cnd *sqls.Cnd) int64 {
	return repositories.RolePermissionRepository.Count(sqls.DB(), cnd)
}

func (s *rolePermissionService) Create(t *models.RolePermission) error {
	return repositories.RolePermissionRepository.Create(sqls.DB(), t)
}

func (s *rolePermissionService) Update(t *models.RolePermission) error {
	return repositories.RolePermissionRepository.Update(sqls.DB(), t)
}

func (s *rolePermissionService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.RolePermissionRepository.Updates(sqls.DB(), id, columns)
}

func (s *rolePermissionService) UpdateColumn(id int64, name string, value interface{}) error {
	return repositories.RolePermissionRepository.UpdateColumn(sqls.DB(), id, name, value)
}

func (s *rolePermissionService) Delete(id int64) {
	repositories.RolePermissionRepository.Delete(sqls.DB(), id)
}
