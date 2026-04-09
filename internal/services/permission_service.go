package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/repositories"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
)

var PermissionService = newPermissionService()

func newPermissionService() *permissionService {
	return &permissionService{}
}

type permissionService struct {
}

func (s *permissionService) Get(id int64) *models.Permission {
	return repositories.PermissionRepository.Get(sqls.DB(), id)
}

func (s *permissionService) Take(where ...interface{}) *models.Permission {
	return repositories.PermissionRepository.Take(sqls.DB(), where...)
}

func (s *permissionService) Find(cnd *sqls.Cnd) []models.Permission {
	return repositories.PermissionRepository.Find(sqls.DB(), cnd)
}

func (s *permissionService) FindOne(cnd *sqls.Cnd) *models.Permission {
	return repositories.PermissionRepository.FindOne(sqls.DB(), cnd)
}

func (s *permissionService) FindPageByParams(params *params.QueryParams) (list []models.Permission, paging *sqls.Paging) {
	return repositories.PermissionRepository.FindPageByParams(sqls.DB(), params)
}

func (s *permissionService) FindPageByCnd(cnd *sqls.Cnd) (list []models.Permission, paging *sqls.Paging) {
	return repositories.PermissionRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *permissionService) Count(cnd *sqls.Cnd) int64 {
	return repositories.PermissionRepository.Count(sqls.DB(), cnd)
}

func (s *permissionService) Create(t *models.Permission) error {
	return repositories.PermissionRepository.Create(sqls.DB(), t)
}

func (s *permissionService) Update(t *models.Permission) error {
	return repositories.PermissionRepository.Update(sqls.DB(), t)
}

func (s *permissionService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.PermissionRepository.Updates(sqls.DB(), id, columns)
}

func (s *permissionService) UpdateColumn(id int64, name string, value interface{}) error {
	return repositories.PermissionRepository.UpdateColumn(sqls.DB(), id, name, value)
}

func (s *permissionService) Delete(id int64) {
	repositories.PermissionRepository.Delete(sqls.DB(), id)
}
