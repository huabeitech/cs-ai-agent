package response

import "cs-agent/internal/pkg/enums"

type MessageResponse struct {
	ID              int64               `json:"id"`
	ConversationID  int64               `json:"conversationId"`
	ClientMsgID     string              `json:"clientMsgId,omitempty"`
	SenderType      enums.IMSenderType  `json:"senderType"`
	SenderID        int64               `json:"senderId"`
	SenderName      string              `json:"senderName,omitempty"`
	SenderAvatar    string              `json:"senderAvatar,omitempty"`
	MessageType     enums.IMMessageType `json:"messageType"`
	Content         string              `json:"content"`
	Payload         string              `json:"payload,omitempty"`
	SeqNo           int64               `json:"seqNo"`
	SendStatus      int                 `json:"sendStatus"`
	SentAt          string              `json:"sentAt,omitempty"`
	DeliveredAt     string              `json:"deliveredAt,omitempty"`
	ReadAt          string              `json:"readAt,omitempty"`
	CustomerRead    bool                `json:"customerRead"`
	CustomerReadAt  string              `json:"customerReadAt,omitempty"`
	AgentRead       bool                `json:"agentRead"`
	AgentReadAt     string              `json:"agentReadAt,omitempty"`
	RecalledAt      string              `json:"recalledAt,omitempty"`
	QuotedMessageID int64               `json:"quotedMessageId,omitempty"`
}
