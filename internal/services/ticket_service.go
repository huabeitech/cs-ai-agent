package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/utils"
	"cs-agent/internal/repositories"
	"errors"
	"fmt"
	"time"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
)

var TicketService = newTicketService()

func newTicketService() *ticketService {
	return &ticketService{}
}

type ticketService struct {
}

func (s *ticketService) Get(id int64) *models.Ticket {
	return repositories.TicketRepository.Get(sqls.DB(), id)
}

func (s *ticketService) Take(where ...interface{}) *models.Ticket {
	return repositories.TicketRepository.Take(sqls.DB(), where...)
}

func (s *ticketService) Find(cnd *sqls.Cnd) []models.Ticket {
	return repositories.TicketRepository.Find(sqls.DB(), cnd)
}

func (s *ticketService) FindOne(cnd *sqls.Cnd) *models.Ticket {
	return repositories.TicketRepository.FindOne(sqls.DB(), cnd)
}

func (s *ticketService) FindPageByParams(params *params.QueryParams) (list []models.Ticket, paging *sqls.Paging) {
	return repositories.TicketRepository.FindPageByParams(sqls.DB(), params)
}

func (s *ticketService) FindPageByCnd(cnd *sqls.Cnd) (list []models.Ticket, paging *sqls.Paging) {
	return repositories.TicketRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *ticketService) Count(cnd *sqls.Cnd) int64 {
	return repositories.TicketRepository.Count(sqls.DB(), cnd)
}

func (s *ticketService) Create(t *models.Ticket) error {
	return repositories.TicketRepository.Create(sqls.DB(), t)
}

func (s *ticketService) Update(t *models.Ticket) error {
	return repositories.TicketRepository.Update(sqls.DB(), t)
}

func (s *ticketService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.TicketRepository.Updates(sqls.DB(), id, columns)
}

func (s *ticketService) UpdateColumn(id int64, name string, value interface{}) error {
	return repositories.TicketRepository.UpdateColumn(sqls.DB(), id, name, value)
}

func (s *ticketService) Delete(id int64) {
	repositories.TicketRepository.Delete(sqls.DB(), id)
}

func (s *ticketService) CreateTicket(req request.CreateTicketRequest, principal *dto.AuthPrincipal) (*models.Ticket, error) {
	ticket := &models.Ticket{
		Title:              req.Title,
		Content:            req.Content,
		ChannelType:        req.ChannelType,
		ChannelID:          req.ChannelID,
		CategoryID:         req.CategoryID,
		Priority:           req.Priority,
		SourceUserID:       req.SourceUserID,
		ExternalUserID:     req.ExternalUserID,
		ExternalUserName:   req.ExternalUserName,
		ExternalUserEmail:  req.ExternalUserEmail,
		ExternalUserMobile: req.ExternalUserMobile,
		ConversationID:     req.ConversationID,
		Tags:               req.Tags,
		Remark:             req.Remark,
		Status:             1,
		AuditFields:        utils.BuildAuditFields(principal),
	}
	err := s.Create(ticket)
	if err != nil {
		return nil, err
	}
	ticketNo := s.generateTicketNo(ticket.ID)
	ticket.TicketNo = ticketNo
	err = s.Update(ticket)
	return ticket, err
}

func (s *ticketService) generateTicketNo(id int64) string {
	return time.Now().Format("TK") + fmt.Sprintf("%06d", id)
}

func (s *ticketService) UpdateTicket(req request.UpdateTicketRequest, principal *dto.AuthPrincipal) error {
	updates := map[string]interface{}{
		"title":       req.Title,
		"content":     req.Content,
		"category_id": req.CategoryID,
		"priority":    req.Priority,
		"status":      req.Status,
		"tags":        req.Tags,
		"remark":      req.Remark,
	}
	if principal != nil {
		updates["update_user_id"] = principal.UserID
		updates["update_user_name"] = principal.Nickname
	}
	return s.Updates(req.ID, updates)
}

func (s *ticketService) DeleteTicket(id int64) error {
	ticket := s.Get(id)
	if ticket == nil {
		return nil
	}
	s.Delete(id)
	return nil
}

func (s *ticketService) AssignTicket(req request.AssignTicketRequest, principal *dto.AuthPrincipal) error {
	ticket := s.Get(req.ID)
	if ticket == nil {
		return errors.New("工单不存在")
	}
	updates := map[string]interface{}{
		"current_assignee_id": req.ToUserID,
		"current_team_id":     req.ToTeamID,
	}
	if principal != nil {
		updates["update_user_id"] = principal.UserID
		updates["update_user_name"] = principal.Nickname
	}
	err := s.Updates(req.ID, updates)
	if err != nil {
		return err
	}
	assignment := &models.TicketAssignment{
		TicketID:    req.ID,
		FromUserID:  principal.UserID,
		ToUserID:    req.ToUserID,
		ToTeamID:    req.ToTeamID,
		AssignType:  string(enums.TicketAssignTypeDistribute),
		Reason:      req.Reason,
		Status:      int(enums.TicketAssignStatusProcessing),
		AuditFields: utils.BuildAuditFields(principal),
	}
	return repositories.TicketAssignmentRepository.Create(sqls.DB(), assignment)
}

func (s *ticketService) CloseTicket(id int64, principal *dto.AuthPrincipal) error {
	ticket := s.Get(id)
	if ticket == nil {
		return errors.New("工单不存在")
	}
	if ticket.Status == int(enums.TicketStatusClosed) {
		return errors.New("工单已关闭")
	}
	updates := map[string]interface{}{
		"status":    int(enums.TicketStatusClosed),
		"closed_at": time.Now(),
	}
	if principal != nil {
		updates["update_user_id"] = principal.UserID
		updates["update_user_name"] = principal.Nickname
	}
	return s.Updates(id, updates)
}

func (s *ticketService) ReopenTicket(id int64, principal *dto.AuthPrincipal) error {
	ticket := s.Get(id)
	if ticket == nil {
		return errors.New("工单不存在")
	}
	if ticket.Status != int(enums.TicketStatusClosed) {
		return errors.New("只能重开已关闭的工单")
	}
	updates := map[string]interface{}{
		"status":    int(enums.TicketStatusPending),
		"closed_at": nil,
	}
	if principal != nil {
		updates["update_user_id"] = principal.UserID
		updates["update_user_name"] = principal.Nickname
	}
	return s.Updates(id, updates)
}
