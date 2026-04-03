package services

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/errorsx"
	"cs-agent/internal/pkg/utils"
	"cs-agent/internal/repositories"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var TicketService = newTicketService()

func newTicketService() *ticketService {
	return &ticketService{}
}

type TicketDetailAggregate struct {
	Ticket   *models.Ticket
	Customer *models.Customer
	SLAs     []models.TicketSLARecord
	Watchers []models.TicketWatcher
	Comments []models.TicketComment
	Events   []models.TicketEventLog
}

type TicketSummaryAggregate struct {
	All             int64
	Mine            int64
	Watching        int64
	PendingCustomer int64
	DueSoon         int64
	Overdue         int64
}

type ticketService struct {
}

var ticketNoSequence uint64

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

func (s *ticketService) GetDetail(id int64) (*TicketDetailAggregate, error) {
	ticket := s.Get(id)
	if ticket == nil {
		return nil, errorsx.InvalidParam("工单不存在")
	}
	aggregate := &TicketDetailAggregate{
		Ticket: ticket,
		SLAs: TicketSLARecordService.Find(
			sqls.NewCnd().Eq("ticket_id", id).Asc("id"),
		),
		Watchers: TicketWatcherService.Find(
			sqls.NewCnd().Eq("ticket_id", id).Asc("id"),
		),
		Comments: TicketCommentService.Find(
			sqls.NewCnd().Eq("ticket_id", id).Asc("id"),
		),
		Events: TicketEventLogService.Find(
			sqls.NewCnd().Eq("ticket_id", id).Desc("id"),
		),
	}
	if ticket.CustomerID > 0 {
		aggregate.Customer = CustomerService.Get(ticket.CustomerID)
	}
	return aggregate, nil
}

func (s *ticketService) GetSummary(operator *dto.AuthPrincipal) *TicketSummaryAggregate {
	if operator == nil {
		return &TicketSummaryAggregate{}
	}
	now := time.Now()
	dueSoonAt := now.Add(30 * time.Minute)
	return &TicketSummaryAggregate{
		All: s.Count(sqls.NewCnd()),
		Mine: s.Count(
			sqls.NewCnd().Eq("current_assignee_id", operator.UserID),
		),
		Watching: s.Count(
			sqls.NewCnd().Where("id IN (SELECT ticket_id FROM ticket_watchers WHERE user_id = ?)", operator.UserID),
		),
		PendingCustomer: s.Count(
			sqls.NewCnd().Eq("status", enums.TicketStatusPendingCustomer),
		),
		DueSoon: s.Count(
			sqls.NewCnd().
				In("status", []enums.TicketStatus{
					enums.TicketStatusNew,
					enums.TicketStatusOpen,
					enums.TicketStatusPendingCustomer,
					enums.TicketStatusPendingInternal,
				}).
				Where("resolve_deadline_at IS NOT NULL").
				Where("resolve_deadline_at >= ?", now).
				Where("resolve_deadline_at <= ?", dueSoonAt),
		),
		Overdue: s.Count(
			sqls.NewCnd().
				In("status", []enums.TicketStatus{
					enums.TicketStatusNew,
					enums.TicketStatusOpen,
					enums.TicketStatusPendingCustomer,
					enums.TicketStatusPendingInternal,
				}).
				Where("resolve_deadline_at IS NOT NULL").
				Where("resolve_deadline_at < ?", now),
		),
	}
}

