package console

import (
	"context"

	"cs-agent/internal/pkg/constants"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/services"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/web"
	"github.com/mlogclub/simple/web/params"
)

type MCPController struct {
	Ctx iris.Context
}

func (c *MCPController) AnyList_servers() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionMCPView); err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(response.BuildMCPServerInfoResponses(services.MCPDebugService.ListServers()))
}

func (c *MCPController) PostTest_connection() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionMCPView); err != nil {
		return web.JsonError(err)
	}
	req := request.MCPServerDebugRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	result, err := services.MCPDebugService.TestConnection(context.Background(), req.ServerCode)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(response.BuildMCPConnectionResponse(result))
}

func (c *MCPController) PostList_tools() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionMCPView); err != nil {
		return web.JsonError(err)
	}
	req := request.MCPServerDebugRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	result, err := services.MCPDebugService.ListTools(context.Background(), req.ServerCode)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(response.BuildMCPToolInfoResponses(result))
}

func (c *MCPController) PostCall_tool() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionMCPCall); err != nil {
		return web.JsonError(err)
	}
	req := request.MCPCallToolRequest{}
	if err := params.ReadJSON(c.Ctx, &req); err != nil {
		return web.JsonError(err)
	}
	result, err := services.MCPDebugService.CallTool(context.Background(), req.ServerCode, req.ToolName, req.Arguments)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(response.BuildMCPCallToolResponse(result))
}
