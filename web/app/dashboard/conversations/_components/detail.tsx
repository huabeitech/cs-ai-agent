"use client";

import { CheckCheckIcon, EyeIcon, MessageCircleMoreIcon } from "lucide-react";
import Image from "next/image";
import {
  useCallback,
  useEffect,
  useLayoutEffect,
  useRef,
} from "react";

import { ImMessageHTML } from "@/components/im-message-html";
import { useImageLightbox } from "@/components/image-lightbox";
import { ProjectDialog } from "@/components/project-dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
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
  /** 是否还有更早消息（cursor 分页） */
  messagesHasMore?: boolean;
  loadingMoreMessages?: boolean;
  onLoadMoreMessages?: () => void | Promise<void>;
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
  messagesHasMore = false,
  loadingMoreMessages = false,
  onLoadMoreMessages,
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
  const messagesScrollRootRef = useRef<HTMLDivElement | null>(null);
  const loadMoreSentinelRef = useRef<HTMLDivElement | null>(null);
  const pendingScrollAnchorRef = useRef<{
    scrollHeight: number;
    scrollTop: number;
  } | null>(null);
  const prevLoadingMoreRef = useRef(false);
  const { open: openImageLightbox, close: closeImageLightbox } =
    useImageLightbox();

  const getMessagesViewport = useCallback((): HTMLElement | null => {
    return (
      messagesScrollRootRef.current?.querySelector(
        '[data-slot="scroll-area-viewport"]',
      ) ?? null
    );
  }, []);

  useEffect(() => {
    if (!open) {
      closeImageLightbox();
      return;
    }
    if (loading) {
      return;
    }
    const bottom = messageBottomRef.current;
    if (!bottom) {
      return;
    }
    bottom.scrollIntoView({ block: "end", behavior: "smooth" });
  }, [open, loading, closeImageLightbox]);

  useLayoutEffect(() => {
    const wasLoading = prevLoadingMoreRef.current;
    prevLoadingMoreRef.current = loadingMoreMessages;
    if (wasLoading && !loadingMoreMessages && pendingScrollAnchorRef.current) {
      const vp = getMessagesViewport();
      const anchor = pendingScrollAnchorRef.current;
      pendingScrollAnchorRef.current = null;
      if (vp && anchor) {
        const delta = vp.scrollHeight - anchor.scrollHeight;
        vp.scrollTop = anchor.scrollTop + delta;
      }
    }
  }, [loadingMoreMessages, messages, getMessagesViewport]);

  useEffect(() => {
    if (!open || loading || !messagesHasMore || !onLoadMoreMessages) {
      return;
    }
    const root = getMessagesViewport();
    const sentinel = loadMoreSentinelRef.current;
    if (!root || !sentinel) {
      return;
    }
    const observer = new IntersectionObserver(
      (entries) => {
        const hit = entries.some((e) => e.isIntersecting);
        if (!hit || loadingMoreMessages) {
          return;
        }
        const vp = getMessagesViewport();
        if (vp) {
          pendingScrollAnchorRef.current = {
            scrollHeight: vp.scrollHeight,
            scrollTop: vp.scrollTop,
          };
        }
        void onLoadMoreMessages();
      },
      { root, rootMargin: "120px 0px 0px 0px", threshold: 0 },
    );
    observer.observe(sentinel);
    return () => observer.disconnect();
  }, [
    open,
    loading,
    messagesHasMore,
    loadingMoreMessages,
    messages.length,
    onLoadMoreMessages,
    getMessagesViewport,
  ]);

  return (
    <ProjectDialog
      open={open}
      onOpenChange={onOpenChange}
      title={
        <div className="flex items-center gap-3">
          <span>{currentConversation?.subject || "会话详情"}</span>

          <div>
            {statusMeta ? (
              <Badge variant={statusMeta.variant}>{statusMeta.label}</Badge>
            ) : null}
          </div>
        </div>
      }
      size="xl"
      // allowFullscreen
      defaultFullscreen
      bodyScrollable={false}
      bodyClassName="flex min-h-0 flex-1 flex-col overflow-hidden p-0"
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
              disabled={
                saving || !currentConversation || !isPendingConversation
              }
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
        <div className="flex min-h-0 flex-1 items-center justify-center text-sm text-muted-foreground">
          正在加载会话详情...
        </div>
      ) : currentConversation ? (
        <div className="flex min-h-0 flex-1 flex-row overflow-hidden border-t">
          <aside className="flex w-90 h-full shrink-0 flex-col overflow-hidden bg-muted/20 border-r border-b-0">
            <div className="space-y-4 p-6">
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
                  value={currentConversation.externalSource || "-"}
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
            {/* <div className="flex items-center justify-between border-b px-6 py-4">
              <div>
                <p className="text-sm font-medium">聊天记录</p>
              </div>
              <div className="text-xs text-muted-foreground">
                消息数：{messages.length}
                {messagesHasMore ? " · 向上滑可加载更早" : ""}
              </div>
            </div> */}

            <div ref={messagesScrollRootRef} className="min-h-0 flex-1">
              <ScrollArea className="h-full min-h-0 bg-muted/10">
                <div className="space-y-4 px-6 py-5">
                  {messagesHasMore ? (
                    <div
                      ref={loadMoreSentinelRef}
                      className="flex min-h-8 flex-col items-center justify-center py-2 text-xs text-muted-foreground"
                    >
                      {loadingMoreMessages
                        ? "正在加载更早的消息…"
                        : "继续上滑加载更早消息"}
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
                                  onImageClick={openImageLightbox}
                                />
                              ) : isImageMessage ? (
                                <MessageImage
                                  src={getImageMessageUrl(message)}
                                  alt={getMessageContent(message)}
                                  onPreview={openImageLightbox}
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
            </div>
          </section>
        </div>
      ) : (
        <div className="flex min-h-0 flex-1 items-center justify-center text-sm text-muted-foreground">
          暂无可展示的会话详情
        </div>
      )}
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
  onPreview: (src: string, alt?: string) => void;
};

function MessageImage({ src, alt, onPreview }: MessageImageProps) {
  if (!src) {
    return (
      <div className="text-sm whitespace-pre-wrap break-words">
        {alt || "[图片]"}
      </div>
    );
  }

  return (
    <button
      type="button"
      className="block cursor-zoom-in"
      onClick={() => onPreview(src, alt)}
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
