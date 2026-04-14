package factory

import (
	"strings"

	einoadapter "cs-agent/internal/ai/runtime/internal/impl/adapter"
	einocallbacks "cs-agent/internal/ai/runtime/internal/impl/callbacks"
	"cs-agent/internal/ai/runtime/registry"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/toolx"
)

func buildRuntimeTraceToolMetadata(
	dynamicToolDefinitions []einoadapter.MCPToolDefinition,
	staticToolMetadata map[string]registry.ToolMetadata,
	selectedSkill *models.SkillDefinition,
) map[string]einocallbacks.ToolMetadata {
	ret := make(map[string]einocallbacks.ToolMetadata, len(dynamicToolDefinitions)+len(staticToolMetadata)+1)
	for _, item := range dynamicToolDefinitions {
		modelName := strings.TrimSpace(item.ModelName)
		if modelName == "" {
			continue
		}
		ret[modelName] = einocallbacks.ToolMetadata{
			ToolCode:   strings.TrimSpace(item.ToolCode),
			ServerCode: strings.TrimSpace(item.ServerCode),
			ToolName:   strings.TrimSpace(item.ToolName),
			SourceType: "mcp",
		}
	}
	for modelName, metadata := range staticToolMetadata {
		modelName = strings.TrimSpace(modelName)
		metadata.ToolCode = strings.TrimSpace(metadata.ToolCode)
		metadata.ServerCode = strings.TrimSpace(metadata.ServerCode)
		metadata.ToolName = strings.TrimSpace(metadata.ToolName)
		metadata.SourceType = strings.TrimSpace(metadata.SourceType)
		if modelName == "" || metadata.ToolCode == "" {
			continue
		}
		ret[modelName] = einocallbacks.ToolMetadata{
			ToolCode:   metadata.ToolCode,
			ServerCode: metadata.ServerCode,
			ToolName:   metadata.ToolName,
			SourceType: metadata.SourceType,
		}
	}
	if selectedSkill != nil {
		resolved := toolx.ResolveToolMetadata(toolx.BuiltinSkill.Code, toolx.BuiltinSkill.Name)
		ret[toolx.BuiltinSkill.Name] = einocallbacks.ToolMetadata{
			ToolCode:   resolved.ToolCode,
			ServerCode: resolved.ServerCode,
			ToolName:   resolved.ToolName,
			SourceType: resolved.SourceType,
		}
	}
	return ret
}

func buildInstructionAppendices(selectedSkill *models.SkillDefinition, toolDefinitions []einoadapter.MCPToolDefinition, extraToolCodes map[string]string) []string {
	return NewToolAppendixProvider().Build(selectedSkill, toolDefinitions, extraToolCodes)
}
