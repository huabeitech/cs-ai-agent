package services

import (
	"context"
	"strings"
	"time"

	"cs-agent/internal/events"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/errorsx"
	"cs-agent/internal/pkg/eventbus"
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
	Ticket     *models.Ticket
	Tags       []models.Tag
	Customer   *models.Customer
	Progresses []models.TicketProgress
	Users      map[int64]*models.User
}

type TicketSummaryAggregate struct {
	All        int64
	Pending    int64
	InProgress int64
	Done       int64
	Unassigned int64
	Mine       int64
	Stale      int64
}

type TicketListAggregate struct {
	List           []models.Ticket
	Paging         *sqls.Paging
	TagsByTicketID map[int64][]models.Tag
	Users          map[int64]*models.User
	Customers      map[int64]*models.Customer
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

func (s *ticketService) FindPageAggregateByCnd(cnd *sqls.Cnd, _ int64) (*TicketListAggregate, error) {
	list, paging := repositories.TicketRepository.FindPageByCnd(sqls.DB(), cnd)
	return s.buildTicketListAggregate(sqls.DB(), list, paging), nil
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

func (s *ticketService) GetTags(ticketID int64) []models.Tag {
	if ticketID <= 0 {
		return nil
	}
	relations := TicketTagService.Find(sqls.NewCnd().Eq("ticket_id", ticketID).Asc("id"))
	if len(relations) == 0 {
		return nil
	}
	tagIDs := make([]int64, 0, len(relations))
	for i := range relations {
		tagIDs = append(tagIDs, relations[i].TagID)
	}
	tags := repositories.TagRepository.Find(sqls.DB(), sqls.NewCnd().In("id", tagIDs))
	if len(tags) <= 1 {
		return tags
	}
	tagMap := make(map[int64]models.Tag, len(tags))
	for i := range tags {
		tagMap[tags[i].ID] = tags[i]
	}
	ordered := make([]models.Tag, 0, len(relations))
	for _, tagID := range tagIDs {
		if tag, ok := tagMap[tagID]; ok {
			ordered = append(ordered, tag)
		}
	}
	return ordered
}

func (s *ticketService) CreateTicket(req request.CreateTicketRequest, operator *dto.AuthPrincipal) (*models.Ticket, error) {
	if operator == nil {
		return nil, errorsx.Unauthorized("未登录或登录已过期")
	}
	title := strings.TrimSpace(req.Title)
	description := strings.TrimSpace(req.Description)
	if title == "" {
		return nil, errorsx.InvalidParam("工单标题不能为空")
	}
	if description == "" {
		return nil, errorsx.InvalidParam("工单描述不能为空")
	}
	source := enums.TicketSource(strings.TrimSpace(req.Source))
	if source == "" {
		source = enums.TicketSourceManual
	}
	if !enums.IsValidTicketSource(string(source)) {
		return nil, errorsx.InvalidParam("工单来源不合法")
	}
	if err := s.validateTicketRefs(req.CustomerID, req.ConversationID, req.CurrentAssigneeID); err != nil {
		return nil, err
	}
	tagIDs, err := TicketTagService.ValidateTagIDs(req.TagIDs)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	ticket := &models.Ticket{
		Title:             title,
		Description:       description,
		Source:            source,
		Channel:           strings.TrimSpace(req.Channel),
		CustomerID:        req.CustomerID,
		ConversationID:    req.ConversationID,
		Status:            enums.TicketStatusPending,
		CurrentAssigneeID: req.CurrentAssigneeID,
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
		if err := TicketTagService.ReplaceTicketTags(ctx.Tx, ticket.ID, tagIDs, operator); err != nil {
			return err
		}
		return repositories.TicketProgressRepository.Create(ctx.Tx, &models.TicketProgress{
			TicketID:  ticket.ID,
			Content:   "创建工单",
			AuthorID:  operator.UserID,
			CreatedAt: now,
		})
	}); err != nil {
		return nil, err
	}

	eventbus.PublishAsync(context.Background(), events.TicketCreatedEvent{
		TicketID:   ticket.ID,
		OperatorID: operator.UserID,
	})
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
		title = strings.TrimSpace(ConversationService.BuildConversationSummary(conversation))
	}
	if title == "" {
		title = "会话工单"
	}
	description := strings.TrimSpace(req.Description)
	if description == "" {
		description = strings.TrimSpace(conversation.LastMessageSummary)
	}
	if description == "" {
		description = title
	}
	return s.CreateTicket(request.CreateTicketRequest{
		Title:             title,
		Description:       description,
		Source:            string(enums.TicketSourceConversation),
		Channel:           s.resolveConversationChannel(conversation),
		CustomerID:        conversation.CustomerID,
		ConversationID:    conversation.ID,
		TagIDs:            req.TagIDs,
		CurrentAssigneeID: req.CurrentAssigneeID,
	}, operator)
}

