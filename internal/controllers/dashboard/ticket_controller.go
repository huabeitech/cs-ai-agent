package dashboard

import (
	"strings"

	"cs-agent/internal/builders"
	"cs-agent/internal/pkg/constants"
	"cs-agent/internal/pkg/dto/request"
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
		params.QueryFilter{ParamName: "currentAssigneeId"},
		params.QueryFilter{ParamName: "customerId"},
		params.QueryFilter{ParamName: "conversationId"},
		params.QueryFilter{ParamName: "source"},
		params.QueryFilter{ParamName: "channel"},
	).Desc("updated_at").Desc("id")
	if keyword, _ := params.Get(c.Ctx, "keyword"); strings.TrimSpace(keyword) != "" {
		keyword = "%" + strings.TrimSpace(keyword) + "%"
		cnd.Where("ticket_no LIKE ? OR title LIKE ? OR description LIKE ?", keyword, keyword, keyword)
	}
	if tagID, _ := params.GetInt64(c.Ctx, "tagId"); tagID > 0 {
		cnd.Where("id IN (SELECT ticket_id FROM t_ticket_tag WHERE tag_id = ?)", tagID)
	}
	if mine, _ := params.Get(c.Ctx, "mine"); mine == "1" || strings.EqualFold(mine, "true") {
		cnd.Eq("current_assignee_id", operator.UserID)
	}
	if unassigned, _ := params.Get(c.Ctx, "unassigned"); unassigned == "1" || strings.EqualFold(unassigned, "true") {
		cnd.Eq("current_assignee_id", 0)
	}
	if staleHoursValue, _ := params.Get(c.Ctx, "staleHours"); strings.TrimSpace(staleHoursValue) != "" {
		staleHours, _ := params.GetInt(c.Ctx, "staleHours")
		services.TicketService.ApplyStaleFilter(cnd, staleHours)
	}
	aggregate, err := services.TicketService.FindPageAggregateByCnd(cnd, operator.UserID)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(&web.PageResult{
		Results: builders.BuildTicketListWithContext(aggregate.List, &builders.TicketBuildContext{
			TagsByTicketID: aggregate.TagsByTicketID,
			Users:          aggregate.Users,
			Customers:      aggregate.Customers,
		}),
		Page: aggregate.Paging,
	})
}

func (c *TicketController) AnySummary() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketView)
	if err != nil {
		return web.JsonError(err)
	}
	staleHours, _ := params.GetInt(c.Ctx, "staleHours")
	return web.JsonData(builders.BuildTicketSummary(services.TicketService.GetSummary(operator, staleHours)))
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

func (c *TicketController) PostAdd_progress() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketProgress)
	if err != nil {
		return web.JsonError(err)
	}
	req := request.CreateTicketProgressRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	item, err := services.TicketService.AddProgress(req, operator)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(builders.BuildTicketProgress(item))
}
