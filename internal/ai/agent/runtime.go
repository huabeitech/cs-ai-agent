package agent

import (
	"context"
	"encoding/json"
	"slices"
	"strings"

	"cs-agent/internal/ai/skills"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/errorsx"
	"cs-agent/internal/pkg/utils"
	"cs-agent/internal/repositories"

	"github.com/mlogclub/simple/sqls"
)

type Runtime struct {
	ragExecutor  *ragExecutor
	toolExecutor *toolExecutor
	planner      *planner
}

func NewRuntime() *Runtime {
	return &Runtime{
		ragExecutor:  newRAGExecutor(),
		toolExecutor: newToolExecutor(),
		planner:      newPlanner(),
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
		return r.runPlannedSkillOrRAG(ctx, turnCtx, question, plan)
	case ActionTool:
		return r.runPlannedToolOrRAG(ctx, turnCtx, question, plan)
	case ActionHandoff:
		return &TurnResult{
			Action:           ActionHandoff,
			Question:         question,
			Reason:           plan.Reason,
			PlannedAction:    plan.Action,
			PlannedSkillCode: plan.SkillCode,
			PlannedToolCode:  plannedToolCode(plan),
			PlanReason:       plan.Reason,
		}, nil
	}
	return r.ragExecutor.Execute(ctx, turnCtx, question)
}

func (r *Runtime) runPlannedSkillOrRAG(ctx context.Context, turnCtx TurnContext, question string, plan *Plan) (*TurnResult, error) {
	if plan == nil {
		return r.ragExecutor.Execute(ctx, turnCtx, question)
	}
	turnCtx.ManualSkillCode = strings.TrimSpace(plan.SkillCode)
	result, err := r.runManualSkill(ctx, turnCtx, question)
	if err == nil && result != nil && result.Action == ActionReply {
		result.PlanReason = normalizeReason(plan.Reason, "planner_skill")
		return result, nil
	}
	ragResult, ragErr := r.ragExecutor.Execute(ctx, turnCtx, question)
	if ragResult != nil {
		ragResult.PlanReason = normalizeReason(plan.Reason, "planner_skill")
	}
	if ragErr != nil {
		return ragResult, ragErr
	}
	return ragResult, nil
}

func (r *Runtime) runPlannedToolOrRAG(ctx context.Context, turnCtx TurnContext, question string, plan *Plan) (*TurnResult, error) {
	if plan == nil || plan.Tool == nil {
		return r.ragExecutor.Execute(ctx, turnCtx, question)
	}
	result, err := r.runDirectTool(ctx, turnCtx, question, plan.Tool, normalizeReason(plan.Reason, "planner_tool"))
	if err == nil && result != nil && result.Action == ActionReply {
		return result, nil
	}
	ragResult, ragErr := r.ragExecutor.Execute(ctx, turnCtx, question)
	if ragResult != nil {
		ragResult.PlannedAction = ActionTool
		ragResult.PlannedToolCode = plan.Tool.Code()
		ragResult.PlanReason = normalizeReason(plan.Reason, "planner_tool")
	}
	if ragErr != nil {
		return ragResult, ragErr
	}
	return ragResult, nil
}

func (r *Runtime) runManualSkill(ctx context.Context, turnCtx TurnContext, question string) (*TurnResult, error) {
	skillCode := strings.TrimSpace(turnCtx.ManualSkillCode)
	if !isSkillAllowed(turnCtx.AIAgent, skillCode) {
		return &TurnResult{
			Action:           ActionFallback,
			Question:         question,
			Reason:           "Skill未绑定到当前Agent",
			PlannedAction:    ActionSkill,
			PlannedSkillCode: skillCode,
			PlanReason:       "manual_skill",
		}, errorsx.Forbidden("Skill 未绑定到当前 AI Agent")
	}
	result, err := skills.Execute(ctx, skills.RuntimeContext{
		AIAgentID:       turnCtx.AIAgent.ID,
		UserMessage:     question,
		ConversationID:  conversationID(turnCtx.Conversation),
		ManualSkillCode: skillCode,
		IntentCode:      strings.TrimSpace(turnCtx.IntentCode),
	})
	if err != nil {
		return &TurnResult{
			Action:           ActionFallback,
			Question:         question,
			Reason:           "Skill执行失败",
			PlannedAction:    ActionSkill,
			PlannedSkillCode: skillCode,
			PlanReason:       "manual_skill",
		}, err
	}
	if result == nil || strings.TrimSpace(result.ReplyText) == "" {
		return &TurnResult{
			Action:           ActionFallback,
			Question:         question,
			Reason:           "Skill未返回结果",
			PlannedAction:    ActionSkill,
			PlannedSkillCode: skillCode,
			PlanReason:       "manual_skill",
		}, nil
	}
	return &TurnResult{
		Action:           ActionReply,
		Question:         question,
		ReplyText:        result.ReplyText,
		Reason:           "manual_skill",
		PlannedAction:    ActionSkill,
		PlannedSkillCode: skillCode,
		PlanReason:       "manual_skill",
	}, nil
}

func (r *Runtime) runDirectTool(ctx context.Context, turnCtx TurnContext, question string, tool *MCPTool, planReason string) (*TurnResult, error) {
	if !isToolAllowed(turnCtx.AIAgent, tool) {
		return &TurnResult{
			Action:          ActionFallback,
			Question:        question,
			Reason:          "Tool未绑定到当前Agent",
			PlannedAction:   ActionTool,
			PlannedToolCode: tool.Code(),
			PlanReason:      planReason,
		}, errorsx.Forbidden("Tool 未绑定到当前 AI Agent")
	}
	return r.toolExecutor.Execute(ctx, turnCtx, question, tool, planReason)
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

func isSkillAllowed(agent *models.AIAgent, skillCode string) bool {
	skillCode = strings.TrimSpace(skillCode)
	if agent == nil || skillCode == "" {
		return false
	}
	skill := repositories.SkillDefinitionRepository.GetByCode(sqls.DB(), skillCode)
	if skill == nil {
		return false
	}
	return slices.Contains(utils.SplitInt64s(agent.SkillIDs), skill.ID)
}

func isToolAllowed(agent *models.AIAgent, tool *MCPTool) bool {
	if agent == nil || tool == nil || strings.TrimSpace(agent.AllowedMCPTools) == "" {
		return false
	}
	var allowedTools []MCPTool
	if err := json.Unmarshal([]byte(agent.AllowedMCPTools), &allowedTools); err != nil {
		return false
	}
	for _, item := range allowedTools {
		if item.Code() == tool.Code() {
			return true
		}
	}
	return false
}

func plannedToolCode(plan *Plan) string {
	if plan == nil || plan.Tool == nil {
		return ""
	}
	return plan.Tool.Code()
}
