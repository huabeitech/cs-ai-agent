package executor

import (
	"encoding/json"
	"strings"

	"cs-agent/internal/ai/runtime/internal/impl/adapter"
	"cs-agent/internal/ai/runtime/registry"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/toolx"

	einotool "github.com/cloudwego/eino/components/tool"
)

type preparedTooling struct {
	definitions         []adapter.MCPToolDefinition
	toolCodes           []string
	toolDefsByModelName map[string]string
	staticToolCodes     []string
	staticTools         []einotool.BaseTool
	staticToolCodeMap   map[string]string
	staticToolMetadata  map[string]registry.ToolMetadata
}

func prepareTooling(defs []adapter.MCPToolDefinition, selectedSkill *models.SkillDefinition, toolSet *registry.ToolSet, includeSkillTool bool) preparedTooling {
	filteredDefs := filterToolDefinitionsBySkill(defs, selectedSkill)
	ret := preparedTooling{
		definitions:         filteredDefs,
		toolCodes:           make([]string, 0, len(filteredDefs)+2),
		toolDefsByModelName: make(map[string]string, len(filteredDefs)),
		staticToolCodes:     staticToolCodeList(toolSet),
		staticTools:         toolSetStaticTools(toolSet),
		staticToolCodeMap:   toolSetStaticToolCodes(toolSet),
		staticToolMetadata:  toolSetStaticToolMetadata(toolSet),
	}
	for _, item := range filteredDefs {
		toolCode := strings.TrimSpace(item.ToolCode)
		modelName := strings.TrimSpace(item.ModelName)
		if toolCode == "" || modelName == "" {
			continue
		}
		ret.toolCodes = appendIfMissing(ret.toolCodes, toolCode)
		ret.toolDefsByModelName[modelName] = toolCode
	}
	if len(filteredDefs) > 0 {
		ret.toolCodes = appendIfMissing(ret.toolCodes, toolx.BuiltinToolSearch.Code)
		ret.toolDefsByModelName[toolx.BuiltinToolSearch.Name] = toolx.BuiltinToolSearch.Code
	}
	if includeSkillTool {
		ret.toolCodes = appendIfMissing(ret.toolCodes, toolx.BuiltinSkill.Code)
		ret.toolDefsByModelName[toolx.BuiltinSkill.Name] = toolx.BuiltinSkill.Code
	}
	for modelName, toolCode := range ret.staticToolCodeMap {
		modelName = strings.TrimSpace(modelName)
		toolCode = strings.TrimSpace(toolCode)
		if modelName == "" || toolCode == "" {
			continue
		}
		ret.toolCodes = appendIfMissing(ret.toolCodes, toolCode)
		ret.toolDefsByModelName[modelName] = toolCode
	}
	return ret
}

func toolSetStaticTools(toolSet *registry.ToolSet) []einotool.BaseTool {
	if toolSet == nil {
		return nil
	}
	return append([]einotool.BaseTool(nil), toolSet.StaticTools...)
}

func toolSetStaticToolCodes(toolSet *registry.ToolSet) map[string]string {
	if toolSet == nil || len(toolSet.StaticToolCodes) == 0 {
		return nil
	}
	ret := make(map[string]string, len(toolSet.StaticToolCodes))
	for name, code := range toolSet.StaticToolCodes {
		ret[strings.TrimSpace(name)] = strings.TrimSpace(code)
	}
	return ret
}

func toolSetStaticToolMetadata(toolSet *registry.ToolSet) map[string]registry.ToolMetadata {
	if toolSet == nil || len(toolSet.StaticToolMetadata) == 0 {
		return nil
	}
	ret := make(map[string]registry.ToolMetadata, len(toolSet.StaticToolMetadata))
	for name, item := range toolSet.StaticToolMetadata {
		trimmedName := strings.TrimSpace(name)
		if trimmedName == "" {
			continue
		}
		item.ToolCode = strings.TrimSpace(item.ToolCode)
		item.ServerCode = strings.TrimSpace(item.ServerCode)
		item.ToolName = strings.TrimSpace(item.ToolName)
		item.SourceType = strings.TrimSpace(item.SourceType)
		ret[trimmedName] = item
	}
	return ret
}

func staticToolCodeList(toolSet *registry.ToolSet) []string {
	if toolSet == nil {
		return nil
	}
	metadata := toolSetStaticToolMetadata(toolSet)
	ret := make([]string, 0, len(metadata))
	for _, item := range metadata {
		code := strings.TrimSpace(item.ToolCode)
		if code == "" {
			continue
		}
		ret = appendIfMissing(ret, code)
	}
	if len(ret) > 0 {
		return ret
	}
	for _, code := range toolSetStaticToolCodes(toolSet) {
		code = strings.TrimSpace(code)
		if code == "" {
			continue
		}
		ret = appendIfMissing(ret, code)
	}
	return ret
}

func definitionToolCodes(defs []adapter.MCPToolDefinition) []string {
	ret := make([]string, 0, len(defs))
	for _, item := range defs {
		code := strings.TrimSpace(item.ToolCode)
		if code == "" {
			continue
		}
		ret = append(ret, code)
	}
	return ret
}

func filterToolDefinitionsBySkill(defs []adapter.MCPToolDefinition, skill *models.SkillDefinition) []adapter.MCPToolDefinition {
	if skill == nil || strings.TrimSpace(skill.ToolWhitelist) == "" {
		return defs
	}
	var allowed []string
	if err := json.Unmarshal([]byte(skill.ToolWhitelist), &allowed); err != nil {
		return defs
	}
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, item := range allowed {
		item = toolx.NormalizeToolCodeAlias(item)
		if strings.TrimSpace(item) == "" {
			continue
		}
		allowedSet[strings.TrimSpace(item)] = struct{}{}
	}
	if len(allowedSet) == 0 {
		return defs
	}
	ret := make([]adapter.MCPToolDefinition, 0, len(defs))
	for _, item := range defs {
		if _, ok := allowedSet[strings.TrimSpace(item.ToolCode)]; ok {
			ret = append(ret, item)
		}
	}
	return ret
}

func parseJSONArrayList(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var items []string
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil
	}
	ret := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		ret = append(ret, item)
	}
	return ret
}
