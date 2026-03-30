export type WidgetHostConfig = {
  appId: string;
  baseUrl: string;
  apiBaseUrl?: string;
  /** 与后端 enums.ExternalSource 一致，默认 web_chat */
  externalSource?: string;
  title?: string;
  subtitle?: string;
  position?: "left" | "right";
  themeColor?: string;
  width?: string;
  /** 访客展示名，随请求以 X-External-Name / WS query externalName 传给后端作 ExternalName */
  subject?: string;
};

declare global {
  interface Window {
    CSAgentConfig?: WidgetHostConfig;
    __CS_AGENT_WIDGET_CONFIG__?: WidgetHostConfig;
  }
}

export function readWidgetConfig(): WidgetHostConfig {
  if (typeof window === "undefined") {
    return {
      appId: "",
      baseUrl: "",
      apiBaseUrl: "",
    };
  }
  const query = new URLSearchParams(window.location.search);
  const fallback = {
    appId: query.get("appId") ?? "",
    baseUrl: query.get("baseUrl") ?? "",
    apiBaseUrl: query.get("apiBaseUrl") ?? undefined,
    externalSource: query.get("externalSource") ?? undefined,
    title: query.get("title") ?? undefined,
    subtitle: query.get("subtitle") ?? undefined,
    position: (query.get("position") as "left" | "right" | null) ?? undefined,
    themeColor: query.get("themeColor") ?? undefined,
    width: query.get("width") ?? undefined,
    subject: query.get("subject") ?? undefined,
  };
  return window.__CS_AGENT_WIDGET_CONFIG__ ?? window.CSAgentConfig ?? fallback;
}

export function setWidgetConfig(config: WidgetHostConfig) {
  if (typeof window === "undefined") {
    return;
  }
  window.__CS_AGENT_WIDGET_CONFIG__ = config;
}
