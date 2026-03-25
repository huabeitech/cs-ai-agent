package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/repositories"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
)

var LoginCredentialLogService = newLoginCredentialLogService()

func newLoginCredentialLogService() *loginCredentialLogService {
	return &loginCredentialLogService{}
}

type loginCredentialLogService struct {
}

func (s *loginCredentialLogService) Get(id int64) *models.LoginCredentialLog {
	return repositories.LoginCredentialLogRepository.Get(sqls.DB(), id)
}

func (s *loginCredentialLogService) Take(where ...interface{}) *models.LoginCredentialLog {
	return repositories.LoginCredentialLogRepository.Take(sqls.DB(), where...)
}

func (s *loginCredentialLogService) Find(cnd *sqls.Cnd) []models.LoginCredentialLog {
	return repositories.LoginCredentialLogRepository.Find(sqls.DB(), cnd)
}

func (s *loginCredentialLogService) FindOne(cnd *sqls.Cnd) *models.LoginCredentialLog {
	return repositories.LoginCredentialLogRepository.FindOne(sqls.DB(), cnd)
}

func (s *loginCredentialLogService) FindPageByParams(params *params.QueryParams) (list []models.LoginCredentialLog, paging *sqls.Paging) {
	return repositories.LoginCredentialLogRepository.FindPageByParams(sqls.DB(), params)
}

func (s *loginCredentialLogService) FindPageByCnd(cnd *sqls.Cnd) (list []models.LoginCredentialLog, paging *sqls.Paging) {
	return repositories.LoginCredentialLogRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *loginCredentialLogService) Count(cnd *sqls.Cnd) int64 {
	return repositories.LoginCredentialLogRepository.Count(sqls.DB(), cnd)
}

func (s *loginCredentialLogService) Create(t *models.LoginCredentialLog) error {
	return repositories.LoginCredentialLogRepository.Create(sqls.DB(), t)
}

func (s *loginCredentialLogService) Update(t *models.LoginCredentialLog) error {
	return repositories.LoginCredentialLogRepository.Update(sqls.DB(), t)
}

func (s *loginCredentialLogService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.LoginCredentialLogRepository.Updates(sqls.DB(), id, columns)
}

func (s *loginCredentialLogService) UpdateColumn(id int64, name string, value interface{}) error {
	return repositories.LoginCredentialLogRepository.UpdateColumn(sqls.DB(), id, name, value)
}

func (s *loginCredentialLogService) Delete(id int64) {
	repositories.LoginCredentialLogRepository.Delete(sqls.DB(), id)
}
