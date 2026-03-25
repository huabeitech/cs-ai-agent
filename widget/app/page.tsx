"use client";

import { useEffect, useState } from "react";

import type { WidgetHostConfig } from "@/lib/widget/config";

const STORAGE_KEY = "cs-agent-widget-test-config";

type TestConfig = WidgetHostConfig;

function getWidgetRootPath(pathname: string): string {
  return pathname.startsWith("/widget") ? "/widget" : "";
}

function getWidgetSdkUrl(baseUrl: string, pathname: string): string {
  return `${baseUrl.replace(/\/$/, "")}${getWidgetRootPath(pathname)}/sdk/cs-agent-widget.js`;
}

function generateRandomSubject(): string {
  const uuid =
    typeof crypto !== "undefined" && typeof crypto.randomUUID === "function"
      ? crypto.randomUUID()
      : `${Date.now()}${Math.random().toString(16).slice(2)}`;

  return `用户${uuid.replace(/-/g, "").slice(0, 8)}`;
}

function buildDefaultConfig(baseUrl: string): TestConfig {
  return {
    appId: "",
    baseUrl,
    apiBaseUrl: baseUrl,
    title: "在线客服",
    position: "right",
    themeColor: "#0f6cbd",
    width: "680px",
    subject: generateRandomSubject(),
  };
}

function removeMountedWidget() {
  if (typeof window === "undefined") {
    return;
  }

  document
    .querySelectorAll(
      '[data-cs-agent-widget="launcher"], [data-cs-agent-widget="frame"], [data-cs-agent-widget="script"]',
    )
    .forEach((node) => node.remove());

  delete window.CSAgentConfig;
  delete window.__CS_AGENT_WIDGET_CONFIG__;
  delete (window as Window & { __CS_AGENT_WIDGET_LOADED__?: boolean })
    .__CS_AGENT_WIDGET_LOADED__;
}

function injectWidget(config: TestConfig) {
  removeMountedWidget();
  window.CSAgentConfig = config;

  const script = document.createElement("script");
  script.async = true;
  script.src = getWidgetSdkUrl(window.location.origin, window.location.pathname);
  script.dataset.csAgentWidget = "script";
  document.body.appendChild(script);
}

