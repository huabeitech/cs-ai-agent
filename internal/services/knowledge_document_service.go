package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"time"

	"cs-agent/internal/ai/rag"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/errorsx"
	"cs-agent/internal/pkg/utils"
	"cs-agent/internal/repositories"

	"github.com/mlogclub/simple/common/strs"
	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
)

var KnowledgeDocumentService = newKnowledgeDocumentService()

func newKnowledgeDocumentService() *knowledgeDocumentService {
	return &knowledgeDocumentService{}
}

type knowledgeDocumentService struct {
}

func (s *knowledgeDocumentService) Get(id int64) *models.KnowledgeDocument {
	return repositories.KnowledgeDocumentRepository.Get(sqls.DB(), id)
}

func (s *knowledgeDocumentService) Take(where ...interface{}) *models.KnowledgeDocument {
	return repositories.KnowledgeDocumentRepository.Take(sqls.DB(), where...)
}

func (s *knowledgeDocumentService) Find(cnd *sqls.Cnd) []models.KnowledgeDocument {
	return repositories.KnowledgeDocumentRepository.Find(sqls.DB(), cnd)
}

func (s *knowledgeDocumentService) FindOne(cnd *sqls.Cnd) *models.KnowledgeDocument {
	return repositories.KnowledgeDocumentRepository.FindOne(sqls.DB(), cnd)
}

func (s *knowledgeDocumentService) FindPageByParams(params *params.QueryParams) (list []models.KnowledgeDocument, paging *sqls.Paging) {
	return repositories.KnowledgeDocumentRepository.FindPageByParams(sqls.DB(), params)
}

func (s *knowledgeDocumentService) FindPageByCnd(cnd *sqls.Cnd) (list []models.KnowledgeDocument, paging *sqls.Paging) {
	return repositories.KnowledgeDocumentRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *knowledgeDocumentService) FindPageListByCnd(cnd *sqls.Cnd) (list []models.KnowledgeDocument, paging *sqls.Paging) {
	return repositories.KnowledgeDocumentRepository.FindPageListByCnd(sqls.DB(), cnd)
}

func (s *knowledgeDocumentService) Count(cnd *sqls.Cnd) int64 {
	return repositories.KnowledgeDocumentRepository.Count(sqls.DB(), cnd)
}

func (s *knowledgeDocumentService) Create(t *models.KnowledgeDocument) error {
	return repositories.KnowledgeDocumentRepository.Create(sqls.DB(), t)
}

func (s *knowledgeDocumentService) Update(t *models.KnowledgeDocument) error {
	return repositories.KnowledgeDocumentRepository.Update(sqls.DB(), t)
}

func (s *knowledgeDocumentService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.KnowledgeDocumentRepository.Updates(sqls.DB(), id, columns)
}

func (s *knowledgeDocumentService) UpdateColumn(id int64, name string, value interface{}) error {
	return repositories.KnowledgeDocumentRepository.UpdateColumn(sqls.DB(), id, name, value)
}

func (s *knowledgeDocumentService) Delete(id int64) {
	repositories.KnowledgeDocumentRepository.Delete(sqls.DB(), id)
}

func (s *knowledgeDocumentService) CreateKnowledgeDocument(req request.CreateKnowledgeDocumentRequest, operator *dto.AuthPrincipal) (*models.KnowledgeDocument, error) {
	if operator == nil {
		return nil, errorsx.Unauthorized("未登录或登录已过期")
	}
	kb := KnowledgeBaseService.Get(req.KnowledgeBaseID)
	if kb == nil {
		return nil, errorsx.InvalidParam("知识库不存在")
	}
	if kb.KnowledgeType == string(enums.KnowledgeBaseTypeFAQ) {
		return nil, errorsx.InvalidParam("FAQ知识库不支持文档")
	}
	item, err := s.buildKnowledgeDocumentModel(req)
	if err != nil {
		return nil, err
	}
	item.Status = enums.StatusOk
	item.IndexStatus = enums.KnowledgeDocumentIndexStatusPending
	item.IndexError = ""
	item.IndexedAt = nil
	item.AuditFields = utils.BuildAuditFields(operator)
	if err := sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		if err := ctx.Tx.Create(item).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	if err := rag.Index.IndexDocumentByID(context.Background(), item.ID); err != nil {
		slog.Error("failed to index created knowledge document", "document_id", item.ID, "error", err)
	}
	item = s.Get(item.ID)
	return item, nil
}

