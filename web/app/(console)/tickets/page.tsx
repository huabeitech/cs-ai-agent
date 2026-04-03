"use client"

import Link from "next/link"
import { PlusIcon, RefreshCcwIcon, SearchIcon, StarIcon } from "lucide-react"
import { useCallback, useEffect, useState } from "react"
import { toast } from "sonner"

import { ListPagination } from "@/components/list-pagination"
import { OptionCombobox } from "@/components/option-combobox"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import {
  fetchAgentProfilesAll,
  fetchAgentTeamsAll,
  type AdminAgentProfile,
  type AdminAgentTeam,
} from "@/lib/api/admin"
import {
  createTicket,
  fetchTicketSummary,
  fetchTickets,
  updateTicket,
  unwatchTicket,
  watchTicket,
  type Paging,
  type TicketItem,
  type TicketSummary,
} from "@/lib/api/ticket"
import { formatDateTime } from "@/lib/utils"
import { EditDialog } from "./_components/edit"
import { TicketPriorityBadge } from "./_components/ticket-priority-badge"
import { TicketStatusBadge } from "./_components/ticket-status-badge"

const emptyPaging: Paging = { page: 1, limit: 20, total: 0 }
const emptySummary: TicketSummary = {
  all: 0,
  mine: 0,
  watching: 0,
  pendingCustomer: 0,
  dueSoon: 0,
  overdue: 0,
}
type QuickViewKey =
  | "all"
  | "mine"
  | "watching"
  | "pending_customer"
  | "due_soon"
  | "overdue"

