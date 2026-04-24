"use client"

import * as React from "react"
import { Area, AreaChart, CartesianGrid, XAxis } from "recharts"

import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import {
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
  type ChartConfig,
} from "@/components/ui/chart"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import {
  ToggleGroup,
  ToggleGroupItem,
} from "@/components/ui/toggle-group"

const chartData = [
  { date: "2026-03-01", activeSessions: 18, indexedDocs: 12 },
  { date: "2026-03-02", activeSessions: 22, indexedDocs: 15 },
  { date: "2026-03-03", activeSessions: 21, indexedDocs: 16 },
  { date: "2026-03-04", activeSessions: 28, indexedDocs: 20 },
  { date: "2026-03-05", activeSessions: 32, indexedDocs: 24 },
  { date: "2026-03-06", activeSessions: 31, indexedDocs: 26 },
  { date: "2026-03-07", activeSessions: 36, indexedDocs: 30 },
  { date: "2026-03-08", activeSessions: 34, indexedDocs: 32 },
  { date: "2026-03-09", activeSessions: 39, indexedDocs: 34 },
  { date: "2026-03-10", activeSessions: 41, indexedDocs: 37 },
  { date: "2026-03-11", activeSessions: 43, indexedDocs: 40 },
  { date: "2026-03-12", activeSessions: 46, indexedDocs: 44 },
  { date: "2026-03-13", activeSessions: 44, indexedDocs: 46 },
  { date: "2026-03-14", activeSessions: 49, indexedDocs: 50 },
]

const chartConfig = {
  activeSessions: {
    label: "活跃会话",
    color: "var(--primary)",
  },
  indexedDocs: {
    label: "知识文档",
    color: "var(--chart-2)",
  },
} satisfies ChartConfig

export function ChartAreaInteractive() {
  const [timeRange, setTimeRange] = React.useState("14d")

  const filteredData = chartData.slice(timeRange === "7d" ? -7 : -14)

  return (
    <Card className="@container/card">
      <CardHeader>
        <CardTitle>近期开发活跃度</CardTitle>
        <CardDescription>
          以演示数据展示会话增长与知识库准备进度
        </CardDescription>
        <CardAction>
          <ToggleGroup
            multiple={false}
            value={timeRange ? [timeRange] : []}
            onValueChange={(value) => {
              setTimeRange(value[0] ?? "14d")
            }}
            variant="outline"
            className="hidden *:data-[slot=toggle-group-item]:px-4! @[767px]/card:flex"
          >
            <ToggleGroupItem value="14d">近 14 天</ToggleGroupItem>
            <ToggleGroupItem value="7d">近 7 天</ToggleGroupItem>
          </ToggleGroup>
          <Select
            value={timeRange}
            onValueChange={(value) => {
              if (value) {
                setTimeRange(value)
              }
            }}
          >
            <SelectTrigger
              className="flex w-32 @[767px]/card:hidden"
              size="sm"
              aria-label="选择时间范围"
            >
              <SelectValue placeholder="近 14 天" />
            </SelectTrigger>
            <SelectContent className="rounded-xl">
              <SelectItem value="14d" className="rounded-lg">
                近 14 天
              </SelectItem>
              <SelectItem value="7d" className="rounded-lg">
                近 7 天
              </SelectItem>
            </SelectContent>
          </Select>
        </CardAction>
      </CardHeader>
      <CardContent className="px-2 pt-4 sm:px-6 sm:pt-6">
        <ChartContainer
          config={chartConfig}
          className="aspect-auto h-[250px] w-full"
        >
          <AreaChart data={filteredData}>
            <defs>
              <linearGradient id="fillSessions" x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor="var(--color-activeSessions)" stopOpacity={0.9} />
                <stop offset="95%" stopColor="var(--color-activeSessions)" stopOpacity={0.1} />
              </linearGradient>
              <linearGradient id="fillDocs" x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor="var(--color-indexedDocs)" stopOpacity={0.7} />
                <stop offset="95%" stopColor="var(--color-indexedDocs)" stopOpacity={0.08} />
              </linearGradient>
            </defs>
            <CartesianGrid vertical={false} />
            <XAxis
              dataKey="date"
              tickLine={false}
              axisLine={false}
              tickMargin={8}
              tickFormatter={(value) => value.slice(5)}
            />
            <ChartTooltip
              cursor={false}
              content={<ChartTooltipContent indicator="dot" />}
            />
            <Area
              dataKey="activeSessions"
              type="natural"
              fill="url(#fillSessions)"
              stroke="var(--color-activeSessions)"
              stackId="a"
            />
            <Area
              dataKey="indexedDocs"
              type="natural"
              fill="url(#fillDocs)"
              stroke="var(--color-indexedDocs)"
              stackId="b"
            />
          </AreaChart>
        </ChartContainer>
      </CardContent>
    </Card>
  )
}
