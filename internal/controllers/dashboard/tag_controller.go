package dashboard

import (
	"cs-agent/internal/builders"
	"cs-agent/internal/pkg/constants"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/services"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/web"
	"github.com/mlogclub/simple/web/params"
)

type TagController struct {
	Ctx iris.Context
}

func (c *TagController) AnyList() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTagView); err != nil {
		return web.JsonError(err)
	}
	list, paging := services.TagService.FindPageByCnd(params.NewPagedSqlCnd(c.Ctx,
		params.QueryFilter{ParamName: "status"},
		params.QueryFilter{ParamName: "parentId"},
		params.QueryFilter{ParamName: "name", Op: params.Like},
	).Asc("sort_no").Desc("id"))
	results := builders.BuildTagResponses(list)
	return web.JsonData(&web.PageResult{Results: results, Page: paging})
}

func (c *TagController) GetList_all() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTagView); err != nil {
		return web.JsonError(err)
	}
	list := services.TagService.FindAll()
	results := builders.BuildTagTreeResponses(list)
	return web.JsonData(results)
}

func (c *TagController) GetBy(id int64) *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTagView); err != nil {
		return web.JsonError(err)
	}

	item := services.TagService.Get(id)
	if item == nil {
		return web.JsonErrorMsg("标签不存在")
	}
	result := builders.BuildTagResponse(item)
	return web.JsonData(&result)
}

func (c *TagController) PostCreate() *web.JsonResult {
	user, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTagCreate)
	if err != nil {
		return web.JsonError(err)
	}

	req := request.CreateTagRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	item, err := services.TagService.CreateTag(req, user)
	if err != nil {
		return web.JsonError(err)
	}
	result := builders.BuildTagResponse(item)
	return web.JsonData(&result)
}

func (c *TagController) PostUpdate() *web.JsonResult {
	user, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTagUpdate)
	if err != nil {
		return web.JsonError(err)
	}

	req := request.UpdateTagRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.TagService.UpdateTag(req, user); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *TagController) PostDelete() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTagDelete); err != nil {
		return web.JsonError(err)
	}

	req := request.DeleteTagRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.TagService.DeleteTag(req.ID); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *TagController) PostUpdate_sort() *web.JsonResult {
	var ids []int64
	if err := c.Ctx.ReadJSON(&ids); err != nil {
		return web.JsonError(err)
	}
	if err := services.TagService.UpdateSort(ids); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *TagController) PostUpdate_status() *web.JsonResult {
	user, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionTagUpdate)
	if err != nil {
		return web.JsonError(err)
	}

	req := request.UpdateTagStatusRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.TagService.UpdateStatus(req.ID, req.Status, user); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}