func (s *ticketService) CreateTicket(req request.CreateTicketRequest, operator *dto.AuthPrincipal) (*models.Ticket, error) {
	if operator == nil {
		return nil, errorsx.Unauthorized("未登录或登录已过期")
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		return nil, errorsx.InvalidParam("工单标题不能为空")
	}

	if req.CustomerID > 0 && CustomerService.Get(req.CustomerID) == nil {
		return nil, errorsx.InvalidParam("客户不存在")
	}
	if req.ConversationID > 0 && ConversationService.Get(req.ConversationID) == nil {
		return nil, errorsx.InvalidParam("会话不存在")
	}

	source := enums.TicketSource(strings.TrimSpace(req.Source))
	if source == "" {
		source = enums.TicketSourceManual
	}
	if !enums.IsValidTicketSource(string(source)) {
		return nil, errorsx.InvalidParam("工单来源不合法")
	}

	priority := enums.TicketPriority(req.Priority)
	if req.Priority == 0 {
		priority = enums.TicketPriorityNormal
	}
	if !enums.IsValidTicketPriority(int(priority)) {
		return nil, errorsx.InvalidParam("工单优先级不合法")
	}

	severity := enums.TicketSeverity(req.Severity)
	if req.Severity == 0 {
		severity = enums.TicketSeverityMinor
	}
	if !enums.IsValidTicketSeverity(int(severity)) {
		return nil, errorsx.InvalidParam("工单严重度不合法")
	}

	teamID, assigneeID, err := s.normalizeAssignment(req.CurrentTeamID, req.CurrentAssigneeID)
	if err != nil {
		return nil, err
	}

	dueAt, err := parseOptionalDateTime(req.DueAt)
	if err != nil {
		return nil, errorsx.InvalidParam("截止时间格式不合法")
	}
	customFieldsJSON, err := marshalJSON(req.CustomFields)
	if err != nil {
		return nil, errorsx.InvalidParam("自定义字段格式不合法")
	}

	now := time.Now()
	status := enums.TicketStatusNew
	if assigneeID > 0 {
		status = enums.TicketStatusOpen
	}
	firstResponseTarget, resolutionTarget := ticketSLATargetMinutes(priority)
	ticket := &models.Ticket{
		TicketNo:            s.nextTicketNo(),
		Title:               title,
		Description:         strings.TrimSpace(req.Description),
		Source:              source,
		Channel:             strings.TrimSpace(req.Channel),
		CustomerID:          req.CustomerID,
		ConversationID:      req.ConversationID,
		CategoryID:          req.CategoryID,
		Type:                strings.TrimSpace(req.Type),
		Priority:            priority,
		Severity:            severity,
		Status:              status,
		CurrentTeamID:       teamID,
		CurrentAssigneeID:   assigneeID,
		DueAt:               dueAt,
		NextReplyDeadlineAt: buildDeadlineFromNow(now, firstResponseTarget),
		ResolveDeadlineAt:   buildDeadlineFromNow(now, resolutionTarget),
		CustomFieldsJSON:    customFieldsJSON,
		AuditFields:         utils.BuildAuditFields(operator),
	}
	ticket.UpdatedAt = now

	if err := sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		if err := repositories.TicketRepository.Create(ctx.Tx, ticket); err != nil {
			return err
		}
		if err := s.initSLAs(ctx.Tx, ticket, now); err != nil {
			return err
		}
		if err := s.logEvent(ctx.Tx, ticket.ID, enums.TicketEventTypeCreated, operator, "", string(ticket.Status), "创建工单", ""); err != nil {
			return err
		}
		if ticket.ConversationID > 0 {
			if err := ConversationEventLogService.CreateEvent(ctx, ticket.ConversationID, enums.IMEventTypeMessageSend, enums.IMSenderTypeAgent, operator.UserID, fmt.Sprintf("已创建工单 %s", ticket.TicketNo), ""); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return s.Get(ticket.ID), nil
}

func (s *ticketService) CreateFromConversation(req request.CreateTicketFromConversationRequest, operator *dto.AuthPrincipal) (*models.Ticket, error) {
	if operator == nil {
		return nil, errorsx.Unauthorized("未登录或登录已过期")
	}
	conversation := ConversationService.Get(req.ConversationID)
	if conversation == nil {
		return nil, errorsx.InvalidParam("会话不存在")
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		title = strings.TrimSpace(conversation.Subject)
	}
	description := strings.TrimSpace(req.Description)
	if description == "" {
		description = strings.TrimSpace(conversation.LastMessageSummary)
	}
	item, err := s.CreateTicket(request.CreateTicketRequest{
		Title:             title,
		Description:       description,
		Source:            string(enums.TicketSourceConversation),
		Channel:           string(conversation.ExternalSource),
		CustomerID:        conversation.CustomerID,
		ConversationID:    conversation.ID,
		CategoryID:        req.CategoryID,
		Priority:          req.Priority,
		Severity:          req.Severity,
		CurrentTeamID:     req.CurrentTeamID,
		CurrentAssigneeID: req.CurrentAssigneeID,
		CustomFields:      req.CustomFields,
	}, operator)
	if err != nil {
		return nil, err
	}
	if req.SyncToConversation {
		_ = sqls.WithTransaction(func(ctx *sqls.TxContext) error {
			return ConversationEventLogService.CreateEvent(ctx, conversation.ID, enums.IMEventTypeMessageSend, enums.IMSenderTypeAgent, operator.UserID, fmt.Sprintf("会话已转为工单 %s", item.TicketNo), "")
		})
	}
	return item, nil
}

func (s *ticketService) UpdateTicket(req request.UpdateTicketRequest, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	ticket := s.Get(req.TicketID)
	if ticket == nil {
		return errorsx.InvalidParam("工单不存在")
	}
	if !s.isEditableStatus(ticket.Status) {
		return errorsx.InvalidParam("工单当前状态不允许编辑")
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return errorsx.InvalidParam("工单标题不能为空")
	}
	if !enums.IsValidTicketPriority(req.Priority) {
		return errorsx.InvalidParam("工单优先级不合法")
	}
	if !enums.IsValidTicketSeverity(req.Severity) {
		return errorsx.InvalidParam("工单严重度不合法")
	}
	teamID, assigneeID, err := s.normalizeAssignment(req.CurrentTeamID, req.CurrentAssigneeID)
	if err != nil {
		return err
	}
	dueAt, err := parseOptionalDateTime(req.DueAt)
	if err != nil {
		return errorsx.InvalidParam("截止时间格式不合法")
	}
	customFieldsJSON, err := marshalJSON(req.CustomFields)
	if err != nil {
		return errorsx.InvalidParam("自定义字段格式不合法")
	}
	now := time.Now()
	return sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		if err := repositories.TicketRepository.Updates(ctx.Tx, ticket.ID, map[string]any{
			"title":               title,
			"description":         strings.TrimSpace(req.Description),
			"category_id":         req.CategoryID,
			"type":                strings.TrimSpace(req.Type),
			"priority":            req.Priority,
			"severity":            req.Severity,
			"current_team_id":     teamID,
			"current_assignee_id": assigneeID,
			"due_at":              dueAt,
			"custom_fields_json":  customFieldsJSON,
			"update_user_id":      operator.UserID,
			"update_user_name":    operator.Username,
			"updated_at":          now,
		}); err != nil {
			return err
		}
		if err := s.syncSLATargetsAndDeadlines(ctx.Tx, ticket.ID, enums.TicketPriority(req.Priority), now); err != nil {
			return err
		}
		return s.logEvent(ctx.Tx, ticket.ID, enums.TicketEventTypeUpdated, operator, "", "", "更新工单信息", "")
	})
}

