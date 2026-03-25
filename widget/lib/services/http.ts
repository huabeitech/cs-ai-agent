import { readWidgetConfig } from "@/lib/widget/config";
import { getOrCreateVisitorId } from "@/lib/widget/visitor";

export async function requestJson<T>(path: string, init?: RequestInit): Promise<T> {
  const config = readWidgetConfig();
  const baseUrl = (config.apiBaseUrl || config.baseUrl).replace(/\/$/, "");
  const visitorId = getOrCreateVisitorId();
  const headers = new Headers(init?.headers ?? {});
  if (
    !headers.has("Content-Type") &&
    init?.body &&
    !(typeof FormData !== "undefined" && init.body instanceof FormData)
  ) {
    headers.set("Content-Type", "application/json");
  }
  headers.set("X-Visitor-Id", visitorId);
  headers.set("X-Widget-App-Id", config.appId);
  const response = await fetch(`${baseUrl}${path}`, {
    ...init,
    headers,
    cache: "no-store",
  });
  if (!response.ok) {
    throw new Error(`Request failed: ${response.status}`);
  }
  return (await response.json()) as T;
}
