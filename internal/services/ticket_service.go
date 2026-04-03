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
	Ticket            *models.Ticket
	Customer          *models.Customer
	SLAs              []models.TicketSLARecord
	Watchers          []models.TicketWatcher
	Collaborators     []models.TicketCollaborator
	Comments          []models.TicketComment
	Events            []models.TicketEventLog
	RelatedTickets    []models.TicketRelation
	Users             map[int64]*models.User
	Teams             map[int64]*models.AgentTeam
	AgentProfiles     map[int64]*models.AgentProfile
	RelatedMap        map[int64]*models.Ticket
	OperatorUsers     map[int64]*models.User
	OperatorCustomers map[int64]*models.Customer
	OperatorAIAgents  map[int64]*models.AIAgent
}

type TicketSummaryAggregate struct {
	All             int64
	Mine            int64
	Watching        int64
	Collaboration   int64
	Participating   int64
	Mentioned       int64
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

type TicketListAggregate struct {
	List             []models.Ticket
	Paging           *sqls.Paging
	Categories       map[int64]*models.TicketCategory
	ResolutionCodes  map[string]*models.TicketResolutionCode
	Users            map[int64]*models.User
	Teams            map[int64]*models.AgentTeam
	Customers        map[int64]*models.Customer
	SLAByTicketID    map[int64][]models.TicketSLARecord
	WatchedTicketIDs map[int64]struct{}
}

type ticketService struct {
}

type ticketNotePayload struct {
	MentionUserIDs []int64 `json:"mentionUserIds,omitempty"`
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

func (s *ticketService) FindPageAggregateByCnd(cnd *sqls.Cnd, watcherUserID int64) (*TicketListAggregate, error) {
	list, paging := repositories.TicketRepository.FindPageByCnd(sqls.DB(), cnd)
	return s.buildTicketListAggregate(sqls.DB(), list, paging, watcherUserID), nil
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
		Users:             make(map[int64]*models.User),
		Teams:             make(map[int64]*models.AgentTeam),
		AgentProfiles:     make(map[int64]*models.AgentProfile),
		RelatedMap:        make(map[int64]*models.Ticket),
		OperatorUsers:     make(map[int64]*models.User),
		OperatorCustomers: make(map[int64]*models.Customer),
		OperatorAIAgents:  make(map[int64]*models.AIAgent),
	}
	if ticket.CustomerID > 0 {
		aggregate.Customer = CustomerService.Get(ticket.CustomerID)
	}
	s.enrichTicketDetailAggregate(aggregate)
	return aggregate, nil
}

func (s *ticketService) enrichTicketDetailAggregate(aggregate *TicketDetailAggregate) {
	if aggregate == nil {
		return
	}
	userIDs := make([]int64, 0)
	teamIDs := make([]int64, 0)
	relatedTicketIDs := make([]int64, 0)
	operatorCustomerIDs := make([]int64, 0)
	operatorAIIDs := make([]int64, 0)
	userSeen := make(map[int64]struct{})
	teamSeen := make(map[int64]struct{})
	relatedSeen := make(map[int64]struct{})
	customerSeen := make(map[int64]struct{})
	aiSeen := make(map[int64]struct{})
	addUserID := func(userID int64) {
		if userID <= 0 {
			return
		}
		if _, ok := userSeen[userID]; ok {
			return
		}
		userSeen[userID] = struct{}{}
		userIDs = append(userIDs, userID)
	}
	addTeamID := func(teamID int64) {
		if teamID <= 0 {
			return
		}
		if _, ok := teamSeen[teamID]; ok {
			return
		}
		teamSeen[teamID] = struct{}{}
		teamIDs = append(teamIDs, teamID)
	}
	addCustomerID := func(customerID int64) {
		if customerID <= 0 {
			return
		}
		if _, ok := customerSeen[customerID]; ok {
			return
		}
		customerSeen[customerID] = struct{}{}
		operatorCustomerIDs = append(operatorCustomerIDs, customerID)
	}
	addAIID := func(aiID int64) {
		if aiID <= 0 {
			return
		}
		if _, ok := aiSeen[aiID]; ok {
			return
		}
		aiSeen[aiID] = struct{}{}
		operatorAIIDs = append(operatorAIIDs, aiID)
	}
	for i := range aggregate.Watchers {
		addUserID(aggregate.Watchers[i].UserID)
	}
	for i := range aggregate.Collaborators {
		addUserID(aggregate.Collaborators[i].UserID)
	}
	for i := range aggregate.RelatedTickets {
		relatedTicketID := aggregate.RelatedTickets[i].RelatedTicketID
		if relatedTicketID <= 0 {
			continue
		}
		if _, ok := relatedSeen[relatedTicketID]; ok {
			continue
		}
		relatedSeen[relatedTicketID] = struct{}{}
		relatedTicketIDs = append(relatedTicketIDs, relatedTicketID)
	}
	for i := range aggregate.Comments {
		switch aggregate.Comments[i].AuthorType {
		case enums.IMSenderTypeAgent:
			addUserID(aggregate.Comments[i].AuthorID)
		case enums.IMSenderTypeCustomer:
			addCustomerID(aggregate.Comments[i].AuthorID)
		case enums.IMSenderTypeAI:
			addAIID(aggregate.Comments[i].AuthorID)
		}
	}
	for i := range aggregate.Events {
		switch aggregate.Events[i].OperatorType {
		case enums.IMSenderTypeAgent:
			addUserID(aggregate.Events[i].OperatorID)
		case enums.IMSenderTypeCustomer:
			addCustomerID(aggregate.Events[i].OperatorID)
		case enums.IMSenderTypeAI:
			addAIID(aggregate.Events[i].OperatorID)
		}
	}
	if len(userIDs) > 0 {
		users := repositories.UserRepository.FindByIds(sqls.DB(), userIDs)
		for i := range users {
			item := users[i]
			aggregate.Users[item.ID] = &item
			aggregate.OperatorUsers[item.ID] = &item
		}
		profiles := repositories.AgentProfileRepository.Find(sqls.DB(), sqls.NewCnd().In("user_id", userIDs))
		for i := range profiles {
			item := profiles[i]
			aggregate.AgentProfiles[item.UserID] = &item
			addTeamID(item.TeamID)
		}
	}
	if len(relatedTicketIDs) > 0 {
		relatedTickets := repositories.TicketRepository.Find(sqls.DB(), sqls.NewCnd().In("id", relatedTicketIDs))
		for i := range relatedTickets {
			item := relatedTickets[i]
			aggregate.RelatedMap[item.ID] = &item
			addUserID(item.CurrentAssigneeID)
			addTeamID(item.CurrentTeamID)
		}
	}
	if len(teamIDs) > 0 {
		teams := repositories.AgentTeamRepository.FindByIds(sqls.DB(), teamIDs)
		for i := range teams {
			item := teams[i]
			aggregate.Teams[item.ID] = &item
		}
	}
	if len(userIDs) > len(aggregate.Users) {
		users := repositories.UserRepository.FindByIds(sqls.DB(), userIDs)
		for i := range users {
			item := users[i]
			aggregate.Users[item.ID] = &item
			aggregate.OperatorUsers[item.ID] = &item
		}
	}
	if len(operatorCustomerIDs) > 0 {
		customers := repositories.CustomerRepository.Find(sqls.DB(), sqls.NewCnd().In("id", operatorCustomerIDs))
		for i := range customers {
			item := customers[i]
			aggregate.OperatorCustomers[item.ID] = &item
		}
	}
	if len(operatorAIIDs) > 0 {
		aiAgents := repositories.AIAgentRepository.FindByIds(sqls.DB(), operatorAIIDs)
		for i := range aiAgents {
			item := aiAgents[i]
			aggregate.OperatorAIAgents[item.ID] = &item
		}
	}
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
		Collaboration: s.Count(
			sqls.NewCnd().Where(
				"id IN (SELECT ticket_id FROM ticket_collaborators WHERE user_id = ?) OR id IN (SELECT ticket_id FROM ticket_mentions WHERE mentioned_user_id = ?)",
				operator.UserID,
				operator.UserID,
			),
		),
		Participating: s.Count(
			sqls.NewCnd().Where("id IN (SELECT ticket_id FROM ticket_collaborators WHERE user_id = ?)", operator.UserID),
		),
		Mentioned: s.Count(
			sqls.NewCnd().Where("id IN (SELECT ticket_id FROM ticket_mentions WHERE mentioned_user_id = ?)", operator.UserID),
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

func (s *ticketService) GetRiskPageAggregate(riskType string, teamID int64, riskWindowMins int, page, limit int, watcherUserID int64) (*TicketListAggregate, error) {
	cnd := s.buildRiskListCnd(riskType, teamID, riskWindowMins, page, limit)
	list, paging := repositories.TicketRepository.FindPageByCnd(sqls.DB(), cnd)
	return s.buildTicketListAggregate(sqls.DB(), list, paging, watcherUserID), nil
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

func (s *ticketService) buildRiskListCnd(riskType string, teamID int64, riskWindowMins int, page, limit int) *sqls.Cnd {
	if riskWindowMins <= 0 {
		riskWindowMins = 240
	}
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	now := time.Now()
	cnd := sqls.NewCnd().Page(page, limit)
	if teamID > 0 {
		cnd.Eq("current_team_id", teamID)
	}
	switch strings.TrimSpace(riskType) {
	case "overdue":
		cnd.In("status", activeTicketStatuses()).
			Where("resolve_deadline_at IS NOT NULL").
			Where("resolve_deadline_at < ?", now).
			Asc("resolve_deadline_at").
			Asc("id")
	case "high_risk":
		cnd.In("status", activeTicketStatuses()).
			Where("resolve_deadline_at IS NOT NULL").
			Where("resolve_deadline_at >= ?", now).
			Where("resolve_deadline_at <= ?", now.Add(time.Duration(riskWindowMins)*time.Minute)).
			Asc("resolve_deadline_at").
			Desc("priority").
			Desc("id")
	case "unassigned":
		cnd.In("status", activeTicketStatuses()).
			Eq("current_assignee_id", 0).
			Desc("updated_at").
			Desc("id")
	case "pending_internal":
		cnd.Eq("status", enums.TicketStatusPendingInternal).
			Desc("updated_at").
			Desc("id")
	case "pending_customer":
		cnd.Eq("status", enums.TicketStatusPendingCustomer).
			Desc("updated_at").
			Desc("id")
	default:
		cnd.In("status", activeTicketStatuses()).
			Desc("updated_at").
			Desc("id")
	}
	return cnd
}

func (s *ticketService) buildTicketListAggregate(db *gorm.DB, list []models.Ticket, paging *sqls.Paging, watcherUserID int64) *TicketListAggregate {
	aggregate := &TicketListAggregate{
		List:             list,
		Paging:           paging,
		Categories:       make(map[int64]*models.TicketCategory),
		ResolutionCodes:  make(map[string]*models.TicketResolutionCode),
		Users:            make(map[int64]*models.User),
		Teams:            make(map[int64]*models.AgentTeam),
		Customers:        make(map[int64]*models.Customer),
		SLAByTicketID:    make(map[int64][]models.TicketSLARecord),
		WatchedTicketIDs: make(map[int64]struct{}),
	}
	if len(list) == 0 {
		return aggregate
	}

	categoryIDs := make([]int64, 0)
	customerIDs := make([]int64, 0)
	teamIDs := make([]int64, 0)
	userIDs := make([]int64, 0)
	ticketIDs := make([]int64, 0, len(list))
	resolutionCodes := make([]string, 0)

	categorySeen := map[int64]struct{}{}
	customerSeen := map[int64]struct{}{}
	teamSeen := map[int64]struct{}{}
	userSeen := map[int64]struct{}{}
	ticketSeen := map[int64]struct{}{}
	codeSeen := map[string]struct{}{}

	for i := range list {
		item := &list[i]
		if _, ok := ticketSeen[item.ID]; !ok {
			ticketSeen[item.ID] = struct{}{}
			ticketIDs = append(ticketIDs, item.ID)
		}
		if item.CategoryID > 0 {
			if _, ok := categorySeen[item.CategoryID]; !ok {
				categorySeen[item.CategoryID] = struct{}{}
				categoryIDs = append(categoryIDs, item.CategoryID)
			}
		}
		if item.CustomerID > 0 {
			if _, ok := customerSeen[item.CustomerID]; !ok {
				customerSeen[item.CustomerID] = struct{}{}
				customerIDs = append(customerIDs, item.CustomerID)
			}
		}
		if item.CurrentTeamID > 0 {
			if _, ok := teamSeen[item.CurrentTeamID]; !ok {
				teamSeen[item.CurrentTeamID] = struct{}{}
				teamIDs = append(teamIDs, item.CurrentTeamID)
			}
		}
		if item.CurrentAssigneeID > 0 {
			if _, ok := userSeen[item.CurrentAssigneeID]; !ok {
				userSeen[item.CurrentAssigneeID] = struct{}{}
				userIDs = append(userIDs, item.CurrentAssigneeID)
			}
		}
		code := strings.TrimSpace(item.ResolutionCode)
		if code != "" {
			if _, ok := codeSeen[code]; !ok {
				codeSeen[code] = struct{}{}
				resolutionCodes = append(resolutionCodes, code)
			}
		}
	}

	if len(categoryIDs) > 0 {
		categories := repositories.TicketCategoryRepository.Find(db, sqls.NewCnd().In("id", categoryIDs))
		for i := range categories {
			item := categories[i]
			aggregate.Categories[item.ID] = &item
		}
	}
	if len(resolutionCodes) > 0 {
		codeItems := repositories.TicketResolutionCodeRepository.Find(db, sqls.NewCnd().In("code", resolutionCodes).NotEq("status", enums.StatusDeleted))
		for i := range codeItems {
			item := codeItems[i]
			aggregate.ResolutionCodes[item.Code] = &item
		}
	}
	users := repositories.UserRepository.FindByIds(db, userIDs)
	for i := range users {
		item := users[i]
		aggregate.Users[item.ID] = &item
	}
	teams := repositories.AgentTeamRepository.FindByIds(db, teamIDs)
	for i := range teams {
		item := teams[i]
		aggregate.Teams[item.ID] = &item
	}
	if len(customerIDs) > 0 {
		customers := repositories.CustomerRepository.Find(db, sqls.NewCnd().In("id", customerIDs))
		for i := range customers {
			item := customers[i]
			aggregate.Customers[item.ID] = &item
		}
	}
	if len(ticketIDs) > 0 {
		slaRecords := repositories.TicketSLARecordRepository.Find(db, sqls.NewCnd().In("ticket_id", ticketIDs).Asc("id"))
		for i := range slaRecords {
			item := slaRecords[i]
			aggregate.SLAByTicketID[item.TicketID] = append(aggregate.SLAByTicketID[item.TicketID], item)
		}
	}
	if watcherUserID > 0 {
		watchers := repositories.TicketWatcherRepository.Find(db, sqls.NewCnd().Eq("user_id", watcherUserID).In("ticket_id", ticketIDs))
		for i := range watchers {
			aggregate.WatchedTicketIDs[watchers[i].TicketID] = struct{}{}
		}
	}
	return aggregate
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
		ticketNo, err := TicketNoService.Next(ctx.Tx, now)
		if err != nil {
			return err
		}
		ticket.TicketNo = ticketNo
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
	return sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		return s.assignTicketTx(ctx.Tx, req, operator)
	})
}

func (s *ticketService) assignTicketTx(tx *gorm.DB, req request.AssignTicketRequest, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	ticket := repositories.TicketRepository.Get(tx, req.TicketID)
	if ticket == nil {
		return errorsx.InvalidParam("工单不存在")
	}
	teamID, assigneeID, err := s.normalizeAssignmentTx(tx, req.ToTeamID, req.ToUserID)
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
	if err := repositories.TicketRepository.Updates(tx, ticket.ID, map[string]any{
		"current_team_id":     teamID,
		"current_assignee_id": assigneeID,
		"status":              status,
		"update_user_id":      operator.UserID,
		"update_user_name":    operator.Username,
		"updated_at":          now,
	}); err != nil {
		return err
	}
	return s.logEvent(tx, ticket.ID, eventType, operator, fmt.Sprintf("%d", ticket.CurrentAssigneeID), fmt.Sprintf("%d", assigneeID), strings.TrimSpace(content), strings.TrimSpace(req.Reason))
}

func (s *ticketService) ChangeStatus(req request.ChangeTicketStatusRequest, operator *dto.AuthPrincipal) error {
	return sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		return s.changeStatusTx(ctx.Tx, req, operator)
	})
}

