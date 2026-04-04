package console

import (
	"cs-agent/internal/builders"
	"cs-agent/internal/pkg/constants"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/services"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web"
	"github.com/mlogclub/simple/web/params"
)

type TicketResolutionCodeController struct{ Ctx iris.Context }

func (c *TicketResolutionCodeController) AnyList() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketResolutionCodeView); err != nil {
		return web.JsonError(err)
	}
	cnd := params.NewPagedSqlCnd(c.Ctx,
		params.QueryFilter{ParamName: "status"},
		params.QueryFilter{ParamName: "name", Op: params.Like},
	).Asc("sort_no").Desc("id")
	if _, ok := params.Get(c.Ctx, "status"); !ok {
		cnd.Where("status <> ?", enums.StatusDeleted)
	}
	list, paging := services.TicketResolutionCodeService.FindPageByCnd(cnd)
	return web.JsonData(&web.PageResult{Results: builders.BuildTicketResolutionCodeList(list), Page: paging})
}

func (c *TicketResolutionCodeController) GetList_all() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketResolutionCodeView); err != nil {
		return web.JsonError(err)
	}
	list := services.TicketResolutionCodeService.Find(sqls.NewCnd().Eq("status", enums.StatusOk).Asc("sort_no").Desc("id"))
	return web.JsonData(builders.BuildTicketResolutionCodeList(list))
}

func (c *TicketResolutionCodeController) PostCreate() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketResolutionCodeCreate)
	if err != nil {
		return web.JsonError(err)
	}
	req := request.CreateTicketResolutionCodeRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	item, err := services.TicketResolutionCodeService.CreateTicketResolutionCode(req, operator)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(builders.BuildTicketResolutionCode(item))
}

func (c *TicketResolutionCodeController) PostUpdate() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketResolutionCodeUpdate)
	if err != nil {
		return web.JsonError(err)
	}
	req := request.UpdateTicketResolutionCodeRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.TicketResolutionCodeService.UpdateTicketResolutionCode(req, operator); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *TicketResolutionCodeController) PostDelete() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketResolutionCodeDelete)
	if err != nil {
		return web.JsonError(err)
	}
	req := request.DeleteTicketResolutionCodeRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.TicketResolutionCodeService.DeleteTicketResolutionCode(req.ID, operator); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}
