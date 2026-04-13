package app

import (
	"context"
	"encoding/json"
	"strings"

	"cs-agent/internal/ai/runtime/registry"
	"cs-agent/internal/ai/skills"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/toolx"
)

func newPrepareService(registry *registry.Registry) *prepareService {
	return &prepareService{registry: registry}
}

type prepareService struct {
	registry *registry.Registry
}

func (s *prepareService) selectSkill(ctx context.Context, req Request) (*models.SkillDefinition, string, string, error) {
	if req.AIAgent == nil || req.AIConfig == nil || req.UserMessage == nil || req.Conversation == nil {
		return nil, "", "", nil
	}
	result, err := skills.Select(ctx, skills.RuntimeContext{
		AIAgentID:       req.AIAgent.ID,
		UserMessage:     strings.TrimSpace(req.UserMessage.Content),
		ConversationID:  req.Conversation.ID,
		ManualSkillCode: strings.TrimSpace(req.ManualSkillCode),
	})
	if err != nil {
		return nil, "", "", err
	}
	if result == nil || result.Plan == nil || result.Plan.Skill == nil {
		traceData := marshalSkillRouteTrace(result)
		reason := ""
		if result != nil && result.Plan != nil {
			reason = strings.TrimSpace(result.Plan.MatchReason)
		}
		return nil, reason, traceData, nil
	}
	return result.Plan.Skill, strings.TrimSpace(result.Plan.MatchReason), marshalSkillRouteTrace(result), nil
}

func (s *prepareService) prepareToolsForRun(req *Request) error {
	if req == nil || req.ToolSet != nil || s.registry == nil {
		return nil
	}
	toolSet, err := s.registry.Resolve(registry.Context{
		Conversation:     req.Conversation,
		AIAgent:          req.AIAgent,
		AIConfig:         req.AIConfig,
		UserMessage:      req.UserMessage,
		AllowedToolCodes: resolveAllowedToolCodes(req.AIAgent, req.SelectedSkill),
	})
	if err != nil {
		return err
	}
	req.ToolSet = toolSet
	return nil
}

func (s *prepareService) prepareToolsForResume(req *ResumeRequest) error {
	if req == nil || req.ToolSet != nil || s.registry == nil {
		return nil
	}
	toolSet, err := s.registry.Resolve(registry.Context{
		Conversation:     req.Conversation,
		AIAgent:          req.AIAgent,
		AIConfig:         req.AIConfig,
		AllowedToolCodes: parseAgentAllowedToolCodes(req.AIAgent),
	})
	if err != nil {
		return err
	}
	req.ToolSet = toolSet
	return nil
}

func marshalSkillRouteTrace(result *skills.ExecutionResult) string {
	if result == nil || result.Plan == nil || result.Plan.RouteTrace == nil {
		return ""
	}
	buf, err := json.Marshal(result.Plan.RouteTrace)
	if err != nil {
		return ""
	}
	return string(buf)
}

func cloneSkillDefinition(item *models.SkillDefinition) *models.SkillDefinition {
	if item == nil {
		return nil
	}
	clone := *item
	return &clone
}

func parseSkillAllowedToolCodes(skill *models.SkillDefinition) []string {
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
	ret := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		item = toolx.NormalizeToolCodeAlias(item)
		if item == "" {
			continue
		}
		ret = append(ret, item)
	}
	return ret
}

func parseAgentAllowedToolCodes(aiAgent *models.AIAgent) []string {
	if aiAgent == nil || strings.TrimSpace(aiAgent.AllowedMCPTools) == "" {
		return nil
	}
	items, err := toolx.ParseAgentMCPToolsJSON(aiAgent.AllowedMCPTools)
	if err != nil {
		return nil
	}
	ret := make([]string, 0, len(items))
	for _, item := range items {
		toolCode := strings.TrimSpace(item.ToolCode)
		toolCode = toolx.NormalizeToolCodeAlias(toolCode)
		if toolCode == "" {
			continue
		}
		ret = append(ret, toolCode)
	}
	return ret
}

func resolveAllowedToolCodes(aiAgent *models.AIAgent, skill *models.SkillDefinition) []string {
	agentAllowed := parseAgentAllowedToolCodes(aiAgent)
	skillAllowed := parseSkillAllowedToolCodes(skill)
	switch {
	case len(agentAllowed) == 0:
		return skillAllowed
	case len(skillAllowed) == 0:
		return agentAllowed
	default:
		skillSet := make(map[string]struct{}, len(skillAllowed))
		for _, item := range skillAllowed {
			item = strings.TrimSpace(item)
			item = toolx.NormalizeToolCodeAlias(item)
			if item == "" {
				continue
			}
			skillSet[item] = struct{}{}
		}
		ret := make([]string, 0, len(agentAllowed))
		for _, item := range agentAllowed {
			item = strings.TrimSpace(item)
			item = toolx.NormalizeToolCodeAlias(item)
			if item == "" {
				continue
			}
			if _, ok := skillSet[item]; ok {
				ret = append(ret, item)
			}
		}
		return ret
	}
}