func (s *ticketService) AssignTicket(req request.AssignTicketRequest, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	ticket := s.Get(req.TicketID)
	if ticket == nil {
		return errorsx.InvalidParam("工单不存在")
	}
	teamID, assigneeID, err := s.normalizeAssignment(req.ToTeamID, req.ToUserID)
	if err != nil {
		return err
	}
	if assigneeID <= 0 {
		return errorsx.InvalidParam("目标处理人不能为空")
	}
	now := time.Now()
	eventType := enums.TicketEventTypeAssigned
	content := "指派工单"
	if ticket.CurrentAssigneeID > 0 && ticket.CurrentAssigneeID != assigneeID {
		eventType = enums.TicketEventTypeTransferred
		content = "转派工单"
	}
	status := ticket.Status
	if status == enums.TicketStatusNew {
		status = enums.TicketStatusOpen
	}
	return sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		if err := repositories.TicketRepository.Updates(ctx.Tx, ticket.ID, map[string]any{
			"current_team_id":     teamID,
			"current_assignee_id": assigneeID,
			"status":              status,
			"update_user_id":      operator.UserID,
			"update_user_name":    operator.Username,
			"updated_at":          now,
		}); err != nil {
			return err
		}
		return s.logEvent(ctx.Tx, ticket.ID, eventType, operator, fmt.Sprintf("%d", ticket.CurrentAssigneeID), fmt.Sprintf("%d", assigneeID), strings.TrimSpace(content), strings.TrimSpace(req.Reason))
	})
}

