package request

type CreateNotificationRequest struct {
	RecipientUserID  int64  `json:"recipientUserId"`
	Title            string `json:"title"`
	Content          string `json:"content"`
	NotificationType string `json:"notificationType"`
	BizType          string `json:"bizType"`
	BizID            int64  `json:"bizId"`
	ActionURL        string `json:"actionUrl"`
}

type MarkNotificationReadRequest struct {
	ID int64 `json:"id"`
}
