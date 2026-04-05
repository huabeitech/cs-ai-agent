"use client"

import { useEffect, useState } from "react"

import { Badge } from "@/components/ui/badge"
import { getTicketPriorityMap } from "@/lib/ticket-priority"

const priorityClassNameMap: Record<number, string> = {
  0: "bg-slate-500/10 text-slate-700 border-slate-500/20",
  1: "bg-blue-500/10 text-blue-700 border-blue-500/20",
  2: "bg-amber-500/10 text-amber-700 border-amber-500/20",
  3: "bg-red-500/10 text-red-700 border-red-500/20",
  4: "bg-fuchsia-500/10 text-fuchsia-700 border-fuchsia-500/20",
}

export function ticketPriorityLabel(priority: number, priorityName?: string) {
  return priorityName?.trim() || `P${priority}`
}

export function TicketPriorityBadge({
  priority,
  priorityName,
}: {
  priority: number
  priorityName?: string
}) {
  const [priorityMap, setPriorityMap] = useState<Record<number, string>>({})

  useEffect(() => {
    void (async () => {
      setPriorityMap(await getTicketPriorityMap())
    })()
  }, [])

  return (
    <Badge
      variant="outline"
      className={priorityClassNameMap[priority] ?? priorityClassNameMap[0]}
    >
      {ticketPriorityLabel(priority, priorityName || priorityMap[priority])}
    </Badge>
  )
}
