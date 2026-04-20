"use client";

import dynamic from "next/dynamic";
import { useEffect, useState } from "react";
import {
  ArrowRight,
  Bot,
  ChartColumn,
  Cpu,
  DatabaseZap,
  GitBranch,
  MessageSquareText,
  ShieldCheck,
} from "lucide-react";

import { generateUUID } from "@/lib/utils";
import type { WidgetHostConfig } from "@/lib/widget/config";

const STORAGE_KEY = "cs-agent-widget-test-config";
const GITHUB_URL = "https://github.com/huabeitech/cs-ai-agent";
const GITEE_URL = "https://gitee.com/huabeitech/cs-ai-agent";
const OFFICIAL_SITE_URL = "https://aiagent.huabei.pro";

type TestConfig = WidgetHostConfig;

const featureCards = [
  {
    title: "先让 AI 处理重复路径",
    description: "把常见问题和标准流程优先交给 Agent，降低人工压力。",
    icon: MessageSquareText,
  },
  {
    title: "人工接管不脱离上下文",
    description: "在低置信度、命中规则或用户要求时，带着会话上下文平滑转人工。",
    icon: Bot,
  },
  {
    title: "知识、人工和工单形成闭环",
    description: "把知识检索、人工接管、外部系统调用和工单流程放进同一条客服链路。",
    icon: GitBranch,
  },
  {
    title: "面向客服运营而不是 Demo",
    description: "强调稳定运行、可观测流程和结构化升级，不把项目做成聊天壳子。",
    icon: ShieldCheck,
  },
];

const scopeCards = [
  {
    title: "会话运行时",
    description: "接收用户消息、维护会话状态并触发回复执行。",
    icon: MessageSquareText,
  },
  {
    title: "AI 知识检索",
    description: "基于 RAG 的企业知识库，提升答案准确性。",
    icon: DatabaseZap,
  },
  {
    title: "AI Agent 运行时",
    description: "自定义 Skill，通过 MCP 协议对接企业业务系统。",
    icon: Cpu,
  },
  {
    title: "人工协同",
    description: "在 AI 不该继续时，把会话转给人工客服或指定客服组。",
    icon: Bot,
  },
  {
    title: "工单闭环",
    description: "从会话创建工单，跟踪处理过程，直到最终关闭。",
    icon: ChartColumn,
  },
];

const flowSteps = [
  {
    step: "01",
    title: "用户发起咨询",
    description: "用户通过 Web Widget 或开放接口进入系统，创建新会话或匹配已有会话。",
  },
  {
    step: "02",
    title: "系统装载上下文",
    description: "加载会话历史、AI 配置、可用 Skill、知识库和工具能力，为当前问题准备执行上下文。",
  },
  {
    step: "03",
    title: "AI 优先回复",
    description: "AI Agent 结合历史消息、知识检索结果和可用工具，优先处理常见问题与标准流程。",
  },
  {
    step: "04",
    title: "按需调用能力",
    description: "在回复过程中执行 Skill、通过 MCP 调用外部系统，或触发特定业务流程。",
  },
  {
    step: "05",
    title: "转交人工处理",
    description: "当 AI 无法继续有效处理，或用户希望由人工接待时，会话带着上下文转交客服。",
  },
  {
    step: "06",
    title: "转入工单流程",
    description: "当问题需要跨时段、跨角色持续跟进时，可以从会话直接转入工单流程。",
  },
];

function getWidgetRootPath(pathname: string): string {
  return pathname.startsWith("/widget") ? "/widget" : "";
}

function getWidgetSdkUrl(baseUrl: string, pathname: string): string {
  return `${baseUrl.replace(/\/$/, "")}${getWidgetRootPath(pathname)}/sdk/cs-agent-widget.js`;
}

function generateRandomSubject(): string {
  return `设备会话-${generateUUID().replace(/-/g, "").slice(0, 8)}`;
}

