package response

import "time"

type SkillDefinitionResponse struct {
	ID             int64     `json:"id"`
	Code           string    `json:"code"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	Instruction    string    `json:"instruction"`
	Examples       []string  `json:"examples"`
	ToolWhitelist  []string  `json:"toolWhitelist"`
	Priority       int       `json:"priority"`
	Status         int       `json:"status"`
	StatusName     string    `json:"statusName"`
	Remark         string    `json:"remark"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	CreateUserName string    `json:"createUserName"`
	UpdateUserName string    `json:"updateUserName"`
}

type SkillDebugRunResponse struct {
	SkillCode        string   `json:"skillCode"`
	SkillName        string   `json:"skillName"`
	ReplyText        string   `json:"replyText"`
	PlanReason       string   `json:"planReason"`
	SkillRouteTrace  string   `json:"skillRouteTrace"`
	ToolWhitelist    []string `json:"toolWhitelist"`
	ExposedToolCodes []string `json:"exposedToolCodes"`
	InvokedToolCodes []string `json:"invokedToolCodes"`
	ToolSearchTrace  string   `json:"toolSearchTrace"`
	GraphToolTrace   string   `json:"graphToolTrace"`
	GraphToolCode    string   `json:"graphToolCode"`
	InterruptType    string   `json:"interruptType"`
	CheckPointID     string   `json:"checkPointId"`
	Interrupted      bool     `json:"interrupted"`
	TraceData        string   `json:"traceData"`
	ErrorMessage     string   `json:"errorMessage"`
	ConversationID   int64    `json:"conversationId"`
	AIAgentID        int64    `json:"aiAgentId"`
}

type AgentRunLogResponse struct {
	ID                int64  `json:"id"`
	ConversationID    int64  `json:"conversationId"`
	MessageID         int64  `json:"messageId"`
	AIAgentID         int64  `json:"aiAgentId"`
	AIConfigID        int64  `json:"aiConfigId"`
	UserMessage       string `json:"userMessage"`
	PlannedAction     string `json:"plannedAction"`
	PlannedSkillCode  string `json:"plannedSkillCode"`
	PlannedSkillName  string `json:"plannedSkillName"`
	SkillRouteTrace   string `json:"skillRouteTrace"`
	ToolSearchTrace   string `json:"toolSearchTrace"`
	GraphToolTrace    string `json:"graphToolTrace"`
	GraphToolCode     string `json:"graphToolCode"`
	RecommendedAction string `json:"recommendedAction"`
	RiskLevel         string `json:"riskLevel"`
	TicketDraftReady  bool   `json:"ticketDraftReady"`
	HandoffReason     string `json:"handoffReason"`
	PlannedToolCode   string `json:"plannedToolCode"`
	PlanReason        string `json:"planReason"`
	InterruptType     string `json:"interruptType"`
	ResumeSource      string `json:"resumeSource"`
	HitlStatus        string `json:"hitlStatus"`
	HitlStatusName    string `json:"hitlStatusName"`
	HitlSummary       string `json:"hitlSummary"`
	FinalAction       string `json:"finalAction"`
	FinalStatus       string `json:"finalStatus"`
	ReplyText         string `json:"replyText"`
	ErrorMessage      string `json:"errorMessage"`
	LatencyMs         int64  `json:"latencyMs"`
	TraceData         string `json:"traceData"`
	CreatedAt         string `json:"createdAt"`
}

type AgentRunGraphSummaryResponse struct {
	TriageCount              int64 `json:"triageCount"`
	TriagePrepareTicket      int64 `json:"triagePrepareTicket"`
	TriagePrepareTicketReady int64 `json:"triagePrepareTicketReady"`
	TriageHandoff            int64 `json:"triageHandoff"`
	TriageContinueAnswering  int64 `json:"triageContinueAnswering"`
	AnalyzeCount             int64 `json:"analyzeCount"`
	PrepareDraftCount        int64 `json:"prepareDraftCount"`
	CreateTicketCount        int64 `json:"createTicketCount"`
	HandoffCount             int64 `json:"handoffCount"`
}
