package skills

import (
	"context"
	"strings"

	"cs-agent/internal/pkg/errorsx"
	"cs-agent/internal/repositories"

	"github.com/mlogclub/simple/sqls"
)

func newPlanService() *planService {
	return &planService{}
}

type planService struct{}

// BuildExecutionPlan 构建当前请求的 Skill 执行计划。
func (s *planService) BuildExecutionPlan(execCtx context.Context, ctx RuntimeContext) (*ExecutionPlan, error) {
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
