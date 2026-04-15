import { readSession } from "@/lib/auth"
import { request } from "@/lib/api/client"
import { createWebSocketBaseUrl } from "@/lib/api/websocket"

export type Paging = {
  page: number
  limit: number
  total: number
}

export type PageResult<T> = {
  results: T[]
  page: Paging
}

export type CursorResult<T> = {
  results: T[]
  cursor: string
  hasMore: boolean
}

export type AdminUser = {
  id: number
  username: string
  nickname: string
  avatar: string
  mobile?: string
  email?: string
  status: number
  isSystem: boolean
  lastLoginAt?: string
  lastLoginIp?: string
  roles?: AdminRole[]
  permissions?: string[]
}

export type UpdateAdminUserPayload = {
  id: number
  nickname: string
  avatar: string
  mobile: string | null
  email: string | null
  remark: string
}

export type CreateAdminUserPayload = {
  username: string
  nickname: string
  avatar: string
  mobile: string | null
  email: string | null
  remark: string
  roleIds: number[]
}

export type CreateUserResult = {
  user: AdminUser
  password: string
}

export type ResetPasswordResult = {
  password: string
}

export type AdminRole = {
  id: number
  name: string
  code: string
  status: number
  isSystem: boolean
  sortNo: number
  permissions?: string[]
}

export type AdminPermission = {
  id: number
  name: string
  code: string
  type: string
  groupName: string
  method: string
  apiPath: string
  status: number
  sortNo: number
}

export type ConversationTag = {
  id: number
  name: string
  color: string
}

export type ConversationParticipant = {
  id: number
  participantType: string
  participantId: number
  externalParticipantId?: string
  joinedAt?: string
  leftAt?: string
  status: number
}

export type AdminConversation = {
  id: number
  externalSource: string
  externalId: string
  subject: string
  status: number
  serviceMode: number
  priority: number
  currentAssigneeId: number
  currentAssigneeName?: string
  lastMessageId: number
  lastMessageAt?: string
  lastActiveAt?: string
  lastMessageSummary?: string
  customerUnreadCount: number
  agentUnreadCount: number
  customerLastReadMessageId: number
  customerLastReadSeqNo: number
  customerLastReadAt?: string
  agentLastReadMessageId: number
  agentLastReadSeqNo: number
  agentLastReadAt?: string
  closedAt?: string
  closedBy: number
  closedByName?: string
  closeReason?: string
  participants?: ConversationParticipant[]
}

export type AdminConversationDetail = AdminConversation & {
  participants?: ConversationParticipant[]
}

export type AdminMessage = {
  id: number
  conversationId: number
  clientMsgId?: string
  senderType: string
  senderId: number
  senderName?: string
  senderAvatar?: string
  messageType: string
  content: string
  payload?: string
  seqNo: number
  sendStatus: number
  sentAt?: string
  deliveredAt?: string
  readAt?: string
  customerRead: boolean
  customerReadAt?: string
  agentRead: boolean
  agentReadAt?: string
  recalledAt?: string
  quotedMessageId?: number
}

export type AdminQuickReply = {
  id: number
  groupName: string
  title: string
  content: string
  status: number
  sortNo: number
  createdBy: number
}

export type AdminChannel = {
  id: number
  channelType: string
  channelId: string
  aiAgentId: number
  aiAgentName?: string
  name: string
  configJson: string
  status: number
  remark: string
}

export type CreateAdminChannelPayload = {
  channelType: string
  aiAgentId: number
  name: string
  configJson: string
  status: number
  remark: string
}

export type UpdateAdminChannelPayload = CreateAdminChannelPayload & {
  id: number
}

export type AIAgent = {
  id: number
  name: string
  description: string
  status: number
  statusName: string
  aiConfigId: number
  aiConfigName?: string
  serviceMode: number
  serviceModeName: string
  systemPrompt: string
  welcomeMessage: string
  replyTimeoutSeconds: number
  teams: { id: number; name: string }[]
  handoffMode: number
  handoffModeName: string
  fallbackMessage: string
  knowledgeIds: number[]
  knowledgeBaseNames: string[]
  skillIds: number[]
  skills: { id: number; code: string; name: string }[]
  directTools: {
    toolCode: string
    serverCode: string
    toolName: string
    title: string
    description: string
    arguments?: Record<string, string>
  }[]
  graphTools: string[]
  sortNo: number
  remark: string
  createdAt: string
  updatedAt: string
  createUserName: string
  updateUserName: string
}

export type CreateAIAgentPayload = {
  name: string
  description: string
  aiConfigId: number
  serviceMode: number
  systemPrompt: string
  welcomeMessage: string
  replyTimeoutSeconds: number
  teamIds: number[]
  handoffMode: number
  fallbackMessage: string
  knowledgeIds: number[]
  skillIds: number[]
  directTools: {
    toolCode: string
    serverCode: string
    toolName: string
    title: string
    description: string
    arguments?: Record<string, string>
  }[]
  graphTools: string[]
  remark: string
}

