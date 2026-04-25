"use client"

import { useEffect, useRef, useState } from "react"
import { EditorContent, useEditor } from "@tiptap/react"
import Placeholder from "@tiptap/extension-placeholder"
import StarterKit from "@tiptap/starter-kit"
import { ImageIcon, PaperclipIcon, SendHorizonalIcon } from "lucide-react"

import { Button } from "@/components/ui/button"
import {
  buildSendableEditorHTML,
  hasUploadingEditorImages,
  markEditorImageUploadedByTitle,
  MessageImageExtension,
  removeEditorImageByTitle,
  revokeEditorObjectUrl,
  revokeEditorObjectUrls,
  setEditorImageUploadingByTitle,
} from "@/lib/im-editor-image"
import { generateUUID } from "@/lib/utils"

type UploadedImage = {
  assetId: string
  provider: string
  storageKey: string
  url: string
  filename?: string
}

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
  const objectUrlsRef = useRef<Set<string>>(new Set())
  const uploadedImagesRef = useRef(new Map<string, UploadedImage>())
  const isUploading = uploadingAsset || localUploading

  useEffect(() => {
    const objectUrls = objectUrlsRef.current
    return () => {
      revokeEditorObjectUrls(objectUrls)
    }
  }, [])

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
      MessageImageExtension,
      Placeholder.configure({
        placeholder: "输入消息，Enter 发送，Shift + Enter 换行",
      }),
    ],
    content: "",
    editorProps: {
      attributes: {
        class:
          "cs-agent-scrollbar min-h-12 max-h-40 overflow-y-auto px-1.5 py-1 text-sm leading-6 text-foreground outline-none [&_p]:m-0 [&_p+*]:mt-2 [&_.cs-agent-editor-image-wrap]:my-2 [&_.cs-agent-editor-image]:max-h-64 [&_.cs-agent-editor-image]:max-w-full [&_.cs-agent-editor-image]:rounded-lg [&_.cs-agent-editor-image]:object-contain [&_.cs-agent-editor-image-wrap-uploading_.cs-agent-editor-image]:opacity-55",
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
    const rawHTML = editor.getHTML()
    if (hasUploadingEditorImages(rawHTML, uploadedImagesRef.current)) {
      return
    }
    const html = buildSendableEditorHTML(rawHTML, uploadedImagesRef.current)
    if (!isMeaningfulHTML(html)) {
      return
    }
    await onSendRef.current(html)
    editor.commands.clearContent(true)
    revokeEditorObjectUrls(objectUrlsRef.current)
    uploadedImagesRef.current.clear()
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
    objectUrlsRef.current.add(objectUrl)
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
    setEditorImageUploadingByTitle(editor, placeholderId)

    try {
      setLocalUploading(true)
      const uploaded = await onUploadImageRef.current(file)
      if (!uploaded?.assetId || !uploaded.provider || !uploaded.storageKey) {
        removeEditorImageByTitle(editor, placeholderId)
        revokeEditorObjectUrl(objectUrlsRef.current, objectUrl)
        return
      }
      markEditorImageUploadedByTitle(
        editor,
        placeholderId,
        uploaded,
        uploadedImagesRef.current
      )
    } finally {
      setLocalUploading(false)
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
    // <div className="border-t border-border bg-card px-3 pb-3 pt-2">
    <div className="p-3">
      <div className="rounded-xl border border-border bg-background p-2 shadow-[0_8px_24px_rgba(15,23,42,0.05)] dark:shadow-none">
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
              className="text-muted-foreground hover:bg-muted hover:text-foreground"
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
              className="text-muted-foreground hover:bg-muted hover:text-foreground"
            >
              <PaperclipIcon />
            </Button>
          </div>
          <div className="flex items-center gap-2">
            <p className="text-[10px] text-muted-foreground">Enter 发送</p>
            <Button
              type="button"
              size="icon"
              onClick={() => void handleSend()}
              disabled={disabled || isUploading}
              aria-label="发送"
              title="发送"
              className="bg-primary text-white shadow-[0_10px_20px_color-mix(in_srgb,var(--primary)_24%,transparent)] hover:bg-primary hover:brightness-105"
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
