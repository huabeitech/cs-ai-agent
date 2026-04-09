package builders

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto/response"
)

func BuildTicketResolutionCode(item *models.TicketResolutionCode) *response.TicketResolutionCodeResponse {
	if item == nil {
		return nil
	}
	return &response.TicketResolutionCodeResponse{
		ID:     item.ID,
		Name:   item.Name,
		Code:   item.Code,
		SortNo: item.SortNo,
		Status: item.Status,
		Remark: item.Remark,
	}
}

func BuildTicketResolutionCodeList(list []models.TicketResolutionCode) []response.TicketResolutionCodeResponse {
	if len(list) == 0 {
		return make([]response.TicketResolutionCodeResponse, 0)
	}
	ret := make([]response.TicketResolutionCodeResponse, 0, len(list))
	for i := range list {
		if item := BuildTicketResolutionCode(&list[i]); item != nil {
			ret = append(ret, *item)
		}
	}
	return ret
}

func BuildTicketPriorityConfig(item *models.TicketPriorityConfig) *response.TicketPriorityConfigResponse {
	if item == nil {
		return nil
	}
	return &response.TicketPriorityConfigResponse{
		ID:                   item.ID,
		Name:                 item.Name,
		SortNo:               item.SortNo,
		FirstResponseMinutes: item.FirstResponseMinutes,
		ResolutionMinutes:    item.ResolutionMinutes,
		Status:               item.Status,
		Remark:               item.Remark,
	}
}

func BuildTicketPriorityConfigList(list []models.TicketPriorityConfig) []response.TicketPriorityConfigResponse {
	if len(list) == 0 {
		return make([]response.TicketPriorityConfigResponse, 0)
	}
	ret := make([]response.TicketPriorityConfigResponse, 0, len(list))
	for i := range list {
		if item := BuildTicketPriorityConfig(&list[i]); item != nil {
			ret = append(ret, *item)
		}
	}
	return ret
}
