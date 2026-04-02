package services

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"cs-agent/internal/ai/rag"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/errorsx"
	"cs-agent/internal/pkg/utils"
	"cs-agent/internal/repositories"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
)

var KnowledgeFAQService = newKnowledgeFAQService()

func newKnowledgeFAQService() *knowledgeFAQService {
	return &knowledgeFAQService{}
}

type knowledgeFAQService struct{}

func (s *knowledgeFAQService) Get(id int64) *models.KnowledgeFAQ {
	return repositories.KnowledgeFAQRepository.Get(sqls.DB(), id)
}

func (s *knowledgeFAQService) FindPageByCnd(cnd *sqls.Cnd) (list []models.KnowledgeFAQ, paging *sqls.Paging) {
	return repositories.KnowledgeFAQRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *knowledgeFAQService) FindPageByParams(queryParams *params.QueryParams) (list []models.KnowledgeFAQ, paging *sqls.Paging) {
	return repositories.KnowledgeFAQRepository.FindPageByParams(sqls.DB(), queryParams)
}

func (s *knowledgeFAQService) CreateKnowledgeFAQ(req request.CreateKnowledgeFAQRequest, operator *dto.AuthPrincipal) (*models.KnowledgeFAQ, error) {
	if operator == nil {
		return nil, errorsx.Unauthorized("未登录或登录已过期")
	}
	kb, err := s.requireFAQKnowledgeBase(req.KnowledgeBaseID)
	if err != nil {
		return nil, err
	}
	item, err := s.buildKnowledgeFAQModel(req)
	if err != nil {
		return nil, err
	}
	item.Status = kb.Status
	item.AuditFields = utils.BuildAuditFields(operator)
	if err := repositories.KnowledgeFAQRepository.Create(sqls.DB(), item); err != nil {
		return nil, err
	}
	if err := rag.Index.IndexFAQByID(context.Background(), item.ID); err != nil {
		slog.Error("failed to index created knowledge faq", "faq_id", item.ID, "error", err)
		return item, nil
	}
	return s.Get(item.ID), nil
}

func (s *knowledgeFAQService) UpdateKnowledgeFAQ(req request.UpdateKnowledgeFAQRequest, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	current := s.Get(req.ID)
	if current == nil {
		return errorsx.InvalidParam("FAQ不存在")
	}
	if _, err := s.requireFAQKnowledgeBase(req.KnowledgeBaseID); err != nil {
		return err
	}
	item, err := s.buildKnowledgeFAQModel(req.CreateKnowledgeFAQRequest)
	if err != nil {
		return err
	}
	if err := repositories.KnowledgeFAQRepository.Updates(sqls.DB(), req.ID, map[string]any{
		"knowledge_base_id": item.KnowledgeBaseID,
		"question":          item.Question,
		"answer":            item.Answer,
		"similar_questions": item.SimilarQuestions,
		"remark":            item.Remark,
		"update_user_id":    operator.UserID,
		"update_user_name":  operator.Username,
		"updated_at":        time.Now(),
	}); err != nil {
		return err
	}
	return rag.Index.IndexFAQByID(context.Background(), req.ID)
}

func (s *knowledgeFAQService) DeleteKnowledgeFAQ(id int64) error {
	current := s.Get(id)
	if current == nil {
		return errorsx.InvalidParam("FAQ不存在")
	}
	chunks := repositories.KnowledgeChunkRepository.FindByFaqID(sqls.DB(), id)
	if err := sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		ctx.Tx.Delete(&models.KnowledgeFAQ{}, "id = ?", id)
		ctx.Tx.Delete(&models.KnowledgeChunk{}, "faq_id = ?", id)
		return nil
	}); err != nil {
		return err
	}
	return rag.Index.RemoveFAQIndexByChunkModels(context.Background(), current.KnowledgeBaseID, id, chunks)
}

func (s *knowledgeFAQService) buildKnowledgeFAQModel(req request.CreateKnowledgeFAQRequest) (*models.KnowledgeFAQ, error) {
	if req.KnowledgeBaseID <= 0 {
		return nil, errorsx.InvalidParam("知识库不存在")
	}
	if req.Question == "" {
		return nil, errorsx.InvalidParam("问题不能为空")
	}
	if req.Answer == "" {
		return nil, errorsx.InvalidParam("答案不能为空")
	}
	similarQuestions, err := json.Marshal(normalizeSimilarQuestions(req.SimilarQuestions))
	if err != nil {
		return nil, errorsx.InvalidParam("相似问格式不合法")
	}
	return &models.KnowledgeFAQ{
		KnowledgeBaseID:  req.KnowledgeBaseID,
		Question:         req.Question,
		Answer:           req.Answer,
		SimilarQuestions: string(similarQuestions),
		Remark:           req.Remark,
	}, nil
}

func (s *knowledgeFAQService) requireFAQKnowledgeBase(knowledgeBaseID int64) (*models.KnowledgeBase, error) {
	kb := KnowledgeBaseService.Get(knowledgeBaseID)
	if kb == nil {
		return nil, errorsx.InvalidParam("知识库不存在")
	}
	if kb.KnowledgeType != "faq" {
		return nil, errorsx.InvalidParam("当前知识库不是FAQ知识库")
	}
	return kb, nil
}

func normalizeSimilarQuestions(values []string) []string {
	items := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, item := range values {
		value := item
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		items = append(items, value)
	}
	return items
}
