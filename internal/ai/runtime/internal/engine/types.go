package engine

import (
	"cs-agent/internal/models"

	einotool "github.com/cloudwego/eino/components/tool"
)

type Request struct {
	Conversation     *models.Conversation
	UserMessage      *models.Message
	AIAgent          *models.AIAgent
	AIConfig         *models.AIConfig
	SelectedSkill    *models.SkillDefinition
	SkillRouteReason string
	SkillRouteTrace  string
	CheckPointID     string
	ExtraTools       []einotool.BaseTool
	ExtraToolCodes   map[string]string
}

type ResumeRequest struct {
	Conversation   *models.Conversation
	AIAgent        *models.AIAgent
	AIConfig       *models.AIConfig
	CheckPointID   string
	ResumeData     map[string]any
	ExtraTools     []einotool.BaseTool
	ExtraToolCodes map[string]string
}

type InterruptContextSummary struct {
	Type        string `json:"type,omitempty"`
	ID          string `json:"id"`
	InfoPreview string `json:"infoPreview,omitempty"`
}

type Summary struct {
	RunID               string
	Status              string
	ReplyText           string
	SelectedSkillCode   string
	SkillRouteReason    string
	SkillRouteTrace     string
	ModelName           string
	PromptTokens        int
	CompletionTokens    int
	HistoryMessageCount int
	RetrieverCount      int
	ToolCallCount       int
	ToolCodes           []string
	InvokedToolCodes    []string
	CheckPointID        string
	Interrupted         bool
	Interrupts          []InterruptContextSummary
	TraceData           string
	ErrorMessage        string
}
