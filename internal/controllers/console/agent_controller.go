package console

import (
	"cs-agent/internal/builders"
	"cs-agent/internal/pkg/constants"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/services"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/web"
	"github.com/mlogclub/simple/web/params"
)

type AgentController struct {
	Ctx iris.Context
}

func (c *AgentController) AnyList() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionAgentView); err != nil {
		return web.JsonError(err)
	}
	list, paging := services.AgentProfileService.FindPageByCnd(params.NewPagedSqlCnd(c.Ctx,
		params.QueryFilter{ParamName: "userId"},
		params.QueryFilter{ParamName: "teamId"},
		params.QueryFilter{ParamName: "serviceStatus"},
		params.QueryFilter{ParamName: "agentCode", Op: params.Like},
		params.QueryFilter{ParamName: "displayName", Op: params.Like},
	).Desc("id"))
	results := builders.BuildAgentProfileList(list)
	return web.JsonData(&web.PageResult{Results: results, Page: paging})
}

func (c *AgentController) GetList_all() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionAgentView); err != nil {
		return web.JsonError(err)
	}
	list := services.AgentProfileService.Find(params.NewPagedSqlCnd(c.Ctx,
		params.QueryFilter{ParamName: "userId"},
		params.QueryFilter{ParamName: "teamId"},
		params.QueryFilter{ParamName: "serviceStatus"},
		params.QueryFilter{ParamName: "agentCode", Op: params.Like},
	).Desc("id"))

	return web.JsonData(builders.BuildAgentProfileList(list))
}

func (c *AgentController) GetBy(id int64) *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionAgentView); err != nil {
		return web.JsonError(err)
	}
	item := services.AgentProfileService.Get(id)
	if item == nil {
		return web.JsonErrorMsg("客服档案不存在")
	}
	return web.JsonData(builders.BuildAgentProfileResponse(item))
}

func (c *AgentController) PostCreate() *web.JsonResult {
	user, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionAgentCreate)
	if err != nil {
		return web.JsonError(err)
	}
	req := request.CreateAgentProfileRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	item, err := services.AgentProfileService.CreateAgentProfile(req, user)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(builders.BuildAgentProfileResponse(item))
}

func (c *AgentController) PostUpdate() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionAgentUpdate); err != nil {
		return web.JsonError(err)
	}
	req := request.UpdateAgentProfileRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.AgentProfileService.UpdateAgentProfile(req, services.AuthService.GetAuthPrincipal(c.Ctx)); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *AgentController) PostDelete() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionAgentDelete); err != nil {
		return web.JsonError(err)
	}
	req := request.DeleteAgentProfileRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.AgentProfileService.DeleteAgentProfile(req.ID); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}
