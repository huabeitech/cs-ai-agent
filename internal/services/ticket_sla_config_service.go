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

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
)

var TicketSLAConfigService = newTicketSLAConfigService()

func newTicketSLAConfigService() *ticketSLAConfigService { return &ticketSLAConfigService{} }

type ticketSLAConfigService struct{}

func (s *ticketSLAConfigService) Get(id int64) *models.TicketSLAConfig {
	return repositories.TicketSLAConfigRepository.Get(sqls.DB(), id)
}
func (s *ticketSLAConfigService) Take(where ...interface{}) *models.TicketSLAConfig {
	return repositories.TicketSLAConfigRepository.Take(sqls.DB(), where...)
}
func (s *ticketSLAConfigService) Find(cnd *sqls.Cnd) []models.TicketSLAConfig {
	return repositories.TicketSLAConfigRepository.Find(sqls.DB(), cnd)
}
func (s *ticketSLAConfigService) FindOne(cnd *sqls.Cnd) *models.TicketSLAConfig {
	return repositories.TicketSLAConfigRepository.FindOne(sqls.DB(), cnd)
}
func (s *ticketSLAConfigService) FindPageByParams(params *params.QueryParams) (list []models.TicketSLAConfig, paging *sqls.Paging) {
	return repositories.TicketSLAConfigRepository.FindPageByParams(sqls.DB(), params)
}
func (s *ticketSLAConfigService) FindPageByCnd(cnd *sqls.Cnd) (list []models.TicketSLAConfig, paging *sqls.Paging) {
	return repositories.TicketSLAConfigRepository.FindPageByCnd(sqls.DB(), cnd)
}
func (s *ticketSLAConfigService) Count(cnd *sqls.Cnd) int64 {
	return repositories.TicketSLAConfigRepository.Count(sqls.DB(), cnd)
}
func (s *ticketSLAConfigService) Create(t *models.TicketSLAConfig) error {
	return repositories.TicketSLAConfigRepository.Create(sqls.DB(), t)
}
func (s *ticketSLAConfigService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.TicketSLAConfigRepository.Updates(sqls.DB(), id, columns)
}

func (s *ticketSLAConfigService) GetActiveByPriority(priority enums.TicketPriority) *models.TicketSLAConfig {
	return s.FindOne(sqls.NewCnd().Eq("priority", priority).Eq("status", enums.StatusOk))
}

func (s *ticketSLAConfigService) CreateTicketSLAConfig(req request.CreateTicketSLAConfigRequest, operator *dto.AuthPrincipal) (*models.TicketSLAConfig, error) {
	if operator == nil {
		return nil, errorsx.Unauthorized("未登录或登录已过期")
	}
	item, err := s.buildSLAConfigModel(0, req.Name, req.Priority, req.FirstResponseMinutes, req.ResolutionMinutes, int(req.Status), req.Remark)
	if err != nil {
		return nil, err
	}
	item.AuditFields = utils.BuildAuditFields(operator)
	if err := s.Create(item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *ticketSLAConfigService) UpdateTicketSLAConfig(req request.UpdateTicketSLAConfigRequest, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	current := s.Get(req.ID)
	if current == nil || current.Status == enums.StatusDeleted {
		return errorsx.InvalidParam("工单SLA配置不存在")
	}
	item, err := s.buildSLAConfigModel(req.ID, req.Name, req.Priority, req.FirstResponseMinutes, req.ResolutionMinutes, int(req.Status), req.Remark)
	if err != nil {
		return err
	}
	return s.Updates(req.ID, map[string]any{
		"name":                   item.Name,
		"priority":               item.Priority,
		"first_response_minutes": item.FirstResponseMinutes,
		"resolution_minutes":     item.ResolutionMinutes,
		"status":                 item.Status,
		"remark":                 item.Remark,
		"update_user_id":         operator.UserID,
		"update_user_name":       operator.Username,
		"updated_at":             time.Now(),
	})
}

func (s *ticketSLAConfigService) DeleteTicketSLAConfig(id int64, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	current := s.Get(id)
	if current == nil || current.Status == enums.StatusDeleted {
		return errorsx.InvalidParam("工单SLA配置不存在")
	}
	return s.Updates(id, map[string]any{
		"status":           enums.StatusDeleted,
		"update_user_id":   operator.UserID,
		"update_user_name": operator.Username,
		"updated_at":       time.Now(),
	})
}

func (s *ticketSLAConfigService) buildSLAConfigModel(id int64, name string, priority enums.TicketPriority, firstResponseMinutes, resolutionMinutes, status int, remark string) (*models.TicketSLAConfig, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errorsx.InvalidParam("工单SLA配置名称不能为空")
	}
	if !enums.IsValidTicketPriority(int(priority)) {
		return nil, errorsx.InvalidParam("工单优先级不合法")
	}
	if firstResponseMinutes <= 0 || resolutionMinutes <= 0 {
		return nil, errorsx.InvalidParam("SLA 时长必须大于 0")
	}
	if !enums.IsValidStatus(status) || status == int(enums.StatusDeleted) {
		return nil, errorsx.InvalidParam("工单SLA配置状态不合法")
	}
	if exists := s.Take("name = ? AND status <> ? AND id <> ?", name, enums.StatusDeleted, id); exists != nil {
		return nil, errorsx.InvalidParam("工单SLA配置名称已存在")
	}
	if exists := s.Take("priority = ? AND status <> ? AND id <> ?", priority, enums.StatusDeleted, id); exists != nil {
		return nil, errorsx.InvalidParam("该优先级已存在 SLA 配置")
	}
	return &models.TicketSLAConfig{
		Name:                 name,
		Priority:             priority,
		FirstResponseMinutes: firstResponseMinutes,
		ResolutionMinutes:    resolutionMinutes,
		Status:               enums.Status(status),
		Remark:               strings.TrimSpace(remark),
	}, nil
}
