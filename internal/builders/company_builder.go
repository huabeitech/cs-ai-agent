package builders

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto/response"
	"time"
)

func BuildCompany(item *models.Company) *response.CompanyResponse {
	if item == nil {
		return nil
	}
	return &response.CompanyResponse{
		ID:        item.ID,
		Name:      item.Name,
		Code:      item.Code,
		Status:    item.Status,
		Remark:    item.Remark,
		CreatedAt: item.CreatedAt.Format(time.DateTime),
		UpdatedAt: item.UpdatedAt.Format(time.DateTime),
	}
}

func BuildCompanyList(list []models.Company) []response.CompanyResponse {
	results := make([]response.CompanyResponse, 0, len(list))
	for _, item := range list {
		if company := BuildCompany(&item); company != nil {
			results = append(results, *company)
		}
	}
	return results
}
