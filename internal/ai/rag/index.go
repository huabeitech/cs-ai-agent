package rag

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
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

	chunks, err := s.buildDocumentChunks(ctx, document, knowledgeBase)
	if err != nil {
		return fail(err)
	}

	collectionName := s.getCollectionName()
	provider := vectordb.GetProvider()
	if provider == nil {
		return fail(fmt.Errorf("vectordb provider not initialized"))
	}

	if _, err := ai.Embedding.GetModel(ctx); err != nil {
		return fail(fmt.Errorf("failed to get embedding model: %w", err))
	}

	existingVectorIDs := collectExistingVectorIDs(existingChunks)
	vectors, chunkModels, dimension, err := s.prepareDocumentVectors(ctx, knowledgeBase, document, chunks)
	if err != nil {
		return fail(err)
	}

	if err := s.ensureCollection(ctx, provider, collectionName, dimension); err != nil {
		return fail(err)
	}

	if len(existingVectorIDs) > 0 {
		if err := provider.DeleteVectors(ctx, collectionName, existingVectorIDs); err != nil {
			return fail(fmt.Errorf("failed to delete old vectors: %w", err))
		}
	}

	if err := provider.UpsertVectors(ctx, collectionName, vectors); err != nil {
		return fail(fmt.Errorf("failed to upsert vectors: %w", err))
	}

	if err := s.replaceDocumentChunks(document.ID, chunkModels); err != nil {
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

func (s *index) IndexFAQByID(ctx context.Context, faqID int64) error {
	faq := repositories.KnowledgeFAQRepository.Get(sqls.DB(), faqID)
	if faq == nil {
		return fmt.Errorf("faq not found: %d", faqID)
	}
	if err := s.markFAQIndexPending(faq.ID); err != nil {
		slog.Error("Failed to mark knowledge faq index as pending", "faq_id", faq.ID, "error", err)
	}
	fail := func(err error) error {
		if updateErr := s.markFAQIndexFailed(faq.ID, err); updateErr != nil {
			slog.Error("Failed to mark knowledge faq index as failed", "faq_id", faq.ID, "error", updateErr)
		}
		return err
	}
	knowledgeBase := repositories.KnowledgeBaseRepository.Get(sqls.DB(), faq.KnowledgeBaseID)
	if knowledgeBase == nil {
		return fail(fmt.Errorf("knowledge base not found: %d", faq.KnowledgeBaseID))
	}
	if knowledgeBase.KnowledgeType != string(enums.KnowledgeBaseTypeFAQ) {
		return fail(fmt.Errorf("knowledge base %d is not faq type", knowledgeBase.ID))
	}
	existingChunks := repositories.KnowledgeChunkRepository.FindByFaqID(sqls.DB(), faq.ID)
	content := buildFAQChunkContent(faq)
	if content == "" {
		return fail(fmt.Errorf("faq content is empty"))
	}

	provider := vectordb.GetProvider()
	if provider == nil {
		return fail(fmt.Errorf("vectordb provider not initialized"))
	}
	if _, err := ai.Embedding.GetModel(ctx); err != nil {
		return fail(fmt.Errorf("failed to get embedding model: %w", err))
	}
	vector, chunkModel, dimension, err := s.prepareFAQVector(ctx, knowledgeBase, faq, content)
	if err != nil {
		return fail(err)
	}

	collectionName := s.getCollectionName()
	if err := s.ensureCollection(ctx, provider, collectionName, dimension); err != nil {
		return fail(err)
	}

	existingVectorIDs := make([]string, 0, len(existingChunks))
	for _, chunk := range existingChunks {
		if strs.IsNotBlank(chunk.VectorID) {
			existingVectorIDs = append(existingVectorIDs, chunk.VectorID)
		}
	}
	if len(existingVectorIDs) > 0 {
		if err := provider.DeleteVectors(ctx, collectionName, existingVectorIDs); err != nil {
			return fail(fmt.Errorf("failed to delete old vectors: %w", err))
		}
	}

	if err := provider.UpsertVectors(ctx, collectionName, []vectordb.Vector{vector}); err != nil {
		return fail(fmt.Errorf("failed to upsert vectors: %w", err))
	}

	if err := s.replaceFAQChunk(faq.ID, &chunkModel); err != nil {
		return fail(fmt.Errorf("failed to save faq chunk: %w", err))
	}
	if err := s.markFAQIndexIndexed(faq.ID); err != nil {
		slog.Error("Failed to mark knowledge faq index as indexed", "faq_id", faq.ID, "error", err)
	}
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

func (s *index) RemoveFAQIndex(ctx context.Context, faqID int64) error {
	faq := repositories.KnowledgeFAQRepository.Get(sqls.DB(), faqID)
	if faq == nil {
		return nil
	}
	chunks := repositories.KnowledgeChunkRepository.FindByFaqID(sqls.DB(), faqID)
	return s.removeFAQIndexByChunks(ctx, faq.KnowledgeBaseID, faqID, chunks)
}

func (s *index) RemoveFAQIndexByChunkModels(ctx context.Context, knowledgeBaseID int64, faqID int64, chunks []models.KnowledgeChunk) error {
	return s.removeFAQIndexByChunks(ctx, knowledgeBaseID, faqID, chunks)
}

func (s *index) removeFAQIndexByChunks(ctx context.Context, knowledgeBaseID int64, faqID int64, chunks []models.KnowledgeChunk) error {
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
			slog.Error("Failed to delete faq vectors", "error", err)
		}
	}
	if err := sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		return ctx.Tx.Where("faq_id = ?", faqID).Delete(&models.KnowledgeChunk{}).Error
	}); err != nil {
		return fmt.Errorf("failed to delete faq chunks: %w", err)
	}
	slog.Info("FAQ index removed", "faq_id", faqID, "chunks_removed", len(chunks))
	return nil
}

