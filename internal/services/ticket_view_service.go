package services

import (
	"encoding/json"
	"strings"
	"time"

	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/errorsx"
	"cs-agent/internal/pkg/utils"
	"cs-agent/internal/repositories"

	"github.com/mlogclub/simple/sqls"
)

var TicketViewService = newTicketViewService()

func newTicketViewService() *ticketViewService {
	return &ticketViewService{}
}

type ticketViewService struct {
}

func (s *ticketViewService) Get(id int64) *models.TicketView {
	return repositories.TicketViewRepository.Get(sqls.DB(), id)
}

func (s *ticketViewService) Find(cnd *sqls.Cnd) []models.TicketView {
	return repositories.TicketViewRepository.Find(sqls.DB(), cnd)
}

func (s *ticketViewService) ListByUser(userID int64) []models.TicketView {
	if userID <= 0 {
		return nil
	}
	return s.Find(sqls.NewCnd().Eq("user_id", userID).Asc("sort_no").Desc("id"))
}

func (s *ticketViewService) Save(req request.SaveTicketViewRequest, operator *dto.AuthPrincipal) (*models.TicketView, error) {
	if operator == nil {
		return nil, errorsx.Unauthorized("未登录或登录已过期")
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errorsx.InvalidParam("视图名称不能为空")
	}
	filtersJSON, err := json.Marshal(req.Filters)
	if err != nil {
		return nil, errorsx.InvalidParam("视图筛选条件格式不正确")
	}
	now := time.Now()
	if req.ID > 0 {
		item := repositories.TicketViewRepository.Get(sqls.DB(), req.ID)
		if item == nil || item.UserID != operator.UserID {
			return nil, errorsx.InvalidParam("视图不存在")
		}
		if err := repositories.TicketViewRepository.Updates(sqls.DB(), req.ID, map[string]interface{}{
			"name":             name,
			"filters_json":     string(filtersJSON),
			"update_user_id":   operator.UserID,
			"update_user_name": operator.Username,
			"updated_at":       now,
		}); err != nil {
			return nil, err
		}
		return repositories.TicketViewRepository.Get(sqls.DB(), req.ID), nil
	}
	item := &models.TicketView{
		UserID:      operator.UserID,
		Name:        name,
		FiltersJSON: string(filtersJSON),
		AuditFields: utils.BuildAuditFields(operator),
	}
	if err := repositories.TicketViewRepository.Create(sqls.DB(), item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *ticketViewService) Delete(id int64, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	item := repositories.TicketViewRepository.Get(sqls.DB(), id)
	if item == nil || item.UserID != operator.UserID {
		return errorsx.InvalidParam("视图不存在")
	}
	repositories.TicketViewRepository.Delete(sqls.DB(), id)
	return nil
}
