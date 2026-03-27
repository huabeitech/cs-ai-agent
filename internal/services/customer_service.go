package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/repositories"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
)

var CustomerService = newCustomerService()

func newCustomerService() *customerService {
	return &customerService{}
}

type customerService struct {
}

func (s *customerService) Get(id int64) *models.Customer {
	return repositories.CustomerRepository.Get(sqls.DB(), id)
}

func (s *customerService) Take(where ...interface{}) *models.Customer {
	return repositories.CustomerRepository.Take(sqls.DB(), where...)
}

func (s *customerService) Find(cnd *sqls.Cnd) []models.Customer {
	return repositories.CustomerRepository.Find(sqls.DB(), cnd)
}

func (s *customerService) FindOne(cnd *sqls.Cnd) *models.Customer {
	return repositories.CustomerRepository.FindOne(sqls.DB(), cnd)
}

func (s *customerService) FindPageByParams(params *params.QueryParams) (list []models.Customer, paging *sqls.Paging) {
	return repositories.CustomerRepository.FindPageByParams(sqls.DB(), params)
}

func (s *customerService) FindPageByCnd(cnd *sqls.Cnd) (list []models.Customer, paging *sqls.Paging) {
	return repositories.CustomerRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *customerService) Count(cnd *sqls.Cnd) int64 {
	return repositories.CustomerRepository.Count(sqls.DB(), cnd)
}

func (s *customerService) Create(t *models.Customer) error {
	return repositories.CustomerRepository.Create(sqls.DB(), t)
}

func (s *customerService) Update(t *models.Customer) error {
	return repositories.CustomerRepository.Update(sqls.DB(), t)
}

func (s *customerService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.CustomerRepository.Updates(sqls.DB(), id, columns)
}

func (s *customerService) UpdateColumn(id int64, name string, value interface{}) error {
	return repositories.CustomerRepository.UpdateColumn(sqls.DB(), id, name, value)
}

func (s *customerService) Delete(id int64) {
	repositories.CustomerRepository.Delete(sqls.DB(), id)
}