func (s *ticketService) ChangeStatus(req request.ChangeTicketStatusRequest, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	ticket := s.Get(req.TicketID)
	if ticket == nil {
		return errorsx.InvalidParam("工单不存在")
	}
	targetStatus := enums.TicketStatus(strings.TrimSpace(req.Status))
	if !enums.IsValidTicketStatus(string(targetStatus)) {
		return errorsx.InvalidParam("工单状态不合法")
	}
	if !s.canTransition(ticket.Status, targetStatus) {
		return errorsx.InvalidParam("工单当前状态不允许该操作")
	}
	now := time.Now()
	values := map[string]any{
		"status":             targetStatus,
		"pending_reason":     "",
		"close_reason":       "",
		"resolution_code":    "",
		"resolution_summary": "",
		"update_user_id":     operator.UserID,
		"update_user_name":   operator.Username,
		"updated_at":         now,
	}
	switch targetStatus {
	case enums.TicketStatusPendingCustomer, enums.TicketStatusPendingInternal:
		values["pending_reason"] = strings.TrimSpace(req.PendingReason)
	case enums.TicketStatusResolved:
		values["resolution_code"] = strings.TrimSpace(req.ResolutionCode)
		values["resolution_summary"] = strings.TrimSpace(req.ResolutionSummary)
		values["resolved_at"] = now
	case enums.TicketStatusClosed:
		closeReason := strings.TrimSpace(req.CloseReason)
		if closeReason == "" {
			return errorsx.InvalidParam("关闭原因不能为空")
		}
		values["close_reason"] = closeReason
		values["closed_at"] = now
	case enums.TicketStatusOpen:
		values["resolved_at"] = nil
		values["closed_at"] = nil
	}
	return sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		if err := repositories.TicketRepository.Updates(ctx.Tx, ticket.ID, values); err != nil {
			return err
		}
		if err := s.applySLAOnStatusChange(ctx.Tx, ticket.ID, ticket.Status, targetStatus, now); err != nil {
			return err
		}
		return s.logEvent(ctx.Tx, ticket.ID, enums.TicketEventTypeStatusChanged, operator, string(ticket.Status), string(targetStatus), strings.TrimSpace(req.Reason), "")
	})
}

func (s *ticketService) ReplyTicket(req request.ReplyTicketRequest, operator *dto.AuthPrincipal) (*models.TicketComment, error) {
	if operator == nil {
		return nil, errorsx.Unauthorized("未登录或登录已过期")
	}
	ticket := s.Get(req.TicketID)
	if ticket == nil {
		return nil, errorsx.InvalidParam("工单不存在")
	}
	if ticket.Status == enums.TicketStatusClosed || ticket.Status == enums.TicketStatusCancelled {
		return nil, errorsx.InvalidParam("当前工单状态不允许回复")
	}
	content := strings.TrimSpace(req.Content)
	if content == "" {
		return nil, errorsx.InvalidParam("回复内容不能为空")
	}
	now := time.Now()
	comment := &models.TicketComment{
		TicketID:    ticket.ID,
		CommentType: enums.TicketCommentTypePublicReply,
		AuthorType:  enums.IMSenderTypeAgent,
		AuthorID:    operator.UserID,
		ContentType: strings.TrimSpace(req.ContentType),
		Content:     content,
		Payload:     strings.TrimSpace(req.Payload),
		CreatedAt:   now,
	}
	if comment.ContentType == "" {
		comment.ContentType = "text"
	}
	err := sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		if err := repositories.TicketCommentRepository.Create(ctx.Tx, comment); err != nil {
			return err
		}
		updateValues := map[string]any{
			"update_user_id":   operator.UserID,
			"update_user_name": operator.Username,
			"updated_at":       now,
		}
		if ticket.FirstResponseAt == nil {
			updateValues["first_response_at"] = now
		}
		if ticket.Status == enums.TicketStatusPendingCustomer || ticket.Status == enums.TicketStatusNew {
			updateValues["status"] = enums.TicketStatusOpen
			updateValues["pending_reason"] = ""
		}
		if err := repositories.TicketRepository.Updates(ctx.Tx, ticket.ID, updateValues); err != nil {
			return err
		}
		if err := s.completeFirstResponseSLA(ctx.Tx, ticket.ID, now); err != nil {
			return err
		}
		return s.logEvent(ctx.Tx, ticket.ID, enums.TicketEventTypeReplied, operator, "", "", "回复客户", "")
	})
	if err != nil {
		return nil, err
	}
	return comment, nil
}

