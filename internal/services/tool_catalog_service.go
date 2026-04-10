package services

import (
	"context"
	"slices"
	"strings"

	"cs-agent/internal/ai/mcps"
	"cs-agent/internal/pkg/config"
	"cs-agent/internal/pkg/errorsx"
	"cs-agent/internal/pkg/toolx"
)

var ToolCatalogService = newToolCatalogService()

func newToolCatalogService() *toolCatalogService {
	return &toolCatalogService{}
}

type toolCatalogService struct{}

type MCPToolCatalogItem struct {
	ToolCode     string
	ServerCode   string
	ToolName     string
	SourceType   string
	AutoInjected bool
	Title        string
	Description  string
	InputSchema  any
	OutputSchema any
}

func (s *toolCatalogService) ListMCPTools(ctx context.Context) ([]MCPToolCatalogItem, error) {
	cfg := config.Current()
	ret := make([]MCPToolCatalogItem, 0, 3)
	ret = append(ret, MCPToolCatalogItem{
		ToolCode:     toolx.GraphCreateTicketConfirmToolCode,
		ServerCode:   toolx.GraphToolCatalogServerCode,
		ToolName:     toolx.GraphCreateTicketConfirmToolName,
		SourceType:   toolx.GraphToolCatalogServerCode,
		AutoInjected: false,
		Title:        toolx.GraphCreateTicketConfirmToolTitle,
		Description:  toolx.GraphCreateTicketConfirmToolDescription,
	})
	ret = append(ret, MCPToolCatalogItem{
		ToolCode:     toolx.GraphHandoffConversationToolCode,
		ServerCode:   toolx.GraphToolCatalogServerCode,
		ToolName:     toolx.GraphHandoffConversationToolName,
		SourceType:   toolx.GraphToolCatalogServerCode,
		AutoInjected: false,
		Title:        toolx.GraphHandoffConversationToolTitle,
		Description:  toolx.GraphHandoffConversationToolDescription,
	})
	if !cfg.MCP.Enabled {
		return ret, nil
	}
	ret = append(ret, MCPToolCatalogItem{
		ToolCode:     toolx.BuiltinToolSearchToolCode,
		ServerCode:   toolx.BuiltinToolCatalogServerCode,
		ToolName:     toolx.BuiltinToolSearchToolName,
		SourceType:   toolx.BuiltinToolCatalogServerCode,
		AutoInjected: true,
		Title:        toolx.BuiltinToolSearchToolTitle,
		Description:  toolx.BuiltinToolSearchToolDescription,
	})
	serverCodes := make([]string, 0, len(cfg.MCP.Servers))
	for serverCode, server := range cfg.MCP.Servers {
		if !server.Enabled {
			continue
		}
		serverCodes = append(serverCodes, serverCode)
	}
	slices.Sort(serverCodes)
	for _, serverCode := range serverCodes {
		tools, err := mcps.Runtime.ListTools(ctx, serverCode)
		if err != nil {
			return nil, err
		}
		for _, item := range tools {
			ret = append(ret, MCPToolCatalogItem{
				ToolCode:     toolx.BuildMCPToolCode(serverCode, item.Name),
				ServerCode:   serverCode,
				ToolName:     strings.TrimSpace(item.Name),
				SourceType:   "mcp",
				AutoInjected: false,
				Title:        strings.TrimSpace(item.Title),
				Description:  strings.TrimSpace(item.Description),
				InputSchema:  item.InputSchema,
				OutputSchema: item.OutputSchema,
			})
		}
	}
	return ret, nil
}

func (s *toolCatalogService) ValidateMCPToolCode(toolCode string) error {
	return s.ValidateToolCode(toolCode)
}

func (s *toolCatalogService) ValidateToolCode(toolCode string) error {
	cfg := config.Current()
	toolCode = strings.TrimSpace(toolCode)
	if toolCode == "" {
		return errorsx.InvalidParam("toolCode不能为空")
	}
	switch toolCode {
	case toolx.BuiltinToolSearchToolCode, toolx.BuiltinCreateTicketConfirmToolCode, toolx.GraphCreateTicketConfirmToolCode, toolx.GraphHandoffConversationToolCode:
		return nil
	}
	serverCode, toolName := toolx.SplitMCPToolCode(toolCode)
	if serverCode == "" || toolName == "" {
		return errorsx.InvalidParam("toolCode格式不合法")
	}
	if !cfg.MCP.Enabled {
		return errorsx.InvalidParam("MCP未启用")
	}
	server, ok := cfg.MCP.Servers[serverCode]
	if !ok || !server.Enabled {
		return errorsx.InvalidParam("toolCode 绑定的 MCP 服务不存在或未启用")
	}
	return nil
}
