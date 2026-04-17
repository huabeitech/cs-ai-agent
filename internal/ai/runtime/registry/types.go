package registry

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/toolx"

	einotool "github.com/cloudwego/eino/components/tool"
)

type Context struct {
	Conversation     models.Conversation
	AIAgent          models.AIAgent
	AIConfig         models.AIConfig
	UserMessage      models.Message
	AllowedToolCodes []string
}

type ToolMetadata struct {
	ToolCode   string
	ServerCode string
	ToolName   string
	SourceType enums.ToolSourceType
}

// ToolSet 描述当前运行时可直接挂载到 ToolsNode 上的固定工具集合。
//
// 这里不承载通过 tool_search 动态暴露的 MCP 工具；动态工具仍由 engine 层单独装配。
type ToolSet struct {
	// StaticTools 为固定可见工具实例，例如 Graph Tool。
	StaticTools []einotool.BaseTool
	// StaticToolCodes 为固定工具的 modelName -> toolCode 映射，用于 trace 和运行日志归因。
	StaticToolCodes map[string]string
	// StaticToolMetadata 为固定工具的 modelName -> metadata 映射，用于 trace、运行日志和后续装配。
	StaticToolMetadata map[string]ToolMetadata
}

type Tool interface {
	Spec() toolx.ToolSpec
	Name() string
	Code() string
	Enabled(ctx Context) bool
	Build(ctx Context) (einotool.BaseTool, error)
}
