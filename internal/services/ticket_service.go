package services

import (
	"encoding/json"
	"fmt"
	"strings"
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
	Ticket         *models.Ticket
	Customer       *models.Customer
	SLAs           []models.TicketSLARecord
	Watchers       []models.TicketWatcher
	Collaborators  []models.TicketCollaborator
	Comments       []models.TicketComment
	Events         []models.TicketEventLog
	RelatedTickets []models.TicketRelation
}

type TicketSummaryAggregate struct {
	All             int64
	Mine            int64
	Watching        int64
	Unassigned      int64
	PendingCustomer int64
	PendingInternal int64
	Overdue         int64
}

type TicketRiskReasonAggregate struct {
	Code        string
	Title       string
	Description string
	Count       int64
}

type TicketRiskOverviewAggregate struct {
	Overdue         int64
	HighRisk        int64
	Unassigned      int64
	PendingInternal int64
	PendingCustomer int64
	RiskWindowMins  int
	Reasons         []TicketRiskReasonAggregate
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
		Collaborators: TicketCollaboratorService.Find(
			sqls.NewCnd().Eq("ticket_id", id).Asc("id"),
		),
		Comments: TicketCommentService.Find(
			sqls.NewCnd().Eq("ticket_id", id).Asc("id"),
		),
		Events: TicketEventLogService.Find(
			sqls.NewCnd().Eq("ticket_id", id).Desc("id"),
		),
		RelatedTickets: TicketRelationService.Find(
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
	return &TicketSummaryAggregate{
		All: s.Count(sqls.NewCnd()),
		Mine: s.Count(
			sqls.NewCnd().Eq("current_assignee_id", operator.UserID),
		),
		Watching: s.Count(
			sqls.NewCnd().Where("id IN (SELECT ticket_id FROM ticket_watchers WHERE user_id = ?)", operator.UserID),
		),
		Unassigned: s.Count(
			sqls.NewCnd().
				In("status", []enums.TicketStatus{
					enums.TicketStatusNew,
					enums.TicketStatusOpen,
					enums.TicketStatusPendingCustomer,
					enums.TicketStatusPendingInternal,
				}).
				Eq("current_assignee_id", 0),
		),
		PendingCustomer: s.Count(
			sqls.NewCnd().Eq("status", enums.TicketStatusPendingCustomer),
		),
		PendingInternal: s.Count(
			sqls.NewCnd().Eq("status", enums.TicketStatusPendingInternal),
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

func (s *ticketService) GetRiskOverview(teamID int64, riskWindowMins int) *TicketRiskOverviewAggregate {
	if riskWindowMins <= 0 {
		riskWindowMins = 240
	}
	now := time.Now()
	staleAt := now.Add(-24 * time.Hour)
	activeBaseCnd := func() *sqls.Cnd {
		cnd := sqls.NewCnd().In("status", []enums.TicketStatus{
			enums.TicketStatusNew,
			enums.TicketStatusOpen,
			enums.TicketStatusPendingCustomer,
			enums.TicketStatusPendingInternal,
		})
		if teamID > 0 {
			cnd.Eq("current_team_id", teamID)
		}
		return cnd
	}
	pendingCustomerBaseCnd := func() *sqls.Cnd {
		cnd := sqls.NewCnd().Eq("status", enums.TicketStatusPendingCustomer)
		if teamID > 0 {
			cnd.Eq("current_team_id", teamID)
		}
		return cnd
	}
	pendingInternalBaseCnd := func() *sqls.Cnd {
		cnd := sqls.NewCnd().Eq("status", enums.TicketStatusPendingInternal)
		if teamID > 0 {
			cnd.Eq("current_team_id", teamID)
		}
		return cnd
	}

	overview := &TicketRiskOverviewAggregate{
		Overdue: s.Count(
			activeBaseCnd().
				Where("resolve_deadline_at IS NOT NULL").
				Where("resolve_deadline_at < ?", now),
		),
		HighRisk: s.Count(
			activeBaseCnd().
				Where("resolve_deadline_at IS NOT NULL").
				Where("resolve_deadline_at >= ?", now).
				Where("resolve_deadline_at <= ?", now.Add(time.Duration(riskWindowMins)*time.Minute)),
		),
		Unassigned: s.Count(
			activeBaseCnd().Eq("current_assignee_id", 0),
		),
		PendingInternal: s.Count(pendingInternalBaseCnd()),
		PendingCustomer: s.Count(pendingCustomerBaseCnd()),
		RiskWindowMins:  riskWindowMins,
		Reasons: []TicketRiskReasonAggregate{
			{
				Code:        "unassigned_active",
				Title:       "工单未分配",
				Description: "仍处于活跃状态，但没有明确负责人的工单",
				Count: s.Count(
					activeBaseCnd().Eq("current_assignee_id", 0),
				),
			},
			{
				Code:        "pending_internal_stale",
				Title:       "待内部处理滞留超过 24 小时",
				Description: "内部协作未及时推进，容易形成长期积压",
				Count: s.Count(
					pendingInternalBaseCnd().Where("updated_at < ?", staleAt),
				),
			},
			{
				Code:        "pending_customer_stale",
				Title:       "待客户反馈滞留超过 24 小时",
				Description: "客户迟迟未补充信息，建议催办或关单",
				Count: s.Count(
					pendingCustomerBaseCnd().Where("updated_at < ?", staleAt),
				),
			},
			{
				Code:        "active_without_deadline",
				Title:       "活跃工单未设置解决时限",
				Description: "缺少 SLA 截止时间，主管无法有效盯防风险",
				Count: s.Count(
					activeBaseCnd().Where("resolve_deadline_at IS NULL"),
				),
			},
		},
	}
	return overview
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
	if req.CategoryID > 0 {
		category := TicketCategoryService.Get(req.CategoryID)
		if category == nil || category.Status != enums.StatusOk {
			return nil, errorsx.InvalidParam("工单分类不存在")
		}
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
	ticket := &models.Ticket{
		TicketNo:          s.nextTicketNo(),
		Title:             title,
		Description:       strings.TrimSpace(req.Description),
		Source:            source,
		Channel:           strings.TrimSpace(req.Channel),
		CustomerID:        req.CustomerID,
		ConversationID:    req.ConversationID,
		CategoryID:        req.CategoryID,
		Type:              strings.TrimSpace(req.Type),
		Priority:          priority,
		Severity:          severity,
		Status:            status,
		CurrentTeamID:     teamID,
		CurrentAssigneeID: assigneeID,
		DueAt:             dueAt,
		CustomFieldsJSON:  customFieldsJSON,
		AuditFields:       utils.BuildAuditFields(operator),
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
	if req.CategoryID > 0 {
		category := TicketCategoryService.Get(req.CategoryID)
		if category == nil || category.Status != enums.StatusOk {
			return errorsx.InvalidParam("工单分类不存在")
		}
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
			"updated_at":          time.Now(),
		}); err != nil {
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
		resolutionCode := strings.TrimSpace(req.ResolutionCode)
		if resolutionCode != "" {
			code := TicketResolutionCodeService.Take("code = ? AND status = ?", resolutionCode, enums.StatusOk)
			if code == nil {
				return errorsx.InvalidParam("解决码不存在")
			}
		}
		values["resolution_code"] = resolutionCode
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

func (s *ticketService) AddCollaborator(ticketID, userID int64, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	ticket := s.Get(ticketID)
	if ticket == nil {
		return errorsx.InvalidParam("工单不存在")
	}
	if userID <= 0 {
		return errorsx.InvalidParam("协作人不能为空")
	}
	profile := AgentProfileService.GetByUserID(userID)
	if profile == nil || profile.Status != enums.StatusOk {
		return errorsx.InvalidParam("协作人不存在")
	}
	if TicketCollaboratorService.FindOne(sqls.NewCnd().Eq("ticket_id", ticketID).Eq("user_id", userID)) != nil {
		return nil
	}
	now := time.Now()
	userName := ""
	if user := UserService.Get(userID); user != nil {
		userName = user.Nickname
		if userName == "" {
			userName = user.Username
		}
	}
	return sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		if err := repositories.TicketCollaboratorRepository.Create(ctx.Tx, &models.TicketCollaborator{
			TicketID:  ticketID,
			UserID:    userID,
			CreatedAt: now,
		}); err != nil {
			return err
		}
		if TicketWatcherService.FindOne(sqls.NewCnd().Eq("ticket_id", ticketID).Eq("user_id", userID)) == nil {
			if err := repositories.TicketWatcherRepository.Create(ctx.Tx, &models.TicketWatcher{
				TicketID:  ticketID,
				UserID:    userID,
				CreatedAt: now,
			}); err != nil {
				return err
			}
		}
		return s.logEvent(ctx.Tx, ticketID, enums.TicketEventTypeUpdated, operator, "", "", fmt.Sprintf("新增协作人：%s", userName), "")
	})
}

func (s *ticketService) RemoveCollaborator(ticketID, collaboratorID int64, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	ticket := s.Get(ticketID)
	if ticket == nil {
		return errorsx.InvalidParam("工单不存在")
	}
	item := TicketCollaboratorService.Get(collaboratorID)
	if item == nil || item.TicketID != ticketID {
		return errorsx.InvalidParam("协作关系不存在")
	}
	userName := ""
	if user := UserService.Get(item.UserID); user != nil {
		userName = user.Nickname
		if userName == "" {
			userName = user.Username
		}
	}
	return sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		repositories.TicketCollaboratorRepository.Delete(ctx.Tx, collaboratorID)
		return s.logEvent(ctx.Tx, ticketID, enums.TicketEventTypeUpdated, operator, "", "", fmt.Sprintf("移除协作人：%s", userName), "")
	})
}

func (s *ticketService) BatchAssignTickets(req request.BatchAssignTicketRequest, operator *dto.AuthPrincipal) error {
	ticketIDs := normalizeBatchTicketIDs(req.TicketIDs)
	if len(ticketIDs) == 0 {
		return errorsx.InvalidParam("请选择工单")
	}
	for _, ticketID := range ticketIDs {
		if err := s.AssignTicket(request.AssignTicketRequest{
			TicketID: ticketID,
			ToUserID: req.ToUserID,
			ToTeamID: req.ToTeamID,
			Reason:   req.Reason,
		}, operator); err != nil {
			return err
		}
	}
	return nil
}

func (s *ticketService) BatchChangeStatus(req request.BatchChangeTicketStatusRequest, operator *dto.AuthPrincipal) error {
	ticketIDs := normalizeBatchTicketIDs(req.TicketIDs)
	if len(ticketIDs) == 0 {
		return errorsx.InvalidParam("请选择工单")
	}
	for _, ticketID := range ticketIDs {
		if err := s.ChangeStatus(request.ChangeTicketStatusRequest{
			TicketID:          ticketID,
			Status:            req.Status,
			PendingReason:     req.PendingReason,
			CloseReason:       req.CloseReason,
			ResolutionCode:    req.ResolutionCode,
			ResolutionSummary: req.ResolutionSummary,
			Reason:            req.Reason,
		}, operator); err != nil {
			return err
		}
	}
	return nil
}

func (s *ticketService) BatchWatchTickets(req request.BatchWatchTicketRequest, operator *dto.AuthPrincipal) error {
	ticketIDs := normalizeBatchTicketIDs(req.TicketIDs)
	if len(ticketIDs) == 0 {
		return errorsx.InvalidParam("请选择工单")
	}
	for _, ticketID := range ticketIDs {
		var err error
		if req.Watched {
			err = s.WatchTicket(ticketID, operator)
		} else {
			err = s.UnwatchTicket(ticketID, operator)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *ticketService) nextTicketNo() string {
	now := time.Now()
	return fmt.Sprintf("TK%s%03d", now.Format("20060102150405"), now.Nanosecond()/1e6)
}

func normalizeBatchTicketIDs(ticketIDs []int64) []int64 {
	if len(ticketIDs) == 0 {
		return nil
	}
	results := make([]int64, 0, len(ticketIDs))
	exists := make(map[int64]struct{}, len(ticketIDs))
	for _, ticketID := range ticketIDs {
		if ticketID <= 0 {
			continue
		}
		if _, ok := exists[ticketID]; ok {
			continue
		}
		exists[ticketID] = struct{}{}
		results = append(results, ticketID)
	}
	return results
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
	if config := TicketSLAConfigService.GetActiveByPriority(ticket.Priority); config != nil {
		firstResponseTarget = config.FirstResponseMinutes
		resolutionTarget = config.ResolutionMinutes
	}
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
	record := TicketSLARecordService.FindOne(sqls.NewCnd().Eq("ticket_id", ticketID).Eq("sla_type", enums.TicketSLATypeFirstResponse))
	if record == nil || record.Status == enums.TicketSLAStatusCompleted {
		return nil
	}
	elapsed := diffMinutes(record.StartedAt, now)
	return repositories.TicketSLARecordRepository.Updates(tx, record.ID, map[string]any{
		"status":      enums.TicketSLAStatusCompleted,
		"stopped_at":  now,
		"elapsed_min": elapsed,
		"updated_at":  now,
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
	record := TicketSLARecordService.FindOne(sqls.NewCnd().Eq("ticket_id", ticketID).Eq("sla_type", enums.TicketSLATypeResolution))
	if record == nil || record.Status != enums.TicketSLAStatusRunning {
		return nil
	}
	elapsed := record.ElapsedMin + diffMinutes(record.StartedAt, now)
	return repositories.TicketSLARecordRepository.Updates(tx, record.ID, map[string]any{
		"status":      enums.TicketSLAStatusPaused,
		"paused_at":   now,
		"elapsed_min": elapsed,
		"updated_at":  now,
	})
}

func (s *ticketService) resumeResolutionSLA(tx *gorm.DB, ticketID int64, now time.Time) error {
	record := TicketSLARecordService.FindOne(sqls.NewCnd().Eq("ticket_id", ticketID).Eq("sla_type", enums.TicketSLATypeResolution))
	if record == nil || record.Status == enums.TicketSLAStatusCompleted {
		return nil
	}
	return repositories.TicketSLARecordRepository.Updates(tx, record.ID, map[string]any{
		"status":     enums.TicketSLAStatusRunning,
		"started_at": now,
		"paused_at":  nil,
		"updated_at": now,
	})
}

func (s *ticketService) completeResolutionSLA(tx *gorm.DB, ticketID int64, now time.Time) error {
	record := TicketSLARecordService.FindOne(sqls.NewCnd().Eq("ticket_id", ticketID).Eq("sla_type", enums.TicketSLATypeResolution))
	if record == nil || record.Status == enums.TicketSLAStatusCompleted {
		return nil
	}
	elapsed := record.ElapsedMin
	if record.Status == enums.TicketSLAStatusRunning {
		elapsed += diffMinutes(record.StartedAt, now)
	}
	return repositories.TicketSLARecordRepository.Updates(tx, record.ID, map[string]any{
		"status":      enums.TicketSLAStatusCompleted,
		"stopped_at":  now,
		"paused_at":   nil,
		"elapsed_min": elapsed,
		"updated_at":  now,
	})
}

func (s *ticketService) ScanAndMarkBreachedSLAs(limit int) (int, error) {
	if limit <= 0 {
		limit = 200
	}
	now := time.Now()
	cnd := sqls.NewCnd().
		Eq("status", enums.TicketSLAStatusRunning).
		Where("started_at IS NOT NULL").
		Asc("id")
	records, _ := TicketSLARecordService.FindPageByCnd(cnd.Limit(limit))
	if len(records) == 0 {
		return 0, nil
	}

	breachedCount := 0
	for i := range records {
		record := &records[i]
		elapsed := record.ElapsedMin + diffMinutes(record.StartedAt, now)
		if elapsed < record.TargetMinutes {
			continue
		}
		if err := sqls.WithTransaction(func(ctx *sqls.TxContext) error {
			if err := repositories.TicketSLARecordRepository.Updates(ctx.Tx, record.ID, map[string]any{
				"status":      enums.TicketSLAStatusBreached,
				"breached_at": now,
				"elapsed_min": elapsed,
				"updated_at":  now,
			}); err != nil {
				return err
			}
			return s.logEvent(ctx.Tx, record.TicketID, enums.TicketEventTypeSLABreached, nil, string(enums.TicketSLAStatusRunning), string(enums.TicketSLAStatusBreached), fmt.Sprintf("%s SLA 已超时", record.SLAType), "")
		}); err != nil {
			return breachedCount, err
		}
		breachedCount++
	}
	return breachedCount, nil
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