func (s *knowledgeDocumentService) UpdateKnowledgeDocument(req request.UpdateKnowledgeDocumentRequest, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	current := s.Get(req.ID)
	if current == nil {
		return errorsx.InvalidParam("文档不存在")
	}
	kb := KnowledgeBaseService.Get(req.KnowledgeBaseID)
	if kb == nil {
		return errorsx.InvalidParam("知识库不存在")
	}
	if kb.KnowledgeType == string(enums.KnowledgeBaseTypeFAQ) {
		return errorsx.InvalidParam("FAQ知识库不支持文档")
	}
	item, err := s.buildKnowledgeDocumentModel(req.CreateKnowledgeDocumentRequest)
	if err != nil {
		return err
	}
	oldKnowledgeBaseID := current.KnowledgeBaseID
	if err := repositories.KnowledgeDocumentRepository.Updates(sqls.DB(), req.ID, map[string]any{
		"knowledge_base_id": item.KnowledgeBaseID,
		"title":             item.Title,
		"content_type":      item.ContentType,
		"content_hash":      item.ContentHash,
		"content":           item.Content,
		"index_status":      enums.KnowledgeDocumentIndexStatusPending,
		"indexed_at":        nil,
		"index_error":       "",
		"update_user_id":    operator.UserID,
		"update_user_name":  operator.Username,
		"updated_at":        time.Now(),
	}); err != nil {
		return err
	}

	if oldKnowledgeBaseID != item.KnowledgeBaseID {
		if err := rag.Index.RemoveDocumentIndexFromKnowledgeBase(context.Background(), oldKnowledgeBaseID, req.ID); err != nil {
			slog.Error("failed to remove old document index after knowledge base change", "document_id", req.ID, "knowledge_base_id", oldKnowledgeBaseID, "error", err)
		}
	}

	if err := rag.Index.IndexDocumentByID(context.Background(), req.ID); err != nil {
		slog.Error("failed to reindex updated knowledge document", "document_id", req.ID, "error", err)
	}
	return nil
}

func (s *knowledgeDocumentService) DeleteKnowledgeDocument(id int64) error {
	current := s.Get(id)
	if current == nil {
		return errorsx.InvalidParam("文档不存在")
	}
	chunks := repositories.KnowledgeChunkRepository.FindByDocumentID(sqls.DB(), id)
	if err := sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		_ = repositories.KnowledgeDocumentRepository.Updates(ctx.Tx, id, map[string]any{
			"status":     enums.StatusDeleted,
			"updated_at": time.Now(),
		})
		ctx.Tx.Delete(&models.KnowledgeChunk{}, "document_id = ?", id)
		return nil
	}); err != nil {
		return err
	}
	return rag.Index.RemoveDocumentIndexByChunkModels(context.Background(), current.KnowledgeBaseID, id, chunks)
}

func (s *knowledgeDocumentService) buildKnowledgeDocumentModel(req request.CreateKnowledgeDocumentRequest) (*models.KnowledgeDocument, error) {
	if strs.IsBlank(string(req.ContentType)) {
		req.ContentType = enums.KnowledgeDocumentContentTypeHTML
	}
	if req.ContentType != enums.KnowledgeDocumentContentTypeHTML && req.ContentType != enums.KnowledgeDocumentContentTypeMarkdown {
		return nil, errorsx.InvalidParam("内容类型不支持")
	}

	plainText := rag.ExtractPlainText(req.Content, req.ContentType)
	item := &models.KnowledgeDocument{
		KnowledgeBaseID: req.KnowledgeBaseID,
		Title:           req.Title,
		ContentType:     req.ContentType,
		Content:         req.Content,
	}
	if plainText != "" {
		hash := sha256.Sum256([]byte(plainText))
		item.ContentHash = hex.EncodeToString(hash[:])
	}
	return item, nil
}
