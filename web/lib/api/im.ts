import { request } from "@/lib/api/client"
import { generateUUID } from "@/lib/utils"

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

export type ImAsset = {
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

const VISITOR_STORAGE_KEY = "cs_agent_im_visitor_id"
const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL?.trim() || "http://127.0.0.1:8083"
const OPEN_IM_CHANNEL_ID =
  process.env.NEXT_PUBLIC_OPEN_IM_CHANNEL_ID?.trim() || ""
const OPEN_IM_EXTERNAL_SOURCE =
  process.env.NEXT_PUBLIC_OPEN_IM_EXTERNAL_SOURCE?.trim() || "web_chat"

function buildVisitorId() {
  return `visitor_${generateUUID()}`
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
    externalId: getImVisitorId(),
    externalSource: OPEN_IM_EXTERNAL_SOURCE,
    channelId: OPEN_IM_CHANNEL_ID,
  })
  return `${baseUrl}/api/open/im/ws?${params.toString()}`
}

function createImHeaders() {
  return {
    "X-External-Source": OPEN_IM_EXTERNAL_SOURCE,
    "X-External-Id": getImVisitorId(),
    "X-Channel-Id": OPEN_IM_CHANNEL_ID,
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

/** 外部身份仅通过 createImHeaders()（X-External-*）传递，无 JSON body */
export function createOrMatchImConversation() {
  return request<ImConversation>("/api/open/im/conversation/create_or_match", {
    method: "POST",
    headers: createImHeaders(),
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

export function uploadImImage(conversationId: number, file: File) {
  const formData = new FormData()
  formData.set("conversationId", String(conversationId))
  formData.set("file", file)
  return request<ImAsset>("/api/open/im/message/upload_image", {
    method: "POST",
    headers: createImHeaders(),
    body: formData,
  })
}

export function uploadImAttachment(conversationId: number, file: File) {
  const formData = new FormData()
  formData.set("conversationId", String(conversationId))
  formData.set("file", file)
  return request<ImAsset>("/api/open/im/message/upload_attachment", {
    method: "POST",
    headers: createImHeaders(),
    body: formData,
  })
}
