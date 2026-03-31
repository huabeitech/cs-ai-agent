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
func BuildExecutionPlan(ctx RuntimeContext) (*ExecutionPlan, error) {
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

	skill, err := MatchSkill(ctx)
	if err != nil {
		return nil, err
	}

	return &ExecutionPlan{
		AIAgent:  aiAgent,
		AIConfig: aiConfig,
		Skill:    skill,
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
	plan, err := BuildExecutionPlan(runtimeCtx)
	if err != nil {
		log := BuildRunLog(runtimeCtx, nil, err)
		_ = WriteRunLog(log)
		return nil, err
	}
	if plan == nil || plan.Skill == nil {
		log := BuildRunLog(runtimeCtx, plan, nil)
		_ = WriteRunLog(log)
		return nil, nil
	}

	replyText, err := executeByPlan(ctx, plan, runtimeCtx)
	log := BuildRunLog(runtimeCtx, plan, err)
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
	}, nil
}
