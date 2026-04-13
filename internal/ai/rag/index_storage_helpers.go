package rag

import (
	"context"
	"fmt"
	"log/slog"

	"cs-agent/internal/ai/rag/vectordb"
	"cs-agent/internal/models"

	"github.com/mlogclub/simple/sqls"
)

func (s *index) ensureCollection(ctx context.Context, provider vectordb.Provider, collectionName string, dimension int) error {
	collectionInfo, err := provider.GetCollection(ctx, collectionName)
	if err == nil && collectionInfo != nil {
		return nil
	}
	if dimension <= 0 {
		return fmt.Errorf("invalid embedding dimension: %d", dimension)
	}
	if err := provider.CreateCollection(ctx, collectionName, dimension); err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}
	slog.Info("Created collection for knowledge base", "collection", collectionName, "dimension", dimension)
	return nil
}

func (s *index) replaceDocumentChunks(documentID int64, chunkModels []models.KnowledgeChunk) error {
	return sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		if err := ctx.Tx.Where("document_id = ?", documentID).Delete(&models.KnowledgeChunk{}).Error; err != nil {
			return err
		}
		for _, chunk := range chunkModels {
			if err := ctx.Tx.Create(&chunk).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *index) replaceFAQChunk(faqID int64, chunkModel *models.KnowledgeChunk) error {
	if chunkModel == nil {
		return nil
	}
	return sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		if err := ctx.Tx.Where("faq_id = ?", faqID).Delete(&models.KnowledgeChunk{}).Error; err != nil {
			return err
		}
		return ctx.Tx.Create(chunkModel).Error
	})
}
