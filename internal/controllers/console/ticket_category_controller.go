package console

import (
	"cs-agent/internal/pkg/constants"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/services"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/web"
	"github.com/mlogclub/simple/web/params"
)

type TicketCategoryController struct {
	Ctx iris.Context
}

func (c *TicketCategoryController) AnyList() *web.JsonResult {
	if err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketCategoryView); err != nil {
		return web.JsonError(err)
	}

	cnd := params.NewPagedSqlCnd(c.Ctx,
		params.QueryFilter{ParamName: "status"},
		params.QueryFilter{ParamName: "name", Op: params.Like},
		params.QueryFilter{ParamName: "code"},
		params.QueryFilter{ParamName: "parentId"},
	).Desc("sort_no").Desc("id")
	list, paging := services.TicketCategoryService.FindPageByCnd(cnd)
	results := make([]map[string]any, 0, len(list))
	for _, item := range list {
		results = append(results, map[string]any{
			"id":          item.ID,
			"parentId":    item.ParentID,
			"name":        item.Name,
			"code":        item.Code,
			"description": item.Description,
			"sortNo":      item.SortNo,
			"status":      item.Status,
			"remark":      item.Remark,
		})
	}
	return web.JsonData(&web.PageResult{Results: results, Page: paging})
}

func (c *TicketCategoryController) GetTree() *web.JsonResult {
	if err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketCategoryView); err != nil {
		return web.JsonError(err)
	}
	list := services.TicketCategoryService.FindTree()
	results := make([]map[string]any, 0, len(list))
	for _, item := range list {
		results = append(results, map[string]any{
			"id":          item.ID,
			"parentId":    item.ParentID,
			"name":        item.Name,
			"code":        item.Code,
			"description": item.Description,
			"sortNo":      item.SortNo,
			"status":      item.Status,
			"remark":      item.Remark,
		})
	}
	return web.JsonData(results)
}

func (c *TicketCategoryController) PostCreate() *web.JsonResult {
	if err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketCategoryCreate); err != nil {
		return web.JsonError(err)
	}

	req := request.CreateTicketCategoryRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	item, err := services.TicketCategoryService.CreateCategory(req, services.AuthService.GetAuthPrincipal(c.Ctx))
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(map[string]any{
		"id":          item.ID,
		"parentId":    item.ParentID,
		"name":        item.Name,
		"code":        item.Code,
		"description": item.Description,
		"sortNo":      item.SortNo,
		"status":      item.Status,
		"remark":      item.Remark,
	})
}

func (c *TicketCategoryController) PostUpdate() *web.JsonResult {
	if err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketCategoryUpdate); err != nil {
		return web.JsonError(err)
	}

	req := request.UpdateTicketCategoryRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.TicketCategoryService.UpdateCategory(req, services.AuthService.GetAuthPrincipal(c.Ctx)); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *TicketCategoryController) PostDelete() *web.JsonResult {
	if err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketCategoryDelete); err != nil {
		return web.JsonError(err)
	}

	req := request.DeleteTicketCategoryRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.TicketCategoryService.DeleteCategory(req.ID); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *TicketCategoryController) PostUpdate_sort() *web.JsonResult {
	if err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTicketCategoryUpdate); err != nil {
		return web.JsonError(err)
	}

	var ids []int64
	if err := c.Ctx.ReadJSON(&ids); err != nil {
		return web.JsonError(err)
	}
	if err := services.TicketCategoryService.UpdateSort(ids); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}
