package response

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/enums"
)

type ChannelResponse struct {
	ID          int64        `json:"id"`
	ChannelType string       `json:"channelType"`
	ChannelID   string       `json:"channelId"`
	AIAgentID   int64        `json:"aiAgentId"`
	AIAgentName string       `json:"aiAgentName,omitempty"`
	Name        string       `json:"name"`
	ConfigJSON  string       `json:"configJson"`
	Status      enums.Status `json:"status"`
	Remark      string       `json:"remark"`
}

type WxWorkKFAccountResponse struct {
	OpenKfID        string `json:"openKfId"`
	Name            string `json:"name"`
	Avatar          string `json:"avatar"`
	ManagePrivilege bool   `json:"managePrivilege"`
}

func BuildChannelResponse(item *models.Channel) ChannelResponse {
	if item == nil {
		return ChannelResponse{}
	}
	return ChannelResponse{
		ID:          item.ID,
		ChannelType: item.ChannelType,
		ChannelID:   item.ChannelID,
		AIAgentID:   item.AIAgentID,
		Name:        item.Name,
		ConfigJSON:  item.ConfigJSON,
		Status:      item.Status,
		Remark:      item.Remark,
	}
}
