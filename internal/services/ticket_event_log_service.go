package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/repositories"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
)

var TicketEventLogService = newTicketEventLogService()

func newTicketEventLogService() *ticketEventLogService {
	return &ticketEventLogService{}
}

type ticketEventLogService struct {
}

func (s *ticketEventLogService) Get(id int64) *models.TicketEventLog {
	return repositories.TicketEventLogRepository.Get(sqls.DB(), id)
}

func (s *ticketEventLogService) Take(where ...interface{}) *models.TicketEventLog {
	return repositories.TicketEventLogRepository.Take(sqls.DB(), where...)
}

func (s *ticketEventLogService) Find(cnd *sqls.Cnd) []models.TicketEventLog {
	return repositories.TicketEventLogRepository.Find(sqls.DB(), cnd)
}

func (s *ticketEventLogService) FindOne(cnd *sqls.Cnd) *models.TicketEventLog {
	return repositories.TicketEventLogRepository.FindOne(sqls.DB(), cnd)
}

func (s *ticketEventLogService) FindPageByParams(params *params.QueryParams) (list []models.TicketEventLog, paging *sqls.Paging) {
	return repositories.TicketEventLogRepository.FindPageByParams(sqls.DB(), params)
}

func (s *ticketEventLogService) FindPageByCnd(cnd *sqls.Cnd) (list []models.TicketEventLog, paging *sqls.Paging) {
	return repositories.TicketEventLogRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *ticketEventLogService) Count(cnd *sqls.Cnd) int64 {
	return repositories.TicketEventLogRepository.Count(sqls.DB(), cnd)
}

func (s *ticketEventLogService) Create(t *models.TicketEventLog) error {
	return repositories.TicketEventLogRepository.Create(sqls.DB(), t)
}

func (s *ticketEventLogService) Update(t *models.TicketEventLog) error {
	return repositories.TicketEventLogRepository.Update(sqls.DB(), t)
}

func (s *ticketEventLogService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.TicketEventLogRepository.Updates(sqls.DB(), id, columns)
}

func (s *ticketEventLogService) UpdateColumn(id int64, name string, value interface{}) error {
	return repositories.TicketEventLogRepository.UpdateColumn(sqls.DB(), id, name, value)
}

func (s *ticketEventLogService) Delete(id int64) {
	repositories.TicketEventLogRepository.Delete(sqls.DB(), id)
}
