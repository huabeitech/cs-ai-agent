package rag

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	ragchunk "cs-agent/internal/ai/rag/chunk"
	"cs-agent/internal/ai/rag/vectordb"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/enums"

	"cs-agent/internal/ai"
	"github.com/mlogclub/simple/common/strs"
)

func (s *index) buildDocumentChunkRequest(document *models.KnowledgeDocument, knowledgeBase *models.KnowledgeBase) *ragchunk.ChunkRequest {
	return &ragchunk.ChunkRequest{
		KnowledgeBaseID: document.KnowledgeBaseID,
		DocumentID:      document.ID,
		DocumentTitle:   document.Title,
		ContentType:     document.ContentType,
		Content:         document.Content,
		PlainText:       ExtractPlainText(document.Content, document.ContentType),
		Options: ragchunk.ChunkOptions{
			Provider:       firstNonEmptyString(knowledgeBase.ChunkProvider, s.chunkConfig.Provider),
			TargetTokens:   firstPositiveInt(knowledgeBase.ChunkTargetTokens, s.chunkConfig.TargetTokens),
			MaxTokens:      firstPositiveInt(knowledgeBase.ChunkMaxTokens, s.chunkConfig.MaxTokens),
			OverlapTokens:  firstPositiveInt(knowledgeBase.ChunkOverlapTokens, s.chunkConfig.OverlapTokens),
			EnableFallback: s.chunkConfig.EnableFallback,
		},
	}
}

func (s *index) buildDocumentChunks(ctx context.Context, document *models.KnowledgeDocument, knowledgeBase *models.KnowledgeBase) ([]ragchunk.ChunkResult, error) {
	chunks, err := s.registry.Chunk(ctx, s.buildDocumentChunkRequest(document, knowledgeBase))
	if err != nil {
		return nil, fmt.Errorf("failed to chunk document: %w", err)
	}
	if len(chunks) == 0 {
		return nil, fmt.Errorf("no chunks generated from document")
	}
	return chunks, nil
}

func collectExistingVectorIDs(chunks []models.KnowledgeChunk) []string {
	ret := make([]string, 0, len(chunks))
	for _, chunk := range chunks {
		if strs.IsNotBlank(chunk.VectorID) {
			ret = append(ret, chunk.VectorID)
		}
	}
	return ret
}

func (s *index) prepareDocumentVectors(ctx context.Context, knowledgeBase *models.KnowledgeBase, document *models.KnowledgeDocument, chunks []ragchunk.ChunkResult) ([]vectordb.Vector, []models.KnowledgeChunk, int, error) {
	vectors := make([]vectordb.Vector, 0, len(chunks))
	chunkModels := make([]models.KnowledgeChunk, 0, len(chunks))
	dimension := 0

	for i, chunk := range chunks {
		embeddingResult, err := ai.Embedding.GenerateEmbedding(ctx, chunk.Content)
		if err != nil {
			slog.Error("Failed to generate embedding for chunk", "document_id", document.ID, "chunk_index", i, "error", err)
			return nil, nil, 0, fmt.Errorf("failed to generate embedding for chunk %d: %w", i, err)
		}
		if dimension == 0 {
			dimension = embeddingResult.Dimension
		}

		chunkID := buildKnowledgeChunkVectorID(knowledgeBase.ID, document.ID, chunk.ChunkNo)
		providerName := ""
		if chunk.Metadata != nil {
			if value, ok := chunk.Metadata["provider"].(string); ok {
				providerName = value
			}
		}
		now := time.Now()
		chunkModels = append(chunkModels, models.KnowledgeChunk{
			KnowledgeBaseID: knowledgeBase.ID,
			DocumentID:      document.ID,
			ChunkNo:         chunk.ChunkNo,
			Title:           chunk.Title,
			Content:         chunk.Content,
			ContentHash:     buildChunkContentHash(chunk.Content),
			CharCount:       chunk.CharCount,
			TokenCount:      chunk.TokenCount,
			ChunkType:       string(chunk.ChunkType),
			SectionPath:     chunk.SectionPath,
			Provider:        providerName,
			VectorID:        chunkID,
			Status:          enums.StatusOk,
			CreatedAt:       now,
			UpdatedAt:       now,
		})

		vectors = append(vectors, vectordb.Vector{
			ID:     chunkID,
			Vector: embeddingResult.Vector,
			Payload: vectordb.ChunkPayload{
				KnowledgeBaseID: knowledgeBase.ID,
				DocumentID:      document.ID,
				DocumentTitle:   document.Title,
				ChunkNo:         chunk.ChunkNo,
				ChunkType:       string(chunk.ChunkType),
				SectionPath:     chunk.SectionPath,
				Content:         chunk.Content,
				Title:           chunk.Title,
				Provider:        providerName,
			},
		})
	}

	if len(vectors) == 0 {
		return nil, nil, 0, fmt.Errorf("no vectors generated")
	}
	return vectors, chunkModels, dimension, nil
}
