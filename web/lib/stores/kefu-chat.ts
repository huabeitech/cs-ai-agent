"use client"

import { create } from "zustand"

import {
  closeImConversation,
  createOrMatchImConversation,
  fetchImMessages,
  fetchImWidgetConfig,
  markImMessageRead,
  sendImMessage,
  uploadImAttachment,
  uploadImImage,
  type ImAsset,
  type ImConversation,
  type ImMessage,
  type ImWidgetConfig,
} from "@/lib/api/im"
import {
  createImRealtimeConnection,
  type ImRealtimeEnvelope,
} from "@/lib/im-realtime"
import { summarizeIMMessage } from "@/lib/im-message"
import { generateUUID } from "@/lib/utils"

type ChatStatus = "connecting" | "connected" | "disconnected"

const RECONNECT_BASE_DELAY = 2000
const RECONNECT_MAX_DELAY = 30000
const DEFAULT_PAGE_LIMIT = 50

function getNotificationBody(message: ImMessage): string {
  return summarizeIMMessage(message)
}

function showNotification(title: string, body: string, onClick?: () => void) {
  if (typeof window === "undefined" || !("Notification" in window)) {
    return
  }

  const create = () => {
    const notification = new Notification(title, { body })
    notification.onclick = () => {
      window.focus()
      onClick?.()
      notification.close()
    }
  }

  if (Notification.permission === "granted") {
    create()
    return
  }

  if (Notification.permission === "default") {
    void Notification.requestPermission().then((permission) => {
      if (permission === "granted") {
        create()
      }
    })
  }
}

function mergeMessagesByIdAsc(a: ImMessage[], b: ImMessage[]): ImMessage[] {
  const byId = new Map<number, ImMessage>()
  for (const message of a) {
    byId.set(message.id, message)
  }
  for (const message of b) {
    byId.set(message.id, message)
  }
  return Array.from(byId.values()).sort((x, y) => x.id - y.id)
}

function ensureMessageList(value: ImMessage[] | null | undefined): ImMessage[] {
  return Array.isArray(value) ? value : []
}

function parseCursorId(cursor: string): number {
  const value = Number.parseInt(cursor, 10)
  return Number.isFinite(value) && value > 0 ? value : 0
}

function cursorFromLoadedMessages(messages: ImMessage[]): string {
  if (messages.length === 0) {
    return ""
  }
  return String(Math.min(...messages.map((message) => message.id)))
}

function minMessageId(messages: ImMessage[]): number | null {
  if (messages.length === 0) {
    return null
  }
  return Math.min(...messages.map((message) => message.id))
}

function hasMoreAfterLatestSyncMerge(args: {
  previousMessages: ImMessage[]
  previousHasMore: boolean
  merged: ImMessage[]
  apiHasMore: boolean
}): boolean {
  const prevMin = minMessageId(args.previousMessages)
  const mergedMin = minMessageId(args.merged)

  if (mergedMin === null) {
    return Boolean(args.apiHasMore)
  }

  if (!args.previousHasMore && prevMin !== null && mergedMin >= prevMin) {
    return false
  }

  return args.previousHasMore || Boolean(args.apiHasMore)
}

export type KefuChatStore = {
  title: string
  subtitle: string
  themeColor: string
  conversation: ImConversation | null
  messages: ImMessage[]
  messagesCursor: string
  messagesHasMore: boolean
  messagesLoadingMore: boolean
  initialized: boolean
  status: ChatStatus
  error: string
  sending: boolean
  uploadingAsset: boolean
  closingConversation: boolean
  isOpen: boolean
  isVisible: boolean
  socket: WebSocket | null
  readingMessageId: number

  setIsOpen: (isOpen: boolean) => void
  setIsVisible: (isVisible: boolean) => void
  bootstrap: () => void
  disconnectSocket: () => void
  refreshMessages: () => Promise<void>
  syncLatestMessages: () => Promise<void>
  loadOlderMessages: () => Promise<void>
  markConversationRead: () => Promise<void>
  handleSendMessage: (content: string) => Promise<void>
  sendMessage: (content: string) => Promise<void>
  uploadMessageImage: (file: File) => Promise<ImAsset | null>
  sendAttachment: (file: File) => Promise<void>
  closeConversation: () => Promise<void>
  retry: () => Promise<void>
}

let bootstrapToken = 0

