package console

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/constants"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/services"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/web"
	"github.com/mlogclub/simple/web/params"
)

type ChannelController struct {
	Ctx iris.Context
}

func (c *ChannelController) AnyList() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionChannelView); err != nil {
		return web.JsonError(err)
	}
	list, paging := services.ChannelService.FindPageByCnd(params.NewPagedSqlCnd(c.Ctx,
		params.QueryFilter{ParamName: "status"},
		params.QueryFilter{ParamName: "name", Op: params.Like},
		params.QueryFilter{ParamName: "channelType"},
		params.QueryFilter{ParamName: "channelId", Op: params.Like},
	).Where("status <> ?", enums.StatusDeleted).Desc("id"))
	results := make([]response.ChannelResponse, 0, len(list))
	for _, item := range list {
		results = append(results, buildChannelResponse(&item))
	}
	return web.JsonData(&web.PageResult{Results: results, Page: paging})
}

func (c *ChannelController) GetBy(id int64) *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionChannelView); err != nil {
		return web.JsonError(err)
	}
	item := services.ChannelService.Get(id)
	if item == nil || item.Status == enums.StatusDeleted {
		return web.JsonErrorMsg("channel not found")
	}
	return web.JsonData(buildChannelResponse(item))
}

func (c *ChannelController) PostCreate() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionChannelCreate)
	if err != nil {
		return web.JsonError(err)
	}
	req := request.CreateChannelRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	item, err := services.ChannelService.CreateChannel(req, operator)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(buildChannelResponse(item))
}

func (c *ChannelController) PostUpdate() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionChannelUpdate)
	if err != nil {
		return web.JsonError(err)
	}
	req := request.UpdateChannelRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.ChannelService.UpdateChannel(req, operator); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *ChannelController) PostUpdate_status() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionChannelUpdate)
	if err != nil {
		return web.JsonError(err)
	}
	req := request.UpdateChannelStatusRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.ChannelService.UpdateStatus(req.ID, req.Status, operator); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *ChannelController) PostDelete() *web.JsonResult {
	operator, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionChannelDelete)
	if err != nil {
		return web.JsonError(err)
	}
	req := request.DeleteChannelRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	if err := services.ChannelService.DeleteChannel(req.ID, operator); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func buildChannelResponse(item *models.Channel) response.ChannelResponse {
	ret := response.BuildChannelResponse(item)
	if item == nil {
		return ret
	}
	if aiAgent := services.AIAgentService.Get(item.AIAgentID); aiAgent != nil {
		ret.AIAgentName = aiAgent.Name
	}
	return ret
}