func (s *ticketService) AddInternalNote(req request.InternalNoteRequest, operator *dto.AuthPrincipal) (*models.TicketComment, error) {
	if operator == nil {
		return nil, errorsx.Unauthorized("未登录或登录已过期")
	}
	ticket := s.Get(req.TicketID)
	if ticket == nil {
		return nil, errorsx.InvalidParam("工单不存在")
	}
	content := strings.TrimSpace(req.Content)
	if content == "" {
		return nil, errorsx.InvalidParam("备注内容不能为空")
	}
	comment := &models.TicketComment{
		TicketID:    ticket.ID,
		CommentType: enums.TicketCommentTypeInternalNote,
		AuthorType:  enums.IMSenderTypeAgent,
		AuthorID:    operator.UserID,
		ContentType: strings.TrimSpace(req.ContentType),
		Content:     content,
		Payload:     strings.TrimSpace(req.Payload),
		CreatedAt:   time.Now(),
	}
	if comment.ContentType == "" {
		comment.ContentType = "text"
	}
	err := sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		if err := repositories.TicketCommentRepository.Create(ctx.Tx, comment); err != nil {
			return err
		}
		if err := repositories.TicketRepository.Updates(ctx.Tx, ticket.ID, map[string]any{
			"update_user_id":   operator.UserID,
			"update_user_name": operator.Username,
			"updated_at":       time.Now(),
		}); err != nil {
			return err
		}
		return s.logEvent(ctx.Tx, ticket.ID, enums.TicketEventTypeInternalNoted, operator, "", "", "添加内部备注", "")
	})
	if err != nil {
		return nil, err
	}
	return comment, nil
}

func (s *ticketService) CloseTicket(req request.CloseTicketRequest, operator *dto.AuthPrincipal) error {
	return s.ChangeStatus(request.ChangeTicketStatusRequest{
		TicketID:    req.TicketID,
		Status:      string(enums.TicketStatusClosed),
		CloseReason: req.CloseReason,
		Reason:      "关闭工单",
	}, operator)
}

func (s *ticketService) ReopenTicket(req request.ReopenTicketRequest, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	ticket := s.Get(req.TicketID)
	if ticket == nil {
		return errorsx.InvalidParam("工单不存在")
	}
	if ticket.Status != enums.TicketStatusClosed && ticket.Status != enums.TicketStatusResolved {
		return errorsx.InvalidParam("当前工单状态不允许重开")
	}
	now := time.Now()
	return sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		if err := repositories.TicketRepository.Updates(ctx.Tx, ticket.ID, map[string]any{
			"status":           enums.TicketStatusOpen,
			"close_reason":     "",
			"closed_at":        nil,
			"resolved_at":      nil,
			"reopened_count":   ticket.ReopenedCount + 1,
			"update_user_id":   operator.UserID,
			"update_user_name": operator.Username,
			"updated_at":       now,
		}); err != nil {
			return err
		}
		if err := s.resumeResolutionSLA(ctx.Tx, ticket.ID, now); err != nil {
			return err
		}
		return s.logEvent(ctx.Tx, ticket.ID, enums.TicketEventTypeReopened, operator, string(ticket.Status), string(enums.TicketStatusOpen), strings.TrimSpace(req.Reason), "")
	})
}

func (s *ticketService) WatchTicket(ticketID int64, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	ticket := s.Get(ticketID)
	if ticket == nil {
		return errorsx.InvalidParam("工单不存在")
	}
	existing := TicketWatcherService.FindOne(sqls.NewCnd().
		Eq("ticket_id", ticketID).
		Eq("user_id", operator.UserID))
	if existing != nil {
		return nil
	}
	return repositories.TicketWatcherRepository.Create(sqls.DB(), &models.TicketWatcher{
		TicketID:  ticketID,
		UserID:    operator.UserID,
		CreatedAt: time.Now(),
	})
}

