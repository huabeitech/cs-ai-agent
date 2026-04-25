import { parseMessageAssetPayload } from "@/lib/im-message"

export type MergeableImMessage = {
  id: number
  conversationId: number
  clientMsgId?: string
  senderType: string
  senderId: number
  senderName?: string
  senderAvatar?: string
  messageType: string
  content: string
  payload?: string
  seqNo: number
  sendStatus: number
  sentAt?: string
  deliveredAt?: string
  readAt?: string
  customerRead: boolean
  customerReadAt?: string
  agentRead: boolean
  agentReadAt?: string
  recalledAt?: string
  quotedMessageId?: number
}

export function mergeImMessagesByIdAsc<T extends MergeableImMessage>(
  a: T[],
  b: T[]
): T[] {
  const byId = new Map<number, T>()
  for (const message of a) {
    byId.set(message.id, message)
  }
  for (const message of b) {
    const existing = byId.get(message.id)
    byId.set(message.id, existing ? mergeImMessage(existing, message) : message)
  }
  return Array.from(byId.values()).sort((x, y) => x.id - y.id)
}

export function mergeImMessage<T extends MergeableImMessage>(
  existing: T,
  incoming: T
): T {
  const normalizedIncoming = normalizeDynamicImageContent(
    existing,
    normalizeDynamicImagePayload(existing, incoming)
  )
  return isSameImMessage(existing, normalizedIncoming) ? existing : normalizedIncoming
}

function normalizeDynamicImagePayload<T extends MergeableImMessage>(
  existing: T,
  incoming: T
): T {
  if (
    existing.messageType !== "image" ||
    incoming.messageType !== "image" ||
    !existing.payload ||
    !incoming.payload
  ) {
    return incoming
  }

  const existingAsset = parseMessageAssetPayload(existing.payload)
  const incomingAsset = parseMessageAssetPayload(incoming.payload)
  if (
    !existingAsset?.assetId ||
    existingAsset.assetId !== incomingAsset?.assetId ||
    existing.payload === incoming.payload
  ) {
    return incoming
  }

  return {
    ...incoming,
    payload: existing.payload,
  }
}

function normalizeDynamicImageContent<T extends MergeableImMessage>(
  existing: T,
  incoming: T
): T {
  if (
    existing.content === incoming.content ||
    !existing.content.includes("data-asset-id") ||
    !incoming.content.includes("data-asset-id")
  ) {
    return incoming
  }

  if (getStableHTMLContentKey(existing.content) !== getStableHTMLContentKey(incoming.content)) {
    return incoming
  }

  return {
    ...incoming,
    content: existing.content,
  }
}

function getStableHTMLContentKey(html: string): string {
  if (typeof document === "undefined") {
    return html.replace(
      /(<img\b[^>]*\bdata-asset-id=(["'])[^"']+\2[^>]*?)\s+(?:src|srcset)=(["'])[^"']*\3/gi,
      "$1"
    )
  }

  const template = document.createElement("template")
  template.innerHTML = html
  for (const image of Array.from(template.content.querySelectorAll("img"))) {
    if (image.getAttribute("data-asset-id")) {
      image.removeAttribute("src")
      image.removeAttribute("srcset")
    }
  }
  return template.innerHTML
}

function isSameImMessage(a: MergeableImMessage, b: MergeableImMessage): boolean {
  const aKeys = Object.keys(a)
  const bKeys = Object.keys(b)
  if (aKeys.length !== bKeys.length) {
    return false
  }
  return aKeys.every((key) => {
    const field = key as keyof MergeableImMessage
    return Object.prototype.hasOwnProperty.call(b, key) && a[field] === b[field]
  })
}
