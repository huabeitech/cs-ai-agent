package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/repositories"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
)

var WxWorkKFConversationService = newWxWorkKFConversationService()

func newWxWorkKFConversationService() *wxWorkKFConversationService {
	return &wxWorkKFConversationService{}
}

type wxWorkKFConversationService struct {
}

func (s *wxWorkKFConversationService) Get(id int64) *models.WxWorkKFConversation {
	return repositories.WxWorkKFConversationRepository.Get(sqls.DB(), id)
}

func (s *wxWorkKFConversationService) Take(where ...interface{}) *models.WxWorkKFConversation {
	return repositories.WxWorkKFConversationRepository.Take(sqls.DB(), where...)
}

func (s *wxWorkKFConversationService) Find(cnd *sqls.Cnd) []models.WxWorkKFConversation {
	return repositories.WxWorkKFConversationRepository.Find(sqls.DB(), cnd)
}

func (s *wxWorkKFConversationService) FindOne(cnd *sqls.Cnd) *models.WxWorkKFConversation {
	return repositories.WxWorkKFConversationRepository.FindOne(sqls.DB(), cnd)
}

func (s *wxWorkKFConversationService) FindPageByParams(params *params.QueryParams) (list []models.WxWorkKFConversation, paging *sqls.Paging) {
	return repositories.WxWorkKFConversationRepository.FindPageByParams(sqls.DB(), params)
}

func (s *wxWorkKFConversationService) FindPageByCnd(cnd *sqls.Cnd) (list []models.WxWorkKFConversation, paging *sqls.Paging) {
	return repositories.WxWorkKFConversationRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *wxWorkKFConversationService) Count(cnd *sqls.Cnd) int64 {
	return repositories.WxWorkKFConversationRepository.Count(sqls.DB(), cnd)
}

func (s *wxWorkKFConversationService) Create(t *models.WxWorkKFConversation) error {
	return repositories.WxWorkKFConversationRepository.Create(sqls.DB(), t)
}

func (s *wxWorkKFConversationService) Update(t *models.WxWorkKFConversation) error {
	return repositories.WxWorkKFConversationRepository.Update(sqls.DB(), t)
}

func (s *wxWorkKFConversationService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.WxWorkKFConversationRepository.Updates(sqls.DB(), id, columns)
}

func (s *wxWorkKFConversationService) UpdateColumn(id int64, name string, value interface{}) error {
	return repositories.WxWorkKFConversationRepository.UpdateColumn(sqls.DB(), id, name, value)
}

func (s *wxWorkKFConversationService) Delete(id int64) {
	repositories.WxWorkKFConversationRepository.Delete(sqls.DB(), id)
}

