"use client"

import Link from "next/link"
import { AlertTriangleIcon, CircleDashedIcon, RefreshCcwIcon, TimerResetIcon, WrenchIcon } from "lucide-react"
import { useCallback, useEffect, useMemo, useState } from "react"
import { toast } from "sonner"

import { OptionCombobox } from "@/components/option-combobox"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import {
  fetchAgentTeamsAll,
  type AdminAgentTeam,
} from "@/lib/api/admin"
import {
  fetchTicketRiskOverview,
  fetchTicketRiskList,
  type TicketItem,
  type TicketRiskOverview,
} from "@/lib/api/ticket"
import { formatDateTime } from "@/lib/utils"
import { TicketPriorityBadge } from "../tickets/_components/ticket-priority-badge"
import { TicketSLABadge } from "../tickets/_components/ticket-sla-badge"
import { TicketStatusBadge } from "../tickets/_components/ticket-status-badge"

type RiskTableProps = {
  title: string
  description: string
  items: TicketItem[]
  emptyText: string
}

function RiskTable({ title, description, items, emptyText }: RiskTableProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">{title}</CardTitle>
        <CardDescription>{description}</CardDescription>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>工单</TableHead>
              <TableHead>分类</TableHead>
              <TableHead>优先级</TableHead>
              <TableHead>状态</TableHead>
              <TableHead>SLA</TableHead>
              <TableHead>处理人</TableHead>
              <TableHead>更新时间</TableHead>
              <TableHead className="text-right">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {items.length > 0 ? (
              items.map((item) => (
                <TableRow key={item.id}>
                  <TableCell className="min-w-64">
                    <div className="space-y-1">
                      <div className="font-medium">{item.title}</div>
                      <div className="text-xs text-muted-foreground">{item.ticketNo}</div>
                    </div>
                  </TableCell>
                  <TableCell>
                    {item.tags && item.tags.length > 0
                      ? item.tags.map((tag) => tag.name).join(" / ")
                      : "未打标签"}
                  </TableCell>
                  <TableCell>
                    <TicketPriorityBadge priority={item.priority} priorityName={item.priorityName} />
                  </TableCell>
                  <TableCell>
                    <TicketStatusBadge status={item.status} />
                  </TableCell>
                  <TableCell>
                    <TicketSLABadge ticket={item} />
                  </TableCell>
                  <TableCell>{item.currentAssigneeName || "未指派"}</TableCell>
                  <TableCell>{item.updatedAt ? formatDateTime(item.updatedAt) : "—"}</TableCell>
                  <TableCell className="text-right">
                    <Link href={`/tickets/detail?id=${item.id}`} target="_blank" rel="noreferrer">
                      <Button variant="outline" size="sm">
                        查看详情
                      </Button>
                    </Link>
                  </TableCell>
                </TableRow>
              ))
            ) : (
              <TableRow>
                <TableCell colSpan={8} className="h-24 text-center text-muted-foreground">
                  {emptyText}
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  )
}

export default function TicketRiskPage() {
  const [loading, setLoading] = useState(true)
  const [overview, setOverview] = useState<TicketRiskOverview | null>(null)
  const [overdueTickets, setOverdueTickets] = useState<TicketItem[]>([])
  const [highRiskTickets, setHighRiskTickets] = useState<TicketItem[]>([])
  const [unassignedTickets, setUnassignedTickets] = useState<TicketItem[]>([])
  const [pendingInternalTickets, setPendingInternalTickets] = useState<TicketItem[]>([])
  const [teams, setTeams] = useState<AdminAgentTeam[]>([])
  const [teamFilter, setTeamFilter] = useState("all")
  const [riskWindow, setRiskWindow] = useState("240")

  useEffect(() => {
    void (async () => {
      try {
        const data = await fetchAgentTeamsAll()
        setTeams(Array.isArray(data) ? data : [])
      } catch {
        setTeams([])
      }
    })()
  }, [])

  const loadData = useCallback(async () => {
    setLoading(true)
    try {
      const currentTeamId = teamFilter === "all" ? undefined : Number(teamFilter)
      const riskMinutes = Number(riskWindow)
      const [overviewData, overdueData, highRiskData, unassignedData, pendingInternalData] =
        await Promise.all([
          fetchTicketRiskOverview({ currentTeamId, riskWindowMins: riskMinutes }),
          fetchTicketRiskList({ riskType: "overdue", currentTeamId, riskWindowMins: riskMinutes, page: 1, limit: 10 }),
          fetchTicketRiskList({ riskType: "high_risk", currentTeamId, riskWindowMins: riskMinutes, page: 1, limit: 10 }),
          fetchTicketRiskList({ riskType: "unassigned", currentTeamId, riskWindowMins: riskMinutes, page: 1, limit: 10 }),
          fetchTicketRiskList({ riskType: "pending_internal", currentTeamId, riskWindowMins: riskMinutes, page: 1, limit: 10 }),
        ])

      setOverview(overviewData)
      setOverdueTickets(Array.isArray(overdueData.results) ? overdueData.results : [])
      setHighRiskTickets(Array.isArray(highRiskData.results) ? highRiskData.results : [])
      setUnassignedTickets(Array.isArray(unassignedData.results) ? unassignedData.results : [])
      setPendingInternalTickets(
        Array.isArray(pendingInternalData.results) ? pendingInternalData.results : [],
      )
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "加载 SLA 风险页失败")
    } finally {
      setLoading(false)
    }
  }, [riskWindow, teamFilter])

  useEffect(() => {
    void loadData()
  }, [loadData])

  const cards = useMemo(
    () => [
      {
        title: "已超时",
        description: "解决 SLA 已经 breach 的工单",
        value: overview?.overdue ?? 0,
        icon: AlertTriangleIcon,
        tone: "text-red-700 bg-red-500/10",
      },
      {
        title: `${Number(riskWindow) / 60} 小时内到期`,
        description: "建议组长优先盯防的风险队列",
        value: overview?.highRisk ?? 0,
        icon: TimerResetIcon,
        tone: "text-orange-700 bg-orange-500/10",
      },
      {
        title: "待分配",
        description: "目前还没有明确负责人的工单",
        value: overview?.unassigned ?? 0,
        icon: CircleDashedIcon,
        tone: "text-amber-700 bg-amber-500/10",
      },
      {
        title: "待内部处理",
        description: "等待内部团队协作处理的工单",
        value: overview?.pendingInternal ?? 0,
        icon: WrenchIcon,
        tone: "text-blue-700 bg-blue-500/10",
      },
    ],
    [overview, riskWindow],
  )

  return (
    <div className="min-h-0 flex-1 overflow-auto bg-muted/20 p-4 md:p-6">
      <div className="flex flex-col gap-4">
        <div className="flex flex-col gap-3 xl:flex-row xl:items-center xl:justify-between">
          <div>
            <h1 className="text-xl font-semibold">SLA 风险运营</h1>
            <p className="mt-1 text-sm text-muted-foreground">
              给主管和组长使用的风险盯防页，优先查看超时、临近超时和待分配工单
            </p>
          </div>
          <div className="flex gap-2">
            <div className="w-44">
              <OptionCombobox
                value={teamFilter}
                onChange={setTeamFilter}
                placeholder="全部团队"
                options={[
                  { value: "all", label: "全部团队" },
                  ...teams.map((team) => ({ value: String(team.id), label: team.name })),
                ]}
              />
            </div>
            <div className="w-44">
              <OptionCombobox
                value={riskWindow}
                onChange={setRiskWindow}
                placeholder="风险时间窗"
                options={[
                  { value: "60", label: "1 小时内" },
                  { value: "240", label: "4 小时内" },
                  { value: "1440", label: "24 小时内" },
                ]}
              />
            </div>
            <Link href="/tickets">
              <Button variant="outline">前往工单工作台</Button>
            </Link>
            <Button variant="outline" onClick={() => void loadData()} disabled={loading}>
              <RefreshCcwIcon className="size-4" />
              刷新
            </Button>
          </div>
        </div>

        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          {cards.map((item) => {
            const Icon = item.icon
            return (
              <Card key={item.title}>
                <CardHeader className="flex flex-row items-start justify-between space-y-0 pb-3">
                  <div className="space-y-1">
                    <CardTitle className="text-sm font-medium">{item.title}</CardTitle>
                    <CardDescription>{item.description}</CardDescription>
                  </div>
                  <div className={`rounded-full p-2 ${item.tone}`}>
                    <Icon className="size-4" />
                  </div>
                </CardHeader>
                <CardContent>
                  <div className="text-3xl font-semibold tracking-tight">
                    {loading ? "..." : item.value.toLocaleString()}
                  </div>
                </CardContent>
              </Card>
            )
          })}
        </div>

        <RiskTable
          title="已超时工单"
          description="需要立即处理或升级的高风险工单"
          items={overdueTickets}
          emptyText="当前没有已超时工单"
        />

        <RiskTable
          title="4 小时内到期"
          description="建议优先处理，避免进入超时队列"
          items={highRiskTickets}
          emptyText="当前没有临近超时工单"
        />

        <Card>
          <CardHeader>
            <CardTitle className="text-base">滞留原因</CardTitle>
            <CardDescription>帮助主管快速判断风险是由分配、协作还是 SLA 配置问题造成</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
              {(overview?.reasons?.length ?? 0) > 0 ? (
                overview?.reasons?.map((item) => (
                  <div key={item.code} className="rounded-lg border bg-muted/20 p-4">
                    <div className="text-sm font-medium">{item.title}</div>
                    <div className="mt-2 text-2xl font-semibold">{item.count.toLocaleString()}</div>
                    <div className="mt-2 text-xs leading-6 text-muted-foreground">{item.description}</div>
                  </div>
                ))
              ) : (
                <div className="text-sm text-muted-foreground">暂无滞留原因数据</div>
              )}
            </div>
          </CardContent>
        </Card>

        <div className="grid gap-4 xl:grid-cols-2">
          <RiskTable
            title="待分配工单"
            description="进入队列但尚未明确负责人的工单"
            items={unassignedTickets}
            emptyText="当前没有待分配工单"
          />
          <RiskTable
            title="待内部处理"
            description="需要内部团队介入，容易长期滞留的工单"
            items={pendingInternalTickets}
            emptyText="当前没有待内部处理工单"
          />
        </div>
      </div>
    </div>
  )
}
