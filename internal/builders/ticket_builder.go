package builders

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/utils"
	"cs-agent/internal/services"

	"github.com/mlogclub/simple/sqls"
)

func BuildTicket(item *models.Ticket) *response.TicketResponse {
	if item == nil {
		return nil
	}
	ret := &response.TicketResponse{
		ID:                  item.ID,
		TicketNo:            item.TicketNo,
		Title:               item.Title,
		Description:         item.Description,
		Source:              item.Source,
		Channel:             item.Channel,
		CustomerID:          item.CustomerID,
		ConversationID:      item.ConversationID,
		CategoryID:          item.CategoryID,
		Type:                item.Type,
		Priority:            item.Priority,
		Severity:            item.Severity,
		Status:              item.Status,
		CurrentTeamID:       item.CurrentTeamID,
		CurrentAssigneeID:   item.CurrentAssigneeID,
		PendingReason:       item.PendingReason,
		CloseReason:         item.CloseReason,
		ResolutionCode:      item.ResolutionCode,
		ResolutionSummary:   item.ResolutionSummary,
		FirstResponseAt:     utils.FormatTimePtr(item.FirstResponseAt),
		ResolvedAt:          utils.FormatTimePtr(item.ResolvedAt),
		ClosedAt:            utils.FormatTimePtr(item.ClosedAt),
		DueAt:               utils.FormatTimePtr(item.DueAt),
		NextReplyDeadlineAt: utils.FormatTimePtr(item.NextReplyDeadlineAt),
		ResolveDeadlineAt:   utils.FormatTimePtr(item.ResolveDeadlineAt),
		ReopenedCount:       item.ReopenedCount,
		CreatedAt:           utils.FormatTime(item.CreatedAt),
		UpdatedAt:           utils.FormatTime(item.UpdatedAt),
	}
	if item.CurrentAssigneeID > 0 {
		if user := services.UserService.Get(item.CurrentAssigneeID); user != nil {
			ret.CurrentAssigneeName = user.Nickname
			if ret.CurrentAssigneeName == "" {
				ret.CurrentAssigneeName = user.Username
			}
		}
	}
	if item.CurrentTeamID > 0 {
		if team := services.AgentTeamService.Get(item.CurrentTeamID); team != nil {
			ret.CurrentTeamName = team.Name
		}
	}
	if item.CustomerID > 0 {
		ret.Customer = BuildCustomer(services.CustomerService.Get(item.CustomerID))
	}
	ret.SLA = BuildTicketSLAList(
		services.TicketSLARecordService.Find(
			sqls.NewCnd().Eq("ticket_id", item.ID).Asc("id"),
		),
	)
	return ret
}

func BuildTicketList(list []models.Ticket) []response.TicketResponse {
	if len(list) == 0 {
		return nil
	}
	userMap, teamMap, customerMap, slaMap := loadTicketListContext(list)
	results := make([]response.TicketResponse, 0, len(list))
	for i := range list {
		if item := buildTicketWithContext(&list[i], userMap, teamMap, customerMap, slaMap); item != nil {
			results = append(results, *item)
		}
	}
	return results
}

func BuildTicketComment(item *models.TicketComment) *response.TicketCommentResponse {
	if item == nil {
		return nil
	}
	ret := &response.TicketCommentResponse{
		ID:          item.ID,
		TicketID:    item.TicketID,
		CommentType: item.CommentType,
		AuthorType:  item.AuthorType,
		AuthorID:    item.AuthorID,
		ContentType: item.ContentType,
		Content:     item.Content,
		Payload:     item.Payload,
		CreatedAt:   utils.FormatTime(item.CreatedAt),
	}
	ret.AuthorName = buildTicketOperatorName(item.AuthorType, item.AuthorID)
	return ret
}

func BuildTicketCommentList(list []models.TicketComment) []response.TicketCommentResponse {
	if len(list) == 0 {
		return nil
	}
	results := make([]response.TicketCommentResponse, 0, len(list))
	for i := range list {
		if item := BuildTicketComment(&list[i]); item != nil {
			results = append(results, *item)
		}
	}
	return results
}

func BuildTicketEventLog(item *models.TicketEventLog) *response.TicketEventLogResponse {
	if item == nil {
		return nil
	}
	return &response.TicketEventLogResponse{
		ID:           item.ID,
		TicketID:     item.TicketID,
		EventType:    item.EventType,
		OperatorType: item.OperatorType,
		OperatorID:   item.OperatorID,
		OperatorName: buildTicketOperatorName(item.OperatorType, item.OperatorID),
		OldValue:     item.OldValue,
		NewValue:     item.NewValue,
		Content:      item.Content,
		Payload:      item.Payload,
		CreatedAt:    utils.FormatTime(item.CreatedAt),
	}
}

