package request

import "cs-agent/internal/pkg/enums"

type SkillDefinitionListRequest struct {
	Name   string `json:"name"`
	Code   string `json:"code"`
	Status int    `json:"status"`
}

type CreateSkillDefinitionRequest struct {
	Code            string                   `json:"code"`
	Name            string                   `json:"name"`
	Description     string                   `json:"description"`
	Prompt          string                   `json:"prompt"`
	ExecutionMode   enums.SkillExecutionMode `json:"executionMode"`
	ExecutionConfig string                   `json:"executionConfig"`
	Remark          string                   `json:"remark"`
}

type UpdateSkillDefinitionRequest struct {
	ID int64 `json:"id"`
	CreateSkillDefinitionRequest
}

type DeleteSkillDefinitionRequest struct {
	ID int64 `json:"id"`
}

type UpdateSkillDefinitionStatusRequest struct {
	ID     int64 `json:"id"`
	Status int   `json:"status"`
}

type SkillDebugRunRequest struct {
	AIAgentID      int64  `json:"aiAgentId"`
	ConversationID int64  `json:"conversationId"`
	SkillCode      string `json:"skillCode"`
	UserMessage    string `json:"userMessage"`
}
