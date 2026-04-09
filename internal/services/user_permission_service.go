package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/repositories"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
)

var UserPermissionService = newUserPermissionService()

func newUserPermissionService() *userPermissionService {
	return &userPermissionService{}
}

type userPermissionService struct {
}

func (s *userPermissionService) Get(id int64) *models.UserPermission {
	return repositories.UserPermissionRepository.Get(sqls.DB(), id)
}

func (s *userPermissionService) Take(where ...interface{}) *models.UserPermission {
	return repositories.UserPermissionRepository.Take(sqls.DB(), where...)
}

func (s *userPermissionService) Find(cnd *sqls.Cnd) []models.UserPermission {
	return repositories.UserPermissionRepository.Find(sqls.DB(), cnd)
}

func (s *userPermissionService) FindOne(cnd *sqls.Cnd) *models.UserPermission {
	return repositories.UserPermissionRepository.FindOne(sqls.DB(), cnd)
}

func (s *userPermissionService) FindPageByParams(params *params.QueryParams) (list []models.UserPermission, paging *sqls.Paging) {
	return repositories.UserPermissionRepository.FindPageByParams(sqls.DB(), params)
}

func (s *userPermissionService) FindPageByCnd(cnd *sqls.Cnd) (list []models.UserPermission, paging *sqls.Paging) {
	return repositories.UserPermissionRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *userPermissionService) Count(cnd *sqls.Cnd) int64 {
	return repositories.UserPermissionRepository.Count(sqls.DB(), cnd)
}

func (s *userPermissionService) Create(t *models.UserPermission) error {
	return repositories.UserPermissionRepository.Create(sqls.DB(), t)
}

func (s *userPermissionService) Update(t *models.UserPermission) error {
	return repositories.UserPermissionRepository.Update(sqls.DB(), t)
}

func (s *userPermissionService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.UserPermissionRepository.Updates(sqls.DB(), id, columns)
}

func (s *userPermissionService) UpdateColumn(id int64, name string, value interface{}) error {
	return repositories.UserPermissionRepository.UpdateColumn(sqls.DB(), id, name, value)
}

func (s *userPermissionService) Delete(id int64) {
	repositories.UserPermissionRepository.Delete(sqls.DB(), id)
}
