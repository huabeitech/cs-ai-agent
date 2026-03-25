package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/errorsx"
	"cs-agent/internal/repositories"
	"time"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
)

var LoginSessionService = newLoginSessionService()

func newLoginSessionService() *loginSessionService {
	return &loginSessionService{}
}

type loginSessionService struct {
}

func (s *loginSessionService) Get(id int64) *models.LoginSession {
	return repositories.LoginSessionRepository.Get(sqls.DB(), id)
}

func (s *loginSessionService) Take(where ...interface{}) *models.LoginSession {
	return repositories.LoginSessionRepository.Take(sqls.DB(), where...)
}

func (s *loginSessionService) Find(cnd *sqls.Cnd) []models.LoginSession {
	return repositories.LoginSessionRepository.Find(sqls.DB(), cnd)
}

func (s *loginSessionService) FindOne(cnd *sqls.Cnd) *models.LoginSession {
	return repositories.LoginSessionRepository.FindOne(sqls.DB(), cnd)
}

func (s *loginSessionService) FindPageByParams(params *params.QueryParams) (list []models.LoginSession, paging *sqls.Paging) {
	return repositories.LoginSessionRepository.FindPageByParams(sqls.DB(), params)
}

func (s *loginSessionService) FindPageByCnd(cnd *sqls.Cnd) (list []models.LoginSession, paging *sqls.Paging) {
	return repositories.LoginSessionRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *loginSessionService) Count(cnd *sqls.Cnd) int64 {
	return repositories.LoginSessionRepository.Count(sqls.DB(), cnd)
}

func (s *loginSessionService) Create(t *models.LoginSession) error {
	return repositories.LoginSessionRepository.Create(sqls.DB(), t)
}

func (s *loginSessionService) Update(t *models.LoginSession) error {
	return repositories.LoginSessionRepository.Update(sqls.DB(), t)
}

func (s *loginSessionService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.LoginSessionRepository.Updates(sqls.DB(), id, columns)
}

func (s *loginSessionService) UpdateColumn(id int64, name string, value interface{}) error {
	return repositories.LoginSessionRepository.UpdateColumn(sqls.DB(), id, name, value)
}

func (s *loginSessionService) Delete(id int64) {
	repositories.LoginSessionRepository.Delete(sqls.DB(), id)
}

func (s *loginSessionService) Revoke(id int64, operatorID int64, operatorName string) error {
	session := s.Get(id)
	if session == nil {
		return errorsx.InvalidParam("会话不存在")
	}
	now := time.Now()
	return s.Updates(id, map[string]any{
		"revoked_at":       now,
		"update_user_id":   operatorID,
		"update_user_name": operatorName,
		"updated_at":       now,
	})
}

func (s *loginSessionService) RevokeByUser(userID int64, operatorID int64, operatorName string) error {
	now := time.Now()
	return sqls.DB().Model(&models.LoginSession{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Updates(map[string]any{
			"revoked_at":       now,
			"update_user_id":   operatorID,
			"update_user_name": operatorName,
			"updated_at":       now,
		}).Error
}
