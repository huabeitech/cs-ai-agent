package dashboard

import (
	"strings"
	"time"

	"cs-agent/internal/builders"
	"cs-agent/internal/pkg/constants"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/services"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/web"
	"github.com/mlogclub/simple/web/params"
)

type TicketController struct {
	Ctx iris.Context
}

func (c *TicketController) AnyList() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketView)
	if err != nil {
		return web.JsonError(err)
	}
	cnd := params.NewPagedSqlCnd(c.Ctx,
		params.QueryFilter{ParamName: "status"},
		params.QueryFilter{ParamName: "priority"},
		params.QueryFilter{ParamName: "severity"},
		params.QueryFilter{ParamName: "currentTeamId"},
		params.QueryFilter{ParamName: "currentAssigneeId"},
		params.QueryFilter{ParamName: "customerId"},
		params.QueryFilter{ParamName: "conversationId"},
		params.QueryFilter{ParamName: "source"},
	).Desc("updated_at").Desc("id")
	if keyword, _ := params.Get(c.Ctx, "keyword"); strings.TrimSpace(keyword) != "" {
		keyword = "%" + strings.TrimSpace(keyword) + "%"
		cnd.Where("ticket_no LIKE ? OR title LIKE ? OR description LIKE ?", keyword, keyword, keyword)
	}
	if tagID, _ := params.GetInt64(c.Ctx, "tagId"); tagID > 0 {
		cnd.Where("id IN (SELECT ticket_id FROM t_ticket_tag WHERE tag_id = ?)", tagID)
	}
	if watching, _ := params.Get(c.Ctx, "watching"); watching == "1" || strings.EqualFold(watching, "true") {
		cnd.Where("id IN (SELECT ticket_id FROM t_ticket_watcher WHERE user_id = ?)", operator.UserID)
	}
	if collaborating, _ := params.Get(c.Ctx, "collaborating"); collaborating == "1" || strings.EqualFold(collaborating, "true") {
		cnd.Where("id IN (SELECT ticket_id FROM t_ticket_collaborator WHERE user_id = ?)", operator.UserID)
	}
	if collaboration, _ := params.Get(c.Ctx, "collaboration"); collaboration == "1" || strings.EqualFold(collaboration, "true") {
		cnd.Where(
			"(id IN (SELECT ticket_id FROM t_ticket_collaborator WHERE user_id = ?) OR id IN (SELECT ticket_id FROM t_ticket_mention WHERE mentioned_user_id = ?))",
			operator.UserID,
			operator.UserID,
		)
	}
	if mentioned, _ := params.Get(c.Ctx, "mentioned"); mentioned == "1" || strings.EqualFold(mentioned, "true") {
		cnd.Where("id IN (SELECT ticket_id FROM t_ticket_mention WHERE mentioned_user_id = ?)", operator.UserID)
	}
	if mine, _ := params.Get(c.Ctx, "mine"); mine == "1" || strings.EqualFold(mine, "true") {
		cnd.Eq("current_assignee_id", operator.UserID)
	}
	if unassigned, _ := params.Get(c.Ctx, "unassigned"); unassigned == "1" || strings.EqualFold(unassigned, "true") {
		cnd.Eq("current_assignee_id", 0)
	}
	if overdue, _ := params.Get(c.Ctx, "overdue"); overdue == "1" || strings.EqualFold(overdue, "true") {
		cnd.In("status", []string{"new", "open", "pending_customer", "pending_internal"})
		cnd.Where("resolve_deadline_at IS NOT NULL")
		cnd.Where("resolve_deadline_at < ?", time.Now())
	}
	aggregate, err := services.TicketService.FindPageAggregateByCnd(cnd, operator.UserID)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(&web.PageResult{
		Results: builders.BuildTicketListWithContext(aggregate.List, &builders.TicketBuildContext{
			TagsByTicketID:   aggregate.TagsByTicketID,
			Priorities:       aggregate.Priorities,
			ResolutionCodes:  aggregate.ResolutionCodes,
			Users:            aggregate.Users,
			Teams:            aggregate.Teams,
			Customers:        aggregate.Customers,
			SLAByTicketID:    aggregate.SLAByTicketID,
			WatchedTicketIDs: aggregate.WatchedTicketIDs,
		}),
		Page: aggregate.Paging,
	})
}

