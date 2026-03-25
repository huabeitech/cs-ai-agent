package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/utils"
	"cs-agent/internal/repositories"
	"errors"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
)

var TicketCategoryService = newTicketCategoryService()

func newTicketCategoryService() *ticketCategoryService {
	return &ticketCategoryService{}
}

type ticketCategoryService struct {
}

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

func (s *ticketCategoryService) Update(t *models.TicketCategory) error {
	return repositories.TicketCategoryRepository.Update(sqls.DB(), t)
}

func (s *ticketCategoryService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.TicketCategoryRepository.Updates(sqls.DB(), id, columns)
}

func (s *ticketCategoryService) UpdateColumn(id int64, name string, value interface{}) error {
	return repositories.TicketCategoryRepository.UpdateColumn(sqls.DB(), id, name, value)
}

func (s *ticketCategoryService) Delete(id int64) {
	repositories.TicketCategoryRepository.Delete(sqls.DB(), id)
}

func (s *ticketCategoryService) FindTree() []models.TicketCategory {
	categories := repositories.TicketCategoryRepository.Find(sqls.DB(), sqls.NewCnd().Where("status = ?", 1).Asc("sort_no").Asc("id"))
	return s.buildTree(categories, 0)
}

func (s *ticketCategoryService) buildTree(categories []models.TicketCategory, parentID int64) []models.TicketCategory {
	result := make([]models.TicketCategory, 0)
	for i := range categories {
		if categories[i].ParentID == parentID {
			children := s.buildTree(categories, categories[i].ID)
			if len(children) > 0 {
				categories[i].Remark = ""
			}
			result = append(result, categories[i])
			result = append(result, children...)
		}
	}
	return result
}

func (s *ticketCategoryService) GetChildren(parentID int64) []models.TicketCategory {
	return repositories.TicketCategoryRepository.Find(sqls.DB(), sqls.NewCnd().Where("parent_id = ?", parentID).Where("status = ?", 1).Asc("sort_no").Asc("id"))
}

func (s *ticketCategoryService) CreateCategory(req request.CreateTicketCategoryRequest, principal *dto.AuthPrincipal) (*models.TicketCategory, error) {
	category := &models.TicketCategory{
		ParentID:    req.ParentID,
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		SortNo:      s.NextSortNo(req.ParentID),
		Status:      req.Status,
		Remark:      req.Remark,
	}
	category.AuditFields = utils.BuildAuditFields(principal)
	err := s.Create(category)
	return category, err
}

func (s *ticketCategoryService) NextSortNo(parentID int64) int {
	if item := s.FindOne(sqls.NewCnd().Where("parent_id = ?", parentID).Desc("sort_no").Asc("id")); item != nil {
		return item.SortNo + 1
	}
	return 1
}

func (s *ticketCategoryService) UpdateCategory(req request.UpdateTicketCategoryRequest, principal *dto.AuthPrincipal) error {
	updates := map[string]interface{}{
		"parent_id":   req.ParentID,
		"name":        req.Name,
		"code":        req.Code,
		"description": req.Description,
		"status":      req.Status,
		"remark":      req.Remark,
	}
	if principal != nil {
		updates["update_user_id"] = principal.UserID
		updates["update_user_name"] = principal.Nickname
	}
	return s.Updates(req.ID, updates)
}

func (s *ticketCategoryService) DeleteCategory(id int64) error {
	children := s.GetChildren(id)
	if len(children) > 0 {
		return errors.New("请先删除子分类")
	}
	s.Delete(id)
	return nil
}

func (s *ticketCategoryService) UpdateSort(ids []int64) error {
	for i, id := range ids {
		if err := s.UpdateColumn(id, "sort_no", len(ids)-i); err != nil {
			return err
		}
	}
	return nil
}
