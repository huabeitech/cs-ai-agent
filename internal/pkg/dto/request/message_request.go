package request

import "cs-agent/internal/pkg/enums"

type MessageListRequest struct {
	ConversationID int64  `json:"conversationId"`
	SenderType     string `json:"senderType"`
	MessageType    string `json:"messageType"`
}

type SendConversationMessageRequest struct {
	ConversationID int64               `json:"conversationId"`
	MessageType    enums.IMMessageType `json:"messageType"`
	Content        string              `json:"content"`
	Payload        string              `json:"payload"`
	ClientMsgID    string              `json:"clientMsgId"`
}
