"use client"

import { useEffect, useRef, useState } from "react"
import { EditorContent, useEditor } from "@tiptap/react"
import Image from "@tiptap/extension-image"
import Placeholder from "@tiptap/extension-placeholder"
import StarterKit from "@tiptap/starter-kit"
import { ImageIcon, PaperclipIcon, SendHorizonalIcon } from "lucide-react"

import { Button } from "@/components/ui/button"
import { generateUUID } from "@/lib/utils"

type UploadedImage = {
  assetId: string
  provider: string
  storageKey: string
  url: string
  filename?: string
}

const MessageImage = Image.extend({
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
})

type KefuMessageEditorProps = {
  disabled?: boolean
  uploadingAsset?: boolean
  onSend: (html: string) => Promise<void>
  onUploadImage: (file: File) => Promise<UploadedImage | null>
  onSendAttachment: (file: File) => Promise<void>
}

export function KefuMessageEditor({
  disabled = false,
  uploadingAsset = false,
  onSend,
  onUploadImage,
  onSendAttachment,
}: KefuMessageEditorProps) {
  const [localUploading, setLocalUploading] = useState(false)
  const imageInputRef = useRef<HTMLInputElement | null>(null)
  const attachmentInputRef = useRef<HTMLInputElement | null>(null)
  const onSendRef = useRef(onSend)
  const onUploadImageRef = useRef(onUploadImage)
  const onSendAttachmentRef = useRef(onSendAttachment)
  const shouldRestoreFocusRef = useRef(false)
  const isUploading = uploadingAsset || localUploading

  useEffect(() => {
    onSendRef.current = onSend
  }, [onSend])

  useEffect(() => {
    onUploadImageRef.current = onUploadImage
  }, [onUploadImage])

  useEffect(() => {
    onSendAttachmentRef.current = onSendAttachment
  }, [onSendAttachment])

  const editor = useEditor({
    immediatelyRender: false,
    extensions: [
      StarterKit.configure({
        heading: false,
        blockquote: false,
        codeBlock: false,
        bulletList: false,
        orderedList: false,
        horizontalRule: false,
      }),
      MessageImage,
      Placeholder.configure({
        placeholder: "输入消息，Enter 发送，Shift + Enter 换行",
      }),
    ],
    content: "",
    editorProps: {
      attributes: {
        class:
          "min-h-12 max-h-40 overflow-y-auto px-1.5 py-1 text-sm leading-6 text-slate-900 outline-none [&_p]:m-0 [&_p+*]:mt-2 [&_img]:my-2 [&_img]:max-h-64 [&_img]:rounded-lg [&_img]:object-contain",
      },
      handleKeyDown: (_view, event) => {
        if (event.key === "Enter" && !event.shiftKey) {
          event.preventDefault()
          void handleSend()
          return true
        }
        return false
      },
      handlePaste: (_view, event) => {
        if (disabled || isUploading) {
          return false
        }
        const imageFile = getClipboardImageFile(event.clipboardData)
        if (!imageFile) {
          return false
        }
        event.preventDefault()
        void insertUploadedImage(imageFile)
        return true
      },
    },
  })

  useEffect(() => {
    if (!editor) {
      return
    }
    editor.setEditable(!disabled && !isUploading)
  }, [disabled, editor, isUploading])

  useEffect(() => {
    if (!editor || disabled || isUploading || !shouldRestoreFocusRef.current) {
      return
    }
    requestAnimationFrame(() => {
      editor.commands.focus()
    })
  }, [disabled, editor, isUploading])

  async function handleSend() {
    if (!editor || disabled || isUploading) {
      return
    }
    const html = editor.getHTML()
    if (!isMeaningfulHTML(html)) {
      return
    }
    await onSendRef.current(html)
    editor.commands.clearContent(true)
  }

  async function handleSelectImage(event: React.ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0]
    event.target.value = ""
    if (!file || !editor || disabled || isUploading) {
      if (editor && shouldRestoreFocusRef.current) {
        requestAnimationFrame(() => {
          editor.commands.focus()
        })
      }
      return
    }
    await insertUploadedImage(file)
  }

  async function insertUploadedImage(file: File) {
    if (!editor || disabled || isUploading) {
      return
    }

    shouldRestoreFocusRef.current = true
    const objectUrl = URL.createObjectURL(file)
    const placeholderId = `uploading-${generateUUID()}`
    editor
      .chain()
      .focus()
      .setImage({
        src: objectUrl,
        alt: file.name || "uploading-image",
        title: placeholderId,
      })
      .run()

    try {
      setLocalUploading(true)
      const uploaded = await onUploadImageRef.current(file)
      if (!uploaded?.url) {
        removeImageByTitle(editor, placeholderId)
        return
      }
      replaceImageSourceByTitle(editor, placeholderId, uploaded)
    } finally {
      setLocalUploading(false)
      URL.revokeObjectURL(objectUrl)
      requestAnimationFrame(() => {
        if (!disabled && shouldRestoreFocusRef.current) {
          editor.commands.focus()
        }
      })
    }
  }

  async function handleSelectAttachment(
    event: React.ChangeEvent<HTMLInputElement>
  ) {
    const file = event.target.files?.[0]
    event.target.value = ""
    if (!file || disabled || isUploading) {
      if (editor && shouldRestoreFocusRef.current) {
        requestAnimationFrame(() => {
          editor.commands.focus()
        })
      }
      return
    }

    shouldRestoreFocusRef.current = editor?.isFocused ?? true
    setLocalUploading(true)
    try {
      await onSendAttachmentRef.current(file)
    } finally {
      setLocalUploading(false)
      requestAnimationFrame(() => {
        if (editor && !disabled && shouldRestoreFocusRef.current) {
          editor.commands.focus()
        }
      })
    }
  }

  return (
    <div className="border-t border-slate-200/70 bg-white px-3 pb-3 pt-2">
      <div className="rounded-xl border border-slate-200 bg-white p-2 shadow-[0_8px_24px_rgba(15,23,42,0.05)]">
        <input
          ref={imageInputRef}
          type="file"
          accept="image/*"
          className="hidden"
          onChange={handleSelectImage}
        />
        <input
          ref={attachmentInputRef}
          type="file"
          className="hidden"
          onChange={handleSelectAttachment}
        />
        <div className="min-h-10">
          <EditorContent editor={editor} />
        </div>
        <div className="mt-2 flex items-center justify-between">
          <div className="flex items-center gap-1.5">
            <Button
              type="button"
              variant="ghost"
              size="icon"
              onMouseDown={(event) => event.preventDefault()}
              onClick={() => {
                shouldRestoreFocusRef.current = editor?.isFocused ?? true
                imageInputRef.current?.click()
              }}
              disabled={disabled || isUploading}
              aria-label={isUploading ? "图片上传中" : "发送图片"}
              title={isUploading ? "图片上传中" : "发送图片"}
              className="text-slate-500 hover:bg-slate-100 hover:text-slate-800"
            >
              <ImageIcon />
            </Button>
            <Button
              type="button"
              variant="ghost"
              size="icon"
              onMouseDown={(event) => event.preventDefault()}
              onClick={() => {
                shouldRestoreFocusRef.current = editor?.isFocused ?? true
                attachmentInputRef.current?.click()
              }}
              disabled={disabled || isUploading}
              aria-label={isUploading ? "附件上传中" : "发送附件"}
              title={isUploading ? "附件上传中" : "发送附件"}
              className="text-slate-500 hover:bg-slate-100 hover:text-slate-800"
            >
              <PaperclipIcon />
            </Button>
          </div>
          <div className="flex items-center gap-2">
            <p className="text-[10px] text-slate-400">Enter 发送</p>
            <Button
              type="button"
              size="icon"
              onClick={() => void handleSend()}
              disabled={disabled || isUploading}
              aria-label="发送"
              title="发送"
              className="bg-[var(--primary)] text-white shadow-[0_10px_20px_color-mix(in_srgb,var(--primary)_24%,transparent)] hover:bg-[var(--primary)] hover:brightness-105"
            >
              <SendHorizonalIcon />
            </Button>
          </div>
        </div>
      </div>
    </div>
  )
}

