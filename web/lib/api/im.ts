import { request } from "@/lib/api/client"
import { readKefuWidgetConfig } from "@/lib/kefu-widget-config"
import { generateUUID } from "@/lib/utils"

export type Paging = {
  page: number
  limit: number
  total: number
}

export type PageResult<T> = {
  results?: T[] | null
  page: Paging
  cursor?: string
  hasMore?: boolean
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
  channelId: number
  customerName: string
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

export type ImWidgetConfig = {
  channelId?: string
  channelType?: string
  externalSource?: string
  userToken?: string
  title?: string
  subtitle?: string
  themeColor?: string
  position?: "left" | "right"
  width?: string
}

const GUEST_STORAGE_KEY = "cs_agent_im_guest_id"
const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL?.trim() || ""
const OPEN_IM_CHANNEL_ID =
  process.env.NEXT_PUBLIC_OPEN_IM_CHANNEL_ID?.trim() || ""
const OPEN_IM_EXTERNAL_SOURCE =
  process.env.NEXT_PUBLIC_OPEN_IM_EXTERNAL_SOURCE?.trim() || "web_chat"

function buildGuestId() {
  return `guest_${generateUUID()}`
}

export function getGuestId() {
  if (typeof window === "undefined") {
    return ""
  }
  const existing = window.localStorage.getItem(GUEST_STORAGE_KEY)?.trim()
  if (existing) {
    return existing
  }
  const guestId = buildGuestId()
  window.localStorage.setItem(GUEST_STORAGE_KEY, guestId)
  return guestId
}

function getRuntimeImConfig() {
  const widgetConfig = readKefuWidgetConfig()
  const baseUrl = (widgetConfig.apiBaseUrl || widgetConfig.baseUrl || API_BASE_URL)
    .trim()
    .replace(/\/$/, "")
  return {
    baseUrl,
    channelId: widgetConfig.channelId || OPEN_IM_CHANNEL_ID,
    externalSource:
      (widgetConfig.externalSource || OPEN_IM_EXTERNAL_SOURCE).trim() || "web_chat",
    externalId: (widgetConfig.externalId || "").trim(),
    externalName: (widgetConfig.externalName || "").trim(),
    userToken: (widgetConfig.userToken || "").trim(),
  }
}

function createImHeaders() {
  const config = getRuntimeImConfig()
  const headers: Record<string, string> = {
    "X-Channel-Id": config.channelId,
  }
  if (config.userToken) {
    headers.Authorization = `Bearer ${config.userToken}`
  } else {
    headers["X-External-Source"] = config.externalSource
    headers["X-External-Id"] = config.externalId || getGuestId()
    if (config.externalName) {
      headers["X-External-Name"] = encodeURIComponent(config.externalName)
    }
  }
  return {
    ...headers,
  }
}

function createRequestOptions(
  init?: RequestInit
): RequestInit & { baseUrl?: string; skipAuth?: boolean } {
  return {
    ...init,
    skipAuth: true,
    headers: {
      ...createImHeaders(),
      ...(init?.headers as Record<string, string> | undefined),
    },
    baseUrl: getRuntimeImConfig().baseUrl,
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
  return request<ImConversationDetail>(`/api/conversation/${id}`, {
    ...createRequestOptions(),
  })
}

export function fetchImMessages(
  query?: Record<string, string | number | undefined>
) {
  return request<PageResult<ImMessage>>(
    `/api/message/list${toQueryString(query)}`,
    createRequestOptions()
  )
}

/** 外部身份仅通过 createImHeaders()（Authorization 或 X-External-*）传递，无 JSON body */
export function createOrMatchImConversation() {
  return request<ImConversation>("/api/conversation/create_or_match", {
    ...createRequestOptions({ method: "POST" }),
  })
}

export function fetchImWidgetConfig() {
  return request<ImWidgetConfig>(
    `/api/channel/config${toQueryString({
      channelId: getRuntimeImConfig().channelId,
    })}`,
    createRequestOptions()
  )
}

export function closeImConversation(conversationId: number) {
  return request<void>("/api/conversation/close", {
    ...createRequestOptions({
      method: "POST",
      body: JSON.stringify({ conversationId }),
    }),
  })
}

export function sendImMessage(payload: {
  conversationId: number
  messageType: string
  content: string
  payload?: string
  clientMsgId?: string
}) {
  return request<ImMessage>("/api/message/send", {
    ...createRequestOptions({
      method: "POST",
      body: JSON.stringify(payload),
    }),
  })
}

export function markImMessageRead(conversationId: number, messageId = 0) {
  return request<void>("/api/message/read", {
    ...createRequestOptions({
      method: "POST",
      body: JSON.stringify({ conversationId, messageId }),
    }),
  })
}

export function uploadImImage(conversationId: number, file: File) {
  const formData = new FormData()
  formData.set("conversationId", String(conversationId))
  formData.set("file", file)
  return request<ImAsset>("/api/message/upload_image", {
    ...createRequestOptions({
      method: "POST",
      body: formData,
    }),
  })
}

export function uploadImAttachment(conversationId: number, file: File) {
  const formData = new FormData()
  formData.set("conversationId", String(conversationId))
  formData.set("file", file)
  return request<ImAsset>("/api/message/upload_attachment", {
    ...createRequestOptions({
      method: "POST",
      body: formData,
    }),
  })
}
