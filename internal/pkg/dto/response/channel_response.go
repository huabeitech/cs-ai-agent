package response

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/enums"
)

type ChannelResponse struct {
	ID          int64        `json:"id"`
	ChannelType string       `json:"channelType"`
	ChannelCode string       `json:"channelCode"`
	AIAgentID   int64        `json:"aiAgentId"`
	AIAgentName string       `json:"aiAgentName,omitempty"`
	Name        string       `json:"name"`
	AppID       string       `json:"appId"`
	ConfigJSON  string       `json:"configJson"`
	SortNo      int          `json:"sortNo"`
	Status      enums.Status `json:"status"`
	Remark      string       `json:"remark"`
}

func BuildChannelResponse(item *models.Channel) ChannelResponse {
	if item == nil {
		return ChannelResponse{}
	}
	return ChannelResponse{
		ID:          item.ID,
		ChannelType: item.ChannelType,
		ChannelCode: item.ChannelCode,
		AIAgentID:   item.AIAgentID,
		Name:        item.Name,
		AppID:       item.AppID,
		ConfigJSON:  item.ConfigJSON,
		SortNo:      item.SortNo,
		Status:      item.Status,
		Remark:      item.Remark,
	}
}
