"use client"

import { useEffect } from "react"
import { EditorContent, useEditor } from "@tiptap/react"
import Placeholder from "@tiptap/extension-placeholder"
import StarterKit from "@tiptap/starter-kit"
import {
  BoldIcon,
  ItalicIcon,
  ListIcon,
  ListOrderedIcon,
  QuoteIcon,
  RedoIcon,
  UndoIcon,
} from "lucide-react"

import { EditorToolbar } from "./toolbar"
import type { BaseEditorProps } from "./types"

export type HtmlEditorProps = BaseEditorProps & {
  value: string
  onChange: (nextValue: string) => void
}

export function HtmlEditor({
  value,
  onChange,
  placeholder = "请输入内容...",
  disabled = false,
}: HtmlEditorProps) {
  const editor = useEditor({
    immediatelyRender: false,
    extensions: [
      StarterKit.configure({
        heading: {
          levels: [1, 2, 3],
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
      Placeholder.configure({
        placeholder,
      }),
    ],
    content: value,
    editable: !disabled,
    onUpdate: ({ editor }) => {
      onChange(editor.getHTML())
    },
    editorProps: {
      attributes: {
        class:
          "min-h-64 max-h-96 overflow-y-auto px-4 py-3 text-sm leading-7 text-slate-900 outline-none [&_.ProseMirror-focused]:outline-none [&_p]:m-0 [&_p]:mb-2 [&_h1]:text-2xl [&_h1]:font-bold [&_h1]:mb-3 [&_h2]:text-xl [&_h2]:font-semibold [&_h2]:mb-2 [&_h3]:text-lg [&_h3]:font-semibold [&_h3]:mb-2 [&_ul]:list-disc [&_ul]:pl-6 [&_ol]:list-decimal [&_ol]:pl-6 [&_li]:mb-1 [&_blockquote]:border-l-4 [&_blockquote]:border-muted-foreground [&_blockquote]:pl-4 [&_blockquote]:italic [&_blockquote]:text-muted-foreground",
      },
    },
  })

  useEffect(() => {
    if (editor && value !== editor.getHTML()) {
      editor.commands.setContent(value)
    }
  }, [editor, value])

  useEffect(() => {
    if (editor) {
      editor.setEditable(!disabled)
    }
  }, [disabled, editor])

  if (!editor) {
    return null
  }

  const toolbarActions = [
    {
      key: "undo",
      label: "撤销",
      icon: UndoIcon,
      disabled: !editor.can().undo() || disabled,
      onClick: () => editor.chain().focus().undo().run(),
    },
    {
      key: "redo",
      label: "重做",
      icon: RedoIcon,
      disabled: !editor.can().redo() || disabled,
      onClick: () => editor.chain().focus().redo().run(),
    },
    { key: "separator-1", type: "separator" as const },
    {
      key: "bold",
      label: "粗体",
      icon: BoldIcon,
      disabled,
      pressed: editor.isActive("bold"),
      onClick: () => editor.chain().focus().toggleBold().run(),
    },
    {
      key: "italic",
      label: "斜体",
      icon: ItalicIcon,
      disabled,
      pressed: editor.isActive("italic"),
      onClick: () => editor.chain().focus().toggleItalic().run(),
    },
    { key: "separator-2", type: "separator" as const },
    {
      key: "bulletList",
      label: "无序列表",
      icon: ListIcon,
      disabled,
      pressed: editor.isActive("bulletList"),
      onClick: () => editor.chain().focus().toggleBulletList().run(),
    },
    {
      key: "orderedList",
      label: "有序列表",
      icon: ListOrderedIcon,
      disabled,
      pressed: editor.isActive("orderedList"),
      onClick: () => editor.chain().focus().toggleOrderedList().run(),
    },
    {
      key: "blockquote",
      label: "引用",
      icon: QuoteIcon,
      disabled,
      pressed: editor.isActive("blockquote"),
      onClick: () => editor.chain().focus().toggleBlockquote().run(),
    },
  ] as const

  return (
    <div className="rounded-lg border bg-background">
      <EditorToolbar actions={toolbarActions} />
      <div className="p-2">
        <EditorContent editor={editor} />
      </div>
    </div>
  )
}

