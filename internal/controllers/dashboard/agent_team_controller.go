package dashboard

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/constants"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/services"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web"
	"github.com/mlogclub/simple/web/params"
)

type AgentTeamController struct {
	Ctx iris.Context
}

func (c *AgentTeamController) AnyList() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionAgentTeamView); err != nil {
		return web.JsonError(err)
	}
	cnd := params.NewSqlCnd(c.Ctx,
		params.QueryFilter{ParamName: "status"},
		params.QueryFilter{ParamName: "leaderUserId"},
		params.QueryFilter{ParamName: "name", Op: params.Like},
	).Desc("id")
	if _, ok := params.Get(c.Ctx, "status"); !ok {
		cnd.Where("status <> ?", enums.StatusDeleted)
	}
	list := services.AgentTeamService.Find(cnd)
	results := make([]response.AgentTeamResponse, 0, len(list))
	for _, item := range list {
		results = append(results, buildAgentTeamResponse(&item))
	}
	return web.JsonData(results)
}

func (c *AgentTeamController) GetList_all() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionAgentTeamView); err != nil {
		return web.JsonError(err)
	}
	list := services.AgentTeamService.Find(sqls.NewCnd().Eq("status", enums.StatusOk))
	results := make([]response.AgentTeamResponse, 0, len(list))
	for _, item := range list {
		results = append(results, buildAgentTeamResponse(&item))
	}
	return web.JsonData(results)
}

func (c *AgentTeamController) GetBy(id int64) *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionAgentTeamView); err != nil {
		return web.JsonError(err)
	}
	item := services.AgentTeamService.Get(id)
	if item == nil || item.Status == enums.StatusDeleted {
		return web.JsonErrorMsg("客服组不存在")
	}
	return web.JsonData(buildAgentTeamResponse(item))
}

func (c *AgentTeamController) PostCreate() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionAgentTeamCreate)
	if err != nil {
		return web.JsonError(err)
	}
	req := request.CreateAgentTeamRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	item, err := services.AgentTeamService.CreateAgentTeam(req, operator)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(buildAgentTeamResponse(item))
}

func (c *AgentTeamController) PostUpdate() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionAgentTeamUpdate)
	if err != nil {
		return web.JsonError(err)
	}
	req := request.UpdateAgentTeamRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.AgentTeamService.UpdateAgentTeam(req, operator); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *AgentTeamController) PostDelete() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionAgentTeamDelete)
	if err != nil {
		return web.JsonError(err)
	}
	req := request.DeleteAgentTeamRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.AgentTeamService.DeleteAgentTeam(req.ID, operator); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func buildAgentTeamResponse(item *models.AgentTeam) response.AgentTeamResponse {
	ret := response.AgentTeamResponse{
		ID:           item.ID,
		Name:         item.Name,
		LeaderUserID: item.LeaderUserID,
		Status:       item.Status,
		Description:  item.Description,
		Remark:       item.Remark,
	}
	if user := services.UserService.Get(item.LeaderUserID); user != nil {
		ret.LeaderUsername = user.Username
		ret.LeaderNickname = user.Nickname
	}
	return ret
}
