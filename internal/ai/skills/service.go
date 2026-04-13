package skills

import (
	"context"
	"strings"

	"cs-agent/internal/models"
	"cs-agent/internal/pkg/errorsx"
	"cs-agent/internal/repositories"

	"github.com/mlogclub/simple/sqls"
)

var RuntimeService = newService()

func newService() *Service {
	return &Service{}
}

type Service struct{}

// BuildExecutionPlan 构建当前请求的 Skill 执行计划。
func (s *Service) BuildExecutionPlan(execCtx context.Context, ctx RuntimeContext) (*ExecutionPlan, error) {
	if ctx.AIAgentID <= 0 {
		return nil, errorsx.InvalidParam("AIAgentID不能为空")
	}

	aiAgent := repositories.AIAgentRepository.Get(sqls.DB(), ctx.AIAgentID)
	if aiAgent == nil {
		return nil, errorsx.InvalidParam("AI Agent不存在")
	}
	aiConfig := repositories.AIConfigRepository.Get(sqls.DB(), aiAgent.AIConfigID)
	if aiConfig == nil {
		return nil, errorsx.InvalidParam("AI Agent关联的AI配置不存在")
	}

	skill, matchReason, routeTrace, err := MatchSkill(execCtx, ctx, aiAgent, aiConfig)
	if err != nil {
		return nil, err
	}

	return &ExecutionPlan{
		AIAgent:     aiAgent,
		AIConfig:    aiConfig,
		Skill:       skill,
		MatchReason: strings.TrimSpace(matchReason),
		RouteTrace:  routeTrace,
	}, nil
}

// WriteRunLog 写入 Skill 路由日志。
func (s *Service) WriteRunLog(log *models.SkillRunLog) error {
	if log == nil {
		return nil
	}
	return repositories.SkillRunLogRepository.Create(sqls.DB(), log)
}

// Select 执行一次 Skill 路由并记录路由日志。
func (s *Service) Select(ctx context.Context, runtimeCtx RuntimeContext) (*ExecutionResult, error) {
	plan, err := s.BuildExecutionPlan(ctx, runtimeCtx)
	if err != nil {
		trace := &ExecutionTrace{Status: "route_error"}
		log := BuildRunLog(runtimeCtx, nil, trace, err)
		_ = s.WriteRunLog(log)
		return nil, err
	}
	trace := &ExecutionTrace{Status: "ok"}
	if plan == nil || plan.Skill == nil {
		if plan != nil {
			trace.Status = "not_matched"
			trace.MatchReason = strings.TrimSpace(plan.MatchReason)
			trace.Route = plan.RouteTrace
		}
		log := BuildRunLog(runtimeCtx, plan, trace, nil)
		_ = s.WriteRunLog(log)
		return &ExecutionResult{
			Plan:   plan,
			RunLog: log,
			Trace:  trace,
		}, nil
	}
	trace.MatchReason = strings.TrimSpace(plan.MatchReason)
	trace.Route = plan.RouteTrace
	log := BuildRunLog(runtimeCtx, plan, trace, err)
	if writeErr := s.WriteRunLog(log); writeErr != nil && err == nil {
		err = writeErr
	}
	if err != nil {
		return nil, err
	}
	return &ExecutionResult{
		Plan:   plan,
		RunLog: log,
		Trace:  trace,
	}, nil
}
