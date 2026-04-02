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
	results := make([]response.TicketResponse, 0, len(list))
	for i := range list {
		if item := BuildTicket(&list[i]); item != nil {
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
		Overdue:         summary.Overdue,
	}
}

func BuildTicketWatcherList(list []models.TicketWatcher) []response.TicketWatcherResponse {
	if len(list) == 0 {
		return nil
	}
	results := make([]response.TicketWatcherResponse, 0, len(list))
	for i := range list {
		item := &list[i]
		out := response.TicketWatcherResponse{
			ID:     item.ID,
			UserID: item.UserID,
		}
		if user := services.UserService.Get(item.UserID); user != nil {
			out.UserName = user.Nickname
			if out.UserName == "" {
				out.UserName = user.Username
			}
		}
		results = append(results, out)
	}
	return results
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
