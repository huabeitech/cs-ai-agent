"use client"

import { Badge } from "@/components/ui/badge"

const statusLabelMap: Record<string, string> = {
  new: "新建",
  open: "处理中",
  pending_customer: "待客户反馈",
  pending_internal: "待内部处理",
  resolved: "已解决",
  closed: "已关闭",
  cancelled: "已取消",
}

const statusClassNameMap: Record<string, string> = {
  new: "bg-sky-500/10 text-sky-700 border-sky-500/20",
  open: "bg-emerald-500/10 text-emerald-700 border-emerald-500/20",
  pending_customer: "bg-amber-500/10 text-amber-700 border-amber-500/20",
  pending_internal: "bg-orange-500/10 text-orange-700 border-orange-500/20",
  resolved: "bg-lime-500/10 text-lime-700 border-lime-500/20",
  closed: "bg-muted text-muted-foreground border-border",
  cancelled: "bg-rose-500/10 text-rose-700 border-rose-500/20",
}

export function ticketStatusLabel(status: string) {
  return statusLabelMap[status] ?? status
}

export function TicketStatusBadge({ status }: { status: string }) {
  return (
    <Badge variant="outline" className={statusClassNameMap[status] ?? statusClassNameMap.closed}>
      {ticketStatusLabel(status)}
    </Badge>
  )
}
