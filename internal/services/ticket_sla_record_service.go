package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/repositories"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
)

var TicketSLARecordService = newTicketSLARecordService()

func newTicketSLARecordService() *ticketSLARecordService {
	return &ticketSLARecordService{}
}

type ticketSLARecordService struct {
}

func (s *ticketSLARecordService) Get(id int64) *models.TicketSLARecord {
	return repositories.TicketSLARecordRepository.Get(sqls.DB(), id)
}

func (s *ticketSLARecordService) Take(where ...interface{}) *models.TicketSLARecord {
	return repositories.TicketSLARecordRepository.Take(sqls.DB(), where...)
}

func (s *ticketSLARecordService) Find(cnd *sqls.Cnd) []models.TicketSLARecord {
	return repositories.TicketSLARecordRepository.Find(sqls.DB(), cnd)
}

func (s *ticketSLARecordService) FindOne(cnd *sqls.Cnd) *models.TicketSLARecord {
	return repositories.TicketSLARecordRepository.FindOne(sqls.DB(), cnd)
}

func (s *ticketSLARecordService) FindPageByParams(params *params.QueryParams) (list []models.TicketSLARecord, paging *sqls.Paging) {
	return repositories.TicketSLARecordRepository.FindPageByParams(sqls.DB(), params)
}

func (s *ticketSLARecordService) FindPageByCnd(cnd *sqls.Cnd) (list []models.TicketSLARecord, paging *sqls.Paging) {
	return repositories.TicketSLARecordRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *ticketSLARecordService) Count(cnd *sqls.Cnd) int64 {
	return repositories.TicketSLARecordRepository.Count(sqls.DB(), cnd)
}

func (s *ticketSLARecordService) Create(t *models.TicketSLARecord) error {
	return repositories.TicketSLARecordRepository.Create(sqls.DB(), t)
}

func (s *ticketSLARecordService) Update(t *models.TicketSLARecord) error {
	return repositories.TicketSLARecordRepository.Update(sqls.DB(), t)
}

func (s *ticketSLARecordService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.TicketSLARecordRepository.Updates(sqls.DB(), id, columns)
}

func (s *ticketSLARecordService) UpdateColumn(id int64, name string, value interface{}) error {
	return repositories.TicketSLARecordRepository.UpdateColumn(sqls.DB(), id, name, value)
}

func (s *ticketSLARecordService) Delete(id int64) {
	repositories.TicketSLARecordRepository.Delete(sqls.DB(), id)
}

