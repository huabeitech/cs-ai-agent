package rag

import (
	"context"
	"fmt"
	"time"

	"cs-agent/internal/ai"
	"cs-agent/internal/ai/rag/vectordb"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/enums"
)

func buildFAQChunkModel(knowledgeBase *models.KnowledgeBase, faq *models.KnowledgeFAQ, content string) (models.KnowledgeChunk, string) {
	chunkID := buildKnowledgeFAQChunkVectorID(knowledgeBase.ID, faq.ID, 0)
	now := time.Now()
	return models.KnowledgeChunk{
		KnowledgeBaseID: knowledgeBase.ID,
		FaqID:           faq.ID,
		ChunkNo:         0,
		Title:           faq.Question,
		Content:         content,
		ContentHash:     buildChunkContentHash(content),
		CharCount:       len([]rune(content)),
		TokenCount:      len([]rune(content)) / 2,
		ChunkType:       string(enums.KnowledgeChunkTypeFAQ),
		Provider:        string(enums.KnowledgeChunkProviderFAQ),
		VectorID:        chunkID,
		Status:          enums.StatusOk,
		CreatedAt:       now,
		UpdatedAt:       now,
	}, chunkID
}

func (s *index) prepareFAQVector(ctx context.Context, knowledgeBase *models.KnowledgeBase, faq *models.KnowledgeFAQ, content string) (vectordb.Vector, models.KnowledgeChunk, int, error) {
	embeddingResult, err := ai.Embedding.GenerateEmbedding(ctx, content)
	if err != nil {
		return vectordb.Vector{}, models.KnowledgeChunk{}, 0, fmt.Errorf("failed to generate embedding for faq %d: %w", faq.ID, err)
	}

	chunkModel, chunkID := buildFAQChunkModel(knowledgeBase, faq, content)
	vector := vectordb.Vector{
		ID:     chunkID,
		Vector: embeddingResult.Vector,
		Payload: vectordb.ChunkPayload{
			KnowledgeBaseID: knowledgeBase.ID,
			FaqID:           faq.ID,
			FaqQuestion:     faq.Question,
			ChunkNo:         0,
			ChunkType:       string(enums.KnowledgeChunkTypeFAQ),
			Content:         content,
			Title:           faq.Question,
			Provider:        string(enums.KnowledgeChunkProviderFAQ),
		},
	}
	return vector, chunkModel, embeddingResult.Dimension, nil
}
