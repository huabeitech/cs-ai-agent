import { createWebSocketBaseUrl } from "@/lib/api/websocket"
import { getImVisitorId } from "@/lib/api/im"
import { readKefuWidgetConfig } from "@/lib/kefu-widget-config"

export type ImRealtimeEnvelope = {
  type: string
  topic?: string
  data?: {
    conversationId?: number
    messageId?: number
  }
  payload?: {
    conversationId?: number
    messageId?: number
  }
}

export function createImRealtimeConnection() {
  const config = readKefuWidgetConfig()
  const apiBaseUrl = (config.apiBaseUrl || config.baseUrl || "").trim()
  const baseUrl = apiBaseUrl
    ? apiBaseUrl.replace(/^http/, "ws").replace(/\/$/, "")
    : createWebSocketBaseUrl()
  const externalId = encodeURIComponent(getImVisitorId())
  const externalSource = encodeURIComponent(
    (config.externalSource ?? "web_chat").trim() || "web_chat"
  )
  const channelId = encodeURIComponent(config.channelId || "")
  const externalName = (config.subject ?? "").trim()
  const nameQuery =
    externalName !== ""
      ? `&externalName=${encodeURIComponent(externalName)}`
      : ""
  return new WebSocket(
    `${baseUrl}/api/open/im/ws?externalId=${externalId}&externalSource=${externalSource}&channelId=${channelId}${nameQuery}`
  )
}
