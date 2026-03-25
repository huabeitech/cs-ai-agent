package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/repositories"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
)

var TicketAttachmentService = newTicketAttachmentService()

func newTicketAttachmentService() *ticketAttachmentService {
	return &ticketAttachmentService{}
}

type ticketAttachmentService struct {
}

func (s *ticketAttachmentService) Get(id int64) *models.TicketAttachment {
	return repositories.TicketAttachmentRepository.Get(sqls.DB(), id)
}

func (s *ticketAttachmentService) Take(where ...interface{}) *models.TicketAttachment {
	return repositories.TicketAttachmentRepository.Take(sqls.DB(), where...)
}

func (s *ticketAttachmentService) Find(cnd *sqls.Cnd) []models.TicketAttachment {
	return repositories.TicketAttachmentRepository.Find(sqls.DB(), cnd)
}

func (s *ticketAttachmentService) FindOne(cnd *sqls.Cnd) *models.TicketAttachment {
	return repositories.TicketAttachmentRepository.FindOne(sqls.DB(), cnd)
}

func (s *ticketAttachmentService) FindPageByParams(params *params.QueryParams) (list []models.TicketAttachment, paging *sqls.Paging) {
	return repositories.TicketAttachmentRepository.FindPageByParams(sqls.DB(), params)
}

func (s *ticketAttachmentService) FindPageByCnd(cnd *sqls.Cnd) (list []models.TicketAttachment, paging *sqls.Paging) {
	return repositories.TicketAttachmentRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *ticketAttachmentService) Count(cnd *sqls.Cnd) int64 {
	return repositories.TicketAttachmentRepository.Count(sqls.DB(), cnd)
}

func (s *ticketAttachmentService) Create(t *models.TicketAttachment) error {
	return repositories.TicketAttachmentRepository.Create(sqls.DB(), t)
}

func (s *ticketAttachmentService) Update(t *models.TicketAttachment) error {
	return repositories.TicketAttachmentRepository.Update(sqls.DB(), t)
}

func (s *ticketAttachmentService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.TicketAttachmentRepository.Updates(sqls.DB(), id, columns)
}

func (s *ticketAttachmentService) UpdateColumn(id int64, name string, value interface{}) error {
	return repositories.TicketAttachmentRepository.UpdateColumn(sqls.DB(), id, name, value)
}

func (s *ticketAttachmentService) Delete(id int64) {
	repositories.TicketAttachmentRepository.Delete(sqls.DB(), id)
}
