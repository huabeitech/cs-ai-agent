package providers

import (
	"context"
	"cs-agent/internal/pkg/config"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type systemToolProvider struct{}

func NewSystemToolProvider() ToolProvider {
	return &systemToolProvider{}
}

func (p *systemToolProvider) Name() string {
	return "system"
}

func (p *systemToolProvider) Register(server *mcp.Server) error {
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "server_time",
			Description: "获取当前服务端时间，可选传入时区。",
		},
		func(_ context.Context, _ *mcp.CallToolRequest, args serverTimeArgs) (*mcp.CallToolResult, map[string]any, error) {
			loc := time.Local
			timezone := args.Timezone
			if timezone == "" {
				timezone = "Local"
			} else if loaded, err := time.LoadLocation(timezone); err == nil {
				loc = loaded
			}
			now := time.Now().In(loc)
			return nil, map[string]any{
				"timezone":  timezone,
				"timestamp": now.Format("2006-01-02 15:04:05"),
				"unix":      now.Unix(),
			}, nil
		},
	)

	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "service_info",
			Description: "查看当前 cs-agent 服务的基础运行信息。",
		},
		func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, map[string]any, error) {
			cfg := config.Current()
			return nil, map[string]any{
				"name":        "cs-agent",
				"version":     "v1",
				"mcpPath":     "/api/mcp",
				"port":        cfg.Server.Port,
				"mcpEnabled":  cfg.MCP.Enabled,
				"vectorDb":    cfg.VectorDB.Type,
				"storageType": cfg.Storage.Default,
			}, nil
		},
	)
	return nil
}

type serverTimeArgs struct {
	Timezone string `json:"timezone,omitempty" jsonschema:"可选时区名称，例如 Asia/Shanghai 或 UTC"`
}
