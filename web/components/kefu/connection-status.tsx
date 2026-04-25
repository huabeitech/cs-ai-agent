"use client"

import { Badge } from "@/components/ui/badge"
import { cn } from "@/lib/utils"

type KefuConnectionStatusProps = {
  status: "connecting" | "connected" | "disconnected"
}

const statusText: Record<KefuConnectionStatusProps["status"], string> = {
  connecting: "连接中",
  connected: "在线服务",
  disconnected: "连接已断开",
}

export function KefuConnectionStatus({ status }: KefuConnectionStatusProps) {
  const toneClass =
    status === "connected"
      ? "border-emerald-200 bg-emerald-50 text-emerald-700"
      : status === "connecting"
        ? "border-amber-200 bg-amber-50 text-amber-700"
        : "border-slate-200 bg-slate-100 text-slate-600"

  return (
    <Badge
      variant="outline"
      className={cn("h-6 gap-2 px-2.5 text-[11px] font-medium shadow-sm", toneClass)}
    >
      <span
        className={cn(
          "inline-block size-2 rounded-full",
          status === "connected"
            ? "bg-emerald-500 shadow-[0_0_0_4px_rgba(16,185,129,0.14)]"
            : status === "connecting"
              ? "bg-amber-500 shadow-[0_0_0_4px_rgba(245,158,11,0.16)]"
              : "bg-slate-400 shadow-[0_0_0_4px_rgba(148,163,184,0.14)]"
        )}
      />
      <span>{statusText[status]}</span>
    </Badge>
  )
}
