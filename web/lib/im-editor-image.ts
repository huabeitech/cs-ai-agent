import type { Editor } from "@tiptap/react"

export type UploadedEditorImage = {
  assetId: string
  provider: string
  storageKey: string
  filename?: string
}

export type UploadedEditorImageMap = Map<string, UploadedEditorImage>

export function removeEditorImageByTitle(editor: Editor, title: string) {
  const { state } = editor
  let targetPos: number | null = null
  state.doc.descendants((node, pos) => {
    if (node.type.name === "image" && node.attrs.title === title) {
      targetPos = pos
      return false
    }
    return true
  })
  if (targetPos === null) {
    return
  }
  editor.chain().focus().deleteRange({ from: targetPos, to: targetPos + 1 }).run()
}

export function markEditorImageUploadedByTitle(
  editor: Editor,
  title: string,
  uploaded: UploadedEditorImage,
  uploadedImages: UploadedEditorImageMap
) {
  uploadedImages.set(title, uploaded)
  const image = findEditorImageElementByTitle(editor, title)
  if (!image) {
    return
  }
  image.setAttribute("data-asset-id", uploaded.assetId)
  image.setAttribute("data-provider", uploaded.provider)
  image.setAttribute("data-storage-key", uploaded.storageKey)
  image.setAttribute("alt", uploaded.filename || image.getAttribute("alt") || "image")
  image.classList.remove("cs-agent-editor-image-uploading")
  image.removeAttribute("data-uploading")
  image.removeAttribute("title")
}

export function setEditorImageUploadingByTitle(editor: Editor, title: string) {
  const image = findEditorImageElementByTitle(editor, title)
  if (!image) {
    return
  }
  image.classList.add("cs-agent-editor-image-uploading")
  image.setAttribute("data-uploading", "true")
}

export function buildSendableEditorHTML(
  html: string,
  uploadedImages?: UploadedEditorImageMap
) {
  if (typeof document === "undefined" || !html.includes("<img")) {
    return html
  }

  const template = document.createElement("template")
  template.innerHTML = html
  for (const image of Array.from(template.content.querySelectorAll("img"))) {
    const title = image.getAttribute("title") ?? ""
    const uploaded = title ? uploadedImages?.get(title) : undefined
    if (uploaded) {
      image.setAttribute("data-asset-id", uploaded.assetId)
      image.setAttribute("data-provider", uploaded.provider)
      image.setAttribute("data-storage-key", uploaded.storageKey)
      image.setAttribute("alt", uploaded.filename || image.getAttribute("alt") || "image")
      image.removeAttribute("title")
    }
    if (
      image.getAttribute("data-asset-id") &&
      image.getAttribute("src")?.startsWith("blob:")
    ) {
      image.removeAttribute("src")
    }
  }
  return template.innerHTML
}

export function hasUploadingEditorImages(
  html: string,
  uploadedImages?: UploadedEditorImageMap
) {
  if (typeof document === "undefined" || !html.includes("<img")) {
    return /<img\b[^>]*\btitle=(["'])uploading-[^"']+\1/i.test(html)
  }

  const template = document.createElement("template")
  template.innerHTML = html
  return Array.from(template.content.querySelectorAll("img")).some((image) =>
    isUnfinishedUploadingImage(image, uploadedImages)
  )
}

export function revokeEditorObjectUrl(urls: Set<string>, objectUrl: string) {
  if (!urls.delete(objectUrl)) {
    return
  }
  URL.revokeObjectURL(objectUrl)
}

export function revokeEditorObjectUrls(urls: Set<string>) {
  for (const objectUrl of urls) {
    URL.revokeObjectURL(objectUrl)
  }
  urls.clear()
}

function findEditorImageElementByTitle(editor: Editor, title: string) {
  const escapedTitle =
    typeof CSS !== "undefined" && typeof CSS.escape === "function"
      ? CSS.escape(title)
      : title.replace(/["\\]/g, "\\$&")
  return editor.view.dom.querySelector<HTMLImageElement>(
    `img[title="${escapedTitle}"]`
  )
}

function isUnfinishedUploadingImage(
  image: HTMLImageElement,
  uploadedImages?: UploadedEditorImageMap
) {
  const title = image.getAttribute("title") ?? ""
  return title.startsWith("uploading-") && !uploadedImages?.has(title)
}
