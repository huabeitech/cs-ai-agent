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

// ExecutionPlan 表示 Skill Runtime 计算出的最终执行计划。
type ExecutionPlan struct {
	AIAgent  *models.AIAgent         // AIAgent 为本次请求所属的 AI Agent。
	AIConfig *models.AIConfig        // AIConfig 为本次请求实际使用的模型配置。
	Skill    *models.SkillDefinition // Skill 为最终命中的 Skill，未命中时为空。
}
