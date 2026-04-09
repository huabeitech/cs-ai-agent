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

func (f *ChatModelFactory) Build(ctx context.Context, item *models.AIConfig) (model.ToolCallingChatModel, error) {
	if item == nil {
		return nil, nil
	}
	conf := &openai.ChatModelConfig{
		APIKey:  strings.TrimSpace(item.APIKey),
		BaseURL: strings.TrimSpace(item.BaseURL),
		Model:   strings.TrimSpace(item.ModelName),
	}
	if item.TimeoutMS > 0 {
		conf.Timeout = time.Duration(item.TimeoutMS) * time.Millisecond
	}
	if item.MaxOutputTokens > 0 {
		maxCompletionTokens := item.MaxOutputTokens
		conf.MaxCompletionTokens = &maxCompletionTokens
	}
	if item.Provider == enums.AIProviderOpenAI && isAzureOpenAIBaseURL(item.BaseURL) {
		conf.ByAzure = true
		conf.APIVersion = "2024-06-01"
	}
	if extraFields := providerExtraFields(item); len(extraFields) > 0 {
		conf.ExtraFields = extraFields
	}
	return openai.NewChatModel(ctx, conf)
}

func isAzureOpenAIBaseURL(baseURL string) bool {
	baseURL = strings.ToLower(strings.TrimSpace(baseURL))
	return strings.Contains(baseURL, ".openai.azure.com")
}

func providerExtraFields(item *models.AIConfig) map[string]any {
	if item == nil {
		return nil
	}
	baseURL := strings.ToLower(strings.TrimSpace(item.BaseURL))
	modelName := strings.ToLower(strings.TrimSpace(item.ModelName))
	if strings.Contains(baseURL, "dashscope.aliyuncs.com") && strings.HasPrefix(modelName, "qwen3") {
		return map[string]any{
			"enable_thinking": false,
		}
	}
	return nil
}
