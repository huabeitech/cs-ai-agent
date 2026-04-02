package response

import "cs-agent/internal/pkg/enums"

type TicketCategoryResponse struct {
	ID         int64        `json:"id"`
	Name       string       `json:"name"`
	Code       string       `json:"code"`
	ParentID   int64        `json:"parentId"`
	ParentName string       `json:"parentName,omitempty"`
	SortNo     int          `json:"sortNo"`
	Status     enums.Status `json:"status"`
	Remark     string       `json:"remark"`
}

type TicketResolutionCodeResponse struct {
	ID     int64        `json:"id"`
	Name   string       `json:"name"`
	Code   string       `json:"code"`
	SortNo int          `json:"sortNo"`
	Status enums.Status `json:"status"`
	Remark string       `json:"remark"`
}

type TicketSLAConfigResponse struct {
	ID                   int64                `json:"id"`
	Name                 string               `json:"name"`
	Priority             enums.TicketPriority `json:"priority"`
	FirstResponseMinutes int                  `json:"firstResponseMinutes"`
	ResolutionMinutes    int                  `json:"resolutionMinutes"`
	Status               enums.Status         `json:"status"`
	Remark               string               `json:"remark"`
}
