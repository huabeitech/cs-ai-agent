package enums

type TicketStatus string

const (
	TicketStatusNew             TicketStatus = "new"
	TicketStatusOpen            TicketStatus = "open"
	TicketStatusPendingCustomer TicketStatus = "pending_customer"
	TicketStatusPendingInternal TicketStatus = "pending_internal"
	TicketStatusResolved        TicketStatus = "resolved"
	TicketStatusClosed          TicketStatus = "closed"
	TicketStatusCancelled       TicketStatus = "cancelled"
)

var TicketStatusValues = []TicketStatus{
	TicketStatusNew,
	TicketStatusOpen,
	TicketStatusPendingCustomer,
	TicketStatusPendingInternal,
	TicketStatusResolved,
	TicketStatusClosed,
	TicketStatusCancelled,
}

var ticketStatusLabelMap = map[TicketStatus]string{
	TicketStatusNew:             "新建",
	TicketStatusOpen:            "处理中",
	TicketStatusPendingCustomer: "待客户反馈",
	TicketStatusPendingInternal: "待内部处理",
	TicketStatusResolved:        "已解决",
	TicketStatusClosed:          "已关闭",
	TicketStatusCancelled:       "已取消",
}

func GetTicketStatusLabel(status TicketStatus) string {
	return ticketStatusLabelMap[status]
}

func IsValidTicketStatus(status string) bool {
	for _, item := range TicketStatusValues {
		if string(item) == status {
			return true
		}
	}
	return false
}

type TicketPriority int

const (
	TicketPriorityLow    TicketPriority = 1
	TicketPriorityNormal TicketPriority = 2
	TicketPriorityHigh   TicketPriority = 3
	TicketPriorityUrgent TicketPriority = 4
)

var TicketPriorityValues = []TicketPriority{
	TicketPriorityLow,
	TicketPriorityNormal,
	TicketPriorityHigh,
	TicketPriorityUrgent,
}

var ticketPriorityLabelMap = map[TicketPriority]string{
	TicketPriorityLow:    "低",
	TicketPriorityNormal: "普通",
	TicketPriorityHigh:   "高",
	TicketPriorityUrgent: "紧急",
}

func GetTicketPriorityLabel(priority TicketPriority) string {
	return ticketPriorityLabelMap[priority]
}

func IsValidTicketPriority(priority int) bool {
	for _, item := range TicketPriorityValues {
		if int(item) == priority {
			return true
		}
	}
	return false
}

type TicketSeverity int

const (
	TicketSeverityMinor    TicketSeverity = 1
	TicketSeverityMajor    TicketSeverity = 2
	TicketSeverityCritical TicketSeverity = 3
)

var TicketSeverityValues = []TicketSeverity{
	TicketSeverityMinor,
	TicketSeverityMajor,
	TicketSeverityCritical,
}

var ticketSeverityLabelMap = map[TicketSeverity]string{
	TicketSeverityMinor:    "轻微",
	TicketSeverityMajor:    "严重",
	TicketSeverityCritical: "致命",
}

func GetTicketSeverityLabel(severity TicketSeverity) string {
	return ticketSeverityLabelMap[severity]
}

func IsValidTicketSeverity(severity int) bool {
	for _, item := range TicketSeverityValues {
		if int(item) == severity {
			return true
		}
	}
	return false
}

type TicketSource string

const (
	TicketSourceManual       TicketSource = "manual"
	TicketSourceConversation TicketSource = "conversation"
	TicketSourcePortal       TicketSource = "portal"
	TicketSourceAPI          TicketSource = "api"
	TicketSourceRule         TicketSource = "rule"
)

var TicketSourceValues = []TicketSource{
	TicketSourceManual,
	TicketSourceConversation,
	TicketSourcePortal,
	TicketSourceAPI,
	TicketSourceRule,
}

func IsValidTicketSource(source string) bool {
	for _, item := range TicketSourceValues {
		if string(item) == source {
			return true
		}
	}
	return false
}

type TicketCommentType string

const (
	TicketCommentTypePublicReply  TicketCommentType = "public_reply"
	TicketCommentTypeInternalNote TicketCommentType = "internal_note"
	TicketCommentTypeSystemLog    TicketCommentType = "system_log"
)

type TicketEventType string

const (
	TicketEventTypeCreated            TicketEventType = "created"
	TicketEventTypeUpdated            TicketEventType = "updated"
	TicketEventTypeAssigned           TicketEventType = "assigned"
	TicketEventTypeTransferred        TicketEventType = "transferred"
	TicketEventTypeStatusChanged      TicketEventType = "status_changed"
	TicketEventTypeReplied            TicketEventType = "replied"
	TicketEventTypeInternalNoted      TicketEventType = "internal_noted"
	TicketEventTypeClosed             TicketEventType = "closed"
	TicketEventTypeReopened           TicketEventType = "reopened"
	TicketEventTypeSLABreached        TicketEventType = "sla_breached"
	TicketEventTypeLinkedConversation TicketEventType = "linked_conversation"
)

type TicketSLAType string

const (
	TicketSLATypeFirstResponse TicketSLAType = "first_response"
	TicketSLATypeResolution    TicketSLAType = "resolution"
)

type TicketSLAStatus string

const (
	TicketSLAStatusRunning   TicketSLAStatus = "running"
	TicketSLAStatusPaused    TicketSLAStatus = "paused"
	TicketSLAStatusCompleted TicketSLAStatus = "completed"
	TicketSLAStatusBreached  TicketSLAStatus = "breached"
)

type TicketRelationType string

const (
	TicketRelationTypeDuplicate TicketRelationType = "duplicate"
	TicketRelationTypeRelated   TicketRelationType = "related"
	TicketRelationTypeParent    TicketRelationType = "parent"
	TicketRelationTypeChild     TicketRelationType = "child"
)
