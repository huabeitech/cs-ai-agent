"use client";

import { useState } from "react";

import type { AgentConversation } from "@/lib/api/agent";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  IMConversationServiceMode,
  IMConversationServiceModeLabels,
  IMConversationStatus,
  IMConversationStatusLabels,
} from "@/lib/generated/enums";
import { cn, formatDateTime } from "@/lib/utils";

import { CustomerTabPanel } from "./customer-tab-panel";

const infoTabOptions = [
  { value: "conversation", label: "会话" },
  { value: "customer", label: "客户" },
] as const;

type InfoTabValue = (typeof infoTabOptions)[number]["value"];

function InfoRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center gap-0.5 py-2">
      <span className="text-sm text-muted-foreground">{label}:&nbsp;</span>
      <span className="break-all text-sm text-foreground">{value || "-"}</span>
    </div>
  );
}

function ConversationDetails({
  conversation,
}: {
  conversation: AgentConversation;
}) {
  return (
    <div className="divide-y divide-border">
      <section>
        <h3 className="pt-3 pb-1 text-xs font-medium uppercase tracking-wide text-muted-foreground">
          基本信息
        </h3>
        {/* <InfoRow label="主题" value={conversation.subject} /> */}
        <InfoRow label="外部来源" value={conversation.externalSource} />
        <InfoRow label="外部用户标识" value={conversation.externalId} />
      </section>
      <section>
        <h3 className="pt-3 pb-1 text-xs font-medium uppercase tracking-wide text-muted-foreground">
          会话状态
        </h3>
        <InfoRow
          label="状态"
          value={
            IMConversationStatusLabels[conversation.status as IMConversationStatus] ??
            String(conversation.status)
          }
        />
        <InfoRow
          label="服务模式"
          value={
            IMConversationServiceModeLabels[
              conversation.serviceMode as IMConversationServiceMode
            ] ?? String(conversation.serviceMode)
          }
        />
        <InfoRow
          label="当前客服"
          value={conversation.currentAssigneeName ?? "-"}
        />
      </section>
      <section>
        <h3 className="pt-3 pb-1 text-xs font-medium uppercase tracking-wide text-muted-foreground">
          时间
        </h3>
        <InfoRow
          label="最后活跃"
          value={formatDateTime(conversation.lastActiveAt)}
        />
        <InfoRow
          label="最后消息"
          value={formatDateTime(conversation.lastMessageAt)}
        />
        {conversation.closedAt ? (
          <InfoRow
            label="关闭时间"
            value={formatDateTime(conversation.closedAt)}
          />
        ) : null}
      </section>
      {conversation.tags && conversation.tags.length > 0 ? (
        <section>
          <h3 className="pt-3 pb-1 text-xs font-medium uppercase tracking-wide text-muted-foreground">
            标签
          </h3>
          <ul className="flex flex-wrap gap-1.5 py-2">
            {conversation.tags.map((tag) => (
              <li
                key={tag.id}
                className={cn(
                  "rounded-md border px-2 py-0.5 text-xs",
                  !tag.color && "border-border text-foreground",
                )}
                style={
                  tag.color
                    ? {
                        borderColor: tag.color,
                        color: tag.color,
                      }
                    : undefined
                }
              >
                {tag.name}
              </li>
            ))}
          </ul>
        </section>
      ) : null}
      {conversation.participants && conversation.participants.length > 0 ? (
        <section>
          <h3 className="pt-3 pb-1 text-xs font-medium uppercase tracking-wide text-muted-foreground">
            参与者
          </h3>
          <ul className="space-y-2 py-2">
            {conversation.participants.map((p) => (
              <li
                key={p.id}
                className="rounded-md border border-border bg-muted/30 px-2 py-1.5 text-xs text-foreground"
              >
                <div className="font-medium text-foreground">
                  {p.participantType}
                </div>
                <div className="text-muted-foreground">
                  ID {p.participantId}
                  {p.externalParticipantId
                    ? ` · 外部 ${p.externalParticipantId}`
                    : ""}
                </div>
                <div className="text-muted-foreground">
                  加入 {formatDateTime(p.joinedAt)}
                  {p.leftAt ? ` · 离开 ${formatDateTime(p.leftAt)}` : ""}
                </div>
              </li>
            ))}
          </ul>
        </section>
      ) : null}
    </div>
  );
}

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
  const [activeTab, setActiveTab] = useState<InfoTabValue>("conversation");

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
      <div className="flex h-12.5 shrink-0 items-start border-b border-border p-2">
        <Tabs
          value={activeTab}
          onValueChange={(v) => setActiveTab(v as InfoTabValue)}
          className="min-w-0 flex-1 gap-0"
        >
          <TabsList className="h-full min-h-8 w-full min-w-0 justify-start">
            {infoTabOptions.map((opt) => (
              <TabsTrigger
                key={opt.value}
                value={opt.value}
                className="shrink-0 px-2.5 text-xs sm:text-sm"
              >
                {opt.label}
              </TabsTrigger>
            ))}
          </TabsList>
        </Tabs>
      </div>

      <div
        className={cn(
          "min-h-0 flex-1 overflow-y-auto px-3 pb-4",
          embedded && "pb-[max(1rem,env(safe-area-inset-bottom))] pt-1",
        )}
      >
        {activeTab === "conversation" ? (
          !conversation ? (
            <p className="pt-4 text-sm text-muted-foreground">
              {embedded
                ? "请选择会话以查看会话信息"
                : "请选择左侧会话以查看会话信息"}
            </p>
          ) : (
            <ConversationDetails conversation={conversation} />
          )
        ) : !conversation ? (
          <p className="pt-4 text-sm text-muted-foreground">
            {embedded
              ? "请选择会话以查看客户信息"
              : "请选择左侧会话以查看客户信息"}
          </p>
        ) : (
          <CustomerTabPanel conversation={conversation} />
        )}
      </div>
    </div>
  );
}
