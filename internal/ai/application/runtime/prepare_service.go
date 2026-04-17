package runtime

import (
	"context"
	"encoding/json"
	"strings"

	"cs-agent/internal/ai/skills"
	"cs-agent/internal/models"
)

func newPrepareService(catalog *toolCatalog) *prepareService {
	return &prepareService{catalog: catalog}
}

type prepareService struct {
	catalog *toolCatalog
}

func (s *prepareService) selectSkill(ctx context.Context, req Request) (*models.SkillDefinition, string, string, error) {
	if req.UserMessage == nil || req.Conversation == nil {
		return nil, "", "", nil
	}
	result, err := skills.Select(ctx, skills.RuntimeContext{
		AIAgent:         req.AIAgent,
		AIConfig:        req.AIConfig,
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
	if req == nil || req.ToolSet != nil || s.catalog == nil {
		return nil
	}
	toolSet, err := s.catalog.resolveForRun(req)
	if err != nil {
		return err
	}
	if toolSet != nil {
		req.ToolSet = toolSet
	}
	return nil
}

func (s *prepareService) prepareToolsForResume(req *ResumeRequest) error {
	if req == nil || req.ToolSet != nil || s.catalog == nil {
		return nil
	}
	toolSet, err := s.catalog.resolveForResume(req)
	if err != nil {
		return err
	}
	if toolSet != nil {
		req.ToolSet = toolSet
	}
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
