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
import { readSession } from "@/lib/auth"

export const agentConversationFilterOptions = [
  { value: "my-active", label: "我处理中的" },
  { value: "pending", label: "待接入" },
  { value: "active", label: "处理中" },
  { value: "closed", label: "已关闭" },
] as const

export type AgentConversationFilterKey =
  (typeof agentConversationFilterOptions)[number]["value"]

function buildConversationQuery(filter: AgentConversationFilterKey, keyword: string) {
  const session = readSession()
  const query: Record<string, string | number | undefined> = {
    keyword: keyword.trim() || undefined,
    limit: 50,
  }

  switch (filter) {
    case "my-active":
      query.status = 2
      query.currentAssigneeId = session?.user.id
      break
    case "pending":
      query.status = 1
      break
    case "active":
      query.status = 2
      break
    case "closed":
      query.status = 3
      break
  }

  return query
}

type LoadMessagesOptions = {
  forceLoading?: boolean
  reset?: boolean
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
  messagesLoadedConversationId: number | null
  sending: boolean
  uploadingImage: boolean
  readingMessageId: number
  setSearchKeyword: (keyword: string) => void
  setConversationFilter: (filter: AgentConversationFilterKey) => void
  loadConversations: () => Promise<void>
  selectConversation: (conversationId: number) => Promise<void>
  loadMessages: (conversationId: number, options?: LoadMessagesOptions) => Promise<void>
  markSelectedConversationRead: () => Promise<void>
  sendMessage: (html: string) => Promise<AgentMessage | null>
  uploadImage: (file: File) => Promise<AgentAsset | null>
}

let conversationsRequestSeq = 0
let messagesRequestSeq = 0

export const useAgentConversationsStore = create<AgentConversationsStore>((set, get) => ({
  searchKeyword: "",
  conversationFilter: "my-active",
  conversations: [],
  conversationsLoading: false,
  conversationsLoaded: false,
  selectedConversationId: null,
  messages: [],
  messagesLoading: false,
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

      if (requestSeq !== conversationsRequestSeq) {
        return
      }

      const currentSelectedId = get().selectedConversationId
      const hasCurrentSelection =
        currentSelectedId !== null && data.results.some((item) => item.id === currentSelectedId)
      const nextSelectedId = hasCurrentSelection ? currentSelectedId : (data.results[0]?.id ?? null)
      const selectionChanged = nextSelectedId !== currentSelectedId

      set({
        conversations: data.results,
        conversationsLoaded: true,
        conversationsLoading: false,
        selectedConversationId: nextSelectedId,
      })

      if (nextSelectedId === null) {
        set({
          messages: [],
          messagesLoading: false,
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
        ...(options.reset ? { messages: [] } : {}),
      })
    }

    try {
      const data = await fetchAgentMessages({
        conversationId,
        limit: 100,
      })

      if (requestSeq !== messagesRequestSeq) {
        return
      }

      if (get().selectedConversationId !== conversationId) {
        return
      }

      set({
        messages: data.results,
        messagesLoading: false,
        messagesLoadedConversationId: conversationId,
      })
    } catch (error) {
      if (requestSeq === messagesRequestSeq) {
        set({ messagesLoading: false })
      }
      throw error
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
          messages: [...current.messages, message],
          conversations: current.conversations.map((item) =>
            item.id === selectedConversationId
              ? {
                  ...item,
                  lastMessageAt: message.sentAt,
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
