package builders

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/services"
)

func BuildTicketCategory(item *models.TicketCategory) *response.TicketCategoryResponse {
	if item == nil {
		return nil
	}
	ret := &response.TicketCategoryResponse{
		ID:       item.ID,
		Name:     item.Name,
		Code:     item.Code,
		ParentID: item.ParentID,
		SortNo:   item.SortNo,
		Status:   item.Status,
		Remark:   item.Remark,
	}
	if item.ParentID > 0 {
		if parent := services.TicketCategoryService.Get(item.ParentID); parent != nil {
			ret.ParentName = parent.Name
		}
	}
	return ret
}

func BuildTicketCategoryList(list []models.TicketCategory) []response.TicketCategoryResponse {
	if len(list) == 0 {
		return make([]response.TicketCategoryResponse, 0)
	}
	ret := make([]response.TicketCategoryResponse, 0, len(list))
	for i := range list {
		if item := BuildTicketCategory(&list[i]); item != nil {
			ret = append(ret, *item)
		}
	}
	return ret
}

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
