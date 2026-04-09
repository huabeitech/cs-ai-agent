package response

import "time"

type SkillDefinitionResponse struct {
	ID               int64     `json:"id"`
	Code             string    `json:"code"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	Content          string    `json:"content"`
	Examples         []string  `json:"examples"`
	AllowedToolCodes []string  `json:"allowedToolCodes"`
	Priority         int       `json:"priority"`
	Status           int       `json:"status"`
	StatusName       string    `json:"statusName"`
	Remark           string    `json:"remark"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
	CreateUserName   string    `json:"createUserName"`
	UpdateUserName   string    `json:"updateUserName"`
}

type SkillDebugRunResponse struct {
	SkillCode        string   `json:"skillCode"`
	SkillName        string   `json:"skillName"`
	ReplyText        string   `json:"replyText"`
	PlanReason       string   `json:"planReason"`
	SkillRouteTrace  string   `json:"skillRouteTrace"`
	ToolCodes        []string `json:"toolCodes"`
	InvokedToolCodes []string `json:"invokedToolCodes"`
	CheckPointID     string   `json:"checkPointId"`
	Interrupted      bool     `json:"interrupted"`
	TraceData        string   `json:"traceData"`
	ErrorMessage     string   `json:"errorMessage"`
	ConversationID   int64    `json:"conversationId"`
	AIAgentID        int64    `json:"aiAgentId"`
}

type AgentRunLogResponse struct {
	ID               int64  `json:"id"`
	ConversationID   int64  `json:"conversationId"`
	MessageID        int64  `json:"messageId"`
	AIAgentID        int64  `json:"aiAgentId"`
	AIConfigID       int64  `json:"aiConfigId"`
	UserMessage      string `json:"userMessage"`
	PlannedAction    string `json:"plannedAction"`
	PlannedSkillCode string `json:"plannedSkillCode"`
	PlannedToolCode  string `json:"plannedToolCode"`
	PlanReason       string `json:"planReason"`
	FinalAction      string `json:"finalAction"`
	ReplyText        string `json:"replyText"`
	ErrorMessage     string `json:"errorMessage"`
	LatencyMs        int64  `json:"latencyMs"`
	TraceData        string `json:"traceData"`
	CreatedAt        string `json:"createdAt"`
}
