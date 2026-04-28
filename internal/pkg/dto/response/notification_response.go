package response

type NotificationResponse struct {
	ID               int64  `json:"id"`
	RecipientUserID  int64  `json:"recipientUserId"`
	Title            string `json:"title"`
	Content          string `json:"content"`
	NotificationType string `json:"notificationType"`
	BizType          string `json:"bizType"`
	BizID            int64  `json:"bizId"`
	ActionURL        string `json:"actionUrl"`
	ReadAt           string `json:"readAt,omitempty"`
	CreatedAt        string `json:"createdAt,omitempty"`
}

type NotificationUnreadCountResponse struct {
	UnreadCount int64 `json:"unreadCount"`
}
