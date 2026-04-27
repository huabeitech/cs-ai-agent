import { createWebSocketBaseUrl } from "@/lib/api/websocket"
import { getImVisitorId, type ImMessage } from "@/lib/api/im"
import { readKefuWidgetConfig } from "@/lib/kefu-widget-config"
import type {
  RealtimeConversationPatch,
  RealtimeMessageCreatedPayload,
} from "@/lib/im-realtime-state"

export type ImRealtimeEnvelope = {
  type: string
  topic?: string
  data?: RealtimeMessageCreatedPayload<ImMessage> & RealtimeConversationPatch
  payload?: RealtimeMessageCreatedPayload<ImMessage> & RealtimeConversationPatch
}

export function createImRealtimeConnection() {
  const config = readKefuWidgetConfig()
  const apiBaseUrl = (config.apiBaseUrl || "").trim()
  const baseUrl = apiBaseUrl
    ? apiBaseUrl.replace(/^http/, "ws").replace(/\/$/, "")
    : createWebSocketBaseUrl()
  const resolvedExternalId = encodeURIComponent(
    (config.externalId ?? "").trim() || getImVisitorId()
  )
  const externalSource = encodeURIComponent(
    (config.externalSource ?? "web_chat").trim() || "web_chat"
  )
  const channelId = encodeURIComponent(config.channelId || "")
  const externalName = (config.externalName ?? "").trim()
  const nameQuery =
    externalName !== ""
      ? `&externalName=${encodeURIComponent(externalName)}`
      : ""
  return new WebSocket(
    `${baseUrl}/api/ws/open?externalId=${resolvedExternalId}&externalSource=${externalSource}&channelId=${channelId}${nameQuery}`
  )
}
