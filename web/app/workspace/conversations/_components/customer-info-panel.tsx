"use client";

import type { AgentConversation } from "@/lib/api/agent";
import { cn } from "@/lib/utils";

import { CustomerTabPanel } from "./customer-tab-panel";

type CustomerInfoPanelProps = {
  conversation: AgentConversation | null;
  className?: string;
  /** 嵌入 Sheet 等容器时使用：去掉左侧栏样式与顶部标题区 */
  variant?: "default" | "embedded";
};

export function CustomerInfoPanel({
  conversation,
  className,
  variant = "default",
}: CustomerInfoPanelProps) {
  const embedded = variant === "embedded";

  return (
    <div
      className={cn(
        "flex h-full min-h-0 flex-col overflow-hidden",
        embedded
          ? "bg-background text-foreground"
          : "border-l border-border bg-card text-card-foreground",
        className,
      )}
    >
      <div className="flex h-12.5 shrink-0 items-center border-b border-border px-3">
        <h2 className="text-sm font-medium text-foreground">会话信息</h2>
      </div>

      <div
        className={cn(
          "min-h-0 flex-1 overflow-y-auto px-3 pb-4",
          embedded && "pb-[max(1rem,env(safe-area-inset-bottom))] pt-1",
        )}
      >
        {!conversation ? (
          <p className="pt-4 text-sm text-muted-foreground">
            {embedded
              ? "请选择会话以查看会话信息"
              : "请选择左侧会话以查看会话信息"}
          </p>
        ) : (
          <CustomerTabPanel conversation={conversation} />
        )}
      </div>
    </div>
  );
}
