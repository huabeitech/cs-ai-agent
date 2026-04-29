package builders

import (
	"cs-agent/internal/pkg/dto/response"
	"cs-agent/internal/services"
	"time"
)

func BuildAgentTeamScheduleBatchPreviewResponse(result *services.AgentTeamScheduleBatchPreviewResult) *response.AgentTeamScheduleBatchPreviewResponse {
	if result == nil {
		return nil
	}
	items := make([]response.AgentTeamScheduleBatchPreviewItem, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, response.AgentTeamScheduleBatchPreviewItem{
			TeamID:         item.TeamID,
			TeamName:       item.TeamName,
			Date:           item.Date.Format(time.DateOnly),
			Weekday:        item.Weekday,
			StartAt:        item.StartAt.Format(time.DateTime),
			EndAt:          item.EndAt.Format(time.DateTime),
			Remark:         item.Remark,
			Conflict:       item.Conflict,
			ConflictReason: item.ConflictReason,
		})
	}
	return &response.AgentTeamScheduleBatchPreviewResponse{
		Total:    result.Total,
		Conflict: result.Conflict,
		Items:    items,
	}
}

func BuildAgentTeamScheduleBatchGenerateResponse(result *services.AgentTeamScheduleBatchGenerateResult) *response.AgentTeamScheduleBatchGenerateResponse {
	if result == nil {
		return nil
	}
	return &response.AgentTeamScheduleBatchGenerateResponse{Created: result.Created}
}
