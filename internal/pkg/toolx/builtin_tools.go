package toolx

const (
	BuiltinToolCatalogServerCode              = "builtin"
	BuiltinToolSearchToolCode                 = "builtin/tool_search"
	BuiltinToolSearchToolName                 = "tool_search"
	BuiltinToolSearchToolTitle                = "搜索并调用动态工具"
	BuiltinToolSearchToolDescription          = "用于搜索当前允许使用的 MCP 工具，并在确认目标 toolCode 后动态调用该工具。适合处理长尾工具，不应替代固定内置流程工具。"
	GraphToolCatalogServerCode                = "graph"
	GraphCreateTicketConfirmToolCode          = "graph/create_ticket_with_confirmation"
	GraphCreateTicketConfirmToolName          = "create_ticket_with_confirmation"
	GraphCreateTicketConfirmToolTitle         = "创建工单确认流程"
	GraphCreateTicketConfirmToolDescription   = "Graph Tool。用于封装建单参数整理、用户确认、真正建单和结果返回的确定性流程。"
	GraphHandoffConversationToolCode          = "graph/handoff_to_human"
	GraphHandoffConversationToolName          = "handoff_to_human"
	GraphHandoffConversationToolTitle         = "转人工确认流程"
	GraphHandoffConversationToolDescription   = "Graph Tool。用于封装转人工原因整理、用户确认、真正转人工和结果返回的确定性流程。"
	BuiltinCreateTicketConfirmToolCode        = "builtin/create_ticket_with_confirmation"
	BuiltinCreateTicketConfirmToolName        = "create_ticket_with_confirmation"
	BuiltinCreateTicketConfirmToolTitle       = "创建工单并发起确认"
	BuiltinCreateTicketConfirmToolDescription = "当用户明确要求创建工单，且标题和描述已经整理清楚后调用。工具会先向用户确认，确认后才真正创建工单。"
)

func IsAutoInjectedToolCode(toolCode string) bool {
	return toolCode == BuiltinToolSearchToolCode
}

func NormalizeToolCodeAlias(toolCode string) string {
	switch toolCode {
	case BuiltinCreateTicketConfirmToolCode:
		return GraphCreateTicketConfirmToolCode
	default:
		return toolCode
	}
}
