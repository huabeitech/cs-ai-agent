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

func (c *MCPController) AnyCatalog() *web.JsonResult {
	if _, err := services.AuthService.RequirePermission(c.Ctx, constants.PermissionMCPView); err != nil {
		return web.JsonError(err)
	}
	items, err := services.ToolCatalogService.ListMCPTools(context.Background())
	if err != nil {
		return web.JsonError(err)
	}
	ret := make([]response.MCPToolCatalogResponse, 0, len(items))
	for _, item := range items {
		ret = append(ret, response.MCPToolCatalogResponse{
			ToolCode:     item.ToolCode,
			ServerCode:   item.ServerCode,
			ToolName:     item.ToolName,
			SourceType:   item.SourceType,
			AutoInjected: item.AutoInjected,
			Title:        item.Title,
			Description:  item.Description,
			InputSchema:  item.InputSchema,
			OutputSchema: item.OutputSchema,
		})
	}
	return web.JsonData(ret)
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
