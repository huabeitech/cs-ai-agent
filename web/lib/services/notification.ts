import { summarizeIMMessage } from "@/lib/im-message"

type NotificationMessage = {
  messageType: string
  content: string
  payload?: string
}

export function getNotificationBody(message: NotificationMessage): string {
  return summarizeIMMessage(message)
}

export function showNotification(title: string, body: string, onClick?: () => void) {
  if (typeof Notification === "undefined") {
    return
  }

  if (Notification.permission === "granted") {
    const notification = new Notification(title, {
      body,
      icon: "/favicon.ico",
      badge: "/favicon.ico",
    })

    if (onClick) {
      notification.onclick = () => {
        onClick()
        notification.close()
      }
    }

    setTimeout(() => {
      notification.close()
    }, 5000)
  } else if (Notification.permission === "default") {
    Notification.requestPermission().then((permission) => {
      if (permission === "granted") {
        showNotification(title, body, onClick)
      }
    })
  }
}
