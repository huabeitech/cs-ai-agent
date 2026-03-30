package rag

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"time"

	"cs-agent/internal/ai"
	ragchunk "cs-agent/internal/ai/rag/chunk"
	"cs-agent/internal/ai/rag/vectordb"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/repositories"

	"github.com/google/uuid"
	"github.com/mlogclub/simple/common/strs"
	"github.com/mlogclub/simple/sqls"
)

type ChunkingConfig struct {
	Provider       string
	TargetTokens   int
	MaxTokens      int
	OverlapTokens  int
	EnableFallback bool
}

type index struct {
	chunkConfig ChunkingConfig
	registry    *ragchunk.Registry
}

const knowledgeCollectionName = "knowledge_chunks"

var Index = &index{
	chunkConfig: ChunkingConfig{
		Provider:       string(enums.KnowledgeChunkProviderStructured),
		TargetTokens:   300,
		MaxTokens:      400,
		OverlapTokens:  40,
		EnableFallback: true,
	},
	registry: ragchunk.NewDefaultRegistry(),
}

func (s *index) IndexDocumentByID(ctx context.Context, documentID int64) error {
	document := repositories.KnowledgeDocumentRepository.Get(sqls.DB(), documentID)
	if document == nil {
		return fmt.Errorf("document not found: %d", documentID)
	}
	return s.IndexDocument(ctx, document)
}

func (s *index) IndexDocument(ctx context.Context, document *models.KnowledgeDocument) error {
	start := time.Now()
	if err := s.markDocumentIndexPending(document.ID); err != nil {
		slog.Error("Failed to mark knowledge document index as pending", "document_id", document.ID, "error", err)
	}

	fail := func(err error) error {
		if updateErr := s.markDocumentIndexFailed(document.ID, err); updateErr != nil {
			slog.Error("Failed to mark knowledge document index as failed", "document_id", document.ID, "error", updateErr)
		}
		return err
	}

	// TODO 这里每次都查询下知识库不太友好
	knowledgeBase := repositories.KnowledgeBaseRepository.Get(sqls.DB(), document.KnowledgeBaseID)
	if knowledgeBase == nil {
		return fail(fmt.Errorf("knowledge base not found: %d", document.KnowledgeBaseID))
	}

	existingChunks := repositories.KnowledgeChunkRepository.FindByDocumentID(sqls.DB(), document.ID)

	chunks, err := s.registry.Chunk(ctx, &ragchunk.ChunkRequest{
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
	})
	if err != nil {
		return fail(fmt.Errorf("failed to chunk document: %w", err))
	}
	if len(chunks) == 0 {
		return fail(fmt.Errorf("no chunks generated from document"))
	}

	collectionName := s.getCollectionName()
	provider := vectordb.GetProvider()
	if provider == nil {
		return fail(fmt.Errorf("vectordb provider not initialized"))
	}

	if _, err := ai.Embedding.GetModel(ctx); err != nil {
		return fail(fmt.Errorf("failed to get embedding model: %w", err))
	}

	existingVectorIDs := make([]string, 0, len(existingChunks))
	for _, chunk := range existingChunks {
		if strs.IsNotBlank(chunk.VectorID) {
			existingVectorIDs = append(existingVectorIDs, chunk.VectorID)
		}
	}

	vectors := make([]vectordb.Vector, 0, len(chunks))
	chunkModels := make([]models.KnowledgeChunk, 0, len(chunks))
	dimension := 0

	for i, chunk := range chunks {
		embeddingResult, err := ai.Embedding.GenerateEmbedding(ctx, chunk.Content)
		if err != nil {
			slog.Error("Failed to generate embedding for chunk", "document_id", document.ID, "chunk_index", i, "error", err)
			return fail(fmt.Errorf("failed to generate embedding for chunk %d: %w", i, err))
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
		chunkModel := models.KnowledgeChunk{
			KnowledgeBaseID: knowledgeBase.ID,
			DocumentID:      document.ID,
			ChunkNo:         chunk.ChunkNo,
			Title:           chunk.Title,
			Content:         chunk.Content,
			ContentHash:     buildChunkContentHash(chunk.Content),
			CharCount:       chunk.CharCount,
			TokenCount:      chunk.TokenCount,
			VectorID:        chunkID,
			Status:          enums.StatusOk,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}
		chunkModels = append(chunkModels, chunkModel)

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
		return fail(fmt.Errorf("no vectors generated"))
	}

	collectionInfo, err := provider.GetCollection(ctx, collectionName)
	if err != nil || collectionInfo == nil {
		if dimension <= 0 {
			return fail(fmt.Errorf("invalid embedding dimension: %d", dimension))
		}
		if err := provider.CreateCollection(ctx, collectionName, dimension); err != nil {
			return fail(fmt.Errorf("failed to create collection: %w", err))
		}
		slog.Info("Created collection for knowledge base", "collection", collectionName, "dimension", dimension)
	}

	if len(existingVectorIDs) > 0 {
		if err := provider.DeleteVectors(ctx, collectionName, existingVectorIDs); err != nil {
			return fail(fmt.Errorf("failed to delete old vectors: %w", err))
		}
	}

	if err := provider.UpsertVectors(ctx, collectionName, vectors); err != nil {
		return fail(fmt.Errorf("failed to upsert vectors: %w", err))
	}

	if err := sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		if err := ctx.Tx.Where("document_id = ?", document.ID).Delete(&models.KnowledgeChunk{}).Error; err != nil {
			return err
		}
		for _, chunk := range chunkModels {
			if err := ctx.Tx.Create(&chunk).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return fail(fmt.Errorf("failed to save chunks: %w", err))
	}

	if err := s.markDocumentIndexIndexed(document.ID); err != nil {
		slog.Error("Failed to mark knowledge document index as indexed", "document_id", document.ID, "error", err)
	}

	slog.Info("Document indexed successfully",
		slog.Any("document_id", document.ID),
		slog.Any("chunks_count", len(chunks)),
		slog.Any("vectors_count", len(vectors)),
		slog.Any("time_taken", time.Since(start).String()),
	)

	return nil
}

func (s *index) RemoveDocumentIndex(ctx context.Context, documentID int64) error {
	document := repositories.KnowledgeDocumentRepository.Get(sqls.DB(), documentID)
	if document == nil {
		return nil
	}
	chunks := repositories.KnowledgeChunkRepository.Find(sqls.DB(), sqls.NewCnd().Eq("document_id", documentID))
	return s.removeDocumentIndexByChunks(ctx, document.KnowledgeBaseID, documentID, chunks)
}

func (s *index) RemoveDocumentIndexFromKnowledgeBase(ctx context.Context, knowledgeBaseID int64, documentID int64) error {
	chunks := repositories.KnowledgeChunkRepository.Find(sqls.DB(), sqls.NewCnd().Eq("document_id", documentID))
	return s.removeDocumentIndexByChunks(ctx, knowledgeBaseID, documentID, chunks)
}

func (s *index) RemoveDocumentIndexByChunkModels(ctx context.Context, knowledgeBaseID int64, documentID int64, chunks []models.KnowledgeChunk) error {
	return s.removeDocumentIndexByChunks(ctx, knowledgeBaseID, documentID, chunks)
}

func (s *index) removeDocumentIndexByChunks(ctx context.Context, knowledgeBaseID int64, documentID int64, chunks []models.KnowledgeChunk) error {
	if len(chunks) == 0 {
		return nil
	}

	collectionName := s.getCollectionName()
	provider := vectordb.GetProvider()
	if provider == nil {
		return fmt.Errorf("vectordb provider not initialized")
	}

	vectorIDs := make([]string, 0, len(chunks))
	for _, chunk := range chunks {
		if chunk.VectorID != "" {
			vectorIDs = append(vectorIDs, chunk.VectorID)
		}
	}

	if len(vectorIDs) > 0 {
		if err := provider.DeleteVectors(ctx, collectionName, vectorIDs); err != nil {
			slog.Error("Failed to delete vectors", "error", err)
		}
	}

	if err := sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		return ctx.Tx.Where("document_id = ?", documentID).Delete(&models.KnowledgeChunk{}).Error
	}); err != nil {
		return fmt.Errorf("failed to delete chunks: %w", err)
	}

	slog.Info("Document index removed", "document_id", documentID, "chunks_removed", len(chunks))
	return nil
}

