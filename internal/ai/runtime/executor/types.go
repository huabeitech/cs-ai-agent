package executor

import (
	"cs-agent/internal/ai/runtime/registry"
	"cs-agent/internal/models"
)

type RunInput struct {
	Conversation     models.Conversation
	UserMessage      models.Message
	AIAgent          models.AIAgent
	AIConfig         models.AIConfig
	SelectedSkill    *models.SkillDefinition
	SkillRouteReason string
	SkillRouteTrace  string
	CheckPointID     string
	ToolSet          *registry.ToolSet
}

type ResumeInput struct {
	Conversation models.Conversation
	AIAgent      models.AIAgent
	AIConfig     models.AIConfig
	CheckPointID string
	ResumeData   map[string]string
	ToolSet      *registry.ToolSet
}

type InterruptContextSummary struct {
	Type        string `json:"type,omitempty"`
	ID          string `json:"id"`
	InfoPreview string `json:"infoPreview,omitempty"`
}

type RunResult struct {
	RunID                 string
	Status                string
	ReplyText             string
	SelectedSkillCode     string
	SelectedSkillName     string
	SkillRouteReason      string
	SkillRouteTrace       string
	SkillAllowedToolCodes []string
	ModelName             string
	PromptTokens          int
	CompletionTokens      int
	HistoryMessageCount   int
	RetrieverCount        int
	ToolCallCount         int
	ToolCodes             []string
	InvokedToolCodes      []string
	CheckPointID          string
	Interrupted           bool
	Interrupts            []InterruptContextSummary
	TraceData             string
	ErrorMessage          string
}
