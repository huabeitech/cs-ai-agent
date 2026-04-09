package skills

import "cs-agent/internal/models"

// RuntimeContext 表示一次 Skill 运行的输入上下文。
type RuntimeContext struct {
	AIAgentID       int64  // AIAgentID 为当前请求所属的 AI Agent ID，必填。
	UserMessage     string // UserMessage 为当前用户输入。
	ConversationID  int64  // ConversationID 为当前会话 ID，无会话上下文时为 0。
	ManualSkillCode string // ManualSkillCode 为显式指定的 Skill 编码。
	IntentCode      string // IntentCode 为上游识别出的意图编码。
}

// ExecutionPlan 表示 Skill Runtime 计算出的最终路由结果。
type ExecutionPlan struct {
	AIAgent     *models.AIAgent         // AIAgent 为本次请求所属的 AI Agent。
	AIConfig    *models.AIConfig        // AIConfig 为本次请求实际使用的模型配置。
	Skill       *models.SkillDefinition // Skill 为最终命中的 Skill，未命中时为空。
	MatchReason string                  // MatchReason 为命中原因。
	RouteTrace  *RouteTrace             // RouteTrace 为匹配阶段的路由追踪。
}

// ExecutionResult 表示一次 Skill 路由的最终结果。
type ExecutionResult struct {
	Plan   *ExecutionPlan
	RunLog *models.SkillRunLog
	Trace  *ExecutionTrace
}

type ExecutionTrace struct {
	Status      string      `json:"status"`
	MatchReason string      `json:"matchReason,omitempty"`
	Route       *RouteTrace `json:"route,omitempty"`
}

type RouteTrace struct {
	Status              string   `json:"status"`
	CandidateSkillCodes []string `json:"candidateSkillCodes,omitempty"`
	SelectedSkillCode   string   `json:"selectedSkillCode,omitempty"`
	RawDecision         string   `json:"rawDecision,omitempty"`
	LatencyMs           int64    `json:"latencyMs,omitempty"`
	Error               string   `json:"error,omitempty"`
}

type PromptTrace struct {
	Status           string `json:"status"`
	LatencyMs        int64  `json:"latencyMs,omitempty"`
	ModelName        string `json:"modelName,omitempty"`
	PromptTokens     int    `json:"promptTokens,omitempty"`
	CompletionTokens int    `json:"completionTokens,omitempty"`
	Error            string `json:"error,omitempty"`
}
