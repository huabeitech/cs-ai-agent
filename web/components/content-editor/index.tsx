"use client"

import { useCallback } from "react"

import { Button } from "@/components/ui/button"

import { htmlToMarkdown, markdownToHtml } from "./convert"
import { HtmlEditor } from "./html-editor"
import { MarkdownEditor } from "./markdown-editor"
import type { ContentMode, ContentValue, UploadImageHandler } from "./types"

type ContentEditorProps = {
  value: ContentValue
  onChange: (next: ContentValue) => void
  placeholder?: string
  disabled?: boolean
  onUploadImage?: UploadImageHandler
  height?: number | string
}

function normalizeHeight(height?: number | string) {
  if (typeof height === "number") {
    return `${height}px`
  }
  if (typeof height === "string" && height.trim()) {
    return height
  }
  return "400px"
}

function getModeLabel(mode: ContentMode) {
  return mode === "markdown" ? "Markdown" : "HTML"
}

function convertContent(mode: ContentMode, raw: string) {
  if (mode === "markdown") {
    return markdownToHtml(raw)
  }
  return htmlToMarkdown(raw)
}

export function ContentEditor({
  value,
  onChange,
  placeholder,
  disabled = false,
  onUploadImage,
  height,
}: ContentEditorProps) {
  const editorHeight = normalizeHeight(height)

  const handleModeChange = useCallback(
    (nextMode: ContentMode) => {
      if (disabled || nextMode === value.mode) {
        return
      }
      const currentText = value.raw.trim()
      if (!currentText) {
        onChange({ mode: nextMode, raw: "" })
        return
      }

      const confirmed = window.confirm(
        `切换到 ${getModeLabel(nextMode)} 模式会尝试自动转换内容，复杂格式可能有损。是否继续？`
      )
      if (!confirmed) {
        return
      }

      onChange({
        mode: nextMode,
        raw: convertContent(value.mode, value.raw),
      })
    },
    [disabled, onChange, value.mode, value.raw]
  )

  return (
    <div className="space-y-3">
      <div className="flex items-center gap-2">
        <Button
          type="button"
          variant={value.mode === "markdown" ? "default" : "outline"}
          size="sm"
          disabled={disabled}
          onClick={() => handleModeChange("markdown")}
        >
          Markdown
        </Button>
        <Button
          type="button"
          variant={value.mode === "html" ? "default" : "outline"}
          size="sm"
          disabled={disabled}
          onClick={() => handleModeChange("html")}
        >
          HTML
        </Button>
      </div>

      {value.mode === "markdown" ? (
        <MarkdownEditor
          value={value.raw}
          onChange={(nextRaw) => onChange({ mode: "markdown", raw: nextRaw })}
          placeholder={placeholder}
          disabled={disabled}
          onUploadImage={onUploadImage}
          height={editorHeight}
        />
      ) : (
        <HtmlEditor
          value={value.raw}
          onChange={(nextRaw) => onChange({ mode: "html", raw: nextRaw })}
          placeholder={placeholder}
          disabled={disabled}
          onUploadImage={onUploadImage}
          height={editorHeight}
        />
      )}
    </div>
  )
}

export type { ContentMode, ContentValue, UploadImageHandler, UploadImageResult } from "./types"
