package response

import "cs-agent/internal/ai/mcps"

type MCPConnectionResponse struct {
	ServerCode string `json:"serverCode"`
	Endpoint   string `json:"endpoint"`
	Protocol   string `json:"protocol"`
	ServerName string `json:"serverName"`
	Version    string `json:"version"`
}

func BuildMCPConnectionResponse(item *mcps.ConnectionResult) *MCPConnectionResponse {
	if item == nil {
		return nil
	}
	return &MCPConnectionResponse{
		ServerCode: item.ServerCode,
		Endpoint:   item.Endpoint,
		Protocol:   item.Protocol,
		ServerName: item.ServerName,
		Version:    item.Version,
	}
}

type MCPServerInfoResponse struct {
	Code      string `json:"code"`
	Enabled   bool   `json:"enabled"`
	Endpoint  string `json:"endpoint"`
	TimeoutMS int    `json:"timeoutMs"`
}

func BuildMCPServerInfoResponses(items []mcps.ServerInfo) []MCPServerInfoResponse {
	ret := make([]MCPServerInfoResponse, 0, len(items))
	for _, item := range items {
		ret = append(ret, MCPServerInfoResponse{
			Code:      item.Code,
			Enabled:   item.Enabled,
			Endpoint:  item.Endpoint,
			TimeoutMS: item.TimeoutMS,
		})
	}
	return ret
}

type MCPToolInfoResponse struct {
	Name         string `json:"name"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	InputSchema  any    `json:"inputSchema"`
	OutputSchema any    `json:"outputSchema,omitempty"`
}

func BuildMCPToolInfoResponses(items []mcps.ToolInfo) []MCPToolInfoResponse {
	ret := make([]MCPToolInfoResponse, 0, len(items))
	for _, item := range items {
		ret = append(ret, MCPToolInfoResponse{
			Name:         item.Name,
			Title:        item.Title,
			Description:  item.Description,
			InputSchema:  item.InputSchema,
			OutputSchema: item.OutputSchema,
		})
	}
	return ret
}

type MCPToolResultContentResponse struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
	Data any    `json:"data,omitempty"`
}

type MCPCallToolResponse struct {
	ServerCode        string                         `json:"serverCode"`
	ToolName          string                         `json:"toolName"`
	IsError           bool                           `json:"isError"`
	Content           []MCPToolResultContentResponse `json:"content"`
	StructuredContent any                            `json:"structuredContent,omitempty"`
}

func BuildMCPCallToolResponse(item *mcps.ToolCallResult) *MCPCallToolResponse {
	if item == nil {
		return nil
	}
	content := make([]MCPToolResultContentResponse, 0, len(item.Content))
	for _, c := range item.Content {
		content = append(content, MCPToolResultContentResponse{
			Type: c.Type,
			Text: c.Text,
			Data: c.Data,
		})
	}
	return &MCPCallToolResponse{
		ServerCode:        item.ServerCode,
		ToolName:          item.ToolName,
		IsError:           item.IsError,
		Content:           content,
		StructuredContent: item.StructuredContent,
	}
}
