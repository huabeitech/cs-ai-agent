"use client"

import { Badge } from "@/components/ui/badge"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import type { DashboardOverview } from "@/lib/api/dashboard"

type TeamLoadPanelProps = {
  agentStats: DashboardOverview["agentStats"]
}

function getLoadTone(loadRate: number) {
  if (loadRate >= 85) {
    return "bg-red-500"
  }
  if (loadRate >= 60) {
    return "bg-amber-500"
  }
  return "bg-emerald-500"
}

export function TeamLoadPanel({ agentStats }: TeamLoadPanelProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>客服组负载</CardTitle>
        <CardDescription>
          在线 {agentStats.onlineAgents}，忙碌 {agentStats.busyAgents}，离线 {agentStats.offlineAgents}
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {agentStats.teamLoads.length === 0 ? (
          <div className="rounded-xl border border-dashed px-4 py-8 text-center text-sm text-muted-foreground">
            暂无客服组数据
          </div>
        ) : (
          agentStats.teamLoads.map((item) => (
            <div key={item.teamId} className="rounded-2xl border p-4">
              <div className="flex flex-wrap items-center justify-between gap-3">
                <div>
                  <div className="flex items-center gap-2">
                    <div className="font-medium">{item.teamName}</div>
                    {item.hasScheduleNow ? (
                      <Badge variant="secondary">排班中</Badge>
                    ) : (
                      <Badge variant="outline">无当前排班</Badge>
                    )}
                  </div>
                  <div className="mt-1 text-sm text-muted-foreground">
                    总客服 {item.totalAgents}，在线 {item.onlineAgents}，忙碌 {item.busyAgents}，离线{" "}
                    {item.offlineAgents}
                  </div>
                </div>
                <div className="text-right">
                  <div className="text-2xl font-semibold">{item.loadRate.toFixed(1)}%</div>
                  <div className="text-xs text-muted-foreground">
                    负载 {item.processingConversations}/{item.maxConcurrentCapacity || 0}
                  </div>
                </div>
              </div>

              <div className="mt-4 h-2 rounded-full bg-muted">
                <div
                  className={`h-2 rounded-full ${getLoadTone(item.loadRate)}`}
                  style={{ width: `${Math.min(item.loadRate, 100)}%` }}
                />
              </div>

              <div className="mt-4 grid gap-3 text-sm md:grid-cols-2 xl:grid-cols-4">
                <div className="rounded-xl bg-muted/40 px-3 py-2">
                  <div className="text-muted-foreground">待接入</div>
                  <div className="mt-1 text-lg font-semibold">{item.waitingConversations}</div>
                </div>
                <div className="rounded-xl bg-muted/40 px-3 py-2">
                  <div className="text-muted-foreground">处理中</div>
                  <div className="mt-1 text-lg font-semibold">{item.processingConversations}</div>
                </div>
                <div className="rounded-xl bg-muted/40 px-3 py-2">
                  <div className="text-muted-foreground">并发容量</div>
                  <div className="mt-1 text-lg font-semibold">{item.maxConcurrentCapacity}</div>
                </div>
                <div className="rounded-xl bg-muted/40 px-3 py-2">
                  <div className="text-muted-foreground">忙碌客服</div>
                  <div className="mt-1 text-lg font-semibold">{item.busyAgents}</div>
                </div>
              </div>
            </div>
          ))
        )}
      </CardContent>
    </Card>
  )
}
