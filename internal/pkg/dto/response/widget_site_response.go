package response

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/enums"
)

type WidgetSiteResponse struct {
	ID          int64        `json:"id"`
	AIAgentID   int64        `json:"aiAgentId"`
	AIAgentName string       `json:"aiAgentName"`
	Name        string       `json:"name"`
	AppID       string       `json:"appId"`
	Status      enums.Status `json:"status"`
	Remark      string       `json:"remark"`
}

func BuildWidgetSiteResponse(item *models.WidgetSite) WidgetSiteResponse {
	return WidgetSiteResponse{
		ID:        item.ID,
		AIAgentID: item.AIAgentID,
		Name:      item.Name,
		AppID:     item.AppID,
		Status:    item.Status,
		Remark:    item.Remark,
	}
}
