package services

import (
	"strings"
	"time"

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

var AIConfigService = newAIConfigService()

func newAIConfigService() *aIConfigService {
	return &aIConfigService{}
}

type aIConfigService struct {
}

func (s *aIConfigService) Get(id int64) *models.AIConfig {
	return repositories.AIConfigRepository.Get(sqls.DB(), id)
}

func (s *aIConfigService) Take(where ...interface{}) *models.AIConfig {
	return repositories.AIConfigRepository.Take(sqls.DB(), where...)
}

func (s *aIConfigService) Find(cnd *sqls.Cnd) []models.AIConfig {
	return repositories.AIConfigRepository.Find(sqls.DB(), cnd)
}

func (s *aIConfigService) FindOne(cnd *sqls.Cnd) *models.AIConfig {
	return repositories.AIConfigRepository.FindOne(sqls.DB(), cnd)
}

func (s *aIConfigService) FindPageByParams(params *params.QueryParams) (list []models.AIConfig, paging *sqls.Paging) {
	return repositories.AIConfigRepository.FindPageByParams(sqls.DB(), params)
}

func (s *aIConfigService) FindPageByCnd(cnd *sqls.Cnd) (list []models.AIConfig, paging *sqls.Paging) {
	return repositories.AIConfigRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *aIConfigService) Count(cnd *sqls.Cnd) int64 {
	return repositories.AIConfigRepository.Count(sqls.DB(), cnd)
}

func (s *aIConfigService) Create(t *models.AIConfig) error {
	return repositories.AIConfigRepository.Create(sqls.DB(), t)
}

func (s *aIConfigService) Update(t *models.AIConfig) error {
	return repositories.AIConfigRepository.Update(sqls.DB(), t)
}

func (s *aIConfigService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.AIConfigRepository.Updates(sqls.DB(), id, columns)
}

func (s *aIConfigService) UpdateColumn(id int64, name string, value interface{}) error {
	return repositories.AIConfigRepository.UpdateColumn(sqls.DB(), id, name, value)
}

func (s *aIConfigService) Delete(id int64) {
	repositories.AIConfigRepository.Delete(sqls.DB(), id)
}

func (s *aIConfigService) CreateAIConfig(req request.CreateAIConfigRequest, operator *dto.AuthPrincipal) (*models.AIConfig, error) {
	if operator == nil {
		return nil, errorsx.Unauthorized("未登录或登录已过期")
	}
	item, err := s.buildAIConfigModel(req)
	if err != nil {
		return nil, err
	}

	item.Status = enums.StatusDisabled
	item.SortNo = s.nextSortNo()
	item.AuditFields = utils.BuildAuditFields(operator)
	if err := sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		if err := ctx.Tx.Create(item).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *aIConfigService) UpdateAIConfig(req request.UpdateAIConfigRequest, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	current := s.Get(req.ID)
	if current == nil {
		return errorsx.InvalidParam("AI配置不存在")
	}
	item, err := s.buildAIConfigModel(req.CreateAIConfigRequest)
	if err != nil {
		return err
	}

	return repositories.AIConfigRepository.Updates(sqls.DB(), req.ID, map[string]any{
		"name":               item.Name,
		"provider":           item.Provider,
		"base_url":           item.BaseURL,
		"api_key":            item.APIKey,
		"model_type":         item.ModelType,
		"model_name":         item.ModelName,
		"dimension":          item.Dimension,
		"max_context_tokens": item.MaxContextTokens,
		"max_output_tokens":  item.MaxOutputTokens,
		"timeout_ms":         item.TimeoutMS,
		"max_retry_count":    item.MaxRetryCount,
		"rpm_limit":          item.RPMLimit,
		"tpm_limit":          item.TPMLimit,
		"remark":             item.Remark,
		"update_user_id":     operator.UserID,
		"update_user_name":   operator.Username,
		"updated_at":         time.Now(),
	})
}

func (s *aIConfigService) DeleteAIConfig(id int64, operator *dto.AuthPrincipal) error {
	current := s.Get(id)
	if current == nil {
		return nil
	}
	if current.Status == enums.StatusOk {
		return errorsx.Forbidden("启用中的AI配置不允许删除")
	}
	return repositories.AIConfigRepository.Updates(sqls.DB(), id, map[string]any{
		"status":           enums.StatusDeleted,
		"update_user_id":   operator.UserID,
		"update_user_name": operator.Username,
		"updated_at":       time.Now(),
	})
}

func (s *aIConfigService) UpdateStatus(id int64, status enums.Status, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	current := s.Get(id)
	if current == nil {
		return errorsx.InvalidParam("AI配置不存在")
	}
	if status != enums.StatusOk && status != enums.StatusDisabled {
		return errorsx.InvalidParam("状态值不合法")
	}

	return sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		if status == enums.StatusOk {
			if err := s.disableOthersByModelTypeWithTx(ctx, current.ModelType, id); err != nil {
				return err
			}
		}
		return repositories.AIConfigRepository.Updates(ctx.Tx, id, map[string]any{
			"status":           status,
			"update_user_id":   operator.UserID,
			"update_user_name": operator.Username,
			"updated_at":       time.Now(),
		})
	})
}

