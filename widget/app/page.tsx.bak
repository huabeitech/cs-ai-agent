"use client";

import dynamic from "next/dynamic";
import { useEffect, useState } from "react";

import { generateUUID } from "@/lib/utils";
import type { WidgetHostConfig } from "@/lib/widget/config";

const STORAGE_KEY = "cs-agent-widget-test-config";
const GITHUB_URL = "https://github.com/huabeitech/cs-ai-agent";
const GITEE_URL = "https://gitee.com/huabeitech/cs-ai-agent";
const OFFICIAL_SITE_URL = "https://aiagent.huabei.pro";

type TestConfig = WidgetHostConfig;

function getWidgetRootPath(pathname: string): string {
  return pathname.startsWith("/widget") ? "/widget" : "";
}

function getWidgetSdkUrl(baseUrl: string, pathname: string): string {
  return `${baseUrl.replace(/\/$/, "")}${getWidgetRootPath(pathname)}/sdk/cs-agent-widget.js`;
}

function generateRandomSubject(): string {
  return `用户${generateUUID().replace(/-/g, "").slice(0, 8)}`;
}

function buildDefaultConfig(baseUrl: string): TestConfig {
  return {
    channelId: "",
    baseUrl,
    apiBaseUrl: baseUrl,
    title: "在线客服",
    subtitle: "贝壳AI客服为您服务",
    position: "right",
    themeColor: "#0f6cbd",
    width: "880px",
    subject: generateRandomSubject(),
  };
}

