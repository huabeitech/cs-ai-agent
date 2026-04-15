"use client"

import { Area, AreaChart, Bar, BarChart, CartesianGrid, XAxis, YAxis } from "recharts"

import type { DashboardStatusDistributionItem, DashboardTrendItem } from "@/lib/api/dashboard"
import {
  Card,
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

type TrendPanelProps = {
  title: string
  description: string
  trend: DashboardTrendItem[]
  distribution: DashboardStatusDistributionItem[]
}

const trendConfig = {
  newCount: {
    label: "新增",
    color: "hsl(24 95% 53%)",
  },
  closedCount: {
    label: "关闭",
    color: "hsl(190 95% 39%)",
  },
} satisfies ChartConfig

const distributionConfig = {
  count: {
    label: "数量",
    theme: {
      light: "hsl(222 47% 11%)",
      dark: "hsl(210 40% 98%)",
    },
  },
} satisfies ChartConfig

export function TrendPanel({
  title,
  description,
  trend,
  distribution,
}: TrendPanelProps) {
  return (
    <div className="grid gap-4 xl:grid-cols-[1.5fr_0.9fr]">
      <Card>
        <CardHeader>
          <CardTitle>{title}</CardTitle>
          <CardDescription>{description}</CardDescription>
        </CardHeader>
        <CardContent>
          <ChartContainer config={trendConfig} className="h-72 w-full">
            <AreaChart data={trend}>
              <CartesianGrid vertical={false} />
              <XAxis dataKey="date" tickLine={false} axisLine={false} tickMargin={8} />
              <ChartTooltip content={<ChartTooltipContent />} />
              <Area
                type="monotone"
                dataKey="newCount"
                stroke="var(--color-newCount)"
                fill="var(--color-newCount)"
                fillOpacity={0.18}
                strokeWidth={2}
              />
              <Area
                type="monotone"
                dataKey="closedCount"
                stroke="var(--color-closedCount)"
                fill="var(--color-closedCount)"
                fillOpacity={0.08}
                strokeWidth={2}
              />
            </AreaChart>
          </ChartContainer>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>状态分布</CardTitle>
          <CardDescription>当前状态数量分布</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <ChartContainer config={distributionConfig} className="h-72 w-full">
            <BarChart data={distribution} layout="vertical" margin={{ left: 20 }}>
              <CartesianGrid horizontal={false} />
              <YAxis
                type="category"
                dataKey="label"
                tickLine={false}
                axisLine={false}
                width={64}
              />
              <XAxis type="number" hide />
              <ChartTooltip content={<ChartTooltipContent hideLabel />} />
              <Bar
                dataKey="count"
                fill="var(--color-count)"
                radius={[0, 8, 8, 0]}
              />
            </BarChart>
          </ChartContainer>
        </CardContent>
      </Card>
    </div>
  )
}
