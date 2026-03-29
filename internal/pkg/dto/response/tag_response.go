package response

import "cs-agent/internal/pkg/enums"

type TagResponse struct {
	ID        int64        `json:"id"`
	ParentID  int64        `json:"parentId"`
	Name      string       `json:"name"`
	Remark    string       `json:"remark"`
	SortNo    int          `json:"sortNo"`
	Status    enums.Status `json:"status"`
	CreatedAt string       `json:"createdAt"`
	UpdatedAt string       `json:"updatedAt"`
}

type TagTreeResponse struct {
	ID        int64              `json:"id"`
	ParentID  int64              `json:"parentId"`
	Name      string             `json:"name"`
	Remark    string             `json:"remark"`
	SortNo    int                `json:"sortNo"`
	Status    enums.Status       `json:"status"`
	CreatedAt string             `json:"createdAt"`
	UpdatedAt string             `json:"updatedAt"`
	Children  []*TagTreeResponse `json:"children"`
}
