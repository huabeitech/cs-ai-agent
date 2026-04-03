package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/repositories"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
)

var TicketCollaboratorService = newTicketCollaboratorService()

func newTicketCollaboratorService() *ticketCollaboratorService {
	return &ticketCollaboratorService{}
}

type ticketCollaboratorService struct {
}

func (s *ticketCollaboratorService) Get(id int64) *models.TicketCollaborator {
	return repositories.TicketCollaboratorRepository.Get(sqls.DB(), id)
}

func (s *ticketCollaboratorService) Take(where ...interface{}) *models.TicketCollaborator {
	return repositories.TicketCollaboratorRepository.Take(sqls.DB(), where...)
}

func (s *ticketCollaboratorService) Find(cnd *sqls.Cnd) []models.TicketCollaborator {
	return repositories.TicketCollaboratorRepository.Find(sqls.DB(), cnd)
}

func (s *ticketCollaboratorService) FindOne(cnd *sqls.Cnd) *models.TicketCollaborator {
	return repositories.TicketCollaboratorRepository.FindOne(sqls.DB(), cnd)
}

func (s *ticketCollaboratorService) FindPageByParams(params *params.QueryParams) (list []models.TicketCollaborator, paging *sqls.Paging) {
	return repositories.TicketCollaboratorRepository.FindPageByParams(sqls.DB(), params)
}

func (s *ticketCollaboratorService) FindPageByCnd(cnd *sqls.Cnd) (list []models.TicketCollaborator, paging *sqls.Paging) {
	return repositories.TicketCollaboratorRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *ticketCollaboratorService) Count(cnd *sqls.Cnd) int64 {
	return repositories.TicketCollaboratorRepository.Count(sqls.DB(), cnd)
}

func (s *ticketCollaboratorService) Create(t *models.TicketCollaborator) error {
	return repositories.TicketCollaboratorRepository.Create(sqls.DB(), t)
}

func (s *ticketCollaboratorService) Delete(id int64) {
	repositories.TicketCollaboratorRepository.Delete(sqls.DB(), id)
}