func (s *aIConfigService) UpdateSort(ids []int64) error {
	return sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		for i, id := range ids {
			if err := repositories.AIConfigRepository.UpdateColumn(ctx.Tx, id, "sort_no", i); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *aIConfigService) buildAIConfigModel(req request.CreateAIConfigRequest) (*models.AIConfig, error) {
	name := strings.TrimSpace(req.Name)
	baseURL := strings.TrimSpace(req.BaseURL)
	modelName := strings.TrimSpace(req.ModelName)

	if name == "" {
		return nil, errorsx.InvalidParam("配置名称不能为空")
	}
	if strs.IsBlank(string(req.Provider)) {
		return nil, errorsx.InvalidParam("供应商不能为空")
	}
	if baseURL == "" {
		return nil, errorsx.InvalidParam("基础地址不能为空")
	}
	if strs.IsBlank(string(req.ModelType)) {
		return nil, errorsx.InvalidParam("模型类型不能为空")
	}
	if modelName == "" {
		return nil, errorsx.InvalidParam("模型名称不能为空")
	}
	if req.Dimension < 0 {
		req.Dimension = 0
	}
	if req.MaxContextTokens < 0 {
		req.MaxContextTokens = 0
	}
	if req.MaxOutputTokens < 0 {
		req.MaxOutputTokens = 0
	}
	if req.TimeoutMS <= 0 {
		req.TimeoutMS = 30000
	}
	if req.MaxRetryCount < 0 {
		req.MaxRetryCount = 0
	}
	if req.RPMLimit < 0 {
		req.RPMLimit = 0
	}
	if req.TPMLimit < 0 {
		req.TPMLimit = 0
	}

	return &models.AIConfig{
		Name:             name,
		Provider:         req.Provider,
		BaseURL:          baseURL,
		APIKey:           strings.TrimSpace(req.APIKey),
		ModelType:        req.ModelType,
		ModelName:        modelName,
		Dimension:        req.Dimension,
		MaxContextTokens: req.MaxContextTokens,
		MaxOutputTokens:  req.MaxOutputTokens,
		TimeoutMS:        req.TimeoutMS,
		MaxRetryCount:    req.MaxRetryCount,
		RPMLimit:         req.RPMLimit,
		TPMLimit:         req.TPMLimit,
		Remark:           strings.TrimSpace(req.Remark),
	}, nil
}

func (s *aIConfigService) nextSortNo() int {
	if latest := s.FindOne(sqls.NewCnd().Desc("sort_no").Desc("id")); latest != nil {
		return latest.SortNo + 1
	}
	return 1
}

func (s *aIConfigService) disableOthersByModelTypeWithTx(ctx *sqls.TxContext, modelType enums.AIModelType, excludeID int64) error {
	query := ctx.Tx.Model(&models.AIConfig{}).Where("model_type = ?", modelType)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}
	return query.Updates(map[string]any{
		"status":     int(enums.StatusDisabled),
		"updated_at": time.Now(),
	}).Error
}
