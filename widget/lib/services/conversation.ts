import { requestJson } from "@/lib/services/http";
import type { JsonResult, WidgetConversation } from "@/lib/services/types";

export async function createOrMatchConversation(subject?: string) {
  const result = await requestJson<JsonResult<WidgetConversation>>(
    "/api/open/im/conversation/create_or_match",
    {
      method: "POST",
      body: JSON.stringify({
        externalSource: "web_chat",
        subject: subject,
      }),
    },
  );
  if (!result.data) {
    throw new Error(result.message || "conversation init failed");
  }
  return result.data;
}

export async function closeConversation(conversationId: number) {
  const result = await requestJson<JsonResult<null>>(
    "/api/open/im/conversation/close",
    {
      method: "POST",
      body: JSON.stringify({ conversationId }),
    },
  );
  if (result.success === false) {
    throw new Error(result.message || "conversation close failed");
  }
}
