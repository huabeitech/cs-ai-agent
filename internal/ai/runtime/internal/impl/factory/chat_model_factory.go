package factory

import (
	"context"
	"strings"
	"time"

	"cs-agent/internal/models"
	"cs-agent/internal/pkg/enums"

	openai "github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
)

type ChatModelFactory struct{}

func NewChatModelFactory() *ChatModelFactory {
	return &ChatModelFactory{}
}

func (f *ChatModelFactory) Build(ctx context.Context, aiConfig models.AIConfig) (model.ToolCallingChatModel, error) {
	conf := &openai.ChatModelConfig{
		APIKey:  strings.TrimSpace(aiConfig.APIKey),
		BaseURL: strings.TrimSpace(aiConfig.BaseURL),
		Model:   strings.TrimSpace(aiConfig.ModelName),
	}
	if aiConfig.TimeoutMS > 0 {
		conf.Timeout = time.Duration(aiConfig.TimeoutMS) * time.Millisecond
	}
	if aiConfig.MaxOutputTokens > 0 {
		maxCompletionTokens := aiConfig.MaxOutputTokens
		conf.MaxCompletionTokens = &maxCompletionTokens
	}
	if aiConfig.Provider == enums.AIProviderOpenAI && isAzureOpenAIBaseURL(aiConfig.BaseURL) {
		conf.ByAzure = true
		conf.APIVersion = "2024-06-01"
	}
	if extraFields := providerExtraFields(aiConfig); len(extraFields) > 0 {
		conf.ExtraFields = extraFields
	}
	return openai.NewChatModel(ctx, conf)
}

func isAzureOpenAIBaseURL(baseURL string) bool {
	baseURL = strings.ToLower(strings.TrimSpace(baseURL))
	return strings.Contains(baseURL, ".openai.azure.com")
}

func providerExtraFields(aiConfig models.AIConfig) map[string]any {
	baseURL := strings.ToLower(strings.TrimSpace(aiConfig.BaseURL))
	modelName := strings.ToLower(strings.TrimSpace(aiConfig.ModelName))
	if strings.Contains(baseURL, "dashscope.aliyuncs.com") && strings.HasPrefix(modelName, "qwen3") {
		return map[string]any{
			"enable_thinking": false,
		}
	}
	return nil
}
