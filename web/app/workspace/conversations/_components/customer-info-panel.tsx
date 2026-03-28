"use client";

import type { AgentConversation } from "@/lib/api/agent";
import { cn, formatDateTime } from "@/lib/utils";

const conversationStatusLabels: Record<number, string> = {
  1: "待接入",
  2: "处理中",
  3: "已关闭",
};

const serviceModeLabels: Record<number, string> = {
  1: "仅AI",
  2: "仅人工",
  3: "AI优先",
};

function InfoRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex flex-col gap-0.5 py-2">
      <span className="text-xs text-muted-foreground">{label}</span>
      <span className="break-all text-sm">{value || "-"}</span>
    </div>
  );
}

type CustomerInfoPanelProps = {
  conversation: AgentConversation | null;
  className?: string;
};

export function CustomerInfoPanel({ conversation, className }: CustomerInfoPanelProps) {
  return (
    <div
      className={cn(
        "flex h-full min-h-0 flex-col overflow-hidden border-l bg-white",
        className,
      )}
    >
      <div className="shrink-0 border-b px-3 py-2.5">
        <h2 className="text-sm font-semibold">客户信息</h2>
        <p className="text-xs text-muted-foreground">当前会话关联的客户与会话属性</p>
      </div>
      <div className="min-h-0 flex-1 overflow-y-auto px-3 pb-4">
        {!conversation ? (
          <p className="pt-4 text-sm text-muted-foreground">请选择左侧会话以查看客户信息</p>
        ) : (
          <div className="divide-y">
            <section>
              <h3 className="pt-3 pb-1 text-xs font-medium uppercase tracking-wide text-muted-foreground">
                基本信息
              </h3>
              <InfoRow label="主题" value={conversation.subject} />
              <InfoRow label="外部来源" value={conversation.externalSource} />
              <InfoRow label="外部用户标识" value={conversation.externalId} />
              <InfoRow label="来源用户 ID" value={String(conversation.sourceUserId)} />
            </section>
            <section>
              <h3 className="pt-3 pb-1 text-xs font-medium uppercase tracking-wide text-muted-foreground">
                会话状态
              </h3>
              <InfoRow
                label="状态"
                value={conversationStatusLabels[conversation.status] ?? String(conversation.status)}
              />
              <InfoRow
                label="服务模式"
                value={
                  serviceModeLabels[conversation.serviceMode] ?? String(conversation.serviceMode)
                }
              />
              <InfoRow label="优先级" value={String(conversation.priority)} />
              <InfoRow label="当前客服" value={conversation.currentAssigneeName ?? "-"} />
            </section>
            <section>
              <h3 className="pt-3 pb-1 text-xs font-medium uppercase tracking-wide text-muted-foreground">
                时间
              </h3>
              <InfoRow label="最后活跃" value={formatDateTime(conversation.lastActiveAt)} />
              <InfoRow label="最后消息" value={formatDateTime(conversation.lastMessageAt)} />
              <InfoRow label="关闭时间" value={formatDateTime(conversation.closedAt)} />
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
                      className="rounded-md border px-2 py-0.5 text-xs"
                      style={{
                        borderColor: tag.color || undefined,
                        color: tag.color || undefined,
                      }}
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
                      className="rounded-md border bg-muted/30 px-2 py-1.5 text-xs"
                    >
                      <div className="font-medium">{p.participantType}</div>
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
        )}
      </div>
    </div>
  );
}
