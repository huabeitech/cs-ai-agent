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

var TicketResolutionCodeService = newTicketResolutionCodeService()

func newTicketResolutionCodeService() *ticketResolutionCodeService {
	return &ticketResolutionCodeService{}
}

type ticketResolutionCodeService struct{}

func (s *ticketResolutionCodeService) Get(id int64) *models.TicketResolutionCode {
	return repositories.TicketResolutionCodeRepository.Get(sqls.DB(), id)
}
func (s *ticketResolutionCodeService) Take(where ...interface{}) *models.TicketResolutionCode {
	return repositories.TicketResolutionCodeRepository.Take(sqls.DB(), where...)
}
func (s *ticketResolutionCodeService) Find(cnd *sqls.Cnd) []models.TicketResolutionCode {
	return repositories.TicketResolutionCodeRepository.Find(sqls.DB(), cnd)
}
func (s *ticketResolutionCodeService) FindOne(cnd *sqls.Cnd) *models.TicketResolutionCode {
	return repositories.TicketResolutionCodeRepository.FindOne(sqls.DB(), cnd)
}
func (s *ticketResolutionCodeService) FindPageByParams(params *params.QueryParams) (list []models.TicketResolutionCode, paging *sqls.Paging) {
	return repositories.TicketResolutionCodeRepository.FindPageByParams(sqls.DB(), params)
}
func (s *ticketResolutionCodeService) FindPageByCnd(cnd *sqls.Cnd) (list []models.TicketResolutionCode, paging *sqls.Paging) {
	return repositories.TicketResolutionCodeRepository.FindPageByCnd(sqls.DB(), cnd)
}
func (s *ticketResolutionCodeService) Count(cnd *sqls.Cnd) int64 {
	return repositories.TicketResolutionCodeRepository.Count(sqls.DB(), cnd)
}
func (s *ticketResolutionCodeService) Create(t *models.TicketResolutionCode) error {
	return repositories.TicketResolutionCodeRepository.Create(sqls.DB(), t)
}
func (s *ticketResolutionCodeService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.TicketResolutionCodeRepository.Updates(sqls.DB(), id, columns)
}

func (s *ticketResolutionCodeService) CreateTicketResolutionCode(req request.CreateTicketResolutionCodeRequest, operator *dto.AuthPrincipal) (*models.TicketResolutionCode, error) {
	if operator == nil {
		return nil, errorsx.Unauthorized("未登录或登录已过期")
	}
	item, err := s.buildResolutionCodeModel(0, req.Name, req.Code, int(req.Status), req.SortNo, req.Remark)
	if err != nil {
		return nil, err
	}
	item.AuditFields = utils.BuildAuditFields(operator)
	if err := s.Create(item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *ticketResolutionCodeService) UpdateTicketResolutionCode(req request.UpdateTicketResolutionCodeRequest, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	current := s.Get(req.ID)
	if current == nil || current.Status == enums.StatusDeleted {
		return errorsx.InvalidParam("工单解决码不存在")
	}
	item, err := s.buildResolutionCodeModel(req.ID, req.Name, req.Code, int(req.Status), req.SortNo, req.Remark)
	if err != nil {
		return err
	}
	return s.Updates(req.ID, map[string]any{
		"name":             item.Name,
		"code":             item.Code,
		"status":           item.Status,
		"sort_no":          item.SortNo,
		"remark":           item.Remark,
		"update_user_id":   operator.UserID,
		"update_user_name": operator.Username,
		"updated_at":       time.Now(),
	})
}

func (s *ticketResolutionCodeService) DeleteTicketResolutionCode(id int64, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	current := s.Get(id)
	if current == nil || current.Status == enums.StatusDeleted {
		return errorsx.InvalidParam("工单解决码不存在")
	}
	return s.Updates(id, map[string]any{
		"status":           enums.StatusDeleted,
		"update_user_id":   operator.UserID,
		"update_user_name": operator.Username,
		"updated_at":       time.Now(),
	})
}

func (s *ticketResolutionCodeService) buildResolutionCodeModel(id int64, name, code string, status, sortNo int, remark string) (*models.TicketResolutionCode, error) {
	name = strings.TrimSpace(name)
	code = strings.TrimSpace(code)
	if name == "" {
		return nil, errorsx.InvalidParam("工单解决码名称不能为空")
	}
	if code == "" {
		return nil, errorsx.InvalidParam("工单解决码编码不能为空")
	}
	if !enums.IsValidStatus(status) || status == int(enums.StatusDeleted) {
		return nil, errorsx.InvalidParam("工单解决码状态不合法")
	}
	if exists := s.Take("name = ? AND status <> ? AND id <> ?", name, enums.StatusDeleted, id); exists != nil {
		return nil, errorsx.InvalidParam("工单解决码名称已存在")
	}
	if exists := s.Take("code = ? AND status <> ? AND id <> ?", code, enums.StatusDeleted, id); exists != nil {
		return nil, errorsx.InvalidParam("工单解决码编码已存在")
	}
	return &models.TicketResolutionCode{
		Name:   name,
		Code:   code,
		SortNo: sortNo,
		Status: enums.Status(status),
		Remark: strings.TrimSpace(remark),
	}, nil
}
