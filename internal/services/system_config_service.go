package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/repositories"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
)

var SystemConfigService = newSystemConfigService()

func newSystemConfigService() *systemConfigService {
	return &systemConfigService{}
}

type systemConfigService struct {
}

func (s *systemConfigService) Get(id int64) *models.SystemConfig {
	return repositories.SystemConfigRepository.Get(sqls.DB(), id)
}

func (s *systemConfigService) Take(where ...interface{}) *models.SystemConfig {
	return repositories.SystemConfigRepository.Take(sqls.DB(), where...)
}

func (s *systemConfigService) Find(cnd *sqls.Cnd) []models.SystemConfig {
	return repositories.SystemConfigRepository.Find(sqls.DB(), cnd)
}

func (s *systemConfigService) FindOne(cnd *sqls.Cnd) *models.SystemConfig {
	return repositories.SystemConfigRepository.FindOne(sqls.DB(), cnd)
}

func (s *systemConfigService) FindPageByParams(params *params.QueryParams) (list []models.SystemConfig, paging *sqls.Paging) {
	return repositories.SystemConfigRepository.FindPageByParams(sqls.DB(), params)
}

func (s *systemConfigService) FindPageByCnd(cnd *sqls.Cnd) (list []models.SystemConfig, paging *sqls.Paging) {
	return repositories.SystemConfigRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *systemConfigService) Count(cnd *sqls.Cnd) int64 {
	return repositories.SystemConfigRepository.Count(sqls.DB(), cnd)
}

func (s *systemConfigService) Create(t *models.SystemConfig) error {
	return repositories.SystemConfigRepository.Create(sqls.DB(), t)
}

func (s *systemConfigService) Update(t *models.SystemConfig) error {
	return repositories.SystemConfigRepository.Update(sqls.DB(), t)
}

func (s *systemConfigService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.SystemConfigRepository.Updates(sqls.DB(), id, columns)
}

func (s *systemConfigService) UpdateColumn(id int64, name string, value interface{}) error {
	return repositories.SystemConfigRepository.UpdateColumn(sqls.DB(), id, name, value)
}

func (s *systemConfigService) Delete(id int64) {
	repositories.SystemConfigRepository.Delete(sqls.DB(), id)
}

