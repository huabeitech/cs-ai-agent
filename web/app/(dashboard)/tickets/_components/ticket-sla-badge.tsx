"use client"

import { Badge } from "@/components/ui/badge"
import type { TicketItem } from "@/lib/api/ticket"

function isClosedStatus(status: string) {
  return status === "resolved" || status === "closed" || status === "cancelled"
}

export function TicketSLABadge({ ticket }: { ticket: TicketItem }) {
  if (isClosedStatus(ticket.status)) {
    return (
      <Badge variant="outline" className="border-border bg-muted text-muted-foreground">
        已结束
      </Badge>
    )
  }
  if (!ticket.resolveDeadlineAt) {
    return (
      <Badge variant="outline" className="border-border bg-muted text-muted-foreground">
        未设置
      </Badge>
    )
  }

  const deadline = new Date(ticket.resolveDeadlineAt.replace(" ", "T"))
  if (Number.isNaN(deadline.getTime())) {
    return (
      <Badge variant="outline" className="border-border bg-muted text-muted-foreground">
        未设置
      </Badge>
    )
  }

  const remainingMinutes = Math.floor((deadline.getTime() - Date.now()) / 60000)
  if (remainingMinutes < 0) {
    return (
      <Badge variant="outline" className="border-red-500/20 bg-red-500/10 text-red-700">
        已超时
      </Badge>
    )
  }
  if (remainingMinutes <= 60) {
    return (
      <Badge variant="outline" className="border-red-500/20 bg-red-500/10 text-red-700">
        1 小时内
      </Badge>
    )
  }
  if (remainingMinutes <= 240) {
    return (
      <Badge variant="outline" className="border-amber-500/20 bg-amber-500/10 text-amber-700">
        今日风险
      </Badge>
    )
  }
  return (
    <Badge variant="outline" className="border-emerald-500/20 bg-emerald-500/10 text-emerald-700">
      正常
    </Badge>
  )
}
