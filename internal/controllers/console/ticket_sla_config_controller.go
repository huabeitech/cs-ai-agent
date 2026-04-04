package console

import (
	"cs-agent/internal/builders"
	"cs-agent/internal/pkg/constants"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/services"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/web"
	"github.com/mlogclub/simple/web/params"
)

type TicketSLAConfigController struct{ Ctx iris.Context }

func (c *TicketSLAConfigController) AnyList() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketSLAConfigView); err != nil {
		return web.JsonError(err)
	}
	cnd := params.NewPagedSqlCnd(c.Ctx,
		params.QueryFilter{ParamName: "status"},
		params.QueryFilter{ParamName: "priority"},
		params.QueryFilter{ParamName: "name", Op: params.Like},
	).Asc("priority").Desc("id")
	if _, ok := params.Get(c.Ctx, "status"); !ok {
		cnd.Where("status <> ?", enums.StatusDeleted)
	}
	list, paging := services.TicketSLAConfigService.FindPageByCnd(cnd)
	return web.JsonData(&web.PageResult{Results: builders.BuildTicketSLAConfigList(list), Page: paging})
}

func (c *TicketSLAConfigController) PostCreate() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketSLAConfigCreate)
	if err != nil {
		return web.JsonError(err)
	}
	req := request.CreateTicketSLAConfigRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	item, err := services.TicketSLAConfigService.CreateTicketSLAConfig(req, operator)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(builders.BuildTicketSLAConfig(item))
}

func (c *TicketSLAConfigController) PostUpdate() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketSLAConfigUpdate)
	if err != nil {
		return web.JsonError(err)
	}
	req := request.UpdateTicketSLAConfigRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.TicketSLAConfigService.UpdateTicketSLAConfig(req, operator); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *TicketSLAConfigController) PostDelete() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketSLAConfigDelete)
	if err != nil {
		return web.JsonError(err)
	}
	req := request.DeleteTicketSLAConfigRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.TicketSLAConfigService.DeleteTicketSLAConfig(req.ID, operator); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}
