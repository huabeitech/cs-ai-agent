package skills

import "cs-agent/internal/models"

// BuildRunLog 根据执行计划与运行结果构建 Skill 运行日志。
func BuildRunLog(ctx RuntimeContext, plan *ExecutionPlan, err error) *models.SkillRunLog {
	log := &models.SkillRunLog{
		ConversationID:  ctx.ConversationID,
		AIAgentID:       ctx.AIAgentID,
		ManualSkillCode: ctx.ManualSkillCode,
		IntentCode:      ctx.IntentCode,
		UserMessage:     ctx.UserMessage,
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
		}
	}
	if err != nil {
		log.ErrorMessage = err.Error()
	}
	return log
}
