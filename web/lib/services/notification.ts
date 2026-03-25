type NotificationMessage = {
  messageType: string
  content: string
}

export function getNotificationBody(message: NotificationMessage): string {
  if (message.messageType === "image") {
    return "[图片]"
  }
  if (message.messageType === "html") {
    const text = extractTextFromHTML(message.content)
    if (text.trim()) {
      return text.substring(0, 100)
    }
    if (message.content.includes("<img")) {
      return "[图片]"
    }
    return "[消息]"
  }
  return message.content?.substring(0, 100) || "[消息]"
}

function extractTextFromHTML(html: string): string {
  if (typeof document === "undefined") {
    return ""
  }
  const div = document.createElement("div")
  div.innerHTML = html
  return div.textContent || div.innerText || ""
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
