"use client"

import { useEffect } from "react"
import { EditorContent, useEditor } from "@tiptap/react"
import StarterKit from "@tiptap/starter-kit"
import Placeholder from "@tiptap/extension-placeholder"
import {
  BoldIcon,
  ItalicIcon,
  ListIcon,
  ListOrderedIcon,
  QuoteIcon,
  RedoIcon,
  UndoIcon,
} from "lucide-react"

import { Button } from "@/components/ui/button"
import {
  ToggleGroup,
  ToggleGroupItem,
} from "@/components/ui/toggle-group"
import { Separator } from "@/components/ui/separator"

type RichTextEditorProps = {
  content: string
  onChange: (html: string) => void
  placeholder?: string
  disabled?: boolean
}

export function RichTextEditor({
  content,
  onChange,
  placeholder = "输入内容...",
  disabled = false,
}: RichTextEditorProps) {
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
    content,
    editable: !disabled,
    onUpdate: ({ editor }) => {
      const html = editor.getHTML()
      onChange(html)
    },
    editorProps: {
      attributes: {
        class:
          "min-h-64 max-h-96 overflow-y-auto px-4 py-3 text-sm leading-7 text-slate-900 outline-none [&_.ProseMirror-focused]:outline-none [&_p]:m-0 [&_p]:mb-2 [&_h1]:text-2xl [&_h1]:font-bold [&_h1]:mb-3 [&_h2]:text-xl [&_h2]:font-semibold [&_h2]:mb-2 [&_h3]:text-lg [&_h3]:font-semibold [&_h3]:mb-2 [&_ul]:list-disc [&_ul]:pl-6 [&_ol]:list-decimal [&_ol]:pl-6 [&_li]:mb-1 [&_blockquote]:border-l-4 [&_blockquote]:border-muted-foreground [&_blockquote]:pl-4 [&_blockquote]:italic [&_blockquote]:text-muted-foreground",
      },
    },
  })

  useEffect(() => {
    if (editor && content !== editor.getHTML()) {
      editor.commands.setContent(content)
    }
  }, [content, editor])

  useEffect(() => {
    if (editor) {
      editor.setEditable(!disabled)
    }
  }, [disabled, editor])

  if (!editor) {
    return null
  }

  return (
    <div className="rounded-lg border bg-background">
      <div className="flex items-center gap-1 border-b p-2">
        <ToggleGroup className="flex-wrap gap-1">
          <ToggleGroupItem
            value="undo"
            aria-label="撤销"
            disabled={!editor.can().undo() || disabled}
            onClick={() => editor.chain().focus().undo().run()}
          >
            <UndoIcon className="size-4" />
          </ToggleGroupItem>
          <ToggleGroupItem
            value="redo"
            aria-label="重做"
            disabled={!editor.can().redo() || disabled}
            onClick={() => editor.chain().focus().redo().run()}
          >
            <RedoIcon className="size-4" />
          </ToggleGroupItem>
          <Separator orientation="vertical" className="mx-1 h-6" />
          <ToggleGroupItem
            value="bold"
            aria-label="粗体"
            disabled={disabled}
            pressed={editor.isActive("bold")}
            onClick={() => editor.chain().focus().toggleBold().run()}
          >
            <BoldIcon className="size-4" />
          </ToggleGroupItem>
          <ToggleGroupItem
            value="italic"
            aria-label="斜体"
            disabled={disabled}
            pressed={editor.isActive("italic")}
            onClick={() => editor.chain().focus().toggleItalic().run()}
          >
            <ItalicIcon className="size-4" />
          </ToggleGroupItem>
          <Separator orientation="vertical" className="mx-1 h-6" />
          <ToggleGroupItem
            value="bulletList"
            aria-label="无序列表"
            disabled={disabled}
            pressed={editor.isActive("bulletList")}
            onClick={() => editor.chain().focus().toggleBulletList().run()}
          >
            <ListIcon className="size-4" />
          </ToggleGroupItem>
          <ToggleGroupItem
            value="orderedList"
            aria-label="有序列表"
            disabled={disabled}
            pressed={editor.isActive("orderedList")}
            onClick={() => editor.chain().focus().toggleOrderedList().run()}
          >
            <ListOrderedIcon className="size-4" />
          </ToggleGroupItem>
          <ToggleGroupItem
            value="blockquote"
            aria-label="引用"
            disabled={disabled}
            pressed={editor.isActive("blockquote")}
            onClick={() => editor.chain().focus().toggleBlockquote().run()}
          >
            <QuoteIcon className="size-4" />
          </ToggleGroupItem>
        </ToggleGroup>
      </div>
      <div className="p-2">
        <EditorContent editor={editor} />
      </div>
    </div>
  )
}
