package skills

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/errorsx"
	"cs-agent/internal/repositories"
	"cs-agent/internal/services"

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
	return services.SkillRunLogService.Create(log)
}
