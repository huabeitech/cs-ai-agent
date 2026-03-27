package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/repositories"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
)

var CustomerContactService = newCustomerContactService()

func newCustomerContactService() *customerContactService {
	return &customerContactService{}
}

type customerContactService struct {
}

func (s *customerContactService) Get(id int64) *models.CustomerContact {
	return repositories.CustomerContactRepository.Get(sqls.DB(), id)
}

func (s *customerContactService) Take(where ...interface{}) *models.CustomerContact {
	return repositories.CustomerContactRepository.Take(sqls.DB(), where...)
}

func (s *customerContactService) Find(cnd *sqls.Cnd) []models.CustomerContact {
	return repositories.CustomerContactRepository.Find(sqls.DB(), cnd)
}

func (s *customerContactService) FindOne(cnd *sqls.Cnd) *models.CustomerContact {
	return repositories.CustomerContactRepository.FindOne(sqls.DB(), cnd)
}

func (s *customerContactService) FindPageByParams(params *params.QueryParams) (list []models.CustomerContact, paging *sqls.Paging) {
	return repositories.CustomerContactRepository.FindPageByParams(sqls.DB(), params)
}

func (s *customerContactService) FindPageByCnd(cnd *sqls.Cnd) (list []models.CustomerContact, paging *sqls.Paging) {
	return repositories.CustomerContactRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *customerContactService) Count(cnd *sqls.Cnd) int64 {
	return repositories.CustomerContactRepository.Count(sqls.DB(), cnd)
}

func (s *customerContactService) Create(t *models.CustomerContact) error {
	return repositories.CustomerContactRepository.Create(sqls.DB(), t)
}

func (s *customerContactService) Update(t *models.CustomerContact) error {
	return repositories.CustomerContactRepository.Update(sqls.DB(), t)
}

func (s *customerContactService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.CustomerContactRepository.Updates(sqls.DB(), id, columns)
}

func (s *customerContactService) UpdateColumn(id int64, name string, value interface{}) error {
	return repositories.CustomerContactRepository.UpdateColumn(sqls.DB(), id, name, value)
}

func (s *customerContactService) Delete(id int64) {
	repositories.CustomerContactRepository.Delete(sqls.DB(), id)
}