function isMeaningfulHTML(html: string) {
  const normalized = html
    .replace(/<p><\/p>/g, "")
    .replace(/<p><br><\/p>/g, "")
    .replace(/\s+/g, "")
  if (/<img[\s\S]*?>/i.test(normalized)) {
    return true
  }
  const plainText = normalized.replace(/<[^>]+>/g, "").trim()
  return plainText !== ""
}

function getClipboardImageFile(data: DataTransfer | null) {
  if (!data) {
    return null
  }

  for (const item of Array.from(data.items)) {
    if (item.kind === "file" && item.type.startsWith("image/")) {
      return item.getAsFile()
    }
  }
  return null
}

function removeImageByTitle(editor: NonNullable<ReturnType<typeof useEditor>>, title: string) {
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

function replaceImageSourceByTitle(
  editor: NonNullable<ReturnType<typeof useEditor>>,
  title: string,
  uploaded: UploadedImage
) {
  const { state, view } = editor
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
  const transaction = view.state.tr.setNodeMarkup(targetPos, undefined, {
    ...view.state.doc.nodeAt(targetPos)?.attrs,
    src: uploaded.url,
    alt: uploaded.filename || "image",
    dataAssetId: uploaded.assetId,
    dataProvider: uploaded.provider,
    dataStorageKey: uploaded.storageKey,
    title: "",
  })
  view.dispatch(transaction)
}
