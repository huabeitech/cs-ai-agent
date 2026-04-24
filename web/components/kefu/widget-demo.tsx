"use client"

import { useEffect, useMemo, useState } from "react"

import type { KefuWidgetHostConfig } from "@/lib/kefu-widget-config"

const STORAGE_KEY = "cs-agent-web-widget-test-config"
const INITIAL_CONFIG: KefuWidgetHostConfig = {
  channelId: "",
  baseUrl: "",
  apiBaseUrl: "",
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

function getDefaultConfig(): KefuWidgetHostConfig {
  if (typeof window === "undefined") {
    return INITIAL_CONFIG
  }

  const savedText = window.localStorage.getItem(STORAGE_KEY)
  const savedConfig = savedText
    ? (JSON.parse(savedText) as Partial<KefuWidgetHostConfig>)
    : {}
  const query = new URLSearchParams(window.location.search)

  return {
    channelId: query.get("channelId") ?? savedConfig.channelId ?? "",
    baseUrl: "",
    apiBaseUrl: "",
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
  script.src = `${window.location.origin}/sdk/cs-ai-agent-sdk.min.js`
  script.dataset.csAgentWidget = "script"
  document.body.appendChild(script)
}

export function KefuWidgetDemo() {
  const [config, setConfig] = useState<KefuWidgetHostConfig>(INITIAL_CONFIG)
  const [status, setStatus] = useState("请填写 channelId")
  const [origin, setOrigin] = useState("")

  useEffect(() => {
    const timer = window.setTimeout(() => {
      const initialConfig = getDefaultConfig()
      setOrigin(window.location.origin)
      setConfig(initialConfig)
      setStatus(initialConfig.channelId ? "Widget 已挂载" : "请填写 channelId")

      if (initialConfig.channelId) {
        injectWidget(initialConfig)
      }
    }, 0)

    return () => {
      window.clearTimeout(timer)
      removeMountedWidget()
    }
  }, [])

  const snippet = useMemo(() => {
    const scriptSrc = origin
      ? `${origin}/sdk/cs-ai-agent-sdk.min.js`
      : "/sdk/cs-ai-agent-sdk.min.js"

    return `<script>
  window.CSAgentConfig = {
    channelId: "${config.channelId || ""}"
  };
</script>
<script async src="${scriptSrc}"></script>`
  }, [config, origin])

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
      baseUrl: "",
      apiBaseUrl: "",
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
