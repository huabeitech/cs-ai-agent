"use client";

import {
  CheckCheckIcon,
  EyeIcon,
  MessageCircleMoreIcon,
} from "lucide-react";
import Image from "next/image";
import { useEffect, useRef, useState } from "react";

import { ImMessageHTML } from "@/components/im-message-html";
import { ProjectDialog } from "@/components/project-dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent } from "@/components/ui/dialog";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Separator } from "@/components/ui/separator";
import {
  type AdminConversation,
  type AdminConversationDetail,
  type AdminMessage,
} from "@/lib/api/admin";
import { formatDateTime } from "@/lib/utils";

type ConversationDetailDialogProps = {
  open: boolean;
  loading: boolean;
  saving: boolean;
  item: AdminConversation | null;
  detail: AdminConversationDetail | null;
  messages: AdminMessage[];
  onOpenChange: (open: boolean) => void;
  onOpenAssign: () => void;
  onDispatch: () => Promise<void>;
  onOpenTransfer: () => void;
  onRead: () => Promise<void>;
  onOpenClose: () => void;
};

function getStatusMeta(status: number) {
  switch (status) {
    case 1:
      return { label: "待接入", variant: "outline" as const };
    case 2:
      return { label: "处理中", variant: "secondary" as const };
    case 3:
      return { label: "已关闭", variant: "outline" as const };
    case 4:
      return { label: "已归档", variant: "outline" as const };
    default:
      return { label: "未知", variant: "outline" as const };
  }
}

function getServiceModeLabel(mode: number) {
  switch (mode) {
    case 1:
      return "AI 接待";
    case 2:
      return "人工接待";
    case 3:
      return "AI 优先";
    default:
      return "未定义";
  }
}

function getSenderLabel(message: AdminMessage) {
  switch (message.senderType) {
    case "agent":
      return message.senderName || "客服";
    case "customer":
      return message.senderName || "用户";
    case "ai":
      return "AI";
    case "system":
      return "系统";
    default:
      return message.senderType;
  }
}

function getMessageContent(message: AdminMessage) {
  return message.content || message.payload || "-";
}

function getImageMessageUrl(message: AdminMessage) {
  if (message.messageType === "image") {
    try {
      const payload = JSON.parse(message.payload || "{}") as {
        url?: string;
      };
      if (payload.url) {
        return payload.url;
      }
    } catch {
      return message.content || "";
    }
  }
  return "";
}

function getParticipantIdentity(
  participant: NonNullable<AdminConversationDetail["participants"]>[number],
) {
  return participant.participantId || participant.externalParticipantId || "-";
}

function getMessageLayout(message: AdminMessage) {
  if (message.senderType === "customer") {
    return {
      rowClassName: "justify-start",
      bubbleClassName: "bg-muted text-foreground border-border",
      metaClassName: "text-left",
    };
  }
  if (message.senderType === "system") {
    return {
      rowClassName: "justify-center",
      bubbleClassName:
        "bg-muted/60 text-muted-foreground border-dashed border-border",
      metaClassName: "text-center",
    };
  }
  if (message.senderType === "ai") {
    return {
      rowClassName: "justify-end",
      bubbleClassName: "bg-primary/10 text-foreground border-primary/20",
      metaClassName: "text-right",
    };
  }
  return {
    rowClassName: "justify-end",
    bubbleClassName: "bg-primary text-primary-foreground border-primary",
    metaClassName: "text-right",
  };
}

