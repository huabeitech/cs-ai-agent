package skills

import (
	"context"
	"strings"

	"cs-agent/internal/models"
	"cs-agent/internal/pkg/errorsx"
	"cs-agent/internal/repositories"

	"github.com/mlogclub/simple/sqls"
)

// BuildExecutionPlan 构建当前请求的 Skill 执行计划。
func BuildExecutionPlan(execCtx context.Context, ctx RuntimeContext) (*ExecutionPlan, error) {
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

// WriteRunLog 写入 Skill 运行日志。
func WriteRunLog(log *models.SkillRunLog) error {
	if log == nil {
		return nil
	}
	return repositories.SkillRunLogRepository.Create(sqls.DB(), log)
}

// Execute 执行一次 Skill 运行，当前阶段仅支持 prompt_only 风格的手动 Skill。
func Execute(ctx context.Context, runtimeCtx RuntimeContext) (*ExecutionResult, error) {
	plan, err := BuildExecutionPlan(ctx, runtimeCtx)
	if err != nil {
		trace := &ExecutionTrace{Status: "plan_error"}
		log := BuildRunLog(runtimeCtx, nil, trace, err)
		_ = WriteRunLog(log)
		return nil, err
	}
	if plan == nil || plan.Skill == nil {
		trace := &ExecutionTrace{Status: "noop"}
		if plan != nil {
			trace.MatchReason = strings.TrimSpace(plan.MatchReason)
			trace.Route = plan.RouteTrace
		}
		log := BuildRunLog(runtimeCtx, plan, trace, nil)
		_ = WriteRunLog(log)
		return nil, nil
	}

	replyText, trace, err := executeByPlan(ctx, plan, runtimeCtx)
	if trace != nil {
		trace.MatchReason = strings.TrimSpace(plan.MatchReason)
		if trace.Route == nil {
			trace.Route = plan.RouteTrace
		}
	}
	log := BuildRunLog(runtimeCtx, plan, trace, err)
	if strings.TrimSpace(replyText) != "" && strings.TrimSpace(log.MatchReason) == "" {
		log.MatchReason = string(plan.Skill.ExecutionMode)
	}
	if writeErr := WriteRunLog(log); writeErr != nil && err == nil {
		err = writeErr
	}
	if err != nil {
		return nil, err
	}
	return &ExecutionResult{
		Plan:      plan,
		ReplyText: strings.TrimSpace(replyText),
		RunLog:    log,
		Trace:     trace,
	}, nil
}
