package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/repositories"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
)

var TicketMentionService = newTicketMentionService()

func newTicketMentionService() *ticketMentionService {
	return &ticketMentionService{}
}

type ticketMentionService struct {
}

func (s *ticketMentionService) Get(id int64) *models.TicketMention {
	return repositories.TicketMentionRepository.Get(sqls.DB(), id)
}

func (s *ticketMentionService) Take(where ...interface{}) *models.TicketMention {
	return repositories.TicketMentionRepository.Take(sqls.DB(), where...)
}

func (s *ticketMentionService) Find(cnd *sqls.Cnd) []models.TicketMention {
	return repositories.TicketMentionRepository.Find(sqls.DB(), cnd)
}

func (s *ticketMentionService) FindOne(cnd *sqls.Cnd) *models.TicketMention {
	return repositories.TicketMentionRepository.FindOne(sqls.DB(), cnd)
}

func (s *ticketMentionService) FindPageByParams(params *params.QueryParams) (list []models.TicketMention, paging *sqls.Paging) {
	return repositories.TicketMentionRepository.FindPageByParams(sqls.DB(), params)
}

func (s *ticketMentionService) FindPageByCnd(cnd *sqls.Cnd) (list []models.TicketMention, paging *sqls.Paging) {
	return repositories.TicketMentionRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *ticketMentionService) Count(cnd *sqls.Cnd) int64 {
	return repositories.TicketMentionRepository.Count(sqls.DB(), cnd)
}
