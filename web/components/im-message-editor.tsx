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

type UploadedImage = {
  url: string
  filename?: string
}

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
      Image,
      Placeholder.configure({
        placeholder: "输入消息，Enter 发送，Shift + Enter 换行",
      }),
    ],
    content: "",
    editorProps: {
      attributes: {
        class:
          "h-full min-h-12 overflow-y-auto px-1.5 py-1 text-sm leading-6 text-foreground outline-none [&_.ProseMirror-focused]:outline-none [&_p]:m-0 [&_p+img]:mt-2 [&_img]:my-2 [&_img]:max-h-64 [&_img]:rounded-md [&_img]:object-contain [&_p.is-editor-empty:first-child]:before:text-muted-foreground",
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
    const html = editor.getHTML()
    if (!isMeaningfulHTML(html)) {
      return
    }
    await onSendRef.current(html)
    editor.commands.clearContent(true)
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
    const placeholderId = `uploading-${crypto.randomUUID()}`
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
      const uploaded = await onUploadImageRef.current(file)
      if (!uploaded?.url) {
        removeImageByTitle(editor, placeholderId)
        return
      }
      replaceImageSourceByTitle(editor, placeholderId, uploaded.url, uploaded.filename || "image")
    } finally {
      URL.revokeObjectURL(objectUrl)
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
      <div className="flex h-full min-h-0 flex-col rounded-sm border border-border bg-card">
        <div className="min-h-0 flex-1 px-2 py-1">
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
  src: string,
  alt: string
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
    src,
    alt,
    title: "",
  })
  view.dispatch(transaction)
}
