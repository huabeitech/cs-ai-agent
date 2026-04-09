package builders

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/pkg/utils"
	"time"
)

func BuildCustomerContactResponse(item *models.CustomerContact) response.CustomerContactResponse {
	if item == nil {
		return response.CustomerContactResponse{}
	}
	return response.CustomerContactResponse{
		ID:           item.ID,
		CustomerID:   item.CustomerID,
		ContactType:  item.ContactType,
		ContactValue: item.ContactValue,
		IsPrimary:    item.IsPrimary,
		IsVerified:   item.IsVerified,
		VerifiedAt:   utils.FormatTimePtr(item.VerifiedAt),
		Source:       item.Source,
		Status:       item.Status,
		Remark:       item.Remark,
		CreatedAt:    item.CreatedAt.Format(time.DateTime),
		UpdatedAt:    item.UpdatedAt.Format(time.DateTime),
	}
}

func BuildCustomerContactList(list []models.CustomerContact) []response.CustomerContactResponse {
	results := make([]response.CustomerContactResponse, 0, len(list))
	for i := range list {
		results = append(results, BuildCustomerContactResponse(&list[i]))
	}
	return results
}
