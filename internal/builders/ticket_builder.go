package builders

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/utils"
	"cs-agent/internal/services"
	"encoding/json"
	"strings"

	"github.com/mlogclub/simple/sqls"
)

type TicketBuildContext struct {
	Categories       map[int64]*models.TicketCategory
	ResolutionCodes  map[string]*models.TicketResolutionCode
	Users            map[int64]*models.User
	Teams            map[int64]*models.AgentTeam
	Customers        map[int64]*models.Customer
	SLAByTicketID    map[int64][]models.TicketSLARecord
	WatchedTicketIDs map[int64]struct{}
}

type TicketDetailBuildContext struct {
	Users          map[int64]*models.User
	Teams          map[int64]*models.AgentTeam
	AgentProfiles  map[int64]*models.AgentProfile
	RelatedTickets map[int64]*models.Ticket
	Customers      map[int64]*models.Customer
	AIAgents       map[int64]*models.AIAgent
}

func BuildTicket(item *models.Ticket) *response.TicketResponse {
	return BuildTicketWithContext(item, nil)
}

func BuildTicketWithContext(item *models.Ticket, ctx *TicketBuildContext) *response.TicketResponse {
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
	if ctx != nil {
		if _, ok := ctx.WatchedTicketIDs[item.ID]; ok {
			ret.WatchedByMe = true
		}
	}
	if item.CategoryID > 0 && ctx != nil && ctx.Categories != nil {
		if category := ctx.Categories[item.CategoryID]; category != nil {
			ret.CategoryName = category.Name
		}
	} else if item.CategoryID > 0 {
		if category := services.TicketCategoryService.Get(item.CategoryID); category != nil {
			ret.CategoryName = category.Name
		}
	}
	if item.ResolutionCode != "" && ctx != nil && ctx.ResolutionCodes != nil {
		if code := ctx.ResolutionCodes[item.ResolutionCode]; code != nil {
			ret.ResolutionCodeName = code.Name
		}
	} else if item.ResolutionCode != "" {
		if code := services.TicketResolutionCodeService.Take("code = ? AND status <> ?", item.ResolutionCode, enums.StatusDeleted); code != nil {
			ret.ResolutionCodeName = code.Name
		}
	}
	if item.CurrentAssigneeID > 0 && ctx != nil && ctx.Users != nil {
		if user := ctx.Users[item.CurrentAssigneeID]; user != nil {
			ret.CurrentAssigneeName = user.Nickname
			if ret.CurrentAssigneeName == "" {
				ret.CurrentAssigneeName = user.Username
			}
		}
	} else if item.CurrentAssigneeID > 0 {
		if user := services.UserService.Get(item.CurrentAssigneeID); user != nil {
			ret.CurrentAssigneeName = user.Nickname
			if ret.CurrentAssigneeName == "" {
				ret.CurrentAssigneeName = user.Username
			}
		}
	}
	if item.CurrentTeamID > 0 && ctx != nil && ctx.Teams != nil {
		if team := ctx.Teams[item.CurrentTeamID]; team != nil {
			ret.CurrentTeamName = team.Name
		}
	} else if item.CurrentTeamID > 0 {
		if team := services.AgentTeamService.Get(item.CurrentTeamID); team != nil {
			ret.CurrentTeamName = team.Name
		}
	}
	if item.CustomerID > 0 && ctx != nil && ctx.Customers != nil {
		ret.Customer = BuildCustomer(ctx.Customers[item.CustomerID])
	} else if item.CustomerID > 0 {
		ret.Customer = BuildCustomer(services.CustomerService.Get(item.CustomerID))
	}
	if ctx != nil && ctx.SLAByTicketID != nil {
		ret.SLA = BuildTicketSLAList(ctx.SLAByTicketID[item.ID])
	} else {
		ret.SLA = BuildTicketSLAList(
			services.TicketSLARecordService.Find(
				sqls.NewCnd().Eq("ticket_id", item.ID).Asc("id"),
			),
		)
	}
	return ret
}

func BuildTicketList(list []models.Ticket) []response.TicketResponse {
	return BuildTicketListWithContext(list, nil)
}

func BuildTicketListWithContext(list []models.Ticket, ctx *TicketBuildContext) []response.TicketResponse {
	if len(list) == 0 {
		return nil
	}
	results := make([]response.TicketResponse, 0, len(list))
	for i := range list {
		if item := BuildTicketWithContext(&list[i], ctx); item != nil {
			results = append(results, *item)
		}
	}
	return results
}

