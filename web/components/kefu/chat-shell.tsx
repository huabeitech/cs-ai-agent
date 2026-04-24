"use client"

import {
  Maximize2Icon,
  Minimize2Icon,
  MinusIcon,
  RotateCwIcon,
  XIcon,
} from "lucide-react"
import { useCallback, useEffect, useRef, useState } from "react"
import { useShallow } from "zustand/react/shallow"

import { KefuConnectionStatus } from "@/components/kefu/connection-status"
import { KefuMessageEditor } from "@/components/kefu/message-editor"
import {
  KefuMessageList,
  type KefuMessageListHandle,
} from "@/components/kefu/message-list"
import {
  bindKefuHostBridge,
  requestKefuHostClose,
  requestKefuHostMinimize,
  requestKefuHostToggleMaximize,
} from "@/lib/kefu-host-bridge"
import { useKefuChatStore } from "@/lib/stores/kefu-chat"

export function KefuChatShell() {
  const messageListRef = useRef<KefuMessageListHandle | null>(null)
  const [isMaximized, setIsMaximized] = useState(false)
  const [isCloseDialogOpen, setIsCloseDialogOpen] = useState(false)
  const [isClosingConversation, setIsClosingConversation] = useState(false)

  const {
    title,
    subtitle,
    themeColor,
    conversation,
    messages,
    messagesHasMore,
    messagesLoadingMore,
    loadOlderMessages,
    status,
    error,
    isOpen,
    isVisible,
    setIsOpen,
    setIsVisible,
    bootstrap,
    handleSendMessage,
    uploadMessageImage,
    sendAttachment,
    retry,
    disconnectSocket,
    markConversationRead,
    closeConversation,
  } = useKefuChatStore(
    useShallow((state) => ({
      title: state.title,
      subtitle: state.subtitle,
      themeColor: state.themeColor,
      conversation: state.conversation,
      messages: state.messages,
      messagesHasMore: state.messagesHasMore,
      messagesLoadingMore: state.messagesLoadingMore,
      loadOlderMessages: state.loadOlderMessages,
      status: state.status,
      error: state.error,
      isOpen: state.isOpen,
      isVisible: state.isVisible,
      setIsOpen: state.setIsOpen,
      setIsVisible: state.setIsVisible,
      bootstrap: state.bootstrap,
      handleSendMessage: state.handleSendMessage,
      uploadMessageImage: state.uploadMessageImage,
      sendAttachment: state.sendAttachment,
      retry: state.retry,
      disconnectSocket: state.disconnectSocket,
      markConversationRead: state.markConversationRead,
      closeConversation: state.closeConversation,
    }))
  )
  const safeMessages = Array.isArray(messages) ? messages : []

  const maybeMarkConversationRead = useCallback(() => {
    if (!isVisible || !conversation || typeof document === "undefined") {
      return
    }
    if (document.visibilityState !== "visible") {
      return
    }
    void markConversationRead().catch((readError) => {
      console.error("Failed to mark kefu conversation read", readError)
    })
  }, [conversation, isVisible, markConversationRead])

  useEffect(() => {
    return bindKefuHostBridge({
      onOpen: () => {
        setIsOpen(true)
        setIsVisible(true)
      },
      onMinimize: () => {
        setIsVisible(false)
      },
      onMaximizedChange: (nextIsMaximized) => {
        setIsMaximized(nextIsMaximized)
      },
    })
  }, [setIsOpen, setIsVisible])

  useEffect(() => {
    bootstrap()

    return () => {
      if (!isOpen) {
        disconnectSocket()
      }
    }
  }, [isOpen, bootstrap, disconnectSocket])

  useEffect(() => {
    maybeMarkConversationRead()
  }, [maybeMarkConversationRead, safeMessages.length])

  useEffect(() => {
    const handleVisibilityChange = () => {
      if (document.visibilityState === "visible") {
        maybeMarkConversationRead()
      }
    }
    const handleFocus = () => {
      maybeMarkConversationRead()
    }

    document.addEventListener("visibilitychange", handleVisibilityChange)
    window.addEventListener("focus", handleFocus)
    return () => {
      document.removeEventListener("visibilitychange", handleVisibilityChange)
      window.removeEventListener("focus", handleFocus)
    }
  }, [maybeMarkConversationRead])

  async function handleSend(content: string) {
    await handleSendMessage(content)
    messageListRef.current?.scrollToBottom()
  }

  function handleMinimize() {
    setIsVisible(false)
    requestKefuHostMinimize()
  }

  function handleToggleMaximize() {
    requestKefuHostToggleMaximize()
  }

  async function confirmCloseConversation() {
    if (isClosingConversation) {
      return
    }
    setIsClosingConversation(true)
    try {
      if (conversation?.id) {
        await closeConversation()
      }
      setIsCloseDialogOpen(false)
      requestKefuHostClose()
    } catch (closeError) {
      window.alert(closeError instanceof Error ? closeError.message : "关闭会话失败")
    } finally {
      setIsClosingConversation(false)
    }
  }

  useEffect(() => {
    if (!isCloseDialogOpen) {
      return
    }
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape" && !isClosingConversation) {
        setIsCloseDialogOpen(false)
      }
    }
    window.addEventListener("keydown", handleKeyDown)
    return () => window.removeEventListener("keydown", handleKeyDown)
  }, [isCloseDialogOpen, isClosingConversation])

  return (
    <main
      className="cs-agent-shell relative flex h-screen overflow-hidden bg-(--background)"
      style={{ "--primary": themeColor } as React.CSSProperties}
    >
      <section className="cs-agent-panel flex h-full w-full flex-col overflow-hidden border border-white/70">
        <header className="relative shrink-0 overflow-hidden border-b border-white/60 px-4 pb-3 pt-3 shadow-[0_10px_24px_rgba(15,23,42,0.05)]">
          <div className="absolute inset-x-0 top-0 h-20 bg-[radial-gradient(circle_at_top_right,rgba(37,99,235,0.14),transparent_52%)]" />
          <div className="relative flex items-center justify-between gap-3">
            <div className="min-w-0">
              <div className="truncate text-[16px] font-semibold tracking-[0.01em] text-slate-950">
                {title}
              </div>
              <div className="mt-1 text-[12px] text-slate-500">{subtitle}</div>
            </div>
            <div className="flex items-center gap-2">
              {status !== "connected" ? (
                <KefuConnectionStatus status={status} />
              ) : null}
              <div className="inline-flex items-center gap-1 rounded-[18px] border border-white/70 bg-[linear-gradient(180deg,rgba(255,255,255,0.88),rgba(241,245,249,0.82))] p-1 shadow-[inset_0_1px_0_rgba(255,255,255,0.9),0_12px_28px_rgba(15,23,42,0.07)] backdrop-blur-xl">
                <button
                  type="button"
                  onClick={retry}
                  aria-label="重新连接"
                  title="重新连接"
                  className="group inline-flex h-6 w-6 items-center justify-center rounded-[14px] text-slate-400 transition duration-200 hover:-translate-y-0.5 hover:bg-[linear-gradient(180deg,rgba(255,255,255,0.98),rgba(239,246,255,0.96))] hover:text-sky-600 hover:shadow-[inset_0_1px_0_rgba(255,255,255,0.95),0_10px_18px_rgba(56,189,248,0.16)]"
                >
                  <RotateCwIcon className="size-3.75 transition duration-200 group-hover:rotate-[-20deg]" />
                </button>
                <button
                  type="button"
                  onClick={handleMinimize}
                  aria-label="收起聊天窗口"
                  title="收起聊天窗口"
                  className="group inline-flex h-6 w-6 items-center justify-center rounded-[14px] text-slate-400 transition duration-200 hover:-translate-y-0.5 hover:bg-[linear-gradient(180deg,rgba(255,255,255,0.98),rgba(248,250,252,0.96))] hover:text-slate-700 hover:shadow-[inset_0_1px_0_rgba(255,255,255,0.95),0_10px_18px_rgba(15,23,42,0.10)]"
                >
                  <MinusIcon className="size-3.75 transition duration-200 group-hover:scale-x-[0.88]" />
                </button>
                <button
                  type="button"
                  onClick={handleToggleMaximize}
                  aria-label={isMaximized ? "取消最大化" : "最大化聊天窗口"}
                  title={isMaximized ? "取消最大化" : "最大化聊天窗口"}
                  className="group inline-flex h-6 w-6 items-center justify-center rounded-[14px] text-slate-400 transition duration-200 hover:-translate-y-0.5 hover:bg-[linear-gradient(180deg,rgba(255,255,255,0.98),rgba(238,249,244,0.96))] hover:text-emerald-700 hover:shadow-[inset_0_1px_0_rgba(255,255,255,0.95),0_10px_18px_rgba(16,185,129,0.14)]"
                >
                  {isMaximized ? (
                    <Minimize2Icon className="size-3.75 transition duration-200 group-hover:scale-[0.94]" />
                  ) : (
                    <Maximize2Icon className="size-3.75 transition duration-200 group-hover:scale-[1.04]" />
                  )}
                </button>
                <button
                  type="button"
                  onClick={() => setIsCloseDialogOpen(true)}
                  aria-label="关闭聊天窗口"
                  title="关闭聊天窗口"
                  className="group inline-flex h-6 w-6 items-center justify-center rounded-[14px] text-rose-400 transition duration-200 hover:-translate-y-0.5 hover:bg-[linear-gradient(180deg,rgba(255,255,255,0.98),rgba(255,241,242,0.98))] hover:text-rose-600 hover:shadow-[inset_0_1px_0_rgba(255,255,255,0.95),0_10px_18px_rgba(244,63,94,0.16)]"
                >
                  <XIcon className="size-3.75 transition duration-200 group-hover:scale-[0.92]" />
                </button>
              </div>
            </div>
          </div>
        </header>

        <div className=".cs-agent-grid-bg grid min-h-0 flex-1 grid-rows-[minmax(0,1fr)_auto] overflow-hidden">
          <KefuMessageList
            ref={messageListRef}
            messages={safeMessages}
            onNearBottomVisible={maybeMarkConversationRead}
            hasMoreOlder={messagesHasMore}
            loadingOlder={messagesLoadingMore}
            onLoadOlder={loadOlderMessages}
          />
          <KefuMessageEditor
            disabled={!conversation}
            onSend={handleSend}
            onUploadImage={uploadMessageImage}
            onSendAttachment={sendAttachment}
          />
        </div>

        {error ? (
          <div className="border-t border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700">
            {error}
          </div>
        ) : null}
      </section>

      {isCloseDialogOpen ? (
        <div className="absolute inset-0 z-50 flex items-center justify-center bg-[radial-gradient(circle_at_top,rgba(37,99,235,0.16),transparent_38%),rgba(15,23,42,0.24)] px-5 backdrop-blur-sm">
          <div className="cs-agent-fade-up w-full max-w-[320px] rounded-[26px] border border-white/70 bg-[linear-gradient(180deg,rgba(255,255,255,0.96),rgba(241,245,249,0.94))] p-5 shadow-[0_28px_80px_rgba(15,23,42,0.18),inset_0_1px_0_rgba(255,255,255,0.92)]">
            <div className="flex items-start gap-3">
              <div className="min-w-0">
                <div className="text-[15px] font-semibold tracking-[0.01em] text-slate-950">
                  结束当前对话？
                </div>
                <div className="mt-1.5 text-[12px] leading-5 text-slate-500">
                  结束会话，客服将无法再查看您的消息记录，如需再次联系请重新发起对话。
                </div>
              </div>
            </div>

            <div className="mt-5 flex items-center justify-end gap-2">
              <button
                type="button"
                disabled={isClosingConversation}
                onClick={() => setIsCloseDialogOpen(false)}
                className="inline-flex h-9 items-center justify-center rounded-2xl border border-slate-200/80 bg-white/80 px-4 text-[12px] font-medium text-slate-600 shadow-[0_10px_20px_rgba(15,23,42,0.05)] transition hover:-translate-y-0.5 hover:border-slate-300 hover:bg-white hover:text-slate-800 disabled:translate-y-0 disabled:cursor-not-allowed disabled:opacity-60"
              >
                继续对话
              </button>
              <button
                type="button"
                disabled={isClosingConversation}
                onClick={() => void confirmCloseConversation()}
                className="inline-flex h-9 items-center justify-center rounded-2xl bg-[linear-gradient(135deg,#f43f5e,#fb7185)] px-4 text-[12px] font-semibold text-white shadow-[0_14px_28px_rgba(244,63,94,0.24)] transition hover:-translate-y-0.5 hover:opacity-92 disabled:translate-y-0 disabled:cursor-not-allowed disabled:opacity-60"
              >
                {isClosingConversation ? "结束中..." : "确认结束"}
              </button>
            </div>
          </div>
        </div>
      ) : null}
    </main>
  )
}
