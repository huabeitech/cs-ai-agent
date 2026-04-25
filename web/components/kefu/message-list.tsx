"use client"

import {
  forwardRef,
  memo,
  useCallback,
  useEffect,
  useImperativeHandle,
  useLayoutEffect,
  useRef,
} from "react"

import { ImMessageHTML } from "@/components/im-message-html"
import { useImageLightbox } from "@/components/image-lightbox"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import type { ImMessage } from "@/lib/api/im"
import { renderIMMessageHTML } from "@/lib/im-message"
import { cn, formatDateTime } from "@/lib/utils"

type KefuMessageListProps = {
  messages?: ImMessage[] | null
  onNearBottomVisible?: () => void
  hasMoreOlder?: boolean
  loadingOlder?: boolean
  onLoadOlder?: () => Promise<void>
}

export type KefuMessageListHandle = {
  scrollToBottom: () => void
}

function getDayKey(value?: string) {
  if (!value) {
    return "unknown"
  }
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value.slice(0, 10)
  }
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}-${String(
    date.getDate()
  ).padStart(2, "0")}`
}

function getTimelineLabel(value?: string) {
  if (!value) {
    return "刚刚"
  }
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }
  const currentDayKey = getDayKey(value)
  const todayDayKey = getDayKey(new Date().toISOString())
  const timeText = `${String(date.getHours()).padStart(2, "0")}:${String(
    date.getMinutes()
  ).padStart(2, "0")}`
  if (currentDayKey === todayDayKey) {
    return `今天 ${timeText}`
  }
  return `${currentDayKey} ${timeText}`
}

export const KefuMessageList = forwardRef<KefuMessageListHandle, KefuMessageListProps>(
  function KefuMessageList(
    {
      messages,
      onNearBottomVisible,
      hasMoreOlder = false,
      loadingOlder = false,
      onLoadOlder,
    },
    ref
  ) {
    const containerRef = useRef<HTMLDivElement>(null)
    const contentRef = useRef<HTMLDivElement>(null)
    const frameRef = useRef<number | null>(null)
    const shouldStickToBottomRef = useRef(true)
    const onNearBottomVisibleRef = useRef(onNearBottomVisible)
    const safeMessages = Array.isArray(messages) ? messages : []
    const lastMessageId = safeMessages.at(-1)?.id

    useEffect(() => {
      onNearBottomVisibleRef.current = onNearBottomVisible
    }, [onNearBottomVisible])

    const isNearBottom = useCallback(
      (element: HTMLElement, threshold = 80) =>
        element.scrollHeight - element.scrollTop - element.clientHeight <= threshold,
      []
    )

    const scrollToBottom = useCallback(() => {
      const container = containerRef.current
      if (!container) {
        return
      }
      container.scrollTop = container.scrollHeight
    }, [])

    const scheduleScrollToBottom = useCallback(
      (attempts = 4) => {
        if (frameRef.current !== null) {
          cancelAnimationFrame(frameRef.current)
        }

        const run = (remaining: number, previousHeight = -1) => {
          frameRef.current = requestAnimationFrame(() => {
            const container = containerRef.current
            if (!container) {
              frameRef.current = null
              return
            }

            const currentHeight = container.scrollHeight
            scrollToBottom()
            if (remaining > 1 && currentHeight !== previousHeight) {
              run(remaining - 1, currentHeight)
              return
            }
            frameRef.current = null
          })
        }

        run(attempts)
      },
      [scrollToBottom]
    )

    const handleImageSettled = useCallback(() => {
      if (shouldStickToBottomRef.current) {
        scheduleScrollToBottom()
        onNearBottomVisibleRef.current?.()
      }
    }, [scheduleScrollToBottom])

    useImperativeHandle(ref, () => ({
      scrollToBottom,
    }))

    useLayoutEffect(() => {
      shouldStickToBottomRef.current = true
      scheduleScrollToBottom()
      return () => {
        if (frameRef.current !== null) {
          cancelAnimationFrame(frameRef.current)
          frameRef.current = null
        }
      }
    }, [lastMessageId, scheduleScrollToBottom])

    useEffect(() => {
      const container = containerRef.current
      const content = contentRef.current
      if (!container || !content) {
        return
      }

      const handleScroll = () => {
        shouldStickToBottomRef.current = isNearBottom(container)
        if (shouldStickToBottomRef.current) {
          onNearBottomVisible?.()
        }
      }

      const resizeObserver = new ResizeObserver(() => {
        if (shouldStickToBottomRef.current) {
          scheduleScrollToBottom()
        }
      })

      handleScroll()
      container.addEventListener("scroll", handleScroll)
      resizeObserver.observe(container)
      resizeObserver.observe(content)
      scrollToBottom()

      return () => {
        container.removeEventListener("scroll", handleScroll)
        resizeObserver.disconnect()
      }
    }, [isNearBottom, onNearBottomVisible, scheduleScrollToBottom, scrollToBottom])

    const handleLoadOlder = useCallback(async () => {
      if (!onLoadOlder || loadingOlder || !hasMoreOlder) {
        return
      }
      const container = containerRef.current
      if (!container) {
        return
      }
      const anchor = {
        height: container.scrollHeight,
        top: container.scrollTop,
      }
      try {
        await onLoadOlder()
      } catch {
        return
      }
      requestAnimationFrame(() => {
        requestAnimationFrame(() => {
          const current = containerRef.current
          if (!current) {
            return
          }
          current.scrollTop = current.scrollHeight - anchor.height + anchor.top
        })
      })
    }, [hasMoreOlder, loadingOlder, onLoadOlder])

    return (
      <div
        ref={containerRef}
        className="cs-agent-scrollbar flex min-h-0 flex-1 flex-col gap-4 overflow-y-auto px-4 py-4"
      >
        <div ref={contentRef} className="flex flex-col gap-4">
          {hasMoreOlder && onLoadOlder ? (
            <div className="flex justify-center py-1">
              <Button
                type="button"
                variant="outline"
                size="sm"
                disabled={loadingOlder}
                onClick={() => void handleLoadOlder()}
                className="h-7 rounded-full bg-background/90 text-xs text-muted-foreground shadow-sm hover:bg-background hover:text-sky-700 dark:hover:text-sky-400"
              >
                {loadingOlder ? "加载中…" : "加载更早的消息"}
              </Button>
            </div>
          ) : null}

          {safeMessages.map((message, index) => {
            const previousMessage = index > 0 ? safeMessages[index - 1] : null
            const showTimeline =
              index === 0 ||
              getDayKey(previousMessage?.sentAt) !== getDayKey(message.sentAt)

            return (
              <MessageItem
                key={message.id}
                message={message}
                showTimeline={showTimeline}
                onImageSettled={handleImageSettled}
              />
            )
          })}
        </div>
      </div>
    )
  }
)

type MessageItemProps = {
  message: ImMessage
  showTimeline: boolean
  onImageSettled: () => void
}

const MessageItem = memo(
  function MessageItem({ message, showTimeline, onImageSettled }: MessageItemProps) {
    const { open } = useImageLightbox()
    const isCustomer = message.senderType === "customer"
    const senderName = isCustomer ? "我" : message.senderName?.trim() || "客服"
    const avatarSrc =
      !isCustomer && message.senderAvatar?.trim() ? message.senderAvatar.trim() : undefined
    const htmlContent = renderIMMessageHTML(message)
    const fallbackName = senderName.slice(0, 1).toUpperCase()

    return (
      <div>
        {showTimeline ? (
          <div className="mb-3 flex items-center justify-center">
            <Badge
              variant="outline"
              className="border-border bg-background/85 text-[11px] font-medium text-muted-foreground shadow-sm"
            >
              {getTimelineLabel(message.sentAt)}
            </Badge>
          </div>
        ) : null}

        <div className={cn("flex gap-2.5", isCustomer ? "justify-end" : "justify-start")}>
          {!isCustomer ? (
            <Avatar className="mt-5">
              {avatarSrc ? <AvatarImage src={avatarSrc} alt="" /> : null}
              <AvatarFallback className="bg-muted text-muted-foreground">
                {fallbackName || "客"}
              </AvatarFallback>
            </Avatar>
          ) : null}

          <div
            className={cn(
              "flex max-w-[86%] flex-col gap-1.5",
              isCustomer ? "items-end" : "items-start"
            )}
          >
            <div className="flex flex-wrap items-center gap-x-2 gap-y-1 px-1 text-[11px] text-muted-foreground">
              <span className="font-medium">{senderName}</span>
              <span>{formatDateTime(message.sentAt)}</span>
              {isCustomer ? (
                <span>{message.agentRead ? "客服已读" : "客服未读"}</span>
              ) : null}
            </div>
            <div
              className={cn(
                "rounded-lg px-3 py-2 text-sm leading-normal shadow-[0_10px_22px_rgba(15,23,42,0.06)]",
                isCustomer
                  ? "bg-[#a9ea7a] text-[#161616] dark:bg-emerald-500 dark:text-emerald-950"
                  : "border border-border bg-card text-card-foreground dark:bg-background"
              )}
            >
              <ImMessageHTML
                html={htmlContent}
                className={cn(
                  isCustomer
                    ? "[&_p]:text-[#161616] dark:[&_p]:text-emerald-950 [&_a]:text-[#161616] dark:[&_a]:text-emerald-950 [&_a]:underline [&_img]:cursor-zoom-in"
                    : "[&_a]:text-card-foreground [&_a]:underline [&_img]:cursor-zoom-in"
                )}
                onImageSettled={onImageSettled}
                onImageClick={open}
              />
            </div>
          </div>
        </div>
      </div>
    )
  },
  (prevProps, nextProps) =>
    isSameMessageItemRender(prevProps.message, nextProps.message) &&
    prevProps.showTimeline === nextProps.showTimeline &&
    prevProps.onImageSettled === nextProps.onImageSettled
)

function isSameMessageItemRender(prev: ImMessage, next: ImMessage) {
  return (
    prev.id === next.id &&
    prev.senderType === next.senderType &&
    prev.senderName === next.senderName &&
    prev.senderAvatar === next.senderAvatar &&
    prev.messageType === next.messageType &&
    prev.content === next.content &&
    prev.payload === next.payload &&
    prev.sentAt === next.sentAt &&
    prev.agentRead === next.agentRead
  )
}
