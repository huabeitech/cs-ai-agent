"use client"

import Link from "next/link"
import {
  ArrowRightLeftIcon,
  CircleOffIcon,
  ClipboardListIcon,
  PlusIcon,
  RefreshCcwIcon,
  SearchIcon,
  SaveIcon,
  Settings2Icon,
  SquarePenIcon,
  StarIcon,
  Trash2Icon,
} from "lucide-react"
import { useCallback, useEffect, useMemo, useState } from "react"
import { toast } from "sonner"

import { ListPagination } from "@/components/list-pagination"
import { OptionCombobox } from "@/components/option-combobox"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
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
import { fetchTicketCategoriesAll, type TicketCategory } from "@/lib/api/ticket-config"
import {
  createTicket,
  batchWatchTickets,
  deleteTicketView,
  fetchTicketViews,
  fetchTicketSummary,
  fetchTickets,
  saveTicketView,
  updateTicket,
  unwatchTicket,
  watchTicket,
  type Paging,
  type TicketSavedView,
  type TicketItem,
  type TicketSummary,
} from "@/lib/api/ticket"
import { readSession } from "@/lib/auth"
import { formatDateTime } from "@/lib/utils"
import { EditDialog } from "./_components/edit"
import { TicketAssignDialog } from "./_components/ticket-assign-dialog"
import { TicketPriorityBadge } from "./_components/ticket-priority-badge"
import { TicketReasonDialog } from "./_components/ticket-reason-dialog"
import { TicketSLABadge } from "./_components/ticket-sla-badge"
import { TicketStatusDialog } from "./_components/ticket-status-dialog"
import { TicketStatusBadge } from "./_components/ticket-status-badge"

const emptyPaging: Paging = { page: 1, limit: 20, total: 0 }
const emptySummary: TicketSummary = {
  all: 0,
  mine: 0,
  watching: 0,
  collaboration: 0,
  participating: 0,
  mentioned: 0,
  unassigned: 0,
  pendingCustomer: 0,
  pendingInternal: 0,
  overdue: 0,
}

type QuickViewKey =
  | "all"
  | "mine"
  | "collaboration"
  | "participating"
  | "mentioned"
  | "watching"
  | "unassigned"
  | "pending_customer"
  | "pending_internal"
  | "overdue"

type SavedTicketView = {
  id: number
  name: string
  keywordInput: string
  keyword: string
  statusFilter: string
  priorityFilter: string
  severityFilter: string
  categoryFilter: string
  teamFilter: string
  assigneeFilter: string
  sourceFilter: string
  watchFilter: string
  quickView: QuickViewKey
}

function isClosedStatus(status: string) {
  return status === "resolved" || status === "closed" || status === "cancelled"
}

function getResolveDeadlineTime(ticket: TicketItem) {
  if (!ticket.resolveDeadlineAt) {
    return null
  }
  const deadline = new Date(ticket.resolveDeadlineAt.replace(" ", "T"))
  if (Number.isNaN(deadline.getTime())) {
    return null
  }
  return deadline.getTime()
}

function isOverdueTicket(ticket: TicketItem) {
  const deadline = getResolveDeadlineTime(ticket)
  if (deadline === null || isClosedStatus(ticket.status)) {
    return false
  }
  return deadline < Date.now()
}

function isRiskTicket(ticket: TicketItem) {
  const deadline = getResolveDeadlineTime(ticket)
  if (deadline === null || isClosedStatus(ticket.status)) {
    return false
  }
  const remainingMinutes = Math.floor((deadline - Date.now()) / 60000)
  return remainingMinutes >= 0 && remainingMinutes <= 240
}

function getTicketRowClassName(ticket: TicketItem) {
  if (isOverdueTicket(ticket)) {
    return "bg-red-50/80 hover:bg-red-50"
  }
  if (ticket.currentAssigneeId <= 0 && !isClosedStatus(ticket.status)) {
    return "bg-amber-50/70 hover:bg-amber-50"
  }
  if (isRiskTicket(ticket)) {
    return "bg-orange-50/60 hover:bg-orange-50"
  }
  return ""
}