func (c *TicketController) AnySummary() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketView)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(builders.BuildTicketSummary(services.TicketService.GetSummary(operator)))
}

func (c *TicketController) AnyRisk_overview() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketView); err != nil {
		return web.JsonError(err)
	}
	teamID, _ := params.GetInt64(c.Ctx, "currentTeamId")
	riskWindowMins, _ := params.GetInt(c.Ctx, "riskWindowMins")
	return web.JsonData(builders.BuildTicketRiskOverview(services.TicketService.GetRiskOverview(teamID, riskWindowMins)))
}

func (c *TicketController) AnyRisk_list() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketView)
	if err != nil {
		return web.JsonError(err)
	}
	riskType, _ := params.Get(c.Ctx, "riskType")
	teamID, _ := params.GetInt64(c.Ctx, "currentTeamId")
	riskWindowMins, _ := params.GetInt(c.Ctx, "riskWindowMins")
	page, _ := params.GetInt(c.Ctx, "page")
	limit, _ := params.GetInt(c.Ctx, "limit")
	aggregate, err := services.TicketService.GetRiskPageAggregate(riskType, teamID, riskWindowMins, page, limit, operator.UserID)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(&web.PageResult{
		Results: builders.BuildTicketListWithContext(aggregate.List, &builders.TicketBuildContext{
			TagsByTicketID:   aggregate.TagsByTicketID,
			Priorities:       aggregate.Priorities,
			ResolutionCodes:  aggregate.ResolutionCodes,
			Users:            aggregate.Users,
			Teams:            aggregate.Teams,
			Customers:        aggregate.Customers,
			SLAByTicketID:    aggregate.SLAByTicketID,
			WatchedTicketIDs: aggregate.WatchedTicketIDs,
		}),
		Page: aggregate.Paging,
	})
}

func (c *TicketController) AnyView_list() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketView)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(builders.BuildTicketViewList(services.TicketViewService.ListByUser(operator.UserID)))
}

func (c *TicketController) PostSave_view() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketView)
	if err != nil {
		return web.JsonError(err)
	}
	req := request.SaveTicketViewRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	item, err := services.TicketViewService.Save(req, operator)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(builders.BuildTicketView(item))
}

func (c *TicketController) PostDelete_view() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketView)
	if err != nil {
		return web.JsonError(err)
	}
	req := request.DeleteTicketViewRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.TicketViewService.Delete(req.ID, operator); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *TicketController) GetBy(id int64) *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketView); err != nil {
		return web.JsonError(err)
	}
	detail, err := services.TicketService.GetDetail(id)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(builders.BuildTicketDetail(detail))
}

func (c *TicketController) PostCreate() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketCreate)
	if err != nil {
		return web.JsonError(err)
	}
	req := request.CreateTicketRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	item, err := services.TicketService.CreateTicket(req, operator)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(builders.BuildTicket(item))
}

func (c *TicketController) PostCreate_from_conversation() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketCreate)
	if err != nil {
		return web.JsonError(err)
	}
	req := request.CreateTicketFromConversationRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	item, err := services.TicketService.CreateFromConversation(req, operator)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(builders.BuildTicket(item))
}

func (c *TicketController) PostUpdate() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketUpdate)
	if err != nil {
		return web.JsonError(err)
	}
	req := request.UpdateTicketRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.TicketService.UpdateTicket(req, operator); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *TicketController) PostLink_customer() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketUpdate)
	if err != nil {
		return web.JsonError(err)
	}
	req := request.LinkTicketCustomerRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.TicketService.LinkTicketCustomer(req.TicketID, req.CustomerID, operator); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *TicketController) PostAssign() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketAssign)
	if err != nil {
		return web.JsonError(err)
	}
	req := request.AssignTicketRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.TicketService.AssignTicket(req, operator); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *TicketController) PostBatch_assign() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketAssign)
	if err != nil {
		return web.JsonError(err)
	}
	req := request.BatchAssignTicketRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.TicketService.BatchAssignTickets(req, operator); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *TicketController) PostChange_status() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketChangeStatus)
	if err != nil {
		return web.JsonError(err)
	}
	req := request.ChangeTicketStatusRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.TicketService.ChangeStatus(req, operator); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *TicketController) PostBatch_change_status() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketChangeStatus)
	if err != nil {
		return web.JsonError(err)
	}
	req := request.BatchChangeTicketStatusRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.TicketService.BatchChangeStatus(req, operator); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *TicketController) PostReply() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketReply)
	if err != nil {
		return web.JsonError(err)
	}
	req := request.ReplyTicketRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	item, err := services.TicketService.ReplyTicket(req, operator)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(builders.BuildTicketComment(item))
}

