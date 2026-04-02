package ai

import (
	"time"

	"github.com/mlogclub/simple/sqls"
	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"

	"cs-agent/internal/models"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/errorsx"
	"cs-agent/internal/repositories"
)

func newOpenAIClient(config *models.AIConfig) openai.Client {
	opts := []option.RequestOption{
		option.WithAPIKey(config.APIKey),
		option.WithBaseURL(config.BaseURL),
	}
	if config.TimeoutMS > 0 {
		opts = append(opts, option.WithRequestTimeout(time.Duration(config.TimeoutMS)*time.Millisecond))
	}
	if config.MaxRetryCount >= 0 {
		opts = append(opts, option.WithMaxRetries(config.MaxRetryCount))
	}

	return openai.NewClient(opts...)
}

func GetEnabledAIConfig(modelType enums.AIModelType) (*models.AIConfig, error) {
	item := repositories.AIConfigRepository.GetEnabled(sqls.DB(), modelType)
	if item == nil {
		return nil, errorsx.BusinessError(2005, "未配置可用的 AI 配置")
	}
	return item, nil
}
