"use client"

import Link from "next/link"
import {
  BotMessageSquareIcon,
  CircleDashedIcon,
  HeadsetIcon,
  SparklesIcon,
  WavesIcon,
} from "lucide-react"

import type { DashboardOverview } from "@/lib/api/dashboard"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"

type SummaryCardsProps = {
  summary: DashboardOverview["summary"]
}

type SummaryCardItem = {
  key: keyof DashboardOverview["summary"]
  title: string
  description: string
  link: string
  icon: typeof BotMessageSquareIcon
  format?: (value: number) => string
}

const cards: SummaryCardItem[] = [
  {
    key: "todayNewConversations",
    title: "今日新增会话",
    description: "今日进入系统的新增咨询量",
    link: "/conversations",
    icon: BotMessageSquareIcon,
  },
  {
    key: "processingConversations",
    title: "当前处理中",
    description: "正在由 AI 或人工接待的会话",
    link: "/conversations",
    icon: WavesIcon,
  },
  {
    key: "pendingDispatchConversations",
    title: "待分配会话",
    description: "仍在待接入池中等待分配",
    link: "/conversations",
    icon: CircleDashedIcon,
  },
  {
    key: "onlineAgents",
    title: "在线客服",
    description: "近 15 分钟内仍有活跃心跳的客服",
    link: "/agents",
    icon: HeadsetIcon,
  },
  {
    key: "aiServiceRate",
    title: "AI 接待占比",
    description: "当前活跃会话中 AI 参与服务比例",
    link: "/ai-agents",
    icon: SparklesIcon,
    format: (value: number) => `${value.toFixed(1)}%`,
  },
]

export function SummaryCards({ summary }: SummaryCardsProps) {
  return (
    <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-6">
      {cards.map((item) => {
        const Icon = item.icon
        const rawValue = summary[item.key]
        const value =
          typeof item.format === "function"
            ? item.format(Number(rawValue))
            : Number(rawValue).toLocaleString()

        return (
          <Link key={item.key} href={item.link}>
            <Card className="h-full transition-colors hover:border-primary/40">
              <CardHeader className="flex flex-row items-start justify-between space-y-0 pb-3">
                <div className="space-y-1">
                  <CardTitle className="text-sm font-medium">{item.title}</CardTitle>
                  <CardDescription>{item.description}</CardDescription>
                </div>
                <div className="rounded-full bg-primary/10 p-2 text-primary">
                  <Icon className="size-4" />
                </div>
              </CardHeader>
              <CardContent>
                <div className="text-3xl font-semibold tracking-tight">{value}</div>
              </CardContent>
            </Card>
          </Link>
        )
      })}
    </div>
  )
}
