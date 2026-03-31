package agent

import (
	"context"
	"strings"

	"cs-agent/internal/ai/skills"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/errorsx"
	"cs-agent/internal/repositories"

	"github.com/mlogclub/simple/sqls"
)

type Runtime struct {
	ragExecutor *ragExecutor
	planner     *planner
}

func NewRuntime() *Runtime {
	return &Runtime{
		ragExecutor: newRAGExecutor(),
		planner:     newPlanner(),
	}
}

func (r *Runtime) RunConversationTurn(ctx context.Context, turnCtx TurnContext) (*TurnResult, error) {
	if turnCtx.Message == nil {
		return nil, errorsx.InvalidParam("消息不存在")
	}
	if turnCtx.Conversation == nil {
		return nil, errorsx.InvalidParam("会话不存在")
	}
	if turnCtx.AIAgent == nil {
		return &TurnResult{Action: ActionNoop}, nil
	}

	question := r.buildQuestion(turnCtx.Message, turnCtx.Conversation)
	if question == "" {
		return &TurnResult{
			Action: ActionNoop,
			Reason: "问题为空",
		}, nil
	}
	if turnCtx.AIConfig == nil {
		return nil, errorsx.InvalidParam("AI Agent 关联的 AI 配置不可用")
	}
	if strings.TrimSpace(turnCtx.ManualSkillCode) != "" {
		return r.runManualSkill(ctx, turnCtx, question)
	}
	plan, err := r.planner.Plan(ctx, turnCtx, question)
	if err != nil {
		return nil, err
	}
	switch plan.Action {
	case ActionSkill:
		turnCtx.ManualSkillCode = plan.SkillCode
		return r.runManualSkill(ctx, turnCtx, question)
	case ActionHandoff:
		return &TurnResult{
			Action:           ActionHandoff,
			Question:         question,
			Reason:           plan.Reason,
			PlannedAction:    plan.Action,
			PlannedSkillCode: plan.SkillCode,
			PlanReason:       plan.Reason,
		}, nil
	}
	return r.ragExecutor.Execute(ctx, turnCtx, question)
}

func (r *Runtime) runManualSkill(ctx context.Context, turnCtx TurnContext, question string) (*TurnResult, error) {
	result, err := skills.Execute(ctx, skills.RuntimeContext{
		AIAgentID:       turnCtx.AIAgent.ID,
		UserMessage:     question,
		ConversationID:  conversationID(turnCtx.Conversation),
		ManualSkillCode: strings.TrimSpace(turnCtx.ManualSkillCode),
		IntentCode:      strings.TrimSpace(turnCtx.IntentCode),
	})
	if err != nil {
		return &TurnResult{
			Action:           ActionFallback,
			Question:         question,
			Reason:           "Skill执行失败",
			PlannedAction:    ActionSkill,
			PlannedSkillCode: strings.TrimSpace(turnCtx.ManualSkillCode),
			PlanReason:       "manual_skill",
		}, err
	}
	if result == nil || strings.TrimSpace(result.ReplyText) == "" {
		return &TurnResult{
			Action:           ActionFallback,
			Question:         question,
			Reason:           "Skill未返回结果",
			PlannedAction:    ActionSkill,
			PlannedSkillCode: strings.TrimSpace(turnCtx.ManualSkillCode),
			PlanReason:       "manual_skill",
		}, nil
	}
	return &TurnResult{
		Action:           ActionReply,
		Question:         question,
		ReplyText:        result.ReplyText,
		Reason:           "manual_skill",
		PlannedAction:    ActionSkill,
		PlannedSkillCode: strings.TrimSpace(turnCtx.ManualSkillCode),
		PlanReason:       "manual_skill",
	}, nil
}

func (r *Runtime) buildQuestion(message *models.Message, conversation *models.Conversation) string {
	if message == nil {
		return ""
	}
	question := strings.TrimSpace(message.Content)
	if question != "" {
		return question
	}
	if conversation != nil {
		return strings.TrimSpace(conversation.LastMessageSummary)
	}
	return ""
}

func findKnowledgeBase(turnCtx TurnContext, knowledgeBaseID int64) *models.KnowledgeBase {
	if knowledgeBaseID <= 0 {
		return nil
	}
	return repositories.KnowledgeBaseRepository.Get(sqls.DB(), knowledgeBaseID)
}

func conversationID(conversation *models.Conversation) int64 {
	if conversation == nil {
		return 0
	}
	return conversation.ID
}
