export type KefuWidgetHostConfig = {
  channelId: string
  baseUrl: string
  apiBaseUrl?: string
  /** 与后端 enums.ExternalSource 一致，默认 web_chat */
  externalSource?: string
  title?: string
  subtitle?: string
  position?: "left" | "right"
  themeColor?: string
  width?: string
  /** 访客展示名，随请求以 X-External-Name / WS query externalName 传给后端 */
  subject?: string
}

declare global {
  interface Window {
    CSAgentConfig?: KefuWidgetHostConfig
    __CS_AGENT_WIDGET_CONFIG__?: KefuWidgetHostConfig
    __CS_AGENT_WIDGET_STATE__?: unknown
  }
}

export function readKefuWidgetConfig(): KefuWidgetHostConfig {
  if (typeof window === "undefined") {
    return {
      channelId: "",
      baseUrl: "",
      apiBaseUrl: "",
    }
  }

  const query = new URLSearchParams(window.location.search)
  const fallback: KefuWidgetHostConfig = {
    channelId:
      query.get("channelId") ??
      process.env.NEXT_PUBLIC_OPEN_IM_CHANNEL_ID?.trim() ??
      "",
    baseUrl:
      query.get("baseUrl") ??
      process.env.NEXT_PUBLIC_API_BASE_URL?.trim() ??
      window.location.origin,
    apiBaseUrl:
      query.get("apiBaseUrl") ??
      process.env.NEXT_PUBLIC_API_BASE_URL?.trim() ??
      undefined,
    externalSource:
      query.get("externalSource") ??
      process.env.NEXT_PUBLIC_OPEN_IM_EXTERNAL_SOURCE?.trim() ??
      undefined,
    title: query.get("title") ?? undefined,
    subtitle: query.get("subtitle") ?? undefined,
    position: (query.get("position") as "left" | "right" | null) ?? undefined,
    themeColor: query.get("themeColor") ?? undefined,
    width: query.get("width") ?? undefined,
    subject: query.get("subject") ?? undefined,
  }

  return window.__CS_AGENT_WIDGET_CONFIG__ ?? window.CSAgentConfig ?? fallback
}

export function setKefuWidgetConfig(config: KefuWidgetHostConfig) {
  if (typeof window === "undefined") {
    return
  }
  window.__CS_AGENT_WIDGET_CONFIG__ = config
}
