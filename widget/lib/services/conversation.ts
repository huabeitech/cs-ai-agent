import { requestJson } from "@/lib/services/http";
import type { JsonResult, WidgetConversation } from "@/lib/services/types";

/** 身份由请求头 X-External-Source / X-External-Id（及可选 X-External-Name）提供，与 GetExternalInfo 一致 */
export async function createOrMatchConversation() {
  const result = await requestJson<JsonResult<WidgetConversation>>(
    "/api/open/im/conversation/create_or_match",
    {
      method: "POST",
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
