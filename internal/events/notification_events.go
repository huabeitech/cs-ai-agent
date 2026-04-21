package events

const (
	ConversationAssignTypeAssign     = "assign"
	ConversationAssignTypeTransfer   = "transfer"
	ConversationAssignTypeAutoAssign = "auto_assign"
)

type TicketCreatedEvent struct {
	TicketID   int64
	OperatorID int64
}

type TicketAssignedEvent struct {
	TicketID   int64
	FromUserID int64
	ToUserID   int64
	OperatorID int64
	Reason     string
}

type ConversationAssignedEvent struct {
	ConversationID int64
	FromUserID     int64
	ToUserID       int64
	OperatorID     int64
	Reason         string
	AssignType     string
}
