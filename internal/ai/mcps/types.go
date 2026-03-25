package mcps

type ServerConfig struct {
	Code      string
	Endpoint  string
	TimeoutMS int
	Headers   map[string]string
}

type ServerInfo struct {
	Code      string `json:"code"`
	Enabled   bool   `json:"enabled"`
	Endpoint  string `json:"endpoint"`
	TimeoutMS int    `json:"timeoutMs"`
}

type ConnectionResult struct {
	ServerCode string `json:"serverCode"`
	Endpoint   string `json:"endpoint"`
	Protocol   string `json:"protocol"`
	ServerName string `json:"serverName"`
	Version    string `json:"version"`
}

type ToolInfo struct {
	Name         string `json:"name"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	InputSchema  any    `json:"inputSchema"`
	OutputSchema any    `json:"outputSchema,omitempty"`
}

type ToolResultContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
	Data any    `json:"data,omitempty"`
}

type ToolCallResult struct {
	ServerCode        string              `json:"serverCode"`
	ToolName          string              `json:"toolName"`
	IsError           bool                `json:"isError"`
	Content           []ToolResultContent `json:"content"`
	StructuredContent any                 `json:"structuredContent,omitempty"`
}
