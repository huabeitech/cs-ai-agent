package request

import "cs-agent/internal/pkg/enums"

type CreateAIConfigRequest struct {
	Name             string            `json:"name"`
	Provider         enums.AIProvider  `json:"provider"`
	BaseURL          string            `json:"baseUrl"`
	APIKey           string            `json:"apiKey"`
	ModelType        enums.AIModelType `json:"modelType"`
	ModelName        string            `json:"modelName"`
	Dimension        int               `json:"dimension"`
	MaxContextTokens int               `json:"maxContextTokens"`
	MaxOutputTokens  int               `json:"maxOutputTokens"`
	TimeoutMS        int               `json:"timeoutMs"`
	MaxRetryCount    int               `json:"maxRetryCount"`
	RPMLimit         int               `json:"rpmLimit"`
	TPMLimit         int               `json:"tpmLimit"`
	Remark           string            `json:"remark"`
}

type UpdateAIConfigRequest struct {
	ID int64 `json:"id"`
	CreateAIConfigRequest
}

type DeleteAIConfigRequest struct {
	ID int64 `json:"id"`
}

type UpdateAIConfigStatusRequest struct {
	ID     int64        `json:"id"`
	Status enums.Status `json:"status"`
}

type CreateAIAgentRequest struct {
	Name             string                          `json:"name"`
	Description      string                          `json:"description"`
	AIConfigID       int64                           `json:"aiConfigId"`
	ServiceMode      enums.IMConversationServiceMode `json:"serviceMode"`
	SystemPrompt     string                          `json:"systemPrompt"`
	WelcomeMessage   string                          `json:"welcomeMessage"`
	TeamIDs          []int64                         `json:"teamIds"`
	HandoffMode      enums.AIAgentHandoffMode        `json:"handoffMode"`
	MaxAIReplyRounds int                             `json:"maxAiReplyRounds"`
	FallbackMode     enums.AIAgentFallbackMode       `json:"fallbackMode"`
	KnowledgeIDs     []int64                         `json:"knowledgeIds"`
	Remark           string                          `json:"remark"`
}

type UpdateAIAgentRequest struct {
	ID int64 `json:"id"`
	CreateAIAgentRequest
}

type DeleteAIAgentRequest struct {
	ID int64 `json:"id"`
}

type UpdateAIAgentStatusRequest struct {
	ID     int64 `json:"id"`
	Status int   `json:"status"`
}