export type UpdateAIAgentPayload = CreateAIAgentPayload & {
  id: number
}

export type CreateAdminQuickReplyPayload = {
  groupName: string
  title: string
  content: string
  status: number
  sortNo: number
}

export type UpdateAdminQuickReplyPayload = CreateAdminQuickReplyPayload & {
  id: number
}

export type SkillDefinition = {
  id: number
  code: string
  name: string
  description: string
  instruction: string
  examples: string[]
  toolWhitelist: string[]
  status: number
  statusName: string
  remark: string
  createdAt: string
  updatedAt: string
  createUserName: string
  updateUserName: string
}

export type CreateSkillDefinitionPayload = {
  code: string
  name: string
  description: string
  instruction: string
  examples: string[]
  toolWhitelist: string[]
  remark: string
}

export type UpdateSkillDefinitionPayload = CreateSkillDefinitionPayload & {
  id: number
}

export type SkillDebugRunPayload = {
  aiAgentId: number
  conversationId?: number
  skillCode: string
  userMessage: string
}

export type SkillDebugResumePayload = {
  aiAgentId: number
  conversationId?: number
  checkPointId: string
  userMessage: string
}

export type SkillDebugRunResult = {
  skillCode: string
  skillName: string
  replyText: string
  planReason: string
  skillRouteTrace: string
  toolWhitelist: string[]
  exposedToolCodes: string[]
  invokedToolCodes: string[]
  toolSearchTrace: string
  graphToolTrace: string
  graphToolCode: string
  interruptType: string
  checkPointId: string
  interrupted: boolean
  traceData: string
  errorMessage: string
  conversationId: number
  aiAgentId: number
}

export type MCPConnectionResult = {
  serverCode: string
  endpoint: string
  protocol: string
  serverName: string
  version: string
}

export type MCPServerInfo = {
  code: string
  enabled: boolean
  endpoint: string
  timeoutMs: number
}

export type MCPToolInfo = {
  name: string
  title: string
  description: string
  inputSchema: unknown
  outputSchema?: unknown
}

export type MCPToolSourceType = "mcp" | "graph" | "builtin"

export type MCPToolCatalogItem = {
  toolCode: string
  serverCode: string
  toolName: string
  sourceType: MCPToolSourceType
  autoInjected: boolean
  title: string
  description: string
  inputSchema: unknown
  outputSchema?: unknown
}

export type MCPToolResultContent = {
  type: string
  text?: string
  data?: unknown
}

export type MCPToolCallResult = {
  serverCode: string
  toolName: string
  isError: boolean
  content: MCPToolResultContent[]
  structuredContent?: unknown
}

export type AgentRunLog = {
  id: number
  conversationId: number
  messageId: number
  aiAgentId: number
  aiConfigId: number
  userMessage: string
  plannedAction: string
  plannedSkillCode: string
  plannedSkillName: string
  skillRouteTrace: string
  toolSearchTrace: string
  graphToolTrace: string
  graphToolCode: string
  recommendedAction: string
  riskLevel: string
  ticketDraftReady: boolean
  handoffReason: string
  plannedToolCode: string
  planReason: string
  interruptType: string
  resumeSource: string
  hitlStatus: string
  hitlStatusName: string
  hitlSummary: string
  finalAction: string
  finalStatus: string
  replyText: string
  errorMessage: string
  latencyMs: number
  traceData: string
  createdAt: string
}

export type AdminAgentProfile = {
  id: number
  userId: number
  teamId: number
  teamName?: string
  username?: string
  nickname?: string
  agentCode: string
  displayName: string
  avatar: string
  serviceStatus: number
  maxConcurrentCount: number
  priorityLevel: number
  autoAssignEnabled: boolean
  receiveOfflineMessage: boolean
  lastOnlineAt?: string
  lastStatusAt?: string
  remark: string
}

export type CreateAdminAgentProfilePayload = {
  userId: number
  teamId: number
  agentCode: string
  displayName: string
  avatar: string
  serviceStatus: number
  maxConcurrentCount: number
  priorityLevel: number
  autoAssignEnabled: boolean
  receiveOfflineMessage: boolean
  lastOnlineAt?: string
  lastStatusAt?: string
  remark: string
}

export type UpdateAdminAgentProfilePayload =
  CreateAdminAgentProfilePayload & {
    id: number
  }

export type AdminAgentTeam = {
  id: number
  name: string
  leaderUserId: number
  leaderUsername?: string
  leaderNickname?: string
  status: number
  description: string
  remark: string
}

export type CreateAdminAgentTeamPayload = {
  name: string
  leaderUserId: number
  status: number
  description: string
  remark: string
}

export type UpdateAdminAgentTeamPayload = CreateAdminAgentTeamPayload & {
  id: number
}

export type AdminAgentTeamSchedule = {
  id: number
  teamId: number
  teamName?: string
  startAt: string
  endAt: string
  sourceType: string
  remark: string
}

