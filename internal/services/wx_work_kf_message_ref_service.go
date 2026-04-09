package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/repositories"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
)

var WxWorkKFMessageRefService = newWxWorkKFMessageRefService()

func newWxWorkKFMessageRefService() *wxWorkKFMessageRefService {
	return &wxWorkKFMessageRefService{}
}

type wxWorkKFMessageRefService struct {
}

func (s *wxWorkKFMessageRefService) Get(id int64) *models.WxWorkKFMessageRef {
	return repositories.WxWorkKFMessageRefRepository.Get(sqls.DB(), id)
}

func (s *wxWorkKFMessageRefService) Take(where ...interface{}) *models.WxWorkKFMessageRef {
	return repositories.WxWorkKFMessageRefRepository.Take(sqls.DB(), where...)
}

func (s *wxWorkKFMessageRefService) Find(cnd *sqls.Cnd) []models.WxWorkKFMessageRef {
	return repositories.WxWorkKFMessageRefRepository.Find(sqls.DB(), cnd)
}

func (s *wxWorkKFMessageRefService) FindOne(cnd *sqls.Cnd) *models.WxWorkKFMessageRef {
	return repositories.WxWorkKFMessageRefRepository.FindOne(sqls.DB(), cnd)
}

func (s *wxWorkKFMessageRefService) FindPageByParams(params *params.QueryParams) (list []models.WxWorkKFMessageRef, paging *sqls.Paging) {
	return repositories.WxWorkKFMessageRefRepository.FindPageByParams(sqls.DB(), params)
}

func (s *wxWorkKFMessageRefService) FindPageByCnd(cnd *sqls.Cnd) (list []models.WxWorkKFMessageRef, paging *sqls.Paging) {
	return repositories.WxWorkKFMessageRefRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *wxWorkKFMessageRefService) Count(cnd *sqls.Cnd) int64 {
	return repositories.WxWorkKFMessageRefRepository.Count(sqls.DB(), cnd)
}

func (s *wxWorkKFMessageRefService) Create(t *models.WxWorkKFMessageRef) error {
	return repositories.WxWorkKFMessageRefRepository.Create(sqls.DB(), t)
}

func (s *wxWorkKFMessageRefService) Update(t *models.WxWorkKFMessageRef) error {
	return repositories.WxWorkKFMessageRefRepository.Update(sqls.DB(), t)
}

func (s *wxWorkKFMessageRefService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.WxWorkKFMessageRefRepository.Updates(sqls.DB(), id, columns)
}

func (s *wxWorkKFMessageRefService) UpdateColumn(id int64, name string, value interface{}) error {
	return repositories.WxWorkKFMessageRefRepository.UpdateColumn(sqls.DB(), id, name, value)
}

func (s *wxWorkKFMessageRefService) Delete(id int64) {
	repositories.WxWorkKFMessageRefRepository.Delete(sqls.DB(), id)
}

func (s *wxWorkKFMessageRefService) GetByWxMsgID(wxMsgID string) *models.WxWorkKFMessageRef {
	return repositories.WxWorkKFMessageRefRepository.Take(sqls.DB(), "wx_msg_id = ?", wxMsgID)
}
