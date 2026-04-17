package skills

import (
	"context"
	"strings"
)

func newPlanService() *planService {
	return &planService{}
}

type planService struct{}

// BuildExecutionPlan 构建当前请求的 Skill 执行计划。
func (s *planService) BuildExecutionPlan(execCtx context.Context, ctx RuntimeContext) (*ExecutionPlan, error) {
	skill, matchReason, routeTrace, err := MatchSkill(execCtx, ctx)
	if err != nil {
		return nil, err
	}

	return &ExecutionPlan{
		AIAgent:     ctx.AIAgent,
		AIConfig:    ctx.AIConfig,
		Skill:       skill,
		MatchReason: strings.TrimSpace(matchReason),
		RouteTrace:  routeTrace,
	}, nil
}