export type CreateAdminAgentTeamSchedulePayload = {
  teamId: number
  startAt: string
  endAt: string
  sourceType: string
  remark: string
}

export type UpdateAdminAgentTeamSchedulePayload =
  CreateAdminAgentTeamSchedulePayload & {
    id: number
  }

function toQueryString(query?: Record<string, string | number | undefined>) {
  if (!query) {
    return ""
  }

  const params = new URLSearchParams()
  Object.entries(query).forEach(([key, value]) => {
    if (value === undefined || value === "") {
      return
    }
    params.set(key, String(value))
  })
  const output = params.toString()
  return output ? `?${output}` : ""
}

export function createAdminWebSocketUrl() {
  const session = readSession()
  if (!session?.accessToken) {
    throw new Error("未登录或登录已过期")
  }

  const baseUrl = createWebSocketBaseUrl()
  const params = new URLSearchParams({
    accessToken: session.accessToken,
  })
  return `${baseUrl}/api/dashboard/ws?${params.toString()}`
}

export function fetchChannels(
  query?: Record<string, string | number | undefined>
) {
  return request<PageResult<AdminChannel>>(
    `/api/dashboard/channel/list${toQueryString(query)}`
  )
}

export function fetchChannel(id: number) {
  return request<AdminChannel>(`/api/dashboard/channel/${id}`)
}

