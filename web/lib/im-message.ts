export type MessageAssetPayload = {
  assetId: string
  filename?: string
  fileSize?: number
  mimeType?: string
  url?: string
}

export function parseMessageAssetPayload(payload?: string): MessageAssetPayload | null {
  if (!payload?.trim()) {
    return null
  }
  try {
    const parsed = JSON.parse(payload) as MessageAssetPayload
    if (!parsed?.assetId?.trim()) {
      return null
    }
    return parsed
  } catch {
    return null
  }
}

export function renderIMMessageHTML(message: {
  messageType: string
  content: string
  payload?: string
}) {
  if (message.messageType === "html") {
    return message.content
  }

  const asset = parseMessageAssetPayload(message.payload)
  if (message.messageType === "image") {
    if (asset?.url) {
      return `<p><img src="${escapeHTMLAttr(asset.url)}" alt="${escapeHTMLAttr(
        asset.filename || "image"
      )}"></p>`
    }
    return "<p>[图片]</p>"
  }

  if (message.messageType === "attachment") {
    if (asset?.url) {
      const title = escapeHTML(asset.filename || message.content || "附件")
      const meta = [
        asset.mimeType?.trim() || "",
        formatFileSize(asset.fileSize ?? 0),
      ]
        .filter(Boolean)
        .join(" · ")
      const metaHTML = meta ? `<div class="im-attachment-meta">${escapeHTML(meta)}</div>` : ""
      return `<div class="im-attachment"><a href="${escapeHTMLAttr(
        asset.url
      )}" target="_blank" rel="noreferrer" download="${escapeHTMLAttr(
        asset.filename || ""
      )}">${title}</a>${metaHTML}</div>`
    }
    return `<p>${escapeHTML(message.content || "[附件]")}</p>`
  }

  return `<p>${escapeHTML(message.content || "")}</p>`
}

export function summarizeIMMessage(message: {
  messageType: string
  content: string
  payload?: string
}) {
  if (message.messageType === "image") {
    return "[图片]"
  }
  if (message.messageType === "attachment") {
    const asset = parseMessageAssetPayload(message.payload)
    return asset?.filename?.trim() ? `[附件] ${asset.filename.trim()}` : "[附件]"
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

export function formatFileSize(size: number) {
  if (!Number.isFinite(size) || size <= 0) {
    return ""
  }
  const units = ["B", "KB", "MB", "GB"]
  let value = size
  let index = 0
  while (value >= 1024 && index < units.length - 1) {
    value /= 1024
    index += 1
  }
  const digits = value >= 10 || index === 0 ? 0 : 1
  return `${value.toFixed(digits)} ${units[index]}`
}

function extractTextFromHTML(html: string): string {
  if (typeof document === "undefined") {
    return ""
  }
  const div = document.createElement("div")
  div.innerHTML = html
  return div.textContent || div.innerText || ""
}

function escapeHTML(value: string) {
  return value
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#39;")
    .replaceAll("\n", "<br>")
}

function escapeHTMLAttr(value: string) {
  return value
    .replaceAll("&", "&amp;")
    .replaceAll('"', "&quot;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
}
