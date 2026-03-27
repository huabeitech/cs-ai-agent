package console

import (
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
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketView); err != nil {
		return web.JsonError(err)
	}

	cnd := params.NewPagedSqlCnd(c.Ctx,
		params.QueryFilter{ParamName: "status"},
		params.QueryFilter{ParamName: "priority"},
		params.QueryFilter{ParamName: "categoryId"},
		params.QueryFilter{ParamName: "currentAssigneeId"},
		params.QueryFilter{ParamName: "channelType"},
		params.QueryFilter{ParamName: "title", Op: params.Like},
	).Desc("id")
	list, paging := services.TicketService.FindPageByCnd(cnd)
	results := make([]map[string]any, 0, len(list))
	for _, item := range list {
		results = append(results, map[string]any{
			"id":                 item.ID,
			"ticketNo":           item.TicketNo,
			"title":              item.Title,
			"content":            item.Content,
			"channelType":        item.ChannelType,
			"channelId":          item.ChannelID,
			"categoryId":         item.CategoryID,
			"priority":           item.Priority,
			"status":             item.Status,
			"sourceUserId":       item.SourceUserID,
			"externalUserId":     item.ExternalUserID,
			"externalUserName":   item.ExternalUserName,
			"externalUserEmail":  item.ExternalUserEmail,
			"externalUserMobile": item.ExternalUserMobile,
			"currentAssigneeId":  item.CurrentAssigneeID,
			"currentTeamId":      item.CurrentTeamID,
			"conversationId":     item.ConversationID,
			"replyCount":         item.ReplyCount,
			"satisfied":          item.Satisfied,
			"satisfiedRemark":    item.SatisfiedRemark,
			"tags":               item.Tags,
			"remark":             item.Remark,
			"createdAt":          item.CreatedAt.Unix(),
			"updatedAt":          item.UpdatedAt.Unix(),
		})
	}
	return web.JsonData(&web.PageResult{Results: results, Page: paging})
}

func (c *TicketController) GetBy(id int64) *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketView); err != nil {
		return web.JsonError(err)
	}
	ticket := services.TicketService.Get(id)
	if ticket == nil {
		return web.JsonErrorMsg("工单不存在")
	}
	return web.JsonData(map[string]any{
		"id":                 ticket.ID,
		"ticketNo":           ticket.TicketNo,
		"title":              ticket.Title,
		"content":            ticket.Content,
		"channelType":        ticket.ChannelType,
		"channelId":          ticket.ChannelID,
		"categoryId":         ticket.CategoryID,
		"priority":           ticket.Priority,
		"status":             ticket.Status,
		"sourceUserId":       ticket.SourceUserID,
		"externalUserId":     ticket.ExternalUserID,
		"externalUserName":   ticket.ExternalUserName,
		"externalUserEmail":  ticket.ExternalUserEmail,
		"externalUserMobile": ticket.ExternalUserMobile,
		"currentAssigneeId":  ticket.CurrentAssigneeID,
		"currentTeamId":      ticket.CurrentTeamID,
		"conversationId":     ticket.ConversationID,
		"replyCount":         ticket.ReplyCount,
		"satisfied":          ticket.Satisfied,
		"satisfiedRemark":    ticket.SatisfiedRemark,
		"tags":               ticket.Tags,
		"remark":             ticket.Remark,
		"createdAt":          ticket.CreatedAt.Unix(),
		"updatedAt":          ticket.UpdatedAt.Unix(),
	})
}

func (c *TicketController) PostCreate() *web.JsonResult {
	user, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketCreate)
	if err != nil {
		return web.JsonError(err)
	}

	req := request.CreateTicketRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	item, err := services.TicketService.CreateTicket(req, user)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(map[string]any{
		"id":        item.ID,
		"ticketNo":  item.TicketNo,
		"title":     item.Title,
		"status":    item.Status,
		"createdAt": item.CreatedAt.Unix(),
	})
}

func (c *TicketController) PostUpdate() *web.JsonResult {
	user, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketUpdate)
	if err != nil {
		return web.JsonError(err)
	}

	req := request.UpdateTicketRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.TicketService.UpdateTicket(req, user); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *TicketController) PostDelete() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketDelete); err != nil {
		return web.JsonError(err)
	}

	req := request.DeleteTicketRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.TicketService.DeleteTicket(req.ID); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *TicketController) PostAssign() *web.JsonResult {
	user, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketAssign)
	if err != nil {
		return web.JsonError(err)
	}

	req := request.AssignTicketRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.TicketService.AssignTicket(req, user); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *TicketController) PostClose() *web.JsonResult {
	user, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketClose)
	if err != nil {
		return web.JsonError(err)
	}

	req := request.CloseTicketRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.TicketService.CloseTicket(req.ID, user); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *TicketController) PostReopen() *web.JsonResult {
	user, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketReopen)
	if err != nil {
		return web.JsonError(err)
	}

	req := request.ReopenTicketRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.TicketService.ReopenTicket(req.ID, user); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}
