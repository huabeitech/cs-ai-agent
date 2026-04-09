package adapter

import "cs-agent/internal/models"

type AIConfigSnapshot struct {
	ID              int64
	Provider        string
	ModelName       string
	BaseURL         string
	MaxOutputTokens int
	TimeoutMS       int
}

func BuildAIConfigSnapshot(item *models.AIConfig) *AIConfigSnapshot {
	if item == nil {
		return nil
	}
	return &AIConfigSnapshot{
		ID:              item.ID,
		Provider:        string(item.Provider),
		ModelName:       item.ModelName,
		BaseURL:         item.BaseURL,
		MaxOutputTokens: item.MaxOutputTokens,
		TimeoutMS:       item.TimeoutMS,
	}
}
