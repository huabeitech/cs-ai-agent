package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/repositories"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
)

var TicketReplyService = newTicketReplyService()

func newTicketReplyService() *ticketReplyService {
	return &ticketReplyService{}
}

type ticketReplyService struct {
}

func (s *ticketReplyService) Get(id int64) *models.TicketReply {
	return repositories.TicketReplyRepository.Get(sqls.DB(), id)
}

func (s *ticketReplyService) Take(where ...interface{}) *models.TicketReply {
	return repositories.TicketReplyRepository.Take(sqls.DB(), where...)
}

func (s *ticketReplyService) Find(cnd *sqls.Cnd) []models.TicketReply {
	return repositories.TicketReplyRepository.Find(sqls.DB(), cnd)
}

func (s *ticketReplyService) FindOne(cnd *sqls.Cnd) *models.TicketReply {
	return repositories.TicketReplyRepository.FindOne(sqls.DB(), cnd)
}

func (s *ticketReplyService) FindPageByParams(params *params.QueryParams) (list []models.TicketReply, paging *sqls.Paging) {
	return repositories.TicketReplyRepository.FindPageByParams(sqls.DB(), params)
}

func (s *ticketReplyService) FindPageByCnd(cnd *sqls.Cnd) (list []models.TicketReply, paging *sqls.Paging) {
	return repositories.TicketReplyRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *ticketReplyService) Count(cnd *sqls.Cnd) int64 {
	return repositories.TicketReplyRepository.Count(sqls.DB(), cnd)
}

func (s *ticketReplyService) Create(t *models.TicketReply) error {
	return repositories.TicketReplyRepository.Create(sqls.DB(), t)
}

func (s *ticketReplyService) Update(t *models.TicketReply) error {
	return repositories.TicketReplyRepository.Update(sqls.DB(), t)
}

func (s *ticketReplyService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.TicketReplyRepository.Updates(sqls.DB(), id, columns)
}

func (s *ticketReplyService) UpdateColumn(id int64, name string, value interface{}) error {
	return repositories.TicketReplyRepository.UpdateColumn(sqls.DB(), id, name, value)
}

func (s *ticketReplyService) Delete(id int64) {
	repositories.TicketReplyRepository.Delete(sqls.DB(), id)
}
