package constants

const (
	RoleCodeSuperAdmin = "super_admin"
	RoleCodeAdmin      = "admin"
	RoleCodeOperator   = "operator"
	RoleCodeViewer     = "viewer"
)

const (
	AccessTokenPrefix  = "atk_"
	RefreshTokenPrefix = "rtk_"
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

const (
	ClientTypeAdminWeb = "admin_web"
)

const (
	BootstrapAdminUsername = "admin"
	BootstrapAdminPassword = "ChangeMe123!"
	BootstrapAdminNickname = "超级管理员"
)

// Permission 权限结构体
type Permission struct {
	Name      string
	Code      string
	Type      string
	GroupName string
	Method    string
	APIPath   string
	SortNo    int
}

// 权限常量定义
var (
	// 用户相关权限
	PermissionUserView       = Permission{Name: "查看用户", Code: "user.view", Type: "api", GroupName: "user", Method: "ANY", APIPath: "/api/console/user/list", SortNo: 10}
	PermissionUserCreate     = Permission{Name: "创建用户", Code: "user.create", Type: "api", GroupName: "user", Method: "POST", APIPath: "/api/console/user/create", SortNo: 20}
	PermissionUserUpdate     = Permission{Name: "更新用户", Code: "user.update", Type: "api", GroupName: "user", Method: "POST", APIPath: "/api/console/user/update", SortNo: 30}
	PermissionUserDelete     = Permission{Name: "删除用户", Code: "user.delete", Type: "api", GroupName: "user", Method: "POST", APIPath: "/api/console/user/delete", SortNo: 40}
	PermissionUserAssignRole = Permission{Name: "分配用户角色", Code: "user.assignRole", Type: "api", GroupName: "user", Method: "POST", APIPath: "/api/console/user/assign_role", SortNo: 50}

	// 角色相关权限
	PermissionRoleView             = Permission{Name: "查看角色", Code: "role.view", Type: "api", GroupName: "role", Method: "ANY", APIPath: "/api/console/role/list", SortNo: 110}
	PermissionRoleCreate           = Permission{Name: "创建角色", Code: "role.create", Type: "api", GroupName: "role", Method: "POST", APIPath: "/api/console/role/create", SortNo: 120}
	PermissionRoleUpdate           = Permission{Name: "更新角色", Code: "role.update", Type: "api", GroupName: "role", Method: "POST", APIPath: "/api/console/role/update", SortNo: 130}
	PermissionRoleDelete           = Permission{Name: "删除角色", Code: "role.delete", Type: "api", GroupName: "role", Method: "POST", APIPath: "/api/console/role/delete", SortNo: 140}
	PermissionRoleAssignPermission = Permission{Name: "分配角色权限", Code: "role.assignPermission", Type: "api", GroupName: "role", Method: "POST", APIPath: "/api/console/role/assign_permission", SortNo: 150}

	// 权限相关权限
	PermissionPermissionView = Permission{Name: "查看权限", Code: "permission.view", Type: "api", GroupName: "permission", Method: "ANY", APIPath: "/api/console/permission/list", SortNo: 210}
	PermissionPermissionSync = Permission{Name: "同步权限", Code: "permission.sync", Type: "api", GroupName: "permission", Method: "POST", APIPath: "/api/console/permission/sync", SortNo: 220}

	// 会话相关权限
	PermissionSessionView   = Permission{Name: "查看会话", Code: "session.view", Type: "api", GroupName: "session", Method: "ANY", APIPath: "/api/console/session/list", SortNo: 310}
	PermissionSessionRevoke = Permission{Name: "踢除会话", Code: "session.revoke", Type: "api", GroupName: "session", Method: "POST", APIPath: "/api/console/session/revoke", SortNo: 320}

	// 客服会话相关权限
	PermissionConversationView     = Permission{Name: "查看会话", Code: "conversation.view", Type: "api", GroupName: "conversation", Method: "ANY", APIPath: "/api/console/conversation/list", SortNo: 410}
	PermissionConversationAssign   = Permission{Name: "分配会话", Code: "conversation.assign", Type: "api", GroupName: "conversation", Method: "POST", APIPath: "/api/console/conversation/assign", SortNo: 430}
	PermissionConversationTransfer = Permission{Name: "转接会话", Code: "conversation.transfer", Type: "api", GroupName: "conversation", Method: "POST", APIPath: "/api/console/conversation/transfer", SortNo: 440}
	PermissionConversationClose    = Permission{Name: "关闭会话", Code: "conversation.close", Type: "api", GroupName: "conversation", Method: "POST", APIPath: "/api/console/conversation/close", SortNo: 450}
	PermissionConversationSend     = Permission{Name: "发送会话消息", Code: "conversation.send", Type: "api", GroupName: "conversation", Method: "POST", APIPath: "/api/console/conversation/send_message", SortNo: 460}
	PermissionConversationTag      = Permission{Name: "管理会话标签", Code: "conversation.tag", Type: "api", GroupName: "conversation", Method: "POST", APIPath: "/api/console/conversation/add_tag", SortNo: 470}
	PermissionConversationHandover = Permission{Name: "处理会话交接", Code: "conversation.handover", Type: "api", GroupName: "conversation", Method: "ANY", APIPath: "/api/console/conversation/handover_list", SortNo: 480}
	PermissionConversationRecycle  = Permission{Name: "回收会话", Code: "conversation.recycle", Type: "api", GroupName: "conversation", Method: "POST", APIPath: "/api/console/conversation/recycle", SortNo: 490}

	// 快捷回复相关权限
	PermissionQuickReplyView   = Permission{Name: "查看快捷回复", Code: "quickReply.view", Type: "api", GroupName: "quickReply", Method: "ANY", APIPath: "/api/console/quick-reply/list", SortNo: 510}
	PermissionQuickReplyCreate = Permission{Name: "创建快捷回复", Code: "quickReply.create", Type: "api", GroupName: "quickReply", Method: "POST", APIPath: "/api/console/quick-reply/create", SortNo: 520}
	PermissionQuickReplyUpdate = Permission{Name: "更新快捷回复", Code: "quickReply.update", Type: "api", GroupName: "quickReply", Method: "POST", APIPath: "/api/console/quick-reply/update", SortNo: 530}
	PermissionQuickReplyDelete = Permission{Name: "删除快捷回复", Code: "quickReply.delete", Type: "api", GroupName: "quickReply", Method: "POST", APIPath: "/api/console/quick-reply/delete", SortNo: 540}

	// 标签相关权限
	PermissionTagView   = Permission{Name: "查看标签", Code: "tag.view", Type: "api", GroupName: "tag", Method: "ANY", APIPath: "/api/console/tag/list", SortNo: 550}
	PermissionTagCreate = Permission{Name: "创建标签", Code: "tag.create", Type: "api", GroupName: "tag", Method: "POST", APIPath: "/api/console/tag/create", SortNo: 560}
	PermissionTagUpdate = Permission{Name: "更新标签", Code: "tag.update", Type: "api", GroupName: "tag", Method: "POST", APIPath: "/api/console/tag/update", SortNo: 570}
	PermissionTagDelete = Permission{Name: "删除标签", Code: "tag.delete", Type: "api", GroupName: "tag", Method: "POST", APIPath: "/api/console/tag/delete", SortNo: 580}

	// 客服相关权限
	PermissionAgentView         = Permission{Name: "查看客服", Code: "agent.view", Type: "api", GroupName: "agent", Method: "ANY", APIPath: "/api/console/agent/list", SortNo: 610}
	PermissionAgentCreate       = Permission{Name: "创建客服", Code: "agent.create", Type: "api", GroupName: "agent", Method: "POST", APIPath: "/api/console/agent/create", SortNo: 620}
	PermissionAgentUpdate       = Permission{Name: "更新客服", Code: "agent.update", Type: "api", GroupName: "agent", Method: "POST", APIPath: "/api/console/agent/update", SortNo: 630}
	PermissionAgentDelete       = Permission{Name: "删除客服", Code: "agent.delete", Type: "api", GroupName: "agent", Method: "POST", APIPath: "/api/console/agent/delete", SortNo: 640}
	PermissionAgentUpdateStatus = Permission{Name: "更新客服状态", Code: "agent.updateStatus", Type: "api", GroupName: "agent", Method: "POST", APIPath: "/api/console/agent/update_status", SortNo: 650}
	PermissionAgentConfig       = Permission{Name: "配置客服服务规则", Code: "agent.config", Type: "api", GroupName: "agent", Method: "POST", APIPath: "/api/console/agent/update_service_config", SortNo: 660}

	// 客服组相关权限
	PermissionAgentTeamView   = Permission{Name: "查看客服组", Code: "agentTeam.view", Type: "api", GroupName: "agentTeam", Method: "ANY", APIPath: "/api/console/agent-team/list", SortNo: 710}
	PermissionAgentTeamCreate = Permission{Name: "创建客服组", Code: "agentTeam.create", Type: "api", GroupName: "agentTeam", Method: "POST", APIPath: "/api/console/agent-team/create", SortNo: 720}
	PermissionAgentTeamUpdate = Permission{Name: "更新客服组", Code: "agentTeam.update", Type: "api", GroupName: "agentTeam", Method: "POST", APIPath: "/api/console/agent-team/update", SortNo: 730}
	PermissionAgentTeamDelete = Permission{Name: "删除客服组", Code: "agentTeam.delete", Type: "api", GroupName: "agentTeam", Method: "POST", APIPath: "/api/console/agent-team/delete", SortNo: 740}

	// 客服组排班相关权限
	PermissionAgentTeamScheduleView          = Permission{Name: "查看客服组排班", Code: "agentTeamSchedule.view", Type: "api", GroupName: "agentTeamSchedule", Method: "ANY", APIPath: "/api/console/agent-team-schedule/list", SortNo: 810}
	PermissionAgentTeamScheduleCreate        = Permission{Name: "创建客服组排班", Code: "agentTeamSchedule.create", Type: "api", GroupName: "agentTeamSchedule", Method: "POST", APIPath: "/api/console/agent-team-schedule/create", SortNo: 820}
	PermissionAgentTeamScheduleUpdate        = Permission{Name: "更新客服组排班", Code: "agentTeamSchedule.update", Type: "api", GroupName: "agentTeamSchedule", Method: "POST", APIPath: "/api/console/agent-team-schedule/update", SortNo: 830}
	PermissionAgentTeamScheduleDelete        = Permission{Name: "删除客服组排班", Code: "agentTeamSchedule.delete", Type: "api", GroupName: "agentTeamSchedule", Method: "POST", APIPath: "/api/console/agent-team-schedule/delete", SortNo: 840}
	PermissionAgentTeamScheduleBatchGenerate = Permission{Name: "批量生成客服组排班", Code: "agentTeamSchedule.batchGenerate", Type: "api", GroupName: "agentTeamSchedule", Method: "POST", APIPath: "/api/console/agent-team-schedule/batch_generate", SortNo: 850}

	// 工单分类相关权限
	PermissionTicketCategoryView   = Permission{Name: "查看工单分类", Code: "ticketCategory.view", Type: "api", GroupName: "ticketCategory", Method: "ANY", APIPath: "/api/console/ticket-category/list", SortNo: 910}
	PermissionTicketCategoryCreate = Permission{Name: "创建工单分类", Code: "ticketCategory.create", Type: "api", GroupName: "ticketCategory", Method: "POST", APIPath: "/api/console/ticket-category/create", SortNo: 920}
	PermissionTicketCategoryUpdate = Permission{Name: "更新工单分类", Code: "ticketCategory.update", Type: "api", GroupName: "ticketCategory", Method: "POST", APIPath: "/api/console/ticket-category/update", SortNo: 930}
	PermissionTicketCategoryDelete = Permission{Name: "删除工单分类", Code: "ticketCategory.delete", Type: "api", GroupName: "ticketCategory", Method: "POST", APIPath: "/api/console/ticket-category/delete", SortNo: 940}

	// 工单相关权限
	PermissionTicketView   = Permission{Name: "查看工单", Code: "ticket.view", Type: "api", GroupName: "ticket", Method: "ANY", APIPath: "/api/console/ticket/list", SortNo: 1010}
	PermissionTicketCreate = Permission{Name: "创建工单", Code: "ticket.create", Type: "api", GroupName: "ticket", Method: "POST", APIPath: "/api/console/ticket/create", SortNo: 1020}
	PermissionTicketUpdate = Permission{Name: "更新工单", Code: "ticket.update", Type: "api", GroupName: "ticket", Method: "POST", APIPath: "/api/console/ticket/update", SortNo: 1030}
	PermissionTicketDelete = Permission{Name: "删除工单", Code: "ticket.delete", Type: "api", GroupName: "ticket", Method: "POST", APIPath: "/api/console/ticket/delete", SortNo: 1040}
	PermissionTicketAssign = Permission{Name: "分配工单", Code: "ticket.assign", Type: "api", GroupName: "ticket", Method: "POST", APIPath: "/api/console/ticket/assign", SortNo: 1050}
	PermissionTicketClose  = Permission{Name: "关闭工单", Code: "ticket.close", Type: "api", GroupName: "ticket", Method: "POST", APIPath: "/api/console/ticket/close", SortNo: 1060}
	PermissionTicketReopen = Permission{Name: "重开工单", Code: "ticket.reopen", Type: "api", GroupName: "ticket", Method: "POST", APIPath: "/api/console/ticket/reopen", SortNo: 1070}

	// 工单回复相关权限
	PermissionTicketReplyView   = Permission{Name: "查看工单回复", Code: "ticketReply.view", Type: "api", GroupName: "ticketReply", Method: "ANY", APIPath: "/api/console/ticket-reply/list", SortNo: 1110}
	PermissionTicketReplyCreate = Permission{Name: "创建工单回复", Code: "ticketReply.create", Type: "api", GroupName: "ticketReply", Method: "POST", APIPath: "/api/console/ticket-reply/create", SortNo: 1120}
	PermissionTicketReplyUpdate = Permission{Name: "更新工单回复", Code: "ticketReply.update", Type: "api", GroupName: "ticketReply", Method: "POST", APIPath: "/api/console/ticket-reply/update", SortNo: 1130}
	PermissionTicketReplyDelete = Permission{Name: "删除工单回复", Code: "ticketReply.delete", Type: "api", GroupName: "ticketReply", Method: "POST", APIPath: "/api/console/ticket-reply/delete", SortNo: 1140}

	// 文件资源相关权限
	PermissionAssetView   = Permission{Name: "查看文件资源", Code: "asset.view", Type: "api", GroupName: "asset", Method: "ANY", APIPath: "/api/console/asset/list", SortNo: 1210}
	PermissionAssetCreate = Permission{Name: "上传文件资源", Code: "asset.create", Type: "api", GroupName: "asset", Method: "POST", APIPath: "/api/console/asset/create", SortNo: 1220}
	PermissionAssetDelete = Permission{Name: "删除文件资源", Code: "asset.delete", Type: "api", GroupName: "asset", Method: "POST", APIPath: "/api/console/asset/delete", SortNo: 1230}

	// AI 配置相关权限
	PermissionAIConfigView   = Permission{Name: "查看 AI 配置", Code: "aiConfig.view", Type: "api", GroupName: "aiConfig", Method: "ANY", APIPath: "/api/console/ai-config/list", SortNo: 1390}
	PermissionAIConfigCreate = Permission{Name: "创建 AI 配置", Code: "aiConfig.create", Type: "api", GroupName: "aiConfig", Method: "POST", APIPath: "/api/console/ai-config/create", SortNo: 1400}
	PermissionAIConfigUpdate = Permission{Name: "更新 AI 配置", Code: "aiConfig.update", Type: "api", GroupName: "aiConfig", Method: "POST", APIPath: "/api/console/ai-config/update", SortNo: 1410}
	PermissionAIConfigDelete = Permission{Name: "删除 AI 配置", Code: "aiConfig.delete", Type: "api", GroupName: "aiConfig", Method: "POST", APIPath: "/api/console/ai-config/delete", SortNo: 1420}

	// 知识库相关权限
	PermissionKnowledgeBaseView   = Permission{Name: "查看知识库", Code: "knowledgeBase.view", Type: "api", GroupName: "knowledgeBase", Method: "ANY", APIPath: "/api/console/knowledge-base/list", SortNo: 1410}
	PermissionKnowledgeBaseCreate = Permission{Name: "创建知识库", Code: "knowledgeBase.create", Type: "api", GroupName: "knowledgeBase", Method: "POST", APIPath: "/api/console/knowledge-base/create", SortNo: 1420}
	PermissionKnowledgeBaseUpdate = Permission{Name: "更新知识库", Code: "knowledgeBase.update", Type: "api", GroupName: "knowledgeBase", Method: "POST", APIPath: "/api/console/knowledge-base/update", SortNo: 1430}
	PermissionKnowledgeBaseDelete = Permission{Name: "删除知识库", Code: "knowledgeBase.delete", Type: "api", GroupName: "knowledgeBase", Method: "POST", APIPath: "/api/console/knowledge-base/delete", SortNo: 1440}

	// 知识文档相关权限
	PermissionKnowledgeDocumentView   = Permission{Name: "查看知识文档", Code: "knowledgeDocument.view", Type: "api", GroupName: "knowledgeDocument", Method: "ANY", APIPath: "/api/console/knowledge-document/list", SortNo: 1510}
	PermissionKnowledgeDocumentCreate = Permission{Name: "创建知识文档", Code: "knowledgeDocument.create", Type: "api", GroupName: "knowledgeDocument", Method: "POST", APIPath: "/api/console/knowledge-document/create", SortNo: 1520}
	PermissionKnowledgeDocumentUpdate = Permission{Name: "更新知识文档", Code: "knowledgeDocument.update", Type: "api", GroupName: "knowledgeDocument", Method: "POST", APIPath: "/api/console/knowledge-document/update", SortNo: 1530}
	PermissionKnowledgeDocumentDelete = Permission{Name: "删除知识文档", Code: "knowledgeDocument.delete", Type: "api", GroupName: "knowledgeDocument", Method: "POST", APIPath: "/api/console/knowledge-document/delete", SortNo: 1540}

	// Skill 定义相关权限
	PermissionSkillDefinitionView   = Permission{Name: "查看技能定义", Code: "skillDefinition.view", Type: "api", GroupName: "skillDefinition", Method: "ANY", APIPath: "/api/console/skill-definition/list", SortNo: 1610}
	PermissionSkillDefinitionCreate = Permission{Name: "创建技能定义", Code: "skillDefinition.create", Type: "api", GroupName: "skillDefinition", Method: "POST", APIPath: "/api/console/skill-definition/create", SortNo: 1620}
	PermissionSkillDefinitionUpdate = Permission{Name: "更新技能定义", Code: "skillDefinition.update", Type: "api", GroupName: "skillDefinition", Method: "POST", APIPath: "/api/console/skill-definition/update", SortNo: 1630}
	PermissionSkillDefinitionDelete = Permission{Name: "删除技能定义", Code: "skillDefinition.delete", Type: "api", GroupName: "skillDefinition", Method: "POST", APIPath: "/api/console/skill-definition/delete", SortNo: 1640}

	// MCP 调试相关权限
	PermissionMCPView = Permission{Name: "查看MCP调试信息", Code: "mcp.view", Type: "api", GroupName: "mcp", Method: "POST", APIPath: "/api/console/mcp/list_tools", SortNo: 1710}
	PermissionMCPCall = Permission{Name: "调用MCP工具", Code: "mcp.call", Type: "api", GroupName: "mcp", Method: "POST", APIPath: "/api/console/mcp/call_tool", SortNo: 1720}
)

// Permissions 内置权限列表
var Permissions = []Permission{
	PermissionUserView,
	PermissionUserCreate,
	PermissionUserUpdate,
	PermissionUserDelete,
	PermissionUserAssignRole,
	PermissionRoleView,
	PermissionRoleCreate,
	PermissionRoleUpdate,
	PermissionRoleDelete,
	PermissionRoleAssignPermission,
	PermissionPermissionView,
	PermissionPermissionSync,
	PermissionSessionView,
	PermissionSessionRevoke,
	PermissionConversationView,
	PermissionConversationAssign,
	PermissionConversationTransfer,
	PermissionConversationClose,
	PermissionConversationSend,
	PermissionConversationTag,
	PermissionConversationHandover,
	PermissionConversationRecycle,
	PermissionQuickReplyView,
	PermissionQuickReplyCreate,
	PermissionQuickReplyUpdate,
	PermissionQuickReplyDelete,
	PermissionTagView,
	PermissionTagCreate,
	PermissionTagUpdate,
	PermissionTagDelete,
	PermissionAgentView,
	PermissionAgentCreate,
	PermissionAgentUpdate,
	PermissionAgentDelete,
	PermissionAgentUpdateStatus,
	PermissionAgentConfig,
	PermissionAgentTeamView,
	PermissionAgentTeamCreate,
	PermissionAgentTeamUpdate,
	PermissionAgentTeamDelete,
	PermissionAgentTeamScheduleView,
	PermissionAgentTeamScheduleCreate,
	PermissionAgentTeamScheduleUpdate,
	PermissionAgentTeamScheduleDelete,
	PermissionAgentTeamScheduleBatchGenerate,
	PermissionTicketCategoryView,
	PermissionTicketCategoryCreate,
	PermissionTicketCategoryUpdate,
	PermissionTicketCategoryDelete,
	PermissionTicketView,
	PermissionTicketCreate,
	PermissionTicketUpdate,
	PermissionTicketDelete,
	PermissionTicketAssign,
	PermissionTicketClose,
	PermissionTicketReopen,
	PermissionTicketReplyView,
	PermissionTicketReplyCreate,
	PermissionTicketReplyUpdate,
	PermissionTicketReplyDelete,
	PermissionAssetView,
	PermissionAssetCreate,
	PermissionAssetDelete,
	PermissionAIConfigView,
	PermissionAIConfigCreate,
	PermissionAIConfigUpdate,
	PermissionAIConfigDelete,
	PermissionKnowledgeBaseView,
	PermissionKnowledgeBaseCreate,
	PermissionKnowledgeBaseUpdate,
	PermissionKnowledgeBaseDelete,
	PermissionKnowledgeDocumentView,
	PermissionKnowledgeDocumentCreate,
	PermissionKnowledgeDocumentUpdate,
	PermissionKnowledgeDocumentDelete,
	PermissionSkillDefinitionView,
	PermissionSkillDefinitionCreate,
	PermissionSkillDefinitionUpdate,
	PermissionSkillDefinitionDelete,
	PermissionMCPView,
	PermissionMCPCall,
}

// PermissionMap 权限映射，用于通过 Code 查找 Permission
var PermissionMap = make(map[string]Permission)

// init 初始化 PermissionMap
func init() {
	for _, permission := range Permissions {
		PermissionMap[permission.Code] = permission
	}
}

type RoleSpec struct {
	Name   string
	Code   string
	SortNo int
}

var Roles = []RoleSpec{
	{Name: "超级管理员", Code: RoleCodeSuperAdmin, SortNo: 10},
	{Name: "管理员", Code: RoleCodeAdmin, SortNo: 20},
	{Name: "客服主管", Code: RoleCodeOperator, SortNo: 30},
	{Name: "只读成员", Code: RoleCodeViewer, SortNo: 40},
}

var RolePermissions = map[string][]Permission{
	RoleCodeSuperAdmin: Permissions,
	RoleCodeAdmin: {
		PermissionUserView, PermissionUserCreate, PermissionUserUpdate, PermissionUserAssignRole,
		PermissionRoleView, PermissionRoleCreate, PermissionRoleUpdate, PermissionRoleAssignPermission,
		PermissionPermissionView, PermissionPermissionSync,
		PermissionSessionView, PermissionSessionRevoke,
		PermissionConversationView, PermissionConversationAssign, PermissionConversationTransfer, PermissionConversationClose, PermissionConversationSend, PermissionConversationTag, PermissionConversationHandover, PermissionConversationRecycle,
		PermissionQuickReplyView, PermissionQuickReplyCreate, PermissionQuickReplyUpdate, PermissionQuickReplyDelete,
		PermissionTagView, PermissionTagCreate, PermissionTagUpdate, PermissionTagDelete,
		PermissionAgentView, PermissionAgentCreate, PermissionAgentUpdate, PermissionAgentDelete, PermissionAgentUpdateStatus, PermissionAgentConfig,
		PermissionAgentTeamView, PermissionAgentTeamCreate, PermissionAgentTeamUpdate, PermissionAgentTeamDelete,
		PermissionAgentTeamScheduleView, PermissionAgentTeamScheduleCreate, PermissionAgentTeamScheduleUpdate, PermissionAgentTeamScheduleDelete, PermissionAgentTeamScheduleBatchGenerate,
		PermissionTicketCategoryView, PermissionTicketCategoryCreate, PermissionTicketCategoryUpdate, PermissionTicketCategoryDelete,
		PermissionTicketView, PermissionTicketCreate, PermissionTicketUpdate, PermissionTicketDelete, PermissionTicketAssign, PermissionTicketClose, PermissionTicketReopen,
		PermissionTicketReplyView, PermissionTicketReplyCreate, PermissionTicketReplyUpdate, PermissionTicketReplyDelete,
		PermissionAssetView, PermissionAssetCreate, PermissionAssetDelete,
		PermissionAIConfigView, PermissionAIConfigCreate, PermissionAIConfigUpdate, PermissionAIConfigDelete,
		PermissionSkillDefinitionView, PermissionSkillDefinitionCreate, PermissionSkillDefinitionUpdate, PermissionSkillDefinitionDelete,
	},
	RoleCodeOperator: {
		PermissionUserView,
		PermissionRoleView,
		PermissionPermissionView,
		PermissionSessionView,
		PermissionConversationView, PermissionConversationClose, PermissionConversationSend, PermissionConversationTag, PermissionConversationHandover, PermissionConversationRecycle,
		PermissionQuickReplyView, PermissionQuickReplyCreate, PermissionQuickReplyUpdate, PermissionQuickReplyDelete,
		PermissionTagView, PermissionTagCreate, PermissionTagUpdate, PermissionTagDelete,
		PermissionAgentView, PermissionAgentUpdate,
		PermissionAgentTeamView,
		PermissionAgentTeamScheduleView, PermissionAgentTeamScheduleCreate, PermissionAgentTeamScheduleUpdate, PermissionAgentTeamScheduleDelete, PermissionAgentTeamScheduleBatchGenerate,
		PermissionTicketCategoryView, PermissionTicketCategoryCreate, PermissionTicketCategoryUpdate, PermissionTicketCategoryDelete,
		PermissionTicketView, PermissionTicketCreate, PermissionTicketUpdate, PermissionTicketDelete, PermissionTicketAssign, PermissionTicketClose, PermissionTicketReopen,
		PermissionTicketReplyView, PermissionTicketReplyCreate, PermissionTicketReplyUpdate, PermissionTicketReplyDelete,
		PermissionAssetView, PermissionAssetCreate, PermissionAssetDelete,
		PermissionAIConfigView,
		PermissionSkillDefinitionView, PermissionSkillDefinitionCreate, PermissionSkillDefinitionUpdate,
	},
	RoleCodeViewer: {
		PermissionUserView,
		PermissionRoleView,
		PermissionPermissionView,
		PermissionConversationView,
		PermissionQuickReplyView,
		PermissionTagView,
		PermissionAssetView,
		PermissionAgentView,
		PermissionAgentTeamView,
		PermissionAgentTeamScheduleView,
		PermissionTicketCategoryView,
		PermissionTicketView,
		PermissionTicketReplyView,
		PermissionAIConfigView,
		PermissionSkillDefinitionView,
	},
}

func PermissionCodes() []string {
	ret := make([]string, 0, len(Permissions))
	for _, permission := range Permissions {
		ret = append(ret, permission.Code)
	}
	return ret
}
