"use client"

import { create } from "zustand"

import {
  fetchAgentConversations,
  fetchAgentMessages,
  markAgentMessageRead,
  sendAgentMessage,
  uploadAgentConversationImage,
  type AgentAsset,
  type AgentConversation,
  type AgentMessage,
} from "@/lib/api/agent"

export const agentConversationFilterOptions = [
  { value: "mine", label: "我的" },
  { value: "active", label: "处理中" },
  { value: "pending", label: "待接入" },
  // { value: "closed", label: "已关闭" },
] as const

export type AgentConversationFilterKey =
  (typeof agentConversationFilterOptions)[number]["value"]

function buildConversationQuery(filter: AgentConversationFilterKey, keyword: string) {
  const query: Record<string, string | number | undefined> = {
    filter,
    keyword: keyword.trim() || undefined,
    limit: 100,
  }

  return query
}

type LoadMessagesOptions = {
  forceLoading?: boolean
  reset?: boolean
}

function ensureArray<T>(value: T[] | null | undefined): T[] {
  return Array.isArray(value) ? value : []
}

function mergeMessagesByIdAsc(
  a: AgentMessage[],
  b: AgentMessage[]
): AgentMessage[] {
  const byId = new Map<number, AgentMessage>()
  for (const m of a) {
    byId.set(m.id, m)
  }
  for (const m of b) {
    byId.set(m.id, m)
  }
  return Array.from(byId.values()).sort((x, y) => x.id - y.id)
}

function parseCursorId(cursor: string): number {
  const n = Number.parseInt(cursor, 10)
  return Number.isFinite(n) && n > 0 ? n : 0
}

/** 下一页「更旧」请求应传入的游标：当前已加载列表中的最小 message id（后端用 id < cursor） */
function cursorFromLoadedMessages(messages: AgentMessage[]): string {
  if (messages.length === 0) {
    return ""
  }
  return String(Math.min(...messages.map((m) => m.id)))
}

type AgentConversationsStore = {
  searchKeyword: string
  conversationFilter: AgentConversationFilterKey
  conversations: AgentConversation[]
  conversationsLoading: boolean
  conversationsLoaded: boolean
  selectedConversationId: number | null
  messages: AgentMessage[]
  messagesLoading: boolean
  messagesLoadingMore: boolean
  messagesCursor: string
  messagesHasMore: boolean
  messagesLoadedConversationId: number | null
  sending: boolean
  uploadingImage: boolean
  readingMessageId: number
  setSearchKeyword: (keyword: string) => void
  setConversationFilter: (filter: AgentConversationFilterKey) => void
  loadConversations: () => Promise<void>
  selectConversation: (conversationId: number) => Promise<void>
  loadMessages: (conversationId: number, options?: LoadMessagesOptions) => Promise<void>
  loadOlderMessages: () => Promise<void>
  syncLatestMessages: (conversationId: number) => Promise<void>
  markSelectedConversationRead: () => Promise<void>
  sendMessage: (html: string) => Promise<AgentMessage | null>
  uploadImage: (file: File) => Promise<AgentAsset | null>
}

let conversationsRequestSeq = 0
let messagesRequestSeq = 0

