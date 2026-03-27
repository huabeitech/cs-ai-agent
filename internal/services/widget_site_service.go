package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/errorsx"
	"cs-agent/internal/pkg/utils"
	"cs-agent/internal/repositories"
	"strings"
	"time"

	"github.com/mlogclub/simple/common/strs"
	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
)

var WidgetSiteService = newWidgetSiteService()

func newWidgetSiteService() *widgetSiteService {
	return &widgetSiteService{}
}

type widgetSiteService struct {
}

func (s *widgetSiteService) Get(id int64) *models.WidgetSite {
	return repositories.WidgetSiteRepository.Get(sqls.DB(), id)
}

func (s *widgetSiteService) Take(where ...interface{}) *models.WidgetSite {
	return repositories.WidgetSiteRepository.Take(sqls.DB(), where...)
}

func (s *widgetSiteService) Find(cnd *sqls.Cnd) []models.WidgetSite {
	return repositories.WidgetSiteRepository.Find(sqls.DB(), cnd)
}

func (s *widgetSiteService) FindOne(cnd *sqls.Cnd) *models.WidgetSite {
	return repositories.WidgetSiteRepository.FindOne(sqls.DB(), cnd)
}

func (s *widgetSiteService) FindPageByParams(params *params.QueryParams) (list []models.WidgetSite, paging *sqls.Paging) {
	return repositories.WidgetSiteRepository.FindPageByParams(sqls.DB(), params)
}

func (s *widgetSiteService) FindPageByCnd(cnd *sqls.Cnd) (list []models.WidgetSite, paging *sqls.Paging) {
	return repositories.WidgetSiteRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *widgetSiteService) Count(cnd *sqls.Cnd) int64 {
	return repositories.WidgetSiteRepository.Count(sqls.DB(), cnd)
}

func (s *widgetSiteService) Create(t *models.WidgetSite) error {
	return repositories.WidgetSiteRepository.Create(sqls.DB(), t)
}

func (s *widgetSiteService) Update(t *models.WidgetSite) error {
	return repositories.WidgetSiteRepository.Update(sqls.DB(), t)
}

func (s *widgetSiteService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.WidgetSiteRepository.Updates(sqls.DB(), id, columns)
}

func (s *widgetSiteService) UpdateColumn(id int64, name string, value interface{}) error {
	return repositories.WidgetSiteRepository.UpdateColumn(sqls.DB(), id, name, value)
}

func (s *widgetSiteService) Delete(id int64) {
	repositories.WidgetSiteRepository.Delete(sqls.DB(), id)
}

func (s *widgetSiteService) FindEnabledByAppID(appID string) *models.WidgetSite {
	appID = strings.TrimSpace(appID)
	if appID == "" {
		return nil
	}
	return s.Take("app_id = ? AND status = ?", appID, enums.StatusOk)
}

func (s *widgetSiteService) CreateSite(req request.CreateWidgetSiteRequest, operator *dto.AuthPrincipal) (*models.WidgetSite, error) {
	if operator == nil {
		return nil, errorsx.Unauthorized("未登录或登录已过期")
	}
	aiAgentID, err := s.validateAIAgentID(req.AIAgentID)
	if err != nil {
		return nil, err
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errorsx.InvalidParam("站点名称不能为空")
	}
	item := &models.WidgetSite{
		AIAgentID:   aiAgentID,
		Name:        name,
		AppID:       strs.UUID(),
		Status:      enums.StatusOk,
		Remark:      strings.TrimSpace(req.Remark),
		AuditFields: utils.BuildAuditFields(operator),
	}
	if err := repositories.WidgetSiteRepository.Create(sqls.DB(), item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *widgetSiteService) UpdateSite(req request.UpdateWidgetSiteRequest, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	item := s.Get(req.ID)
	if item == nil {
		return errorsx.InvalidParam("接入站点不存在")
	}
	aiAgentID, err := s.validateAIAgentID(req.AIAgentID)
	if err != nil {
		return err
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return errorsx.InvalidParam("站点名称不能为空")
	}
	now := time.Now()
	return s.Updates(req.ID, map[string]any{
		"ai_agent_id":      aiAgentID,
		"name":             name,
		"remark":           strings.TrimSpace(req.Remark),
		"update_user_id":   operator.UserID,
		"update_user_name": operator.Username,
		"updated_at":       now,
	})
}

func (s *widgetSiteService) UpdateStatus(id int64, status int, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	item := s.Get(id)
	if item == nil {
		return errorsx.InvalidParam("接入站点不存在")
	}
	if status != int(enums.StatusOk) && status != int(enums.StatusDisabled) {
		return errorsx.InvalidParam("状态值不合法")
	}
	return s.Updates(id, map[string]any{
		"status":           status,
		"update_user_id":   operator.UserID,
		"update_user_name": operator.Username,
		"updated_at":       time.Now(),
	})
}

func (s *widgetSiteService) DeleteSite(id int64) error {
	item := s.Get(id)
	if item == nil {
		return errorsx.InvalidParam("接入站点不存在")
	}
	s.Delete(id)
	return nil
}

func (s *widgetSiteService) validateAIAgentID(aiAgentID int64) (int64, error) {
	if aiAgentID <= 0 {
		return 0, errorsx.InvalidParam("请选择 AI Agent")
	}
	aiAgent := AIAgentService.Get(aiAgentID)
	if aiAgent == nil {
		return 0, errorsx.InvalidParam("AI Agent 不存在")
	}
	if aiAgent.Status != enums.StatusOk {
		return 0, errorsx.InvalidParam("AI Agent 未启用")
	}
	return aiAgentID, nil
}
