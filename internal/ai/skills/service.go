package skills

import (
	"context"
	"strings"

	"cs-agent/internal/models"
)

var RuntimeService = newService()

func newService() *Service {
	return &Service{
		plan:   newPlanService(),
		runlog: newRunLogService(),
	}
}

type Service struct {
	plan   *planService
	runlog *RunLogService
}

// BuildExecutionPlan 构建当前请求的 Skill 执行计划。
func (s *Service) BuildExecutionPlan(execCtx context.Context, ctx RuntimeContext) (*ExecutionPlan, error) {
	return s.plan.BuildExecutionPlan(execCtx, ctx)
}

// WriteRunLog 写入 Skill 路由日志。
func (s *Service) WriteRunLog(log *models.SkillRunLog) error {
	return s.runlog.Write(log)
}

// Select 执行一次 Skill 路由并记录路由日志。
func (s *Service) Select(ctx context.Context, runtimeCtx RuntimeContext) (*ExecutionResult, error) {
	plan, err := s.BuildExecutionPlan(ctx, runtimeCtx)
	if err != nil {
		trace := &ExecutionTrace{Status: "route_error"}
		log := s.runlog.Build(runtimeCtx, nil, trace, err)
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
		log := s.runlog.Build(runtimeCtx, plan, trace, nil)
		_ = s.WriteRunLog(log)
		return &ExecutionResult{
			Plan:   plan,
			RunLog: log,
			Trace:  trace,
		}, nil
	}
	trace.MatchReason = strings.TrimSpace(plan.MatchReason)
	trace.Route = plan.RouteTrace
	log := s.runlog.Build(runtimeCtx, plan, trace, err)
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