func (s *index) getCollectionName() string {
	return knowledgeCollectionName
}

func buildKnowledgeChunkVectorID(knowledgeBaseID int64, documentID int64, chunkNo int) string {
	raw := fmt.Sprintf("kb:%d:doc:%d:chunk:%d", knowledgeBaseID, documentID, chunkNo)
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(raw)).String()
}

func buildChunkContentHash(content string) string {
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:])
}

func firstPositiveInt(values ...int) int {
	for _, value := range values {
		if value > 0 {
			return value
		}
	}
	return 0
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func (s *index) EnsureCollection(ctx context.Context) error {
	dimension, err := ai.Embedding.GetDimension(ctx)
	if err != nil {
		return fmt.Errorf("failed to get embedding dimension: %w", err)
	}

	collectionName := s.getCollectionName()
	provider := vectordb.GetProvider()
	if provider == nil {
		return fmt.Errorf("vectordb provider not initialized")
	}

	existing, err := provider.GetCollection(ctx, collectionName)
	if err == nil && existing != nil {
		return nil
	}

	return provider.CreateCollection(ctx, collectionName, dimension)
}

func (s *index) RebuildKnowledgeBaseIndex(ctx context.Context, knowledgeBaseID int64) error {
	knowledgeBase := repositories.KnowledgeBaseRepository.Get(sqls.DB(), knowledgeBaseID)
	if knowledgeBase == nil {
		return fmt.Errorf("knowledge base not found: %d", knowledgeBaseID)
	}

	if err := s.resetKnowledgeBaseIndexStorage(ctx, knowledgeBaseID); err != nil {
		return err
	}

	documents := repositories.KnowledgeDocumentRepository.Find(sqls.DB(), sqls.NewCnd().
		Eq("knowledge_base_id", knowledgeBaseID).
		Where("status != ?", enums.StatusDeleted))
	if len(documents) == 0 {
		slog.Info("No documents found in knowledge base, nothing to rebuild", "knowledge_base_id", knowledgeBaseID)
		return nil
	}

	documentIDs := make([]int64, 0, len(documents))
	for _, doc := range documents {
		documentIDs = append(documentIDs, doc.ID)
	}
	if err := s.markKnowledgeBaseDocumentsIndexPending(knowledgeBaseID, documentIDs); err != nil {
		slog.Error("Failed to mark knowledge base documents index as pending", "knowledge_base_id", knowledgeBaseID, "error", err)
	}

	slog.Info("Rebuilding knowledge base index", "knowledge_base_id", knowledgeBaseID, "document_count", len(documents))

	successCount := 0
	failedCount := 0
	for _, doc := range documents {
		if err := s.IndexDocumentByID(ctx, doc.ID); err != nil {
			slog.Error("Failed to index document", "document_id", doc.ID, "error", err)
			failedCount++
		} else {
			successCount++
		}
	}

	slog.Info("Knowledge base index rebuild completed",
		"knowledge_base_id", knowledgeBaseID,
		"success_count", successCount,
		"failed_count", failedCount)

	return nil
}

func (s *index) markDocumentIndexPending(documentID int64) error {
	return repositories.KnowledgeDocumentRepository.Updates(sqls.DB(), documentID, map[string]any{
		"index_status": enums.KnowledgeDocumentIndexStatusPending,
		"indexed_at":   nil,
		"index_error":  "",
		"updated_at":   time.Now(),
	})
}

func (s *index) markDocumentIndexIndexed(documentID int64) error {
	now := time.Now()
	return repositories.KnowledgeDocumentRepository.Updates(sqls.DB(), documentID, map[string]any{
		"index_status": enums.KnowledgeDocumentIndexStatusIndexed,
		"indexed_at":   &now,
		"index_error":  "",
		"updated_at":   now,
	})
}

func (s *index) markDocumentIndexFailed(documentID int64, err error) error {
	return repositories.KnowledgeDocumentRepository.Updates(sqls.DB(), documentID, map[string]any{
		"index_status": enums.KnowledgeDocumentIndexStatusFailed,
		"index_error":  truncateIndexError(err),
		"updated_at":   time.Now(),
	})
}

func (s *index) markKnowledgeBaseDocumentsIndexPending(knowledgeBaseID int64, documentIDs []int64) error {
	if len(documentIDs) == 0 {
		return nil
	}
	return sqls.DB().Model(&models.KnowledgeDocument{}).
		Where("knowledge_base_id = ?", knowledgeBaseID).
		Where("id IN ?", documentIDs).
		Updates(map[string]any{
			"index_status": enums.KnowledgeDocumentIndexStatusPending,
			"indexed_at":   nil,
			"index_error":  "",
			"updated_at":   time.Now(),
		}).Error
}

func truncateIndexError(err error) string {
	if err == nil {
		return ""
	}
	message := err.Error()
	if len(message) <= 1000 {
		return message
	}
	return message[:1000]
}

func (s *index) resetKnowledgeBaseIndexStorage(ctx context.Context, knowledgeBaseID int64) error {
	collectionName := s.getCollectionName()
	provider := vectordb.GetProvider()
	if provider == nil {
		return fmt.Errorf("vectordb provider not initialized")
	}

	chunks := repositories.KnowledgeChunkRepository.Find(sqls.DB(), sqls.NewCnd().Eq("knowledge_base_id", knowledgeBaseID))
	vectorIDs := make([]string, 0, len(chunks))
	for _, chunk := range chunks {
		if strs.IsNotBlank(chunk.VectorID) {
			vectorIDs = append(vectorIDs, chunk.VectorID)
		}
	}
	if len(vectorIDs) > 0 {
		if err := provider.DeleteVectors(ctx, collectionName, vectorIDs); err != nil {
			return fmt.Errorf("failed to delete vectors for knowledge base %d before rebuild: %w", knowledgeBaseID, err)
		}
	}

	if err := sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		return ctx.Tx.Where("knowledge_base_id = ?", knowledgeBaseID).Delete(&models.KnowledgeChunk{}).Error
	}); err != nil {
		return fmt.Errorf("failed to clear chunks before rebuild: %w", err)
	}

	slog.Info("Knowledge base index storage reset",
		"knowledge_base_id", knowledgeBaseID,
		"collection", collectionName)
	return nil
}