func BuildTicketEventLogList(list []models.TicketEventLog) []response.TicketEventLogResponse {
	if len(list) == 0 {
		return nil
	}
	results := make([]response.TicketEventLogResponse, 0, len(list))
	for i := range list {
		if item := BuildTicketEventLog(&list[i]); item != nil {
			results = append(results, *item)
		}
	}
	return results
}

func BuildTicketSLAList(list []models.TicketSLARecord) []response.TicketSLAResponse {
	if len(list) == 0 {
		return nil
	}
	results := make([]response.TicketSLAResponse, 0, len(list))
	for i := range list {
		item := &list[i]
		results = append(results, response.TicketSLAResponse{
			SLAType:       item.SLAType,
			TargetMinutes: item.TargetMinutes,
			Status:        item.Status,
			StartedAt:     utils.FormatTimePtr(item.StartedAt),
			PausedAt:      utils.FormatTimePtr(item.PausedAt),
			StoppedAt:     utils.FormatTimePtr(item.StoppedAt),
			BreachedAt:    utils.FormatTimePtr(item.BreachedAt),
			ElapsedMin:    item.ElapsedMin,
		})
	}
	return results
}

func BuildTicketDetail(aggregate *services.TicketDetailAggregate) *response.TicketDetailResponse {
	if aggregate == nil || aggregate.Ticket == nil {
		return nil
	}
	ret := &response.TicketDetailResponse{
		Ticket:   *BuildTicket(aggregate.Ticket),
		Watchers: BuildTicketWatcherList(aggregate.Watchers),
		Comments: BuildTicketCommentList(aggregate.Comments),
		Events:   BuildTicketEventLogList(aggregate.Events),
	}
	if aggregate.Customer != nil {
		ret.Ticket.Customer = BuildCustomer(aggregate.Customer)
	}
	ret.Ticket.SLA = BuildTicketSLAList(aggregate.SLAs)
	return ret
}

func BuildTicketSummary(summary *services.TicketSummaryAggregate) *response.TicketSummaryResponse {
	if summary == nil {
		return nil
	}
	return &response.TicketSummaryResponse{
		All:             summary.All,
		Mine:            summary.Mine,
		Watching:        summary.Watching,
		PendingCustomer: summary.PendingCustomer,
		DueSoon:         summary.DueSoon,
		Overdue:         summary.Overdue,
	}
}

func BuildTicketWatcherList(list []models.TicketWatcher) []response.TicketWatcherResponse {
	if len(list) == 0 {
		return nil
	}
	userIDs := make([]int64, 0, len(list))
	for i := range list {
		if list[i].UserID > 0 {
			userIDs = append(userIDs, list[i].UserID)
		}
	}
	userMap := buildUserMap(services.UserService.FindByIds(userIDs))
	results := make([]response.TicketWatcherResponse, 0, len(list))
	for i := range list {
		item := &list[i]
		out := response.TicketWatcherResponse{
			ID:     item.ID,
			UserID: item.UserID,
		}
		if user := userMap[item.UserID]; user != nil {
			out.UserName = user.Nickname
			if out.UserName == "" {
				out.UserName = user.Username
			}
		}
		results = append(results, out)
	}
	return results
}

