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

var TicketCategoryService = newTicketCategoryService()

func newTicketCategoryService() *ticketCategoryService { return &ticketCategoryService{} }

type ticketCategoryService struct{}

func (s *ticketCategoryService) Get(id int64) *models.TicketCategory {
	return repositories.TicketCategoryRepository.Get(sqls.DB(), id)
}
func (s *ticketCategoryService) Take(where ...interface{}) *models.TicketCategory {
	return repositories.TicketCategoryRepository.Take(sqls.DB(), where...)
}
func (s *ticketCategoryService) Find(cnd *sqls.Cnd) []models.TicketCategory {
	return repositories.TicketCategoryRepository.Find(sqls.DB(), cnd)
}
func (s *ticketCategoryService) FindOne(cnd *sqls.Cnd) *models.TicketCategory {
	return repositories.TicketCategoryRepository.FindOne(sqls.DB(), cnd)
}
func (s *ticketCategoryService) FindPageByParams(params *params.QueryParams) (list []models.TicketCategory, paging *sqls.Paging) {
	return repositories.TicketCategoryRepository.FindPageByParams(sqls.DB(), params)
}
func (s *ticketCategoryService) FindPageByCnd(cnd *sqls.Cnd) (list []models.TicketCategory, paging *sqls.Paging) {
	return repositories.TicketCategoryRepository.FindPageByCnd(sqls.DB(), cnd)
}
func (s *ticketCategoryService) Count(cnd *sqls.Cnd) int64 {
	return repositories.TicketCategoryRepository.Count(sqls.DB(), cnd)
}
func (s *ticketCategoryService) Create(t *models.TicketCategory) error {
	return repositories.TicketCategoryRepository.Create(sqls.DB(), t)
}
func (s *ticketCategoryService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.TicketCategoryRepository.Updates(sqls.DB(), id, columns)
}

func (s *ticketCategoryService) CreateTicketCategory(req request.CreateTicketCategoryRequest, operator *dto.AuthPrincipal) (*models.TicketCategory, error) {
	if operator == nil {
		return nil, errorsx.Unauthorized("未登录或登录已过期")
	}
	item, err := s.buildCategoryModel(0, req.Name, req.Code, req.ParentID, int(req.Status), req.SortNo, req.Remark)
	if err != nil {
		return nil, err
	}
	item.AuditFields = utils.BuildAuditFields(operator)
	if err := s.Create(item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *ticketCategoryService) UpdateTicketCategory(req request.UpdateTicketCategoryRequest, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	current := s.Get(req.ID)
	if current == nil || current.Status == enums.StatusDeleted {
		return errorsx.InvalidParam("工单分类不存在")
	}
	item, err := s.buildCategoryModel(req.ID, req.Name, req.Code, req.ParentID, int(req.Status), req.SortNo, req.Remark)
	if err != nil {
		return err
	}
	return s.Updates(req.ID, map[string]any{
		"name":             item.Name,
		"code":             item.Code,
		"parent_id":        item.ParentID,
		"status":           item.Status,
		"sort_no":          item.SortNo,
		"remark":           item.Remark,
		"update_user_id":   operator.UserID,
		"update_user_name": operator.Username,
		"updated_at":       time.Now(),
	})
}

func (s *ticketCategoryService) DeleteTicketCategory(id int64, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	current := s.Get(id)
	if current == nil || current.Status == enums.StatusDeleted {
		return errorsx.InvalidParam("工单分类不存在")
	}
	if s.Take("parent_id = ? AND status <> ?", id, enums.StatusDeleted) != nil {
		return errorsx.Forbidden("该分类下仍有子分类，无法删除")
	}
	if TicketService.Take("category_id = ?", id) != nil {
		return errorsx.Forbidden("该分类下仍有关联工单，无法删除")
	}
	return s.Updates(id, map[string]any{
		"status":           enums.StatusDeleted,
		"update_user_id":   operator.UserID,
		"update_user_name": operator.Username,
		"updated_at":       time.Now(),
	})
}

func (s *ticketCategoryService) buildCategoryModel(id int64, name, code string, parentID int64, status, sortNo int, remark string) (*models.TicketCategory, error) {
	name = strings.TrimSpace(name)
	code = strings.TrimSpace(code)
	if name == "" {
		return nil, errorsx.InvalidParam("工单分类名称不能为空")
	}
	if code == "" {
		return nil, errorsx.InvalidParam("工单分类编码不能为空")
	}
	if !enums.IsValidStatus(status) || status == int(enums.StatusDeleted) {
		return nil, errorsx.InvalidParam("工单分类状态不合法")
	}
	if parentID > 0 {
		parent := s.Get(parentID)
		if parent == nil || parent.Status == enums.StatusDeleted {
			return nil, errorsx.InvalidParam("父级分类不存在")
		}
		if parent.ID == id {
			return nil, errorsx.InvalidParam("父级分类不能为自身")
		}
	}
	if exists := s.Take("name = ? AND status <> ? AND id <> ?", name, enums.StatusDeleted, id); exists != nil {
		return nil, errorsx.InvalidParam("工单分类名称已存在")
	}
	if exists := s.Take("code = ? AND status <> ? AND id <> ?", code, enums.StatusDeleted, id); exists != nil {
		return nil, errorsx.InvalidParam("工单分类编码已存在")
	}
	return &models.TicketCategory{
		Name:     name,
		Code:     code,
		ParentID: parentID,
		SortNo:   sortNo,
		Status:   enums.Status(status),
		Remark:   strings.TrimSpace(remark),
	}, nil
}
