package console

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/constants"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/utils"
	"cs-agent/internal/services"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web"
	"github.com/mlogclub/simple/web/params"
)

type TicketReplyController struct {
	Ctx iris.Context
}

func (c *TicketReplyController) AnyList() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketReplyView); err != nil {
		return web.JsonError(err)
	}

	ticketId, _ := params.GetInt64(c.Ctx, "ticketId")
	if ticketId == 0 {
		return web.JsonErrorMsg("缺少工单ID")
	}

	cnd := sqls.NewCnd().Where("ticket_id = ?", ticketId).Desc("id")
	list, paging := services.TicketReplyService.FindPageByCnd(cnd)
	results := make([]map[string]any, 0, len(list))
	for _, item := range list {
		results = append(results, map[string]any{
			"id":            item.ID,
			"ticketId":      item.TicketID,
			"parentId":      item.ParentID,
			"content":       item.Content,
			"senderType":    item.SenderType,
			"senderId":      item.SenderID,
			"senderName":    item.SenderName,
			"isInternal":    item.IsInternal,
			"sendStatus":    item.SendStatus,
			"attachmentIds": item.AttachmentIDs,
			"createdAt":     item.CreatedAt.Unix(),
		})
	}
	return web.JsonData(&web.PageResult{Results: results, Page: paging})
}

func (c *TicketReplyController) PostCreate() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketReplyCreate)
	if err != nil {
		return web.JsonError(err)
	}

	req := request.CreateTicketReplyRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	reply := &models.TicketReply{
		TicketID:      req.TicketID,
		ParentID:      req.ParentID,
		Content:       req.Content,
		SenderType:    string(enums.TicketSenderTypeAgent),
		SenderID:      operator.UserID,
		SenderName:    operator.Nickname,
		IsInternal:    req.IsInternal,
		SendStatus:    int(enums.TicketReplySendStatusSent),
		AttachmentIDs: req.AttachmentIDs,
		AuditFields:   utils.BuildAuditFields(operator),
	}
	if err = services.TicketReplyService.Create(reply); err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(map[string]any{
		"id":         reply.ID,
		"ticketId":   reply.TicketID,
		"content":    reply.Content,
		"senderType": reply.SenderType,
		"senderId":   reply.SenderID,
		"senderName": reply.SenderName,
		"createdAt":  reply.CreatedAt.Unix(),
	})
}

func (c *TicketReplyController) PostUpdate() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketReplyUpdate); err != nil {
		return web.JsonError(err)
	}

	req := request.UpdateTicketReplyRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	updates := map[string]interface{}{
		"content": req.Content,
	}
	if err := services.TicketReplyService.Updates(req.ID, updates); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *TicketReplyController) PostDelete() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketReplyDelete); err != nil {
		return web.JsonError(err)
	}

	req := request.DeleteTicketReplyRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	services.TicketReplyService.Delete(req.ID)
	return web.JsonSuccess()
}
