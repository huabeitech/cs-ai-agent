package response

import "cs-agent/internal/pkg/enums"

type ConversationTagResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type ConversationParticipantResponse struct {
	ID                    int64        `json:"id"`
	ParticipantType       string       `json:"participantType"`
	ParticipantID         int64        `json:"participantId"`
	ExternalParticipantID string       `json:"externalParticipantId,omitempty"`
	JoinedAt              string       `json:"joinedAt,omitempty"`
	LeftAt                string       `json:"leftAt,omitempty"`
	Status                enums.Status `json:"status"`
}

type ConversationResponse struct {
	ID                        int64                           `json:"id"`
	AIAgentID                 int64                           `json:"aiAgentId"`
	CustomerID                int64                           `json:"customerId"`
	ExternalSource            enums.ExternalSource            `json:"externalSource"`
	SourceUserID              int64                           `json:"sourceUserId"`
	ExternalID            string                          `json:"externalUserId"`
	Subject                   string                          `json:"subject"`
	Status                    enums.IMConversationStatus      `json:"status"`
	ServiceMode               enums.IMConversationServiceMode `json:"serviceMode"`
	Priority                  int                             `json:"priority"`
	CurrentAssigneeID         int64                           `json:"currentAssigneeId"`
	CurrentAssigneeName       string                          `json:"currentAssigneeName,omitempty"`
	LastMessageID             int64                           `json:"lastMessageId"`
	LastMessageAt             string                          `json:"lastMessageAt,omitempty"`
	LastActiveAt              string                          `json:"lastActiveAt,omitempty"`
	LastMessageSummary        string                          `json:"lastMessageSummary,omitempty"`
	CustomerUnreadCount       int                             `json:"customerUnreadCount"`
	AgentUnreadCount          int                             `json:"agentUnreadCount"`
	CustomerLastReadMessageID int64                           `json:"customerLastReadMessageId"`
	CustomerLastReadSeqNo     int64                           `json:"customerLastReadSeqNo"`
	CustomerLastReadAt        string                          `json:"customerLastReadAt,omitempty"`
	AgentLastReadMessageID    int64                           `json:"agentLastReadMessageId"`
	AgentLastReadSeqNo        int64                           `json:"agentLastReadSeqNo"`
	AgentLastReadAt           string                          `json:"agentLastReadAt,omitempty"`
	ClosedAt                  string                          `json:"closedAt,omitempty"`
	ClosedBy                  int64                           `json:"closedBy"`
	ClosedByName              string                          `json:"closedByName,omitempty"`
	CloseReason               string                          `json:"closeReason,omitempty"`
	Tags                      []ConversationTagResponse       `json:"tags,omitempty"`
}

type ConversationDetailResponse struct {
	ConversationResponse
	Participants []ConversationParticipantResponse `json:"participants,omitempty"`
}
