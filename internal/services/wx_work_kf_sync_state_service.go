package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/repositories"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
)

var WxWorkKFSyncStateService = newWxWorkKFSyncStateService()

func newWxWorkKFSyncStateService() *wxWorkKFSyncStateService {
	return &wxWorkKFSyncStateService{}
}

type wxWorkKFSyncStateService struct {
}

func (s *wxWorkKFSyncStateService) Get(id int64) *models.WxWorkKFSyncState {
	return repositories.WxWorkKFSyncStateRepository.Get(sqls.DB(), id)
}

func (s *wxWorkKFSyncStateService) Take(where ...interface{}) *models.WxWorkKFSyncState {
	return repositories.WxWorkKFSyncStateRepository.Take(sqls.DB(), where...)
}

func (s *wxWorkKFSyncStateService) Find(cnd *sqls.Cnd) []models.WxWorkKFSyncState {
	return repositories.WxWorkKFSyncStateRepository.Find(sqls.DB(), cnd)
}

func (s *wxWorkKFSyncStateService) FindOne(cnd *sqls.Cnd) *models.WxWorkKFSyncState {
	return repositories.WxWorkKFSyncStateRepository.FindOne(sqls.DB(), cnd)
}

func (s *wxWorkKFSyncStateService) FindPageByParams(params *params.QueryParams) (list []models.WxWorkKFSyncState, paging *sqls.Paging) {
	return repositories.WxWorkKFSyncStateRepository.FindPageByParams(sqls.DB(), params)
}

func (s *wxWorkKFSyncStateService) FindPageByCnd(cnd *sqls.Cnd) (list []models.WxWorkKFSyncState, paging *sqls.Paging) {
	return repositories.WxWorkKFSyncStateRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *wxWorkKFSyncStateService) Count(cnd *sqls.Cnd) int64 {
	return repositories.WxWorkKFSyncStateRepository.Count(sqls.DB(), cnd)
}

func (s *wxWorkKFSyncStateService) Create(t *models.WxWorkKFSyncState) error {
	return repositories.WxWorkKFSyncStateRepository.Create(sqls.DB(), t)
}

func (s *wxWorkKFSyncStateService) Update(t *models.WxWorkKFSyncState) error {
	return repositories.WxWorkKFSyncStateRepository.Update(sqls.DB(), t)
}

func (s *wxWorkKFSyncStateService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.WxWorkKFSyncStateRepository.Updates(sqls.DB(), id, columns)
}

func (s *wxWorkKFSyncStateService) UpdateColumn(id int64, name string, value interface{}) error {
	return repositories.WxWorkKFSyncStateRepository.UpdateColumn(sqls.DB(), id, name, value)
}

func (s *wxWorkKFSyncStateService) Delete(id int64) {
	repositories.WxWorkKFSyncStateRepository.Delete(sqls.DB(), id)
}

// GetByOpenKfID 根据open_kf_id获取同步状态
func (s *wxWorkKFSyncStateService) GetByOpenKfID(openKfID string) *models.WxWorkKFSyncState {
	return repositories.WxWorkKFSyncStateRepository.Take(sqls.DB(), "open_kf_id = ?", openKfID)
}
