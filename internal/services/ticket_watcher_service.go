package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/repositories"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
)

var TicketWatcherService = newTicketWatcherService()

func newTicketWatcherService() *ticketWatcherService {
	return &ticketWatcherService{}
}

type ticketWatcherService struct {
}

func (s *ticketWatcherService) Get(id int64) *models.TicketWatcher {
	return repositories.TicketWatcherRepository.Get(sqls.DB(), id)
}

func (s *ticketWatcherService) Take(where ...interface{}) *models.TicketWatcher {
	return repositories.TicketWatcherRepository.Take(sqls.DB(), where...)
}

func (s *ticketWatcherService) Find(cnd *sqls.Cnd) []models.TicketWatcher {
	return repositories.TicketWatcherRepository.Find(sqls.DB(), cnd)
}

func (s *ticketWatcherService) FindOne(cnd *sqls.Cnd) *models.TicketWatcher {
	return repositories.TicketWatcherRepository.FindOne(sqls.DB(), cnd)
}

func (s *ticketWatcherService) FindPageByParams(params *params.QueryParams) (list []models.TicketWatcher, paging *sqls.Paging) {
	return repositories.TicketWatcherRepository.FindPageByParams(sqls.DB(), params)
}

func (s *ticketWatcherService) FindPageByCnd(cnd *sqls.Cnd) (list []models.TicketWatcher, paging *sqls.Paging) {
	return repositories.TicketWatcherRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *ticketWatcherService) Count(cnd *sqls.Cnd) int64 {
	return repositories.TicketWatcherRepository.Count(sqls.DB(), cnd)
}

func (s *ticketWatcherService) Create(t *models.TicketWatcher) error {
	return repositories.TicketWatcherRepository.Create(sqls.DB(), t)
}

func (s *ticketWatcherService) Update(t *models.TicketWatcher) error {
	return repositories.TicketWatcherRepository.Update(sqls.DB(), t)
}

func (s *ticketWatcherService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.TicketWatcherRepository.Updates(sqls.DB(), id, columns)
}

func (s *ticketWatcherService) UpdateColumn(id int64, name string, value interface{}) error {
	return repositories.TicketWatcherRepository.UpdateColumn(sqls.DB(), id, name, value)
}

func (s *ticketWatcherService) Delete(id int64) {
	repositories.TicketWatcherRepository.Delete(sqls.DB(), id)
}