export const useKefuChatStore = create<KefuChatStore>((set, get) => {
  let reconnectTimer: number | null = null
  let pingTimer: number | null = null
  let reconnectAttempt = 0
  let shouldReconnect = false

  const clearRealtimeTimers = () => {
    if (reconnectTimer !== null) {
      window.clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    if (pingTimer !== null) {
      window.clearInterval(pingTimer)
      pingTimer = null
    }
  }

  const scheduleReconnect = () => {
    if (!shouldReconnect || reconnectTimer !== null) {
      return
    }

    const delay = Math.min(
      RECONNECT_BASE_DELAY * 2 ** reconnectAttempt,
      RECONNECT_MAX_DELAY
    )
    set({ status: "connecting" })
    reconnectTimer = window.setTimeout(() => {
      reconnectTimer = null
      reconnectAttempt += 1
      if (!shouldReconnect || !get().isOpen) {
        return
      }
      connectSocket()
    }, delay)
  }

  const closeSocket = (options?: { reconnect?: boolean }) => {
    shouldReconnect = options?.reconnect ?? false
    clearRealtimeTimers()
    if (!shouldReconnect) {
      reconnectAttempt = 0
    }

    const socket = get().socket
    if (
      socket &&
      (socket.readyState === WebSocket.OPEN ||
        socket.readyState === WebSocket.CONNECTING)
    ) {
      socket.close()
    }

    set({ socket: null })
  }

  const connectSocket = () => {
    const conversationId = get().conversation?.id
    if (!conversationId) {
      return
    }

    closeSocket({ reconnect: false })
    shouldReconnect = true

    const socket = createImRealtimeConnection()
    set({ socket })

    const handleRealtimeEvent = (event: ImRealtimeEnvelope) => {
      const payload = event.data ?? event.payload
      const needsRefresh =
        event.type === "message.created" ||
        event.type?.startsWith("conversation.")

      if (needsRefresh && payload?.conversationId === conversationId) {
        void get()
          .syncLatestMessages()
          .then(() => {
            if (event.type !== "message.created") {
              return
            }
            const state = get()
            const lastMessage = state.messages.at(-1)
            if (
              lastMessage &&
              lastMessage.senderType !== "customer" &&
              typeof document !== "undefined" &&
              document.visibilityState !== "visible"
            ) {
              showNotification("新消息", getNotificationBody(lastMessage), () => {
                state.setIsOpen(true)
                state.setIsVisible(true)
              })
            }
          })
      }
    }

    socket.addEventListener("message", (event) => {
      try {
        handleRealtimeEvent(JSON.parse(event.data) as ImRealtimeEnvelope)
      } catch {
        return
      }
    })

    socket.addEventListener("open", () => {
      clearRealtimeTimers()
      reconnectAttempt = 0
      pingTimer = window.setInterval(() => {
        if (socket.readyState === WebSocket.OPEN) {
          socket.send(JSON.stringify({ type: "ping" }))
        }
      }, 20000)
      if (get().isOpen && get().socket === socket) {
        set({ status: "connected" })
      }
    })

    socket.addEventListener("error", () => {
      if (get().socket === socket) {
        scheduleReconnect()
      }
    })

    socket.addEventListener("close", () => {
      if (pingTimer !== null) {
        window.clearInterval(pingTimer)
        pingTimer = null
      }
      if (get().socket === socket) {
        set({ socket: null })
      }
      if (get().isOpen) {
        if (shouldReconnect) {
          scheduleReconnect()
        } else {
          set({ status: "disconnected" })
        }
      }
    })
  }

  return {
    title: "在线客服",
    subtitle: "",
    themeColor: "#2563eb",
    conversation: null,
    messages: [],
    messagesCursor: "",
    messagesHasMore: false,
    messagesLoadingMore: false,
    initialized: false,
    status: "connecting",
    error: "",
    sending: false,
    uploadingAsset: false,
    closingConversation: false,
    isOpen: typeof window !== "undefined" ? window.self === window.top : false,
    isVisible:
      typeof window !== "undefined" ? window.self === window.top : false,
    socket: null,
    readingMessageId: 0,

    setIsOpen: (isOpen: boolean) => {
      set({ isOpen })
    },

    setIsVisible: (isVisible: boolean) => {
      set({ isVisible })
    },

    bootstrap: () => {
      const token = ++bootstrapToken

      if (!get().isOpen) {
        closeSocket({ reconnect: false })
        set({ status: "disconnected" })
        return
      }

      const activateChat = async () => {
        try {
          set({ error: "", status: "connecting" })

          const widgetConfig: ImWidgetConfig = await fetchImWidgetConfig().catch(
            () => ({})
          )
          if (bootstrapToken !== token || !get().isOpen) {
            return
          }

          set({
            title: widgetConfig.title || "在线客服",
            subtitle: widgetConfig.subtitle || "",
            themeColor: widgetConfig.themeColor || "#2563eb",
          })

          let currentConversation = get().conversation
          if (!get().initialized || !currentConversation) {
            currentConversation = await createOrMatchImConversation()
            if (bootstrapToken !== token || !get().isOpen) {
              return
            }
            set({ initialized: true, conversation: currentConversation })
          }

          await get().refreshMessages()
          if (bootstrapToken !== token || !get().isOpen) {
            return
          }

          connectSocket()
        } catch (error) {
          if (bootstrapToken !== token || !get().isOpen) {
            return
          }
          set({
            status: "disconnected",
            error: error instanceof Error ? error.message : "初始化失败",
          })
        }
      }

      void activateChat()
    },

    disconnectSocket: () => {
      closeSocket({ reconnect: false })
    },

    refreshMessages: async () => {
      const conversationId = get().conversation?.id
      if (!conversationId) {
        return
      }

      try {
        const page = await fetchImMessages({
          conversationId,
          limit: DEFAULT_PAGE_LIMIT,
        })
        const results = ensureMessageList(page.results)
        set({
          messages: results,
          messagesCursor: cursorFromLoadedMessages(results) || page.cursor || "",
          messagesHasMore: Boolean(page.hasMore) || results.length >= DEFAULT_PAGE_LIMIT,
        })
      } catch (error) {
        set({
          error: error instanceof Error ? error.message : "加载消息失败",
        })
        throw error
      }
    },

    syncLatestMessages: async () => {
      const conversationId = get().conversation?.id
      if (!conversationId) {
        return
      }

      try {
        const page = await fetchImMessages({
          conversationId,
          limit: DEFAULT_PAGE_LIMIT,
        })
        const batch = ensureMessageList(page.results)
        if (batch.length === 0) {
          return
        }
        const firstId = batch[0]!.id
        set((state) => {
          const preserved = state.messages.filter((message) => message.id < firstId)
          const merged = mergeMessagesByIdAsc(preserved, batch)
          return {
            messages: merged,
            messagesCursor: cursorFromLoadedMessages(merged) || page.cursor || "",
            messagesHasMore: hasMoreAfterLatestSyncMerge({
              previousMessages: state.messages,
              previousHasMore: state.messagesHasMore,
              merged,
              apiHasMore: Boolean(page.hasMore) || batch.length >= DEFAULT_PAGE_LIMIT,
            }),
          }
        })
      } catch (error) {
        set({
          error: error instanceof Error ? error.message : "同步消息失败",
        })
      }
    },

    loadOlderMessages: async () => {
      const conversationId = get().conversation?.id
      if (
        !conversationId ||
        get().messagesLoadingMore ||
        !get().messagesHasMore
      ) {
        return
      }

      const cursorId = parseCursorId(get().messagesCursor)
      if (cursorId <= 0) {
        return
      }

      set({ messagesLoadingMore: true })
      try {
        const page = await fetchImMessages({
          conversationId,
          cursor: cursorId,
          limit: DEFAULT_PAGE_LIMIT,
        })
        const results = ensureMessageList(page.results)
        set((state) => {
          const merged = mergeMessagesByIdAsc(results, ensureMessageList(state.messages))
          return {
            messages: merged,
            messagesCursor: cursorFromLoadedMessages(merged) || page.cursor || "",
            messagesHasMore: Boolean(page.hasMore) || results.length >= DEFAULT_PAGE_LIMIT,
            messagesLoadingMore: false,
          }
        })
      } catch (error) {
        set({
          messagesLoadingMore: false,
          error: error instanceof Error ? error.message : "加载历史消息失败",
        })
        throw error
      }
    },

    markConversationRead: async () => {
      const state = get()
      const conversation = state.conversation
      const lastMessage = state.messages.at(-1)
      if (!conversation?.id || !lastMessage) {
        return
      }

      if (
        conversation.customerUnreadCount <= 0 &&
        conversation.customerLastReadMessageId >= lastMessage.id
      ) {
        return
      }
      if (state.readingMessageId === lastMessage.id) {
        return
      }

      set({ readingMessageId: lastMessage.id })
      try {
        await markImMessageRead(conversation.id, lastMessage.id)
        set((current) => ({
          readingMessageId: 0,
          messages: current.messages.map((item) =>
            (item.seqNo ?? 0) <= (lastMessage.seqNo ?? 0)
              ? { ...item, customerRead: true }
              : item
          ),
          conversation: current.conversation
            ? {
                ...current.conversation,
                customerUnreadCount: 0,
                customerLastReadMessageId: lastMessage.id,
                customerLastReadSeqNo: lastMessage.seqNo,
              }
            : null,
        }))
      } catch (error) {
        set({ readingMessageId: 0 })
        throw error
      }
    },

    handleSendMessage: async (content: string) => {
      const conversationId = get().conversation?.id
      if (!conversationId) {
        return
      }

      set({ error: "", sending: true })
      try {
        const nextMessage = await sendImMessage({
          conversationId,
          messageType: "html",
          content,
          clientMsgId: `kefu_html_${generateUUID()}`,
        })
        set((state) => ({
          sending: false,
          messages: state.messages.some((message) => message.id === nextMessage.id)
            ? state.messages.map((message) =>
                message.id === nextMessage.id ? nextMessage : message
              )
            : [...state.messages, nextMessage],
          conversation: state.conversation
            ? {
                ...state.conversation,
                customerLastReadMessageId: nextMessage.id,
                customerLastReadSeqNo: nextMessage.seqNo,
                customerUnreadCount: 0,
                lastMessageAt: nextMessage.sentAt,
                lastMessageSummary: summarizeIMMessage(nextMessage),
              }
            : null,
        }))
      } catch (error) {
        set({
          sending: false,
          error: error instanceof Error ? error.message : "发送消息失败",
        })
        throw error
      }
    },

    sendMessage: async (content: string) => {
      return get().handleSendMessage(content)
    },

    uploadMessageImage: async (file: File) => {
      const conversationId = get().conversation?.id
      if (!conversationId) {
        return null
      }

      set({ error: "", uploadingAsset: true })
      try {
        return await uploadImImage(conversationId, file)
      } catch (error) {
        set({
          error: error instanceof Error ? error.message : "上传图片失败",
        })
        return null
      } finally {
        set({ uploadingAsset: false })
      }
    },

    sendAttachment: async (file: File) => {
      const conversationId = get().conversation?.id
      if (!conversationId) {
        return
      }

      set({ error: "", uploadingAsset: true })
      try {
        const asset = await uploadImAttachment(conversationId, file)
        const nextMessage = await sendImMessage({
          conversationId,
          messageType: "attachment",
          content: asset.filename,
          payload: JSON.stringify({ assetId: asset.assetId }),
          clientMsgId: `kefu_attachment_${generateUUID()}`,
        })
        set((state) => ({
          uploadingAsset: false,
          messages: state.messages.some((message) => message.id === nextMessage.id)
            ? state.messages.map((message) =>
                message.id === nextMessage.id ? nextMessage : message
              )
            : [...state.messages, nextMessage],
          conversation: state.conversation
            ? {
                ...state.conversation,
                customerLastReadMessageId: nextMessage.id,
                customerLastReadSeqNo: nextMessage.seqNo,
                customerUnreadCount: 0,
                lastMessageAt: nextMessage.sentAt,
                lastMessageSummary: summarizeIMMessage(nextMessage),
              }
            : null,
        }))
      } catch (error) {
        set({
          uploadingAsset: false,
          error: error instanceof Error ? error.message : "发送附件失败",
        })
        throw error
      }
    },

    closeConversation: async () => {
      const conversationId = get().conversation?.id
      if (!conversationId) {
        return
      }

      set({ error: "", closingConversation: true })
      try {
        await closeImConversation(conversationId)
        closeSocket({ reconnect: false })
        set((state) => ({
          closingConversation: false,
          status: "disconnected",
          conversation: state.conversation
            ? {
                ...state.conversation,
                status: 2,
              }
            : null,
        }))
      } catch (error) {
        set({
          closingConversation: false,
          error: error instanceof Error ? error.message : "关闭会话失败",
        })
        throw error
      }
    },

    retry: async () => {
      if (!get().conversation?.id) {
        return
      }

      set({ error: "", status: "connecting" })
      try {
        await get().refreshMessages()
        if (get().isOpen) {
          shouldReconnect = true
          connectSocket()
        }
      } catch (error) {
        set({
          status: "disconnected",
          error: error instanceof Error ? error.message : "刷新失败",
        })
      }
    },
  }
})
