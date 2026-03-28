"use client";

import { create } from "zustand";

import { createOrMatchConversation } from "@/lib/services/conversation";
import {
  fetchMessages,
  markMessageRead,
  sendMessage,
  uploadImage,
} from "@/lib/services/message";
import { getNotificationBody, showNotification } from "@/lib/services/notification";
import {
  createRealtimeConnection,
  type RealtimeEnvelope,
} from "@/lib/services/realtime";
import { fetchWidgetConfig, readWidgetConfig } from "@/lib/services/widget-config";
import type {
  WidgetConfigResponse,
  WidgetConversation,
  WidgetMessage,
} from "@/lib/services/types";

type ChatStatus = "connecting" | "connected" | "disconnected";

const RECONNECT_BASE_DELAY = 2000;
const RECONNECT_MAX_DELAY = 30000;

export interface ChatStore {
  title: string;
  welcomeText: string;
  themeColor: string;
  conversation: WidgetConversation | null;
  messages: WidgetMessage[];
  status: ChatStatus;
  error: string;
  isOpen: boolean;
  isVisible: boolean;
  initialized: boolean;
  socket: WebSocket | null;
  readingMessageId: number;

  setIsOpen: (isOpen: boolean) => void;
  setIsVisible: (isVisible: boolean) => void;
  bootstrap: () => void;
  handleSendMessage: (html: string) => Promise<void>;
  uploadMessageImage: (file: File) => Promise<{ url: string; filename?: string } | null>;
  retry: () => void;
  disconnectSocket: () => void;
  refreshMessages: () => Promise<void>;
  markConversationRead: () => Promise<void>;
}

let bootstrapToken = 0;

