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

type TicketCategoryController struct{ Ctx iris.Context }

func (c *TicketCategoryController) AnyList() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketCategoryView); err != nil {
		return web.JsonError(err)
	}
	cnd := params.NewPagedSqlCnd(c.Ctx,
		params.QueryFilter{ParamName: "status"},
		params.QueryFilter{ParamName: "parentId"},
		params.QueryFilter{ParamName: "name", Op: params.Like},
	).Asc("sort_no").Desc("id")
	if _, ok := params.Get(c.Ctx, "status"); !ok {
		cnd.Where("status <> ?", enums.StatusDeleted)
	}
	list, paging := services.TicketCategoryService.FindPageByCnd(cnd)
	return web.JsonData(&web.PageResult{Results: builders.BuildTicketCategoryList(list), Page: paging})
}

func (c *TicketCategoryController) GetList_all() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketCategoryView); err != nil {
		return web.JsonError(err)
	}
	list := services.TicketCategoryService.Find(sqls.NewCnd().Eq("status", enums.StatusOk).Asc("sort_no").Desc("id"))
	return web.JsonData(builders.BuildTicketCategoryList(list))
}

func (c *TicketCategoryController) PostCreate() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketCategoryCreate)
	if err != nil {
		return web.JsonError(err)
	}
	req := request.CreateTicketCategoryRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	item, err := services.TicketCategoryService.CreateTicketCategory(req, operator)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(builders.BuildTicketCategory(item))
}

func (c *TicketCategoryController) PostUpdate() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketCategoryUpdate)
	if err != nil {
		return web.JsonError(err)
	}
	req := request.UpdateTicketCategoryRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.TicketCategoryService.UpdateTicketCategory(req, operator); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *TicketCategoryController) PostDelete() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketCategoryDelete)
	if err != nil {
		return web.JsonError(err)
	}
	req := request.DeleteTicketCategoryRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.TicketCategoryService.DeleteTicketCategory(req.ID, operator); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}
