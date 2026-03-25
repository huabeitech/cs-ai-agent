package mcps

import "github.com/modelcontextprotocol/go-sdk/mcp"

type ToolProvider interface {
	Name() string
	Register(server *mcp.Server, ctx ToolContext) error
}

func defaultProviders() []ToolProvider {
	return []ToolProvider{
		NewSystemToolProvider(),
		// 在这里注册其他的 ToolProvider
	}
}

func registerProviders(server *mcp.Server, ctx ToolContext) error {
	for _, provider := range defaultProviders() {
		if err := provider.Register(server, ctx); err != nil {
			return err
		}
	}
	return nil
}
