package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/repositories"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
)

var UserIdentityService = newUserIdentityService()

func newUserIdentityService() *userIdentityService {
	return &userIdentityService{}
}

type userIdentityService struct {
}

func (s *userIdentityService) Get(id int64) *models.UserIdentity {
	return repositories.UserIdentityRepository.Get(sqls.DB(), id)
}

func (s *userIdentityService) Take(where ...interface{}) *models.UserIdentity {
	return repositories.UserIdentityRepository.Take(sqls.DB(), where...)
}

func (s *userIdentityService) Find(cnd *sqls.Cnd) []models.UserIdentity {
	return repositories.UserIdentityRepository.Find(sqls.DB(), cnd)
}

func (s *userIdentityService) FindOne(cnd *sqls.Cnd) *models.UserIdentity {
	return repositories.UserIdentityRepository.FindOne(sqls.DB(), cnd)
}

func (s *userIdentityService) FindPageByParams(params *params.QueryParams) (list []models.UserIdentity, paging *sqls.Paging) {
	return repositories.UserIdentityRepository.FindPageByParams(sqls.DB(), params)
}

func (s *userIdentityService) FindPageByCnd(cnd *sqls.Cnd) (list []models.UserIdentity, paging *sqls.Paging) {
	return repositories.UserIdentityRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *userIdentityService) Count(cnd *sqls.Cnd) int64 {
	return repositories.UserIdentityRepository.Count(sqls.DB(), cnd)
}

func (s *userIdentityService) Create(t *models.UserIdentity) error {
	return repositories.UserIdentityRepository.Create(sqls.DB(), t)
}

func (s *userIdentityService) Update(t *models.UserIdentity) error {
	return repositories.UserIdentityRepository.Update(sqls.DB(), t)
}

func (s *userIdentityService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.UserIdentityRepository.Updates(sqls.DB(), id, columns)
}

func (s *userIdentityService) UpdateColumn(id int64, name string, value interface{}) error {
	return repositories.UserIdentityRepository.UpdateColumn(sqls.DB(), id, name, value)
}

func (s *userIdentityService) Delete(id int64) {
	repositories.UserIdentityRepository.Delete(sqls.DB(), id)
}
