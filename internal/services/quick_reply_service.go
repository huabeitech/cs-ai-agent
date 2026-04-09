package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/errorsx"
	"cs-agent/internal/pkg/utils"
	"cs-agent/internal/repositories"
	"strings"
	"time"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
)

var QuickReplyService = newQuickReplyService()

func newQuickReplyService() *quickReplyService {
	return &quickReplyService{}
}

type quickReplyService struct {
}

func (s *quickReplyService) Get(id int64) *models.QuickReply {
	return repositories.QuickReplyRepository.Get(sqls.DB(), id)
}

func (s *quickReplyService) Take(where ...interface{}) *models.QuickReply {
	return repositories.QuickReplyRepository.Take(sqls.DB(), where...)
}

func (s *quickReplyService) Find(cnd *sqls.Cnd) []models.QuickReply {
	return repositories.QuickReplyRepository.Find(sqls.DB(), cnd)
}

func (s *quickReplyService) FindOne(cnd *sqls.Cnd) *models.QuickReply {
	return repositories.QuickReplyRepository.FindOne(sqls.DB(), cnd)
}

func (s *quickReplyService) FindPageByParams(params *params.QueryParams) (list []models.QuickReply, paging *sqls.Paging) {
	return repositories.QuickReplyRepository.FindPageByParams(sqls.DB(), params)
}

func (s *quickReplyService) FindPageByCnd(cnd *sqls.Cnd) (list []models.QuickReply, paging *sqls.Paging) {
	return repositories.QuickReplyRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *quickReplyService) Count(cnd *sqls.Cnd) int64 {
	return repositories.QuickReplyRepository.Count(sqls.DB(), cnd)
}

func (s *quickReplyService) Create(t *models.QuickReply) error {
	return repositories.QuickReplyRepository.Create(sqls.DB(), t)
}

func (s *quickReplyService) Update(t *models.QuickReply) error {
	return repositories.QuickReplyRepository.Update(sqls.DB(), t)
}

func (s *quickReplyService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.QuickReplyRepository.Updates(sqls.DB(), id, columns)
}

func (s *quickReplyService) UpdateColumn(id int64, name string, value interface{}) error {
	return repositories.QuickReplyRepository.UpdateColumn(sqls.DB(), id, name, value)
}

func (s *quickReplyService) Delete(id int64) {
	repositories.QuickReplyRepository.Delete(sqls.DB(), id)
}

func (s *quickReplyService) CreateQuickReply(req request.CreateQuickReplyRequest, operator *dto.AuthPrincipal) (*models.QuickReply, error) {
	if operator == nil {
		return nil, errorsx.Unauthorized("未登录或登录已过期")
	}
	title := strings.TrimSpace(req.Title)
	content := strings.TrimSpace(req.Content)
	if title == "" || content == "" {
		return nil, errorsx.InvalidParam("标题和内容不能为空")
	}
	item := &models.QuickReply{
		GroupName:   strings.TrimSpace(req.GroupName),
		Title:       title,
		Content:     content,
		Status:      req.Status,
		SortNo:      req.SortNo,
		AuditFields: utils.BuildAuditFields(operator),
	}
	if err := s.Create(item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *quickReplyService) UpdateQuickReply(req request.UpdateQuickReplyRequest, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	item := s.Get(req.ID)
	if item == nil {
		return errorsx.InvalidParam("快捷回复不存在")
	}
	return s.Updates(req.ID, map[string]any{
		"group_name":       strings.TrimSpace(req.GroupName),
		"title":            strings.TrimSpace(req.Title),
		"content":          strings.TrimSpace(req.Content),
		"status":           req.Status,
		"sort_no":          req.SortNo,
		"update_user_id":   operator.UserID,
		"update_user_name": operator.Username,
		"updated_at":       time.Now(),
	})
}

func (s *quickReplyService) DeleteQuickReply(id int64) error {
	item := s.Get(id)
	if item == nil {
		return errorsx.InvalidParam("快捷回复不存在")
	}
	s.Delete(id)
	return nil
}
