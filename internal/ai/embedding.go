package ai

import (
	"context"
	"fmt"

	openai "github.com/openai/openai-go/v3"

	"cs-agent/internal/models"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/errorsx"
)

type EmbeddingResult struct {
	Vector     []float32
	TokensUsed int
	ModelName  string
	Dimension  int
}

type embedding struct{}

var Embedding = &embedding{}

func (s *embedding) GetModel(ctx context.Context) (*models.AIConfig, error) {
	config, err := GetEnabledAIConfig(enums.AIModelTypeEmbedding)
	if err != nil {
		return nil, errorsx.BusinessError(2001, "未配置可用的 Embedding 模型")
	}
	return config, nil
}

func (s *embedding) GenerateEmbedding(ctx context.Context, text string) (*EmbeddingResult, error) {
	if text == "" {
		return nil, errorsx.InvalidParam("文本内容不能为空")
	}

	result, err := s.callEmbeddingAPI(ctx, text)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *embedding) GenerateBatchEmbeddings(ctx context.Context, texts []string) ([]EmbeddingResult, error) {
	if len(texts) == 0 {
		return nil, errorsx.InvalidParam("文本列表不能为空")
	}

	results := make([]EmbeddingResult, 0, len(texts))
	for _, text := range texts {
		result, err := s.callEmbeddingAPI(ctx, text)
		if err != nil {
			return nil, fmt.Errorf("failed to generate embedding for text: %w", err)
		}
		results = append(results, *result)
	}

	return results, nil
}

func (s *embedding) callEmbeddingAPI(ctx context.Context, text string) (*EmbeddingResult, error) {
	config, err := GetEnabledAIConfig(enums.AIModelTypeEmbedding)
	if err != nil {
		return nil, err
	}
	client := newOpenAIClient(*config)
	embeddingResp, err := client.Embeddings.New(ctx, openai.EmbeddingNewParams{
		Input: openai.EmbeddingNewParamsInputUnion{
			OfString: openai.String(text),
		},
		Model: openai.EmbeddingModel(config.ModelName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to call embedding api: %w", err)
	}

	if len(embeddingResp.Data) == 0 {
		return nil, fmt.Errorf("no embedding data in response")
	}
	vector := make([]float32, 0, len(embeddingResp.Data[0].Embedding))
	for _, item := range embeddingResp.Data[0].Embedding {
		vector = append(vector, float32(item))
	}

	return &EmbeddingResult{
		Vector:     vector,
		TokensUsed: int(embeddingResp.Usage.TotalTokens),
		ModelName:  embeddingResp.Model,
		Dimension:  len(vector),
	}, nil
}

func (s *embedding) GetDimension(ctx context.Context) (int, error) {
	model, err := s.GetModel(ctx)
	if err != nil {
		return 0, err
	}
	return model.Dimension, nil
}
