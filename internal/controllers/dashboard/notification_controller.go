package dashboard

import (
	"strings"

	"cs-agent/internal/builders"
	"cs-agent/internal/pkg/constants"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/services"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/web"
	"github.com/mlogclub/simple/web/params"
)

type NotificationController struct {
	Ctx iris.Context
}

func (c *NotificationController) AnyList() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionNotificationView)
	if err != nil {
		return web.JsonError(err)
	}

	cnd := params.NewPagedSqlCnd(c.Ctx,
		params.QueryFilter{ParamName: "type", ColumnName: "notification_type"},
	).Eq("recipient_user_id", operator.UserID).
		Eq("status", enums.StatusOk).
		Desc("id")

	switch strings.TrimSpace(c.Ctx.URLParam("readStatus")) {
	case "unread":
		cnd.Where("read_at IS NULL")
	case "read":
		cnd.Where("read_at IS NOT NULL")
	}

	list, paging := services.NotificationService.FindPageByCnd(cnd)
	return web.JsonData(&web.PageResult{
		Results: builders.BuildNotificationList(list),
		Page:    paging,
	})
}

func (c *NotificationController) GetUnread_count() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionNotificationView)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(&response.NotificationUnreadCountResponse{
		UnreadCount: services.NotificationService.CountUnread(operator.UserID),
	})
}

func (c *NotificationController) PostMark_read() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionNotificationUpdate)
	if err != nil {
		return web.JsonError(err)
	}
	req := request.MarkNotificationReadRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.NotificationService.MarkRead(req.ID, operator.UserID); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *NotificationController) PostMark_all_read() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionNotificationUpdate)
	if err != nil {
		return web.JsonError(err)
	}
	if err := services.NotificationService.MarkAllRead(operator.UserID); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}
