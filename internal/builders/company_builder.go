package builders

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto/response"
	"time"
)

func BuildCompanyResponse(item *models.Company) response.CompanyResponse {
	if item == nil {
		return response.CompanyResponse{}
	}
	return response.CompanyResponse{
		ID:        item.ID,
		Name:      item.Name,
		Code:      item.Code,
		Industry:  item.Industry,
		Website:   item.Website,
		Province:  item.Province,
		City:      item.City,
		Address:   item.Address,
		Status:    item.Status,
		Remark:    item.Remark,
		CreatedAt: item.CreatedAt.Format(time.DateTime),
		UpdatedAt: item.UpdatedAt.Format(time.DateTime),
	}
}

func BuildCompanyList(list []models.Company) []response.CompanyResponse {
	results := make([]response.CompanyResponse, 0, len(list))
	for _, item := range list {
		results = append(results, BuildCompanyResponse(&item))
	}
	return results
}