export function createChannel(payload: CreateAdminChannelPayload) {
  return request<AdminChannel>("/api/dashboard/channel/create", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function updateChannel(payload: UpdateAdminChannelPayload) {
  return request<void>("/api/dashboard/channel/update", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function updateChannelStatus(id: number, status: number) {
  return request<void>("/api/dashboard/channel/update_status", {
    method: "POST",
    body: JSON.stringify({ id, status }),
  })
}

export function deleteChannel(id: number) {
  return request<void>("/api/dashboard/channel/delete", {
    method: "POST",
    body: JSON.stringify({ id }),
  })
}

export function fetchAIAgents(
  query?: Record<string, string | number | undefined>
) {
  return request<PageResult<AIAgent>>(
    `/api/dashboard/ai-agent/list${toQueryString(query)}`
  )
}

export function fetchAIAgentsAll(query?: Record<string, string | number | undefined>) {
  return request<AIAgent[]>(
    `/api/dashboard/ai-agent/list_all${toQueryString(query)}`
  )
}

export function fetchAIAgent(id: number) {
  return request<AIAgent>(`/api/dashboard/ai-agent/${id}`)
}

export function createAIAgent(payload: CreateAIAgentPayload) {
  return request<AIAgent>("/api/dashboard/ai-agent/create", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function updateAIAgent(payload: UpdateAIAgentPayload) {
  return request<void>("/api/dashboard/ai-agent/update", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function deleteAIAgent(id: number) {
  return request<void>("/api/dashboard/ai-agent/delete", {
    method: "POST",
    body: JSON.stringify({ id }),
  })
}

export function updateAIAgentSort(ids: number[]) {
  return request<void>("/api/dashboard/ai-agent/update_sort", {
    method: "POST",
    body: JSON.stringify(ids),
  })
}

export function updateAIAgentStatus(id: number, status: number) {
  return request<void>("/api/dashboard/ai-agent/update_status", {
    method: "POST",
    body: JSON.stringify({ id, status }),
  })
}

export function fetchUsers(query?: Record<string, string | number | undefined>) {
  return request<PageResult<AdminUser>>(
    `/api/dashboard/user/list${toQueryString(query)}`
  )
}

export function fetchUsersAll(query?: Record<string, string | number | undefined>) {
  return request<AdminUser[]>(
    `/api/dashboard/user/list_all${toQueryString(query)}`
  )
}

export function createUser(payload: CreateAdminUserPayload) {
  return request<CreateUserResult>("/api/dashboard/user/create", {
    method: "POST",
    body: JSON.stringify({
      username: payload.username,
      nickname: payload.nickname,
      avatar: payload.avatar,
      mobile: payload.mobile,
      email: payload.email,
      remark: payload.remark,
      roleIds: payload.roleIds,
    }),
  })
}

export function updateUser(payload: UpdateAdminUserPayload) {
  return request<void>("/api/dashboard/user/update", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function fetchUserDetail(id: number) {
  return request<AdminUser>(`/api/dashboard/user/${id}`)
}

export function updateUserStatus(id: number, status: number) {
  return request<void>("/api/dashboard/user/update_status", {
    method: "POST",
    body: JSON.stringify({ id, status }),
  })
}

export function resetUserPassword(userId: number) {
  return request<ResetPasswordResult>("/api/dashboard/user/reset_password", {
    method: "POST",
    body: JSON.stringify({ userId }),
  })
}

export function changeSelfPassword(password: string) {
  return request<void>("/api/dashboard/user/change_password", {
    method: "POST",
    body: JSON.stringify({ password }),
  })
}

export function assignUserRoles(userId: number, roleIds: number[]) {
  return request<void>("/api/dashboard/user/assign_role", {
    method: "POST",
    body: JSON.stringify({ userId, roleIds }),
  })
}

export function fetchRoles(query?: Record<string, string | number | undefined>) {
  return request<PageResult<AdminRole>>(
    `/api/dashboard/role/list${toQueryString(query)}`
  )
}

export function fetchRoleListAll() {
  return request<AdminRole[]>("/api/dashboard/role/list_all")
}

export function fetchRoleDetail(id: number) {
  return request<AdminRole>(`/api/dashboard/role/${id}`)
}

export function assignRolePermissions(roleId: number, permissionIds: number[]) {
  return request<void>("/api/dashboard/role/assign_permission", {
    method: "POST",
    body: JSON.stringify({ roleId, permissionIds }),
  })
}

export function updateRoleSort(ids: number[]) {
  return request<void>("/api/dashboard/role/update_sort", {
    method: "POST",
    body: JSON.stringify(ids),
  })
}

export function fetchPermissions(
  query?: Record<string, string | number | undefined>
) {
  return request<PageResult<AdminPermission>>(
    `/api/dashboard/permission/list${toQueryString(query)}`
  )
}

export function fetchConversations(
  query?: Record<string, string | number | undefined>
) {
  return request<PageResult<AdminConversation>>(
    `/api/dashboard/conversation/list${toQueryString(query)}`
  )
}

export function fetchConversationDetail(id: number) {
  return request<AdminConversationDetail>(`/api/dashboard/conversation/${id}`)
}

export function fetchConversationMessages(
  query?: Record<string, string | number | undefined>
) {
  return request<CursorResult<AdminMessage>>(
    `/api/dashboard/conversation/message_list${toQueryString(query)}`
  )
}

export function assignConversation(
  conversationId: number,
  assigneeId: number,
  reason: string
) {
  return request<void>("/api/dashboard/conversation/assign", {
    method: "POST",
    body: JSON.stringify({ conversationId, assigneeId, reason }),
  })
}

export function dispatchConversation(conversationId: number) {
  return request<void>("/api/dashboard/conversation/dispatch", {
    method: "POST",
    body: JSON.stringify({ conversationId }),
  })
}

export function transferConversation(
  conversationId: number,
  toUserId: number,
  reason: string
) {
  return request<void>("/api/dashboard/conversation/transfer", {
    method: "POST",
    body: JSON.stringify({ conversationId, toUserId, reason }),
  })
}

export function closeConversation(conversationId: number, closeReason: string) {
  return request<void>("/api/dashboard/conversation/close", {
    method: "POST",
    body: JSON.stringify({ conversationId, closeReason }),
  })
}

export function markConversationRead(conversationId: number, messageId = 0) {
  return request<void>("/api/dashboard/conversation/read", {
    method: "POST",
    body: JSON.stringify({ conversationId, messageId }),
  })
}

export function sendConversationMessage(payload: {
  conversationId: number
  messageType: string
  content: string
  payload?: string
  clientMsgId?: string
}) {
  return request<AdminMessage>("/api/dashboard/conversation/send_message", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function recallConversationMessage(messageId: number) {
  return request<AdminMessage>("/api/dashboard/conversation/recall_message", {
    method: "POST",
    body: JSON.stringify({ messageId }),
  })
}

export function fetchQuickReplies(
  query?: Record<string, string | number | undefined>
) {
  return request<PageResult<AdminQuickReply>>(
    `/api/dashboard/quick-reply/list${toQueryString(query)}`
  )
}

export function fetchQuickReplyListAll() {
  return request<AdminQuickReply[]>("/api/dashboard/quick-reply/list_all")
}

export function fetchQuickReply(id: number) {
  return request<AdminQuickReply>(`/api/dashboard/quick-reply/${id}`)
}

export function createQuickReply(payload: CreateAdminQuickReplyPayload) {
  return request<AdminQuickReply>("/api/dashboard/quick-reply/create", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function updateQuickReply(payload: UpdateAdminQuickReplyPayload) {
  return request<void>("/api/dashboard/quick-reply/update", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function deleteQuickReply(id: number) {
  return request<void>("/api/dashboard/quick-reply/delete", {
    method: "POST",
    body: JSON.stringify({ id }),
  })
}

export function fetchSkillDefinitions(
  query?: Record<string, string | number | undefined>
) {
  return request<PageResult<SkillDefinition>>(
    `/api/dashboard/skill-definition/list${toQueryString(query)}`
  )
}

export function fetchSkillDefinitionsAll(
  query?: Record<string, string | number | undefined>
) {
  return request<SkillDefinition[]>(
    `/api/dashboard/skill-definition/list_all${toQueryString(query)}`
  )
}

export function fetchSkillDefinition(id: number) {
  return request<SkillDefinition>(`/api/dashboard/skill-definition/${id}`)
}

export function createSkillDefinition(payload: CreateSkillDefinitionPayload) {
  return request<SkillDefinition>("/api/dashboard/skill-definition/create", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function updateSkillDefinition(payload: UpdateSkillDefinitionPayload) {
  return request<void>("/api/dashboard/skill-definition/update", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function fetchAgentRunLogs(
  query?: Record<string, string | number | undefined>
) {
  return request<PageResult<AgentRunLog>>(
    `/api/dashboard/agent-run-log/list${toQueryString(query)}`
  )
}

export function fetchAgentRunLog(id: number) {
  return request<AgentRunLog>(`/api/dashboard/agent-run-log/${id}`)
}

export function updateSkillDefinitionStatus(id: number, status: number) {
  return request<void>("/api/dashboard/skill-definition/update_status", {
    method: "POST",
    body: JSON.stringify({ id, status }),
  })
}

export function deleteSkillDefinition(id: number) {
  return request<void>("/api/dashboard/skill-definition/delete", {
    method: "POST",
    body: JSON.stringify({ id }),
  })
}

export function debugRunSkillDefinition(payload: SkillDebugRunPayload) {
  return request<SkillDebugRunResult>("/api/dashboard/skill-definition/debug_run", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function debugResumeSkillDefinition(payload: SkillDebugResumePayload) {
  return request<SkillDebugRunResult>("/api/dashboard/skill-definition/debug_resume", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function testMCPConnection(serverCode: string) {
  return request<MCPConnectionResult>("/api/dashboard/mcp/test_connection", {
    method: "POST",
    body: JSON.stringify({ serverCode }),
  })
}

export function listMCPServers() {
  return request<MCPServerInfo[]>("/api/dashboard/mcp/list_servers")
}

export function listMCPTools(serverCode: string) {
  return request<MCPToolInfo[]>("/api/dashboard/mcp/list_tools", {
    method: "POST",
    body: JSON.stringify({ serverCode }),
  })
}

export function fetchMCPCatalog() {
  return request<MCPToolCatalogItem[]>("/api/dashboard/mcp/catalog")
}

export function callMCPTool(payload: {
  serverCode: string
  toolName: string
  arguments: Record<string, unknown>
}) {
  return request<MCPToolCallResult>("/api/dashboard/mcp/call_tool", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function fetchAgentProfiles(
  query?: Record<string, string | number | undefined>
) {
  return request<PageResult<AdminAgentProfile>>(
    `/api/dashboard/agent/list${toQueryString(query)}`
  )
}

export function fetchAgentProfile(id: number) {
  return request<AdminAgentProfile>(`/api/dashboard/agent/${id}`)
}

export function fetchAgentProfilesAll(
  query?: Record<string, string | number | undefined>
) {
  return request<AdminAgentProfile[]>(
    `/api/dashboard/agent/list_all${toQueryString(query)}`
  )
}

export function createAgentProfile(payload: CreateAdminAgentProfilePayload) {
  return request<AdminAgentProfile>("/api/dashboard/agent/create", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function updateAgentProfile(payload: UpdateAdminAgentProfilePayload) {
  return request<void>("/api/dashboard/agent/update", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function deleteAgentProfile(id: number) {
  return request<void>("/api/dashboard/agent/delete", {
    method: "POST",
    body: JSON.stringify({ id }),
  })
}

export function fetchAgentTeams(query?: Record<string, string | number | undefined>) {
  return request<AdminAgentTeam[]>(
    `/api/dashboard/agent-team/list${toQueryString(query)}`
  )
}

export function fetchAgentTeamsAll() {
  return request<AdminAgentTeam[]>(
    `/api/dashboard/agent-team/list_all`
  )
}

export function fetchAgentTeam(id: number) {
  return request<AdminAgentTeam>(`/api/dashboard/agent-team/${id}`)
}

export function createAgentTeam(payload: CreateAdminAgentTeamPayload) {
  return request<AdminAgentTeam>("/api/dashboard/agent-team/create", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function updateAgentTeam(payload: UpdateAdminAgentTeamPayload) {
  return request<void>("/api/dashboard/agent-team/update", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function deleteAgentTeam(id: number) {
  return request<void>("/api/dashboard/agent-team/delete", {
    method: "POST",
    body: JSON.stringify({ id }),
  })
}

export function fetchAgentTeamSchedules(
  query?: Record<string, string | number | undefined>
) {
  return request<PageResult<AdminAgentTeamSchedule>>(
    `/api/dashboard/agent-team-schedule/list${toQueryString(query)}`
  )
}

export function fetchAgentTeamSchedule(id: number) {
  return request<AdminAgentTeamSchedule>(`/api/dashboard/agent-team-schedule/${id}`)
}

export function createAgentTeamSchedule(payload: CreateAdminAgentTeamSchedulePayload) {
  return request<AdminAgentTeamSchedule>("/api/dashboard/agent-team-schedule/create", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function updateAgentTeamSchedule(payload: UpdateAdminAgentTeamSchedulePayload) {
  return request<void>("/api/dashboard/agent-team-schedule/update", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function deleteAgentTeamSchedule(id: number) {
  return request<void>("/api/dashboard/agent-team-schedule/delete", {
    method: "POST",
    body: JSON.stringify({ id }),
  })
}

export type AIConfig = {
  id: number
  name: string
  provider: string
  baseUrl: string
  apiKey: string
  modelType: string
  modelName: string
  dimension: number
  maxContextTokens: number
  maxOutputTokens: number
  timeoutMs: number
  maxRetryCount: number
  rpmLimit: number
  tpmLimit: number
  status: number
  sortNo: number
  remark: string
}

export type CreateAIConfigPayload = {
  name: string
  provider: string
  baseUrl: string
  apiKey: string
  modelType: string
  modelName: string
  dimension: number
  maxContextTokens: number
  maxOutputTokens: number
  timeoutMs: number
  maxRetryCount: number
  rpmLimit: number
  tpmLimit: number
  remark: string
}

export type UpdateAIConfigPayload = CreateAIConfigPayload & {
  id: number
}

export function fetchAIConfigs(
  query?: Record<string, string | number | undefined>
) {
  return request<PageResult<AIConfig>>(
    `/api/dashboard/ai-config/list${toQueryString(query)}`
  )
}

export function fetchAIConfig(id: number) {
  return request<AIConfig>(`/api/dashboard/ai-config/${id}`)
}

export function fetchAIConfigsAll(
  query?: Record<string, string | number | undefined>
) {
  return request<AIConfig[]>(
    `/api/dashboard/ai-config/list_all${toQueryString(query)}`
  )
}

export function createAIConfig(payload: CreateAIConfigPayload) {
  return request<AIConfig>("/api/dashboard/ai-config/create", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function updateAIConfig(payload: UpdateAIConfigPayload) {
  return request<void>("/api/dashboard/ai-config/update", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function deleteAIConfig(id: number) {
  return request<void>("/api/dashboard/ai-config/delete", {
    method: "POST",
    body: JSON.stringify({ id }),
  })
}

export function updateAIConfigStatus(id: number, status: number) {
  return request<void>("/api/dashboard/ai-config/update_status", {
    method: "POST",
    body: JSON.stringify({ id, status }),
  })
}

export function updateAIConfigSort(ids: number[]) {
  return request<void>("/api/dashboard/ai-config/update_sort", {
    method: "POST",
    body: JSON.stringify(ids),
  })
}

export type KnowledgeBase = {
  id: number
  name: string
  description: string
  knowledgeType: string
  knowledgeTypeName: string
  status: number
  statusName: string
  defaultTopK: number
  defaultScoreThreshold: number
  defaultRerankLimit: number
  chunkProvider: string
  chunkTargetTokens: number
  chunkMaxTokens: number
  chunkOverlapTokens: number
  answerMode: number
  answerModeName: string
  fallbackMode: number
  fallbackModeName: string
  documentCount: number
  faqCount: number
  remark: string
  createdAt: string
  updatedAt: string
  createUserName: string
  updateUserName: string
}

export type CreateKnowledgeBasePayload = {
  name: string
  description: string
  knowledgeType: string
  defaultTopK: number
  defaultScoreThreshold: number
  defaultRerankLimit: number
  chunkProvider: string
  chunkTargetTokens: number
  chunkMaxTokens: number
  chunkOverlapTokens: number
  answerMode: number
  fallbackMode: number
  remark: string
}

export type UpdateKnowledgeBasePayload = CreateKnowledgeBasePayload & {
  id: number
}

export type KnowledgeDocument = {
  id: number
  knowledgeBaseId: number
  knowledgeBaseName?: string
  title: string
  contentType: string
  content: string
  status: number
  statusName: string
  indexStatus: string
  indexStatusName: string
  indexedAt?: string | null
  indexError: string
  contentHash: string
  createdAt: string
  updatedAt: string
  createUserName: string
  updateUserName: string
}

export type KnowledgeFAQ = {
  id: number
  knowledgeBaseId: number
  knowledgeBaseName?: string
  question: string
  answer: string
  similarQuestions: string[]
  status: number
  statusName: string
  indexStatus: string
  indexStatusName: string
  indexedAt?: string | null
  indexError: string
  remark: string
  createdAt: string
  updatedAt: string
  createUserName: string
  updateUserName: string
}

export type KnowledgeSearchResult = {
  knowledgeBaseId: number
  chunkId: number
  documentId: number
  documentTitle: string
  faqId: number
  faqQuestion: string
  chunkNo: number
  title: string
  sectionPath: string
  content: string
  score: number
  rerankScore?: number
}

export type KnowledgeSearchResponse = {
  question: string
  results: KnowledgeSearchResult[]
  hitCount: number
  latencyMs: number
}

export type KnowledgeAnswerResponse = {
  question: string
  rewriteQuestion?: string
  answer: string
  answerStatus: number
  answerStatusName: string
  citations: KnowledgeCitation[]
  hits: KnowledgeSearchResult[]
  hitCount: number
  topScore: number
  latencyMs: number
  retrieveMs: number
  generateMs: number
  promptTokens: number
  completionTokens: number
  modelName: string
  retrieveLogId: number
}

export type KnowledgeCitation = {
  documentId: number
  documentTitle: string
  faqId: number
  faqQuestion: string
  chunkNo: number
  title: string
  sectionPath: string
  snippet: string
  score: number
}

export type KnowledgeRetrieveLog = {
  id: number
  knowledgeBaseId: number
  knowledgeBaseName?: string
  channel: string
  channelName: string
  scene: string
  sceneName: string
  sessionId: string
  conversationId: number
  requestId: string
  question: string
  rewriteQuestion: string
  answer: string
  answerStatus: number
  answerStatusName: string
  hitCount: number
  topScore: number
  chunkProvider: string
  chunkTargetTokens: number
  chunkMaxTokens: number
  chunkOverlapTokens: number
  rerankEnabled: boolean
  rerankLimit: number
  citationCount: number
  usedChunkCount: number
  latencyMs: number
  retrieveMs: number
  generateMs: number
  promptTokens: number
  completionTokens: number
  modelName: string
  traceData: string
  createdAt: string
}

export type KnowledgeRetrieveHit = {
  id: number
  retrieveLogId: number
  knowledgeBaseId: number
  chunkId: number
  documentId: number
  documentTitle: string
  faqId: number
  faqQuestion: string
  chunkNo: number
  title: string
  sectionPath: string
  chunkType: string
  chunkTypeName: string
  provider: string
  rankNo: number
  score: number
  rerankScore: number
  usedInAnswer: boolean
  isCitation: boolean
  snippet: string
  createdAt: string
}

export type KnowledgeRetrieveLogDetail = {
  log: KnowledgeRetrieveLog
  hits: KnowledgeRetrieveHit[]
}

export type KnowledgeRetrieveLogListQuery = {
  knowledgeBaseId: number
  question?: string
  channel?: string
  scene?: string
  answerStatus?: number
  chunkProvider?: string
  rerankEnabled?: number
  page?: number
  limit?: number
}

export type KnowledgeSearchPayload = {
  knowledgeBaseIds: number[]
  question: string
  topK?: number
  scoreThreshold?: number
  rerankLimit?: number
}

export type KnowledgeAnswerPayload = KnowledgeSearchPayload & {
  answerMode?: number
  fallbackMode?: number
}

export type CreateKnowledgeDocumentPayload = {
  knowledgeBaseId: number
  title: string
  contentType: string
  content: string
}

export type UpdateKnowledgeDocumentPayload = CreateKnowledgeDocumentPayload & {
  id: number
}

export type CreateKnowledgeFAQPayload = {
  knowledgeBaseId: number
  question: string
  answer: string
  similarQuestions: string[]
  remark: string
}

export type UpdateKnowledgeFAQPayload = CreateKnowledgeFAQPayload & {
  id: number
}

export function fetchKnowledgeBases(
  query?: Record<string, string | number | undefined>
) {
  return request<PageResult<KnowledgeBase>>(
    `/api/dashboard/knowledge-base/list${toQueryString(query)}`
  )
}

export function fetchKnowledgeBasesAll(
  query?: Record<string, string | number | undefined>
) {
  return request<KnowledgeBase[]>(
    `/api/dashboard/knowledge-base/list_all${toQueryString(query)}`
  )
}

export function fetchKnowledgeBase(id: number) {
  return request<KnowledgeBase>(`/api/dashboard/knowledge-base/${id}`)
}

export function createKnowledgeBase(payload: CreateKnowledgeBasePayload) {
  return request<KnowledgeBase>("/api/dashboard/knowledge-base/create", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function updateKnowledgeBase(payload: UpdateKnowledgeBasePayload) {
  return request<void>("/api/dashboard/knowledge-base/update", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function deleteKnowledgeBase(id: number) {
  return request<void>("/api/dashboard/knowledge-base/delete", {
    method: "POST",
    body: JSON.stringify({ id }),
  })
}

export function updateKnowledgeBaseSort(ids: number[]) {
  return request<void>("/api/dashboard/knowledge-base/update_sort", {
    method: "POST",
    body: JSON.stringify(ids),
  })
}

export function rebuildKnowledgeBaseIndex(id: number) {
  return request<void>("/api/dashboard/knowledge-base/rebuild_index", {
    method: "POST",
    body: JSON.stringify({ id }),
  })
}

export function fetchKnowledgeDocuments(
  query?: Record<string, string | number | undefined>
) {
  return request<PageResult<KnowledgeDocument>>(
    `/api/dashboard/knowledge-document/list${toQueryString(query)}`
  )
}

export function fetchKnowledgeDocument(id: number) {
  return request<KnowledgeDocument>(`/api/dashboard/knowledge-document/${id}`)
}

export function createKnowledgeDocument(payload: CreateKnowledgeDocumentPayload) {
  return request<KnowledgeDocument>("/api/dashboard/knowledge-document/create", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function updateKnowledgeDocument(payload: UpdateKnowledgeDocumentPayload) {
  return request<void>("/api/dashboard/knowledge-document/update", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function deleteKnowledgeDocument(id: number) {
  return request<void>("/api/dashboard/knowledge-document/delete", {
    method: "POST",
    body: JSON.stringify({ id }),
  })
}

export function buildKnowledgeDocumentIndex(documentId: number) {
  return request<void>("/api/dashboard/knowledge-retrieve/build", {
    method: "POST",
    body: JSON.stringify({ documentId }),
  })
}

export function fetchKnowledgeFAQs(
  query?: Record<string, string | number | undefined>
) {
  return request<PageResult<KnowledgeFAQ>>(
    `/api/dashboard/knowledge-faq/list${toQueryString(query)}`
  )
}

export function fetchKnowledgeFAQ(id: number) {
  return request<KnowledgeFAQ>(`/api/dashboard/knowledge-faq/${id}`)
}

export function createKnowledgeFAQ(payload: CreateKnowledgeFAQPayload) {
  return request<KnowledgeFAQ>("/api/dashboard/knowledge-faq/create", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function updateKnowledgeFAQ(payload: UpdateKnowledgeFAQPayload) {
  return request<void>("/api/dashboard/knowledge-faq/update", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function deleteKnowledgeFAQ(id: number) {
  return request<void>("/api/dashboard/knowledge-faq/delete", {
    method: "POST",
    body: JSON.stringify({ id }),
  })
}

export function buildKnowledgeFAQIndex(faqId: number) {
  return request<void>("/api/dashboard/knowledge-retrieve/build", {
    method: "POST",
    body: JSON.stringify({ faqId }),
  })
}

export function debugKnowledgeSearch(payload: KnowledgeSearchPayload) {
  return request<KnowledgeSearchResponse>("/api/dashboard/knowledge-retrieve/debug/search", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function debugKnowledgeAnswer(payload: KnowledgeAnswerPayload) {
  return request<KnowledgeAnswerResponse>("/api/dashboard/knowledge-retrieve/debug/answer", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function fetchKnowledgeRetrieveLogs(query: KnowledgeRetrieveLogListQuery) {
  return request<PageResult<KnowledgeRetrieveLog>>(
    `/api/dashboard/knowledge-retrieve-log/list${toQueryString(query)}`
  )
}

export function fetchKnowledgeRetrieveLog(id: number) {
  return request<KnowledgeRetrieveLogDetail>(`/api/dashboard/knowledge-retrieve-log/${id}`)
}

export type AdminAsset = {
  id: number
  assetId: string
  provider: string
  filename: string
  fileSize: number
  mimeType: string
  status: number
  url: string
  createdAt: string
  updatedAt: string
  createUserId: number
  createUserName: string
  updateUserId: number
  updateUserName: string
}

export function uploadAsset(file: File, prefix?: string) {
  const formData = new FormData()
  formData.set("file", file)
  if (prefix) {
    formData.set("prefix", prefix)
  }
  return request<AdminAsset>("/api/dashboard/asset/create", {
    method: "POST",
    body: formData,
  })
}

export type Tag = {
  id: number
  parentId: number
  name: string
  remark: string
  sortNo: number
  status: number
  createdAt: string
  updatedAt: string
}

export type TagTree = {
  id: number
  parentId: number
  name: string
  remark: string
  sortNo: number
  status: number
  createdAt: string
  updatedAt: string
  children: TagTree[]
}

export type CreateTagPayload = {
  parentId: number
  name: string
  remark: string
  status: number
}

export type UpdateTagPayload = CreateTagPayload & {
  id: number
}

export function fetchTags(query?: Record<string, string | number | undefined>) {
  return request<PageResult<Tag>>(
    `/api/dashboard/tag/list${toQueryString(query)}`
  )
}

export function fetchTagsAll() {
  return request<TagTree[]>("/api/dashboard/tag/list_all")
}

export function fetchTag(id: number) {
  return request<Tag>(`/api/dashboard/tag/${id}`)
}

export function createTag(payload: CreateTagPayload) {
  return request<Tag>("/api/dashboard/tag/create", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function updateTag(payload: UpdateTagPayload) {
  return request<void>("/api/dashboard/tag/update", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function deleteTag(id: number) {
  return request<void>("/api/dashboard/tag/delete", {
    method: "POST",
    body: JSON.stringify({ id }),
  })
}

export function updateTagSort(ids: number[]) {
  return request<void>("/api/dashboard/tag/update_sort", {
    method: "POST",
    body: JSON.stringify(ids),
  })
}

export function updateTagStatus(id: number, status: number) {
  return request<void>("/api/dashboard/tag/update_status", {
    method: "POST",
    body: JSON.stringify({ id, status }),
  })
}
