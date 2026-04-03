package builders

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/utils"
	"cs-agent/internal/services"

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
	return BuildTicketDetailWithContext(aggregate, nil)
}

func BuildTicketDetailWithContext(aggregate *services.TicketDetailAggregate, ctx *TicketDetailBuildContext) *response.TicketDetailResponse {
	if aggregate == nil || aggregate.Ticket == nil {
		return nil
	}
	if ctx == nil {
		ctx = &TicketDetailBuildContext{
			Users:          aggregate.Users,
			Teams:          aggregate.Teams,
			AgentProfiles:  aggregate.AgentProfiles,
			RelatedTickets: aggregate.RelatedMap,
		}
	}
	ret := &response.TicketDetailResponse{
		Ticket:         *BuildTicket(aggregate.Ticket),
		Watchers:       BuildTicketWatcherListWithContext(aggregate.Watchers, ctx),
		Collaborators:  BuildTicketCollaboratorListWithContext(aggregate.Collaborators, ctx),
		Comments:       BuildTicketCommentList(aggregate.Comments),
		Events:         BuildTicketEventLogList(aggregate.Events),
		RelatedTickets: BuildTicketRelationListWithContext(aggregate.RelatedTickets, ctx),
	}
	if aggregate.Customer != nil {
		ret.Ticket.Customer = BuildCustomer(aggregate.Customer)
	}
	ret.Ticket.SLA = BuildTicketSLAList(aggregate.SLAs)
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

func buildTicketUserDisplayName(user *models.User) string {
	if user == nil {
		return ""
	}
	if user.Nickname != "" {
		return user.Nickname
	}
	return user.Username
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
