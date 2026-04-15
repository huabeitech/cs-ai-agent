"use client"

import type { CSSProperties } from "react"
import JsonView from "@uiw/react-json-view"

import { cn } from "@/lib/utils"

type JsonTreeViewerProps = {
  value: unknown
  emptyText?: string
  className?: string
  collapsed?: boolean | number
}

const viewerTheme = {
  "--w-rjv-font-family":
    'ui-monospace, SFMono-Regular, "SF Mono", Consolas, "Liberation Mono", Menlo, monospace',
  "--w-rjv-background-color": "transparent",
  "--w-rjv-color": "hsl(var(--foreground))",
  "--w-rjv-line-color": "hsl(var(--border))",
  "--w-rjv-arrow-color": "hsl(var(--muted-foreground))",
  "--w-rjv-info-color": "hsl(var(--muted-foreground))",
  "--w-rjv-curlybraces-color": "hsl(var(--foreground))",
  "--w-rjv-brackets-color": "hsl(var(--foreground))",
  "--w-rjv-colon-color": "hsl(var(--muted-foreground))",
  "--w-rjv-ellipsis-color": "hsl(var(--muted-foreground))",
  "--w-rjv-key-string": "hsl(199 89% 58%)",
  "--w-rjv-type-string-color": "hsl(142 76% 45%)",
  "--w-rjv-type-int-color": "hsl(262 83% 70%)",
  "--w-rjv-type-float-color": "hsl(262 83% 70%)",
  "--w-rjv-type-boolean-color": "hsl(38 92% 50%)",
  "--w-rjv-type-null-color": "hsl(347 77% 50%)",
} as CSSProperties

export function JsonTreeViewer({
  value,
  emptyText = "暂无数据",
  className,
  collapsed = 2,
}: JsonTreeViewerProps) {
  if (value == null || value === "") {
    return (
      <div
        className={cn(
          "rounded-md border bg-muted/20 px-4 py-3 font-mono text-xs text-muted-foreground",
          className
        )}
      >
        {emptyText}
      </div>
    )
  }

  return (
    <div className={cn("rounded-md border bg-muted/20 p-3", className)}>
      <JsonView
        value={value as Record<string, unknown>}
        collapsed={collapsed}
        displayDataTypes={false}
        displayObjectSize
        enableClipboard
        shortenTextAfterLength={120}
        style={viewerTheme}
      />
    </div>
  )
}
