package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/repositories"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
)

var TicketRelationService = newTicketRelationService()

func newTicketRelationService() *ticketRelationService {
	return &ticketRelationService{}
}

type ticketRelationService struct {
}

func (s *ticketRelationService) Get(id int64) *models.TicketRelation {
	return repositories.TicketRelationRepository.Get(sqls.DB(), id)
}

func (s *ticketRelationService) Take(where ...interface{}) *models.TicketRelation {
	return repositories.TicketRelationRepository.Take(sqls.DB(), where...)
}

func (s *ticketRelationService) Find(cnd *sqls.Cnd) []models.TicketRelation {
	return repositories.TicketRelationRepository.Find(sqls.DB(), cnd)
}

func (s *ticketRelationService) FindOne(cnd *sqls.Cnd) *models.TicketRelation {
	return repositories.TicketRelationRepository.FindOne(sqls.DB(), cnd)
}

func (s *ticketRelationService) FindPageByParams(params *params.QueryParams) (list []models.TicketRelation, paging *sqls.Paging) {
	return repositories.TicketRelationRepository.FindPageByParams(sqls.DB(), params)
}

func (s *ticketRelationService) FindPageByCnd(cnd *sqls.Cnd) (list []models.TicketRelation, paging *sqls.Paging) {
	return repositories.TicketRelationRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *ticketRelationService) Count(cnd *sqls.Cnd) int64 {
	return repositories.TicketRelationRepository.Count(sqls.DB(), cnd)
}

func (s *ticketRelationService) Create(t *models.TicketRelation) error {
	return repositories.TicketRelationRepository.Create(sqls.DB(), t)
}

func (s *ticketRelationService) Update(t *models.TicketRelation) error {
	return repositories.TicketRelationRepository.Update(sqls.DB(), t)
}

func (s *ticketRelationService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.TicketRelationRepository.Updates(sqls.DB(), id, columns)
}

func (s *ticketRelationService) UpdateColumn(id int64, name string, value interface{}) error {
	return repositories.TicketRelationRepository.UpdateColumn(sqls.DB(), id, name, value)
}

func (s *ticketRelationService) Delete(id int64) {
	repositories.TicketRelationRepository.Delete(sqls.DB(), id)
}