function buildDefaultConfig(baseUrl: string): TestConfig {
  return {
    channelId: "",
    baseUrl,
    apiBaseUrl: baseUrl,
    title: "IoT 智能客服",
    subtitle: "设备告警、工单协同与在线服务统一入口",
    position: "right",
    themeColor: "#2563eb",
    width: "680px",
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

function WidgetHomePageInner() {
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

  function updateField<K extends keyof TestConfig>(
    key: K,
    value: TestConfig[K],
  ) {
    setConfig((current) => ({ ...current, [key]: value }));
  }

  function handleApply() {
    const nextConfig = {
      ...config,
      channelId: config.channelId.trim(),
      baseUrl: config.baseUrl.trim() || window.location.origin,
      apiBaseUrl:
        config.apiBaseUrl?.trim() ||
        config.baseUrl.trim() ||
        window.location.origin,
      title: config.title?.trim() || "IoT 智能客服",
      subtitle: config.subtitle?.trim() || "设备告警、工单协同与在线服务统一入口",
      themeColor: config.themeColor?.trim() || "#2563eb",
      width: config.width?.trim() || "680px",
      subject: config.subject?.trim() || generateRandomSubject(),
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

  const sdkUrl = getWidgetSdkUrl(config.baseUrl, window.location.pathname);
  const snippet = `<script>
  window.CSAgentConfig = {
    channelId: "${config.channelId || ""}",
    baseUrl: "${config.baseUrl}",
    apiBaseUrl: "${config.apiBaseUrl || config.baseUrl}",
    title: "${config.title || "IoT 智能客服"}",
    subtitle: "${config.subtitle || "设备告警、工单协同与在线服务统一入口"}",
    position: "${config.position || "right"}",
    themeColor: "${config.themeColor || "#2563eb"}",
    width: "${config.width || "680px"}",
    subject: "${config.subject || ""}",
  };
</script>
<script async src="${sdkUrl}"></script>`;

  return (
    <main className="min-h-screen overflow-y-auto overflow-x-hidden scroll-smooth">
      <div className="relative">
        <div className="pointer-events-none absolute inset-x-0 top-0 -z-10 h-[24rem] bg-[radial-gradient(circle_at_top,rgba(37,99,235,0.14),transparent_46%)]" />

        <header className="sticky top-0 z-20 border-b border-slate-200/80 bg-white/90 backdrop-blur-xl">
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
                className="h-10 w-10 rounded-2xl object-contain ring-1 ring-slate-200"
              />
              <span>
                <span className="block text-base font-semibold">
                  贝壳 AI 客服
                </span>
              </span>
            </a>

            <nav className="hidden items-center gap-8 text-sm text-slate-600 md:flex">
              <a href="#features" className="transition hover:text-slate-950">
                产品能力
              </a>
              <a
                href={GITHUB_URL}
                target="_blank"
                rel="noopener noreferrer"
                className="transition hover:text-slate-950"
              >
                GitHub
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
              href="#integration"
              className="inline-flex items-center gap-2 rounded-full bg-blue-600 px-4 py-2 text-sm font-medium text-white transition hover:bg-blue-700"
            >
              立即体验
              <ArrowRight className="h-4 w-4" />
            </a>
          </div>
        </header>

        <section className="px-4 pb-18 pt-16 md:px-6 md:pb-24 md:pt-20">
          <div className="mx-auto flex max-w-4xl flex-col items-center text-center">
            <h1 className="mt-6 text-4xl font-semibold tracking-[-0.05em] text-slate-950 md:text-6xl">
              AI 优先接待，人工无缝协同
            </h1>

            <p className="mt-6 max-w-2xl text-lg leading-8 text-slate-600">
              贝壳 AI 客服是一个以 AI Agent 为核心的智能客服系统，把在线会话、
              知识库检索、工单流转和人工接管串成一条完整链路。
            </p>

            <div className="mt-8 flex flex-wrap items-center justify-center gap-3">
              <a
                href="#integration"
                className="inline-flex items-center gap-2 rounded-full bg-blue-600 px-5 py-3 text-sm font-medium text-white transition hover:bg-blue-700"
              >
                开始接入
                <ArrowRight className="h-4 w-4" />
              </a>
              <a
                href={OFFICIAL_SITE_URL}
                target="_blank"
                rel="noopener noreferrer"
                className="rounded-full border border-slate-200 bg-white px-5 py-3 text-sm font-medium text-slate-700 transition hover:border-slate-300 hover:text-slate-950"
              >
                查看官网
              </a>
            </div>

            <div className="mt-12 grid w-full gap-4 md:grid-cols-3">
              {[
                { label: "接待模式", value: "AI + 人工协同" },
                { label: "回答方式", value: "知识库 RAG 驱动" },
                { label: "运行能力", value: "Skills + MCP + 工单流程" },
              ].map((item) => (
                <div
                  key={item.label}
                  className="rounded-lg border border-white/80 bg-white/90 p-6 text-left shadow-[0_18px_40px_rgba(15,23,42,0.05)]"
                >
                  <div className="text-sm text-slate-500">{item.label}</div>
                  <div className="mt-2 text-xl font-semibold text-slate-950">
                    {item.value}
                  </div>
                </div>
              ))}
            </div>
          </div>
        </section>

        <section id="features" className="px-4 py-8 md:px-6 md:py-10">
          <div className="mx-auto max-w-6xl">
            <div className="max-w-2xl">
              <div className="text-sm font-medium text-blue-700">产品能力</div>
              <h2 className="mt-2 text-3xl font-semibold tracking-[-0.03em] text-slate-950">
                面向真实客服场景的开源系统
              </h2>
              <p className="mt-4 text-base leading-8 text-slate-600">
                用一套系统处理用户咨询、AI 回复、知识检索、人工接管和后续工单。
              </p>
            </div>

            <div className="mt-8 grid gap-4 md:grid-cols-2 xl:grid-cols-4">
              {featureCards.map((item) => {
                const Icon = item.icon;
                return (
                  <article
                    key={item.title}
                    className="rounded-lg border border-slate-200 bg-white p-6 shadow-[0_18px_40px_rgba(15,23,42,0.04)]"
                  >
                    <div className="inline-flex rounded-2xl bg-blue-50 p-3 text-blue-700">
                      <Icon className="h-5 w-5" />
                    </div>
                    <h3 className="mt-4 text-lg font-semibold text-slate-950">
                      {item.title}
                    </h3>
                    <p className="mt-3 text-sm leading-7 text-slate-600">
                      {item.description}
                    </p>
                  </article>
                );
              })}
            </div>
          </div>
        </section>

        <section className="px-4 py-16 md:px-6">
          <div className="mx-auto grid max-w-6xl gap-6 lg:grid-cols-[0.95fr_1.05fr]">
            <div className="rounded-lg border border-slate-200 bg-white p-7 shadow-[0_24px_60px_rgba(15,23,42,0.06)]">
              <div className="mt-2 text-2xl font-semibold tracking-[-0.03em] text-slate-950">
                覆盖完整客服链路
              </div>
              <p className="mt-4 text-base leading-8 text-slate-600">
                从第一条消息到问题关闭，关键环节都在同一套系统里。
              </p>

              <div className="mt-6 grid gap-4 md:grid-cols-1">
                {scopeCards.map((item) => {
                  const Icon = item.icon;
                  return (
                    <div
                      key={item.title}
                      className="rounded-lg border border-slate-200 bg-slate-50 p-5"
                    >
                      <div className="flex items-start gap-4">
                        <div className="rounded-2xl bg-blue-50 p-3 text-blue-700">
                          <Icon className="h-5 w-5" />
                        </div>
                        <div>
                          <div className="text-base font-semibold text-slate-950">
                            {item.title}
                          </div>
                          <div className="mt-2 text-sm leading-7 text-slate-600">
                            {item.description}
                          </div>
                        </div>
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>

            <div className="rounded-lg border border-slate-200 bg-gradient-to-br from-slate-950 via-slate-900 to-blue-950 p-7 text-white shadow-[0_24px_60px_rgba(15,23,42,0.12)]">
              <div className="mt-2 text-2xl font-semibold tracking-[-0.03em]">
                客服处理闭环
              </div>
              <p className="mt-4 text-sm leading-7 text-slate-300">
                从用户发起咨询，到 AI 回复、人工接管、工单跟踪，核心流程都在同一套系统里完成。
              </p>

              <div className="mt-6 space-y-4">
                {flowSteps.map((item) => (
                  <div
                    key={item.step}
                    className="flex items-start gap-4 rounded-lg border border-white/10 bg-white/5 p-5"
                  >
                    <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-white text-sm font-semibold text-slate-950">
                      {item.step}
                    </div>
                    <div>
                      <div className="text-sm font-semibold text-white">
                        {item.title}
                      </div>
                      <div className="mt-2 text-sm leading-7 text-slate-200">
                        {item.description}
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </section>

        <section id="integration" className="px-4 py-16 md:px-6 md:py-20">
          <div className="mx-auto max-w-6xl">
            <div className="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
              <div className="max-w-2xl">
                <div className="text-sm font-medium text-blue-700">接入方式</div>
                <h2 className="mt-2 text-3xl font-semibold tracking-[-0.03em] text-slate-950">
                  保留在线配置与接入脚本，直接拿来演示和集成
                </h2>
                <p className="mt-4 text-base leading-8 text-slate-600">
                  这里可以直接填写 `channelId`、调整宿主参数并挂载 Widget，也可以复制脚本到业务系统完成基础接入。
                </p>
              </div>
            </div>

            <div className="mt-8 grid gap-6 xl:grid-cols-[0.94fr_1.06fr]">
              <section className="rounded-lg border border-slate-200 bg-white p-6 shadow-[0_24px_60px_rgba(15,23,42,0.06)]">
                <div className="text-lg font-semibold text-slate-950">Widget 配置</div>
                <div className="mt-1 text-sm text-slate-500">
                  宿主页面接入时需要的基础参数。
                </div>

                <div className="mt-6 grid gap-4 md:grid-cols-2">
                  <label className="block">
                    <div className="mb-2 text-xs font-medium tracking-[0.18em] text-slate-500">
                      channelId
                    </div>
                    <input
                      value={config.channelId}
                      onChange={(event) => updateField("channelId", event.target.value)}
                      placeholder="请输入后台渠道 channelId"
                      className="w-full rounded-2xl border border-slate-200 bg-white px-4 py-3 text-sm text-slate-900 outline-none transition focus:border-blue-400"
                    />
                  </label>
                  <label className="block">
                    <div className="mb-2 text-xs font-medium tracking-[0.18em] text-slate-500">
                      baseUrl
                    </div>
                    <input
                      value={config.baseUrl}
                      onChange={(event) => updateField("baseUrl", event.target.value)}
                      placeholder="Widget 地址，例如 http://localhost:3001"
                      className="w-full rounded-2xl border border-slate-200 bg-white px-4 py-3 text-sm text-slate-900 outline-none transition focus:border-blue-400"
                    />
                  </label>
                  <label className="block">
                    <div className="mb-2 text-xs font-medium tracking-[0.18em] text-slate-500">
                      apiBaseUrl
                    </div>
                    <input
                      value={config.apiBaseUrl ?? ""}
                      onChange={(event) => updateField("apiBaseUrl", event.target.value)}
                      placeholder="后端地址，例如 http://localhost:8080"
                      className="w-full rounded-2xl border border-slate-200 bg-white px-4 py-3 text-sm text-slate-900 outline-none transition focus:border-blue-400"
                    />
                  </label>
                  <label className="block">
                    <div className="mb-2 text-xs font-medium tracking-[0.18em] text-slate-500">
                      标题
                    </div>
                    <input
                      value={config.title ?? ""}
                      onChange={(event) => updateField("title", event.target.value)}
                      className="w-full rounded-2xl border border-slate-200 bg-white px-4 py-3 text-sm text-slate-900 outline-none transition focus:border-blue-400"
                    />
                  </label>
                  <label className="block">
                    <div className="mb-2 text-xs font-medium tracking-[0.18em] text-slate-500">
                      副标题
                    </div>
                    <input
                      value={config.subtitle ?? ""}
                      onChange={(event) => updateField("subtitle", event.target.value)}
                      placeholder="例如：设备告警、工单协同与在线服务统一入口"
                      className="w-full rounded-2xl border border-slate-200 bg-white px-4 py-3 text-sm text-slate-900 outline-none transition focus:border-blue-400"
                    />
                  </label>
                  <label className="block">
                    <div className="mb-2 text-xs font-medium tracking-[0.18em] text-slate-500">
                      主题色
                    </div>
                    <input
                      value={config.themeColor ?? ""}
                      onChange={(event) => updateField("themeColor", event.target.value)}
                      className="w-full rounded-2xl border border-slate-200 bg-white px-4 py-3 text-sm text-slate-900 outline-none transition focus:border-blue-400"
                    />
                  </label>
                  <label className="block">
                    <div className="mb-2 text-xs font-medium tracking-[0.18em] text-slate-500">
                      宽度
                    </div>
                    <input
                      value={config.width ?? ""}
                      onChange={(event) => updateField("width", event.target.value)}
                      placeholder="例如 680px、50vw"
                      className="w-full rounded-2xl border border-slate-200 bg-white px-4 py-3 text-sm text-slate-900 outline-none transition focus:border-blue-400"
                    />
                  </label>
                  <label className="block">
                    <div className="mb-2 text-xs font-medium tracking-[0.18em] text-slate-500">
                      会话主题
                    </div>
                    <input
                      value={config.subject ?? ""}
                      onChange={(event) => updateField("subject", event.target.value)}
                      placeholder="例如：站点A-储能柜-001"
                      className="w-full rounded-2xl border border-slate-200 bg-white px-4 py-3 text-sm text-slate-900 outline-none transition focus:border-blue-400"
                    />
                  </label>
                  <label className="block">
                    <div className="mb-2 text-xs font-medium tracking-[0.18em] text-slate-500">
                      悬浮位置
                    </div>
                    <select
                      value={config.position ?? "right"}
                      onChange={(event) =>
                        updateField(
                          "position",
                          event.target.value as "left" | "right",
                        )
                      }
                      className="w-full rounded-2xl border border-slate-200 bg-white px-4 py-3 text-sm text-slate-900 outline-none transition focus:border-blue-400"
                    >
                      <option value="right">右侧</option>
                      <option value="left">左侧</option>
                    </select>
                  </label>
                </div>

                <div className="mt-6 flex flex-wrap gap-3">
                  <button
                    type="button"
                    onClick={handleApply}
                    className="rounded-full bg-blue-600 px-5 py-3 text-sm font-medium text-white transition hover:bg-blue-700"
                  >
                    挂载 Widget
                  </button>
                  <button
                    type="button"
                    onClick={() => {
                      const nextConfig = {
                        ...buildDefaultConfig(window.location.origin),
                        subject: generateRandomSubject(),
                      };
                      setConfig(nextConfig);
                      window.localStorage.setItem(STORAGE_KEY, JSON.stringify(nextConfig));
                      window.localStorage.removeItem("cs-agent:external-id");
                      removeMountedWidget();
                      setStatus("已重置");
                    }}
                    className="rounded-full border border-slate-200 bg-white px-5 py-3 text-sm font-medium text-slate-700 transition hover:border-slate-300"
                  >
                    重置配置
                  </button>
                </div>
              </section>

              <section className="rounded-lg border border-slate-200 bg-white p-6 shadow-[0_24px_60px_rgba(15,23,42,0.06)]">
                <div className="text-lg font-semibold text-slate-950">接入脚本</div>
                <div className="mt-1 text-sm text-slate-500">
                  复制到宿主页面即可完成基础接入。
                </div>
                <pre className="cs-agent-scrollbar mt-6 overflow-x-auto rounded-lg border border-slate-200 bg-slate-950 p-5 text-xs leading-6 text-slate-100">
                  <code>{snippet}</code>
                </pre>
                
              </section>
            </div>
          </div>
        </section>

        <footer className="border-t border-slate-200 bg-white px-4 py-8 md:px-6">
          <div className="mx-auto flex max-w-6xl flex-col gap-4 text-sm text-slate-500 md:flex-row md:items-center md:justify-between">
            <div>贝壳 AI 客服</div>
            <div className="flex flex-wrap items-center gap-5">
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

const WidgetHomePage = dynamic(async () => WidgetHomePageInner, {
  ssr: false,
});

export default WidgetHomePage;
