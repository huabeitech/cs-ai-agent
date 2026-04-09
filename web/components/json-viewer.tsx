"use client"

import { cn } from "@/lib/utils"

type JsonViewerProps = {
  value: unknown
  emptyText?: string
  className?: string
}

function escapeHtml(value: string) {
  return value
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
}

function formatJson(value: unknown) {
  if (value === undefined) {
    return ""
  }
  try {
    return JSON.stringify(value, null, 2)
  } catch {
    return String(value)
  }
}

function highlightJson(value: string) {
  const escaped = escapeHtml(value)
  return escaped.replace(
    /("(?:\\u[\da-fA-F]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d+)?(?:[eE][+-]?\d+)?)/g,
    (match) => {
      let className = "text-slate-200"
      if (match.startsWith('"')) {
        className = match.endsWith(":") ? "text-sky-300" : "text-emerald-300"
      } else if (match === "true" || match === "false") {
        className = "text-amber-300"
      } else if (match === "null") {
        className = "text-rose-300"
      } else {
        className = "text-violet-300"
      }
      return `<span class="${className}">${match}</span>`
    }
  )
}

export function JsonViewer({
  value,
  emptyText = "暂无数据",
  className,
}: JsonViewerProps) {
  const formatted = formatJson(value)
  if (!formatted) {
    return (
      <div
        className={cn(
          "rounded-md border bg-slate-950 px-4 py-3 font-mono text-xs leading-6 text-slate-400",
          className
        )}
      >
        {emptyText}
      </div>
    )
  }

  return (
    <pre
      className={cn(
        "overflow-x-auto rounded-md border bg-slate-950 px-4 py-3 font-mono text-xs leading-6 text-slate-100",
        className
      )}
      dangerouslySetInnerHTML={{ __html: highlightJson(formatted) }}
    />
  )
}