export function ConversationDetailDialog({
  open,
  loading,
  saving,
  item,
  detail,
  messages,
  onOpenChange,
  onOpenAssign,
  onDispatch,
  onOpenTransfer,
  onRead,
  onOpenClose,
}: ConversationDetailDialogProps) {
  const currentConversation = detail ?? item;
  const isClosedConversation = currentConversation?.status === 3;
  const isPendingConversation = currentConversation?.status === 1;
  const statusMeta = currentConversation
    ? getStatusMeta(currentConversation.status)
    : null;
  const messageBottomRef = useRef<HTMLDivElement | null>(null);
  const [previewImage, setPreviewImage] = useState("");

  useEffect(() => {
    if (!open) {
      setPreviewImage("");
      return;
    }
    const bottom = messageBottomRef.current;
    if (!bottom) {
      return;
    }
    bottom.scrollIntoView({ block: "end", behavior: "smooth" });
  }, [messages, open]);

  return (
    <ProjectDialog
      open={open}
      onOpenChange={onOpenChange}
      title={currentConversation?.subject || "会话详情"}
      size="xl"
      // allowFullscreen
      defaultFullscreen
      bodyScrollable={false}
      bodyClassName="h-full overflow-hidden p-0"
      contentClassName="h-[calc(100vh-40px)] max-h-[calc(100vh-40px)]"
      footer={
        <div className="flex w-full flex-wrap items-center justify-between gap-3">
          <div className="text-sm text-muted-foreground">
            {currentConversation
              ? `最后活跃：${formatDateTime(currentConversation.lastMessageAt)}`
              : "暂无会话信息"}
          </div>
          <div className="flex flex-wrap gap-2">
            <Button
              variant="outline"
              onClick={onOpenAssign}
              disabled={saving || !currentConversation || isClosedConversation}
            >
              <MessageCircleMoreIcon />
              {saving ? "处理中..." : "分配会话"}
            </Button>
            <Button
              variant="outline"
              onClick={() => void onDispatch()}
              disabled={saving || !currentConversation || !isPendingConversation}
            >
              <MessageCircleMoreIcon />
              {saving ? "处理中..." : "重试分配"}
            </Button>
            <Button
              variant="outline"
              onClick={() => void onRead()}
              disabled={saving || !currentConversation}
            >
              <CheckCheckIcon />
              {saving ? "处理中..." : "标记已读"}
            </Button>
            <Button
              type="button"
              variant="outline"
              onClick={onOpenTransfer}
              disabled={saving || !currentConversation || isClosedConversation}
            >
              <MessageCircleMoreIcon />
              {saving ? "处理中..." : "转接会话"}
            </Button>
            {!isClosedConversation ? (
              <Button
                variant="outline"
                onClick={onOpenClose}
                disabled={saving || !currentConversation}
              >
                <EyeIcon />
                {saving ? "处理中..." : "关闭会话"}
              </Button>
            ) : null}
          </div>
        </div>
      }
    >
      {loading ? (
        <div className="flex h-full min-h-[60vh] items-center justify-center text-sm text-muted-foreground">
          正在加载会话详情...
        </div>
      ) : currentConversation ? (
        <div className="flex h-[calc(100vh)] min-h-[60vh] overflow-hidden flex-row border-t">
          <aside className="flex w-90 h-full shrink-0 flex-col overflow-hidden bg-muted/20 border-r border-b-0">
            <div className="space-y-4 p-6">
              <div className="flex items-start justify-between gap-3">
                <div className="space-y-2">
                  <p className="text-sm text-muted-foreground">
                    来源用户：
                    {currentConversation.externalUserId ||
                      currentConversation.sourceUserId ||
                      "-"}
                  </p>
                </div>
                {statusMeta ? (
                  <Badge variant={statusMeta.variant}>{statusMeta.label}</Badge>
                ) : null}
              </div>

              <div className="grid grid-cols-2 gap-3 text-sm">
                <InfoItem
                  label="接待模式"
                  value={getServiceModeLabel(currentConversation.serviceMode)}
                />
                <InfoItem
                  label="当前客服"
                  value={currentConversation.currentAssigneeName || "-"}
                />
                <InfoItem
                  label="渠道类型"
                  value={currentConversation.channelType || "-"}
                />
                <InfoItem
                  label="客服未读"
                  value={`${currentConversation.agentUnreadCount}`}
                />
                <InfoItem
                  label="用户未读"
                  value={`${currentConversation.customerUnreadCount}`}
                />
                <InfoItem
                  label="最后活跃"
                  value={formatDateTime(currentConversation.lastMessageAt)}
                  fullWidth
                />
                <InfoItem
                  label="关闭时间"
                  value={formatDateTime(currentConversation.closedAt)}
                  fullWidth
                />
              </div>
            </div>

            <Separator />

            <ScrollArea className="min-h-0 flex-1">
              <div className="space-y-4 p-6">
                <section className="space-y-3">
                  <div className="flex items-center justify-between">
                    <p className="text-sm font-medium">参与方</p>
                    <span className="text-xs text-muted-foreground">
                      {detail?.participants?.length ?? 0} 人
                    </span>
                  </div>
                  {detail?.participants?.length ? (
                    <div className="space-y-3">
                      {detail.participants.map((participant) => (
                        <div
                          key={participant.id}
                          className="rounded-lg border bg-background p-3"
                        >
                          <div className="flex items-center justify-between gap-3">
                            <span className="text-sm font-medium">
                              {participant.participantType || "-"}
                            </span>
                          </div>
                          <div className="mt-1 text-sm text-muted-foreground">
                            标识：{getParticipantIdentity(participant)}
                          </div>
                          <div className="mt-1 text-xs text-muted-foreground">
                            加入时间：{formatDateTime(participant.joinedAt)}
                          </div>
                        </div>
                      ))}
                    </div>
                  ) : (
                    <div className="rounded-lg border border-dashed bg-background p-4 text-sm text-muted-foreground">
                      暂无参与方信息
                    </div>
                  )}
                </section>
              </div>
            </ScrollArea>
          </aside>

          <section className="flex h-full min-h-0 min-w-0 flex-1 flex-col overflow-hidden bg-background">
            <div className="flex items-center justify-between border-b px-6 py-4">
              <div>
                <p className="text-sm font-medium">聊天记录</p>
                <p className="mt-1 text-xs text-muted-foreground">
                  展示最近 20 条消息
                </p>
              </div>
              <div className="text-xs text-muted-foreground">
                消息数：{messages.length}
              </div>
            </div>

            <ScrollArea className="min-h-0 flex-1 bg-muted/10">
              <div className="space-y-4 px-6 py-5">
                {isClosedConversation ? (
                  <div className="rounded-xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-900">
                    当前会话已关闭。后台不可继续发送消息；如用户再次咨询，应创建新会话。
                  </div>
                ) : null}
                {messages.length ? (
                  messages.map((message) => {
                    const layout = getMessageLayout(message);
                    const isHtmlMessage = message.messageType === "html";
                    const isImageMessage = message.messageType === "image";

                    return (
                      <div
                        key={message.id}
                        className={`flex ${layout.rowClassName}`}
                      >
                        <div className="max-w-[85%] space-y-2">
                          <div
                            className={`text-xs text-muted-foreground ${layout.metaClassName}`}
                          >
                            <span>{getSenderLabel(message)}</span>
                            <span className="mx-2">·</span>
                            <span>{formatDateTime(message.sentAt)}</span>
                          </div>
                          <div
                            className={`rounded-2xl border px-4 py-3 text-sm leading-6 ${layout.bubbleClassName}`}
                          >
                            {isHtmlMessage ? (
                              <ImMessageHTML
                                html={message.content || "-"}
                                className="[&_a]:underline [&_img]:max-w-full [&_img]:cursor-zoom-in"
                                onImageClick={setPreviewImage}
                              />
                            ) : isImageMessage ? (
                              <MessageImage
                                src={getImageMessageUrl(message)}
                                alt={getMessageContent(message)}
                                onPreview={setPreviewImage}
                              />
                            ) : (
                              <div className="whitespace-pre-wrap break-words">
                                {getMessageContent(message)}
                              </div>
                            )}
                          </div>
                          <div
                            className={`text-xs text-muted-foreground ${layout.metaClassName}`}
                          >
                            客服 {message.agentRead ? "已读" : "未读"} / 用户{" "}
                            {message.customerRead ? "已读" : "未读"}
                          </div>
                        </div>
                      </div>
                    );
                  })
                ) : (
                  <div className="flex h-full min-h-80 items-center justify-center rounded-xl border border-dashed bg-background text-sm text-muted-foreground">
                    暂无消息记录
                  </div>
                )}
                <div ref={messageBottomRef} />
              </div>
            </ScrollArea>
          </section>
        </div>
      ) : (
        <div className="flex h-full min-h-[60vh] items-center justify-center text-sm text-muted-foreground">
          暂无可展示的会话详情
        </div>
      )}
      <Dialog open={Boolean(previewImage)} onOpenChange={(nextOpen) => !nextOpen && setPreviewImage("")}>
        <DialogContent
          className="max-w-[calc(100vw-2rem)] border-0 bg-transparent p-0 shadow-none ring-0 sm:max-w-[calc(100vw-4rem)]"
          showCloseButton
        >
          {previewImage ? (
            <div className="flex max-h-[85vh] items-center justify-center">
              <Image
                src={previewImage}
                alt="消息图片预览"
                width={1600}
                height={1200}
                className="max-h-[85vh] w-auto rounded-lg object-contain"
                unoptimized
              />
            </div>
          ) : null}
        </DialogContent>
      </Dialog>
    </ProjectDialog>
  );
}

type InfoItemProps = {
  label: string;
  value: string;
  fullWidth?: boolean;
};

function InfoItem({ label, value, fullWidth = false }: InfoItemProps) {
  return (
    <div className={fullWidth ? "col-span-2" : undefined}>
      <p className="text-xs text-muted-foreground">{label}</p>
      <p className="mt-1 break-all font-medium">{value || "-"}</p>
    </div>
  );
}

type MessageImageProps = {
  src: string;
  alt: string;
  onPreview: (src: string) => void;
};

function MessageImage({ src, alt, onPreview }: MessageImageProps) {
  if (!src) {
    return <div className="text-sm whitespace-pre-wrap break-words">{alt || "[图片]"}</div>;
  }

  return (
    <button
      type="button"
      className="block cursor-zoom-in"
      onClick={() => onPreview(src)}
    >
      <Image
        src={src}
        alt={alt || "消息图片"}
        width={480}
        height={360}
        className="max-h-64 w-auto max-w-full rounded-md object-contain"
        unoptimized
      />
    </button>
  );
}
