package request

import "cs-agent/internal/pkg/enums"

type QuickReplyListRequest struct {
	GroupName string `json:"groupName"`
	Keyword   string `json:"keyword"`
	Status    int    `json:"status"`
}

type CreateQuickReplyRequest struct {
	GroupName string       `json:"groupName"`
	Title     string       `json:"title"`
	Content   string       `json:"content"`
	Status    enums.Status `json:"status"`
	SortNo    int          `json:"sortNo"`
}

type UpdateQuickReplyRequest struct {
	ID int64 `json:"id"`
	CreateQuickReplyRequest
}

type DeleteQuickReplyRequest struct {
	ID int64 `json:"id"`
}
