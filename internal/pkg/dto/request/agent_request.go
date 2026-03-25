package request

import "cs-agent/internal/pkg/enums"

type CreateAgentProfileRequest struct {
	UserID                int64               `json:"userId"`
	TeamID                int64               `json:"teamId"`
	AgentCode             string              `json:"agentCode"`
	DisplayName           string              `json:"displayName"`
	Avatar                string              `json:"avatar"`
	ServiceStatus         enums.ServiceStatus `json:"serviceStatus"`
	MaxConcurrentCount    int                 `json:"maxConcurrentCount"`
	PriorityLevel         int                 `json:"priorityLevel"`
	AutoAssignEnabled     bool                `json:"autoAssignEnabled"`
	ReceiveOfflineMessage bool                `json:"receiveOfflineMessage"`
	Remark                string              `json:"remark"`
}

type UpdateAgentProfileRequest struct {
	ID int64 `json:"id"`
	CreateAgentProfileRequest
}

type DeleteAgentProfileRequest struct {
	ID int64 `json:"id"`
}

type CreateAgentTeamRequest struct {
	Name         string `json:"name"`
	LeaderUserID int64  `json:"leaderUserId"`
	Status       int    `json:"status"`
	Description  string `json:"description"`
	Remark       string `json:"remark"`
}

type UpdateAgentTeamRequest struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	LeaderUserID int64  `json:"leaderUserId"`
	Status       int    `json:"status"`
	Description  string `json:"description"`
	Remark       string `json:"remark"`
}

type DeleteAgentTeamRequest struct {
	ID int64 `json:"id"`
}

type CreateAgentTeamScheduleRequest struct {
	TeamID     int64  `json:"teamId"`
	StartAt    string `json:"startAt"`
	EndAt      string `json:"endAt"`
	SourceType string `json:"sourceType"`
	Remark     string `json:"remark"`
}

type UpdateAgentTeamScheduleRequest struct {
	ID int64 `json:"id"`
	CreateAgentTeamScheduleRequest
}

type DeleteAgentTeamScheduleRequest struct {
	ID int64 `json:"id"`
}
