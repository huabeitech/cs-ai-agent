package console

import (
	"strings"
	"time"

	"cs-agent/internal/builders"
	"cs-agent/internal/pkg/constants"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/services"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/sqls"
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
		params.QueryFilter{ParamName: "categoryId"},
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
	if watching, _ := params.Get(c.Ctx, "watching"); watching == "1" || strings.EqualFold(watching, "true") {
		cnd.Where("id IN (SELECT ticket_id FROM ticket_watchers WHERE user_id = ?)", operator.UserID)
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
	list, paging := services.TicketService.FindPageByCnd(cnd)
	results := builders.BuildTicketList(list)
	for i := range results {
		results[i].WatchedByMe = services.TicketWatcherService.FindOne(
			sqls.NewCnd().Eq("ticket_id", results[i].ID).Eq("user_id", operator.UserID),
		) != nil
	}
	return web.JsonData(&web.PageResult{Results: results, Page: paging})
}

func (c *TicketController) AnySummary() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketView)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(builders.BuildTicketSummary(services.TicketService.GetSummary(operator)))
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