function getInitialConfig(): TestConfig {
  if (typeof window === "undefined") {
    return buildDefaultConfig("");
  }

  const origin = window.location.origin;
  const query = new URLSearchParams(window.location.search);
  const savedText = window.localStorage.getItem(STORAGE_KEY);
  const savedConfig = savedText
    ? (JSON.parse(savedText) as Partial<TestConfig>)
    : {};

  return {
    ...buildDefaultConfig(origin),
    ...savedConfig,
    channelId: query.get("channelId") ?? savedConfig.channelId ?? "",
    baseUrl: query.get("baseUrl") ?? savedConfig.baseUrl ?? origin,
    apiBaseUrl:
      query.get("apiBaseUrl") ??
      savedConfig.apiBaseUrl ??
      savedConfig.baseUrl ??
      origin,
    width: query.get("width") ?? savedConfig.width ?? "680px",
    subject:
      query.get("subject") ?? savedConfig.subject ?? generateRandomSubject(),
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

function WidgetTestPageInner() {
  const [config, setConfig] = useState<TestConfig>(getInitialConfig);
  const [status, setStatus] = useState(() =>
    getInitialConfig().channelId ? "Widget 已挂载" : "请先填写 channelId",
  );

  useEffect(() => {
    if (config.channelId) {
      injectWidget(config);
    } else {
      removeMountedWidget();
    }

    return () => {
      removeMountedWidget();
    };
  }, [config]);

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
      channelId: currentConfig.channelId.trim(),
      baseUrl: currentConfig.baseUrl.trim() || window.location.origin,
      apiBaseUrl:
        currentConfig.apiBaseUrl?.trim() ||
        currentConfig.baseUrl.trim() ||
        window.location.origin,
      title: currentConfig.title?.trim() || "在线客服",
      subtitle:
        currentConfig.subtitle?.trim() || "",
      themeColor: currentConfig.themeColor?.trim() || "#0f6cbd",
      width: currentConfig.width?.trim() || "680px",
    };
    setConfig(nextConfig);
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(nextConfig));

    if (!nextConfig.channelId) {
      removeMountedWidget();
      setStatus("请先填写 channelId");
      return;
    }

    injectWidget(nextConfig);
    setStatus("Widget 已挂载");
  }

  const sdkUrl = getWidgetSdkUrl(currentConfig.baseUrl, window.location.pathname);

  const snippet = `<script>
  window.CSAgentConfig = {
    channelId: "${currentConfig.channelId || ""}",
    baseUrl: "${currentConfig.baseUrl}",
    apiBaseUrl: "${currentConfig.apiBaseUrl || currentConfig.baseUrl}",
    title: "${currentConfig.title || "在线客服"}",
    subtitle: "${currentConfig.subtitle || ""}",
    position: "${currentConfig.position || "right"}",
    themeColor: "${currentConfig.themeColor || "#0f6cbd"}",
    width: "${currentConfig.width || "680px"}",
    subject: "${currentConfig.subject || ""}",
  };
</script>
<script async src="${sdkUrl}"></script>`;

  return (
    <main className="h-screen overflow-y-auto overflow-x-hidden scroll-smooth">
      <div className="relative isolate">
        <div className="pointer-events-none absolute inset-x-0 top-0 -z-10 h-[520px] bg-[radial-gradient(circle_at_top,rgba(59,130,246,0.18),transparent_45%),radial-gradient(circle_at_20%_20%,rgba(14,165,233,0.10),transparent_30%)]" />

        <header className="sticky top-0 z-20 border-b border-white/60 bg-white/75 backdrop-blur-xl">
          <div className="mx-auto flex max-w-6xl items-center justify-between px-4 py-4 md:px-6">
            <a
              href={OFFICIAL_SITE_URL}
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center gap-3 text-slate-950"
            >
              <img
                src="/images/logo.png"
                alt="贝壳 AI 客服"
                className="h-11 w-11 rounded-2xl object-contain shadow-[0_14px_30px_rgba(15,23,42,0.18)]"
              />
              <span>
                <span className="block text-base font-semibold">
                  贝壳 AI 客服
                </span>
              </span>
            </a>

            <nav className="hidden items-center gap-7 text-sm text-slate-600 md:flex">
              <a
                href={GITHUB_URL}
                target="_blank"
                rel="noopener noreferrer"
                className="transition hover:text-slate-950"
              >
                GitHub
              </a>
              <a
                href={GITEE_URL}
                target="_blank"
                rel="noopener noreferrer"
                className="transition hover:text-slate-950"
              >
                Gitee
              </a>
              <a
                href={OFFICIAL_SITE_URL}
                target="_blank"
                rel="noopener noreferrer"
                className="transition hover:text-slate-950"
              >
                官网
              </a>
            </nav>

            <a
              href={OFFICIAL_SITE_URL}
              target="_blank"
              rel="noopener noreferrer"
              className="rounded-full bg-slate-950 px-4 py-2 text-sm font-medium text-white transition hover:bg-slate-800"
            >
              访问官网
            </a>
          </div>
        </header>

        <section id="playground" className="px-4 pb-8 pt-8 md:px-6 md:pb-10 md:pt-10">
          <div className="mx-auto grid w-full max-w-6xl gap-4 lg:grid-cols-[1.15fr_0.85fr]">
            <section className="rounded-lg border border-white/70 bg-white/82 p-5 shadow-[0_24px_80px_rgba(15,23,42,0.08)] backdrop-blur md:p-6">
              <div className="mb-5 flex flex-wrap items-start justify-between gap-3">
                <div>
                  <div className="mt-2 text-2xl font-semibold tracking-[-0.03em] text-slate-950 md:text-3xl">
                    插件配置
                  </div>
                  <div className="mt-1 text-sm text-slate-500">
                    这里是首页主体。先填渠道和展示参数，再直接挂载 Widget 验证效果。
                  </div>
                </div>
                {/* <div className="rounded-full border border-slate-200 bg-slate-50 px-3 py-1 text-xs text-slate-600">
                  {status}
                </div> */}
              </div>

              <div className="grid gap-3 md:grid-cols-2">
                <label className="block">
                  <div className="mb-1.5 text-xs font-medium text-slate-700">
                    channelId
                  </div>
                  <input
                    value={currentConfig.channelId}
                    onChange={(event) => updateField("channelId", event.target.value)}
                    placeholder="请输入后台渠道 channelId"
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
                    副标题
                  </div>
                  <input
                    value={currentConfig.subtitle ?? ""}
                    onChange={(event) =>
                      updateField("subtitle", event.target.value)
                    }
                    placeholder="例如：通常几分钟内回复，支持连续会话记录"
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
                    placeholder="例如 880px、50vw"
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

              <div className="mt-4 flex flex-wrap gap-2.5">
                <button
                  type="button"
                  onClick={handleApply}
                  className="rounded-full bg-(--primary) px-5 py-2.5 text-sm font-medium text-white transition hover:opacity-92"
                >
                  挂载 Widget
                </button>
                <button
                  type="button"
                  onClick={() => {
                    const nextConfig = {
                      ...currentConfig,
                      subject: generateRandomSubject(),
                    };
                    setConfig(nextConfig);
                    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(nextConfig));
                    window.localStorage.removeItem("cs-agent:external-id");
                    removeMountedWidget();
                    setStatus("已重置");
                  }}
                  className="rounded-full border border-slate-200 bg-white px-5 py-2.5 text-sm font-medium text-slate-700 transition hover:border-slate-300"
                >
                  重置
                </button>
              </div>
            </section>

            <section
              id="integration"
              className="rounded-lg border border-slate-200 bg-slate-50/95 p-5 text-slate-800 shadow-[0_24px_80px_rgba(15,23,42,0.06)] md:p-6"
            >
              <div className="mt-2 text-2xl font-semibold tracking-[-0.03em] text-slate-950 md:text-3xl">
                宿主页面接入脚本
              </div>
              <div className="mt-1 text-sm text-slate-500">
                这段脚本就是外部业务系统接入时需要放到宿主页面里的最小配置。 Widget 验证效果。
              </div>
              <pre className="mt-4 overflow-x-auto rounded-2xl border border-slate-200 bg-white p-4 text-xs leading-5 text-slate-700">
                <code>{snippet}</code>
              </pre>
            </section>
          </div>
        </section>


        <footer className="border-t border-white/60 bg-white/70 px-4 py-8 backdrop-blur-xl md:px-6">
          <div className="mx-auto flex max-w-6xl flex-col gap-6 md:flex-row md:items-end md:justify-between">
            <div className="flex flex-wrap gap-6 text-sm text-slate-500">
              <a
                href={GITHUB_URL}
                target="_blank"
                rel="noopener noreferrer"
                className="transition hover:text-slate-900"
              >
                GitHub
              </a>
              <a
                href={GITEE_URL}
                target="_blank"
                rel="noopener noreferrer"
                className="transition hover:text-slate-900"
              >
                Gitee
              </a>
              <a
                href={OFFICIAL_SITE_URL}
                target="_blank"
                rel="noopener noreferrer"
                className="transition hover:text-slate-900"
              >
                官网
              </a>
            </div>
          </div>
        </footer>
      </div>
    </main>
  );
}

const WidgetTestPage = dynamic(async () => WidgetTestPageInner, {
  ssr: false,
});

export default WidgetTestPage;
