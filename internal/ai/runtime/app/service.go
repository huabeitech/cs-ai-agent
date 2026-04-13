package app

import (
	"context"

	"cs-agent/internal/ai/runtime/internal/executor"
	"cs-agent/internal/ai/runtime/registry"
	"cs-agent/internal/ai/runtime/tools"
)

type Service struct {
	runtime  *executor.Service
	registry *registry.Registry
	prepare  *prepareService
}

func NewService() *Service {
	return &Service{
		runtime: executor.NewService(),
		registry: registry.NewRegistry(
			tools.NewTriageServiceRequestTool(),
			tools.NewAnalyzeConversationTool(),
			tools.NewPrepareTicketDraftTool(),
			tools.NewCreateTicketGraphTool(),
			tools.NewHandoffGraphTool(),
		),
	}
}

// TODO 这个方法真的要这样吗？ 不能直接在NewService中直接初始化吗？
func (s *Service) initPrepareService() {
	if s.prepare == nil {
		s.prepare = newPrepareService(s.registry)
	}
}

func (s *Service) Run(ctx context.Context, req Request) (*Summary, error) {
	s.initPrepareService()
	selectedSkill, skillReason, skillTrace, skillErr := s.prepare.selectSkill(ctx, req)
	req.SelectedSkill = selectedSkill
	req.SkillRouteReason = skillReason
	req.SkillRouteTrace = skillTrace
	if req.SelectedSkill != nil {
		req.SelectedSkill = cloneSkillDefinition(req.SelectedSkill)
	}
	if err := s.prepare.prepareToolsForRun(&req); err != nil {
		return nil, err
	}
	summary, err := s.runtime.ExecuteRun(ctx, executor.RunInput{
		Conversation:     req.Conversation,
		UserMessage:      req.UserMessage,
		AIAgent:          req.AIAgent,
		AIConfig:         req.AIConfig,
		SelectedSkill:    req.SelectedSkill,
		SkillRouteReason: req.SkillRouteReason,
		SkillRouteTrace:  req.SkillRouteTrace,
		CheckPointID:     req.CheckPointID,
		ToolSet:          req.ToolSet,
	})
	if err != nil {
		ret := toSummary(summary)
		if ret != nil && skillErr != nil && ret.PlanReason == "" {
			ret.PlanReason = "skill_failed_fallback_runtime"
		}
		return ret, err
	}
	ret := toSummary(summary)
	if ret != nil && skillErr != nil && ret.PlanReason == "" {
		ret.PlanReason = "skill_failed_fallback_runtime"
	}
	return ret, nil
}

func (s *Service) Resume(ctx context.Context, req ResumeRequest) (*Summary, error) {
	s.initPrepareService()
	if err := s.prepare.prepareToolsForResume(&req); err != nil {
		return nil, err
	}
	summary, err := s.runtime.ExecuteResume(ctx, executor.ResumeInput{
		Conversation: req.Conversation,
		AIAgent:      req.AIAgent,
		AIConfig:     req.AIConfig,
		CheckPointID: req.CheckPointID,
		ResumeData:   req.ResumeData,
		ToolSet:      req.ToolSet,
	})
	if err != nil {
		return toSummary(summary), err
	}
	return toSummary(summary), nil
}
