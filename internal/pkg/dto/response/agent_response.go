package response

import "cs-agent/internal/pkg/enums"

type AgentProfileResponse struct {
	ID                    int64               `json:"id"`
	UserID                int64               `json:"userId"`
	TeamID                int64               `json:"teamId"`
	TeamName              string              `json:"teamName,omitempty"`
	Username              string              `json:"username,omitempty"`
	Nickname              string              `json:"nickname,omitempty"`
	AgentCode             string              `json:"agentCode"`
	DisplayName           string              `json:"displayName"`
	Avatar                string              `json:"avatar"`
	ServiceStatus         enums.ServiceStatus `json:"serviceStatus"`
	MaxConcurrentCount    int                 `json:"maxConcurrentCount"`
	PriorityLevel         int                 `json:"priorityLevel"`
	AutoAssignEnabled     bool                `json:"autoAssignEnabled"`
	ReceiveOfflineMessage bool                `json:"receiveOfflineMessage"`
	LastOnlineAt          string              `json:"lastOnlineAt,omitempty"`
	LastStatusAt          string              `json:"lastStatusAt,omitempty"`
	Remark                string              `json:"remark"`
}

type AgentTeamResponse struct {
	ID             int64        `json:"id"`
	Name           string       `json:"name"`
	LeaderUserID   int64        `json:"leaderUserId"`
	LeaderUsername string       `json:"leaderUsername,omitempty"`
	LeaderNickname string       `json:"leaderNickname,omitempty"`
	Status         enums.Status `json:"status"`
	Description    string       `json:"description"`
	Remark         string       `json:"remark"`
}

type AgentTeamScheduleResponse struct {
	ID         int64  `json:"id"`
	TeamID     int64  `json:"teamId"`
	TeamName   string `json:"teamName,omitempty"`
	StartAt    string `json:"startAt"`
	EndAt      string `json:"endAt"`
	SourceType string `json:"sourceType"`
	Remark     string `json:"remark"`
}
