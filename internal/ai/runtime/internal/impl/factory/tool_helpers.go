package factory

import (
	"strings"

	runtimeinstruction "cs-agent/internal/ai/runtime/instruction"
	einocallbacks "cs-agent/internal/ai/runtime/internal/impl/callbacks"
	"cs-agent/internal/ai/runtime/registry"
	runtimetooling "cs-agent/internal/ai/runtime/tooling"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/toolx"
)

func buildInstructionTraceSummary(summary runtimeinstruction.AssemblySummary) einocallbacks.InstructionTraceSummary {
	return einocallbacks.InstructionTraceSummary{
		SectionTitles:     append([]string(nil), summary.SectionTitles...),
		HasGovernanceRule: summary.HasGovernanceRule,
		HasAgentRule:      summary.HasAgentRule,
		HasSkillRule:      summary.HasSkillRule,
		HasToolRule:       summary.HasToolRule,
	}
}

func buildRuntimeTraceToolMetadata(
	dynamicToolDefinitions []runtimetooling.MCPToolDefinition,
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
			SourceType: enums.ToolSourceTypeMCP,
		}
	}
	for modelName, metadata := range staticToolMetadata {
		modelName = strings.TrimSpace(modelName)
		metadata.ToolCode = strings.TrimSpace(metadata.ToolCode)
		metadata.ServerCode = strings.TrimSpace(metadata.ServerCode)
		metadata.ToolName = strings.TrimSpace(metadata.ToolName)
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
