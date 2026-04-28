import { readSession } from "@/lib/auth"
import { request } from "@/lib/api/client"
import { createWebSocketBaseUrl } from "@/lib/api/websocket"
import type { PageResult } from "@/lib/api/admin"

export type NotificationReadStatus = "all" | "unread" | "read"

export type NotificationItem = {
  id: number
  recipientUserId: number
  title: string
  content: string
  notificationType: string
  bizType: string
  bizId: number
  actionUrl: string
  readAt?: string
  createdAt?: string
}

export type NotificationUnreadCount = {
  unreadCount: number
}

export type NotificationListQuery = {
  page?: number
  limit?: number
  readStatus?: NotificationReadStatus
  type?: string
}

function toQueryString(query?: Record<string, string | number | undefined>) {
  if (!query) {
    return ""
  }

  const params = new URLSearchParams()
  Object.entries(query).forEach(([key, value]) => {
    if (value === undefined || value === "") {
      return
    }
    params.set(key, String(value))
  })
  const output = params.toString()
  return output ? `?${output}` : ""
}

export function fetchNotifications(query?: NotificationListQuery) {
  return request<PageResult<NotificationItem>>(
    `/api/dashboard/notification/list${toQueryString(query)}`
  )
}

export function fetchNotificationUnreadCount() {
  return request<NotificationUnreadCount>("/api/dashboard/notification/unread_count")
}

export function markNotificationRead(id: number) {
  return request<void>("/api/dashboard/notification/mark_read", {
    method: "POST",
    body: JSON.stringify({ id }),
  })
}

export function markAllNotificationsRead() {
  return request<void>("/api/dashboard/notification/mark_all_read", {
    method: "POST",
  })
}

export function createNotificationWebSocketUrl() {
  const session = readSession()
  if (!session?.accessToken) {
    throw new Error("未登录或登录已过期")
  }

  const params = new URLSearchParams({
    accessToken: session.accessToken,
  })
  return `${createWebSocketBaseUrl()}/api/ws/dashboard/notification?${params.toString()}`
}
