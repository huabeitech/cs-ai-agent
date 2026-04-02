"use client"

import { Badge } from "@/components/ui/badge"

const priorityLabelMap: Record<number, string> = {
  1: "低",
  2: "普通",
  3: "高",
  4: "紧急",
}

const priorityClassNameMap: Record<number, string> = {
  1: "bg-slate-500/10 text-slate-700 border-slate-500/20",
  2: "bg-blue-500/10 text-blue-700 border-blue-500/20",
  3: "bg-orange-500/10 text-orange-700 border-orange-500/20",
  4: "bg-red-500/10 text-red-700 border-red-500/20",
}

export function ticketPriorityLabel(priority: number) {
  return priorityLabelMap[priority] ?? `P${priority}`
}

export function TicketPriorityBadge({ priority }: { priority: number }) {
  return (
    <Badge
      variant="outline"
      className={priorityClassNameMap[priority] ?? priorityClassNameMap[2]}
    >
      {ticketPriorityLabel(priority)}
    </Badge>
  )
}