export const useChatStore = create<ChatStore>((set, get) => {
  let reconnectTimer: number | null = null;
  let pingTimer: number | null = null;
  let reconnectAttempt = 0;
  let shouldReconnect = false;

  const clearRealtimeTimers = () => {
    if (reconnectTimer !== null) {
      window.clearTimeout(reconnectTimer);
      reconnectTimer = null;
    }
    if (pingTimer !== null) {
      window.clearInterval(pingTimer);
      pingTimer = null;
    }
  };

  const scheduleReconnect = () => {
    if (!shouldReconnect || reconnectTimer !== null) {
      return;
    }

    const delay = Math.min(
      RECONNECT_BASE_DELAY * 2 ** reconnectAttempt,
      RECONNECT_MAX_DELAY,
    );
    set({ status: "connecting" });
    reconnectTimer = window.setTimeout(() => {
      reconnectTimer = null;
      reconnectAttempt += 1;
      if (!shouldReconnect || !get().isOpen) {
        return;
      }
      connectSocket();
    }, delay);
  };

  const closeSocket = (options?: { reconnect?: boolean }) => {
    shouldReconnect = options?.reconnect ?? false;
    clearRealtimeTimers();
    if (!shouldReconnect) {
      reconnectAttempt = 0;
    }

    const socket = get().socket;
    if (
      socket &&
      (socket.readyState === WebSocket.OPEN ||
        socket.readyState === WebSocket.CONNECTING)
    ) {
      socket.close();
    }

    set({ socket: null });
  };

  const connectSocket = () => {
    const conversationId = get().conversation?.id;
    if (!conversationId) {
      return;
    }
    closeSocket({ reconnect: false });
    shouldReconnect = true;

    const handleRealtimeEvent = (event: RealtimeEnvelope) => {
      const payload = event.data ?? event.payload;
      const needsRefresh =
        event.type === "message.created" ||
        event.type?.startsWith("conversation.");

      if (needsRefresh && payload?.conversationId === conversationId) {
        void get().refreshMessages().then(() => {
          if (event.type === "message.created") {
            const state = get();
            const lastMessage = state.messages.at(-1);
            if (lastMessage && lastMessage.senderType !== "customer" && typeof document !== "undefined" && document.visibilityState !== "visible") {
              showNotification(
                "新消息",
                getNotificationBody(lastMessage),
                () => {
                  state.setIsOpen(true);
                  state.setIsVisible(true);
                }
              );
            }
          }
        });
      }
    };

    const socket = createRealtimeConnection(handleRealtimeEvent);
    set({ socket });

    socket.addEventListener("open", () => {
      clearRealtimeTimers();
      reconnectAttempt = 0;
      pingTimer = window.setInterval(() => {
        if (socket.readyState === WebSocket.OPEN) {
          socket.send(JSON.stringify({ type: "ping" }));
        }
      }, 20000);
      if (get().isOpen && get().socket === socket) {
        set({ status: "connected" });
      }
    });

    socket.addEventListener("error", () => {
      if (get().socket === socket) {
        scheduleReconnect();
      }
    });

    socket.addEventListener("close", () => {
      if (pingTimer !== null) {
        window.clearInterval(pingTimer);
        pingTimer = null;
      }
      if (get().socket === socket) {
        set({ socket: null });
      }
      if (get().isOpen) {
        if (shouldReconnect) {
          scheduleReconnect();
        } else {
          set({ status: "disconnected" });
        }
      }
    });
  };

  return {
    title: "在线客服",
    welcomeText: "",
    themeColor: "#2563eb",
    conversation: null,
    messages: [],
    status: "connecting",
    error: "",
    isOpen:
      typeof window !== "undefined" ? window.self === window.top : false,
    isVisible:
      typeof window !== "undefined" ? window.self === window.top : false,
    initialized: false,
    socket: null,
    readingMessageId: 0,

    setIsOpen: (isOpen: boolean) => {
      set({ isOpen });
    },

    setIsVisible: (isVisible: boolean) => {
      set({ isVisible });
    },

    disconnectSocket: () => {
      closeSocket({ reconnect: false });
    },

    refreshMessages: async () => {
      const conversationId = get().conversation?.id;
      if (!conversationId) return;

      try {
        const history = await fetchMessages(conversationId);
        const currentConversation = get().conversation;
        set({
          messages: history,
          conversation: currentConversation,
        });
      } catch (e) {
        console.error("Failed to refresh messages", e);
      }
    },

    markConversationRead: async () => {
      const state = get();
      const conversation = state.conversation;
      const lastMessage = state.messages.at(-1);
      if (!conversation?.id || !lastMessage) {
        return;
      }
      if (
        (conversation.customerUnreadCount ?? 0) <= 0 &&
        (conversation.customerLastReadMessageId ?? 0) >= lastMessage.id
      ) {
        return;
      }
      if (state.readingMessageId === lastMessage.id) {
        return;
      }

      set({ readingMessageId: lastMessage.id });
      try {
        await markMessageRead(conversation.id, lastMessage.id);
        set((current) => ({
          readingMessageId: 0,
          messages: current.messages.map((item) =>
            (item.seqNo ?? 0) <= (lastMessage.seqNo ?? 0)
              ? { ...item, customerRead: true }
              : item,
          ),
          conversation: current.conversation
            ? {
                ...current.conversation,
                customerUnreadCount: 0,
                customerLastReadMessageId: lastMessage.id,
                customerLastReadSeqNo: lastMessage.seqNo,
              }
            : null,
        }));
      } catch (error) {
        set({ readingMessageId: 0 });
        throw error;
      }
    },

    bootstrap: () => {
      const token = ++bootstrapToken;

      if (!get().isOpen) {
        closeSocket({ reconnect: false });
        set({ status: "disconnected" });
        return;
      }

      const activateChat = async () => {
        try {
          set({ error: "", status: "connecting" });

          const widgetConfig: WidgetConfigResponse =
            await fetchWidgetConfig().catch(() => ({}));
          if (bootstrapToken !== token || !get().isOpen) return;

          set({
            title: widgetConfig.title || "在线客服",
            welcomeText: widgetConfig.welcomeText || "",
            themeColor: widgetConfig.themeColor || "#2563eb",
          });

          let currentConversation = get().conversation;
          if (!get().initialized || !currentConversation) {
            const widgetConfig = readWidgetConfig();
            currentConversation = await createOrMatchConversation();
            if (bootstrapToken !== token || !get().isOpen) return;
            set({ initialized: true, conversation: currentConversation });
          }

          await get().refreshMessages();
          if (bootstrapToken !== token || !get().isOpen) return;

          connectSocket();
        } catch (bootstrapError) {
          if (bootstrapToken !== token || !get().isOpen) return;
          set({
            status: "disconnected",
            error:
              bootstrapError instanceof Error
                ? bootstrapError.message
                : "初始化失败",
          });
        }
      };

      void activateChat();
    },

    handleSendMessage: async (content: string) => {
      const conversationId = get().conversation?.id;
      if (!conversationId) return;
      set({ error: "" });
      try {
        const nextMessage = await sendMessage(conversationId, content);
        set((state) => ({
          messages: [...state.messages, nextMessage],
          conversation: state.conversation
            ? {
                ...state.conversation,
                customerLastReadMessageId: nextMessage.id,
                customerLastReadSeqNo: nextMessage.seqNo,
                customerUnreadCount: 0,
              }
            : null,
        }));
      } catch (e) {
        set({ error: e instanceof Error ? e.message : "发送消息失败" });
      }
    },

    uploadMessageImage: async (file: File) => {
      const conversationId = get().conversation?.id;
      if (!conversationId) return null;
      set({ error: "" });
      try {
        const asset = await uploadImage(conversationId, file);
        return { url: asset.url, filename: asset.filename };
      } catch (e) {
        set({ error: e instanceof Error ? e.message : "发送图片失败" });
        return null;
      }
    },

    retry: async () => {
      if (!get().conversation?.id) return;

      set({ error: "", status: "connecting" });
      try {
        await get().refreshMessages();
        if (get().isOpen) {
          shouldReconnect = true;
          connectSocket();
        }
      } catch (retryError) {
        set({
          status: "disconnected",
          error:
            retryError instanceof Error ? retryError.message : "刷新失败",
        });
      }
    },
  };
});