func BuildTicketComment(item *models.TicketComment) *response.TicketCommentResponse {
	return BuildTicketCommentWithContext(item, nil)
}

func BuildTicketCommentWithContext(item *models.TicketComment, ctx *TicketDetailBuildContext) *response.TicketCommentResponse {
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
	ret.AuthorName = buildTicketOperatorName(item.AuthorType, item.AuthorID, ctx)
	return ret
}

func BuildTicketCommentList(list []models.TicketComment) []response.TicketCommentResponse {
	return BuildTicketCommentListWithContext(list, nil)
}

func BuildTicketCommentListWithContext(list []models.TicketComment, ctx *TicketDetailBuildContext) []response.TicketCommentResponse {
	if len(list) == 0 {
		return nil
	}
	results := make([]response.TicketCommentResponse, 0, len(list))
	for i := range list {
		if item := BuildTicketCommentWithContext(&list[i], ctx); item != nil {
			results = append(results, *item)
		}
	}
	return results
}

func BuildTicketEventLog(item *models.TicketEventLog) *response.TicketEventLogResponse {
	return BuildTicketEventLogWithContext(item, nil)
}

func BuildTicketEventLogWithContext(item *models.TicketEventLog, ctx *TicketDetailBuildContext) *response.TicketEventLogResponse {
	if item == nil {
		return nil
	}
	return &response.TicketEventLogResponse{
		ID:           item.ID,
		TicketID:     item.TicketID,
		EventType:    item.EventType,
		OperatorType: item.OperatorType,
		OperatorID:   item.OperatorID,
		OperatorName: buildTicketOperatorName(item.OperatorType, item.OperatorID, ctx),
		OldValue:     item.OldValue,
		NewValue:     item.NewValue,
		Content:      item.Content,
		Payload:      item.Payload,
		CreatedAt:    utils.FormatTime(item.CreatedAt),
	}
}

func BuildTicketEventLogList(list []models.TicketEventLog) []response.TicketEventLogResponse {
	return BuildTicketEventLogListWithContext(list, nil)
}