func (s *ticketService) UnwatchTicket(ticketID int64, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	ticket := s.Get(ticketID)
	if ticket == nil {
		return errorsx.InvalidParam("工单不存在")
	}
	existing := TicketWatcherService.FindOne(sqls.NewCnd().
		Eq("ticket_id", ticketID).
		Eq("user_id", operator.UserID))
	if existing == nil {
		return nil
	}
	repositories.TicketWatcherRepository.Delete(sqls.DB(), existing.ID)
	return nil
}

func (s *ticketService) nextTicketNo() string {
	now := time.Now()
	seq := atomic.AddUint64(&ticketNoSequence, 1) % 1000
	return fmt.Sprintf("TK%s%09d%03d", now.Format("20060102150405"), now.UnixNano()%1e9, seq)
}

func (s *ticketService) normalizeAssignment(teamID, assigneeID int64) (int64, int64, error) {
	if assigneeID > 0 {
		profile := AgentProfileService.GetByUserID(assigneeID)
		if profile == nil || profile.Status != enums.StatusOk {
			return 0, 0, errorsx.InvalidParam("目标处理人不存在")
		}
		if teamID <= 0 {
			teamID = profile.TeamID
		}
	}
	if teamID > 0 {
		team := AgentTeamService.Get(teamID)
		if team == nil || team.Status != enums.StatusOk {
			return 0, 0, errorsx.InvalidParam("处理团队不存在")
		}
	}
	return teamID, assigneeID, nil
}

func (s *ticketService) initSLAs(tx *gorm.DB, ticket *models.Ticket, now time.Time) error {
	firstResponseTarget, resolutionTarget := ticketSLATargetMinutes(ticket.Priority)
	records := []*models.TicketSLARecord{
		{
			TicketID:      ticket.ID,
			SLAType:       enums.TicketSLATypeFirstResponse,
			TargetMinutes: firstResponseTarget,
			Status:        enums.TicketSLAStatusRunning,
			StartedAt:     &now,
			CreatedAt:     now,
			UpdatedAt:     now,
		},
		{
			TicketID:      ticket.ID,
			SLAType:       enums.TicketSLATypeResolution,
			TargetMinutes: resolutionTarget,
			Status:        enums.TicketSLAStatusRunning,
			StartedAt:     &now,
			CreatedAt:     now,
			UpdatedAt:     now,
		},
	}
	for _, item := range records {
		if err := repositories.TicketSLARecordRepository.Create(tx, item); err != nil {
			return err
		}
	}
	return nil
}

func (s *ticketService) logEvent(tx *gorm.DB, ticketID int64, eventType enums.TicketEventType, operator *dto.AuthPrincipal, oldValue, newValue, content, payload string) error {
	operatorID := int64(0)
	if operator != nil {
		operatorID = operator.UserID
	}
	return repositories.TicketEventLogRepository.Create(tx, &models.TicketEventLog{
		TicketID:     ticketID,
		EventType:    eventType,
		OperatorType: enums.IMSenderTypeAgent,
		OperatorID:   operatorID,
		OldValue:     strings.TrimSpace(oldValue),
		NewValue:     strings.TrimSpace(newValue),
		Content:      strings.TrimSpace(content),
		Payload:      strings.TrimSpace(payload),
		CreatedAt:    time.Now(),
	})
}

func (s *ticketService) completeFirstResponseSLA(tx *gorm.DB, ticketID int64, now time.Time) error {
	record := s.findSLARecord(tx, ticketID, enums.TicketSLATypeFirstResponse)
	if record == nil || record.Status == enums.TicketSLAStatusCompleted {
		return nil
	}
	elapsed := diffMinutes(record.StartedAt, now)
	if err := repositories.TicketSLARecordRepository.Updates(tx, record.ID, map[string]any{
		"status":      enums.TicketSLAStatusCompleted,
		"stopped_at":  now,
		"elapsed_min": elapsed,
		"updated_at":  now,
	}); err != nil {
		return err
	}
	return repositories.TicketRepository.Updates(tx, ticketID, map[string]any{
		"next_reply_deadline_at": nil,
		"updated_at":             now,
	})
}

