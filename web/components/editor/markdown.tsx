"use client"

import { useRef } from "react"
import {
  BoldIcon,
  CodeIcon,
  Heading1Icon,
  ItalicIcon,
  LinkIcon,
  ListIcon,
  ListOrderedIcon,
  QuoteIcon,
} from "lucide-react"

import { Textarea } from "@/components/ui/textarea"

import { EditorToolbar } from "./toolbar"
import type { BaseEditorProps } from "./types"

export type MarkdownEditorProps = BaseEditorProps & {
  value: string
  onChange: (nextValue: string) => void
  rows?: number
}

export function MarkdownEditor({
  value,
  onChange,
  placeholder = "请输入 Markdown 内容...",
  disabled = false,
  rows = 16,
  className,
}: MarkdownEditorProps) {
  const textareaRef = useRef<HTMLTextAreaElement | null>(null)

  const handleWrapSelection = (prefix: string, suffix = prefix) => {
    const textarea = textareaRef.current
    if (!textarea || disabled) {
      return
    }
    const start = textarea.selectionStart ?? 0
    const end = textarea.selectionEnd ?? 0
    const selected = value.slice(start, end)
    const next = `${value.slice(0, start)}${prefix}${selected}${suffix}${value.slice(end)}`
    onChange(next)
    requestAnimationFrame(() => {
      textarea.focus()
      textarea.setSelectionRange(start + prefix.length, end + prefix.length)
    })
  }

  const handleInsertLinePrefix = (prefix: string) => {
    const textarea = textareaRef.current
    if (!textarea || disabled) {
      return
    }
    const start = textarea.selectionStart ?? 0
    const end = textarea.selectionEnd ?? 0
    const lineStart = value.lastIndexOf("\n", start - 1) + 1
    const lineEndRaw = value.indexOf("\n", end)
    const lineEnd = lineEndRaw === -1 ? value.length : lineEndRaw
    const selectedLines = value.slice(lineStart, lineEnd)
    const nextLines = selectedLines
      .split("\n")
      .map((line) => `${prefix}${line}`)
      .join("\n")
    const next = `${value.slice(0, lineStart)}${nextLines}${value.slice(lineEnd)}`
    onChange(next)
    requestAnimationFrame(() => {
      textarea.focus()
      textarea.setSelectionRange(lineStart, lineStart + nextLines.length)
    })
  }

  const handleInsertLink = () => {
    const textarea = textareaRef.current
    if (!textarea || disabled) {
      return
    }
    const start = textarea.selectionStart ?? 0
    const end = textarea.selectionEnd ?? 0
    const selected = value.slice(start, end) || "链接文本"
    const markdown = `[${selected}](https://)`
    const next = `${value.slice(0, start)}${markdown}${value.slice(end)}`
    onChange(next)
    requestAnimationFrame(() => {
      textarea.focus()
      const urlStart = start + markdown.lastIndexOf("https://")
      textarea.setSelectionRange(urlStart, urlStart + "https://".length)
    })
  }

  const toolbarActions = [
    {
      key: "heading1",
      label: "一级标题",
      icon: Heading1Icon,
      disabled,
      onClick: () => handleInsertLinePrefix("# "),
    },
    { key: "separator-1", type: "separator" as const },
    {
      key: "bold",
      label: "粗体",
      icon: BoldIcon,
      disabled,
      onClick: () => handleWrapSelection("**"),
    },
    {
      key: "italic",
      label: "斜体",
      icon: ItalicIcon,
      disabled,
      onClick: () => handleWrapSelection("*"),
    },
    {
      key: "code",
      label: "行内代码",
      icon: CodeIcon,
      disabled,
      onClick: () => handleWrapSelection("`"),
    },
    { key: "separator-2", type: "separator" as const },
    {
      key: "bulletList",
      label: "无序列表",
      icon: ListIcon,
      disabled,
      onClick: () => handleInsertLinePrefix("- "),
    },
    {
      key: "orderedList",
      label: "有序列表",
      icon: ListOrderedIcon,
      disabled,
      onClick: () => handleInsertLinePrefix("1. "),
    },
    {
      key: "blockquote",
      label: "引用",
      icon: QuoteIcon,
      disabled,
      onClick: () => handleInsertLinePrefix("> "),
    },
    {
      key: "link",
      label: "链接",
      icon: LinkIcon,
      disabled,
      onClick: handleInsertLink,
    },
  ] as const

  return (
    <div className="rounded-lg border bg-background">
      <EditorToolbar actions={toolbarActions} />
      <div className="p-2">
        <Textarea
          ref={textareaRef}
          value={value}
          rows={rows}
          disabled={disabled}
          placeholder={placeholder}
          className={`min-h-64 max-h-96 resize-y border-0 px-2 py-2 text-sm leading-7 shadow-none focus-visible:ring-0 ${className ?? ""}`}
          onChange={(event) => onChange(event.target.value)}
        />
      </div>
    </div>
  )
}

