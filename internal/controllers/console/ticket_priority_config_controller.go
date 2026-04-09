package console

import (
	"cs-agent/internal/builders"
	"cs-agent/internal/pkg/constants"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/services"
	"strings"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web"
	"github.com/mlogclub/simple/web/params"
)

type TicketPriorityConfigController struct{ Ctx iris.Context }

func (c *TicketPriorityConfigController) AnyList() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketPriorityConfigView); err != nil {
		return web.JsonError(err)
	}
	cnd := sqls.NewCnd().Asc("sort_no").Asc("id")
	if status, ok := params.Get(c.Ctx, "status"); ok && strings.TrimSpace(status) != "" {
		cnd.Eq("status", status)
	} else {
		cnd.Where("status <> ?", enums.StatusDeleted)
	}
	if name, ok := params.Get(c.Ctx, "name"); ok && strings.TrimSpace(name) != "" {
		cnd.Where("name LIKE ?", "%"+strings.TrimSpace(name)+"%")
	}
	list := services.TicketPriorityConfigService.Find(cnd)
	return web.JsonData(builders.BuildTicketPriorityConfigList(list))
}

func (c *TicketPriorityConfigController) GetList_all() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketPriorityConfigView); err != nil {
		return web.JsonError(err)
	}
	list := services.TicketPriorityConfigService.Find(sqls.NewCnd().Eq("status", enums.StatusOk).Asc("sort_no").Asc("id"))
	return web.JsonData(builders.BuildTicketPriorityConfigList(list))
}

func (c *TicketPriorityConfigController) PostCreate() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketPriorityConfigCreate)
	if err != nil {
		return web.JsonError(err)
	}
	req := request.CreateTicketPriorityConfigRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	item, err := services.TicketPriorityConfigService.CreateTicketPriorityConfig(req, operator)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(builders.BuildTicketPriorityConfig(item))
}

func (c *TicketPriorityConfigController) PostUpdate() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketPriorityConfigUpdate)
	if err != nil {
		return web.JsonError(err)
	}
	req := request.UpdateTicketPriorityConfigRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.TicketPriorityConfigService.UpdateTicketPriorityConfig(req, operator); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *TicketPriorityConfigController) PostUpdate_sort() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketPriorityConfigUpdate); err != nil {
		return web.JsonError(err)
	}
	var ids []int64
	if err := c.Ctx.ReadJSON(&ids); err != nil {
		return web.JsonError(err)
	}
	if err := services.TicketPriorityConfigService.UpdateSort(ids); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *TicketPriorityConfigController) PostDelete() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketPriorityConfigDelete)
	if err != nil {
		return web.JsonError(err)
	}
	req := request.DeleteTicketPriorityConfigRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.TicketPriorityConfigService.DeleteTicketPriorityConfig(req.ID, operator); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}
