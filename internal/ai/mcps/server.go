package mcps

import (
	"fmt"
	"net/http"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func NewHTTPHandler() http.Handler {
	server := newServer()
	return mcp.NewStreamableHTTPHandler(func(_ *http.Request) *mcp.Server {
		return server
	}, &mcp.StreamableHTTPOptions{
		JSONResponse:   true,
		SessionTimeout: 2 * time.Minute,
	})
}

func newServer() *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:       "cs-agent-mcp-server",
		Title:      "CS Agent MCP Server",
		Version:    "v1",
		WebsiteURL: "https://github.com/modelcontextprotocol",
	}, nil)
	if err := registerProviders(server); err != nil {
		panic(fmt.Sprintf("register mcp providers failed: %v", err))
	}
	return server
}
