package ai

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/mlogclub/simple/common/strs"
	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/shared"

	"cs-agent/internal/models"
	"cs-agent/internal/pkg/enums"
)

type ChatCompletionResult struct {
	Content          string
	ModelName        string
	PromptTokens     int
	CompletionTokens int
}

type llm struct{}

var LLM = &llm{}

func (s *llm) Chat(ctx context.Context, systemPrompt string, userPrompt string) (*ChatCompletionResult, error) {
	config, err := GetEnabledAIConfig(enums.AIModelTypeLLM)
	if err != nil {
		return nil, err
	}
	return s.ChatWithConfig(ctx, config, systemPrompt, userPrompt)
}

func (s *llm) ChatWithConfig(ctx context.Context, config *models.AIConfig, systemPrompt string, userPrompt string) (*ChatCompletionResult, error) {
	if config == nil {
		return nil, fmt.Errorf("ai config is nil")
	}

	messages := make([]openai.ChatCompletionMessageParamUnion, 0, 2)
	if strs.IsNotBlank(systemPrompt) {
		messages = append(messages, openai.ChatCompletionMessageParamUnion{
			OfSystem: &openai.ChatCompletionSystemMessageParam{
				Content: openai.ChatCompletionSystemMessageParamContentUnion{
					OfString: openai.String(systemPrompt),
				},
			},
		})
	}
	messages = append(messages, openai.ChatCompletionMessageParamUnion{
		OfUser: &openai.ChatCompletionUserMessageParam{
			Content: openai.ChatCompletionUserMessageParamContentUnion{
				OfString: openai.String(userPrompt),
			},
		},
	})

	params := openai.ChatCompletionNewParams{
		Messages: messages,
		Model:    shared.ChatModel(config.ModelName),
	}
	if config.MaxOutputTokens > 0 {
		params.MaxCompletionTokens = openai.Int(int64(config.MaxOutputTokens))
	}
	applyProviderSpecificChatParams(&params, config)

	client := newOpenAIClient(config)
	chatResp, err := client.Chat.Completions.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to call llm api (model=%s provider=%s system_chars=%d user_chars=%d max_output_tokens=%d): %w",
			config.ModelName, config.Provider, utf8.RuneCountInString(systemPrompt), utf8.RuneCountInString(userPrompt), config.MaxOutputTokens, err)
	}
	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("no llm choices in response")
	}

	content := strings.TrimSpace(chatResp.Choices[0].Message.Content)
	return &ChatCompletionResult{
		Content:          content,
		ModelName:        config.ModelName,
		PromptTokens:     int(chatResp.Usage.PromptTokens),
		CompletionTokens: int(chatResp.Usage.CompletionTokens),
	}, nil
}

func applyProviderSpecificChatParams(params *openai.ChatCompletionNewParams, config *models.AIConfig) {
	if params == nil || config == nil {
		return
	}
	if isDashScopeQwenThinkingModel(config) {
		params.SetExtraFields(map[string]any{
			"enable_thinking": false,
		})
	}
}

func isDashScopeQwenThinkingModel(config *models.AIConfig) bool {
	if config == nil {
		return false
	}
	baseURL := strings.ToLower(strings.TrimSpace(config.BaseURL))
	modelName := strings.ToLower(strings.TrimSpace(config.ModelName))
	return strings.Contains(baseURL, "dashscope.aliyuncs.com") && strings.HasPrefix(modelName, "qwen3")
}