func (s *ticketService) applySLAOnStatusChange(tx *gorm.DB, ticketID int64, fromStatus, toStatus enums.TicketStatus, now time.Time) error {
	switch toStatus {
	case enums.TicketStatusPendingCustomer:
		return s.pauseResolutionSLA(tx, ticketID, now)
	case enums.TicketStatusOpen:
		return s.resumeResolutionSLA(tx, ticketID, now)
	case enums.TicketStatusResolved, enums.TicketStatusClosed, enums.TicketStatusCancelled:
		return s.completeResolutionSLA(tx, ticketID, now)
	default:
		if fromStatus == enums.TicketStatusPendingCustomer && toStatus != enums.TicketStatusPendingCustomer {
			return s.resumeResolutionSLA(tx, ticketID, now)
		}
	}
	return nil
}

func (s *ticketService) pauseResolutionSLA(tx *gorm.DB, ticketID int64, now time.Time) error {
	record := s.findSLARecord(tx, ticketID, enums.TicketSLATypeResolution)
	if record == nil || record.Status != enums.TicketSLAStatusRunning {
		return nil
	}
	elapsed := record.ElapsedMin + diffMinutes(record.StartedAt, now)
	if err := repositories.TicketSLARecordRepository.Updates(tx, record.ID, map[string]any{
		"status":      enums.TicketSLAStatusPaused,
		"paused_at":   now,
		"elapsed_min": elapsed,
		"updated_at":  now,
	}); err != nil {
		return err
	}
	return repositories.TicketRepository.Updates(tx, ticketID, map[string]any{
		"resolve_deadline_at": nil,
		"updated_at":          now,
	})
}

func (s *ticketService) resumeResolutionSLA(tx *gorm.DB, ticketID int64, now time.Time) error {
	record := s.findSLARecord(tx, ticketID, enums.TicketSLATypeResolution)
	if record == nil || record.Status == enums.TicketSLAStatusCompleted {
		return nil
	}
	if err := repositories.TicketSLARecordRepository.Updates(tx, record.ID, map[string]any{
		"status":     enums.TicketSLAStatusRunning,
		"started_at": now,
		"paused_at":  nil,
		"updated_at": now,
	}); err != nil {
		return err
	}
	targetMinutes := record.TargetMinutes - record.ElapsedMin
	return repositories.TicketRepository.Updates(tx, ticketID, map[string]any{
		"resolve_deadline_at": buildDeadlineFromNow(now, targetMinutes),
		"updated_at":          now,
	})
}

func (s *ticketService) completeResolutionSLA(tx *gorm.DB, ticketID int64, now time.Time) error {
	record := s.findSLARecord(tx, ticketID, enums.TicketSLATypeResolution)
	if record == nil || record.Status == enums.TicketSLAStatusCompleted {
		return nil
	}
	elapsed := record.ElapsedMin
	if record.Status == enums.TicketSLAStatusRunning {
		elapsed += diffMinutes(record.StartedAt, now)
	}
	if err := repositories.TicketSLARecordRepository.Updates(tx, record.ID, map[string]any{
		"status":      enums.TicketSLAStatusCompleted,
		"stopped_at":  now,
		"paused_at":   nil,
		"elapsed_min": elapsed,
		"updated_at":  now,
	}); err != nil {
		return err
	}
	return repositories.TicketRepository.Updates(tx, ticketID, map[string]any{
		"resolve_deadline_at": nil,
		"updated_at":          now,
	})
}

func (s *ticketService) findSLARecord(tx *gorm.DB, ticketID int64, slaType enums.TicketSLAType) *models.TicketSLARecord {
	return repositories.TicketSLARecordRepository.FindOne(
		tx,
		sqls.NewCnd().Eq("ticket_id", ticketID).Eq("sla_type", slaType),
	)
}

