package request

type MCPServerDebugRequest struct {
	ServerCode string `json:"serverCode"`
}

type MCPCallToolRequest struct {
	ServerCode string         `json:"serverCode"`
	ToolName   string         `json:"toolName"`
	Arguments  map[string]any `json:"arguments"`
}
