import { request } from "@/lib/api/client"

export type Paging = {
  page: number
  limit: number
  total: number
}

export type PageResult<T> = {
  results: T[]
  page: Paging
}

export type ImConversationTag = {
  id: number
  name: string
  color: string
}

export type ImConversationParticipant = {
  id: number
  participantType: string
  participantId: number
  externalParticipantId?: string
  joinedAt?: string
  leftAt?: string
  status: number
}

export type ImConversation = {
  id: number
  channelType: string
  subject: string
  status: number
  serviceMode: number
  priority: number
  sourceUserId: number
  externalUserId: string
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
  tags?: ImConversationTag[]
  participants?: ImConversationParticipant[]
}

export type ImConversationDetail = ImConversation

export type ImMessage = {
  id: number
  conversationId: number
  clientMsgId?: string
  senderType: string
  senderId: number
  senderName?: string
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

export type ImAsset = {
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

const VISITOR_STORAGE_KEY = "cs_agent_im_visitor_id"
const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL?.trim() || "http://127.0.0.1:8083"
const OPEN_IM_APP_ID =
  process.env.NEXT_PUBLIC_OPEN_IM_APP_ID?.trim() || ""

function buildVisitorId() {
  if (typeof crypto !== "undefined" && typeof crypto.randomUUID === "function") {
    return `visitor_${crypto.randomUUID()}`
  }
  return `visitor_${Date.now()}_${Math.random().toString(36).slice(2, 10)}`
}

export function getImVisitorId() {
  if (typeof window === "undefined") {
    return "visitor_ssr"
  }
  const existing = window.localStorage.getItem(VISITOR_STORAGE_KEY)?.trim()
  if (existing) {
    return existing
  }
  const visitorId = buildVisitorId()
  window.localStorage.setItem(VISITOR_STORAGE_KEY, visitorId)
  return visitorId
}

export function createImWebSocketUrl() {
  const baseUrl = API_BASE_URL.replace(/^http/, "ws")
  const params = new URLSearchParams({
    visitorId: getImVisitorId(),
    appId: OPEN_IM_APP_ID,
  })
  return `${baseUrl}/api/open/im/ws?${params.toString()}`
}

function createImHeaders() {
  return {
    "X-Visitor-Id": getImVisitorId(),
    "X-Widget-App-Id": OPEN_IM_APP_ID,
  }
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

export function fetchImConversations(
  query?: Record<string, string | number | undefined>
) {
  return request<PageResult<ImConversation>>(
    `/api/open/im/conversation/list${toQueryString(query)}`,
    { headers: createImHeaders() }
  )
}

export function fetchImConversationDetail(id: number) {
  return request<ImConversationDetail>(`/api/open/im/conversation/${id}`, {
    headers: createImHeaders(),
  })
}

export function fetchImMessages(
  query?: Record<string, string | number | undefined>
) {
  return request<PageResult<ImMessage>>(
    `/api/open/im/message/list${toQueryString(query)}`,
    { headers: createImHeaders() }
  )
}

export function createOrMatchImConversation(payload?: {
  channelType?: string
  subject?: string
}) {
  return request<ImConversation>("/api/open/im/conversation/create_or_match", {
    method: "POST",
    headers: createImHeaders(),
    body: JSON.stringify(payload ?? {}),
  })
}

export function sendImMessage(payload: {
  conversationId: number
  messageType: string
  content: string
  payload?: string
  clientMsgId?: string
}) {
  return request<ImMessage>("/api/open/im/message/send", {
    method: "POST",
    headers: createImHeaders(),
    body: JSON.stringify(payload),
  })
}

export function markImMessageRead(conversationId: number, messageId = 0) {
  return request<void>("/api/open/im/message/read", {
    method: "POST",
    headers: createImHeaders(),
    body: JSON.stringify({ conversationId, messageId }),
  })
}

export function uploadImImage(file: File) {
  const formData = new FormData()
  formData.set("file", file)
  return request<ImAsset>("/api/open/im/message/upload_image", {
    method: "POST",
    headers: createImHeaders(),
    body: formData,
  })
}