func (s *ticketService) syncSLATargetsAndDeadlines(tx *gorm.DB, ticketID int64, priority enums.TicketPriority, now time.Time) error {
	firstResponseTarget, resolutionTarget := ticketSLATargetMinutes(priority)
	firstResponseRecord := s.findSLARecord(tx, ticketID, enums.TicketSLATypeFirstResponse)
	if firstResponseRecord != nil && firstResponseRecord.TargetMinutes != firstResponseTarget {
		if err := repositories.TicketSLARecordRepository.Updates(tx, firstResponseRecord.ID, map[string]any{
			"target_minutes": firstResponseTarget,
			"updated_at":     now,
		}); err != nil {
			return err
		}
		firstResponseRecord.TargetMinutes = firstResponseTarget
	}
	resolutionRecord := s.findSLARecord(tx, ticketID, enums.TicketSLATypeResolution)
	if resolutionRecord != nil && resolutionRecord.TargetMinutes != resolutionTarget {
		if err := repositories.TicketSLARecordRepository.Updates(tx, resolutionRecord.ID, map[string]any{
			"target_minutes": resolutionTarget,
			"updated_at":     now,
		}); err != nil {
			return err
		}
		resolutionRecord.TargetMinutes = resolutionTarget
	}
	return repositories.TicketRepository.Updates(tx, ticketID, map[string]any{
		"next_reply_deadline_at": s.calculateTicketDeadline(firstResponseRecord),
		"resolve_deadline_at":    s.calculateTicketDeadline(resolutionRecord),
		"updated_at":             now,
	})
}

func (s *ticketService) calculateTicketDeadline(record *models.TicketSLARecord) *time.Time {
	if record == nil {
		return nil
	}
	switch record.Status {
	case enums.TicketSLAStatusCompleted, enums.TicketSLAStatusBreached, enums.TicketSLAStatusPaused:
		return nil
	case enums.TicketSLAStatusRunning:
		if record.StartedAt == nil || record.StartedAt.IsZero() {
			return nil
		}
		remainingMinutes := record.TargetMinutes - record.ElapsedMin
		deadline := record.StartedAt.Add(time.Duration(remainingMinutes) * time.Minute)
		return &deadline
	default:
		return nil
	}
}

func (s *ticketService) canTransition(from, to enums.TicketStatus) bool {
	if from == to {
		return true
	}
	switch from {
	case enums.TicketStatusNew:
		return to == enums.TicketStatusOpen || to == enums.TicketStatusCancelled || to == enums.TicketStatusPendingInternal
	case enums.TicketStatusOpen:
		return to == enums.TicketStatusPendingCustomer || to == enums.TicketStatusPendingInternal || to == enums.TicketStatusResolved || to == enums.TicketStatusClosed || to == enums.TicketStatusCancelled
	case enums.TicketStatusPendingCustomer:
		return to == enums.TicketStatusOpen || to == enums.TicketStatusResolved || to == enums.TicketStatusClosed || to == enums.TicketStatusCancelled
	case enums.TicketStatusPendingInternal:
		return to == enums.TicketStatusOpen || to == enums.TicketStatusResolved || to == enums.TicketStatusClosed || to == enums.TicketStatusCancelled
	case enums.TicketStatusResolved:
		return to == enums.TicketStatusOpen || to == enums.TicketStatusClosed
	case enums.TicketStatusClosed:
		return to == enums.TicketStatusOpen
	default:
		return false
	}
}

func (s *ticketService) isEditableStatus(status enums.TicketStatus) bool {
	return status != enums.TicketStatusClosed && status != enums.TicketStatusCancelled
}

func ticketSLATargetMinutes(priority enums.TicketPriority) (int, int) {
	switch priority {
	case enums.TicketPriorityHigh:
		return 10, 240
	case enums.TicketPriorityUrgent:
		return 5, 120
	default:
		return 30, 1440
	}
}

func marshalJSON(value any) (string, error) {
	if value == nil {
		return "", nil
	}
	buf, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func parseOptionalDateTime(value string) (*time.Time, error) {
	raw := strings.TrimSpace(value)
	if raw == "" {
		return nil, nil
	}
	if t, err := time.ParseInLocation(time.DateTime, raw, time.Local); err == nil {
		return &t, nil
	}
	if t, err := time.ParseInLocation("2006-01-02T15:04:05", raw, time.Local); err == nil {
		return &t, nil
	}
	return nil, fmt.Errorf("invalid datetime")
}

func diffMinutes(start *time.Time, end time.Time) int {
	if start == nil || start.IsZero() {
		return 0
	}
	return int(end.Sub(*start).Minutes())
}

func buildDeadlineFromNow(now time.Time, targetMinutes int) *time.Time {
	deadline := now.Add(time.Duration(targetMinutes) * time.Minute)
	return &deadline
}
