package console

import (
	"cs-agent/internal/pkg/constants"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/services"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/web"
	"github.com/mlogclub/simple/web/params"
)

type AIConfigController struct {
	Ctx iris.Context
}

func (c *AIConfigController) AnyList() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionAIConfigView); err != nil {
		return web.JsonError(err)
	}
	list, paging := services.AIConfigService.FindPageByCnd(params.NewPagedSqlCnd(c.Ctx,
		params.QueryFilter{ParamName: "status"},
		params.QueryFilter{ParamName: "provider"},
		params.QueryFilter{ParamName: "modelType"},
		params.QueryFilter{ParamName: "name", Op: params.Like},
		params.QueryFilter{ParamName: "modelName", Op: params.Like},
	).Desc("sort_no").Desc("id"))
	results := make([]response.AIConfigResponse, 0, len(list))
	for _, item := range list {
		results = append(results, response.BuildAIConfigResponse(&item))
	}
	return web.JsonData(&web.PageResult{Results: results, Page: paging})
}

func (c *AIConfigController) AnyList_all() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionAIConfigView); err != nil {
		return web.JsonError(err)
	}

	list := services.AIConfigService.Find(params.NewSqlCnd(c.Ctx,
		params.QueryFilter{ParamName: "modelType"},
	).Eq("status", enums.StatusOk).Desc("sort_no").Desc("id"))

	results := make([]response.AIConfigResponse, 0, len(list))
	for _, item := range list {
		results = append(results, response.BuildAIConfigResponse(&item))
	}
	return web.JsonData(results)
}

func (c *AIConfigController) GetBy(id int64) *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionAIConfigView); err != nil {
		return web.JsonError(err)
	}

	item := services.AIConfigService.Get(id)
	if item == nil || item.Status == enums.StatusDeleted {
		return web.JsonErrorMsg("AI配置不存在")
	}
	return web.JsonData(response.BuildAIConfigResponse(item))
}

func (c *AIConfigController) PostCreate() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionAIConfigCreate)
	if err != nil {
		return web.JsonError(err)
	}

	req := request.CreateAIConfigRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	item, err := services.AIConfigService.CreateAIConfig(req, operator)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(response.BuildAIConfigResponse(item))
}

func (c *AIConfigController) PostUpdate() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionAIConfigUpdate)
	if err != nil {
		return web.JsonError(err)
	}

	req := request.UpdateAIConfigRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.AIConfigService.UpdateAIConfig(req, operator); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *AIConfigController) PostDelete() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionAIConfigDelete); err != nil {
		return web.JsonError(err)
	}

	req := request.DeleteAIConfigRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.AIConfigService.DeleteAIConfig(req.ID); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *AIConfigController) PostUpdate_status() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionAIConfigUpdate)
	if err != nil {
		return web.JsonError(err)
	}

	req := request.UpdateAIConfigStatusRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.AIConfigService.UpdateStatus(req.ID, req.Status, operator); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *AIConfigController) PostUpdate_sort() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionAIConfigUpdate); err != nil {
		return web.JsonError(err)
	}

	var ids []int64
	if err := c.Ctx.ReadJSON(&ids); err != nil {
		return web.JsonError(err)
	}
	if err := services.AIConfigService.UpdateSort(ids); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}