func (s *ticketService) UpdateTicket(req request.UpdateTicketRequest, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	title := strings.TrimSpace(req.Title)
	description := strings.TrimSpace(req.Description)
	if title == "" {
		return errorsx.InvalidParam("工单标题不能为空")
	}
	if description == "" {
		return errorsx.InvalidParam("工单描述不能为空")
	}
	ticket := s.Get(req.TicketID)
	if ticket == nil {
		return errorsx.InvalidParam("工单不存在")
	}
	if err := s.validateAssignee(req.CurrentAssigneeID); err != nil {
		return err
	}
	tagIDs, err := TicketTagService.ValidateTagIDs(req.TagIDs)
	if err != nil {
		return err
	}
	now := time.Now()
	return sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		if err := repositories.TicketRepository.Updates(ctx.Tx, ticket.ID, map[string]any{
			"title":               title,
			"description":         description,
			"current_assignee_id": req.CurrentAssigneeID,
			"updated_at":          now,
			"update_user_id":      operator.UserID,
			"update_user_name":    operator.Username,
		}); err != nil {
			return err
		}
		return TicketTagService.ReplaceTicketTags(ctx.Tx, ticket.ID, tagIDs, operator)
	})
}

func (s *ticketService) AssignTicket(req request.AssignTicketRequest, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	var assignedEvent *events.TicketAssignedEvent
	if err := sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		event, err := s.assignTicketTx(ctx.Tx, req, operator)
		if err != nil {
			return err
		}
		assignedEvent = event
		return nil
	}); err != nil {
		return err
	}
	if assignedEvent != nil {
		eventbus.PublishAsync(context.Background(), *assignedEvent)
	}
	return nil
}

func (s *ticketService) BatchAssignTickets(req request.BatchAssignTicketRequest, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	ticketIDs := normalizeInt64IDs(req.TicketIDs)
	if len(ticketIDs) == 0 {
		return errorsx.InvalidParam("请选择工单")
	}
	assignedEvents := make([]events.TicketAssignedEvent, 0, len(ticketIDs))
	if err := sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		for _, ticketID := range ticketIDs {
			event, err := s.assignTicketTx(ctx.Tx, request.AssignTicketRequest{
				TicketID: ticketID,
				ToUserID: req.ToUserID,
				Reason:   req.Reason,
			}, operator)
			if err != nil {
				return err
			}
			if event != nil {
				assignedEvents = append(assignedEvents, *event)
			}
		}
		return nil
	}); err != nil {
		return err
	}
	for i := range assignedEvents {
		eventbus.PublishAsync(context.Background(), assignedEvents[i])
	}
	return nil
}

func (s *ticketService) ChangeStatus(req request.ChangeTicketStatusRequest, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	status := strings.TrimSpace(req.Status)
	if !enums.IsValidTicketStatus(status) {
		return errorsx.InvalidParam("工单状态不合法")
	}
	ticket := s.Get(req.TicketID)
	if ticket == nil {
		return errorsx.InvalidParam("工单不存在")
	}
	now := time.Now()
	var handledAt *time.Time
	if enums.TicketStatus(status) == enums.TicketStatusDone {
		handledAt = &now
	}
	return repositories.TicketRepository.Updates(sqls.DB(), ticket.ID, map[string]any{
		"status":           enums.TicketStatus(status),
		"handled_at":       handledAt,
		"updated_at":       now,
		"update_user_id":   operator.UserID,
		"update_user_name": operator.Username,
	})
}

func (s *ticketService) AddProgress(req request.CreateTicketProgressRequest, operator *dto.AuthPrincipal) (*models.TicketProgress, error) {
	if operator == nil {
		return nil, errorsx.Unauthorized("未登录或登录已过期")
	}
	content := strings.TrimSpace(req.Content)
	if content == "" {
		return nil, errorsx.InvalidParam("处理进展不能为空")
	}
	ticket := s.Get(req.TicketID)
	if ticket == nil {
		return nil, errorsx.InvalidParam("工单不存在")
	}
	now := time.Now()
	progress := &models.TicketProgress{
		TicketID:  ticket.ID,
		Content:   content,
		AuthorID:  operator.UserID,
		CreatedAt: now,
	}
	if err := sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		if err := repositories.TicketProgressRepository.Create(ctx.Tx, progress); err != nil {
			return err
		}
		return repositories.TicketRepository.Updates(ctx.Tx, ticket.ID, map[string]any{
			"updated_at":       now,
			"update_user_id":   operator.UserID,
			"update_user_name": operator.Username,
		})
	}); err != nil {
		return nil, err
	}
	return progress, nil
}

