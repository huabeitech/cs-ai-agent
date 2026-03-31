package agent

import (
	"cs-agent/internal/ai/rag"
	"cs-agent/internal/models"
	"strings"
)

type Action string

const (
	ActionNoop     Action = "noop"
	ActionSkill    Action = "skill"
	ActionTool     Action = "tool"
	ActionRAG      Action = "rag"
	ActionReply    Action = "reply"
	ActionFallback Action = "fallback"
	ActionHandoff  Action = "handoff"
)

type MCPTool struct {
	ServerCode  string            `json:"serverCode"`
	ToolName    string            `json:"toolName"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Arguments   map[string]string `json:"arguments"`
}

func (t MCPTool) Code() string {
	return strings.TrimSpace(t.ServerCode) + "/" + strings.TrimSpace(t.ToolName)
}

// TurnContext 表示客服 Agent 处理一轮消息所需的最小上下文。
type TurnContext struct {
	Message         *models.Message
	Conversation    *models.Conversation
	AIAgent         *models.AIAgent
	AIConfig        *models.AIConfig
	ManualSkillCode string
	IntentCode      string
}

// TurnResult 表示 Agent Runtime 对当前轮次给出的执行结果。
type TurnResult struct {
	Action           Action
	Question         string
	ReplyText        string
	Reason           string
	PlannedAction    Action
	PlannedSkillCode string
	PlannedToolCode  string
	PlanReason       string
	KnowledgeBase    *models.KnowledgeBase
	RetrieveHits     []rag.RetrieveResult
}

type Plan struct {
	Action    Action
	SkillCode string
	Tool      *MCPTool
	Reason    string
}
