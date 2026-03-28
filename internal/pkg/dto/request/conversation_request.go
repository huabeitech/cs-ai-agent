package request

import "cs-agent/internal/pkg/enums"

type AgentConversationFilter string

const (
	AgentConversationFilterMine    AgentConversationFilter = "mine"
	AgentConversationFilterActive  AgentConversationFilter = "active"
	AgentConversationFilterPending AgentConversationFilter = "pending"
	AgentConversationFilterClosed  AgentConversationFilter = "closed"
)

type ConversationListRequest struct {
	Status            int    `json:"status"`
	ChannelType       string `json:"channelType"`
	ServiceMode       int    `json:"serviceMode"`
	CurrentAssigneeID int64  `json:"currentAssigneeId"`
	SourceUserID      int64  `json:"sourceUserId"`
	Keyword           string `json:"keyword"`
	TagID             int64  `json:"tagId"`
}

type AssignConversationRequest struct {
	ConversationID int64  `json:"conversationId"`
	AssigneeID     int64  `json:"assigneeId"`
	Reason         string `json:"reason"`
}

type DispatchConversationRequest struct {
	ConversationID int64 `json:"conversationId"`
}

type TransferConversationRequest struct {
	ConversationID int64  `json:"conversationId"`
	ToUserID       int64  `json:"toUserId"`
	Reason         string `json:"reason"`
}

type CloseConversationRequest struct {
	ConversationID int64  `json:"conversationId"`
	CloseReason    string `json:"closeReason"`
}

type ReadConversationRequest struct {
	ConversationID int64 `json:"conversationId"`
	MessageID      int64 `json:"messageId"`
}

type CreateOrMatchConversationRequest struct {
	ChannelType enums.ExternalSource `json:"channelType"`
	Subject     string                       `json:"subject"`
}

type AddConversationTagRequest struct {
	ConversationID int64 `json:"conversationId"`
	TagID          int64 `json:"tagId"`
}

type RemoveConversationTagRequest struct {
	ConversationID int64 `json:"conversationId"`
	TagID          int64 `json:"tagId"`
}
