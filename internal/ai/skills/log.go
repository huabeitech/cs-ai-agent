package skills

import (
	"time"

	"cs-agent/internal/models"
)

// BuildRunLog 根据执行计划与运行结果构建 Skill 运行日志。
func BuildRunLog(ctx RuntimeContext, plan *ExecutionPlan, err error) *models.SkillRunLog {
	log := &models.SkillRunLog{
		ConversationID:  ctx.ConversationID,
		AIAgentID:       ctx.AIAgentID,
		ManualSkillCode: ctx.ManualSkillCode,
		IntentCode:      ctx.IntentCode,
		UserMessage:     ctx.UserMessage,
		CreatedAt:       time.Now(),
	}
	if plan != nil {
		if plan.AIConfig != nil {
			log.AIConfigID = plan.AIConfig.ID
			log.UsedModel = plan.AIConfig.ModelName
			log.UsedProvider = plan.AIConfig.Provider
		}
		if plan.Skill != nil {
			log.SkillDefinitionID = plan.Skill.ID
			log.SkillCode = plan.Skill.Code
			log.Matched = true
			log.FinalSelected = true
			log.MatchReason = "manual_skill_code"
		}
	}
	if err != nil {
		log.ErrorMessage = err.Error()
	} else if !log.Matched {
		log.MatchReason = "not_matched"
	}
	return log
}
