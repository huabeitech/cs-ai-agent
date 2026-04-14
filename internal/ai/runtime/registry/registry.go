package registry

import (
	"strings"

	"cs-agent/internal/pkg/toolx"

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
		StaticTools:        make([]einotool.BaseTool, 0, len(r.tools)),
		StaticToolCodes:    make(map[string]string),
		StaticToolMetadata: make(map[string]ToolMetadata),
	}
	allowedToolCodes := makeAllowedToolCodeSet(ctx.AllowedToolCodes)
	for _, toolDef := range r.tools {
		if toolDef == nil || !toolDef.Enabled(ctx) {
			continue
		}
		spec := toolDef.Spec()
		toolCode := strings.TrimSpace(spec.Code)
		if toolCode == "" {
			toolCode = strings.TrimSpace(toolDef.Code())
		}
		if len(allowedToolCodes) > 0 && !isAllowedToolCode(toolCode, allowedToolCodes) {
			continue
		}
		tool, err := toolDef.Build(ctx)
		if err != nil {
			return nil, err
		}
		if tool == nil {
			continue
		}
		toolName := strings.TrimSpace(spec.Name)
		if toolName == "" {
			toolName = strings.TrimSpace(toolDef.Name())
		}
		if toolName == "" || toolCode == "" {
			continue
		}
		ret.StaticTools = append(ret.StaticTools, tool)
		ret.StaticToolCodes[toolName] = toolCode
		resolvedMetadata := toolx.ResolveToolMetadata(toolCode, toolName)
		if strings.TrimSpace(resolvedMetadata.ServerCode) == "" {
			resolvedMetadata.ServerCode = strings.TrimSpace(spec.ServerCode)
		}
		if strings.TrimSpace(resolvedMetadata.ToolName) == "" {
			resolvedMetadata.ToolName = toolName
		}
		if strings.TrimSpace(resolvedMetadata.SourceType) == "" {
			resolvedMetadata.SourceType = strings.TrimSpace(spec.SourceType)
		}
		ret.StaticToolMetadata[toolName] = ToolMetadata{
			ToolCode:   resolvedMetadata.ToolCode,
			ServerCode: resolvedMetadata.ServerCode,
			ToolName:   resolvedMetadata.ToolName,
			SourceType: resolvedMetadata.SourceType,
		}
	}
	return ret, nil
}

func isAllowedToolCode(toolCode string, allowedToolCodes map[string]struct{}) bool {
	if len(allowedToolCodes) == 0 {
		return true
	}
	toolCode = toolx.NormalizeToolCodeAlias(strings.TrimSpace(toolCode))
	if _, ok := allowedToolCodes[toolCode]; ok {
		return true
	}
	if toolx.IsAlwaysAllowedToolCode(toolCode) {
		return true
	}
	return toolx.IsImpliedAllowedToolCode(toolCode, allowedToolCodes)
}

func makeAllowedToolCodeSet(input []string) map[string]struct{} {
	if len(input) == 0 {
		return nil
	}
	ret := make(map[string]struct{}, len(input))
	for _, item := range input {
		item = toolx.NormalizeToolCodeAlias(strings.TrimSpace(item))
		if item == "" {
			continue
		}
		ret[item] = struct{}{}
	}
	return ret
}
