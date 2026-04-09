"use client"

import { HtmlEditor } from "./html"
import { MarkdownEditor } from "./markdown"
import type { BaseEditorProps, EditorValue } from "./types"

export type UnifiedEditorProps = BaseEditorProps & {
  value: EditorValue
  onChange: (next: EditorValue) => void
  markdownRows?: number
}

export function UnifiedEditor({
  value,
  onChange,
  placeholder,
  disabled,
  features,
  className,
  markdownRows,
}: UnifiedEditorProps) {
  if (value.mode === "markdown") {
    return (
      <MarkdownEditor
        value={value.raw}
        onChange={(nextRaw) => onChange({ ...value, raw: nextRaw })}
        placeholder={placeholder}
        disabled={disabled}
        features={features}
        className={className}
        rows={markdownRows}
      />
    )
  }

  return (
    <HtmlEditor
      value={value.raw}
      onChange={(nextRaw) => onChange({ ...value, raw: nextRaw })}
      placeholder={placeholder}
      disabled={disabled}
      features={features}
      className={className}
    />
  )
}

export * from "./types"

