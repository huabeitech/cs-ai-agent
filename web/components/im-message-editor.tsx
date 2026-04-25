"use client"

import { useEffect, useRef, useState } from "react"
import { EditorContent, useEditor } from "@tiptap/react"
import StarterKit from "@tiptap/starter-kit"
import Image from "@tiptap/extension-image"
import Placeholder from "@tiptap/extension-placeholder"
import { ImageIcon, MessageSquareTextIcon, PaperclipIcon, SendIcon } from "lucide-react"
import { toast } from "sonner"

import { Button } from "@/components/ui/button"
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { fetchQuickReplyListAll, type AdminQuickReply } from "@/lib/api/admin"
import {
  buildSendableEditorHTML,
  hasUploadingEditorImages,
  markEditorImageUploadedByTitle,
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

type ImMessageEditorProps = {
  disabled?: boolean
  uploadingAsset?: boolean
  onSend: (html: string) => Promise<void>
  onUploadImage: (file: File) => Promise<UploadedImage | null>
  onSendAttachment: (file: File) => Promise<void>
}

export function ImMessageEditor({
  disabled = false,
  uploadingAsset = false,
  onSend,
  onUploadImage,
  onSendAttachment,
}: ImMessageEditorProps) {
  const imageInputRef = useRef<HTMLInputElement>(null)
  const attachmentInputRef = useRef<HTMLInputElement>(null)
  const onSendRef = useRef(onSend)
  const onUploadImageRef = useRef(onUploadImage)
  const onSendAttachmentRef = useRef(onSendAttachment)
  const shouldRestoreFocusRef = useRef(false)
  const objectUrlsRef = useRef<Set<string>>(new Set())
  const uploadedImagesRef = useRef(new Map<string, UploadedImage>())
  const [quickReplies, setQuickReplies] = useState<AdminQuickReply[]>([])
  const [loadingQuickReplies, setLoadingQuickReplies] = useState(false)
  const [quickReplyPickerOpen, setQuickReplyPickerOpen] = useState(false)

  useEffect(() => {
    onSendRef.current = onSend
  }, [onSend])

  useEffect(() => {
    onUploadImageRef.current = onUploadImage
  }, [onUploadImage])

  useEffect(() => {
    onSendAttachmentRef.current = onSendAttachment
  }, [onSendAttachment])

  useEffect(() => {
    const objectUrls = objectUrlsRef.current
    return () => {
      revokeEditorObjectUrls(objectUrls)
    }
  }, [])

  useEffect(() => {
    let cancelled = false
    setLoadingQuickReplies(true)
    void fetchQuickReplyListAll()
      .then((list) => {
        if (!cancelled) {
          setQuickReplies(list)
        }
      })
      .catch((error) => {
        if (!cancelled) {
          toast.error(error instanceof Error ? error.message : "加载快捷回复失败")
        }
      })
      .finally(() => {
        if (!cancelled) {
          setLoadingQuickReplies(false)
        }
      })
    return () => {
      cancelled = true
    }
  }, [])

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
          "h-full min-h-12 max-h-[20vh] overflow-y-auto px-1.5 py-1 text-sm leading-6 text-foreground outline-none sm:max-h-none [&_.ProseMirror-focused]:outline-none [&_p]:m-0 [&_p+img]:mt-2 [&_img]:my-2 [&_img]:max-h-64 [&_img]:rounded-md [&_img]:object-contain [&_img.cs-agent-editor-image-uploading]:animate-pulse [&_img.cs-agent-editor-image-uploading]:opacity-55 [&_img.cs-agent-editor-image-uploading]:ring-2 [&_img.cs-agent-editor-image-uploading]:ring-primary/35 [&_p.is-editor-empty:first-child]:before:text-muted-foreground",
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
        if (disabled || uploadingAsset) {
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
    editor.setEditable(!disabled && !uploadingAsset)
  }, [disabled, editor, uploadingAsset])

  useEffect(() => {
    if (!editor || disabled || uploadingAsset || !shouldRestoreFocusRef.current) {
      return
    }
    requestAnimationFrame(() => {
      editor.commands.focus()
    })
  }, [disabled, editor, uploadingAsset])

  const handleSend = async () => {
    if (!editor || disabled || uploadingAsset) {
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
    requestAnimationFrame(() => {
      editor.commands.focus("end")
    })
  }

  const handleSelectImage = async (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0]
    event.target.value = ""
    if (!file || !editor || disabled || uploadingAsset) {
      if (editor && shouldRestoreFocusRef.current) {
        requestAnimationFrame(() => {
          editor.commands.focus()
        })
      }
      return
    }
    await insertUploadedImage(file)
  }

  const insertUploadedImage = async (file: File) => {
    if (!editor || disabled || uploadingAsset) {
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
      requestAnimationFrame(() => {
        if (!disabled && shouldRestoreFocusRef.current) {
          editor.commands.focus()
        }
      })
    }
  }

  const handleSelectAttachment = async (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0]
    event.target.value = ""
    if (!file || disabled || uploadingAsset) {
      if (editor && shouldRestoreFocusRef.current) {
        requestAnimationFrame(() => {
          editor.commands.focus()
        })
      }
      return
    }
    shouldRestoreFocusRef.current = editor?.isFocused ?? true
    await onSendAttachmentRef.current(file)
    requestAnimationFrame(() => {
      if (editor && !disabled && shouldRestoreFocusRef.current) {
        editor.commands.focus()
      }
    })
  }

  const handleInsertQuickReply = (item: AdminQuickReply) => {
    if (!editor || disabled || uploadingAsset) {
      return
    }
    if (!item.content.trim()) {
      return
    }
    editor.chain().focus().insertContent(item.content).run()
    setQuickReplyPickerOpen(false)
  }

  return (
    <div className="flex h-full min-h-0 flex-col p-2">
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
      <div className="flex h-full min-h-0 flex-col overflow-hidden rounded-sm border border-border bg-card">
        <div className="min-h-0 flex-1 overflow-hidden px-2 py-1">
          <EditorContent editor={editor} className="h-full" />
        </div>
        <div className="flex items-center justify-between rounded-b-sm border-t border-border bg-card px-2 pt-1 pb-2">
          <div className="flex items-center gap-1">
            <Button
              variant="ghost"
              size="icon"
              className="size-8"
              onMouseDown={(event) => event.preventDefault()}
              onClick={() => {
                shouldRestoreFocusRef.current = editor?.isFocused ?? true
                imageInputRef.current?.click()
              }}
              disabled={disabled || uploadingAsset}
            >
              <ImageIcon className="size-4" />
            </Button>
            <Button
              variant="ghost"
              size="icon"
              className="size-8"
              onMouseDown={(event) => event.preventDefault()}
              onClick={() => {
                shouldRestoreFocusRef.current = editor?.isFocused ?? true
                attachmentInputRef.current?.click()
              }}
              disabled={disabled || uploadingAsset}
            >
              <PaperclipIcon className="size-4" />
            </Button>
            <Popover open={quickReplyPickerOpen} onOpenChange={setQuickReplyPickerOpen}>
              <PopoverTrigger
                render={
                  <Button
                    variant="ghost"
                    size="icon"
                    className="size-8"
                    disabled={disabled || uploadingAsset || loadingQuickReplies}
                    onMouseDown={(event) => event.preventDefault()}
                  />
                }
              >
                <MessageSquareTextIcon className="size-4" />
              </PopoverTrigger>
              <PopoverContent className="w-[30rem] p-0" align="start">
                <Command>
                  <CommandInput placeholder="搜索快捷回复" />
                  <CommandList>
                    <CommandEmpty>暂无快捷回复</CommandEmpty>
                    <CommandGroup>
                      {quickReplies.map((item) => (
                        <CommandItem
                          key={item.id}
                          value={`${item.groupName} ${item.title} ${item.content}`}
                          onSelect={() => handleInsertQuickReply(item)}
                        >
                          <div className="flex min-w-0 flex-col gap-0.5 py-0.5">
                            <span className="line-clamp-1 text-sm">
                              {item.groupName ? `${item.groupName} / ${item.title}` : item.title}
                            </span>
                            <span className="line-clamp-2 text-xs text-muted-foreground">
                              {item.content}
                            </span>
                          </div>
                        </CommandItem>
                      ))}
                    </CommandGroup>
                  </CommandList>
                </Command>
              </PopoverContent>
            </Popover>
          </div>
          <div className="flex items-center gap-2">
            <p className="text-xs text-muted-foreground">Enter 发送</p>
            <Button size="sm" onClick={() => void handleSend()} disabled={disabled || uploadingAsset}>
              <SendIcon className="mr-1 size-4" />
              {uploadingAsset ? "上传中..." : "发送"}
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

function getClipboardImageFile(clipboardData: DataTransfer | null) {
  if (!clipboardData) {
    return null
  }
  for (const item of Array.from(clipboardData.items)) {
    if (item.kind === "file" && item.type.startsWith("image/")) {
      return item.getAsFile()
    }
  }
  return null
}
