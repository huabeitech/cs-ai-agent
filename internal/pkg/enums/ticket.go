package enums

type TicketStatus int

const (
	TicketStatusPending    TicketStatus = 1
	TicketStatusProcessing TicketStatus = 2
	TicketStatusWaiting    TicketStatus = 3
	TicketStatusResolved   TicketStatus = 4
	TicketStatusClosed     TicketStatus = 5
	TicketStatusCancelled  TicketStatus = 6
)

var ticketStatusLabelMap = map[TicketStatus]string{
	TicketStatusPending:    "待处理",
	TicketStatusProcessing: "处理中",
	TicketStatusWaiting:    "待确认",
	TicketStatusResolved:   "已解决",
	TicketStatusClosed:     "已关闭",
	TicketStatusCancelled:  "已取消",
}

func GetTicketStatusLabel(status TicketStatus) string {
	return ticketStatusLabelMap[status]
}

type TicketPriority int

const (
	TicketPriorityNormal TicketPriority = 0
	TicketPriorityLow    TicketPriority = 1
	TicketPriorityMedium TicketPriority = 2
	TicketPriorityHigh   TicketPriority = 3
	TicketPriorityUrgent TicketPriority = 4
)

var ticketPriorityLabelMap = map[TicketPriority]string{
	TicketPriorityNormal: "普通",
	TicketPriorityLow:    "低",
	TicketPriorityMedium: "中",
	TicketPriorityHigh:   "高",
	TicketPriorityUrgent: "紧急",
}

func GetTicketPriorityLabel(priority TicketPriority) string {
	return ticketPriorityLabelMap[priority]
}

type TicketSatisfied int

const (
	TicketSatisfiedNone TicketSatisfied = 0
	TicketSatisfiedYes  TicketSatisfied = 1
	TicketSatisfiedNo   TicketSatisfied = 2
)

var ticketSatisfiedLabelMap = map[TicketSatisfied]string{
	TicketSatisfiedNone: "未评价",
	TicketSatisfiedYes:  "满意",
	TicketSatisfiedNo:   "不满意",
}

func GetTicketSatisfiedLabel(satisfied TicketSatisfied) string {
	return ticketSatisfiedLabelMap[satisfied]
}

type TicketChannelType string

const (
	TicketChannelConversation TicketChannelType = "conversation"
	TicketChannelSelfService  TicketChannelType = "self_service"
	TicketChannelAdmin        TicketChannelType = "admin"
	TicketChannelAPI          TicketChannelType = "api"
)

var ticketChannelTypeLabelMap = map[TicketChannelType]string{
	TicketChannelConversation: "客服会话转化",
	TicketChannelSelfService:  "自助提交",
	TicketChannelAdmin:        "后台创建",
	TicketChannelAPI:          "API接口",
}

func GetTicketChannelTypeLabel(channelType TicketChannelType) string {
	return ticketChannelTypeLabelMap[channelType]
}

type TicketAssignType string

const (
	TicketAssignTypeCreate     TicketAssignType = "create"
	TicketAssignTypeTransfer   TicketAssignType = "transfer"
	TicketAssignTypeClaim      TicketAssignType = "claim"
	TicketAssignTypeWithdraw   TicketAssignType = "withdraw"
	TicketAssignTypeDistribute TicketAssignType = "distribute"
)

var ticketAssignTypeLabelMap = map[TicketAssignType]string{
	TicketAssignTypeCreate:     "创建时分配",
	TicketAssignTypeTransfer:   "转接",
	TicketAssignTypeClaim:      "领取",
	TicketAssignTypeWithdraw:   "退回",
	TicketAssignTypeDistribute: "分配",
}

func GetTicketAssignTypeLabel(assignType TicketAssignType) string {
	return ticketAssignTypeLabelMap[assignType]
}

type TicketAssignStatus int

const (
	TicketAssignStatusProcessing TicketAssignStatus = 1
	TicketAssignStatusAccepted   TicketAssignStatus = 2
	TicketAssignStatusRejected   TicketAssignStatus = 3
	TicketAssignStatusCancelled  TicketAssignStatus = 4
)

var ticketAssignStatusLabelMap = map[TicketAssignStatus]string{
	TicketAssignStatusProcessing: "进行中",
	TicketAssignStatusAccepted:   "已接受",
	TicketAssignStatusRejected:   "已退回",
	TicketAssignStatusCancelled:  "已取消",
}

func GetTicketAssignStatusLabel(status TicketAssignStatus) string {
	return ticketAssignStatusLabelMap[status]
}

type TicketSenderType string

const (
	TicketSenderTypeAgent    TicketSenderType = "agent"
	TicketSenderTypeCustomer TicketSenderType = "customer"
	TicketSenderTypeSystem   TicketSenderType = "system"
)

var ticketSenderTypeLabelMap = map[TicketSenderType]string{
	TicketSenderTypeAgent:    "客服",
	TicketSenderTypeCustomer: "客户",
	TicketSenderTypeSystem:   "系统",
}

func GetTicketSenderTypeLabel(senderType TicketSenderType) string {
	return ticketSenderTypeLabelMap[senderType]
}

type TicketReplySendStatus int

const (
	TicketReplySendStatusPending TicketReplySendStatus = 1
	TicketReplySendStatusSent    TicketReplySendStatus = 2
	TicketReplySendStatusFailed  TicketReplySendStatus = 3
)

var ticketReplySendStatusLabelMap = map[TicketReplySendStatus]string{
	TicketReplySendStatusPending: "待发送",
	TicketReplySendStatusSent:    "已发送",
	TicketReplySendStatusFailed:  "发送失败",
}

func GetTicketReplySendStatusLabel(status TicketReplySendStatus) string {
	return ticketReplySendStatusLabelMap[status]
}
