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
	TeamID  int64  `json:"teamId"`
	StartAt string `json:"startAt"`
	EndAt   string `json:"endAt"`
	Remark  string `json:"remark"`
}

type UpdateAgentTeamScheduleRequest struct {
	ID int64 `json:"id"`
	CreateAgentTeamScheduleRequest
}

type DeleteAgentTeamScheduleRequest struct {
	ID int64 `json:"id"`
}

type AgentTeamScheduleCalendarRequest struct {
	StartAt string `json:"startAt"`
	EndAt   string `json:"endAt"`
	TeamID  int64  `json:"teamId"`
}

type AgentTeamScheduleBatchRequest struct {
	TeamIDs   []int64 `json:"teamIds"`
	StartDate string  `json:"startDate"`
	EndDate   string  `json:"endDate"`
	Weekdays  []int   `json:"weekdays"`
	StartTime string  `json:"startTime"`
	EndTime   string  `json:"endTime"`
	Remark    string  `json:"remark"`
}