func (s *ticketService) GetDetail(id int64) (*TicketDetailAggregate, error) {
	ticket := s.Get(id)
	if ticket == nil {
		return nil, errorsx.InvalidParam("工单不存在")
	}
	aggregate := &TicketDetailAggregate{
		Ticket:     ticket,
		Tags:       s.GetTags(id),
		Progresses: repositories.TicketProgressRepository.Find(sqls.DB(), sqls.NewCnd().Eq("ticket_id", id).Asc("id")),
		Users:      make(map[int64]*models.User),
	}
	if ticket.CustomerID > 0 {
		aggregate.Customer = CustomerService.Get(ticket.CustomerID)
	}
	userIDs := make([]int64, 0)
	seen := make(map[int64]struct{})
	addUserID := func(userID int64) {
		if userID <= 0 {
			return
		}
		if _, ok := seen[userID]; ok {
			return
		}
		seen[userID] = struct{}{}
		userIDs = append(userIDs, userID)
	}
	addUserID(ticket.CurrentAssigneeID)
	for i := range aggregate.Progresses {
		addUserID(aggregate.Progresses[i].AuthorID)
	}
	if len(userIDs) > 0 {
		users := repositories.UserRepository.FindByIds(sqls.DB(), userIDs)
		for i := range users {
			item := users[i]
			aggregate.Users[item.ID] = &item
		}
	}
	return aggregate, nil
}

func (s *ticketService) GetSummary(operator *dto.AuthPrincipal, staleHours ...int) *TicketSummaryAggregate {
	staleHour := 24
	if len(staleHours) > 0 && staleHours[0] > 0 {
		staleHour = staleHours[0]
	}
	summary := &TicketSummaryAggregate{
		All:        s.Count(sqls.NewCnd()),
		Pending:    s.Count(sqls.NewCnd().Eq("status", enums.TicketStatusPending)),
		InProgress: s.Count(sqls.NewCnd().Eq("status", enums.TicketStatusInProgress)),
		Done:       s.Count(sqls.NewCnd().Eq("status", enums.TicketStatusDone)),
		Unassigned: s.Count(sqls.NewCnd().Eq("current_assignee_id", 0)),
		Stale: s.Count(
			sqls.NewCnd().
				NotEq("status", enums.TicketStatusDone).
				Where("updated_at < ?", time.Now().Add(-time.Duration(staleHour)*time.Hour)),
		),
	}
	if operator != nil {
		summary.Mine = s.Count(sqls.NewCnd().Eq("current_assignee_id", operator.UserID))
	}
	return summary
}

func (s *ticketService) assignTicketTx(tx *gorm.DB, req request.AssignTicketRequest, operator *dto.AuthPrincipal) (*events.TicketAssignedEvent, error) {
	ticket := repositories.TicketRepository.Get(tx, req.TicketID)
	if ticket == nil {
		return nil, errorsx.InvalidParam("工单不存在")
	}
	if err := s.validateRequiredAssignee(req.ToUserID); err != nil {
		return nil, err
	}
	now := time.Now()
	if err := repositories.TicketRepository.Updates(tx, ticket.ID, map[string]any{
		"current_assignee_id": req.ToUserID,
		"updated_at":          now,
		"update_user_id":      operator.UserID,
		"update_user_name":    operator.Username,
	}); err != nil {
		return nil, err
	}
	return &events.TicketAssignedEvent{
		TicketID:   ticket.ID,
		FromUserID: ticket.CurrentAssigneeID,
		ToUserID:   req.ToUserID,
		OperatorID: operator.UserID,
		Reason:     strings.TrimSpace(req.Reason),
	}, nil
}

