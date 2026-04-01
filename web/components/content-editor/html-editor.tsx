"use client"

import {
  forwardRef,
  useEffect,
  useImperativeHandle,
  useRef,
  useState,
  type ChangeEvent,
} from "react"
import { EditorContent, useEditor } from "@tiptap/react"
import Image from "@tiptap/extension-image"
import Link from "@tiptap/extension-link"
import Placeholder from "@tiptap/extension-placeholder"
import StarterKit from "@tiptap/starter-kit"
import Underline from "@tiptap/extension-underline"
import {
  BoldIcon,
  Code2Icon,
  Heading1Icon,
  Heading2Icon,
  ImageIcon,
  ItalicIcon,
  LinkIcon,
  ListIcon,
  ListOrderedIcon,
  Columns2Icon,
  QuoteIcon,
  RedoIcon,
  RotateCcwIcon,
  StrikethroughIcon,
  EyeIcon,
  Maximize2Icon,
  Minimize2Icon,
  UnderlineIcon,
} from "lucide-react"

import { EditorToolbar } from "./toolbar"
import type { ContentMode, EditorToolbarAction, UploadImageHandler } from "./types"

export type HtmlEditorRef = {
  focus: () => void
}

type HtmlEditorProps = {
  value: string
  onChange: (nextValue: string) => void
  mode: ContentMode
  onModeChange: (nextMode: ContentMode) => void
  fullscreen: boolean
  onToggleFullscreen: () => void
  placeholder?: string
  disabled?: boolean
  onUploadImage?: UploadImageHandler
  height: string
}