func (s *ticketService) changeStatusTx(tx *gorm.DB, req request.ChangeTicketStatusRequest, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	ticket := repositories.TicketRepository.Get(tx, req.TicketID)
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
	if targetStatus == enums.TicketStatusClosed {
		if childNos := s.findOpenChildTicketNosTx(tx, ticket.ID); len(childNos) > 0 {
			if len(childNos) > 3 {
				return errorsx.InvalidParam(fmt.Sprintf("仍有未完成子工单，暂不可关闭：%s 等 %d 张", strings.Join(childNos[:3], "、"), len(childNos)))
			}
			return errorsx.InvalidParam(fmt.Sprintf("仍有未完成子工单，暂不可关闭：%s", strings.Join(childNos, "、")))
		}
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
			code := repositories.TicketResolutionCodeRepository.Take(tx, "code = ? AND status = ?", resolutionCode, enums.StatusOk)
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
	if err := repositories.TicketRepository.Updates(tx, ticket.ID, values); err != nil {
		return err
	}
	if err := s.applySLAOnStatusChange(tx, ticket.ID, ticket.Status, targetStatus, now); err != nil {
		return err
	}
	if err := s.refreshTicketSLAFields(tx, ticket.ID, now); err != nil {
		return err
	}
	return s.logEvent(tx, ticket.ID, enums.TicketEventTypeStatusChanged, operator, string(ticket.Status), string(targetStatus), strings.TrimSpace(req.Reason), "")
}

func (s *ticketService) findOpenChildTicketNos(ticketID int64) []string {
	return s.findOpenChildTicketNosTx(sqls.DB(), ticketID)
}

func (s *ticketService) findOpenChildTicketNosTx(db *gorm.DB, ticketID int64) []string {
	relations := repositories.TicketRelationRepository.Find(db,
		sqls.NewCnd().
			Eq("ticket_id", ticketID).
			Eq("relation_type", enums.TicketRelationTypeChild),
	)
	if len(relations) == 0 {
		return nil
	}
	results := make([]string, 0, len(relations))
	for _, relation := range relations {
		child := repositories.TicketRepository.Get(db, relation.RelatedTicketID)
		if child == nil {
			continue
		}
		switch child.Status {
		case enums.TicketStatusResolved, enums.TicketStatusClosed, enums.TicketStatusCancelled:
			continue
		default:
			if child.TicketNo != "" {
				results = append(results, child.TicketNo)
			} else {
				results = append(results, fmt.Sprintf("#%d", child.ID))
			}
		}
	}
	return results
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
		if err := s.refreshTicketSLAFields(ctx.Tx, ticket.ID, now); err != nil {
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
	notePayload := parseTicketNotePayload(comment.Payload)
	err := sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		if err := repositories.TicketCommentRepository.Create(ctx.Tx, comment); err != nil {
			return err
		}
		mentionedNames := make([]string, 0, len(notePayload.MentionUserIDs))
		for _, userID := range notePayload.MentionUserIDs {
			if userID <= 0 || userID == operator.UserID {
				continue
			}
			if user := UserService.Get(userID); user != nil {
				name := user.Nickname
				if name == "" {
					name = user.Username
				}
				if name != "" {
					mentionedNames = append(mentionedNames, name)
				}
			}
			if repositories.TicketCollaboratorRepository.TakeByTicketIDAndUserID(ctx.Tx, ticket.ID, userID) == nil {
				if err := repositories.TicketCollaboratorRepository.Create(ctx.Tx, &models.TicketCollaborator{
					TicketID:  ticket.ID,
					UserID:    userID,
					CreatedAt: time.Now(),
				}); err != nil {
					return err
				}
			}
			if repositories.TicketWatcherRepository.TakeByTicketIDAndUserID(ctx.Tx, ticket.ID, userID) == nil {
				if err := repositories.TicketWatcherRepository.Create(ctx.Tx, &models.TicketWatcher{
					TicketID:  ticket.ID,
					UserID:    userID,
					CreatedAt: time.Now(),
				}); err != nil {
					return err
				}
			}
			if repositories.TicketMentionRepository.TakeByCommentAndUserID(ctx.Tx, ticket.ID, comment.ID, userID) == nil {
				if err := repositories.TicketMentionRepository.Create(ctx.Tx, &models.TicketMention{
					TicketID:        ticket.ID,
					CommentID:       comment.ID,
					MentionedUserID: userID,
					CreatedAt:       time.Now(),
				}); err != nil {
					return err
				}
			}
		}
		if err := repositories.TicketRepository.Updates(ctx.Tx, ticket.ID, map[string]any{
			"update_user_id":   operator.UserID,
			"update_user_name": operator.Username,
			"updated_at":       time.Now(),
		}); err != nil {
			return err
		}
		if err := s.logEvent(ctx.Tx, ticket.ID, enums.TicketEventTypeInternalNoted, operator, "", "", "添加内部备注", ""); err != nil {
			return err
		}
		if len(mentionedNames) > 0 {
			return s.logEvent(ctx.Tx, ticket.ID, enums.TicketEventTypeMentioned, operator, "", "", fmt.Sprintf("提及协作人：%s", strings.Join(mentionedNames, "、")), comment.Payload)
		}
		return nil
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
		if err := s.refreshTicketSLAFields(ctx.Tx, ticket.ID, now); err != nil {
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
	return sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		for _, ticketID := range ticketIDs {
			if err := s.assignTicketTx(ctx.Tx, request.AssignTicketRequest{
				TicketID: ticketID,
				ToUserID: req.ToUserID,
				ToTeamID: req.ToTeamID,
				Reason:   req.Reason,
			}, operator); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *ticketService) BatchChangeStatus(req request.BatchChangeTicketStatusRequest, operator *dto.AuthPrincipal) error {
	ticketIDs := normalizeBatchTicketIDs(req.TicketIDs)
	if len(ticketIDs) == 0 {
		return errorsx.InvalidParam("请选择工单")
	}
	return sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		for _, ticketID := range ticketIDs {
			if err := s.changeStatusTx(ctx.Tx, request.ChangeTicketStatusRequest{
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
	})
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
	return s.normalizeAssignmentTx(sqls.DB(), teamID, assigneeID)
}

func (s *ticketService) normalizeAssignmentTx(db *gorm.DB, teamID, assigneeID int64) (int64, int64, error) {
	if assigneeID > 0 {
		profile := repositories.AgentProfileRepository.FindOne(db, sqls.NewCnd().Eq("user_id", assigneeID))
		if profile == nil || profile.Status != enums.StatusOk {
			return 0, 0, errorsx.InvalidParam("目标处理人不存在")
		}
		if teamID <= 0 {
			teamID = profile.TeamID
		}
	}
	if teamID > 0 {
		team := repositories.AgentTeamRepository.Get(db, teamID)
		if team == nil || team.Status != enums.StatusOk {
			return 0, 0, errorsx.InvalidParam("处理团队不存在")
		}
	}
	return teamID, assigneeID, nil
}

func (s *ticketService) initSLAs(tx *gorm.DB, ticket *models.Ticket, now time.Time) error {
	firstResponseTarget, resolutionTarget := ticketSLATargetMinutes(ticket.Priority)
	if config := repositories.TicketSLAConfigRepository.FindOne(tx, sqls.NewCnd().Eq("priority", ticket.Priority).Eq("status", enums.StatusOk)); config != nil {
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
	return s.refreshTicketSLAFields(tx, ticket.ID, now)
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
	record := repositories.TicketSLARecordRepository.TakeByTicketIDAndType(tx, ticketID, string(enums.TicketSLATypeFirstResponse))
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
	record := repositories.TicketSLARecordRepository.TakeByTicketIDAndType(tx, ticketID, string(enums.TicketSLATypeResolution))
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
	record := repositories.TicketSLARecordRepository.TakeByTicketIDAndType(tx, ticketID, string(enums.TicketSLATypeResolution))
	if record == nil {
		return nil
	}
	values := map[string]any{
		"status":     enums.TicketSLAStatusRunning,
		"started_at": now,
		"paused_at":  nil,
		"updated_at": now,
	}
	if record.Status == enums.TicketSLAStatusCompleted || record.Status == enums.TicketSLAStatusBreached {
		values["elapsed_min"] = 0
		values["stopped_at"] = nil
		values["breached_at"] = nil
	}
	return repositories.TicketSLARecordRepository.Updates(tx, record.ID, values)
}

func (s *ticketService) completeResolutionSLA(tx *gorm.DB, ticketID int64, now time.Time) error {
	record := repositories.TicketSLARecordRepository.TakeByTicketIDAndType(tx, ticketID, string(enums.TicketSLATypeResolution))
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

func (s *ticketService) refreshTicketSLAFields(tx *gorm.DB, ticketID int64, now time.Time) error {
	firstResponse := repositories.TicketSLARecordRepository.TakeByTicketIDAndType(tx, ticketID, string(enums.TicketSLATypeFirstResponse))
	resolution := repositories.TicketSLARecordRepository.TakeByTicketIDAndType(tx, ticketID, string(enums.TicketSLATypeResolution))
	return repositories.TicketRepository.Updates(tx, ticketID, map[string]any{
		"next_reply_deadline_at": calcTicketSLADeadline(firstResponse, now),
		"resolve_deadline_at":    calcTicketSLADeadline(resolution, now),
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
			if err := s.refreshTicketSLAFields(ctx.Tx, record.TicketID, now); err != nil {
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

func activeTicketStatuses() []enums.TicketStatus {
	return []enums.TicketStatus{
		enums.TicketStatusNew,
		enums.TicketStatusOpen,
		enums.TicketStatusPendingCustomer,
		enums.TicketStatusPendingInternal,
	}
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

func parseTicketNotePayload(value string) ticketNotePayload {
	value = strings.TrimSpace(value)
	if value == "" {
		return ticketNotePayload{}
	}
	var payload ticketNotePayload
	if err := json.Unmarshal([]byte(value), &payload); err != nil {
		return ticketNotePayload{}
	}
	return payload
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

func calcTicketSLADeadline(record *models.TicketSLARecord, now time.Time) *time.Time {
	if record == nil {
		return nil
	}
	switch record.Status {
	case enums.TicketSLAStatusCompleted:
		return nil
	case enums.TicketSLAStatusBreached:
		if record.BreachedAt != nil && !record.BreachedAt.IsZero() {
			value := *record.BreachedAt
			return &value
		}
	case enums.TicketSLAStatusPaused:
		base := now
		if record.PausedAt != nil && !record.PausedAt.IsZero() {
			base = *record.PausedAt
		}
		remaining := record.TargetMinutes - record.ElapsedMin
		if remaining < 0 {
			remaining = 0
		}
		deadline := base.Add(time.Duration(remaining) * time.Minute)
		return &deadline
	}
	if record.StartedAt == nil || record.StartedAt.IsZero() {
		return nil
	}
	remaining := record.TargetMinutes - record.ElapsedMin
	deadline := record.StartedAt.Add(time.Duration(remaining) * time.Minute)
	return &deadline
}
