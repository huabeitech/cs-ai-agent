import { readWidgetConfig } from "@/lib/widget/config";
import { getOrCreateExternalID } from "@/lib/widget/visitor";

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
  const externalId = encodeURIComponent(getOrCreateExternalID());
  const externalSource = encodeURIComponent(
    (config.externalSource ?? "web_chat").trim() || "web_chat",
  );
  const externalName = (config.subject ?? "").trim();
  const nameQuery =
    externalName !== ""
      ? `&externalName=${encodeURIComponent(externalName)}`
      : "";
  const socket = new WebSocket(
    `${baseUrl}/api/open/im/ws?externalId=${externalId}&externalSource=${externalSource}&appId=${encodeURIComponent(config.appId)}${nameQuery}`,
  );
  socket.addEventListener("message", (event) => {
    try {
      onEvent(JSON.parse(event.data) as RealtimeEnvelope);
    } catch {
      return;
    }
  });
  return socket;
}