func (c *TicketController) PostInternal_note() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketReply)
	if err != nil {
		return web.JsonError(err)
	}
	req := request.InternalNoteRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	item, err := services.TicketService.AddInternalNote(req, operator)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(builders.BuildTicketComment(item))
}

func (c *TicketController) PostClose() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketClose)
	if err != nil {
		return web.JsonError(err)
	}
	req := request.CloseTicketRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.TicketService.CloseTicket(req, operator); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *TicketController) PostReopen() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketReopen)
	if err != nil {
		return web.JsonError(err)
	}
	req := request.ReopenTicketRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.TicketService.ReopenTicket(req, operator); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *TicketController) PostWatch() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketView)
	if err != nil {
		return web.JsonError(err)
	}
	req := request.WatchTicketRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.TicketService.WatchTicket(req.TicketID, operator); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *TicketController) PostUnwatch() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketView)
	if err != nil {
		return web.JsonError(err)
	}
	req := request.WatchTicketRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.TicketService.UnwatchTicket(req.TicketID, operator); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *TicketController) PostBatch_watch() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketView)
	if err != nil {
		return web.JsonError(err)
	}
	req := request.BatchWatchTicketRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.TicketService.BatchWatchTickets(req, operator); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *TicketController) PostAdd_relation() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketUpdate)
	if err != nil {
		return web.JsonError(err)
	}
	req := request.AddTicketRelationRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	relatedTicketID := req.RelatedTicketID
	if relatedTicketID <= 0 && strings.TrimSpace(req.RelatedTicketNo) != "" {
		if relatedTicket := services.TicketService.Take("ticket_no = ?", strings.TrimSpace(req.RelatedTicketNo)); relatedTicket != nil {
			relatedTicketID = relatedTicket.ID
		}
	}
	if err := services.TicketRelationService.AddRelation(req.TicketID, relatedTicketID, enums.TicketRelationType(strings.TrimSpace(req.RelationType)), operator); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *TicketController) PostDelete_relation() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketUpdate)
	if err != nil {
		return web.JsonError(err)
	}
	req := request.DeleteTicketRelationRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.TicketRelationService.DeleteRelation(req.TicketID, req.RelationID, operator); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *TicketController) PostAdd_collaborator() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketUpdate)
	if err != nil {
		return web.JsonError(err)
	}
	req := request.AddTicketCollaboratorRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.TicketService.AddCollaborator(req.TicketID, req.UserID, operator); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *TicketController) PostDelete_collaborator() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketUpdate)
	if err != nil {
		return web.JsonError(err)
	}
	req := request.DeleteTicketCollaboratorRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.TicketService.RemoveCollaborator(req.TicketID, req.CollaboratorID, operator); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *TicketController) AnyComment_list() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketView); err != nil {
		return web.JsonError(err)
	}
	ticketID, _ := params.GetInt64(c.Ctx, "ticketId")
	if ticketID <= 0 {
		return web.JsonData(&web.PageResult{Results: []any{}, Page: params.GetPaging(c.Ctx)})
	}
	cnd := params.NewPagedSqlCnd(c.Ctx, params.QueryFilter{ParamName: "ticketId"}).Asc("id")
	results, paging := services.TicketCommentService.FindPageByCnd(cnd)
	return web.JsonData(&web.PageResult{Results: builders.BuildTicketCommentList(results), Page: paging})
}

func (c *TicketController) AnyEvent_list() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketView); err != nil {
		return web.JsonError(err)
	}
	cnd := params.NewPagedSqlCnd(c.Ctx, params.QueryFilter{ParamName: "ticketId"}).Desc("id")
	list, paging := services.TicketEventLogService.FindPageByCnd(cnd)
	return web.JsonData(&web.PageResult{Results: builders.BuildTicketEventLogList(list), Page: paging})
}