func BuildTicketEventLogListWithContext(list []models.TicketEventLog, ctx *TicketDetailBuildContext) []response.TicketEventLogResponse {
	if len(list) == 0 {
		return nil
	}
	results := make([]response.TicketEventLogResponse, 0, len(list))
	for i := range list {
		if item := BuildTicketEventLogWithContext(&list[i], ctx); item != nil {
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
	return BuildTicketDetailWithContext(aggregate, nil)
}

func BuildTicketDetailWithContext(aggregate *services.TicketDetailAggregate, ctx *TicketDetailBuildContext) *response.TicketDetailResponse {
	if aggregate == nil || aggregate.Ticket == nil {
		return nil
	}
	if ctx == nil {
		users := make(map[int64]*models.User, len(aggregate.Users)+len(aggregate.OperatorUsers))
		for id, item := range aggregate.Users {
			users[id] = item
		}
		for id, item := range aggregate.OperatorUsers {
			users[id] = item
		}
		ctx = &TicketDetailBuildContext{
			Users:          users,
			Teams:          aggregate.Teams,
			AgentProfiles:  aggregate.AgentProfiles,
			RelatedTickets: aggregate.RelatedMap,
			Customers:      aggregate.OperatorCustomers,
			AIAgents:       aggregate.OperatorAIAgents,
		}
	}
	ret := &response.TicketDetailResponse{
		Ticket:         *BuildTicket(aggregate.Ticket),
		Watchers:       BuildTicketWatcherListWithContext(aggregate.Watchers, ctx),
		Collaborators:  BuildTicketCollaboratorListWithContext(aggregate.Collaborators, ctx),
		RelatedTickets: BuildTicketRelationListWithContext(aggregate.RelatedTickets, ctx),
	}
	if len(aggregate.Comments) > 0 {
		ret.Comments = make([]response.TicketCommentResponse, 0, len(aggregate.Comments))
		for i := range aggregate.Comments {
			if item := BuildTicketCommentWithContext(&aggregate.Comments[i], ctx); item != nil {
				ret.Comments = append(ret.Comments, *item)
			}
		}
	}
	if len(aggregate.Events) > 0 {
		ret.Events = make([]response.TicketEventLogResponse, 0, len(aggregate.Events))
		for i := range aggregate.Events {
			if item := BuildTicketEventLogWithContext(&aggregate.Events[i], ctx); item != nil {
				ret.Events = append(ret.Events, *item)
			}
		}
	}
	if aggregate.Customer != nil {
		ret.Ticket.Customer = BuildCustomer(aggregate.Customer)
	}
	ret.Ticket.SLA = BuildTicketSLAList(aggregate.SLAs)
	for i := range ret.Comments {
		if ret.Comments[i].AuthorName == "" {
			ret.Comments[i].AuthorName = buildTicketOperatorName(ret.Comments[i].AuthorType, ret.Comments[i].AuthorID, ctx)
		}
	}
	for i := range ret.Events {
		if ret.Events[i].OperatorName == "" {
			ret.Events[i].OperatorName = buildTicketOperatorName(ret.Events[i].OperatorType, ret.Events[i].OperatorID, ctx)
		}
	}
	return ret
}

func BuildTicketRelation(item *models.TicketRelation) *response.TicketRelationResponse {
	return BuildTicketRelationWithContext(item, nil)
}

func BuildTicketRelationWithContext(item *models.TicketRelation, ctx *TicketDetailBuildContext) *response.TicketRelationResponse {
	if item == nil {
		return nil
	}
	ret := &response.TicketRelationResponse{
		ID:              item.ID,
		TicketID:        item.TicketID,
		RelatedTicketID: item.RelatedTicketID,
		RelationType:    item.RelationType,
	}
	related := (*models.Ticket)(nil)
	if ctx != nil && ctx.RelatedTickets != nil {
		related = ctx.RelatedTickets[item.RelatedTicketID]
	}
	if related == nil {
		related = services.TicketService.Get(item.RelatedTicketID)
	}
	if related != nil {
		ret.RelatedTicketNo = related.TicketNo
		ret.RelatedTicketTitle = related.Title
		ret.RelatedTicketStatus = related.Status
		ret.UpdatedAt = utils.FormatTime(related.UpdatedAt)
		if related.CurrentTeamID > 0 && ctx != nil && ctx.Teams != nil {
			if team := ctx.Teams[related.CurrentTeamID]; team != nil {
				ret.CurrentTeamName = team.Name
			}
		} else if related.CurrentTeamID > 0 {
			if team := services.AgentTeamService.Get(related.CurrentTeamID); team != nil {
				ret.CurrentTeamName = team.Name
			}
		}
		if related.CurrentAssigneeID > 0 && ctx != nil && ctx.Users != nil {
			ret.CurrentAssigneeName = buildTicketUserDisplayName(ctx.Users[related.CurrentAssigneeID])
		} else if related.CurrentAssigneeID > 0 {
			if user := services.UserService.Get(related.CurrentAssigneeID); user != nil {
				ret.CurrentAssigneeName = user.Nickname
				if ret.CurrentAssigneeName == "" {
					ret.CurrentAssigneeName = user.Username
				}
			}
		}
	}
	return ret
}

func BuildTicketRelationList(list []models.TicketRelation) []response.TicketRelationResponse {
	return BuildTicketRelationListWithContext(list, nil)
}

func BuildTicketRelationListWithContext(list []models.TicketRelation, ctx *TicketDetailBuildContext) []response.TicketRelationResponse {
	if len(list) == 0 {
		return nil
	}
	results := make([]response.TicketRelationResponse, 0, len(list))
	for i := range list {
		if item := BuildTicketRelationWithContext(&list[i], ctx); item != nil {
			results = append(results, *item)
		}
	}
	return results
}

func BuildTicketSummary(summary *services.TicketSummaryAggregate) *response.TicketSummaryResponse {
	if summary == nil {
		return nil
	}
	return &response.TicketSummaryResponse{
		All:             summary.All,
		Mine:            summary.Mine,
		Watching:        summary.Watching,
		Collaboration:   summary.Collaboration,
		Participating:   summary.Participating,
		Mentioned:       summary.Mentioned,
		Unassigned:      summary.Unassigned,
		PendingCustomer: summary.PendingCustomer,
		PendingInternal: summary.PendingInternal,
		Overdue:         summary.Overdue,
	}
}

func BuildTicketRiskOverview(overview *services.TicketRiskOverviewAggregate) *response.TicketRiskOverviewResponse {
	if overview == nil {
		return nil
	}
	ret := &response.TicketRiskOverviewResponse{
		Overdue:         overview.Overdue,
		HighRisk:        overview.HighRisk,
		Unassigned:      overview.Unassigned,
		PendingInternal: overview.PendingInternal,
		PendingCustomer: overview.PendingCustomer,
		RiskWindowMins:  overview.RiskWindowMins,
	}
	if len(overview.Reasons) > 0 {
		ret.Reasons = make([]response.TicketRiskReasonResponse, 0, len(overview.Reasons))
		for _, item := range overview.Reasons {
			ret.Reasons = append(ret.Reasons, response.TicketRiskReasonResponse{
				Code:        item.Code,
				Title:       item.Title,
				Description: item.Description,
				Count:       item.Count,
			})
		}
	}
	return ret
}

func BuildTicketWatcherList(list []models.TicketWatcher) []response.TicketWatcherResponse {
	return BuildTicketWatcherListWithContext(list, nil)
}

func BuildTicketWatcherListWithContext(list []models.TicketWatcher, ctx *TicketDetailBuildContext) []response.TicketWatcherResponse {
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
		if ctx != nil && ctx.Users != nil {
			out.UserName = buildTicketUserDisplayName(ctx.Users[item.UserID])
		} else if user := services.UserService.Get(item.UserID); user != nil {
			out.UserName = buildTicketUserDisplayName(user)
		}
		results = append(results, out)
	}
	return results
}

func BuildTicketCollaboratorList(list []models.TicketCollaborator) []response.TicketCollaboratorResponse {
	return BuildTicketCollaboratorListWithContext(list, nil)
}

func BuildTicketCollaboratorListWithContext(list []models.TicketCollaborator, ctx *TicketDetailBuildContext) []response.TicketCollaboratorResponse {
	if len(list) == 0 {
		return nil
	}
	results := make([]response.TicketCollaboratorResponse, 0, len(list))
	for i := range list {
		item := &list[i]
		out := response.TicketCollaboratorResponse{
			ID:     item.ID,
			UserID: item.UserID,
		}
		if ctx != nil && ctx.Users != nil {
			out.UserName = buildTicketUserDisplayName(ctx.Users[item.UserID])
		} else if user := services.UserService.Get(item.UserID); user != nil {
			out.UserName = buildTicketUserDisplayName(user)
		}
		if ctx != nil && ctx.AgentProfiles != nil {
			if profile := ctx.AgentProfiles[item.UserID]; profile != nil && profile.TeamID > 0 {
				if team := ctx.Teams[profile.TeamID]; team != nil {
					out.TeamName = team.Name
				}
			}
		} else if profile := services.AgentProfileService.GetByUserID(item.UserID); profile != nil && profile.TeamID > 0 {
			if team := services.AgentTeamService.Get(profile.TeamID); team != nil {
				out.TeamName = team.Name
			}
		}
		results = append(results, out)
	}
	return results
}

func BuildTicketView(item *models.TicketView) *response.TicketViewResponse {
	if item == nil {
		return nil
	}
	ret := &response.TicketViewResponse{
		ID:     item.ID,
		Name:   item.Name,
		SortNo: item.SortNo,
	}
	if strings.TrimSpace(item.FiltersJSON) != "" {
		_ = json.Unmarshal([]byte(item.FiltersJSON), &ret.Filters)
	}
	return ret
}

func BuildTicketViewList(list []models.TicketView) []response.TicketViewResponse {
	if len(list) == 0 {
		return nil
	}
	results := make([]response.TicketViewResponse, 0, len(list))
	for i := range list {
		if item := BuildTicketView(&list[i]); item != nil {
			results = append(results, *item)
		}
	}
	return results
}

func buildTicketUserDisplayName(user *models.User) string {
	if user == nil {
		return ""
	}
	if user.Nickname != "" {
		return user.Nickname
	}
	return user.Username
}

func buildTicketOperatorName(senderType enums.IMSenderType, senderID int64, ctx *TicketDetailBuildContext) string {
	if senderID <= 0 {
		return ""
	}
	switch senderType {
	case enums.IMSenderTypeAgent:
		if ctx != nil && ctx.Users != nil {
			if user := ctx.Users[senderID]; user != nil {
				return buildTicketUserDisplayName(user)
			}
		}
		if user := services.UserService.Get(senderID); user != nil {
			return buildTicketUserDisplayName(user)
		}
	case enums.IMSenderTypeCustomer:
		if ctx != nil && ctx.Customers != nil {
			if customer := ctx.Customers[senderID]; customer != nil {
				return customer.Name
			}
		}
		if customer := services.CustomerService.Get(senderID); customer != nil {
			return customer.Name
		}
	case enums.IMSenderTypeAI:
		if ctx != nil && ctx.AIAgents != nil {
			if ai := ctx.AIAgents[senderID]; ai != nil {
				return ai.Name
			}
		}
		if ai := services.AIAgentService.Get(senderID); ai != nil {
			return ai.Name
		}
	}
	return ""
}