func buildTicketWithContext(item *models.Ticket, userMap map[int64]*models.User, teamMap map[int64]*models.AgentTeam, customerMap map[int64]*models.Customer, slaMap map[int64][]models.TicketSLARecord) *response.TicketResponse {
	if item == nil {
		return nil
	}
	ret := &response.TicketResponse{
		ID:                  item.ID,
		TicketNo:            item.TicketNo,
		Title:               item.Title,
		Description:         item.Description,
		Source:              item.Source,
		Channel:             item.Channel,
		CustomerID:          item.CustomerID,
		ConversationID:      item.ConversationID,
		CategoryID:          item.CategoryID,
		Type:                item.Type,
		Priority:            item.Priority,
		Severity:            item.Severity,
		Status:              item.Status,
		CurrentTeamID:       item.CurrentTeamID,
		CurrentAssigneeID:   item.CurrentAssigneeID,
		PendingReason:       item.PendingReason,
		CloseReason:         item.CloseReason,
		ResolutionCode:      item.ResolutionCode,
		ResolutionSummary:   item.ResolutionSummary,
		FirstResponseAt:     utils.FormatTimePtr(item.FirstResponseAt),
		ResolvedAt:          utils.FormatTimePtr(item.ResolvedAt),
		ClosedAt:            utils.FormatTimePtr(item.ClosedAt),
		DueAt:               utils.FormatTimePtr(item.DueAt),
		NextReplyDeadlineAt: utils.FormatTimePtr(item.NextReplyDeadlineAt),
		ResolveDeadlineAt:   utils.FormatTimePtr(item.ResolveDeadlineAt),
		ReopenedCount:       item.ReopenedCount,
		CreatedAt:           utils.FormatTime(item.CreatedAt),
		UpdatedAt:           utils.FormatTime(item.UpdatedAt),
	}
	if user := userMap[item.CurrentAssigneeID]; user != nil {
		ret.CurrentAssigneeName = user.Nickname
		if ret.CurrentAssigneeName == "" {
			ret.CurrentAssigneeName = user.Username
		}
	}
	if team := teamMap[item.CurrentTeamID]; team != nil {
		ret.CurrentTeamName = team.Name
	}
	if customer := customerMap[item.CustomerID]; customer != nil {
		ret.Customer = BuildCustomer(customer)
	}
	ret.SLA = BuildTicketSLAList(slaMap[item.ID])
	return ret
}

func loadTicketListContext(list []models.Ticket) (map[int64]*models.User, map[int64]*models.AgentTeam, map[int64]*models.Customer, map[int64][]models.TicketSLARecord) {
	userIDs := make([]int64, 0, len(list))
	teamIDs := make([]int64, 0, len(list))
	customerIDs := make([]int64, 0, len(list))
	ticketIDs := make([]int64, 0, len(list))
	for i := range list {
		item := &list[i]
		if item.CurrentAssigneeID > 0 {
			userIDs = append(userIDs, item.CurrentAssigneeID)
		}
		if item.CurrentTeamID > 0 {
			teamIDs = append(teamIDs, item.CurrentTeamID)
		}
		if item.CustomerID > 0 {
			customerIDs = append(customerIDs, item.CustomerID)
		}
		if item.ID > 0 {
			ticketIDs = append(ticketIDs, item.ID)
		}
	}
	userMap := buildUserMap(services.UserService.FindByIds(userIDs))
	teamMap := buildTeamMap(services.AgentTeamService.FindByIds(teamIDs))
	customerMap := map[int64]*models.Customer{}
	if len(customerIDs) > 0 {
		customerMap = buildCustomerMap(services.CustomerService.Find(sqls.NewCnd().In("id", customerIDs)))
	}
	slaMap := map[int64][]models.TicketSLARecord{}
	if len(ticketIDs) > 0 {
		slaMap = buildTicketSLAMap(services.TicketSLARecordService.Find(sqls.NewCnd().In("ticket_id", ticketIDs).Asc("id")))
	}
	return userMap, teamMap, customerMap, slaMap
}

func buildUserMap(list []models.User) map[int64]*models.User {
	ret := make(map[int64]*models.User, len(list))
	for i := range list {
		item := &list[i]
		ret[item.ID] = item
	}
	return ret
}

func buildTeamMap(list []models.AgentTeam) map[int64]*models.AgentTeam {
	ret := make(map[int64]*models.AgentTeam, len(list))
	for i := range list {
		item := &list[i]
		ret[item.ID] = item
	}
	return ret
}

func buildCustomerMap(list []models.Customer) map[int64]*models.Customer {
	ret := make(map[int64]*models.Customer, len(list))
	for i := range list {
		item := &list[i]
		ret[item.ID] = item
	}
	return ret
}

func buildTicketSLAMap(list []models.TicketSLARecord) map[int64][]models.TicketSLARecord {
	ret := make(map[int64][]models.TicketSLARecord)
	for i := range list {
		item := list[i]
		ret[item.TicketID] = append(ret[item.TicketID], item)
	}
	return ret
}

func buildTicketOperatorName(senderType enums.IMSenderType, senderID int64) string {
	if senderID <= 0 {
		return ""
	}
	switch senderType {
	case enums.IMSenderTypeAgent:
		if user := services.UserService.Get(senderID); user != nil {
			if user.Nickname != "" {
				return user.Nickname
			}
			return user.Username
		}
	case enums.IMSenderTypeCustomer:
		if customer := services.CustomerService.Get(senderID); customer != nil {
			return customer.Name
		}
	case enums.IMSenderTypeAI:
		if ai := services.AIAgentService.Get(senderID); ai != nil {
			return ai.Name
		}
	}
	return ""
}