func (s *ticketService) buildTicketListAggregate(db *gorm.DB, list []models.Ticket, paging *sqls.Paging) *TicketListAggregate {
	aggregate := &TicketListAggregate{
		List:           list,
		Paging:         paging,
		TagsByTicketID: make(map[int64][]models.Tag),
		Users:          make(map[int64]*models.User),
		Customers:      make(map[int64]*models.Customer),
	}
	if len(list) == 0 {
		return aggregate
	}
	ticketIDs := make([]int64, 0, len(list))
	customerIDs := make([]int64, 0)
	userIDs := make([]int64, 0)
	ticketSeen := make(map[int64]struct{})
	customerSeen := make(map[int64]struct{})
	userSeen := make(map[int64]struct{})
	for i := range list {
		item := &list[i]
		if _, ok := ticketSeen[item.ID]; !ok {
			ticketSeen[item.ID] = struct{}{}
			ticketIDs = append(ticketIDs, item.ID)
		}
		if item.CustomerID > 0 {
			if _, ok := customerSeen[item.CustomerID]; !ok {
				customerSeen[item.CustomerID] = struct{}{}
				customerIDs = append(customerIDs, item.CustomerID)
			}
		}
		if item.CurrentAssigneeID > 0 {
			if _, ok := userSeen[item.CurrentAssigneeID]; !ok {
				userSeen[item.CurrentAssigneeID] = struct{}{}
				userIDs = append(userIDs, item.CurrentAssigneeID)
			}
		}
	}
	s.enrichTicketTags(db, aggregate, ticketIDs)
	if len(userIDs) > 0 {
		users := repositories.UserRepository.FindByIds(db, userIDs)
		for i := range users {
			item := users[i]
			aggregate.Users[item.ID] = &item
		}
	}
	if len(customerIDs) > 0 {
		customers := repositories.CustomerRepository.Find(db, sqls.NewCnd().In("id", customerIDs))
		for i := range customers {
			item := customers[i]
			aggregate.Customers[item.ID] = &item
		}
	}
	return aggregate
}

func (s *ticketService) enrichTicketTags(db *gorm.DB, aggregate *TicketListAggregate, ticketIDs []int64) {
	if len(ticketIDs) == 0 {
		return
	}
	ticketTags := repositories.TicketTagRepository.Find(db, sqls.NewCnd().In("ticket_id", ticketIDs).Asc("id"))
	if len(ticketTags) == 0 {
		return
	}
	tagIDs := make([]int64, 0)
	tagSeen := make(map[int64]struct{})
	ticketTagMap := make(map[int64][]int64, len(ticketIDs))
	for i := range ticketTags {
		relation := ticketTags[i]
		ticketTagMap[relation.TicketID] = append(ticketTagMap[relation.TicketID], relation.TagID)
		if _, ok := tagSeen[relation.TagID]; !ok {
			tagSeen[relation.TagID] = struct{}{}
			tagIDs = append(tagIDs, relation.TagID)
		}
	}
	tags := repositories.TagRepository.Find(db, sqls.NewCnd().In("id", tagIDs))
	tagMap := make(map[int64]models.Tag, len(tags))
	for i := range tags {
		tagMap[tags[i].ID] = tags[i]
	}
	for ticketID, orderedTagIDs := range ticketTagMap {
		orderedTags := make([]models.Tag, 0, len(orderedTagIDs))
		for _, tagID := range orderedTagIDs {
			if tag, ok := tagMap[tagID]; ok {
				orderedTags = append(orderedTags, tag)
			}
		}
		aggregate.TagsByTicketID[ticketID] = orderedTags
	}
}

func (s *ticketService) validateTicketRefs(customerID, conversationID, assigneeID int64) error {
	if customerID > 0 && CustomerService.Get(customerID) == nil {
		return errorsx.InvalidParam("客户不存在")
	}
	if conversationID > 0 && ConversationService.Get(conversationID) == nil {
		return errorsx.InvalidParam("会话不存在")
	}
	return s.validateAssignee(assigneeID)
}

func (s *ticketService) validateAssignee(userID int64) error {
	if userID <= 0 {
		return nil
	}
	return s.validateRequiredAssignee(userID)
}

func (s *ticketService) validateRequiredAssignee(userID int64) error {
	if userID <= 0 {
		return errorsx.InvalidParam("负责人不存在")
	}
	user := UserService.Get(userID)
	if user == nil || user.Status == enums.StatusDeleted {
		return errorsx.InvalidParam("负责人不存在")
	}
	return nil
}

func (s *ticketService) resolveConversationChannel(conversation *models.Conversation) string {
	if conversation == nil || conversation.ChannelID <= 0 {
		return ""
	}
	if channel := ChannelService.Get(conversation.ChannelID); channel != nil {
		return channel.ChannelType
	}
	return ""
}

func normalizeInt64IDs(ids []int64) []int64 {
	if len(ids) == 0 {
		return nil
	}
	seen := make(map[int64]struct{}, len(ids))
	result := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}
