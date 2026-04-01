"use client"

import { useCallback, useEffect, useState } from "react"
import { createPortal } from "react-dom"

import { cn } from "@/lib/utils"

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
  const [fullscreen, setFullscreen] = useState(false)
  const [mounted, setMounted] = useState(false)

  useEffect(() => {
    setMounted(true)
  }, [])

  useEffect(() => {
    if (!fullscreen) {
      return
    }

    const previousOverflow = document.body.style.overflow
    document.body.style.overflow = "hidden"

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        setFullscreen(false)
      }
    }

    window.addEventListener("keydown", handleKeyDown)
    return () => {
      document.body.style.overflow = previousOverflow
      window.removeEventListener("keydown", handleKeyDown)
    }
  }, [fullscreen])

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

  const content = (
    <div
      className={cn(
        "w-full",
        fullscreen && "fixed inset-0 z-[10000] overflow-hidden bg-background p-4"
      )}
    >
      {value.mode === "markdown" ? (
        <MarkdownEditor
          value={value.raw}
          onChange={(nextRaw) => onChange({ mode: "markdown", raw: nextRaw })}
          mode={value.mode}
          onModeChange={handleModeChange}
          fullscreen={fullscreen}
          onToggleFullscreen={() => setFullscreen((current) => !current)}
          placeholder={placeholder}
          disabled={disabled}
          onUploadImage={onUploadImage}
          height={fullscreen ? "calc(100vh - 5rem)" : editorHeight}
        />
      ) : (
        <HtmlEditor
          value={value.raw}
          onChange={(nextRaw) => onChange({ mode: "html", raw: nextRaw })}
          mode={value.mode}
          onModeChange={handleModeChange}
          fullscreen={fullscreen}
          onToggleFullscreen={() => setFullscreen((current) => !current)}
          placeholder={placeholder}
          disabled={disabled}
          onUploadImage={onUploadImage}
          height={fullscreen ? "calc(100vh - 5rem)" : editorHeight}
        />
      )}
    </div>
  )

  if (fullscreen && mounted) {
    return createPortal(content, document.body)
  }

  return content
}

export type { ContentMode, ContentValue, UploadImageHandler, UploadImageResult } from "./types"
