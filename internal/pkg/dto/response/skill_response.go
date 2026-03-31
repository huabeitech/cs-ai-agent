package response

import "time"

type SkillDefinitionResponse struct {
	ID                int64     `json:"id"`
	Code              string    `json:"code"`
	Name              string    `json:"name"`
	Description       string    `json:"description"`
	Prompt            string    `json:"prompt"`
	ExecutionMode     string    `json:"executionMode"`
	ExecutionModeName string    `json:"executionModeName"`
	ExecutionConfig   string    `json:"executionConfig"`
	Priority          int       `json:"priority"`
	Status            int       `json:"status"`
	StatusName        string    `json:"statusName"`
	Remark            string    `json:"remark"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
	CreateUserName    string    `json:"createUserName"`
	UpdateUserName    string    `json:"updateUserName"`
}

type SkillDebugRunResponse struct {
	SkillCode      string `json:"skillCode"`
	SkillName      string `json:"skillName"`
	ReplyText      string `json:"replyText"`
	RunLogID       int64  `json:"runLogId"`
	ConversationID int64  `json:"conversationId"`
	AIAgentID      int64  `json:"aiAgentId"`
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
	PlanReason       string `json:"planReason"`
	FinalAction      string `json:"finalAction"`
	ReplyText        string `json:"replyText"`
	ErrorMessage     string `json:"errorMessage"`
	LatencyMs        int64  `json:"latencyMs"`
	CreatedAt        string `json:"createdAt"`
}
