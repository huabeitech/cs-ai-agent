package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/repositories"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
)

var TicketCommentService = newTicketCommentService()

func newTicketCommentService() *ticketCommentService {
	return &ticketCommentService{}
}

type ticketCommentService struct {
}

func (s *ticketCommentService) Get(id int64) *models.TicketComment {
	return repositories.TicketCommentRepository.Get(sqls.DB(), id)
}

func (s *ticketCommentService) Take(where ...interface{}) *models.TicketComment {
	return repositories.TicketCommentRepository.Take(sqls.DB(), where...)
}

func (s *ticketCommentService) Find(cnd *sqls.Cnd) []models.TicketComment {
	return repositories.TicketCommentRepository.Find(sqls.DB(), cnd)
}

func (s *ticketCommentService) FindOne(cnd *sqls.Cnd) *models.TicketComment {
	return repositories.TicketCommentRepository.FindOne(sqls.DB(), cnd)
}

func (s *ticketCommentService) FindPageByParams(params *params.QueryParams) (list []models.TicketComment, paging *sqls.Paging) {
	return repositories.TicketCommentRepository.FindPageByParams(sqls.DB(), params)
}

func (s *ticketCommentService) FindPageByCnd(cnd *sqls.Cnd) (list []models.TicketComment, paging *sqls.Paging) {
	return repositories.TicketCommentRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *ticketCommentService) Count(cnd *sqls.Cnd) int64 {
	return repositories.TicketCommentRepository.Count(sqls.DB(), cnd)
}

func (s *ticketCommentService) Create(t *models.TicketComment) error {
	return repositories.TicketCommentRepository.Create(sqls.DB(), t)
}

func (s *ticketCommentService) Update(t *models.TicketComment) error {
	return repositories.TicketCommentRepository.Update(sqls.DB(), t)
}

func (s *ticketCommentService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.TicketCommentRepository.Updates(sqls.DB(), id, columns)
}

func (s *ticketCommentService) UpdateColumn(id int64, name string, value interface{}) error {
	return repositories.TicketCommentRepository.UpdateColumn(sqls.DB(), id, name, value)
}

func (s *ticketCommentService) Delete(id int64) {
	repositories.TicketCommentRepository.Delete(sqls.DB(), id)
}
