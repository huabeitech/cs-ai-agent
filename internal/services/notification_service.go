package services

import (
	"strings"
	"time"

	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/errorsx"
	"cs-agent/internal/pkg/utils"
	"cs-agent/internal/repositories"

	"github.com/mlogclub/simple/sqls"
)

var NotificationService = newNotificationService()

func newNotificationService() *notificationService {
	return &notificationService{}
}

type notificationService struct {
}

func (s *notificationService) Create(req request.CreateNotificationRequest) (*models.Notification, error) {
	if req.RecipientUserID <= 0 {
		return nil, errorsx.InvalidParam("接收人不能为空")
	}
	now := time.Now()
	item := &models.Notification{
		RecipientUserID:  req.RecipientUserID,
		Title:            strings.TrimSpace(req.Title),
		Content:          strings.TrimSpace(req.Content),
		NotificationType: strings.TrimSpace(req.NotificationType),
		BizType:          strings.TrimSpace(req.BizType),
		BizID:            req.BizID,
		ActionURL:        strings.TrimSpace(req.ActionURL),
		Status:           enums.StatusOk,
		CreatedAt:        now,
	}
	if err := repositories.NotificationRepository.Create(sqls.DB(), item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *notificationService) CreateAndPush(req request.CreateNotificationRequest) (*models.Notification, error) {
	item, err := s.Create(req)
	if err != nil {
		return nil, err
	}
	WsService.PublishNotificationCreated(item.RecipientUserID, response.NotificationResponse{
		ID:               item.ID,
		RecipientUserID:  item.RecipientUserID,
		Title:            item.Title,
		Content:          item.Content,
		NotificationType: item.NotificationType,
		BizType:          item.BizType,
		BizID:            item.BizID,
		ActionURL:        item.ActionURL,
		ReadAt:           utils.FormatTimePtr(item.ReadAt),
		CreatedAt:        utils.FormatTime(item.CreatedAt),
	})
	return item, nil
}

func (s *notificationService) FindPageByCnd(cnd *sqls.Cnd) ([]models.Notification, *sqls.Paging) {
	return repositories.NotificationRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *notificationService) CountUnread(userID int64) int64 {
	if userID <= 0 {
		return 0
	}
	return repositories.NotificationRepository.Count(sqls.DB(), sqls.NewCnd().
		Eq("recipient_user_id", userID).
		Eq("status", enums.StatusOk).
		Where("read_at IS NULL"))
}

func (s *notificationService) MarkRead(id int64, userID int64) error {
	if id <= 0 {
		return errorsx.InvalidParam("通知不存在")
	}
	item := repositories.NotificationRepository.Get(sqls.DB(), id)
	if item == nil || item.RecipientUserID != userID {
		return errorsx.InvalidParam("通知不存在")
	}
	if item.ReadAt != nil {
		return nil
	}
	now := time.Now()
	return repositories.NotificationRepository.Updates(sqls.DB(), id, map[string]any{
		"read_at": now,
	})
}

func (s *notificationService) MarkAllRead(userID int64) error {
	if userID <= 0 {
		return errorsx.InvalidParam("接收人不能为空")
	}
	return repositories.NotificationRepository.MarkAllRead(sqls.DB(), userID, time.Now())
}
