package skills

import (
	"encoding/json"
	"time"

	"cs-agent/internal/models"
	"cs-agent/internal/repositories"

	"github.com/mlogclub/simple/sqls"
)

func newRunLogService() *RunLogService {
	return &RunLogService{}
}

type RunLogService struct{}

// Build 根据执行计划与运行结果构建 Skill 运行日志。
func (s *RunLogService) Build(ctx RuntimeContext, plan *ExecutionPlan, trace *ExecutionTrace, err error) *models.SkillRunLog {
	log := &models.SkillRunLog{
		ConversationID:  ctx.ConversationID,
		AIAgentID:       ctx.AIAgentID,
		ManualSkillCode: ctx.ManualSkillCode,
		IntentCode:      ctx.IntentCode,
		UserMessage:     ctx.UserMessage,
		TraceData:       s.buildTraceData(trace),
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

// Write 写入 Skill 路由日志。
func (s *RunLogService) Write(log *models.SkillRunLog) error {
	if log == nil {
		return nil
	}
	return repositories.SkillRunLogRepository.Create(sqls.DB(), log)
}

func (s *RunLogService) buildTraceData(trace *ExecutionTrace) string {
	if trace == nil {
		return ""
	}
	data, err := json.Marshal(trace)
	if err != nil {
		return ""
	}
	return string(data)
}