export const useAgentConversationsStore = create<AgentConversationsStore>((set, get) => ({
  searchKeyword: "",
  conversationFilter: "mine",
  conversations: [],
  conversationsLoading: false,
  conversationsLoaded: false,
  selectedConversationId: null,
  messages: [],
  messagesLoading: false,
  messagesLoadingMore: false,
  messagesCursor: "",
  messagesHasMore: false,
  messagesLoadedConversationId: null,
  sending: false,
  uploadingImage: false,
  readingMessageId: 0,

  setSearchKeyword: (keyword) => {
    set({ searchKeyword: keyword })
  },

  setConversationFilter: (filter) => {
    set({ conversationFilter: filter })
  },

  loadConversations: async () => {
    const requestSeq = ++conversationsRequestSeq
    const store = get()

    if (!store.conversationsLoaded) {
      set({ conversationsLoading: true })
    }

    try {
      const data = await fetchAgentConversations(
        buildConversationQuery(store.conversationFilter, store.searchKeyword)
      )
      const conversations = ensureArray(data.results)

      if (requestSeq !== conversationsRequestSeq) {
        return
      }

      const currentSelectedId = get().selectedConversationId
      const hasCurrentSelection =
        currentSelectedId !== null && conversations.some((item) => item.id === currentSelectedId)
      const nextSelectedId = hasCurrentSelection ? currentSelectedId : (conversations[0]?.id ?? null)
      const selectionChanged = nextSelectedId !== currentSelectedId

      set({
        conversations,
        conversationsLoaded: true,
        conversationsLoading: false,
        selectedConversationId: nextSelectedId,
      })

      if (nextSelectedId === null) {
        set({
          messages: [],
          messagesLoading: false,
          messagesLoadingMore: false,
          messagesCursor: "",
          messagesHasMore: false,
          messagesLoadedConversationId: null,
        })
        return
      }

      if (selectionChanged || get().messagesLoadedConversationId === null) {
        await get().loadMessages(nextSelectedId, {
          forceLoading: true,
          reset: true,
        })
      }
    } catch (error) {
      if (requestSeq === conversationsRequestSeq) {
        set({ conversationsLoading: false })
      }
      throw error
    }
  },

  selectConversation: async (conversationId) => {
    if (get().selectedConversationId === conversationId) {
      return
    }

    set({
      selectedConversationId: conversationId,
      messages: [],
      messagesLoading: true,
      messagesLoadingMore: false,
      messagesCursor: "",
      messagesHasMore: false,
      messagesLoadedConversationId: null,
    })

    await get().loadMessages(conversationId, {
      forceLoading: true,
      reset: true,
    })
  },

  loadMessages: async (conversationId, options = {}) => {
    const requestSeq = ++messagesRequestSeq
    const store = get()
    const shouldShowLoading =
      options.forceLoading || store.messagesLoadedConversationId !== conversationId

    if (shouldShowLoading) {
      set({
        messagesLoading: true,
        ...(options.reset
          ? {
              messages: [],
              messagesCursor: "",
              messagesHasMore: false,
            }
          : {}),
      })
    }

    try {
      const data = await fetchAgentMessages({
        conversationId,
        limit: 50,
      })

      if (requestSeq !== messagesRequestSeq) {
        return
      }

      if (get().selectedConversationId !== conversationId) {
        return
      }

      const list = ensureArray(data.results)
      set({
        messages: list,
        messagesLoading: false,
        messagesLoadedConversationId: conversationId,
        messagesCursor:
          cursorFromLoadedMessages(list) || (data.cursor ?? ""),
        messagesHasMore: Boolean(data.hasMore),
      })
    } catch (error) {
      if (requestSeq === messagesRequestSeq) {
        set({ messagesLoading: false })
      }
      throw error
    }
  },

  loadOlderMessages: async () => {
    const conversationId = get().selectedConversationId
    if (!conversationId || get().messagesLoadingMore || !get().messagesHasMore) {
      return
    }
    const cursorId = parseCursorId(get().messagesCursor)
    if (cursorId <= 0) {
      return
    }

    set({ messagesLoadingMore: true })
    try {
      const data = await fetchAgentMessages({
        conversationId,
        cursor: cursorId,
        limit: 50,
      })
      if (get().selectedConversationId !== conversationId) {
        return
      }
      const incoming = ensureArray(data.results)
      set((state) => {
        const merged = mergeMessagesByIdAsc(incoming, state.messages)
        return {
          messages: merged,
          messagesCursor:
            cursorFromLoadedMessages(merged) ||
            (data.cursor ?? state.messagesCursor),
          messagesHasMore: Boolean(data.hasMore),
          messagesLoadingMore: false,
        }
      })
    } catch (error) {
      set({ messagesLoadingMore: false })
      throw error
    }
  },

  syncLatestMessages: async (conversationId) => {
    if (conversationId <= 0) {
      return
    }
    try {
      const data = await fetchAgentMessages({
        conversationId,
        limit: 50,
      })
      if (get().selectedConversationId !== conversationId) {
        return
      }
      const batch = ensureArray(data.results)
      if (batch.length === 0) {
        return
      }
      const firstId = batch[0]!.id
      set((state) => {
        const preserved = state.messages.filter((m) => m.id < firstId)
        const merged = mergeMessagesByIdAsc(preserved, batch)
        return {
          messages: merged,
          messagesCursor:
            cursorFromLoadedMessages(merged) ||
            (data.cursor ?? state.messagesCursor),
          // 接口 hasMore 只反映「最新一页」以下是否还有；已 prepend 的历史要保留「还能往上翻」
          messagesHasMore:
            state.messagesHasMore || Boolean(data.hasMore),
        }
      })
    } catch {
      // 实时同步失败不抛给 WS 回调
    }
  },

  markSelectedConversationRead: async () => {
    const store = get()
    const conversationId = store.selectedConversationId
    const conversation = store.conversations.find((item) => item.id === conversationId)
    const lastMessage = store.messages.at(-1)
    if (!conversationId || !conversation || !lastMessage) {
      return
    }
    if (
      conversation.agentUnreadCount <= 0 &&
      (conversation.agentLastReadMessageId ?? 0) >= lastMessage.id
    ) {
      return
    }
    if (store.readingMessageId === lastMessage.id) {
      return
    }

    set({ readingMessageId: lastMessage.id })
    try {
      await markAgentMessageRead(conversationId, lastMessage.id)
      set((current) => {
        if (current.selectedConversationId !== conversationId) {
          return { readingMessageId: 0 }
        }
        return {
          readingMessageId: 0,
          messages: current.messages.map((item) =>
            item.seqNo <= lastMessage.seqNo
              ? {
                  ...item,
                  agentRead: true,
                }
              : item
          ),
          conversations: current.conversations.map((item) =>
            item.id === conversationId
              ? {
                  ...item,
                  agentUnreadCount: 0,
                  agentLastReadMessageId: lastMessage.id,
                  agentLastReadSeqNo: lastMessage.seqNo,
                }
              : item
          ),
        }
      })
    } catch (error) {
      set({ readingMessageId: 0 })
      throw error
    }
  },

  sendMessage: async (html) => {
    const trimmedContent = html.trim()
    const { selectedConversationId, sending } = get()
    if (!selectedConversationId || !trimmedContent || sending) {
      return null
    }

    set({ sending: true })
    try {
      const message = await sendAgentMessage({
        conversationId: selectedConversationId,
        messageType: "html",
        content: trimmedContent,
        clientMsgId: `agent_${crypto.randomUUID()}`,
      })

      if (get().selectedConversationId === selectedConversationId) {
        set((current) => ({
          messages: current.messages.some((m) => m.id === message.id)
            ? current.messages.map((m) => (m.id === message.id ? message : m))
            : [...current.messages, message],
          conversations: current.conversations.map((item) =>
            item.id === selectedConversationId
              ? {
                  ...item,
                  lastMessageAt: message.sentAt,
                  lastActiveAt: message.sentAt,
                  lastMessageSummary: summarizeHTML(trimmedContent),
                  agentUnreadCount: 0,
                  customerUnreadCount: (item.customerUnreadCount ?? 0) + 1,
                  agentLastReadMessageId: message.id,
                  agentLastReadSeqNo: message.seqNo,
                }
              : item
          ),
        }))
      }

      return message
    } finally {
      set({ sending: false })
    }
  },

  uploadImage: async (file) => {
    const { selectedConversationId, sending, uploadingImage } = get()
    if (!selectedConversationId || sending || uploadingImage) {
      return null
    }

    set({ uploadingImage: true })
    try {
      return await uploadAgentConversationImage(file)
    } finally {
      set({ uploadingImage: false })
    }
  },
}))

export const agentConversationSelectors = {
  selectedConversation: (state: AgentConversationsStore) =>
    state.conversations.find((item) => item.id === state.selectedConversationId) ?? null,
}

function summarizeHTML(html: string) {
  const withImagePlaceholder = html.replace(/<img[\s\S]*?>/gi, " [图片] ")
  const plainText = withImagePlaceholder.replace(/<[^>]+>/g, " ").replace(/\s+/g, " ").trim()
  return plainText || "[图片]"
}
