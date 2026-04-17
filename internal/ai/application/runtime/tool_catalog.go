package runtime

import (
	"encoding/json"
	"strings"

	"cs-agent/internal/ai/runtime/registry"
	"cs-agent/internal/ai/runtime/tools"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/toolx"
)

type toolCatalog struct {
	registry *registry.Registry
}

func newToolCatalog() *toolCatalog {
	return &toolCatalog{
		registry: registry.NewRegistry(buildRuntimeStaticTools()...),
	}
}

func buildRuntimeStaticTools() []registry.Tool {
	ret := make([]registry.Tool, 0, len(toolx.ListRuntimeStaticToolSpecs()))
	for _, spec := range toolx.ListRuntimeStaticToolSpecs() {
		tool := tools.NewRuntimeStaticTool(spec.Code)
		if tool == nil {
			continue
		}
		ret = append(ret, tool)
	}
	return ret
}

func (c *toolCatalog) resolveForRun(req *Request) (*registry.ToolSet, error) {
	if req == nil || req.ToolSet != nil || c == nil || c.registry == nil {
		return nil, nil
	}
	return c.registry.Resolve(registry.Context{
		Conversation:     req.Conversation,
		AIAgent:          req.AIAgent,
		AIConfig:         req.AIConfig,
		UserMessage:      req.UserMessage,
		AllowedToolCodes: c.resolveAllowedToolCodes(req.AIAgent, req.SelectedSkill),
	})
}

func (c *toolCatalog) resolveForResume(req *ResumeRequest) (*registry.ToolSet, error) {
	if req == nil || req.ToolSet != nil || c == nil || c.registry == nil {
		return nil, nil
	}
	return c.registry.Resolve(registry.Context{
		Conversation:     req.Conversation,
		AIAgent:          req.AIAgent,
		AIConfig:         req.AIConfig,
		AllowedToolCodes: c.parseAgentAllowedToolCodes(req.AIAgent),
	})
}

func (c *toolCatalog) parseSkillAllowedToolCodes(skill *models.SkillDefinition) []string {
	if skill == nil {
		return nil
	}
	raw := strings.TrimSpace(skill.ToolWhitelist)
	if raw == "" {
		return nil
	}
	var items []string
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil
	}
	return toolx.NormalizeToolCodes(items)
}

func (c *toolCatalog) parseAgentAllowedToolCodes(aiAgent models.AIAgent) []string {
	ret := make([]string, 0)
	if raw := strings.TrimSpace(aiAgent.AllowedMCPTools); raw != "" {
		items, err := toolx.ParseAgentMCPToolsJSON(raw)
		if err == nil {
			for _, item := range items {
				ret = append(ret, item.ToolCode)
			}
		}
	}
	if raw := strings.TrimSpace(aiAgent.AllowedGraphTools); raw != "" {
		var graphTools []string
		if err := json.Unmarshal([]byte(raw), &graphTools); err == nil {
			ret = append(ret, graphTools...)
		}
	}
	return toolx.NormalizeToolCodes(ret)
}

func (c *toolCatalog) resolveAllowedToolCodes(aiAgent models.AIAgent, skill *models.SkillDefinition) []string {
	return toolx.IntersectToolCodes(c.parseAgentAllowedToolCodes(aiAgent), c.parseSkillAllowedToolCodes(skill))
}
