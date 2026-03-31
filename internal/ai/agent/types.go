package agent

import (
	"cs-agent/internal/ai/rag"
	"cs-agent/internal/models"
)

type Action string

const (
	ActionNoop     Action = "noop"
	ActionReply    Action = "reply"
	ActionFallback Action = "fallback"
)

// TurnContext 表示客服 Agent 处理一轮消息所需的最小上下文。
type TurnContext struct {
	Message      *models.Message
	Conversation *models.Conversation
	AIAgent      *models.AIAgent
	AIConfig     *models.AIConfig
}

// TurnResult 表示 Agent Runtime 对当前轮次给出的执行结果。
type TurnResult struct {
	Action        Action
	Question      string
	ReplyText     string
	Reason        string
	KnowledgeBase *models.KnowledgeBase
	RetrieveHits  []rag.RetrieveResult
}