function parseSavedTicketView(item: TicketSavedView): SavedTicketView | null {
  const filters = item.filters
  if (!filters || typeof filters !== "object") {
    return null
  }
  return {
    id: item.id,
    name: item.name,
    keywordInput: typeof filters.keywordInput === "string" ? filters.keywordInput : "",
    keyword: typeof filters.keyword === "string" ? filters.keyword : "",
    statusFilter: typeof filters.statusFilter === "string" ? filters.statusFilter : "all",
    priorityFilter: typeof filters.priorityFilter === "string" ? filters.priorityFilter : "all",
    severityFilter: typeof filters.severityFilter === "string" ? filters.severityFilter : "all",
    categoryFilter: typeof filters.categoryFilter === "string" ? filters.categoryFilter : "all",
    teamFilter: typeof filters.teamFilter === "string" ? filters.teamFilter : "all",
    assigneeFilter: typeof filters.assigneeFilter === "string" ? filters.assigneeFilter : "all",
    sourceFilter: typeof filters.sourceFilter === "string" ? filters.sourceFilter : "all",
    watchFilter: typeof filters.watchFilter === "string" ? filters.watchFilter : "all",
    quickView: typeof filters.quickView === "string" ? (filters.quickView as QuickViewKey) : "all",
  }
}

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
  const [severityFilter, setSeverityFilter] = useState("all")
  const [categoryFilter, setCategoryFilter] = useState("all")
  const [teamFilter, setTeamFilter] = useState("all")
  const [assigneeFilter, setAssigneeFilter] = useState("all")
  const [sourceFilter, setSourceFilter] = useState("all")
  const [watchFilter, setWatchFilter] = useState("all")
  const [quickView, setQuickView] = useState<QuickViewKey>("all")
  const [savedViews, setSavedViews] = useState<SavedTicketView[]>([])
  const [activeSavedViewId, setActiveSavedViewId] = useState("all")
  const [teams, setTeams] = useState<AdminAgentTeam[]>([])
  const [agents, setAgents] = useState<AdminAgentProfile[]>([])
  const [categories, setCategories] = useState<TicketCategory[]>([])
  const [actionLoadingId, setActionLoadingId] = useState<number | null>(null)
  const [selectedTicketIds, setSelectedTicketIds] = useState<number[]>([])
  const [assignDialogOpen, setAssignDialogOpen] = useState(false)
  const [statusDialogOpen, setStatusDialogOpen] = useState(false)
  const [closeDialogOpen, setCloseDialogOpen] = useState(false)
  const [reopenDialogOpen, setReopenDialogOpen] = useState(false)
  const [actionTicket, setActionTicket] = useState<TicketItem | null>(null)
  const currentUserId = readSession()?.user?.id ?? 0

  const activeStatusFilter = useMemo(() => {
    if (quickView === "pending_customer") {
      return "pending_customer"
    }
    if (quickView === "pending_internal") {
      return "pending_internal"
    }
    return statusFilter === "all" ? undefined : statusFilter
  }, [quickView, statusFilter])

  const activeWatchFilter = quickView === "watching" ? 1 : watchFilter === "watching" ? 1 : undefined
  const activeCollaborationFilter = quickView === "collaboration" ? 1 : undefined
  const activeCollaboratingFilter = quickView === "participating" ? 1 : undefined
  const activeMentionedFilter = quickView === "mentioned" ? 1 : undefined
  const activeMineFilter = quickView === "mine" ? 1 : undefined
  const activeUnassignedFilter = quickView === "unassigned" ? 1 : undefined
  const activeOverdueFilter = quickView === "overdue" ? 1 : undefined

  const loadSummary = useCallback(async () => {
    try {
      const data = await fetchTicketSummary()
      setSummary(data)
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "加载工单汇总失败")
    }
  }, [])

  const loadSavedViews = useCallback(async () => {
    try {
      const data = await fetchTicketViews()
      const views = Array.isArray(data)
        ? data
            .map(parseSavedTicketView)
            .filter((item): item is SavedTicketView => item !== null)
        : []
      setSavedViews(views)
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "加载保存视图失败")
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
        severity: severityFilter === "all" ? undefined : Number(severityFilter),
        categoryId: categoryFilter === "all" ? undefined : Number(categoryFilter),
        currentTeamId: teamFilter === "all" ? undefined : Number(teamFilter),
        currentAssigneeId: assigneeFilter === "all" ? undefined : Number(assigneeFilter),
        source: sourceFilter === "all" ? undefined : sourceFilter,
        watching: activeWatchFilter,
        collaboration: activeCollaborationFilter,
        collaborating: activeCollaboratingFilter,
        mentioned: activeMentionedFilter,
        mine: activeMineFilter,
        unassigned: activeUnassignedFilter,
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
    activeMineFilter,
    activeOverdueFilter,
    activeStatusFilter,
    activeUnassignedFilter,
    activeWatchFilter,
    activeCollaborationFilter,
    activeCollaboratingFilter,
    activeMentionedFilter,
    assigneeFilter,
    categoryFilter,
    keyword,
    priorityFilter,
    result.page.limit,
    result.page.page,
    severityFilter,
    sourceFilter,
    teamFilter,
  ])

  useEffect(() => {
    void loadTickets()
  }, [loadTickets])

  useEffect(() => {
    setSelectedTicketIds((current) =>
      current.filter((ticketId) => result.results.some((item) => item.id === ticketId)),
    )
  }, [result.results])

  useEffect(() => {
    void loadSummary()
  }, [loadSummary])

  useEffect(() => {
    void loadSavedViews()
  }, [loadSavedViews])

  useEffect(() => {
    void (async () => {
      const [teamData, agentData, categoryData] = await Promise.all([
        fetchAgentTeamsAll(),
        fetchAgentProfilesAll(),
        fetchTicketCategoriesAll(),
      ])
      setTeams(Array.isArray(teamData) ? teamData : [])
      setAgents(Array.isArray(agentData) ? agentData : [])
      setCategories(Array.isArray(categoryData) ? categoryData : [])
    })()
  }, [])

  function resetToFirstPage() {
    setResult((current) => ({
      ...current,
      page: { ...current.page, page: 1 },
    }))
  }

  function applyFilters() {
    setKeyword(keywordInput)
    setActiveSavedViewId("all")
    resetToFirstPage()
  }

  function buildCurrentView(name: string, existingId?: number): SavedTicketView {
    return {
      id: existingId ?? 0,
      name,
      keywordInput,
      keyword,
      statusFilter,
      priorityFilter,
      severityFilter,
      categoryFilter,
      teamFilter,
      assigneeFilter,
      sourceFilter,
      watchFilter,
      quickView,
    }
  }

  function applySavedView(view: SavedTicketView) {
    setKeywordInput(view.keywordInput)
    setKeyword(view.keyword)
    setStatusFilter(view.statusFilter)
    setPriorityFilter(view.priorityFilter)
    setSeverityFilter(view.severityFilter)
    setCategoryFilter(view.categoryFilter)
    setTeamFilter(view.teamFilter)
    setAssigneeFilter(view.assigneeFilter)
    setSourceFilter(view.sourceFilter)
    setWatchFilter(view.watchFilter)
    setQuickView(view.quickView)
    setActiveSavedViewId(String(view.id))
    setResult((current) => ({
      ...current,
      page: { ...current.page, page: 1 },
    }))
  }

  function resetAllFilters() {
    setKeywordInput("")
    setKeyword("")
    setStatusFilter("all")
    setPriorityFilter("all")
    setSeverityFilter("all")
    setCategoryFilter("all")
    setTeamFilter("all")
    setAssigneeFilter("all")
    setSourceFilter("all")
    setWatchFilter("all")
    setQuickView("all")
    setActiveSavedViewId("all")
    resetToFirstPage()
  }

  async function handleSaveCurrentView() {
    const currentView = activeSavedViewId === "all"
      ? null
      : savedViews.find((item) => String(item.id) === activeSavedViewId) ?? null
    const defaultName = currentView?.name ?? ""
    const name = window.prompt("输入视图名称", defaultName)?.trim()
    if (!name) {
      return
    }
    const nextView = buildCurrentView(name, currentView?.id)
    try {
      const saved = await saveTicketView({
        id: nextView.id > 0 ? nextView.id : undefined,
        name: nextView.name,
        filters: {
          keywordInput: nextView.keywordInput,
          keyword: nextView.keyword,
          statusFilter: nextView.statusFilter,
          priorityFilter: nextView.priorityFilter,
          severityFilter: nextView.severityFilter,
          categoryFilter: nextView.categoryFilter,
          teamFilter: nextView.teamFilter,
          assigneeFilter: nextView.assigneeFilter,
          sourceFilter: nextView.sourceFilter,
          watchFilter: nextView.watchFilter,
          quickView: nextView.quickView,
        },
      })
      const parsed = parseSavedTicketView(saved)
      if (parsed) {
        setSavedViews((current) => {
          const exists = current.some((item) => item.id === parsed.id)
          return exists ? current.map((item) => (item.id === parsed.id ? parsed : item)) : [parsed, ...current]
        })
        setActiveSavedViewId(String(parsed.id))
      }
      toast.success(currentView ? "视图已更新" : "视图已保存")
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "保存视图失败")
    }
  }

  async function handleDeleteCurrentView() {
    if (activeSavedViewId === "all") {
      toast.error("当前不是已保存视图")
      return
    }
    const currentView = savedViews.find((item) => String(item.id) === activeSavedViewId)
    if (!currentView) {
      return
    }
    const confirmed = window.confirm(`确认删除视图「${currentView.name}」吗？`)
    if (!confirmed) {
      return
    }
    try {
      await deleteTicketView(currentView.id)
      setSavedViews((current) => current.filter((item) => item.id !== currentView.id))
      setActiveSavedViewId("all")
      toast.success("视图已删除")
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "删除视图失败")
    }
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

  function openAssignDialog(item: TicketItem) {
    setActionTicket(item)
    setAssignDialogOpen(true)
  }

  function openStatusDialog(item: TicketItem) {
    setActionTicket(item)
    setStatusDialogOpen(true)
  }

  function openCloseDialog(item: TicketItem) {
    setActionTicket(item)
    setCloseDialogOpen(true)
  }

  function openReopenDialog(item: TicketItem) {
    setActionTicket(item)
    setReopenDialogOpen(true)
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

  async function refreshAll() {
    await Promise.all([loadSummary(), loadTickets()])
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
      await refreshAll()
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
      await refreshAll()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "更新关注状态失败")
    } finally {
      setActionLoadingId(null)
    }
  }

  function toggleSelectTicket(ticketId: number, checked: boolean) {
    setSelectedTicketIds((current) => {
      if (checked) {
        if (current.includes(ticketId)) {
          return current
        }
        return current.concat(ticketId)
      }
      return current.filter((item) => item !== ticketId)
    })
  }

  function handleSelectAll(checked: boolean) {
    setSelectedTicketIds(checked ? result.results.map((item) => item.id) : [])
  }

  async function handleBatchWatch(watched: boolean) {
    if (selectedTicketIds.length === 0) {
      toast.error("请先选择工单")
      return
    }
    setSaving(true)
    try {
      await batchWatchTickets({ ticketIds: selectedTicketIds, watched })
      toast.success(watched ? `已批量关注 ${selectedTicketIds.length} 张工单` : `已批量取消关注 ${selectedTicketIds.length} 张工单`)
      setSelectedTicketIds([])
      await refreshAll()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "批量更新关注失败")
    } finally {
      setSaving(false)
    }
  }

  const categoryOptions = [{ value: "all", label: "全部分类" }].concat(
    categories.map((category) => ({
      value: String(category.id),
      label: category.parentName ? `${category.parentName} / ${category.name}` : category.name,
    })),
  )

  const teamOptions = [{ value: "all", label: "全部团队" }].concat(
    teams.map((team) => ({ value: String(team.id), label: team.name })),
  )

  const agentOptions = [{ value: "all", label: "全部处理人" }].concat(
    agents.map((agent) => ({
      value: String(agent.userId),
      label: agent.displayName || agent.nickname || agent.username || `客服#${agent.userId}`,
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
    { key: "collaboration", label: "协作相关", description: "我参与或提及我的工单", count: summary.collaboration },
    { key: "participating", label: "我参与的", description: "我是协作人的工单", count: summary.participating },
    { key: "mentioned", label: "提及我的", description: "内部备注里提到了我", count: summary.mentioned },
    { key: "unassigned", label: "待分配", description: "尚未指派处理人", count: summary.unassigned },
    { key: "watching", label: "我的关注", description: "我在跟进的工单", count: summary.watching },
    {
      key: "pending_customer",
      label: "待客户反馈",
      description: "等待客户补充信息",
      count: summary.pendingCustomer,
    },
    {
      key: "pending_internal",
      label: "待内部处理",
      description: "等待内部协作处理",
      count: summary.pendingInternal,
    },
    { key: "overdue", label: "已超时", description: "解决 SLA 已超时", count: summary.overdue },
  ] 

  const allSelected =
    result.results.length > 0 && selectedTicketIds.length === result.results.length

  const displayResults = useMemo(() => {
    const list = [...result.results]
    list.sort((left, right) => {
      const leftOverdue = isOverdueTicket(left)
      const rightOverdue = isOverdueTicket(right)
      if (leftOverdue !== rightOverdue) {
        return leftOverdue ? -1 : 1
      }

      if (quickView === "unassigned") {
        const leftUnassigned = left.currentAssigneeId <= 0 ? 1 : 0
        const rightUnassigned = right.currentAssigneeId <= 0 ? 1 : 0
        if (leftUnassigned !== rightUnassigned) {
          return rightUnassigned - leftUnassigned
        }
      }

      const leftDeadline = getResolveDeadlineTime(left)
      const rightDeadline = getResolveDeadlineTime(right)
      if (leftDeadline !== null || rightDeadline !== null) {
        if (leftDeadline === null) {
          return 1
        }
        if (rightDeadline === null) {
          return -1
        }
        if (leftDeadline !== rightDeadline) {
          return leftDeadline - rightDeadline
        }
      }

      if (left.priority !== right.priority) {
        return right.priority - left.priority
      }

      return (right.id ?? 0) - (left.id ?? 0)
    })
    return list
  }, [quickView, result.results])

  const savedViewOptions = useMemo(
    () =>
      [{ value: "all", label: "当前筛选" }].concat(
        savedViews.map((item) => ({
          value: String(item.id),
          label: item.name,
        })),
      ),
    [savedViews],
  )

  return (
    <div className="min-h-0 flex-1 overflow-auto bg-muted/20 p-4 md:p-6">
      <div className="flex w-full flex-col gap-4">
        <div className="flex flex-col gap-3 xl:flex-row xl:items-center xl:justify-between">
          <div>
            <h1 className="text-xl font-semibold">工单</h1>
            <p className="mt-1 text-sm text-muted-foreground">
              作为队列工作台处理异步问题、分派、流转与关闭
            </p>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-48">
              <OptionCombobox
                value={activeSavedViewId}
                onChange={(value) => {
                  if (value === "all") {
                    setActiveSavedViewId("all")
                    return
                  }
                  const nextView = savedViews.find((item) => String(item.id) === value)
                  if (nextView) {
                    applySavedView(nextView)
                  }
                }}
                placeholder="保存视图"
                options={savedViewOptions}
              />
            </div>
            <Button variant="outline" onClick={handleSaveCurrentView}>
              <SaveIcon className="size-4" />
              保存视图
            </Button>
            <Button variant="outline" onClick={handleDeleteCurrentView} disabled={activeSavedViewId === "all"}>
              <Trash2Icon className="size-4" />
              删除视图
            </Button>
            <Button onClick={openCreateDialog}>
              <PlusIcon className="size-4" />
              新建工单
            </Button>
          </div>
        </div>

        <div className="flex flex-wrap gap-2 text-xs text-muted-foreground">
          <span className="rounded bg-red-50 px-2 py-1 text-red-700">红色：已超时</span>
          <span className="rounded bg-amber-50 px-2 py-1 text-amber-700">黄色：待分配</span>
          <span className="rounded bg-orange-50 px-2 py-1 text-orange-700">橙色：即将超时</span>
        </div>

        <div className="rounded-lg border bg-background/80 p-2">
          <div className="flex flex-wrap gap-2">
            {quickViews.map((view) => {
              const active = quickView === view.key
              return (
                <button
                  key={view.key}
                  type="button"
                  className={`inline-flex min-w-0 items-center gap-2 rounded-md border px-3 py-1.5 text-sm transition ${
                    active
                      ? "border-primary bg-primary/8 text-primary shadow-sm"
                      : "border-border bg-background text-foreground hover:border-primary/40 hover:bg-muted/60"
                  }`}
                  onClick={() => {
                    setQuickView(view.key)
                    resetToFirstPage()
                  }}
                >
                  <span className="truncate font-medium">{view.label}</span>
                  <span
                    className={`rounded-full px-1.5 py-0.5 text-[11px] font-semibold tabular-nums ${
                      active ? "bg-primary/12 text-primary" : "bg-muted text-muted-foreground"
                    }`}
                  >
                    {view.count}
                  </span>
                </button>
              )
            })}
          </div>
        </div>

        <div className="grid gap-3 lg:grid-cols-4 xl:grid-cols-[minmax(0,1.5fr)_repeat(8,minmax(0,1fr))_auto]">
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
              resetToFirstPage()
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
            value={categoryFilter}
            onChange={(value) => {
              setCategoryFilter(value)
              setQuickView("all")
              resetToFirstPage()
            }}
            placeholder="全部分类"
            options={categoryOptions}
          />
          <OptionCombobox
            value={priorityFilter}
            onChange={(value) => {
              setPriorityFilter(value)
              setQuickView("all")
              resetToFirstPage()
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
            value={severityFilter}
            onChange={(value) => {
              setSeverityFilter(value)
              setQuickView("all")
              resetToFirstPage()
            }}
            placeholder="全部严重度"
            options={[
              { value: "all", label: "全部严重度" },
              { value: "1", label: "轻微" },
              { value: "2", label: "一般" },
              { value: "3", label: "严重" },
            ]}
          />
          <OptionCombobox
            value={sourceFilter}
            onChange={(value) => {
              setSourceFilter(value)
              setQuickView("all")
              resetToFirstPage()
            }}
            placeholder="全部来源"
            options={[
              { value: "all", label: "全部来源" },
              { value: "manual", label: "手动创建" },
              { value: "conversation", label: "会话转工单" },
              { value: "email", label: "邮件" },
              { value: "api", label: "API" },
              { value: "system", label: "系统" },
            ]}
          />
          <OptionCombobox
            value={teamFilter}
            onChange={(value) => {
              setTeamFilter(value)
              setQuickView("all")
              resetToFirstPage()
            }}
            placeholder="全部团队"
            options={teamOptions}
          />
          <OptionCombobox
            value={assigneeFilter}
            onChange={(value) => {
              setAssigneeFilter(value)
              setQuickView("all")
              resetToFirstPage()
            }}
            placeholder="全部处理人"
            options={agentOptions}
          />
          <OptionCombobox
            value={watchFilter}
            onChange={(value) => {
              setWatchFilter(value)
              setQuickView("all")
              resetToFirstPage()
            }}
            placeholder="全部关注"
            options={[
              { value: "all", label: "全部关注" },
              { value: "watching", label: "我关注的" },
            ]}
          />
          <div className="flex gap-2">
            <Button variant="outline" onClick={applyFilters}>
              查询
            </Button>
            <Button variant="outline" onClick={() => void refreshAll()}>
              <RefreshCcwIcon className="size-4" />
            </Button>
            <Button variant="ghost" onClick={resetAllFilters}>
              重置
            </Button>
          </div>
        </div>

        <div className="overflow-hidden rounded-lg border bg-background">
          {selectedTicketIds.length > 0 ? (
            <div className="flex flex-col gap-3 border-b bg-muted/30 px-4 py-3 lg:flex-row lg:items-center lg:justify-between">
              <div className="text-sm text-muted-foreground">
                已选择 <span className="font-medium text-foreground">{selectedTicketIds.length}</span> 张工单
              </div>
              <div className="flex flex-wrap gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => {
                    setActionTicket(null)
                    setAssignDialogOpen(true)
                  }}
                  disabled={saving}
                >
                  批量指派
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => {
                    setActionTicket(null)
                    setStatusDialogOpen(true)
                  }}
                  disabled={saving}
                >
                  批量改状态
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => void handleBatchWatch(true)}
                  disabled={saving}
                >
                  批量关注
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => void handleBatchWatch(false)}
                  disabled={saving}
                >
                  取消关注
                </Button>
                <Button variant="ghost" size="sm" onClick={() => setSelectedTicketIds([])} disabled={saving}>
                  清空选择
                </Button>
              </div>
            </div>
          ) : null}
          <Table>
            <TableHeader className="bg-muted/35">
              <TableRow>
                <TableHead className="w-10">
                  <Checkbox
                    checked={allSelected}
                    onCheckedChange={(checked) => handleSelectAll(Boolean(checked))}
                    aria-label="全选工单"
                  />
                </TableHead>
                <TableHead>工单</TableHead>
                <TableHead>客户</TableHead>
                <TableHead>分类</TableHead>
                <TableHead>优先级</TableHead>
                <TableHead>状态</TableHead>
                <TableHead>SLA 风险</TableHead>
                <TableHead>协作</TableHead>
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
                  <TableCell colSpan={13} className="h-32 text-center text-muted-foreground">
                    加载中...
                  </TableCell>
                </TableRow>
              ) : result.results.length > 0 ? (
                displayResults.map((item) => (
                  <TableRow key={item.id} className={getTicketRowClassName(item)}>
                    <TableCell>
                      <Checkbox
                        checked={selectedTicketIds.includes(item.id)}
                        onCheckedChange={(checked) => toggleSelectTicket(item.id, Boolean(checked))}
                        aria-label={`选择工单 ${item.ticketNo}`}
                      />
                    </TableCell>
                    <TableCell className="min-w-64">
                      <div className="space-y-1">
                        <div className="font-medium">{item.title}</div>
                        <div className="flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
                          <span>{item.ticketNo}</span>
                          <span>{item.source || "manual"}</span>
                          {item.currentAssigneeId <= 0 && !isClosedStatus(item.status) ? (
                            <span className="rounded bg-amber-100 px-1.5 py-0.5 text-amber-700">
                              待分配
                            </span>
                          ) : null}
                          {item.pendingReason && !isClosedStatus(item.status) ? (
                            <span className="rounded bg-muted px-1.5 py-0.5">{item.pendingReason}</span>
                          ) : null}
                        </div>
                      </div>
                    </TableCell>
                    <TableCell>{item.customer?.name || "未绑定客户"}</TableCell>
                    <TableCell>
                      {item.categoryName ? (
                        item.categoryName
                      ) : (
                        <Link
                          href="/ticket-categories"
                          className="inline-flex rounded-full border border-amber-300 bg-amber-50 px-2 py-1 text-xs text-amber-800 hover:bg-amber-100"
                        >
                          未分类，去配置
                        </Link>
                      )}
                    </TableCell>
                    <TableCell>
                      <TicketPriorityBadge priority={item.priority} />
                    </TableCell>
                    <TableCell>
                      <TicketStatusBadge status={item.status} />
                    </TableCell>
                    <TableCell>
                      <TicketSLABadge ticket={item} />
                    </TableCell>
                    <TableCell>
                      <div className="flex flex-wrap gap-1">
                        {quickView === "participating" || quickView === "collaboration" ? (
                          <Badge variant="outline">协作中</Badge>
                        ) : null}
                        {quickView === "mentioned" || quickView === "collaboration" ? (
                          <Badge variant="outline" className="border-amber-300 bg-amber-50 text-amber-800">
                            提及我
                          </Badge>
                        ) : null}
                        {item.currentAssigneeId === currentUserId ? (
                          <Badge variant="outline" className="border-blue-300 bg-blue-50 text-blue-800">
                            负责人
                          </Badge>
                        ) : null}
                      </div>
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
                          className={`size-4 ${item.watchedByMe ? "fill-current text-amber-500" : ""}`}
                        />
                        {item.watchedByMe ? "已关注" : "关注"}
                      </Button>
                    </TableCell>
                    <TableCell>{item.updatedAt ? formatDateTime(item.updatedAt) : "—"}</TableCell>
                    <TableCell className="text-right">
                      <div className="flex justify-end gap-2">
                        <Link href={`/tickets/detail?id=${item.id}`} target="_blank" rel="noreferrer">
                          <Button variant="outline" size="sm">
                            <ClipboardListIcon className="size-4" />
                            详情
                          </Button>
                        </Link>
                        <Button variant="ghost" size="sm" onClick={() => openAssignDialog(item)}>
                          <ArrowRightLeftIcon className="size-4" />
                          指派
                        </Button>
                        <Button variant="ghost" size="sm" onClick={() => openStatusDialog(item)}>
                          <SquarePenIcon className="size-4" />
                          状态
                        </Button>
                        {item.status === "closed" ? (
                          <Button variant="ghost" size="sm" onClick={() => openReopenDialog(item)}>
                            重开
                          </Button>
                        ) : (
                          <Button variant="ghost" size="sm" onClick={() => openCloseDialog(item)}>
                            <CircleOffIcon className="size-4" />
                            关闭
                          </Button>
                        )}
                        <Button variant="ghost" size="sm" onClick={() => openEditDialog(item)}>
                          编辑
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))
              ) : (
                <TableRow>
                  <TableCell colSpan={13} className="h-32 text-center text-muted-foreground">
                    暂无符合当前筛选条件的工单
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
      <TicketAssignDialog
        open={assignDialogOpen}
        ticketId={actionTicket?.id ?? null}
        ticketIds={actionTicket ? undefined : selectedTicketIds}
        currentTeamId={actionTicket?.currentTeamId}
        currentAssigneeId={actionTicket?.currentAssigneeId}
        onOpenChange={(open) => {
          setAssignDialogOpen(open)
          if (!open) {
            setActionTicket(null)
            setSelectedTicketIds((current) => current)
          }
        }}
        onSuccess={async () => {
          setSelectedTicketIds([])
          await refreshAll()
        }}
      />
      <TicketStatusDialog
        open={statusDialogOpen}
        ticketId={actionTicket?.id ?? null}
        ticketIds={actionTicket ? undefined : selectedTicketIds}
        currentStatus={actionTicket?.status}
        onOpenChange={(open) => {
          setStatusDialogOpen(open)
          if (!open) {
            setActionTicket(null)
          }
        }}
        onSuccess={async () => {
          setSelectedTicketIds([])
          await refreshAll()
        }}
      />
      <TicketReasonDialog
        open={closeDialogOpen}
        mode="close"
        ticketId={actionTicket?.id ?? null}
        defaultReason={actionTicket?.closeReason || "处理完成"}
        onOpenChange={(open) => {
          setCloseDialogOpen(open)
          if (!open) {
            setActionTicket(null)
          }
        }}
        onSuccess={refreshAll}
      />
      <TicketReasonDialog
        open={reopenDialogOpen}
        mode="reopen"
        ticketId={actionTicket?.id ?? null}
        defaultReason="客户有新反馈"
        onOpenChange={(open) => {
          setReopenDialogOpen(open)
          if (!open) {
            setActionTicket(null)
          }
        }}
        onSuccess={refreshAll}
      />
    </div>
  )
}
