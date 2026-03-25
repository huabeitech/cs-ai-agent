package services

import (
	"time"

	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/errorsx"
	"cs-agent/internal/pkg/utils"
	"cs-agent/internal/repositories"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
)

var KnowledgeBaseService = newKnowledgeBaseService()

func newKnowledgeBaseService() *knowledgeBaseService {
	return &knowledgeBaseService{}
}

type knowledgeBaseService struct {
}

func (s *knowledgeBaseService) Get(id int64) *models.KnowledgeBase {
	return repositories.KnowledgeBaseRepository.Get(sqls.DB(), id)
}

func (s *knowledgeBaseService) Take(where ...interface{}) *models.KnowledgeBase {
	return repositories.KnowledgeBaseRepository.Take(sqls.DB(), where...)
}

func (s *knowledgeBaseService) Find(cnd *sqls.Cnd) []models.KnowledgeBase {
	return repositories.KnowledgeBaseRepository.Find(sqls.DB(), cnd)
}

func (s *knowledgeBaseService) FindOne(cnd *sqls.Cnd) *models.KnowledgeBase {
	return repositories.KnowledgeBaseRepository.FindOne(sqls.DB(), cnd)
}

func (s *knowledgeBaseService) FindPageByParams(params *params.QueryParams) (list []models.KnowledgeBase, paging *sqls.Paging) {
	return repositories.KnowledgeBaseRepository.FindPageByParams(sqls.DB(), params)
}

func (s *knowledgeBaseService) FindPageByCnd(cnd *sqls.Cnd) (list []models.KnowledgeBase, paging *sqls.Paging) {
	return repositories.KnowledgeBaseRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *knowledgeBaseService) Count(cnd *sqls.Cnd) int64 {
	return repositories.KnowledgeBaseRepository.Count(sqls.DB(), cnd)
}

func (s *knowledgeBaseService) Create(t *models.KnowledgeBase) error {
	return repositories.KnowledgeBaseRepository.Create(sqls.DB(), t)
}

func (s *knowledgeBaseService) Update(t *models.KnowledgeBase) error {
	return repositories.KnowledgeBaseRepository.Update(sqls.DB(), t)
}

func (s *knowledgeBaseService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.KnowledgeBaseRepository.Updates(sqls.DB(), id, columns)
}

func (s *knowledgeBaseService) UpdateColumn(id int64, name string, value interface{}) error {
	return repositories.KnowledgeBaseRepository.UpdateColumn(sqls.DB(), id, name, value)
}

func (s *knowledgeBaseService) Delete(id int64) {
	repositories.KnowledgeBaseRepository.Delete(sqls.DB(), id)
}

func (s *knowledgeBaseService) CreateKnowledgeBase(req request.CreateKnowledgeBaseRequest, operator *dto.AuthPrincipal) (*models.KnowledgeBase, error) {
	if operator == nil {
		return nil, errorsx.Unauthorized("未登录或登录已过期")
	}
	item, err := s.buildKnowledgeBaseModel(req)
	if err != nil {
		return nil, err
	}
	item.Status = enums.StatusOk
	item.AuditFields = utils.BuildAuditFields(operator)
	if err := sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		return ctx.Tx.Create(item).Error
	}); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *knowledgeBaseService) UpdateKnowledgeBase(req request.UpdateKnowledgeBaseRequest, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	current := s.Get(req.ID)
	if current == nil {
		return errorsx.InvalidParam("知识库不存在")
	}
	item, err := s.buildKnowledgeBaseModel(req.CreateKnowledgeBaseRequest)
	if err != nil {
		return err
	}
	return sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		return ctx.Tx.Model(&models.KnowledgeBase{}).Where("id = ?", req.ID).Updates(map[string]any{
			"name":                    item.Name,
			"description":             item.Description,
			"default_top_k":           item.DefaultTopK,
			"default_score_threshold": item.DefaultScoreThreshold,
			"default_rerank_limit":    item.DefaultRerankLimit,
			"chunk_provider":          item.ChunkProvider,
			"chunk_target_tokens":     item.ChunkTargetTokens,
			"chunk_max_tokens":        item.ChunkMaxTokens,
			"chunk_overlap_tokens":    item.ChunkOverlapTokens,
			"answer_mode":             item.AnswerMode,
			"fallback_mode":           item.FallbackMode,
			"remark":                  item.Remark,
			"update_user_id":          operator.UserID,
			"update_user_name":        operator.Username,
			"updated_at":              time.Now(),
		}).Error
	})
}

func (s *knowledgeBaseService) DeleteKnowledgeBase(id int64) error {
	current := s.Get(id)
	if current == nil {
		return errorsx.InvalidParam("知识库不存在")
	}
	docCount := repositories.KnowledgeDocumentRepository.CountByKnowledgeBaseID(sqls.DB(), id)
	if docCount > 0 {
		return errorsx.InvalidParam("知识库下存在文档，无法删除")
	}
	return sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		ctx.Tx.Delete(&models.KnowledgeBase{}, "id = ?", id)
		return nil
	})
}

func (s *knowledgeBaseService) UpdateSort(ids []int64) error {
	return sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		for i, id := range ids {
			err := ctx.Tx.Model(&models.KnowledgeBase{}).Where("id = ?", id).Update("sort_no", i).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *knowledgeBaseService) buildKnowledgeBaseModel(req request.CreateKnowledgeBaseRequest) (*models.KnowledgeBase, error) {
	item := &models.KnowledgeBase{
		Name:                  req.Name,
		Description:           req.Description,
		DefaultTopK:           req.DefaultTopK,
		DefaultScoreThreshold: req.DefaultScoreThreshold,
		DefaultRerankLimit:    req.DefaultRerankLimit,
		ChunkProvider:         req.ChunkProvider,
		ChunkTargetTokens:     req.ChunkTargetTokens,
		ChunkMaxTokens:        req.ChunkMaxTokens,
		ChunkOverlapTokens:    req.ChunkOverlapTokens,
		AnswerMode:            req.AnswerMode,
		FallbackMode:          req.FallbackMode,
		Remark:                req.Remark,
	}
	if item.DefaultTopK == 0 {
		item.DefaultTopK = 10
	}
	if item.DefaultScoreThreshold == 0 {
		item.DefaultScoreThreshold = 0.2
	}
	if item.DefaultRerankLimit == 0 {
		item.DefaultRerankLimit = 5
	}
	if item.ChunkProvider == "" {
		item.ChunkProvider = string(enums.KnowledgeChunkProviderStructured)
	}
	if !isValidChunkProvider(item.ChunkProvider) {
		return nil, errorsx.InvalidParam("分块策略不支持")
	}
	if item.ChunkTargetTokens == 0 {
		item.ChunkTargetTokens = 300
	}
	if item.ChunkMaxTokens == 0 {
		item.ChunkMaxTokens = 400
	}
	if item.ChunkMaxTokens < item.ChunkTargetTokens {
		item.ChunkMaxTokens = item.ChunkTargetTokens
	}
	if item.ChunkOverlapTokens == 0 {
		item.ChunkOverlapTokens = 40
	}
	if item.AnswerMode == 0 {
		item.AnswerMode = 1
	}
	if item.FallbackMode == 0 {
		item.FallbackMode = 1
	}
	return item, nil
}

func isValidChunkProvider(provider string) bool {
	switch provider {
	case string(enums.KnowledgeChunkProviderFixed),
		string(enums.KnowledgeChunkProviderStructured),
		string(enums.KnowledgeChunkProviderFAQ),
		string(enums.KnowledgeChunkProviderSemantic):
		return true
	default:
		return false
	}
}
