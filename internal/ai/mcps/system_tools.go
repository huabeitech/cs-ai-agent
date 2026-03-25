package mcps

import (
	"context"
	"cs-agent/internal/pkg/enums"
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

func (p *systemToolProvider) Register(server *mcp.Server, ctx ToolContext) error {
	registerServerTimeTool(server)
	registerServiceInfoTool(server, ctx)
	return nil
}

type serverTimeArgs struct {
	Timezone string `json:"timezone,omitempty" jsonschema:"可选时区名称，例如 Asia/Shanghai 或 UTC"`
}

type serverTimeResult struct {
	Timezone  string `json:"timezone"`
	Timestamp string `json:"timestamp"`
	Unix      int64  `json:"unix"`
}

func registerServerTimeTool(server *mcp.Server) {
	mcp.AddTool(server,
		&mcp.Tool{
			Name:        "server_time",
			Description: "获取当前服务端时间，可选传入时区。",
		},
		func(_ context.Context, _ *mcp.CallToolRequest, args serverTimeArgs) (*mcp.CallToolResult, serverTimeResult, error) {
			loc := time.Local
			timezone := args.Timezone
			if timezone == "" {
				timezone = "Local"
			} else if loaded, err := time.LoadLocation(timezone); err == nil {
				loc = loaded
			}
			now := time.Now().In(loc)
			return nil, serverTimeResult{
				Timezone:  timezone,
				Timestamp: now.Format("2006-01-02 15:04:05"),
				Unix:      now.Unix(),
			}, nil
		},
	)
}

type serviceInfoArgs struct{}

type serviceInfoResult struct {
	Name        string              `json:"name"`
	Version     string              `json:"version"`
	MCPPath     string              `json:"mcpPath"`
	Port        int                 `json:"port"`
	MCPEnabled  bool                `json:"mcpEnabled"`
	VectorDB    string              `json:"vectorDb"`
	StorageType enums.AssetProvider `json:"storageType"`
}

func registerServiceInfoTool(server *mcp.Server, toolCtx ToolContext) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "service_info",
		Description: "查看当前 cs-agent 服务的基础运行信息。",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ serviceInfoArgs) (*mcp.CallToolResult, serviceInfoResult, error) {
		cfg := toolCtx.Config
		result := serviceInfoResult{
			Name:        "cs-agent",
			Version:     "v1",
			MCPPath:     "/api/mcp",
			MCPEnabled:  cfg != nil && cfg.MCP.Enabled,
			VectorDB:    "",
			StorageType: "",
		}
		if cfg != nil {
			result.Port = cfg.Server.Port
			result.VectorDB = cfg.VectorDB.Type
			result.StorageType = cfg.Storage.Default
		}
		return nil, result, nil
	})
}
