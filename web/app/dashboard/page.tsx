"use client"

import { useCallback, useEffect, useState } from "react"
import { RefreshCwIcon } from "lucide-react"
import { toast } from "sonner"

import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { Card, CardContent } from "@/components/ui/card"
import {
  fetchDashboardOverview,
  type DashboardOverview,
  type DashboardRange,
} from "@/lib/api/dashboard"
import { SummaryCards } from "./_components/summary-cards"
import { TrendPanel } from "./_components/trend-panel"
import { TeamLoadPanel } from "./_components/team-load-panel"
import { AlertList } from "./_components/alert-list"

const rangeOptions: Array<{ value: DashboardRange; label: string }> = [
  { value: "today", label: "今天" },
  { value: "7d", label: "近 7 天" },
  { value: "30d", label: "近 30 天" },
]

function LoadingCards() {
  return (
    <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-6">
      {Array.from({ length: 6 }).map((_, index) => (
        <Card key={index}>
          <CardContent className="space-y-3 p-6">
            <Skeleton className="h-4 w-24" />
            <Skeleton className="h-8 w-20" />
            <Skeleton className="h-4 w-full" />
          </CardContent>
        </Card>
      ))}
    </div>
  )
}

export default function DashboardPage() {
  const [range, setRange] = useState<DashboardRange>("7d")
  const [data, setData] = useState<DashboardOverview | null>(null)
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)

  const loadData = useCallback(
    async (nextRange: DashboardRange, showRefreshing = false) => {
      if (showRefreshing) {
        setRefreshing(true)
      } else {
        setLoading(true)
      }
      try {
        const result = await fetchDashboardOverview(nextRange)
        setData(result)
      } catch (error) {
        toast.error(error instanceof Error ? error.message : "加载首页概览失败")
      } finally {
        setLoading(false)
        setRefreshing(false)
      }
    },
    []
  )

  useEffect(() => {
    void loadData(range)
  }, [loadData, range])

  return (
    <div className="flex flex-1 flex-col gap-6 p-4 lg:p-6">
      <div className="flex flex-col gap-3 xl:flex-row xl:items-start xl:justify-between">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">后台总览</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            聚焦会话接待、工单处理、客服负载与 AI 运行状态
            {data ? `，更新于 ${data.generatedAt}` : ""}
          </p>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <div className="rounded-xl border bg-muted/30 p-1">
            {rangeOptions.map((item) => (
              <Button
                key={item.value}
                variant={range === item.value ? "secondary" : "ghost"}
                size="sm"
                onClick={() => setRange(item.value)}
              >
                {item.label}
              </Button>
            ))}
          </div>
          <Button
            variant="outline"
            size="sm"
            onClick={() => void loadData(range, true)}
            disabled={loading || refreshing}
          >
            <RefreshCwIcon className={refreshing ? "mr-2 size-4 animate-spin" : "mr-2 size-4"} />
            刷新
          </Button>
        </div>
      </div>

      {loading && !data ? (
        <LoadingCards />
      ) : data ? (
        <>
          <SummaryCards summary={data.summary} />

          <TrendPanel
            title="会话趋势"
            description="观察新增与关闭会话变化，快速判断接待压力是否持续上升"
            trend={data.conversationStats.trend}
            distribution={data.conversationStats.statusDistribution}
          />

          <div className="grid gap-4 xl:grid-cols-[1.15fr_0.85fr]">
            <TeamLoadPanel agentStats={data.agentStats} />

            <Card>
              <CardContent className="grid gap-4 p-6 sm:grid-cols-2">
                <div className="rounded-2xl border bg-muted/30 p-4">
                  <div className="text-sm text-muted-foreground">启用中的 AI Agent</div>
                  <div className="mt-2 text-3xl font-semibold">{data.aiStats.enabledAiAgents}</div>
                </div>
                <div className="rounded-2xl border bg-muted/30 p-4">
                  <div className="text-sm text-muted-foreground">启用中的接入站点</div>
                  <div className="mt-2 text-3xl font-semibold">{data.aiStats.enabledWidgetSites}</div>
                </div>
                <div className="rounded-2xl border bg-muted/30 p-4">
                  <div className="text-sm text-muted-foreground">今日知识检索次数</div>
                  <div className="mt-2 text-3xl font-semibold">
                    {data.aiStats.todayKnowledgeRetrieves}
                  </div>
                </div>
                <div className="rounded-2xl border bg-muted/30 p-4">
                  <div className="text-sm text-muted-foreground">今日检索失败率</div>
                  <div className="mt-2 text-3xl font-semibold">
                    {data.aiStats.todayKnowledgeRetrieveFailRate.toFixed(1)}%
                  </div>
                </div>
                <div className="rounded-2xl border bg-muted/30 p-4">
                  <div className="text-sm text-muted-foreground">今日 Skill 失败次数</div>
                  <div className="mt-2 text-3xl font-semibold">
                    {data.aiStats.todaySkillRunFailCount}
                  </div>
                </div>
                <div className="rounded-2xl border bg-muted/30 p-4">
                  <div className="text-sm text-muted-foreground">今日 AI 转人工次数</div>
                  <div className="mt-2 text-3xl font-semibold">
                    {data.aiStats.todayAiHandoffCount}
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>

          <AlertList alerts={data.alerts} />
        </>
      ) : (
        <Card>
          <CardContent className="flex min-h-60 items-center justify-center p-6 text-sm text-muted-foreground">
            暂无首页概览数据
          </CardContent>
        </Card>
      )}
    </div>
  )
}
