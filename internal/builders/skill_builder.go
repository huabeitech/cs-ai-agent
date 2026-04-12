package builders

import (
	"encoding/json"

	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/pkg/enums"
)

func BuildSkillDefinitionResponse(item *models.SkillDefinition) response.SkillDefinitionResponse {
	examples := make([]string, 0)
	if raw := item.Examples; raw != "" {
		_ = json.Unmarshal([]byte(raw), &examples)
	}
	allowedToolCodes := make([]string, 0)
	if raw := item.AllowedToolCodes; raw != "" {
		_ = json.Unmarshal([]byte(raw), &allowedToolCodes)
	}
	return response.SkillDefinitionResponse{
		ID:               item.ID,
		Code:             item.Code,
		Name:             item.Name,
		Description:      item.Description,
		Instruction:      item.Instruction,
		Examples:         examples,
		AllowedToolCodes: allowedToolCodes,
		Priority:         item.Priority,
		Status:           int(item.Status),
		StatusName:       getSkillStatusName(item.Status),
		Remark:           item.Remark,
		CreatedAt:        item.CreatedAt,
		UpdatedAt:        item.UpdatedAt,
		CreateUserName:   item.CreateUserName,
		UpdateUserName:   item.UpdateUserName,
	}
}

func getSkillStatusName(status enums.Status) string {
	if label := enums.GetStatusLabel(status); label != "" {
		return label
	}
	return "未知"
}
