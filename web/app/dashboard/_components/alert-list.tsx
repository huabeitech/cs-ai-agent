"use client"

import Link from "next/link"
import { ArrowRightIcon, Link2Icon, ShieldAlertIcon } from "lucide-react"

import type { DashboardAlert, DashboardQuickLink } from "@/lib/api/dashboard"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"

type AlertListProps = {
  alerts: DashboardAlert[]
  quickLinks: DashboardQuickLink[]
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

export function AlertList({ alerts, quickLinks }: AlertListProps) {
  return (
    <div className="grid gap-4 xl:grid-cols-[1.2fr_0.8fr]">
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

      <Card>
        <CardHeader>
          <CardTitle>快捷入口</CardTitle>
          <CardDescription>常用后台模块的直接跳转入口</CardDescription>
        </CardHeader>
        <CardContent className="grid gap-3">
          {quickLinks.map((item) => (
            <Link key={item.link} href={item.link}>
              <div className="rounded-2xl border p-4 transition-colors hover:border-primary/40">
                <div className="flex items-start gap-3">
                  <div className="rounded-full bg-primary/10 p-2 text-primary">
                    <Link2Icon className="size-4" />
                  </div>
                  <div className="min-w-0">
                    <div className="font-medium">{item.title}</div>
                    <div className="mt-1 text-sm text-muted-foreground">
                      {item.description}
                    </div>
                  </div>
                </div>
              </div>
            </Link>
          ))}
          <Link href="/dashboard/conversations">
            <Button variant="outline" className="w-full justify-between">
              进入会话管理
              <ArrowRightIcon className="size-4" />
            </Button>
          </Link>
        </CardContent>
      </Card>
    </div>
  )
}
