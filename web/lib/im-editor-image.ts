import type { Editor } from "@tiptap/react"
import Image from "@tiptap/extension-image"

export const MessageImageExtension = Image.extend({
  addAttributes() {
    return {
      ...this.parent?.(),
      dataAssetId: {
        default: null,
        parseHTML: (element) => element.getAttribute("data-asset-id"),
        renderHTML: (attributes) =>
          attributes.dataAssetId ? { "data-asset-id": attributes.dataAssetId } : {},
      },
      dataProvider: {
        default: null,
        parseHTML: (element) => element.getAttribute("data-provider"),
        renderHTML: (attributes) =>
          attributes.dataProvider ? { "data-provider": attributes.dataProvider } : {},
      },
      dataStorageKey: {
        default: null,
        parseHTML: (element) => element.getAttribute("data-storage-key"),
        renderHTML: (attributes) =>
          attributes.dataStorageKey ? { "data-storage-key": attributes.dataStorageKey } : {},
      },
    }
  },

  addNodeView() {
    return ({ node }) => {
      const wrapper = document.createElement("span")
      wrapper.className =
        "cs-agent-editor-image-wrap relative inline-block max-w-full align-middle"

      const image = document.createElement("img")
      image.className = "cs-agent-editor-image"
      image.draggable = true

      const overlay = document.createElement("span")
      overlay.className =
        "cs-agent-editor-image-loading pointer-events-none absolute inset-0 hidden items-center justify-center rounded-lg bg-background/55 backdrop-blur-[1px]"

      const spinner = document.createElement("span")
      spinner.className =
        "size-6 animate-spin rounded-full border-2 border-primary/25 border-t-primary shadow-sm"

      overlay.appendChild(spinner)
      wrapper.appendChild(image)
      wrapper.appendChild(overlay)

      const applyAttrs = (attrs: Record<string, unknown>) => {
        setImageAttr(image, "src", attrs.src)
        setImageAttr(image, "alt", attrs.alt)
        setImageAttr(image, "title", attrs.title)
        setImageAttr(image, "data-asset-id", attrs.dataAssetId)
        setImageAttr(image, "data-provider", attrs.dataProvider)
        setImageAttr(image, "data-storage-key", attrs.dataStorageKey)
        setImageUploading(wrapper, Boolean(String(attrs.title ?? "").startsWith("uploading-")))
      }

      applyAttrs(node.attrs)

      return {
        dom: wrapper,
        update: (updatedNode) => {
          if (updatedNode.type.name !== node.type.name) {
            return false
          }
          applyAttrs(updatedNode.attrs)
          return true
        },
        ignoreMutation: () => true,
      }
    }
  },
})

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
  setImageUploading(image.closest(".cs-agent-editor-image-wrap"), false)
}

export function setEditorImageUploadingByTitle(editor: Editor, title: string) {
  const image = findEditorImageElementByTitle(editor, title)
  if (!image) {
    return
  }
  image.classList.add("cs-agent-editor-image-uploading")
  image.setAttribute("data-uploading", "true")
  setImageUploading(image.closest(".cs-agent-editor-image-wrap"), true)
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

function setImageAttr(image: HTMLImageElement, name: string, value: unknown) {
  const normalized = typeof value === "string" ? value : ""
  if (!normalized) {
    image.removeAttribute(name)
    return
  }
  if (image.getAttribute(name) !== normalized) {
    image.setAttribute(name, normalized)
  }
}

function setImageUploading(wrapper: Element | null, uploading: boolean) {
  if (!(wrapper instanceof HTMLElement)) {
    return
  }
  wrapper.classList.toggle("cs-agent-editor-image-wrap-uploading", uploading)
  const overlay = wrapper.querySelector<HTMLElement>(".cs-agent-editor-image-loading")
  if (overlay) {
    overlay.classList.toggle("hidden", !uploading)
    overlay.classList.toggle("flex", uploading)
  }
}
