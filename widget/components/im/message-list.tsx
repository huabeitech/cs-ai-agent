"use client";

import {
  forwardRef,
  memo,
  useCallback,
  useEffect,
  useImperativeHandle,
  useLayoutEffect,
  useRef,
} from "react";
import Image from "next/image";

import { MessageHTML } from "@/components/im/message-html";
import { useImageLightbox } from "@/components/image-lightbox";
import { renderMessageHTML } from "@/lib/services/message-asset";
import type { WidgetMessage } from "@/lib/services/types";
import { cn, formatDateTime } from "@/lib/utils";

type MessageListProps = {
  messages: WidgetMessage[];
  onNearBottomVisible?: () => void;
  hasMoreOlder?: boolean;
  loadingOlder?: boolean;
  onLoadOlder?: () => Promise<void>;
};

export type MessageListHandle = {
  scrollToBottom: () => void;
};

function getDayKey(value?: string) {
  if (!value) {
    return "unknown";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value.slice(0, 10);
  }
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}-${String(
    date.getDate(),
  ).padStart(2, "0")}`;
}

function getTimelineLabel(value?: string) {
  if (!value) {
    return "刚刚";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  const now = new Date();
  const currentDayKey = getDayKey(value);
  const todayDayKey = getDayKey(now.toISOString());
  const timeText = `${String(date.getHours()).padStart(2, "0")}:${String(date.getMinutes()).padStart(2, "0")}`;
  if (currentDayKey === todayDayKey) {
    return `今天 ${timeText}`;
  }
  return `${currentDayKey} ${timeText}`;
}

export const MessageList = forwardRef<MessageListHandle, MessageListProps>(
  function MessageList(
    {
      messages,
      onNearBottomVisible,
      hasMoreOlder = false,
      loadingOlder = false,
      onLoadOlder,
    },
    ref,
  ) {
    const containerRef = useRef<HTMLDivElement>(null);
    const contentRef = useRef<HTMLDivElement>(null);
    const frameRef = useRef<number | null>(null);
    const shouldStickToBottomRef = useRef(true);
    const lastMessageId = messages.at(-1)?.id;

    const isNearBottom = useCallback(
      (element: HTMLElement, threshold = 80) =>
        element.scrollHeight - element.scrollTop - element.clientHeight <=
        threshold,
      [],
    );

    const scrollToBottom = useCallback(() => {
      const container = containerRef.current;
      if (!container) {
        return;
      }
      container.scrollTop = container.scrollHeight;
    }, []);

    const scheduleScrollToBottom = useCallback(
      (attempts = 4) => {
        if (frameRef.current !== null) {
          cancelAnimationFrame(frameRef.current);
        }

        const run = (remaining: number, previousHeight = -1) => {
          frameRef.current = requestAnimationFrame(() => {
            const container = containerRef.current;
            if (!container) {
              frameRef.current = null;
              return;
            }

            const currentHeight = container.scrollHeight;
            scrollToBottom();
            if (remaining > 1 && currentHeight !== previousHeight) {
              run(remaining - 1, currentHeight);
              return;
            }
            frameRef.current = null;
          });
        };

        run(attempts);
      },
      [scrollToBottom],
    );

    const handleImageSettled = useCallback(() => {
      if (shouldStickToBottomRef.current) {
        scheduleScrollToBottom();
        onNearBottomVisible?.();
      }
    }, [onNearBottomVisible, scheduleScrollToBottom]);

    useImperativeHandle(ref, () => ({
      scrollToBottom,
    }));

    useLayoutEffect(() => {
      shouldStickToBottomRef.current = true;
      scheduleScrollToBottom();
      return () => {
        if (frameRef.current !== null) {
          cancelAnimationFrame(frameRef.current);
          frameRef.current = null;
        }
      };
    }, [lastMessageId, scheduleScrollToBottom]);

    useEffect(() => {
      const container = containerRef.current;
      const content = contentRef.current;
      if (!container || !content) {
        return;
      }

      const handleScroll = () => {
        shouldStickToBottomRef.current = isNearBottom(container);
        if (shouldStickToBottomRef.current) {
          onNearBottomVisible?.();
        }
      };

      const resizeObserver = new ResizeObserver(() => {
        if (shouldStickToBottomRef.current) {
          scheduleScrollToBottom();
        }
      });

      handleScroll();
      container.addEventListener("scroll", handleScroll);
      resizeObserver.observe(content);
      scrollToBottom();

      return () => {
        container.removeEventListener("scroll", handleScroll);
        resizeObserver.disconnect();
      };
    }, [
      isNearBottom,
      onNearBottomVisible,
      scheduleScrollToBottom,
      scrollToBottom,
    ]);

    const handleLoadOlder = useCallback(async () => {
      if (!onLoadOlder || loadingOlder || !hasMoreOlder) {
        return;
      }
      const container = containerRef.current;
      if (!container) {
        return;
      }
      const anchor = {
        height: container.scrollHeight,
        top: container.scrollTop,
      };
      try {
        await onLoadOlder();
      } catch {
        return;
      }
      requestAnimationFrame(() => {
        requestAnimationFrame(() => {
          const c = containerRef.current;
          if (!c) {
            return;
          }
          c.scrollTop = c.scrollHeight - anchor.height + anchor.top;
        });
      });
    }, [hasMoreOlder, loadingOlder, onLoadOlder]);

    return (
      <div
        ref={containerRef}
        className="flex min-h-0 flex-1 flex-col gap-4 overflow-y-auto px-4 py-4 cs-agent-scrollbar"
      >
        <div ref={contentRef} className="flex flex-col gap-4">
          {hasMoreOlder && onLoadOlder ? (
            <div className="flex justify-center py-1">
              <button
                type="button"
                disabled={loadingOlder}
                onClick={() => void handleLoadOlder()}
                className="rounded-full border border-white/70 bg-white/75 px-3 py-1 text-[11px] font-medium text-slate-500 shadow-[0_8px_18px_rgba(15,23,42,0.04)] backdrop-blur transition hover:-translate-y-0.5 hover:border-sky-200 hover:text-sky-700 disabled:translate-y-0 disabled:opacity-60"
              >
                {loadingOlder ? "加载中…" : "加载更早的消息"}
              </button>
            </div>
          ) : null}
          {/* <WelcomePanel title={title} welcomeText={welcomeText} /> */}

          {/* {messages.length === 0 ? (
        <div className="cs-agent-fade-up rounded-3xl border border-dashed border-slate-200 bg-white/72 px-4 py-5 text-sm leading-6 text-slate-500 shadow-[0_10px_22px_rgba(15,23,42,0.04)] backdrop-blur">
          开始发送第一条消息后，会在这里保留完整会话记录。
        </div>
      ) : null} */}

          {messages.map((message, index) => {
            const previousMessage = index > 0 ? messages[index - 1] : null;
            const showTimeline =
              index === 0 ||
              getDayKey(previousMessage?.sentAt) !== getDayKey(message.sentAt);

            return (
              <MessageItem
                key={message.id}
                message={message}
                showTimeline={showTimeline}
                onImageSettled={handleImageSettled}
              />
            );
          })}
        </div>
      </div>
    );
  },
);

type MessageItemProps = {
  message: WidgetMessage;
  showTimeline: boolean;
  onImageSettled: () => void;
};

const MessageItem = memo(
  function MessageItem({
    message,
    showTimeline,
    onImageSettled,
  }: MessageItemProps) {
    const { open: openImageLightbox } = useImageLightbox();
    const isCustomer = message.senderType === "customer";
    const senderName = isCustomer ? "我" : message.senderName?.trim() || "客服";
    const agentAvatarSrc =
      !isCustomer && message.senderAvatar?.trim()
        ? message.senderAvatar.trim()
        : undefined;
    const htmlContent = buildMessageHTML(message);

    return (
      <div className="cs-agent-fade-up">
        {showTimeline ? (
          <div className="mb-3 flex items-center justify-center">
            <div className="rounded-full border border-white/70 bg-white/80 px-3 py-1 text-[11px] font-medium text-slate-500 shadow-[0_8px_18px_rgba(15,23,42,0.04)] backdrop-blur">
              {getTimelineLabel(message.sentAt)}
            </div>
          </div>
        ) : null}

        <div
          className={cn(
            "flex gap-2",
            isCustomer ? "justify-end" : "justify-start",
          )}
        >
          {!isCustomer && agentAvatarSrc ? (
            <Image
              src={agentAvatarSrc}
              alt=""
              width={32}
              height={32}
              className="size-8 shrink-0 rounded-full object-cover ring-1 ring-white/80"
            />
          ) : null}
          <div
            className={cn(
              "flex max-w-[86%] flex-col gap-1",
              isCustomer ? "items-end" : "items-start",
            )}
          >
            <div className="flex items-center gap-2 px-1 text-[11px] text-slate-400">
              <span className="font-medium">{senderName}</span>
              <span>{formatDateTime(message.sentAt)}</span>
              {isCustomer ? (
                <span>{message.agentRead ? "客服已读" : "客服未读"}</span>
              ) : null}
            </div>
            <div
              className={cn(
                "rounded-lg px-3 py-2 text-sm leading-normal shadow-[0_14px_28px_rgba(15,23,42,0.08)]",
                isCustomer
                  ? "bg-[linear-gradient(135deg,var(--primary),color-mix(in_srgb,var(--primary)_78%,white_22%))] text-white"
                  : "border border-white/80 bg-white/94 text-slate-900",
              )}
            >
              <MessageHTML
                html={htmlContent}
                className={cn(
                  isCustomer
                    ? "[&_p]:text-white [&_a]:text-white [&_a]:underline [&_img]:cursor-zoom-in"
                    : "[&_a]:text-slate-900 [&_a]:underline [&_img]:cursor-zoom-in",
                )}
                onImageSettled={onImageSettled}
                onImageClick={openImageLightbox}
              />
            </div>
          </div>
        </div>
      </div>
    );
  },
  (prevProps, nextProps) =>
    prevProps.message === nextProps.message &&
    prevProps.showTimeline === nextProps.showTimeline &&
    prevProps.onImageSettled === nextProps.onImageSettled,
);

function buildMessageHTML(message: WidgetMessage) {
  return renderMessageHTML(message);
}