func (s *index) getCollectionName() string {
	return knowledgeCollectionName
}

func buildKnowledgeChunkVectorID(knowledgeBaseID int64, documentID int64, chunkNo int) string {
	raw := fmt.Sprintf("kb:%d:doc:%d:chunk:%d", knowledgeBaseID, documentID, chunkNo)
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(raw)).String()
}

func buildKnowledgeFAQChunkVectorID(knowledgeBaseID int64, faqID int64, chunkNo int) string {
	raw := fmt.Sprintf("kb:%d:faq:%d:chunk:%d", knowledgeBaseID, faqID, chunkNo)
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

	successCount := 0
	failedCount := 0
	if knowledgeBase.KnowledgeType == string(enums.KnowledgeBaseTypeFAQ) {
		faqs := repositories.KnowledgeFAQRepository.Find(sqls.DB(), sqls.NewCnd().
			Eq("knowledge_base_id", knowledgeBaseID).
			Where("status != ?", enums.StatusDeleted))
		if len(faqs) == 0 {
			slog.Info("No faqs found in knowledge base, nothing to rebuild", "knowledge_base_id", knowledgeBaseID)
			return nil
		}
		slog.Info("Rebuilding faq knowledge base index", "knowledge_base_id", knowledgeBaseID, "faq_count", len(faqs))
		for _, faq := range faqs {
			if err := s.IndexFAQByID(ctx, faq.ID); err != nil {
				slog.Error("Failed to index faq", "faq_id", faq.ID, "error", err)
				failedCount++
			} else {
				successCount++
			}
		}
	} else {
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
		for _, doc := range documents {
			if err := s.IndexDocumentByID(ctx, doc.ID); err != nil {
				slog.Error("Failed to index document", "document_id", doc.ID, "error", err)
				failedCount++
			} else {
				successCount++
			}
		}
	}

	slog.Info("Knowledge base index rebuild completed",
		"knowledge_base_id", knowledgeBaseID,
		"success_count", successCount,
		"failed_count", failedCount)

	return nil
}

func buildFAQChunkContent(faq *models.KnowledgeFAQ) string {
	if faq == nil {
		return ""
	}
	parts := []string{fmt.Sprintf("问题：%s", faq.Question)}
	var similarQuestions []string
	if faq.SimilarQuestions != "" {
		_ = json.Unmarshal([]byte(faq.SimilarQuestions), &similarQuestions)
	}
	if len(similarQuestions) > 0 {
		parts = append(parts, fmt.Sprintf("相似问：%s", joinSimilarQuestions(similarQuestions)))
	}
	parts = append(parts, fmt.Sprintf("回答：%s", faq.Answer))
	content := ""
	for _, part := range parts {
		if part == "" {
			continue
		}
		if content != "" {
			content += "\n"
		}
		content += part
	}
	return content
}

func joinSimilarQuestions(items []string) string {
	result := ""
	for _, item := range items {
		if item == "" {
			continue
		}
		if result != "" {
			result += "；"
		}
		result += item
	}
	return result
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
