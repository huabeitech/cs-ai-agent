import { requestJson } from "@/lib/services/http";
import type {
  JsonResult,
  PageResult,
  WidgetAsset,
  WidgetMessage,
} from "@/lib/services/types";

export async function fetchMessages(conversationId: number) {
  const result = await requestJson<JsonResult<PageResult<WidgetMessage>>>(`/api/open/im/message/list?conversationId=${conversationId}`);
  return result.data?.results ?? [];
}

export async function sendMessage(conversationId: number, content: string) {
  const result = await requestJson<JsonResult<WidgetMessage>>("/api/open/im/message/send", {
    method: "POST",
    body: JSON.stringify({
      conversationId,
      clientMsgId: `client_${crypto.randomUUID()}`,
      messageType: "html",
      content,
      payload: "",
    }),
  });
  if (!result.data) {
    throw new Error(result.message || "send message failed");
  }
  return result.data;
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
