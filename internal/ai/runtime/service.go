package runtime

import (
	"context"
	"strings"

	"cs-agent/internal/ai/runtime/internal/engine"
	"cs-agent/internal/ai/runtime/registry"
	"cs-agent/internal/ai/runtime/tools"
	"cs-agent/internal/ai/skills"
)

var Service = newService()

func newService() *service {
	return &service{
		runtime: engine.NewService(),
		registry: registry.NewRegistry(
			tools.NewCreateTicketConfirmTool(),
		),
	}
}

type service struct {
	runtime  *engine.Service
	registry *registry.Registry
}

func (s *service) Run(ctx context.Context, req Request) (*Summary, error) {
	skillSummary, skillErr := s.tryRunSkill(ctx, req)
	if skillSummary != nil && strings.TrimSpace(skillSummary.ReplyText) != "" {
		return skillSummary, nil
	}
	if err := s.prepareToolsForRun(&req); err != nil {
		return nil, err
	}
	summary, err := s.runtime.Run(ctx, engine.Request{
		Conversation:   req.Conversation,
		UserMessage:    req.UserMessage,
		AIAgent:        req.AIAgent,
		AIConfig:       req.AIConfig,
		CheckPointID:   req.CheckPointID,
		ExtraTools:     req.ExtraTools,
		ExtraToolCodes: req.ExtraToolCodes,
	})
	if err != nil {
		ret := toSummary(summary)
		if ret != nil && skillErr != nil && strings.TrimSpace(ret.PlanReason) == "" {
			ret.PlanReason = "skill_failed_fallback_runtime"
		}
		return ret, err
	}
	ret := toSummary(summary)
	if ret != nil && skillErr != nil && strings.TrimSpace(ret.PlanReason) == "" {
		ret.PlanReason = "skill_failed_fallback_runtime"
	}
	return ret, nil
}

func (s *service) Resume(ctx context.Context, req ResumeRequest) (*Summary, error) {
	if err := s.prepareToolsForResume(&req); err != nil {
		return nil, err
	}
	summary, err := s.runtime.Resume(ctx, engine.ResumeRequest{
		Conversation:   req.Conversation,
		AIAgent:        req.AIAgent,
		AIConfig:       req.AIConfig,
		CheckPointID:   req.CheckPointID,
		ResumeData:     req.ResumeData,
		ExtraTools:     req.ExtraTools,
		ExtraToolCodes: req.ExtraToolCodes,
	})
	if err != nil {
		return toSummary(summary), err
	}
	return toSummary(summary), nil
}

func (s *service) prepareToolsForRun(req *Request) error {
	if req == nil || len(req.ExtraTools) > 0 || len(req.ExtraToolCodes) > 0 || s.registry == nil {
		return nil
	}
	toolSet, err := s.registry.Resolve(registry.Context{
		Conversation: req.Conversation,
		AIAgent:      req.AIAgent,
		AIConfig:     req.AIConfig,
		UserMessage:  req.UserMessage,
	})
	if err != nil {
		return err
	}
	req.ExtraTools = toolSet.Tools
	req.ExtraToolCodes = toolSet.ToolCodes
	return nil
}

func (s *service) prepareToolsForResume(req *ResumeRequest) error {
	if req == nil || len(req.ExtraTools) > 0 || len(req.ExtraToolCodes) > 0 || s.registry == nil {
		return nil
	}
	toolSet, err := s.registry.Resolve(registry.Context{
		Conversation: req.Conversation,
		AIAgent:      req.AIAgent,
		AIConfig:     req.AIConfig,
	})
	if err != nil {
		return err
	}
	req.ExtraTools = toolSet.Tools
	req.ExtraToolCodes = toolSet.ToolCodes
	return nil
}

func toSummary(summary *engine.Summary) *Summary {
	if summary == nil {
		return nil
	}
	ret := &Summary{
		RunID:               summary.RunID,
		Status:              summary.Status,
		ReplyText:           summary.ReplyText,
		PlannedSkillCode:    "",
		PlanReason:          "",
		ModelName:           summary.ModelName,
		PromptTokens:        summary.PromptTokens,
		CompletionTokens:    summary.CompletionTokens,
		HistoryMessageCount: summary.HistoryMessageCount,
		RetrieverCount:      summary.RetrieverCount,
		ToolCallCount:       summary.ToolCallCount,
		ToolCodes:           append([]string(nil), summary.ToolCodes...),
		InvokedToolCodes:    append([]string(nil), summary.InvokedToolCodes...),
		CheckPointID:        summary.CheckPointID,
		Interrupted:         summary.Interrupted,
		TraceData:           summary.TraceData,
		ErrorMessage:        summary.ErrorMessage,
	}
	if len(summary.Interrupts) > 0 {
		ret.Interrupts = make([]InterruptContextSummary, 0, len(summary.Interrupts))
		for _, item := range summary.Interrupts {
			ret.Interrupts = append(ret.Interrupts, InterruptContextSummary{
				Type:        item.Type,
				ID:          item.ID,
				InfoPreview: item.InfoPreview,
			})
		}
	}
	return ret
}

func (s *service) tryRunSkill(ctx context.Context, req Request) (*Summary, error) {
	if req.AIAgent == nil || req.AIConfig == nil || req.UserMessage == nil || req.Conversation == nil {
		return nil, nil
	}
	result, err := skills.Execute(ctx, skills.RuntimeContext{
		AIAgentID:      req.AIAgent.ID,
		UserMessage:    strings.TrimSpace(req.UserMessage.Content),
		ConversationID: req.Conversation.ID,
	})
	if err != nil {
		return nil, err
	}
	if result == nil || result.Plan == nil || result.Plan.Skill == nil {
		return nil, nil
	}
	traceData := ""
	if result.RunLog != nil {
		traceData = result.RunLog.TraceData
	}
	return &Summary{
		Status:           "completed",
		ReplyText:        strings.TrimSpace(result.ReplyText),
		PlannedSkillCode: strings.TrimSpace(result.Plan.Skill.Code),
		PlanReason:       strings.TrimSpace(result.Plan.MatchReason),
		ModelName:        req.AIConfig.ModelName,
		TraceData:        traceData,
	}, nil
}
