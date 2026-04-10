package toolx

const (
	BuiltinToolCatalogServerCode              = "builtin"
	BuiltinToolSearchToolCode                 = "builtin/tool_search"
	BuiltinToolSearchToolName                 = "tool_search"
	BuiltinToolSearchToolTitle                = "搜索并调用动态工具"
	BuiltinToolSearchToolDescription          = "用于搜索当前允许使用的 MCP 工具，并在确认目标 toolCode 后动态调用该工具。适合处理长尾工具，不应替代固定内置流程工具。"
	BuiltinCreateTicketConfirmToolCode        = "builtin/create_ticket_with_confirmation"
	BuiltinCreateTicketConfirmToolName        = "create_ticket_with_confirmation"
	BuiltinCreateTicketConfirmToolTitle       = "创建工单并发起确认"
	BuiltinCreateTicketConfirmToolDescription = "当用户明确要求创建工单，且标题和描述已经整理清楚后调用。工具会先向用户确认，确认后才真正创建工单。"
)
