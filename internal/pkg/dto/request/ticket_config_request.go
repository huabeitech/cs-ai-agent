package request

import "cs-agent/internal/pkg/enums"

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

type CreateTicketPriorityConfigRequest struct {
	Name                 string       `json:"name"`
	FirstResponseMinutes int          `json:"firstResponseMinutes"`
	ResolutionMinutes    int          `json:"resolutionMinutes"`
	Status               enums.Status `json:"status"`
	Remark               string       `json:"remark"`
}

type UpdateTicketPriorityConfigRequest struct {
	ID int64 `json:"id"`
	CreateTicketPriorityConfigRequest
}

type DeleteTicketPriorityConfigRequest struct {
	ID int64 `json:"id"`
}
