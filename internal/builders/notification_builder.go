package builders

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/pkg/utils"
)

func BuildNotification(item *models.Notification) *response.NotificationResponse {
	if item == nil {
		return nil
	}
	return &response.NotificationResponse{
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
	}
}

func BuildNotificationList(list []models.Notification) []response.NotificationResponse {
	if len(list) == 0 {
		return nil
	}
	results := make([]response.NotificationResponse, 0, len(list))
	for i := range list {
		if item := BuildNotification(&list[i]); item != nil {
			results = append(results, *item)
		}
	}
	return results
}
