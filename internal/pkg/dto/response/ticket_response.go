package response

import "cs-agent/internal/pkg/enums"

type TicketSLAResponse struct {
	SLAType       enums.TicketSLAType   `json:"slaType"`
	TargetMinutes int                   `json:"targetMinutes"`
	Status        enums.TicketSLAStatus `json:"status"`
	StartedAt     string                `json:"startedAt,omitempty"`
	PausedAt      string                `json:"pausedAt,omitempty"`
	StoppedAt     string                `json:"stoppedAt,omitempty"`
	BreachedAt    string                `json:"breachedAt,omitempty"`
	ElapsedMin    int                   `json:"elapsedMin"`
}

type TicketWatcherResponse struct {
	ID       int64  `json:"id"`
	UserID   int64  `json:"userId"`
	UserName string `json:"userName,omitempty"`
}

type TicketCommentResponse struct {
	ID          int64                   `json:"id"`
	TicketID    int64                   `json:"ticketId"`
	CommentType enums.TicketCommentType `json:"commentType"`
	AuthorType  enums.IMSenderType      `json:"authorType"`
	AuthorID    int64                   `json:"authorId"`
	AuthorName  string                  `json:"authorName,omitempty"`
	ContentType string                  `json:"contentType"`
	Content     string                  `json:"content"`
	Payload     string                  `json:"payload,omitempty"`
	CreatedAt   string                  `json:"createdAt,omitempty"`
}

type TicketEventLogResponse struct {
	ID           int64                 `json:"id"`
	TicketID     int64                 `json:"ticketId"`
	EventType    enums.TicketEventType `json:"eventType"`
	OperatorType enums.IMSenderType    `json:"operatorType"`
	OperatorID   int64                 `json:"operatorId"`
	OperatorName string                `json:"operatorName,omitempty"`
	OldValue     string                `json:"oldValue,omitempty"`
	NewValue     string                `json:"newValue,omitempty"`
	Content      string                `json:"content,omitempty"`
	Payload      string                `json:"payload,omitempty"`
	CreatedAt    string                `json:"createdAt,omitempty"`
}

type TicketResponse struct {
	ID                  int64                `json:"id"`
	TicketNo            string               `json:"ticketNo"`
	Title               string               `json:"title"`
	Description         string               `json:"description"`
	Source              enums.TicketSource   `json:"source"`
	Channel             string               `json:"channel"`
	CustomerID          int64                `json:"customerId"`
	ConversationID      int64                `json:"conversationId"`
	CategoryID          int64                `json:"categoryId"`
	CategoryName        string               `json:"categoryName,omitempty"`
	Type                string               `json:"type"`
	Priority            enums.TicketPriority `json:"priority"`
	Severity            enums.TicketSeverity `json:"severity"`
	Status              enums.TicketStatus   `json:"status"`
	CurrentTeamID       int64                `json:"currentTeamId"`
	CurrentTeamName     string               `json:"currentTeamName,omitempty"`
	CurrentAssigneeID   int64                `json:"currentAssigneeId"`
	CurrentAssigneeName string               `json:"currentAssigneeName,omitempty"`
	WatchedByMe         bool                 `json:"watchedByMe"`
	PendingReason       string               `json:"pendingReason,omitempty"`
	CloseReason         string               `json:"closeReason,omitempty"`
	ResolutionCode      string               `json:"resolutionCode,omitempty"`
	ResolutionCodeName  string               `json:"resolutionCodeName,omitempty"`
	ResolutionSummary   string               `json:"resolutionSummary,omitempty"`
	FirstResponseAt     string               `json:"firstResponseAt,omitempty"`
	ResolvedAt          string               `json:"resolvedAt,omitempty"`
	ClosedAt            string               `json:"closedAt,omitempty"`
	DueAt               string               `json:"dueAt,omitempty"`
	NextReplyDeadlineAt string               `json:"nextReplyDeadlineAt,omitempty"`
	ResolveDeadlineAt   string               `json:"resolveDeadlineAt,omitempty"`
	ReopenedCount       int                  `json:"reopenedCount"`
	CreatedAt           string               `json:"createdAt,omitempty"`
	UpdatedAt           string               `json:"updatedAt,omitempty"`
	Customer            *CustomerResponse    `json:"customer,omitempty"`
	SLA                 []TicketSLAResponse  `json:"sla,omitempty"`
}

type TicketDetailResponse struct {
	Ticket   TicketResponse           `json:"ticket"`
	Watchers []TicketWatcherResponse  `json:"watchers,omitempty"`
	Comments []TicketCommentResponse  `json:"comments,omitempty"`
	Events   []TicketEventLogResponse `json:"events,omitempty"`
}

type TicketSummaryResponse struct {
	All             int64 `json:"all"`
	Mine            int64 `json:"mine"`
	Watching        int64 `json:"watching"`
	Unassigned      int64 `json:"unassigned"`
	PendingCustomer int64 `json:"pendingCustomer"`
	PendingInternal int64 `json:"pendingInternal"`
	Overdue         int64 `json:"overdue"`
}
