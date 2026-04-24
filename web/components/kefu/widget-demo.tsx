"use client"

import { useEffect, useMemo, useState } from "react"

import type { KefuWidgetHostConfig } from "@/lib/kefu-widget-config"
import { generateUUID } from "@/lib/utils"

const STORAGE_KEY = "cs-agent-web-widget-test-config"
const INITIAL_CONFIG: KefuWidgetHostConfig = {
  channelId: "",
  baseUrl: "",
  apiBaseUrl: "",
  externalSource: "web_chat",
  title: "在线客服",
  subtitle: "欢迎咨询",
  position: "right",
  themeColor: "#2563eb",
  width: "680px",
  subject: "",
}

declare global {
  interface Window {
    CSAgentWidget?: {
      mount: (config: KefuWidgetHostConfig) => void
      destroy: () => void
      open: () => void
      close: () => void
    }
  }
}

function generateRandomSubject() {
  return `访客-${generateUUID().replace(/-/g, "").slice(0, 8)}`
}

function getDefaultConfig(): KefuWidgetHostConfig {
  if (typeof window === "undefined") {
    return INITIAL_CONFIG
  }

  const savedText = window.localStorage.getItem(STORAGE_KEY)
  const savedConfig = savedText
    ? (JSON.parse(savedText) as Partial<KefuWidgetHostConfig>)
    : {}
  const query = new URLSearchParams(window.location.search)
  const origin = window.location.origin

  return {
    channelId: query.get("channelId") ?? savedConfig.channelId ?? "",
    baseUrl: query.get("baseUrl") ?? savedConfig.baseUrl ?? origin,
    apiBaseUrl:
      query.get("apiBaseUrl") ??
      savedConfig.apiBaseUrl ??
      savedConfig.baseUrl ??
      origin,
    externalSource:
      query.get("externalSource") ?? savedConfig.externalSource ?? "web_chat",
    title: query.get("title") ?? savedConfig.title ?? "在线客服",
    subtitle: query.get("subtitle") ?? savedConfig.subtitle ?? "欢迎咨询",
    position:
      (query.get("position") as "left" | "right" | null) ??
      savedConfig.position ??
      "right",
    themeColor: query.get("themeColor") ?? savedConfig.themeColor ?? "#2563eb",
    width: query.get("width") ?? savedConfig.width ?? "680px",
    subject: query.get("subject") ?? savedConfig.subject ?? generateRandomSubject(),
  }
}

function removeMountedWidget() {
  if (typeof window === "undefined") {
    return
  }

  window.CSAgentWidget?.destroy()
  document
    .querySelectorAll(
      '[data-cs-agent-widget="launcher"], [data-cs-agent-widget="frame"], [data-cs-agent-widget="script"]'
    )
    .forEach((node) => node.remove())

  delete window.CSAgentConfig
  delete window.__CS_AGENT_WIDGET_CONFIG__
  delete window.__CS_AGENT_WIDGET_STATE__
  delete window.CSAgentWidget
}

function injectWidget(config: KefuWidgetHostConfig) {
  removeMountedWidget()
  window.CSAgentConfig = config

  const script = document.createElement("script")
  script.async = true
  script.src = `${window.location.origin}/sdk/cs-agent-widget.js`
  script.dataset.csAgentWidget = "script"
  document.body.appendChild(script)
}

