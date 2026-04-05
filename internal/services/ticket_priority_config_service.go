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

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
)

var TicketPriorityConfigService = newTicketPriorityConfigService()

func newTicketPriorityConfigService() *ticketPriorityConfigService { return &ticketPriorityConfigService{} }

type ticketPriorityConfigService struct{}

func (s *ticketPriorityConfigService) Get(id int64) *models.TicketPriorityConfig {
	return repositories.TicketPriorityConfigRepository.Get(sqls.DB(), id)
}

func (s *ticketPriorityConfigService) Take(where ...interface{}) *models.TicketPriorityConfig {
	return repositories.TicketPriorityConfigRepository.Take(sqls.DB(), where...)
}

func (s *ticketPriorityConfigService) Find(cnd *sqls.Cnd) []models.TicketPriorityConfig {
	return repositories.TicketPriorityConfigRepository.Find(sqls.DB(), cnd)
}

func (s *ticketPriorityConfigService) FindOne(cnd *sqls.Cnd) *models.TicketPriorityConfig {
	return repositories.TicketPriorityConfigRepository.FindOne(sqls.DB(), cnd)
}

func (s *ticketPriorityConfigService) FindPageByParams(queryParams *params.QueryParams) (list []models.TicketPriorityConfig, paging *sqls.Paging) {
	return repositories.TicketPriorityConfigRepository.FindPageByParams(sqls.DB(), queryParams)
}

func (s *ticketPriorityConfigService) FindPageByCnd(cnd *sqls.Cnd) (list []models.TicketPriorityConfig, paging *sqls.Paging) {
	return repositories.TicketPriorityConfigRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *ticketPriorityConfigService) Create(t *models.TicketPriorityConfig) error {
	return repositories.TicketPriorityConfigRepository.Create(sqls.DB(), t)
}

func (s *ticketPriorityConfigService) Updates(id int64, columns map[string]any) error {
	return repositories.TicketPriorityConfigRepository.Updates(sqls.DB(), id, columns)
}

func (s *ticketPriorityConfigService) GetDefaultActive() *models.TicketPriorityConfig {
	return s.FindOne(sqls.NewCnd().Eq("status", enums.StatusOk).Asc("sort_no").Asc("id"))
}

func (s *ticketPriorityConfigService) CreateTicketPriorityConfig(req request.CreateTicketPriorityConfigRequest, operator *dto.AuthPrincipal) (*models.TicketPriorityConfig, error) {
	if operator == nil {
		return nil, errorsx.Unauthorized("未登录或登录已过期")
	}
	item, err := s.buildPriorityConfigModel(0, req.Name, req.SortNo, req.FirstResponseMinutes, req.ResolutionMinutes, int(req.Status), req.Remark)
	if err != nil {
		return nil, err
	}
	item.AuditFields = utils.BuildAuditFields(operator)
	if err := s.Create(item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *ticketPriorityConfigService) UpdateTicketPriorityConfig(req request.UpdateTicketPriorityConfigRequest, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	current := s.Get(req.ID)
	if current == nil || current.Status == enums.StatusDeleted {
		return errorsx.InvalidParam("工单优先级配置不存在")
	}
	item, err := s.buildPriorityConfigModel(req.ID, req.Name, req.SortNo, req.FirstResponseMinutes, req.ResolutionMinutes, int(req.Status), req.Remark)
	if err != nil {
		return err
	}
	return s.Updates(req.ID, map[string]any{
		"name":                   item.Name,
		"sort_no":                item.SortNo,
		"first_response_minutes": item.FirstResponseMinutes,
		"resolution_minutes":     item.ResolutionMinutes,
		"status":                 item.Status,
		"remark":                 item.Remark,
		"update_user_id":         operator.UserID,
		"update_user_name":       operator.Username,
		"updated_at":             time.Now(),
	})
}

func (s *ticketPriorityConfigService) DeleteTicketPriorityConfig(id int64, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	current := s.Get(id)
	if current == nil || current.Status == enums.StatusDeleted {
		return errorsx.InvalidParam("工单优先级配置不存在")
	}
	if TicketService.Take("priority = ?", id) != nil {
		return errorsx.Forbidden("该优先级仍有关联工单，无法删除")
	}
	return s.Updates(id, map[string]any{
		"status":           enums.StatusDeleted,
		"update_user_id":   operator.UserID,
		"update_user_name": operator.Username,
		"updated_at":       time.Now(),
	})
}

func (s *ticketPriorityConfigService) buildPriorityConfigModel(id int64, name string, sortNo, firstResponseMinutes, resolutionMinutes, status int, remark string) (*models.TicketPriorityConfig, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errorsx.InvalidParam("工单优先级名称不能为空")
	}
	if firstResponseMinutes <= 0 || resolutionMinutes <= 0 {
		return nil, errorsx.InvalidParam("SLA 时长必须大于 0")
	}
	if !enums.IsValidStatus(status) || status == int(enums.StatusDeleted) {
		return nil, errorsx.InvalidParam("工单优先级状态不合法")
	}
	if exists := s.Take("name = ? AND status <> ? AND id <> ?", name, enums.StatusDeleted, id); exists != nil {
		return nil, errorsx.InvalidParam("工单优先级名称已存在")
	}
	return &models.TicketPriorityConfig{
		Name:                 name,
		SortNo:               sortNo,
		FirstResponseMinutes: firstResponseMinutes,
		ResolutionMinutes:    resolutionMinutes,
		Status:               enums.Status(status),
		Remark:               strings.TrimSpace(remark),
	}, nil
}
