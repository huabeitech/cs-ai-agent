package skills

import (
	"encoding/json"
	"time"

	"cs-agent/internal/models"
)

// BuildRunLog 根据执行计划与运行结果构建 Skill 运行日志。
func BuildRunLog(ctx RuntimeContext, plan *ExecutionPlan, trace *ExecutionTrace, err error) *models.SkillRunLog {
	log := &models.SkillRunLog{
		ConversationID:  ctx.ConversationID,
		AIAgentID:       ctx.AIAgentID,
		ManualSkillCode: ctx.ManualSkillCode,
		IntentCode:      ctx.IntentCode,
		UserMessage:     ctx.UserMessage,
		TraceData:       buildTraceData(trace),
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
			log.MatchReason = plan.MatchReason
		}
	}
	if err != nil {
		log.ErrorMessage = err.Error()
	} else if !log.Matched {
		if plan != nil && plan.MatchReason != "" {
			log.MatchReason = plan.MatchReason
		} else {
			log.MatchReason = "not_matched"
		}
	}
	return log
}

func buildTraceData(trace *ExecutionTrace) string {
	if trace == nil {
		return ""
	}
	data, err := json.Marshal(trace)
	if err != nil {
		return ""
	}
	return string(data)
}
