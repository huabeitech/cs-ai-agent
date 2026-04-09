package registry

import (
	"strings"

	einotool "github.com/cloudwego/eino/components/tool"
)

type Registry struct {
	tools []Tool
}

func NewRegistry(tools ...Tool) *Registry {
	return &Registry{
		tools: tools,
	}
}

func (r *Registry) Resolve(ctx Context) (*ToolSet, error) {
	ret := &ToolSet{
		Tools:     make([]einotool.BaseTool, 0, len(r.tools)),
		ToolCodes: make(map[string]string),
	}
	for _, toolDef := range r.tools {
		if toolDef == nil || !toolDef.Enabled(ctx) {
			continue
		}
		tool, err := toolDef.Build(ctx)
		if err != nil {
			return nil, err
		}
		if tool == nil {
			continue
		}
		toolName := strings.TrimSpace(toolDef.Name())
		toolCode := strings.TrimSpace(toolDef.Code())
		if toolName == "" || toolCode == "" {
			continue
		}
		ret.Tools = append(ret.Tools, tool)
		ret.ToolCodes[toolName] = toolCode
	}
	return ret, nil
}
