package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/repositories"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
)

var TicketAssignmentService = newTicketAssignmentService()

func newTicketAssignmentService() *ticketAssignmentService {
	return &ticketAssignmentService{}
}

type ticketAssignmentService struct {
}

func (s *ticketAssignmentService) Get(id int64) *models.TicketAssignment {
	return repositories.TicketAssignmentRepository.Get(sqls.DB(), id)
}

func (s *ticketAssignmentService) Take(where ...interface{}) *models.TicketAssignment {
	return repositories.TicketAssignmentRepository.Take(sqls.DB(), where...)
}

func (s *ticketAssignmentService) Find(cnd *sqls.Cnd) []models.TicketAssignment {
	return repositories.TicketAssignmentRepository.Find(sqls.DB(), cnd)
}

func (s *ticketAssignmentService) FindOne(cnd *sqls.Cnd) *models.TicketAssignment {
	return repositories.TicketAssignmentRepository.FindOne(sqls.DB(), cnd)
}

func (s *ticketAssignmentService) FindPageByParams(params *params.QueryParams) (list []models.TicketAssignment, paging *sqls.Paging) {
	return repositories.TicketAssignmentRepository.FindPageByParams(sqls.DB(), params)
}

func (s *ticketAssignmentService) FindPageByCnd(cnd *sqls.Cnd) (list []models.TicketAssignment, paging *sqls.Paging) {
	return repositories.TicketAssignmentRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *ticketAssignmentService) Count(cnd *sqls.Cnd) int64 {
	return repositories.TicketAssignmentRepository.Count(sqls.DB(), cnd)
}

func (s *ticketAssignmentService) Create(t *models.TicketAssignment) error {
	return repositories.TicketAssignmentRepository.Create(sqls.DB(), t)
}

func (s *ticketAssignmentService) Update(t *models.TicketAssignment) error {
	return repositories.TicketAssignmentRepository.Update(sqls.DB(), t)
}

func (s *ticketAssignmentService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.TicketAssignmentRepository.Updates(sqls.DB(), id, columns)
}

func (s *ticketAssignmentService) UpdateColumn(id int64, name string, value interface{}) error {
	return repositories.TicketAssignmentRepository.UpdateColumn(sqls.DB(), id, name, value)
}

func (s *ticketAssignmentService) Delete(id int64) {
	repositories.TicketAssignmentRepository.Delete(sqls.DB(), id)
}
