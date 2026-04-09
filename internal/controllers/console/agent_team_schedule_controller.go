package console

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/constants"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/services"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/web"
	"github.com/mlogclub/simple/web/params"
)

type AgentTeamScheduleController struct {
	Ctx iris.Context
}

func (c *AgentTeamScheduleController) AnyList() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionAgentTeamScheduleView); err != nil {
		return web.JsonError(err)
	}
	cnd := params.NewPagedSqlCnd(c.Ctx,
		params.QueryFilter{ParamName: "teamId"},
	).Desc("start_at").Desc("id")
	list, paging := services.AgentTeamScheduleService.FindPageByCnd(cnd)
	results := make([]response.AgentTeamScheduleResponse, 0, len(list))
	for _, item := range list {
		results = append(results, buildAgentTeamScheduleResponse(&item))
	}
	return web.JsonData(&web.PageResult{Results: results, Page: paging})
}

func (c *AgentTeamScheduleController) GetBy(id int64) *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionAgentTeamScheduleView); err != nil {
		return web.JsonError(err)
	}
	item := services.AgentTeamScheduleService.Get(id)
	if item == nil {
		return web.JsonErrorMsg("客服组排班不存在")
	}
	return web.JsonData(buildAgentTeamScheduleResponse(item))
}

func (c *AgentTeamScheduleController) PostCreate() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionAgentTeamScheduleCreate)
	if err != nil {
		return web.JsonError(err)
	}
	req := request.CreateAgentTeamScheduleRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	item, err := services.AgentTeamScheduleService.CreateAgentTeamSchedule(req, operator)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(buildAgentTeamScheduleResponse(item))
}

func (c *AgentTeamScheduleController) PostUpdate() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionAgentTeamScheduleUpdate)
	if err != nil {
		return web.JsonError(err)
	}
	req := request.UpdateAgentTeamScheduleRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.AgentTeamScheduleService.UpdateAgentTeamSchedule(req, operator); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *AgentTeamScheduleController) PostDelete() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionAgentTeamScheduleDelete); err != nil {
		return web.JsonError(err)
	}
	req := request.DeleteAgentTeamScheduleRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.AgentTeamScheduleService.DeleteAgentTeamSchedule(req.ID); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func buildAgentTeamScheduleResponse(item *models.AgentTeamSchedule) response.AgentTeamScheduleResponse {
	ret := response.AgentTeamScheduleResponse{
		ID:         item.ID,
		TeamID:     item.TeamID,
		StartAt:    item.StartAt.Format("2006-01-02 15:04:05"),
		EndAt:      item.EndAt.Format("2006-01-02 15:04:05"),
		SourceType: item.SourceType,
		Remark:     item.Remark,
	}
	if team := services.AgentTeamService.Get(item.TeamID); team != nil {
		ret.TeamName = team.Name
	}
	return ret
}
