package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/repositories"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
)

var CompanyService = newCompanyService()

func newCompanyService() *companyService {
	return &companyService{}
}

type companyService struct {
}

func (s *companyService) Get(id int64) *models.Company {
	return repositories.CompanyRepository.Get(sqls.DB(), id)
}

func (s *companyService) Take(where ...interface{}) *models.Company {
	return repositories.CompanyRepository.Take(sqls.DB(), where...)
}

func (s *companyService) Find(cnd *sqls.Cnd) []models.Company {
	return repositories.CompanyRepository.Find(sqls.DB(), cnd)
}

func (s *companyService) FindOne(cnd *sqls.Cnd) *models.Company {
	return repositories.CompanyRepository.FindOne(sqls.DB(), cnd)
}

func (s *companyService) FindPageByParams(params *params.QueryParams) (list []models.Company, paging *sqls.Paging) {
	return repositories.CompanyRepository.FindPageByParams(sqls.DB(), params)
}

func (s *companyService) FindPageByCnd(cnd *sqls.Cnd) (list []models.Company, paging *sqls.Paging) {
	return repositories.CompanyRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *companyService) Count(cnd *sqls.Cnd) int64 {
	return repositories.CompanyRepository.Count(sqls.DB(), cnd)
}

func (s *companyService) Create(t *models.Company) error {
	return repositories.CompanyRepository.Create(sqls.DB(), t)
}

func (s *companyService) Update(t *models.Company) error {
	return repositories.CompanyRepository.Update(sqls.DB(), t)
}

func (s *companyService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.CompanyRepository.Updates(sqls.DB(), id, columns)
}

func (s *companyService) UpdateColumn(id int64, name string, value interface{}) error {
	return repositories.CompanyRepository.UpdateColumn(sqls.DB(), id, name, value)
}

func (s *companyService) Delete(id int64) {
	repositories.CompanyRepository.Delete(sqls.DB(), id)
}

