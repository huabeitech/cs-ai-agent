"use client"

import Link from "next/link"
import { ArrowRightIcon, ShieldAlertIcon } from "lucide-react"

import type { DashboardAlert } from "@/lib/api/dashboard"
import { Badge } from "@/components/ui/badge"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"

type AlertListProps = {
  alerts: DashboardAlert[]
}

function getAlertBadgeVariant(level: DashboardAlert["level"]) {
  if (level === "error") {
    return "destructive" as const
  }
  if (level === "warning") {
    return "secondary" as const
  }
  return "outline" as const
}

export function AlertList({ alerts }: AlertListProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>风险提醒</CardTitle>
        <CardDescription>优先处理会直接影响接待效率和 AI 稳定性的项目</CardDescription>
      </CardHeader>
      <CardContent className="space-y-3">
        {alerts.length === 0 ? (
          <div className="rounded-2xl border border-dashed px-4 py-10 text-center">
            <ShieldAlertIcon className="mx-auto mb-3 size-8 text-muted-foreground" />
            <div className="text-sm font-medium">当前没有需要优先处理的风险项</div>
            <div className="mt-1 text-sm text-muted-foreground">
              首页将持续监控会话堆积、客服排班与 AI 配置异常
            </div>
          </div>
        ) : (
          alerts.map((item) => (
            <Link key={item.id} href={item.link} className="block">
              <div className="rounded-2xl border p-4 transition-colors hover:border-primary/40">
                <div className="flex items-start justify-between gap-3">
                  <div>
                    <div className="flex items-center gap-2">
                      <div className="font-medium">{item.title}</div>
                      <Badge variant={getAlertBadgeVariant(item.level)}>
                        {item.count}
                      </Badge>
                    </div>
                    <div className="mt-1 text-sm text-muted-foreground">
                      {item.description}
                    </div>
                  </div>
                  <ArrowRightIcon className="mt-0.5 size-4 text-muted-foreground" />
                </div>
              </div>
            </Link>
          ))
        )}
      </CardContent>
    </Card>
  )
}
