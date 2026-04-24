"use client"

import {
  forwardRef,
  useId,
  useImperativeHandle,
  useMemo,
  useRef,
} from "react"
import { Maximize2Icon, Minimize2Icon } from "lucide-react"
import { MdEditor, NormalToolbar, type ExposeParam } from "md-editor-rt"
import { useTheme } from "next-themes"

import "./markdown-editor.css"

import { EditorModeSwitch } from "./editor-mode-switch"
import type { ContentMode, UploadImageHandler } from "./types"

export type MarkdownEditorRef = {
  focus: () => void
}

type MarkdownEditorProps = {
  value: string
  onChange: (nextValue: string) => void
  mode: ContentMode
  allowedModes: ReadonlyArray<ContentMode>
  onModeChange: (nextMode: ContentMode) => void
  fullscreen: boolean
  onToggleFullscreen: () => void
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
      mode,
      allowedModes,
      onModeChange,
      fullscreen,
      onToggleFullscreen,
      placeholder = "",
      disabled = false,
      onUploadImage,
      height,
    },
    ref
  ) {
    const editorId = useId()
    const editorRef = useRef<ExposeParam>(null)
    const { resolvedTheme } = useTheme()
    const defToolbars = useMemo(
      () => [
        <EditorModeSwitch
          key="mode-switch"
          value={mode}
          allowedModes={allowedModes}
          disabled={disabled}
          onChange={onModeChange}
        />,
        <NormalToolbar
          key="toggle-fullscreen"
          title={fullscreen ? "退出全屏" : "全屏"}
          disabled={disabled}
          onClick={onToggleFullscreen}
        >
          {fullscreen ? (
            <Minimize2Icon className="h-[16px] w-[16px]" />
          ) : (
            <Maximize2Icon className="h-[16px] w-[16px]" />
          )}
        </NormalToolbar>,
      ],
      [allowedModes, disabled, fullscreen, mode, onModeChange, onToggleFullscreen]
    )

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
              0,
              "-",
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
              1,
              "=",
              "preview",
              "previewOnly",
            ]}
            defToolbars={defToolbars}
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
