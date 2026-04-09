package registry

import (
	"cs-agent/internal/models"

	einotool "github.com/cloudwego/eino/components/tool"
)

type Context struct {
	Conversation *models.Conversation
	AIAgent      *models.AIAgent
	AIConfig     *models.AIConfig
	UserMessage  *models.Message
}

type ToolSet struct {
	Tools     []einotool.BaseTool
	ToolCodes map[string]string
}

type Tool interface {
	Name() string
	Code() string
	Enabled(ctx Context) bool
	Build(ctx Context) (einotool.BaseTool, error)
}
