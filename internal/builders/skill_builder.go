package builders

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/pkg/enums"
)

func BuildSkillDefinitionResponse(item *models.SkillDefinition) response.SkillDefinitionResponse {
	return response.SkillDefinitionResponse{
		ID:             item.ID,
		Code:           item.Code,
		Name:           item.Name,
		Description:    item.Description,
		Prompt:         item.Prompt,
		Priority:       item.Priority,
		Status:         int(item.Status),
		StatusName:     getSkillStatusName(item.Status),
		Remark:         item.Remark,
		CreatedAt:      item.CreatedAt,
		UpdatedAt:      item.UpdatedAt,
		CreateUserName: item.CreateUserName,
		UpdateUserName: item.UpdateUserName,
	}
}

func getSkillStatusName(status enums.Status) string {
	if label := enums.GetStatusLabel(status); label != "" {
		return label
	}
	return "未知"
}
