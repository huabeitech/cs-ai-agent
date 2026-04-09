package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/repositories"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
)

var CustomerIdentityService = newCustomerIdentityService()

func newCustomerIdentityService() *customerIdentityService {
	return &customerIdentityService{}
}

type customerIdentityService struct {
}

func (s *customerIdentityService) Get(id int64) *models.CustomerIdentity {
	return repositories.CustomerIdentityRepository.Get(sqls.DB(), id)
}

func (s *customerIdentityService) Take(where ...interface{}) *models.CustomerIdentity {
	return repositories.CustomerIdentityRepository.Take(sqls.DB(), where...)
}

func (s *customerIdentityService) Find(cnd *sqls.Cnd) []models.CustomerIdentity {
	return repositories.CustomerIdentityRepository.Find(sqls.DB(), cnd)
}

func (s *customerIdentityService) FindOne(cnd *sqls.Cnd) *models.CustomerIdentity {
	return repositories.CustomerIdentityRepository.FindOne(sqls.DB(), cnd)
}

func (s *customerIdentityService) FindPageByParams(params *params.QueryParams) (list []models.CustomerIdentity, paging *sqls.Paging) {
	return repositories.CustomerIdentityRepository.FindPageByParams(sqls.DB(), params)
}

func (s *customerIdentityService) FindPageByCnd(cnd *sqls.Cnd) (list []models.CustomerIdentity, paging *sqls.Paging) {
	return repositories.CustomerIdentityRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *customerIdentityService) Count(cnd *sqls.Cnd) int64 {
	return repositories.CustomerIdentityRepository.Count(sqls.DB(), cnd)
}

func (s *customerIdentityService) Create(t *models.CustomerIdentity) error {
	return repositories.CustomerIdentityRepository.Create(sqls.DB(), t)
}

func (s *customerIdentityService) Update(t *models.CustomerIdentity) error {
	return repositories.CustomerIdentityRepository.Update(sqls.DB(), t)
}

func (s *customerIdentityService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.CustomerIdentityRepository.Updates(sqls.DB(), id, columns)
}

func (s *customerIdentityService) UpdateColumn(id int64, name string, value interface{}) error {
	return repositories.CustomerIdentityRepository.UpdateColumn(sqls.DB(), id, name, value)
}

func (s *customerIdentityService) Delete(id int64) {
	repositories.CustomerIdentityRepository.Delete(sqls.DB(), id)
}

