package agent

import (
	"context"
	"strings"

	"cs-agent/internal/models"
	"cs-agent/internal/pkg/errorsx"
	"cs-agent/internal/repositories"

	"github.com/mlogclub/simple/sqls"
)

type Runtime struct {
	ragExecutor *ragExecutor
}

func NewRuntime() *Runtime {
	return &Runtime{
		ragExecutor: newRAGExecutor(),
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
	return r.ragExecutor.Execute(ctx, turnCtx, question)
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
