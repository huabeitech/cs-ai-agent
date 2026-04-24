"use client"

import { useEffect, useRef } from "react"
import { toast } from "sonner"

import { createAdminWebSocketUrl } from "@/lib/api/admin"
import { readSession } from "@/lib/auth"
import { useAgentConversationsStore } from "@/lib/stores/agent-conversations"
import { getNotificationBody, showNotification } from "@/lib/services/notification"

const RECONNECT_BASE_DELAY = 2000
const RECONNECT_MAX_DELAY = 30000

export function useAgentConversationRealtime() {
  const selectedConversationId = useAgentConversationsStore((state) => state.selectedConversationId)
  const loadConversations = useAgentConversationsStore((state) => state.loadConversations)
  const syncLatestMessages = useAgentConversationsStore(
    (state) => state.syncLatestMessages,
  )
  const setRealtimeStatus = useAgentConversationsStore((state) => state.setRealtimeStatus)
  const websocketRef = useRef<WebSocket | null>(null)
  const reconnectTimerRef = useRef<number | null>(null)
  const pingTimerRef = useRef<number | null>(null)
  const reconnectAttemptRef = useRef(0)
  const subscribedConversationIdRef = useRef<number | null>(null)
  const selectedConversationIdRef = useRef<number | null>(selectedConversationId)
  const currentUserIdRef = useRef<number>(readSession()?.user.id ?? 0)

  useEffect(() => {
    selectedConversationIdRef.current = selectedConversationId
  }, [selectedConversationId])

  useEffect(() => {
    let cancelled = false

    const clearTimers = () => {
      if (reconnectTimerRef.current) {
        window.clearTimeout(reconnectTimerRef.current)
        reconnectTimerRef.current = null
      }
      if (pingTimerRef.current) {
        window.clearInterval(pingTimerRef.current)
        pingTimerRef.current = null
      }
    }

    const scheduleReconnect = () => {
      if (cancelled || reconnectTimerRef.current) {
        return
      }
      const delay = Math.min(
        RECONNECT_BASE_DELAY * 2 ** reconnectAttemptRef.current,
        RECONNECT_MAX_DELAY
      )
      reconnectTimerRef.current = window.setTimeout(() => {
        reconnectTimerRef.current = null
        reconnectAttemptRef.current += 1
        if (!cancelled) {
          connect()
        }
      }, delay)
    }

    const connect = () => {
      if (cancelled) {
        return
      }

      let socket: WebSocket
      try {
        setRealtimeStatus("connecting")
        socket = new WebSocket(createAdminWebSocketUrl())
      } catch (error) {
        setRealtimeStatus("disconnected")
        toast.error(error instanceof Error ? error.message : "连接实时服务失败")
        scheduleReconnect()
        return
      }

      websocketRef.current = socket

      socket.onopen = () => {
        console.info("[agent-realtime] websocket connected", {
          url: socket.url,
        })
        setRealtimeStatus("connected")
        reconnectAttemptRef.current = 0
        if (pingTimerRef.current) {
          window.clearInterval(pingTimerRef.current)
        }
        pingTimerRef.current = window.setInterval(() => {
          if (socket.readyState === WebSocket.OPEN) {
            socket.send(JSON.stringify({ type: "ping" }))
          }
        }, 20000)

        const conversationId = selectedConversationIdRef.current
        if (conversationId) {
          socket.send(
            JSON.stringify({
              type: "subscribe",
              topics: [`conversation:${conversationId}`],
            })
          )
          subscribedConversationIdRef.current = conversationId
        } else {
          subscribedConversationIdRef.current = null
        }
      }

      socket.onmessage = (event) => {
        try {
          const payload = JSON.parse(event.data) as {
            eventId?: string
            type?: string
            data?: {
              conversationId?: number
              messageId?: number
              status?: number
              currentAssigneeId?: number
              senderType?: string
              messageType?: string
              content?: string
            }
          }
          const eventType = payload.type ?? ""
          const conversationId = payload.data?.conversationId ?? 0
          const eventId = payload.eventId?.trim() ?? ""

          if (
            eventType === "" ||
            eventType === "connected" ||
            eventType === "pong" ||
            eventType === "subscribed" ||
            eventType === "unsubscribed"
          ) {
            return
          }

          if (eventId && socket.readyState === WebSocket.OPEN) {
            socket.send(JSON.stringify({ type: "ack", eventId }))
          }

          if (eventType === "message.created" && conversationId > 0) {
            const senderType = payload.data?.senderType ?? ""
            const status = payload.data?.status ?? 0
            const currentAssigneeId = payload.data?.currentAssigneeId ?? 0

            void loadConversations().then(() => {
              const store = useAgentConversationsStore.getState()
              const shouldNotify =
                senderType === "customer" &&
                status === 2 &&
                currentAssigneeId > 0 &&
                currentAssigneeId === currentUserIdRef.current &&
                typeof document !== "undefined" &&
                document.visibilityState !== "visible"

              if (!shouldNotify) {
                return
              }

              showNotification(
                "新消息",
                getNotificationBody({
                  messageType: payload.data?.messageType ?? "",
                  content: payload.data?.content ?? "",
                }),
                () => {
                  void store.selectConversation(conversationId)
                }
              )
            }).catch((error) => {
              toast.error(error instanceof Error ? error.message : "加载消息失败")
            })
          } else {
            void loadConversations().catch((error) => {
              toast.error(error instanceof Error ? error.message : "加载会话列表失败")
            })
          }

          if (conversationId > 0 && selectedConversationIdRef.current === conversationId) {
            void syncLatestMessages(conversationId)
          }
        } catch {
          // ignore invalid ws payload
        }
      }

      socket.onclose = (event) => {
        console.log("[agent-realtime] websocket closed", {
          url: socket.url,
          readyState: socket.readyState,
          code: event.code,
          reason: event.reason,
          wasClean: event.wasClean,
        })
        setRealtimeStatus("disconnected")
        if (pingTimerRef.current) {
          window.clearInterval(pingTimerRef.current)
          pingTimerRef.current = null
        }
        if (websocketRef.current === socket) {
          websocketRef.current = null
        }
        subscribedConversationIdRef.current = null
        scheduleReconnect()
      }

      socket.onerror = () => {
        console.log("[agent-realtime] websocket error", {
          url: socket.url,
          readyState: socket.readyState,
        })
        setRealtimeStatus("disconnected")
        scheduleReconnect()
      }
    }

    connect()

    return () => {
      cancelled = true
      clearTimers()
      reconnectAttemptRef.current = 0
      const socket = websocketRef.current
      websocketRef.current = null
      if (socket) {
        socket.close()
      }
      setRealtimeStatus("disconnected")
      subscribedConversationIdRef.current = null
    }
  }, [loadConversations, setRealtimeStatus, syncLatestMessages])

  useEffect(() => {
    const socket = websocketRef.current
    if (!socket || socket.readyState !== WebSocket.OPEN) {
      return
    }

    const previousConversationId = subscribedConversationIdRef.current
    const nextConversationId = selectedConversationId ?? null

    if (previousConversationId && previousConversationId !== nextConversationId) {
      socket.send(
        JSON.stringify({
          type: "unsubscribe",
          topics: [`conversation:${previousConversationId}`],
        })
      )
    }

    if (nextConversationId && nextConversationId !== previousConversationId) {
      socket.send(
        JSON.stringify({
          type: "subscribe",
          topics: [`conversation:${nextConversationId}`],
        })
      )
      subscribedConversationIdRef.current = nextConversationId
      return
    }

    if (!nextConversationId) {
      subscribedConversationIdRef.current = null
    }
  }, [selectedConversationId])
}
