package rag

import (
	"time"

	"cs-agent/internal/models"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/repositories"

	"github.com/mlogclub/simple/sqls"
)

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

func (s *index) markFAQIndexPending(faqID int64) error {
	return repositories.KnowledgeFAQRepository.Updates(sqls.DB(), faqID, map[string]any{
		"index_status": enums.KnowledgeDocumentIndexStatusPending,
		"indexed_at":   nil,
		"index_error":  "",
		"updated_at":   time.Now(),
	})
}

func (s *index) markFAQIndexIndexed(faqID int64) error {
	now := time.Now()
	return repositories.KnowledgeFAQRepository.Updates(sqls.DB(), faqID, map[string]any{
		"index_status": enums.KnowledgeDocumentIndexStatusIndexed,
		"indexed_at":   &now,
		"index_error":  "",
		"updated_at":   now,
	})
}

func (s *index) markFAQIndexFailed(faqID int64, err error) error {
	return repositories.KnowledgeFAQRepository.Updates(sqls.DB(), faqID, map[string]any{
		"index_status": enums.KnowledgeDocumentIndexStatusFailed,
		"index_error":  truncateIndexError(err),
		"updated_at":   time.Now(),
	})
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