export default function TicketsPage() {
  const [result, setResult] = useState<{ results: TicketItem[]; page: Paging }>({
    results: [],
    page: emptyPaging,
  })
  const [summary, setSummary] = useState<TicketSummary>(emptySummary)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editingItemId, setEditingItemId] = useState<number | null>(null)
  const [keywordInput, setKeywordInput] = useState("")
  const [keyword, setKeyword] = useState("")
  const [statusFilter, setStatusFilter] = useState("all")
  const [priorityFilter, setPriorityFilter] = useState("all")
  const [teamFilter, setTeamFilter] = useState("all")
  const [assigneeFilter, setAssigneeFilter] = useState("all")
  const [watchFilter, setWatchFilter] = useState("all")
  const [quickView, setQuickView] = useState<QuickViewKey>("all")
  const [teams, setTeams] = useState<AdminAgentTeam[]>([])
  const [agents, setAgents] = useState<AdminAgentProfile[]>([])
  const [actionLoadingId, setActionLoadingId] = useState<number | null>(null)

  const activeStatusFilter =
    quickView === "pending_customer"
      ? "pending_customer"
      : statusFilter === "all"
        ? undefined
        : statusFilter
  const activeWatchFilter = quickView === "watching" ? 1 : watchFilter === "watching" ? 1 : undefined
  const activeMineFilter = quickView === "mine" ? 1 : undefined
  const activeDueSoonFilter = quickView === "due_soon" ? 1 : undefined
  const activeOverdueFilter = quickView === "overdue" ? 1 : undefined

  const loadSummary = useCallback(async () => {
    try {
      const data = await fetchTicketSummary()
      setSummary(data)
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "加载工单汇总失败")
    }
  }, [])

  const loadTickets = useCallback(async () => {
    setLoading(true)
    try {
      const data = await fetchTickets({
        page: result.page.page,
        limit: result.page.limit,
        keyword: keyword.trim() || undefined,
        status: activeStatusFilter,
        priority: priorityFilter === "all" ? undefined : Number(priorityFilter),
        currentTeamId: teamFilter === "all" ? undefined : Number(teamFilter),
        currentAssigneeId:
          assigneeFilter === "all" ? undefined : Number(assigneeFilter),
        watching: activeWatchFilter,
        mine: activeMineFilter,
        dueSoon: activeDueSoonFilter,
        overdue: activeOverdueFilter,
      })
      setResult({
        results: Array.isArray(data.results) ? data.results : [],
        page: data.page ?? emptyPaging,
      })
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "加载工单失败")
    } finally {
      setLoading(false)
    }
  }, [
    assigneeFilter,
    keyword,
    result.page.limit,
    result.page.page,
    priorityFilter,
    activeStatusFilter,
    activeWatchFilter,
    activeMineFilter,
    activeDueSoonFilter,
    activeOverdueFilter,
    teamFilter,
  ])

  useEffect(() => {
    void loadTickets()
  }, [loadTickets])

  useEffect(() => {
    void loadSummary()
  }, [loadSummary])

  useEffect(() => {
    void (async () => {
      const [teamData, agentData] = await Promise.all([
        fetchAgentTeamsAll(),
        fetchAgentProfilesAll(),
      ])
      setTeams(Array.isArray(teamData) ? teamData : [])
      setAgents(Array.isArray(agentData) ? agentData : [])
    })()
  }, [])

  function applyFilters() {
    setKeyword(keywordInput)
    setResult((current) => ({
      ...current,
      page: { ...current.page, page: 1 },
    }))
  }

  function handleFilterKeyDown(event: React.KeyboardEvent<HTMLInputElement>) {
    if (event.key !== "Enter") {
      return
    }
    event.preventDefault()
    applyFilters()
  }

  function openCreateDialog() {
    setEditingItemId(null)
    setDialogOpen(true)
  }

  function openEditDialog(item: TicketItem) {
    setEditingItemId(item.id)
    setDialogOpen(true)
  }

  function handleDialogOpenChange(open: boolean) {
    if (saving) {
      return
    }
    if (!open) {
      setEditingItemId(null)
    }
    setDialogOpen(open)
  }

  async function handleSubmit(
    payload: Parameters<typeof createTicket>[0] | Parameters<typeof updateTicket>[0],
  ) {
    if (saving) {
      return
    }
    setSaving(true)
    try {
      if ("ticketId" in payload) {
        await updateTicket(payload)
        toast.success("工单已更新")
      } else {
        await createTicket(payload)
        toast.success("工单已创建")
      }
      setDialogOpen(false)
      setEditingItemId(null)
      await loadSummary()
      await loadTickets()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "保存工单失败")
    } finally {
      setSaving(false)
    }
  }

  async function handleWatchToggle(item: TicketItem) {
    setActionLoadingId(item.id)
    try {
      if (item.watchedByMe) {
        await unwatchTicket(item.id)
        toast.success("已取消关注")
      } else {
        await watchTicket(item.id)
        toast.success("已关注工单")
      }
      await loadSummary()
      await loadTickets()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "更新关注状态失败")
    } finally {
      setActionLoadingId(null)
    }
  }

  const teamOptions = [{ value: "all", label: "全部团队" }].concat(
    teams.map((team) => ({ value: String(team.id), label: team.name })),
  )
  const agentOptions = [{ value: "all", label: "全部处理人" }].concat(
    agents.map((agent) => ({
      value: String(agent.userId),
      label:
        agent.displayName ||
        agent.nickname ||
        agent.username ||
        `客服#${agent.userId}`,
    })),
  )
  const quickViews: Array<{
    key: QuickViewKey
    label: string
    description: string
    count: number
  }> = [
    { key: "all", label: "全部工单", description: "工单总量", count: summary.all },
    { key: "mine", label: "我的工单", description: "当前指派给我", count: summary.mine },
    { key: "watching", label: "我的关注", description: "我在跟进的工单", count: summary.watching },
    {
      key: "pending_customer",
      label: "待客户反馈",
      description: "等待客户补充信息",
      count: summary.pendingCustomer,
    },
    { key: "due_soon", label: "即将超时", description: "30 分钟内到期", count: summary.dueSoon },
    { key: "overdue", label: "已超时", description: "解决 SLA 已超时", count: summary.overdue },
  ]

  return (
    <div className="min-h-0 flex-1 overflow-auto bg-muted/20 p-4 md:p-6">
      <div className="flex w-full flex-col gap-4">
        <div className="flex flex-col gap-3 xl:flex-row xl:items-center xl:justify-between">
          <div>
            <h1 className="text-xl font-semibold">工单</h1>
            <p className="mt-1 text-sm text-muted-foreground">
              集中处理异步问题、转派、回复与关闭
            </p>
          </div>
          <Button onClick={openCreateDialog}>
            <PlusIcon className="size-4" />
            新建工单
          </Button>
        </div>

        <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-5">
          {quickViews.map((view) => {
            const active = quickView === view.key
            return (
              <button
                key={view.key}
                type="button"
                className={`rounded-xl border p-4 text-left transition ${
                  active
                    ? "border-primary bg-primary/5 shadow-sm"
                    : "border-border bg-background hover:border-primary/40"
                }`}
                onClick={() => {
                  setQuickView(view.key)
                  setResult((current) => ({
                    ...current,
                    page: { ...current.page, page: 1 },
                  }))
                }}
              >
                <div className="text-sm font-medium">{view.label}</div>
                <div className="mt-1 text-2xl font-semibold">{view.count}</div>
                <div className="mt-2 text-xs text-muted-foreground">{view.description}</div>
              </button>
            )
          })}
        </div>

        <div className="grid gap-3 lg:grid-cols-[minmax(0,1.4fr)_repeat(5,minmax(0,1fr))_auto]">
          <div className="relative">
            <SearchIcon className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              className="pl-9"
              value={keywordInput}
              placeholder="搜索工单号、标题或描述"
              onChange={(event) => setKeywordInput(event.target.value)}
              onKeyDown={handleFilterKeyDown}
            />
          </div>
          <OptionCombobox
            value={statusFilter}
            onChange={(value) => {
              setStatusFilter(value)
              setQuickView("all")
              setResult((current) => ({
                ...current,
                page: { ...current.page, page: 1 },
              }))
            }}
            placeholder="全部状态"
            options={[
              { value: "all", label: "全部状态" },
              { value: "new", label: "新建" },
              { value: "open", label: "处理中" },
              { value: "pending_customer", label: "待客户反馈" },
              { value: "pending_internal", label: "待内部处理" },
              { value: "resolved", label: "已解决" },
              { value: "closed", label: "已关闭" },
              { value: "cancelled", label: "已取消" },
            ]}
          />
          <OptionCombobox
            value={priorityFilter}
            onChange={(value) => {
              setPriorityFilter(value)
              setQuickView("all")
              setResult((current) => ({
                ...current,
                page: { ...current.page, page: 1 },
              }))
            }}
            placeholder="全部优先级"
            options={[
              { value: "all", label: "全部优先级" },
              { value: "1", label: "低" },
              { value: "2", label: "普通" },
              { value: "3", label: "高" },
              { value: "4", label: "紧急" },
            ]}
          />
          <OptionCombobox
            value={teamFilter}
            onChange={(value) => {
              setTeamFilter(value)
              setQuickView("all")
              setResult((current) => ({
                ...current,
                page: { ...current.page, page: 1 },
              }))
            }}
            placeholder="全部团队"
            options={teamOptions}
          />
          <OptionCombobox
            value={assigneeFilter}
            onChange={(value) => {
              setAssigneeFilter(value)
              setQuickView("all")
              setResult((current) => ({
                ...current,
                page: { ...current.page, page: 1 },
              }))
            }}
            placeholder="全部处理人"
            options={agentOptions}
          />
          <OptionCombobox
            value={watchFilter}
            onChange={(value) => {
              setWatchFilter(value)
              setQuickView("all")
              setResult((current) => ({
                ...current,
                page: { ...current.page, page: 1 },
              }))
            }}
            placeholder="全部工单"
            options={[
              { value: "all", label: "全部工单" },
              { value: "watching", label: "我的关注" },
            ]}
          />
          <div className="flex gap-2">
            <Button variant="outline" onClick={applyFilters}>
              查询
            </Button>
            <Button variant="outline" onClick={() => void loadTickets()}>
              <RefreshCcwIcon className="size-4" />
            </Button>
          </div>
        </div>

        <div className="overflow-hidden rounded-lg border bg-background">
          <Table>
            <TableHeader className="bg-muted/35">
              <TableRow>
                <TableHead>工单</TableHead>
                <TableHead>客户</TableHead>
                <TableHead>优先级</TableHead>
                <TableHead>状态</TableHead>
                <TableHead>处理人</TableHead>
                <TableHead>团队</TableHead>
                <TableHead>关注</TableHead>
                <TableHead>更新时间</TableHead>
                <TableHead className="text-right">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? (
                <TableRow>
                  <TableCell colSpan={9} className="h-32 text-center text-muted-foreground">
                    加载中...
                  </TableCell>
                </TableRow>
              ) : result.results.length > 0 ? (
                result.results.map((item) => (
                  <TableRow key={item.id}>
                    <TableCell>
                      <div className="space-y-1">
                        <div className="font-medium">{item.title}</div>
                        <div className="flex items-center gap-2 text-xs text-muted-foreground">
                          <span>{item.ticketNo}</span>
                          {quickView !== "all" ? (
                            <span className="rounded bg-muted px-1.5 py-0.5">
                              {quickViews.find((view) => view.key === quickView)?.label}
                            </span>
                          ) : null}
                        </div>
                      </div>
                    </TableCell>
                    <TableCell>{item.customer?.name || "未绑定客户"}</TableCell>
                    <TableCell>
                      <TicketPriorityBadge priority={item.priority} />
                    </TableCell>
                    <TableCell>
                      <TicketStatusBadge status={item.status} />
                    </TableCell>
                    <TableCell>{item.currentAssigneeName || "未指派"}</TableCell>
                    <TableCell>{item.currentTeamName || "未分组"}</TableCell>
                    <TableCell>
                      <Button
                        variant="ghost"
                        size="sm"
                        disabled={actionLoadingId === item.id}
                        onClick={() => void handleWatchToggle(item)}
                      >
                        <StarIcon
                          className={`size-4 ${
                            item.watchedByMe ? "fill-current text-amber-500" : ""
                          }`}
                        />
                        {item.watchedByMe ? "已关注" : "关注"}
                      </Button>
                    </TableCell>
                    <TableCell>{item.updatedAt ? formatDateTime(item.updatedAt) : "—"}</TableCell>
                    <TableCell className="text-right">
                      <div className="flex justify-end gap-2">
                        <Link
                          href={`/tickets/detail?id=${item.id}`}
                          target="_blank"
                          rel="noreferrer"
                        >
                          <Button variant="outline" size="sm">详情</Button>
                        </Link>
                        <Button variant="ghost" size="sm" onClick={() => openEditDialog(item)}>
                          编辑
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))
              ) : (
                <TableRow>
                  <TableCell colSpan={9} className="h-32 text-center text-muted-foreground">
                    暂无工单
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </div>

        <ListPagination
          page={result.page.page}
          total={result.page.total}
          limit={result.page.limit}
          loading={loading}
          onPageChange={(page) =>
            setResult((current) => ({
              ...current,
              page: { ...current.page, page },
            }))
          }
          onLimitChange={(limit) =>
            setResult((current) => ({
              ...current,
              page: { ...current.page, page: 1, limit },
            }))
          }
        />
      </div>

      <EditDialog
        open={dialogOpen}
        saving={saving}
        itemId={editingItemId}
        onOpenChange={handleDialogOpenChange}
        onSubmit={handleSubmit}
      />
    </div>
  )
}
