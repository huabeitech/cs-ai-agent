package response

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/enums"
)

type AIAgentTeamResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type AIAgentSkillResponse struct {
	ID   int64  `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type AIConfigResponse struct {
	ID               int64             `json:"id"`
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
	Status           enums.Status      `json:"status"`
	SortNo           int               `json:"sortNo"`
	Remark           string            `json:"remark"`
}

func BuildAIConfigResponse(item *models.AIConfig) AIConfigResponse {
	return AIConfigResponse{
		ID:               item.ID,
		Name:             item.Name,
		Provider:         item.Provider,
		BaseURL:          item.BaseURL,
		APIKey:           item.APIKey,
		ModelType:        item.ModelType,
		ModelName:        item.ModelName,
		Dimension:        item.Dimension,
		MaxContextTokens: item.MaxContextTokens,
		MaxOutputTokens:  item.MaxOutputTokens,
		TimeoutMS:        item.TimeoutMS,
		MaxRetryCount:    item.MaxRetryCount,
		RPMLimit:         item.RPMLimit,
		TPMLimit:         item.TPMLimit,
		Status:           item.Status,
		SortNo:           item.SortNo,
		Remark:           item.Remark,
	}
}

type AIAgentResponse struct {
	ID                  int64                           `json:"id"`
	Name                string                          `json:"name"`
	Description         string                          `json:"description"`
	Status              enums.Status                    `json:"status"`
	StatusName          string                          `json:"statusName"`
	AIConfigID          int64                           `json:"aiConfigId"`
	AIConfigName        string                          `json:"aiConfigName"`
	ServiceMode         enums.IMConversationServiceMode `json:"serviceMode"`
	ServiceModeName     string                          `json:"serviceModeName"`
	SystemPrompt        string                          `json:"systemPrompt"`
	WelcomeMessage      string                          `json:"welcomeMessage"`
	ReplyTimeoutSeconds int                             `json:"replyTimeoutSeconds"`
	Teams               []AIAgentTeamResponse           `json:"teams"`
	HandoffMode         enums.AIAgentHandoffMode        `json:"handoffMode"`
	HandoffModeName     string                          `json:"handoffModeName"`
	MaxAIReplyRounds    int                             `json:"maxAiReplyRounds"`
	FallbackMode        enums.AIAgentFallbackMode       `json:"fallbackMode"`
	FallbackModeName    string                          `json:"fallbackModeName"`
	FallbackMessage     string                          `json:"fallbackMessage"`
	KnowledgeIDs        []int64                         `json:"knowledgeIds"`
	KnowledgeBaseNames  []string                        `json:"knowledgeBaseNames"`
	SkillIDs            []int64                         `json:"skillIds"`
	Skills              []AIAgentSkillResponse          `json:"skills"`
	SortNo              int                             `json:"sortNo"`
	Remark              string                          `json:"remark"`
	CreatedAt           string                          `json:"createdAt"`
	UpdatedAt           string                          `json:"updatedAt"`
	CreateUserName      string                          `json:"createUserName"`
	UpdateUserName      string                          `json:"updateUserName"`
}
