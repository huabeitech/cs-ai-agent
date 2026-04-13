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
		registry: registry.NewRegistry(
			tools.NewTriageServiceRequestTool(),
			tools.NewAnalyzeConversationTool(),
			tools.NewPrepareTicketDraftTool(),
			tools.NewCreateTicketGraphTool(),
			tools.NewHandoffGraphTool(),
		),
	}
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
	return normalizeAllowedToolCodes(items)
}

func (c *toolCatalog) parseAgentAllowedToolCodes(aiAgent *models.AIAgent) []string {
	if aiAgent == nil || strings.TrimSpace(aiAgent.AllowedMCPTools) == "" {
		return nil
	}
	items, err := toolx.ParseAgentMCPToolsJSON(aiAgent.AllowedMCPTools)
	if err != nil {
		return nil
	}
	ret := make([]string, 0, len(items))
	for _, item := range items {
		ret = append(ret, item.ToolCode)
	}
	return normalizeAllowedToolCodes(ret)
}

func (c *toolCatalog) resolveAllowedToolCodes(aiAgent *models.AIAgent, skill *models.SkillDefinition) []string {
	agentAllowed := c.parseAgentAllowedToolCodes(aiAgent)
	skillAllowed := c.parseSkillAllowedToolCodes(skill)
	switch {
	case len(agentAllowed) == 0:
		return skillAllowed
	case len(skillAllowed) == 0:
		return agentAllowed
	default:
		skillSet := make(map[string]struct{}, len(skillAllowed))
		for _, item := range skillAllowed {
			skillSet[item] = struct{}{}
		}
		ret := make([]string, 0, len(agentAllowed))
		for _, item := range agentAllowed {
			if _, ok := skillSet[item]; ok {
				ret = append(ret, item)
			}
		}
		return ret
	}
}

func normalizeAllowedToolCodes(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	ret := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		item = toolx.NormalizeToolCodeAlias(strings.TrimSpace(item))
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		ret = append(ret, item)
	}
	return ret
}