export default function WidgetTestPage() {
  const [config, setConfig] = useState<TestConfig | null>(null);
  const [status, setStatus] = useState("准备中");

  useEffect(() => {
    const origin = window.location.origin;
    const query = new URLSearchParams(window.location.search);
    const savedText = window.localStorage.getItem(STORAGE_KEY);
    const savedConfig = savedText
      ? (JSON.parse(savedText) as Partial<TestConfig>)
      : {};
    const nextConfig: TestConfig = {
      ...buildDefaultConfig(origin),
      ...savedConfig,
      appId: query.get("appId") ?? savedConfig.appId ?? "",
      baseUrl: query.get("baseUrl") ?? savedConfig.baseUrl ?? origin,
      apiBaseUrl:
        query.get("apiBaseUrl") ??
        savedConfig.apiBaseUrl ??
        savedConfig.baseUrl ??
        origin,
      width: query.get("width") ?? savedConfig.width ?? "680px",
      subject: query.get("subject") ?? savedConfig.subject ?? generateRandomSubject(),
    };
    setConfig(nextConfig);

    if (nextConfig.appId) {
      injectWidget(nextConfig);
      setStatus("Widget 已挂载");
    } else {
      removeMountedWidget();
      setStatus("请先填写 appId");
    }

    return () => {
      removeMountedWidget();
    };
  }, []);

  if (!config) {
    return null;
  }

  const currentConfig = config;

  function updateField<K extends keyof TestConfig>(
    key: K,
    value: TestConfig[K],
  ) {
    setConfig((current) => (current ? { ...current, [key]: value } : current));
  }

  function handleApply() {
    const nextConfig = {
      ...currentConfig,
      appId: currentConfig.appId.trim(),
      baseUrl: currentConfig.baseUrl.trim() || window.location.origin,
      apiBaseUrl:
        currentConfig.apiBaseUrl?.trim() ||
        currentConfig.baseUrl.trim() ||
        window.location.origin,
      title: currentConfig.title?.trim() || "在线客服",
      themeColor: currentConfig.themeColor?.trim() || "#0f6cbd",
      width: currentConfig.width?.trim() || "680px",
    };
    setConfig(nextConfig);
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(nextConfig));

    if (!nextConfig.appId) {
      removeMountedWidget();
      setStatus("请先填写 appId");
      return;
    }

    injectWidget(nextConfig);
    setStatus("Widget 已挂载");
  }

  const sdkUrl = getWidgetSdkUrl(currentConfig.baseUrl, window.location.pathname);

  const snippet = `<script>
  window.CSAgentConfig = {
    appId: "${currentConfig.appId || ""}",
    baseUrl: "${currentConfig.baseUrl}",
    apiBaseUrl: "${currentConfig.apiBaseUrl || currentConfig.baseUrl}",
    title: "${currentConfig.title || "在线客服"}",
    position: "${currentConfig.position || "right"}",
    themeColor: "${currentConfig.themeColor || "#0f6cbd"}",
    width: "${currentConfig.width || "680px"}",
    subject: "${currentConfig.subject || ""}",
  };
</script>
<script async src="${sdkUrl}"></script>`;

  return (
    <main className="min-h-screen px-4 py-6 md:px-6 md:py-7">
      <div className="mx-auto grid w-full max-w-6xl gap-4 lg:grid-cols-[1.15fr_0.85fr]">
        <section className="rounded-lg border border-white/70 bg-white/80 p-4 backdrop-blur md:p-5">
          <div className="mb-4 flex flex-wrap items-start justify-between gap-3">
            <div className="rounded-full border border-slate-200 bg-slate-50 px-3 py-1 text-xs text-slate-600">
              {status}
            </div>
          </div>

          <div className="grid gap-3 md:grid-cols-2">
            <label className="block">
              <div className="mb-1.5 text-xs font-medium text-slate-700">
                appId
              </div>
              <input
                value={currentConfig.appId}
                onChange={(event) => updateField("appId", event.target.value)}
                placeholder="请输入后台嵌入站点 appId"
                className="w-full rounded-lg border border-slate-200 bg-white px-3 py-2.5 text-sm text-slate-900 outline-none transition focus:border-sky-400"
              />
            </label>
            <label className="block">
              <div className="mb-1.5 text-xs font-medium text-slate-700">
                baseUrl
              </div>
              <input
                value={currentConfig.baseUrl}
                onChange={(event) => updateField("baseUrl", event.target.value)}
                placeholder="Widget 地址，例如 http://localhost:3001"
                className="w-full rounded-lg border border-slate-200 bg-white px-3 py-2.5 text-sm text-slate-900 outline-none transition focus:border-sky-400"
              />
            </label>
            <label className="block">
              <div className="mb-1.5 text-xs font-medium text-slate-700">
                apiBaseUrl
              </div>
              <input
                value={currentConfig.apiBaseUrl ?? ""}
                onChange={(event) =>
                  updateField("apiBaseUrl", event.target.value)
                }
                placeholder="后端地址，例如 http://localhost:8080"
                className="w-full rounded-lg border border-slate-200 bg-white px-3 py-2.5 text-sm text-slate-900 outline-none transition focus:border-sky-400"
              />
            </label>
            <label className="block">
              <div className="mb-1.5 text-xs font-medium text-slate-700">
                标题
              </div>
              <input
                value={currentConfig.title ?? ""}
                onChange={(event) => updateField("title", event.target.value)}
                className="w-full rounded-lg border border-slate-200 bg-white px-3 py-2.5 text-sm text-slate-900 outline-none transition focus:border-sky-400"
              />
            </label>
            <label className="block">
              <div className="mb-1.5 text-xs font-medium text-slate-700">
                主题色
              </div>
              <input
                value={currentConfig.themeColor ?? ""}
                onChange={(event) =>
                  updateField("themeColor", event.target.value)
                }
                className="w-full rounded-lg border border-slate-200 bg-white px-3 py-2.5 text-sm text-slate-900 outline-none transition focus:border-sky-400"
              />
            </label>
            <label className="block">
              <div className="mb-1.5 text-xs font-medium text-slate-700">
                宽度
              </div>
              <input
                value={currentConfig.width ?? ""}
                onChange={(event) =>
                  updateField("width", event.target.value)
                }
                placeholder="例如 680px、50vw"
                className="w-full rounded-lg border border-slate-200 bg-white px-3 py-2.5 text-sm text-slate-900 outline-none transition focus:border-sky-400"
              />
            </label>
            <label className="block">
              <div className="mb-1.5 text-xs font-medium text-slate-700">
                会话主题
              </div>
              <input
                value={currentConfig.subject ?? ""}
                onChange={(event) =>
                  updateField("subject", event.target.value)
                }
                placeholder="可选，例如：张三的咨询、订单#12345"
                className="w-full rounded-lg border border-slate-200 bg-white px-3 py-2.5 text-sm text-slate-900 outline-none transition focus:border-sky-400"
              />
            </label>
            <label className="block">
              <div className="mb-1.5 text-xs font-medium text-slate-700">
                悬浮位置
              </div>
              <select
                value={currentConfig.position ?? "right"}
                onChange={(event) =>
                  updateField(
                    "position",
                    event.target.value as "left" | "right",
                  )
                }
                className="w-full rounded-lg border border-slate-200 bg-white px-3 py-2.5 text-sm text-slate-900 outline-none transition focus:border-sky-400"
              >
                <option value="right">右侧</option>
                <option value="left">左侧</option>
              </select>
            </label>
          </div>

          <div className="mt-3 flex flex-wrap gap-2.5">
            <button
              type="button"
              onClick={handleApply}
              className="rounded-md bg-(--primary) px-4 py-2 text-sm text-white transition hover:opacity-92"
            >
              挂载 Widget
            </button>
            <button
              type="button"
              onClick={() => {
                const nextConfig = buildDefaultConfig(window.location.origin);
                setConfig(nextConfig);
                window.localStorage.removeItem(STORAGE_KEY);
                window.localStorage.removeItem(":cs-agent:visitor-id");
                removeMountedWidget();
                setStatus("已重置");
              }}
              className="rounded-md border border-slate-200 bg-white px-4 py-2 text-sm text-slate-700 transition hover:border-slate-300"
            >
              重置
            </button>
          </div>
        </section>

        <section className="rounded-lg border border-slate-200 bg-slate-50/95 p-4 text-slate-800 md:p-5">
          <p className="mt-1 text-sm leading-6 text-slate-600">
            这段脚本就是外部业务系统接入时需要放到宿主页面里的最小配置。
          </p>
          <pre className="mt-3 overflow-x-auto rounded-md border border-slate-200 bg-white p-3 text-xs leading-5 text-slate-700">
            <code>{snippet}</code>
          </pre>
        </section>
      </div>
    </main>
  );
}
