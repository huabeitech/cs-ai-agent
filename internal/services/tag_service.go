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

var TagService = newTagService()

func newTagService() *tagService {
	return &tagService{}
}

type tagService struct {
}

func (s *tagService) Get(id int64) *models.Tag {
	return repositories.TagRepository.Get(sqls.DB(), id)
}

func (s *tagService) Take(where ...interface{}) *models.Tag {
	return repositories.TagRepository.Take(sqls.DB(), where...)
}

func (s *tagService) Find(cnd *sqls.Cnd) []models.Tag {
	return repositories.TagRepository.Find(sqls.DB(), cnd)
}

func (s *tagService) FindOne(cnd *sqls.Cnd) *models.Tag {
	return repositories.TagRepository.FindOne(sqls.DB(), cnd)
}

func (s *tagService) FindPageByParams(params *params.QueryParams) (list []models.Tag, paging *sqls.Paging) {
	return repositories.TagRepository.FindPageByParams(sqls.DB(), params)
}

func (s *tagService) FindPageByCnd(cnd *sqls.Cnd) (list []models.Tag, paging *sqls.Paging) {
	return repositories.TagRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *tagService) Count(cnd *sqls.Cnd) int64 {
	return repositories.TagRepository.Count(sqls.DB(), cnd)
}

func (s *tagService) Create(t *models.Tag) error {
	return repositories.TagRepository.Create(sqls.DB(), t)
}

func (s *tagService) Update(t *models.Tag) error {
	return repositories.TagRepository.Update(sqls.DB(), t)
}

func (s *tagService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.TagRepository.Updates(sqls.DB(), id, columns)
}

func (s *tagService) UpdateColumn(id int64, name string, value interface{}) error {
	return repositories.TagRepository.UpdateColumn(sqls.DB(), id, name, value)
}

func (s *tagService) Delete(id int64) {
	repositories.TagRepository.Delete(sqls.DB(), id)
}

func (s *tagService) GetChildren(parentID int64) []models.Tag {
	return s.Find(sqls.NewCnd().Eq("parent_id", parentID).Asc("sort_no").Asc("id"))
}

func (s *tagService) HasChildren(parentID int64) bool {
	return s.Count(sqls.NewCnd().Eq("parent_id", parentID)) > 0
}

func (s *tagService) FindByNameAndParentID(name string, parentID int64) *models.Tag {
	return s.FindOne(sqls.NewCnd().Eq("name", name).Eq("parent_id", parentID))
}

func (s *tagService) CreateTag(req request.CreateTagRequest, operator *dto.AuthPrincipal) (*models.Tag, error) {
	if operator == nil {
		return nil, errorsx.Unauthorized("未登录或登录已过期")
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errorsx.InvalidParam("标签名称不能为空")
	}

	if req.ParentID > 0 {
		parent := s.Get(req.ParentID)
		if parent == nil {
			return nil, errorsx.InvalidParam("父标签不存在")
		}
	}

	existing := s.FindByNameAndParentID(name, req.ParentID)
	if existing != nil {
		return nil, errorsx.InvalidParam("同级下已存在相同名称的标签")
	}

	item := &models.Tag{
		ParentID:    req.ParentID,
		Name:        name,
		Remark:      strings.TrimSpace(req.Remark),
		Status:      enums.StatusOk,
		AuditFields: utils.BuildAuditFields(operator),
	}

	item.SortNo = s.NextSortNo(req.ParentID)
	if err := s.Create(item); err != nil {
		return nil, err
	}

	return item, nil
}

func (s *tagService) NextSortNo(parentID int64) int {
	if temp := s.FindOne(sqls.NewCnd().Eq("parent_id", parentID).Desc("sort_no").Desc("id")); temp != nil {
		return temp.SortNo + 1
	}
	return 1
}

func (s *tagService) UpdateTag(req request.UpdateTagRequest, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}

	item := s.Get(req.ID)
	if item == nil {
		return errorsx.InvalidParam("标签不存在")
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return errorsx.InvalidParam("标签名称不能为空")
	}

	if req.ParentID > 0 {
		if req.ParentID == req.ID {
			return errorsx.InvalidParam("不能将标签设为自己的子标签")
		}
		parent := s.Get(req.ParentID)
		if parent == nil {
			return errorsx.InvalidParam("父标签不存在")
		}
	}

	existing := s.FindByNameAndParentID(name, req.ParentID)
	if existing != nil && existing.ID != req.ID {
		return errorsx.InvalidParam("同级下已存在相同名称的标签")
	}

	return s.Updates(req.ID, map[string]any{
		"parent_id":        req.ParentID,
		"name":             name,
		"remark":           strings.TrimSpace(req.Remark),
		"update_user_id":   operator.UserID,
		"update_user_name": operator.Username,
		"updated_at":       time.Now(),
	})
}

func (s *tagService) UpdateSort(ids []int64) error {
	return sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		for i, id := range ids {
			if err := repositories.TagRepository.UpdateColumn(ctx.Tx, id, "sort_no", i+1); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *tagService) DeleteTag(id int64) error {
	item := s.Get(id)
	if item == nil {
		return errorsx.InvalidParam("标签不存在")
	}

	if s.HasChildren(id) {
		return errorsx.InvalidParam("该标签下存在子标签，无法删除")
	}

	s.Delete(id)
	return nil
}

func (s *tagService) FindAll() []models.Tag {
	return s.Find(sqls.NewCnd().Asc("sort_no").Asc("id"))
}

func (s *tagService) GetSelfAndDescendantIDs(tagID int64) []int64 {
	if tagID <= 0 {
		return nil
	}

	allTags := s.FindAll()
	if len(allTags) == 0 {
		return nil
	}

	exists := false
	childrenMap := make(map[int64][]int64, len(allTags))
	for _, item := range allTags {
		if item.ID == tagID {
			exists = true
		}
		childrenMap[item.ParentID] = append(childrenMap[item.ParentID], item.ID)
	}
	if !exists {
		return nil
	}

	result := make([]int64, 0, 8)
	visited := make(map[int64]bool, len(allTags))
	var walk func(id int64)
	walk = func(id int64) {
		if visited[id] {
			return
		}
		visited[id] = true
		result = append(result, id)
		for _, childID := range childrenMap[id] {
			walk(childID)
		}
	}
	walk(tagID)

	return result
}

func (s *tagService) UpdateStatus(id int64, status int, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}

	item := s.Get(id)
	if item == nil {
		return errorsx.InvalidParam("标签不存在")
	}

	if status != int(enums.StatusOk) && status != int(enums.StatusDisabled) {
		return errorsx.InvalidParam("状态值不合法")
	}

	now := time.Now()
	return s.Updates(id, map[string]any{
		"status":           status,
		"update_user_id":   operator.UserID,
		"update_user_name": operator.Username,
		"updated_at":       now,
	})
}
