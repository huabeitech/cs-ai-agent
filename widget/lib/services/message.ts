import { requestJson } from "@/lib/services/http";
import type {
  CursorResult,
  JsonResult,
  WidgetAsset,
  WidgetMessage,
} from "@/lib/services/types";
import { generateUUID } from "@/lib/utils";

const DEFAULT_PAGE_LIMIT = 50;

function buildListQuery(
  conversationId: number,
  options?: { cursor?: number; limit?: number },
) {
  const params = new URLSearchParams({
    conversationId: String(conversationId),
  });
  const cursor = options?.cursor;
  if (cursor !== undefined && cursor > 0) {
    params.set("cursor", String(cursor));
  }
  const limit = options?.limit ?? DEFAULT_PAGE_LIMIT;
  if (limit > 0) {
    params.set("limit", String(limit));
  }
  return params.toString();
}

export async function fetchMessagesPage(
  conversationId: number,
  options?: { cursor?: number; limit?: number },
): Promise<CursorResult<WidgetMessage>> {
  const qs = buildListQuery(conversationId, options);
  const result = await requestJson<
    JsonResult<CursorResult<WidgetMessage>>
  >(`/api/open/im/message/list?${qs}`);
  if (result.success === false) {
    throw new Error(result.message || "加载消息失败");
  }
  const data = result.data;
  return {
    results: data?.results ?? [],
    cursor: data?.cursor ?? "",
    hasMore: Boolean(data?.hasMore),
  };
}

/** @deprecated 使用 fetchMessagesPage；保留别名供渐进迁移 */
export async function fetchMessages(conversationId: number) {
  const page = await fetchMessagesPage(conversationId);
  return page.results;
}

export async function sendMessageWithPayload(
  conversationId: number,
  payload: {
    messageType: string;
    content: string;
    payload?: string;
    clientMsgId?: string;
  },
) {
  const result = await requestJson<JsonResult<WidgetMessage>>("/api/open/im/message/send", {
    method: "POST",
    body: JSON.stringify({
      conversationId,
      clientMsgId: payload.clientMsgId || `client_${generateUUID()}`,
      messageType: payload.messageType,
      content: payload.content,
      payload: payload.payload || "",
    }),
  });
  if (!result.data) {
    throw new Error(result.message || "send message failed");
  }
  return result.data;
}

export async function sendMessage(conversationId: number, content: string) {
  return sendMessageWithPayload(conversationId, {
    messageType: "html",
    content,
    payload: "",
  });
}

export async function markMessageRead(conversationId: number, messageId = 0) {
  await requestJson<JsonResult<void>>("/api/open/im/message/read", {
    method: "POST",
    body: JSON.stringify({ conversationId, messageId }),
  });
}

export async function uploadImage(conversationId: number, file: File) {
  const formData = new FormData();
  formData.set("conversationId", String(conversationId));
  formData.set("file", file);
  const result = await requestJson<JsonResult<WidgetAsset>>(
    "/api/open/im/message/upload_image",
    {
      method: "POST",
      body: formData,
    },
  );
  if (!result.data) {
    throw new Error(result.message || "upload image failed");
  }
  return result.data;
}

export async function uploadAttachment(conversationId: number, file: File) {
  const formData = new FormData();
  formData.set("conversationId", String(conversationId));
  formData.set("file", file);
  const result = await requestJson<JsonResult<WidgetAsset>>(
    "/api/open/im/message/upload_attachment",
    {
      method: "POST",
      body: formData,
    },
  );
  if (!result.data) {
    throw new Error(result.message || "upload attachment failed");
  }
  return result.data;
}
