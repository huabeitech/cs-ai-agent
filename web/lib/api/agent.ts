import { request } from "@/lib/api/client"
import { readSession } from "@/lib/auth"

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

export type AgentConversationTag = {
  id: number
  name: string
  color: string
}

export type AgentConversationParticipant = {
  id: number
  participantType: string
  participantId: number
  externalParticipantId?: string
  joinedAt?: string
  leftAt?: string
  status: number
}

export type AgentConversation = {
  id: number
  aiAgentId?: number
  customerId?: number
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
  tags?: AgentConversationTag[]
  participants?: AgentConversationParticipant[]
}

export type AgentConversationDetail = AgentConversation & {
  participants?: AgentConversationParticipant[]
}

export type AgentMessage = {
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

export type AgentAsset = {
  id: number
  assetId: string
  provider: string
  storageKey: string
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

export function fetchAgentConversations(
  query?: Record<string, string | number | undefined>
) {
  return request<PageResult<AgentConversation>>(
    `/api/console/conversation/conversations${toQueryString(query)}`
  )
}

export function fetchAgentConversationDetail(id: number) {
  return request<AgentConversationDetail>(`/api/console/conversation/${id}`)
}

export function fetchAgentMessages(
  query?: Record<string, string | number | undefined>
) {
  return request<CursorResult<AgentMessage>>(
    `/api/console/conversation/message_list${toQueryString(query)}`
  )
}

export function sendAgentMessage(payload: {
  conversationId: number
  messageType: string
  content: string
  payload?: string
  clientMsgId?: string
}) {
  return request<AgentMessage>("/api/console/conversation/send_message", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function markAgentMessageRead(conversationId: number, messageId = 0) {
  return request<void>("/api/console/conversation/read", {
    method: "POST",
    body: JSON.stringify({ conversationId, messageId }),
  })
}

export function uploadAgentConversationImage(file: File) {
  const formData = new FormData()
  formData.set("file", file)
  return request<AgentAsset>("/api/console/conversation/upload_image", {
    method: "POST",
    body: formData,
  })
}

export function closeAgentConversation(
  conversationId: number,
  closeReason: string
) {
  return request<void>("/api/console/conversation/close", {
    method: "POST",
    body: JSON.stringify({ conversationId, closeReason }),
  })
}

export function assignAgentConversation(
  conversationId: number,
  assigneeId: number,
  reason: string
) {
  return request<void>("/api/console/conversation/assign", {
    method: "POST",
    body: JSON.stringify({ conversationId, assigneeId, reason }),
  })
}

export function transferAgentConversation(
  conversationId: number,
  toUserId: number,
  reason: string
) {
  return request<void>("/api/console/conversation/transfer", {
    method: "POST",
    body: JSON.stringify({ conversationId, toUserId, reason }),
  })
}

export function linkConversationToCustomer(payload: {
  conversationId: number
  customerId: number
}) {
  return request<void>("/api/console/conversation/link_customer", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function createAgentWebSocketUrl() {
  const session = readSession()
  if (!session?.accessToken) {
    throw new Error("未登录或登录已过期")
  }

  const baseUrl = (
    process.env.NEXT_PUBLIC_API_BASE_URL?.trim() || "http://127.0.0.1:8083"
  ).replace(/^http/, "ws")
  const params = new URLSearchParams({
    accessToken: session.accessToken,
  })
  return `${baseUrl}/api/admin/ws?${params.toString()}`
}
