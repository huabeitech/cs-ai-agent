package providers

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ToolProvider interface {
	Name() string
	Register(server *mcp.Server) error
}
