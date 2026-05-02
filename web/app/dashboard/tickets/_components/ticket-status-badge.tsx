"use client"

import { Badge } from "@/components/ui/badge"
import type { TicketStatus } from "@/lib/api/ticket"

const statusMap = {
  pending: { label: "待处理", className: "border-amber-200 bg-amber-50 text-amber-700" },
  in_progress: { label: "处理中", className: "border-blue-200 bg-blue-50 text-blue-700" },
  done: { label: "已处理", className: "border-emerald-200 bg-emerald-50 text-emerald-700" },
} as const

export function ticketStatusLabel(status: string) {
  return statusMap[status as TicketStatus]?.label ?? status
}

export function TicketStatusBadge({ status }: { status: string }) {
  const option = statusMap[status as TicketStatus]

  return (
    <Badge variant="outline" className={option?.className ?? "border-border bg-muted text-muted-foreground"}>
      {option?.label ?? status}
    </Badge>
  )
}
