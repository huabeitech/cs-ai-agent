import { readWidgetConfig } from "@/lib/widget/config";
import { getOrCreateVisitorId } from "@/lib/widget/visitor";

export type RealtimeEnvelope = {
  type: string;
  topic?: string;
  data?: {
    conversationId?: number;
    messageId?: number;
  };
  payload?: {
    conversationId?: number;
    messageId?: number;
  };
};

export function createRealtimeConnection(onEvent: (event: RealtimeEnvelope) => void) {
  const config = readWidgetConfig();
  const baseUrl = (config.apiBaseUrl || config.baseUrl)
    .replace(/^http/, "ws")
    .replace(/\/$/, "");
  const visitorId = encodeURIComponent(getOrCreateVisitorId());
  const socket = new WebSocket(`${baseUrl}/api/open/im/ws?visitorId=${visitorId}&appId=${encodeURIComponent(config.appId)}`);
  socket.addEventListener("message", (event) => {
    try {
      onEvent(JSON.parse(event.data) as RealtimeEnvelope);
    } catch {
      return;
    }
  });
  return socket;
}
