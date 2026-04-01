"use client"

import { forwardRef, useId, useImperativeHandle, useRef } from "react"
import { MdEditor, type ExposeParam } from "md-editor-rt"
import { useTheme } from "next-themes"

import "./markdown-editor.css"

import type { UploadImageHandler } from "./types"

export type MarkdownEditorRef = {
  focus: () => void
}

type MarkdownEditorProps = {
  value: string
  onChange: (nextValue: string) => void
  placeholder?: string
  disabled?: boolean
  onUploadImage?: UploadImageHandler
  height: string
}

export const MarkdownEditor = forwardRef<MarkdownEditorRef, MarkdownEditorProps>(
  function MarkdownEditor(
    {
      value,
      onChange,
      placeholder = "请输入 Markdown 内容...",
      disabled = false,
      onUploadImage,
      height,
    },
    ref
  ) {
    const editorId = useId()
    const editorRef = useRef<ExposeParam>(null)
    const { resolvedTheme } = useTheme()

    useImperativeHandle(ref, () => ({
      focus() {
        editorRef.current?.focus()
      },
    }))

    return (
      <div
        className="w-full rounded-lg border bg-background"
        style={{ height }}
      >
        <div className="content-editor-markdown h-full">
          <MdEditor
            ref={editorRef}
            id={editorId}
            value={value}
            onChange={onChange}
            theme={resolvedTheme === "dark" ? "dark" : "light"}
            preview={false}
            toolbars={[
              "bold",
              "underline",
              "italic",
              "strikeThrough",
              "-",
              "title",
              "quote",
              "unorderedList",
              "orderedList",
              "-",
              "codeRow",
              "code",
              "link",
              "image",
              "-",
              "revoke",
              "next",
              "=",
              "previewOnly",
            ]}
            footers={[]}
            noMermaid
            noKatex
            noHighlight
            placeholder={placeholder}
            disabled={disabled}
            style={{ height: "100%" }}
            onUploadImg={
              onUploadImage
                ? async (files, callback) => {
                    const uploadedUrls: string[] = []
                    for (const file of files) {
                      const uploaded = await onUploadImage(file)
                      if (uploaded?.url) {
                        uploadedUrls.push(uploaded.url)
                      }
                    }
                    callback(uploadedUrls)
                  }
                : undefined
            }
          />
        </div>
      </div>
    )
  }
)
