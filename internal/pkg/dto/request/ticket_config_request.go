package request

import "cs-agent/internal/pkg/enums"

type CreateTicketCategoryRequest struct {
	Name     string       `json:"name"`
	Code     string       `json:"code"`
	ParentID int64        `json:"parentId"`
	SortNo   int          `json:"sortNo"`
	Status   enums.Status `json:"status"`
	Remark   string       `json:"remark"`
}

type UpdateTicketCategoryRequest struct {
	ID int64 `json:"id"`
	CreateTicketCategoryRequest
}

type DeleteTicketCategoryRequest struct {
	ID int64 `json:"id"`
}

type CreateTicketResolutionCodeRequest struct {
	Name   string       `json:"name"`
	Code   string       `json:"code"`
	SortNo int          `json:"sortNo"`
	Status enums.Status `json:"status"`
	Remark string       `json:"remark"`
}

type UpdateTicketResolutionCodeRequest struct {
	ID int64 `json:"id"`
	CreateTicketResolutionCodeRequest
}

type DeleteTicketResolutionCodeRequest struct {
	ID int64 `json:"id"`
}

type CreateTicketSLAConfigRequest struct {
	Name                 string               `json:"name"`
	Priority             enums.TicketPriority `json:"priority"`
	FirstResponseMinutes int                  `json:"firstResponseMinutes"`
	ResolutionMinutes    int                  `json:"resolutionMinutes"`
	Status               enums.Status         `json:"status"`
	Remark               string               `json:"remark"`
}

type UpdateTicketSLAConfigRequest struct {
	ID int64 `json:"id"`
	CreateTicketSLAConfigRequest
}

type DeleteTicketSLAConfigRequest struct {
	ID int64 `json:"id"`
}
