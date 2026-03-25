package mcps

import (
	"fmt"
	"net/http"
	"time"

	"cs-agent/internal/pkg/config"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func NewHTTPHandler(cfg *config.Config) http.Handler {
	server := newServer(cfg)
	return mcp.NewStreamableHTTPHandler(func(_ *http.Request) *mcp.Server {
		return server
	}, &mcp.StreamableHTTPOptions{
		JSONResponse:   true,
		SessionTimeout: 2 * time.Minute,
	})
}

func newServer(cfg *config.Config) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:       "cs-agent-mcp-server",
		Title:      "CS Agent MCP Server",
		Version:    "v1",
		WebsiteURL: "https://github.com/modelcontextprotocol",
	}, nil)
	if err := registerProviders(server, ToolContext{Config: cfg}); err != nil {
		panic(fmt.Sprintf("register mcp providers failed: %v", err))
	}
	return server
}
