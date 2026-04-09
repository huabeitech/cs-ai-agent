package mcps

import (
	"cs-agent/internal/ai/mcps/providers"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func defaultProviders() []providers.ToolProvider {
	return []providers.ToolProvider{
		providers.NewSystemToolProvider(),
		// 在这里注册其他的 ToolProvider
	}
}

func registerProviders(server *mcp.Server) error {
	for _, provider := range defaultProviders() {
		if err := provider.Register(server); err != nil {
			return err
		}
	}
	return nil
}
