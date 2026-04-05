package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/repositories"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
)

var ChannelMessageOutboxService = newChannelMessageOutboxService()

func newChannelMessageOutboxService() *channelMessageOutboxService {
	return &channelMessageOutboxService{}
}

type channelMessageOutboxService struct {
}

func (s *channelMessageOutboxService) Get(id int64) *models.ChannelMessageOutbox {
	return repositories.ChannelMessageOutboxRepository.Get(sqls.DB(), id)
}

func (s *channelMessageOutboxService) Take(where ...interface{}) *models.ChannelMessageOutbox {
	return repositories.ChannelMessageOutboxRepository.Take(sqls.DB(), where...)
}

func (s *channelMessageOutboxService) Find(cnd *sqls.Cnd) []models.ChannelMessageOutbox {
	return repositories.ChannelMessageOutboxRepository.Find(sqls.DB(), cnd)
}

func (s *channelMessageOutboxService) FindOne(cnd *sqls.Cnd) *models.ChannelMessageOutbox {
	return repositories.ChannelMessageOutboxRepository.FindOne(sqls.DB(), cnd)
}

func (s *channelMessageOutboxService) FindPageByParams(params *params.QueryParams) (list []models.ChannelMessageOutbox, paging *sqls.Paging) {
	return repositories.ChannelMessageOutboxRepository.FindPageByParams(sqls.DB(), params)
}

func (s *channelMessageOutboxService) FindPageByCnd(cnd *sqls.Cnd) (list []models.ChannelMessageOutbox, paging *sqls.Paging) {
	return repositories.ChannelMessageOutboxRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *channelMessageOutboxService) Count(cnd *sqls.Cnd) int64 {
	return repositories.ChannelMessageOutboxRepository.Count(sqls.DB(), cnd)
}

func (s *channelMessageOutboxService) Create(t *models.ChannelMessageOutbox) error {
	return repositories.ChannelMessageOutboxRepository.Create(sqls.DB(), t)
}

func (s *channelMessageOutboxService) Update(t *models.ChannelMessageOutbox) error {
	return repositories.ChannelMessageOutboxRepository.Update(sqls.DB(), t)
}

func (s *channelMessageOutboxService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.ChannelMessageOutboxRepository.Updates(sqls.DB(), id, columns)
}

func (s *channelMessageOutboxService) UpdateColumn(id int64, name string, value interface{}) error {
	return repositories.ChannelMessageOutboxRepository.UpdateColumn(sqls.DB(), id, name, value)
}

func (s *channelMessageOutboxService) Delete(id int64) {
	repositories.ChannelMessageOutboxRepository.Delete(sqls.DB(), id)
}