export const HtmlEditor = forwardRef<HtmlEditorRef, HtmlEditorProps>(
  function HtmlEditor(
    {
      value,
      onChange,
      mode,
      onModeChange,
      fullscreen,
      onToggleFullscreen,
      placeholder = "请输入内容...",
      disabled = false,
      onUploadImage,
      height,
    },
    ref
  ) {
    const imageInputRef = useRef<HTMLInputElement>(null)
    const [showPreview, setShowPreview] = useState(false)
    const [previewOnly, setPreviewOnly] = useState(false)
    const proseClassName =
      "h-full overflow-y-auto px-4 py-3 text-sm leading-7 text-foreground outline-none [&_.ProseMirror-focused]:outline-none [&_p]:m-0 [&_p]:mb-2 [&_h1]:mb-3 [&_h1]:text-2xl [&_h1]:font-bold [&_h2]:mb-2 [&_h2]:text-xl [&_h2]:font-semibold [&_ul]:list-disc [&_ul]:pl-6 [&_ol]:list-decimal [&_ol]:pl-6 [&_li]:mb-1 [&_blockquote]:border-l-4 [&_blockquote]:border-border [&_blockquote]:pl-4 [&_blockquote]:italic [&_blockquote]:text-muted-foreground [&_pre]:overflow-x-auto [&_pre]:rounded-md [&_pre]:bg-muted [&_pre]:p-3 [&_code]:rounded-sm [&_code]:bg-muted [&_code]:px-1.5 [&_code]:py-0.5 [&_img]:my-2 [&_img]:max-h-80 [&_img]:rounded-md [&_img]:object-contain [&_p.is-editor-empty:first-child]:before:text-muted-foreground"

    const editor = useEditor({
      immediatelyRender: false,
      extensions: [
        StarterKit.configure({
          heading: {
            levels: [1, 2],
          },
          bulletList: {
            keepMarks: true,
            keepAttributes: false,
          },
          orderedList: {
            keepMarks: true,
            keepAttributes: false,
          },
        }),
        Image,
        Link.configure({
          openOnClick: false,
          autolink: true,
        }),
        Underline,
        Placeholder.configure({
          placeholder,
        }),
      ],
      content: value,
      editable: !disabled,
      onUpdate: ({ editor: currentEditor }) => {
        onChange(currentEditor.getHTML())
      },
      editorProps: {
        attributes: {
          class: proseClassName,
        },
      },
    })

    useImperativeHandle(ref, () => ({
      focus() {
        editor?.commands.focus()
      },
    }), [editor])

    useEffect(() => {
      if (editor && value !== editor.getHTML()) {
        editor.commands.setContent(value, { emitUpdate: false })
      }
    }, [editor, value])

    useEffect(() => {
      if (editor) {
        editor.setEditable(!disabled)
      }
    }, [disabled, editor])

    const handleTogglePreview = () => {
      setPreviewOnly(false)
      setShowPreview((current) => !current)
    }

    const handleTogglePreviewOnly = () => {
      setPreviewOnly((current) => {
        const next = !current
        setShowPreview(next)
        return next
      })
    }

    const handleInsertLink = () => {
      if (!editor || disabled) {
        return
      }
      const previousUrl = editor.getAttributes("link").href as string | undefined
      const url = window.prompt("输入链接地址", previousUrl || "https://")
      if (url === null) {
        return
      }
      if (!url.trim()) {
        editor.chain().focus().unsetLink().run()
        return
      }
      editor.chain().focus().extendMarkRange("link").setLink({ href: url.trim() }).run()
    }

    const handleSelectImage = async (event: ChangeEvent<HTMLInputElement>) => {
      const file = event.target.files?.[0]
      event.target.value = ""
      if (!file || !editor || !onUploadImage || disabled) {
        return
      }
      const uploaded = await onUploadImage(file)
      if (!uploaded?.url) {
        return
      }
      editor
        .chain()
        .focus()
        .setImage({
          src: uploaded.url,
          alt: uploaded.alt || file.name || "image",
          title: uploaded.title || "",
        })
        .run()
    }

    const actions: EditorToolbarAction[] = [
      {
        key: "mode-markdown",
        label: "Markdown 模式",
        content: "Markdown",
        disabled,
        pressed: mode === "markdown",
        onClick: () => onModeChange("markdown"),
      },
      {
        key: "mode-html",
        label: "HTML 模式",
        content: "HTML",
        disabled,
        pressed: mode === "html",
        onClick: () => onModeChange("html"),
      },
      { key: "separator-mode", type: "separator" },
      {
        key: "bold",
        label: "粗体",
        icon: BoldIcon,
        disabled,
        pressed: !!editor?.isActive("bold"),
        onClick: () => editor?.chain().focus().toggleBold().run(),
      },
      {
        key: "underline",
        label: "下划线",
        icon: UnderlineIcon,
        disabled,
        pressed: !!editor?.isActive("underline"),
        onClick: () => editor?.chain().focus().toggleUnderline().run(),
      },
      {
        key: "italic",
        label: "斜体",
        icon: ItalicIcon,
        disabled,
        pressed: !!editor?.isActive("italic"),
        onClick: () => editor?.chain().focus().toggleItalic().run(),
      },
      {
        key: "strike",
        label: "删除线",
        icon: StrikethroughIcon,
        disabled,
        pressed: !!editor?.isActive("strike"),
        onClick: () => editor?.chain().focus().toggleStrike().run(),
      },
      { key: "separator-1", type: "separator" },
      {
        key: "h1",
        label: "一级标题",
        icon: Heading1Icon,
        disabled,
        pressed: !!editor?.isActive("heading", { level: 1 }),
        onClick: () => editor?.chain().focus().toggleHeading({ level: 1 }).run(),
      },
      {
        key: "h2",
        label: "二级标题",
        icon: Heading2Icon,
        disabled,
        pressed: !!editor?.isActive("heading", { level: 2 }),
        onClick: () => editor?.chain().focus().toggleHeading({ level: 2 }).run(),
      },
      {
        key: "quote",
        label: "引用",
        icon: QuoteIcon,
        disabled,
        pressed: !!editor?.isActive("blockquote"),
        onClick: () => editor?.chain().focus().toggleBlockquote().run(),
      },
      {
        key: "bullet-list",
        label: "无序列表",
        icon: ListIcon,
        disabled,
        pressed: !!editor?.isActive("bulletList"),
        onClick: () => editor?.chain().focus().toggleBulletList().run(),
      },
      {
        key: "ordered-list",
        label: "有序列表",
        icon: ListOrderedIcon,
        disabled,
        pressed: !!editor?.isActive("orderedList"),
        onClick: () => editor?.chain().focus().toggleOrderedList().run(),
      },
      { key: "separator-2", type: "separator" },
      {
        key: "code",
        label: "行内代码",
        icon: Code2Icon,
        disabled,
        pressed: !!editor?.isActive("code"),
        onClick: () => editor?.chain().focus().toggleCode().run(),
      },
      {
        key: "code-block",
        label: "代码块",
        icon: Code2Icon,
        disabled,
        pressed: !!editor?.isActive("codeBlock"),
        onClick: () => editor?.chain().focus().toggleCodeBlock().run(),
      },
      { key: "separator-3", type: "separator" },
      {
        key: "link",
        label: "链接",
        icon: LinkIcon,
        disabled,
        pressed: !!editor?.isActive("link"),
        onClick: handleInsertLink,
      },
      {
        key: "image",
        label: "图片",
        icon: ImageIcon,
        disabled: disabled || !onUploadImage,
        onClick: () => imageInputRef.current?.click(),
      },
      { key: "separator-4", type: "separator" },
      {
        key: "undo-tail",
        label: "撤销",
        icon: RotateCcwIcon,
        disabled: disabled || !editor?.can().undo(),
        onClick: () => editor?.chain().focus().undo().run(),
      },
      {
        key: "redo-tail",
        label: "重做",
        icon: RedoIcon,
        disabled: disabled || !editor?.can().redo(),
        onClick: () => editor?.chain().focus().redo().run(),
      },
      { key: "separator-fullscreen", type: "separator" },
      {
        key: "fullscreen",
        label: fullscreen ? "退出全屏" : "全屏",
        icon: fullscreen ? Minimize2Icon : Maximize2Icon,
        disabled,
        pressed: fullscreen,
        onClick: onToggleFullscreen,
      },
      { key: "separator-preview", type: "separator" },
      {
        key: "preview",
        label: "分屏预览",
        icon: Columns2Icon,
        disabled,
        pressed: showPreview && !previewOnly,
        onClick: handleTogglePreview,
      },
      {
        key: "preview-only",
        label: "仅预览",
        icon: EyeIcon,
        disabled,
        pressed: previewOnly,
        onClick: handleTogglePreviewOnly,
      },
    ]

    if (!editor) {
      return null
    }

    return (
      <div
        className="flex w-full flex-col rounded-lg border bg-background"
        style={{ height }}
      >
        <input
          ref={imageInputRef}
          type="file"
          accept="image/*"
          className="hidden"
          onChange={(event) => {
            void handleSelectImage(event)
          }}
        />
        <EditorToolbar actions={actions} />
        <div className="min-h-0 flex-1 p-2">
          {previewOnly ? (
            <div
              className={proseClassName}
              dangerouslySetInnerHTML={{ __html: value }}
            />
          ) : showPreview ? (
            <div className="flex h-full overflow-hidden rounded-md border border-border">
              <div className="min-w-0 flex-1 border-r border-border">
                <EditorContent editor={editor} className="h-full" />
              </div>
              <div className="min-w-0 flex-1">
                <div
                  className={proseClassName}
                  dangerouslySetInnerHTML={{ __html: value }}
                />
              </div>
            </div>
          ) : (
            <EditorContent editor={editor} className="h-full" />
          )}
        </div>
      </div>
    )
  }
)