export function KefuWidgetDemo() {
  const [config, setConfig] = useState<KefuWidgetHostConfig>(INITIAL_CONFIG)
  const [status, setStatus] = useState("请填写 channelId")

  useEffect(() => {
    const initialConfig = getDefaultConfig()
    setConfig(initialConfig)
    setStatus(initialConfig.channelId ? "Widget 已挂载" : "请填写 channelId")

    if (initialConfig.channelId) {
      injectWidget(initialConfig)
    }

    return () => {
      removeMountedWidget()
    }
  }, [])

  const snippet = useMemo(() => {
    const scriptSrc = config.baseUrl
      ? `${config.baseUrl.replace(/\/$/, "")}/sdk/cs-agent-widget.js`
      : "/sdk/cs-agent-widget.js"

    return `<script>
  window.CSAgentConfig = {
    channelId: "${config.channelId || ""}",
    baseUrl: "${config.baseUrl || ""}",
    apiBaseUrl: "${config.apiBaseUrl || config.baseUrl || ""}",
    externalSource: "${config.externalSource || "web_chat"}",
    title: "${config.title || "在线客服"}",
    subtitle: "${config.subtitle || ""}",
    position: "${config.position || "right"}",
    themeColor: "${config.themeColor || "#2563eb"}",
    width: "${config.width || "380px"}",
    subject: "${config.subject || ""}",
  };
</script>
<script async src="${scriptSrc}"></script>`
  }, [config])

  function updateField<K extends keyof KefuWidgetHostConfig>(
    key: K,
    value: KefuWidgetHostConfig[K]
  ) {
    setConfig((current) => ({ ...current, [key]: value }))
  }

  function handleMount() {
    const nextConfig: KefuWidgetHostConfig = {
      ...config,
      channelId: config.channelId.trim(),
      baseUrl: config.baseUrl.trim() || window.location.origin,
      apiBaseUrl:
        config.apiBaseUrl?.trim() || config.baseUrl.trim() || window.location.origin,
      externalSource: config.externalSource?.trim() || "web_chat",
      title: config.title?.trim() || "在线客服",
      subtitle: config.subtitle?.trim() || "",
      themeColor: config.themeColor?.trim() || "#2563eb",
      width: config.width?.trim() || "380px",
      subject: config.subject?.trim() || generateRandomSubject(),
    }

    setConfig(nextConfig)
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(nextConfig))

    if (!nextConfig.channelId) {
      removeMountedWidget()
      setStatus("请填写 channelId")
      return
    }

    injectWidget(nextConfig)
    setStatus("Widget 已挂载")
  }

  return (
    <main className="min-h-svh bg-slate-50 px-6 py-8 text-slate-950">
      <div className="mx-auto grid max-w-6xl gap-6 lg:grid-cols-[360px_minmax(0,1fr)]">
        <section className="rounded-lg border border-slate-200 bg-white p-5 shadow-sm">
          <div className="text-base font-semibold">Widget 挂载测试</div>
          <div className="mt-1 text-sm text-slate-500">{status}</div>

          <div className="mt-5 grid gap-3">
            <TextField
              label="channelId"
              value={config.channelId}
              onChange={(value) => updateField("channelId", value)}
            />
            <TextField
              label="baseUrl"
              value={config.baseUrl}
              onChange={(value) => updateField("baseUrl", value)}
            />
            <TextField
              label="apiBaseUrl"
              value={config.apiBaseUrl || ""}
              onChange={(value) => updateField("apiBaseUrl", value)}
            />
            <TextField
              label="title"
              value={config.title || ""}
              onChange={(value) => updateField("title", value)}
            />
            <TextField
              label="subtitle"
              value={config.subtitle || ""}
              onChange={(value) => updateField("subtitle", value)}
            />
            <TextField
              label="themeColor"
              value={config.themeColor || ""}
              onChange={(value) => updateField("themeColor", value)}
            />
            <TextField
              label="width"
              value={config.width || ""}
              onChange={(value) => updateField("width", value)}
            />
            <TextField
              label="subject"
              value={config.subject || ""}
              onChange={(value) => updateField("subject", value)}
            />
          </div>

          <div className="mt-5 flex gap-2">
            <button
              type="button"
              onClick={handleMount}
              className="rounded-md bg-slate-950 px-4 py-2 text-sm font-medium text-white"
            >
              挂载
            </button>
            <button
              type="button"
              onClick={() => window.CSAgentWidget?.open()}
              className="rounded-md border border-slate-200 bg-white px-4 py-2 text-sm font-medium"
            >
              打开
            </button>
            <button
              type="button"
              onClick={() => {
                removeMountedWidget()
                setStatus("Widget 已卸载")
              }}
              className="rounded-md border border-slate-200 bg-white px-4 py-2 text-sm font-medium"
            >
              卸载
            </button>
          </div>
        </section>

        <section className="rounded-lg border border-slate-200 bg-white p-5 shadow-sm">
          <div className="text-base font-semibold">接入代码</div>
          <pre className="mt-4 overflow-x-auto rounded-md bg-slate-950 p-4 text-xs leading-5 text-slate-100">
            <code>{snippet}</code>
          </pre>
        </section>
      </div>
    </main>
  )
}

function TextField({
  label,
  value,
  onChange,
}: {
  label: string
  value?: string
  onChange: (value: string) => void
}) {
  return (
    <label className="grid gap-1.5 text-sm">
      <span className="font-medium text-slate-700">{label}</span>
      <input
        value={value || ""}
        onChange={(event) => onChange(event.target.value)}
        className="h-9 rounded-md border border-slate-200 px-3 text-sm outline-none focus:border-slate-400"
      />
    </label>
  )
}
