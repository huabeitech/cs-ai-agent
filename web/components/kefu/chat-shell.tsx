"use client"

import {
  Maximize2Icon,
  Minimize2Icon,
  MinusIcon,
  RotateCwIcon,
  XIcon,
} from "lucide-react"
import {
  useCallback,
  useEffect,
  useLayoutEffect,
  useRef,
  useState,
  type CSSProperties,
} from "react"
import { useShallow } from "zustand/react/shallow"

import { KefuConnectionStatus } from "@/components/kefu/connection-status"
import { CustomerMessageEditor } from "@/components/kefu/customer-message-editor"
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
import { Button } from "@/components/ui/button"
import { ButtonGroup } from "@/components/ui/button-group"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"

function useKefuSystemTheme() {
  useLayoutEffect(() => {
    if (typeof window === "undefined") {
      return
    }
    const root = document.documentElement
    const query = window.matchMedia("(prefers-color-scheme: dark)")
    const previousDarkClass = root.classList.contains("dark")
    const previousColorScheme = root.style.colorScheme

    const syncTheme = () => {
      const isDark = query.matches
      root.classList.toggle("dark", isDark)
      root.style.colorScheme = isDark ? "dark" : "light"
    }

    syncTheme()
    query.addEventListener("change", syncTheme)

    return () => {
      query.removeEventListener("change", syncTheme)
      root.classList.toggle("dark", previousDarkClass)
      root.style.colorScheme = previousColorScheme
    }
  }, [])
}

export function KefuChatShell() {
  useKefuSystemTheme()

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
      className="relative flex h-screen overflow-hidden bg-muted text-foreground"
      style={{ "--primary": themeColor } as CSSProperties}
    >
      <section className="flex h-full w-full flex-col overflow-hidden border border-border bg-card text-card-foreground">
        <header className="shrink-0 border-b border-border bg-card/95 px-4 py-3 shadow-[0_8px_24px_rgba(15,23,42,0.04)] dark:shadow-none">
          <div className="flex items-center justify-between gap-3">
            <div className="min-w-0">
              <div className="truncate text-base font-semibold text-foreground">
                {title}
              </div>
              <div className="mt-1 truncate text-xs text-muted-foreground">{subtitle}</div>
            </div>
            <div className="flex items-center gap-2">
              {status !== "connected" ? (
                <KefuConnectionStatus status={status} />
              ) : null}
              <ButtonGroup>
                <Button
                  type="button"
                  variant="outline"
                  size="icon-sm"
                  onClick={retry}
                  aria-label="重新连接"
                  title="重新连接"
                  className="bg-background text-muted-foreground hover:text-sky-600 dark:hover:text-sky-400"
                >
                  <RotateCwIcon className="size-4" />
                </Button>
                <Button
                  type="button"
                  variant="outline"
                  size="icon-sm"
                  onClick={handleMinimize}
                  aria-label="收起聊天窗口"
                  title="收起聊天窗口"
                  className="bg-background text-muted-foreground hover:text-foreground"
                >
                  <MinusIcon className="size-4" />
                </Button>
                <Button
                  type="button"
                  variant="outline"
                  size="icon-sm"
                  onClick={handleToggleMaximize}
                  aria-label={isMaximized ? "取消最大化" : "最大化聊天窗口"}
                  title={isMaximized ? "取消最大化" : "最大化聊天窗口"}
                  className="bg-background text-muted-foreground hover:text-emerald-700 dark:hover:text-emerald-400"
                >
                  {isMaximized ? (
                    <Minimize2Icon className="size-4" />
                  ) : (
                    <Maximize2Icon className="size-4" />
                  )}
                </Button>
                <Button
                  type="button"
                  variant="outline"
                  size="icon-sm"
                  onClick={() => setIsCloseDialogOpen(true)}
                  aria-label="关闭聊天窗口"
                  title="关闭聊天窗口"
                  className="bg-background text-rose-500 hover:bg-rose-50 hover:text-rose-600 dark:hover:bg-rose-950/40 dark:hover:text-rose-300"
                >
                  <XIcon className="size-4" />
                </Button>
              </ButtonGroup>
            </div>
          </div>
        </header>

        <div className="grid min-h-0 flex-1 grid-rows-[minmax(0,1fr)_auto] overflow-hidden bg-muted/70">
          <KefuMessageList
            ref={messageListRef}
            messages={safeMessages}
            onNearBottomVisible={maybeMarkConversationRead}
            hasMoreOlder={messagesHasMore}
            loadingOlder={messagesLoadingMore}
            onLoadOlder={loadOlderMessages}
          />
          <CustomerMessageEditor
            disabled={!conversation}
            onSend={handleSend}
            onUploadImage={uploadMessageImage}
            onSendAttachment={sendAttachment}
          />
        </div>

        {error ? (
          <div className="border-t border-destructive/20 bg-destructive/10 px-4 py-3 text-sm text-destructive">
            {error}
          </div>
        ) : null}
      </section>

      <Dialog
        open={isCloseDialogOpen}
        onOpenChange={(open) => {
          if (!isClosingConversation) {
            setIsCloseDialogOpen(open)
          }
        }}
      >
        <DialogContent className="max-w-[320px]" showCloseButton={!isClosingConversation}>
          <DialogHeader>
            <DialogTitle>结束当前对话？</DialogTitle>
            <DialogDescription className="text-xs leading-5">
              结束会话，客服将无法再查看您的消息记录，如需再次联系请重新发起对话。
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              disabled={isClosingConversation}
              onClick={() => setIsCloseDialogOpen(false)}
            >
              继续对话
            </Button>
            <Button
              type="button"
              variant="destructive"
              disabled={isClosingConversation}
              onClick={() => void confirmCloseConversation()}
            >
              {isClosingConversation ? "结束中..." : "确认结束"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </main>
  )
}
