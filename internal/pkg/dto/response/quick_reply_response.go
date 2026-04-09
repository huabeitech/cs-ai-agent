package response

import "cs-agent/internal/pkg/enums"

type QuickReplyResponse struct {
	ID        int64        `json:"id"`
	GroupName string       `json:"groupName"`
	Title     string       `json:"title"`
	Content   string       `json:"content"`
	Status    enums.Status `json:"status"`
	SortNo    int          `json:"sortNo"`
	CreatedBy int64        `json:"createdBy"`
}
