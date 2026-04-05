package request

type CreateTicketRequest struct {
	Title             string         `json:"title"`
	Description       string         `json:"description"`
	Source            string         `json:"source"`
	Channel           string         `json:"channel"`
	CustomerID        int64          `json:"customerId"`
	ConversationID    int64          `json:"conversationId"`
	CategoryID        int64          `json:"categoryId"`
	Type              string         `json:"type"`
	Priority          int64          `json:"priority"`
	Severity          int            `json:"severity"`
	CurrentTeamID     int64          `json:"currentTeamId"`
	CurrentAssigneeID int64          `json:"currentAssigneeId"`
	DueAt             string         `json:"dueAt"`
	CustomFields      map[string]any `json:"customFields"`
}

type CreateTicketFromConversationRequest struct {
	ConversationID     int64          `json:"conversationId"`
	Title              string         `json:"title"`
	Description        string         `json:"description"`
	CategoryID         int64          `json:"categoryId"`
	Priority           int64          `json:"priority"`
	Severity           int            `json:"severity"`
	CurrentTeamID      int64          `json:"currentTeamId"`
	CurrentAssigneeID  int64          `json:"currentAssigneeId"`
	SyncToConversation bool           `json:"syncToConversation"`
	CustomFields       map[string]any `json:"customFields"`
}

type UpdateTicketRequest struct {
	TicketID          int64          `json:"ticketId"`
	Title             string         `json:"title"`
	Description       string         `json:"description"`
	CategoryID        int64          `json:"categoryId"`
	Type              string         `json:"type"`
	Priority          int64          `json:"priority"`
	Severity          int            `json:"severity"`
	CurrentTeamID     int64          `json:"currentTeamId"`
	CurrentAssigneeID int64          `json:"currentAssigneeId"`
	DueAt             string         `json:"dueAt"`
	CustomFields      map[string]any `json:"customFields"`
}

type AssignTicketRequest struct {
	TicketID int64  `json:"ticketId"`
	ToUserID int64  `json:"toUserId"`
	ToTeamID int64  `json:"toTeamId"`
	Reason   string `json:"reason"`
}

type ChangeTicketStatusRequest struct {
	TicketID          int64  `json:"ticketId"`
	Status            string `json:"status"`
	PendingReason     string `json:"pendingReason"`
	CloseReason       string `json:"closeReason"`
	ResolutionCode    string `json:"resolutionCode"`
	ResolutionSummary string `json:"resolutionSummary"`
	Reason            string `json:"reason"`
}

type ReplyTicketRequest struct {
	TicketID    int64  `json:"ticketId"`
	ContentType string `json:"contentType"`
	Content     string `json:"content"`
	Payload     string `json:"payload"`
}

type InternalNoteRequest struct {
	TicketID    int64  `json:"ticketId"`
	ContentType string `json:"contentType"`
	Content     string `json:"content"`
	Payload     string `json:"payload"`
}

type CloseTicketRequest struct {
	TicketID    int64  `json:"ticketId"`
	CloseReason string `json:"closeReason"`
}

type ReopenTicketRequest struct {
	TicketID int64  `json:"ticketId"`
	Reason   string `json:"reason"`
}

type WatchTicketRequest struct {
	TicketID int64 `json:"ticketId"`
}

type BatchAssignTicketRequest struct {
	TicketIDs []int64 `json:"ticketIds"`
	ToUserID  int64   `json:"toUserId"`
	ToTeamID  int64   `json:"toTeamId"`
	Reason    string  `json:"reason"`
}

type BatchChangeTicketStatusRequest struct {
	TicketIDs         []int64 `json:"ticketIds"`
	Status            string  `json:"status"`
	PendingReason     string  `json:"pendingReason"`
	CloseReason       string  `json:"closeReason"`
	ResolutionCode    string  `json:"resolutionCode"`
	ResolutionSummary string  `json:"resolutionSummary"`
	Reason            string  `json:"reason"`
}

type BatchWatchTicketRequest struct {
	TicketIDs []int64 `json:"ticketIds"`
	Watched   bool    `json:"watched"`
}

type AddTicketRelationRequest struct {
	TicketID        int64  `json:"ticketId"`
	RelatedTicketID int64  `json:"relatedTicketId"`
	RelatedTicketNo string `json:"relatedTicketNo"`
	RelationType    string `json:"relationType"`
}

type DeleteTicketRelationRequest struct {
	TicketID   int64 `json:"ticketId"`
	RelationID int64 `json:"relationId"`
}

type AddTicketCollaboratorRequest struct {
	TicketID int64 `json:"ticketId"`
	UserID   int64 `json:"userId"`
}

type DeleteTicketCollaboratorRequest struct {
	TicketID       int64 `json:"ticketId"`
	CollaboratorID int64 `json:"collaboratorId"`
}

type LinkTicketCustomerRequest struct {
	TicketID   int64 `json:"ticketId"`
	CustomerID int64 `json:"customerId"`
}

type SaveTicketViewRequest struct {
	ID      int64          `json:"id"`
	Name    string         `json:"name"`
	Filters map[string]any `json:"filters"`
}

type DeleteTicketViewRequest struct {
	ID int64 `json:"id"`
}
