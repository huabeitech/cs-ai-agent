package factory

import (
	"context"
	"strings"

	"cs-agent/internal/ai/mcps"
	impladapter "cs-agent/internal/ai/runtime/internal/impl/adapter"
	runtimetooling "cs-agent/internal/ai/runtime/tooling"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/toolx"

	einotool "github.com/cloudwego/eino/components/tool"
)

type ToolFactory struct{}

func NewToolFactory() *ToolFactory {
	return &ToolFactory{}
}

func (f *ToolFactory) BuildMCPTools(aiAgent models.AIAgent) ([]runtimetooling.MCPToolDefinition, error) {
	raw, err := toolx.ParseAgentMCPToolsJSON(aiAgent.AllowedMCPTools)
	if err != nil {
		return nil, err
	}
	ret := make([]runtimetooling.MCPToolDefinition, 0, len(raw))
	for _, item := range raw {
		toolCode := strings.TrimSpace(item.ToolCode)
		toolCode = toolx.NormalizeToolCodeAlias(toolCode)
		if toolCode == "" {
			toolCode = toolx.BuildMCPToolCode(item.ServerCode, item.ToolName)
		}
		if toolx.ResolveToolSourceType(toolCode) != enums.ToolSourceTypeMCP {
			continue
		}
		serverCode, toolName := toolx.SplitMCPToolCode(toolCode)
		if serverCode == "" || toolName == "" {
			continue
		}
		definition := runtimetooling.MCPToolDefinition{
			ToolCode:    toolCode,
			ServerCode:  serverCode,
			ToolName:    toolName,
			Title:       strings.TrimSpace(item.Title),
			Description: strings.TrimSpace(item.Description),
			FixedArgs:   cloneStringMap(item.Arguments),
		}
		definition.ModelName = runtimetooling.BuildModelToolName(definition)
		ret = append(ret, definition)
	}
	return ret, nil
}

func (f *ToolFactory) BuildBaseTools(ctx context.Context, aiAgent models.AIAgent) ([]einotool.BaseTool, error) {
	definitions, err := f.BuildMCPTools(aiAgent)
	if err != nil {
		return nil, err
	}
	return f.BuildBaseToolsByDefinitions(ctx, definitions)
}

func (f *ToolFactory) BuildBaseToolsByDefinitions(ctx context.Context, definitions []runtimetooling.MCPToolDefinition) ([]einotool.BaseTool, error) {
	if len(definitions) == 0 {
		return nil, nil
	}
	metadataByCode, err := f.loadToolMetadata(ctx, definitions)
	if err != nil {
		return nil, err
	}
	ret := make([]einotool.BaseTool, 0, len(definitions))
	for _, item := range definitions {
		ret = append(ret, impladapter.NewMCPTool(item, metadataByCode[item.ToolCode]))
	}
	return ret, nil
}

func (f *ToolFactory) loadToolMetadata(ctx context.Context, definitions []runtimetooling.MCPToolDefinition) (map[string]*mcps.ToolInfo, error) {
	toolsByCode := make(map[string]*mcps.ToolInfo, len(definitions))
	serverCodes := make(map[string]struct{})
	for _, item := range definitions {
		serverCodes[item.ServerCode] = struct{}{}
	}
	for serverCode := range serverCodes {
		toolInfos, err := mcps.Runtime.ListTools(ctx, serverCode)
		if err != nil {
			return nil, err
		}
		for i := range toolInfos {
			toolInfo := toolInfos[i]
			toolCode := toolx.BuildMCPToolCode(serverCode, toolInfo.Name)
			toolInfoCopy := toolInfo
			toolsByCode[toolCode] = &toolInfoCopy
		}
	}
	return toolsByCode, nil
}

func cloneStringMap(input map[string]string) map[string]string {
	if len(input) == 0 {
		return nil
	}
	ret := make(map[string]string, len(input))
	for key, value := range input {
		ret[key] = value
	}
	return ret
}
