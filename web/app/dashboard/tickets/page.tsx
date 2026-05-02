"use client"

import { PlusIcon, RefreshCcwIcon, SearchXIcon } from "lucide-react"
import { useCallback, useEffect, useMemo, useState } from "react"
import { toast } from "sonner"

import { OptionCombobox, type ComboboxOption } from "@/components/option-combobox"
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
  fetchTagsAll,
  type AdminAgentProfile,
  type TagTree,
} from "@/lib/api/admin"
import {
  createTicket,
  fetchTicketSummary,
  fetchTickets,
  type CreateTicketPayload,
  type TicketItem,
  type TicketListQuery,
  type TicketStatus,
  type TicketSummary,
} from "@/lib/api/ticket"
import { cn, formatDateTime } from "@/lib/utils"
import { EditDialog } from "./_components/edit"
import { TicketAssignDialog } from "./_components/ticket-assign-dialog"
import { TicketDetailDialog } from "./_components/ticket-detail-dialog"
import { TicketStatusBadge } from "./_components/ticket-status-badge"

type QuickViewKey =
  | "all"
  | "pending"
  | "in_progress"
  | "done"
  | "unassigned"
  | "mine"
  | "stale"

const emptySummary: TicketSummary = {
  all: 0,
  pending: 0,
  inProgress: 0,
  done: 0,
  unassigned: 0,
  mine: 0,
  stale: 0,
}

const assigneeAllOption: ComboboxOption = { value: "0", label: "全部负责人" }
const tagAllOption: ComboboxOption = { value: "0", label: "全部标签" }
const staleHourOptions: ComboboxOption[] = [
  { value: "24", label: "24 小时" },
  { value: "48", label: "48 小时" },
  { value: "168", label: "168 小时" },
]

function buildTagOptions(nodes: TagTree[], parentPath = ""): ComboboxOption[] {
  const result: ComboboxOption[] = []
  nodes.forEach((item) => {
    const currentPath = parentPath ? `${parentPath}/${item.name}` : item.name
    result.push({
      value: String(item.id),
      label: currentPath,
    })
    if (item.children.length > 0) {
      result.push(...buildTagOptions(item.children, currentPath))
    }
  })
  return result
}

function sourceLabel(source: string) {
  switch (source) {
    case "manual":
      return "手动"
    case "conversation":
      return "会话"
    default:
      return source || "-"
  }
}

function assigneeLabel(ticket: TicketItem) {
  if (ticket.currentAssigneeName) {
    return ticket.currentAssigneeName
  }
  if (ticket.currentAssigneeId > 0) {
    return `客服#${ticket.currentAssigneeId}`
  }
  return "未分配"
}

export default function TicketsPage() {
  const [tickets, setTickets] = useState<TicketItem[]>([])
  const [summary, setSummary] = useState<TicketSummary>(emptySummary)
  const [quickView, setQuickView] = useState<QuickViewKey>("all")
  const [keyword, setKeyword] = useState("")
  const [assigneeId, setAssigneeId] = useState("0")
  const [tagId, setTagId] = useState("0")
  const [staleHours, setStaleHours] = useState("24")
  const [assigneeOptions, setAssigneeOptions] = useState<ComboboxOption[]>([assigneeAllOption])
  const [tagOptions, setTagOptions] = useState<ComboboxOption[]>([tagAllOption])
  const [loading, setLoading] = useState(false)
  const [selectedTicketIds, setSelectedTicketIds] = useState<Set<number>>(new Set())
  const [batchAssignOpen, setBatchAssignOpen] = useState(false)
  const [selectedTicketId, setSelectedTicketId] = useState<number | null>(null)
  const [detailOpen, setDetailOpen] = useState(false)
  const [createOpen, setCreateOpen] = useState(false)
  const [savingCreate, setSavingCreate] = useState(false)

  const quickViews = useMemo(
    () =>
      [
        { key: "all", label: "全部工单", count: summary.all },
        { key: "pending", label: "待处理", count: summary.pending },
        { key: "in_progress", label: "处理中", count: summary.inProgress },
        { key: "done", label: "已处理", count: summary.done },
        { key: "unassigned", label: "待分配", count: summary.unassigned },
        { key: "mine", label: "我的工单", count: summary.mine },
        { key: "stale", label: "长时间未更新", count: summary.stale },
      ] satisfies Array<{ key: QuickViewKey; label: string; count: number }>,
    [summary],
  )

  const selectedIds = useMemo(() => Array.from(selectedTicketIds), [selectedTicketIds])
  const allPageSelected = tickets.length > 0 && tickets.every((ticket) => selectedTicketIds.has(ticket.id))

  const loadData = useCallback(async () => {
    setLoading(true)
    try {
      const staleThreshold = Number(staleHours)
      const query: TicketListQuery = {
        page: 1,
        limit: 50,
        keyword: keyword.trim() || undefined,
        currentAssigneeId: assigneeId !== "0" ? Number(assigneeId) : undefined,
        tagId: tagId !== "0" ? Number(tagId) : undefined,
      }

      if (quickView === "pending" || quickView === "in_progress" || quickView === "done") {
        query.status = quickView as TicketStatus
      }
      if (quickView === "unassigned") {
        query.unassigned = 1
      }
      if (quickView === "mine") {
        query.mine = 1
      }
      if (quickView === "stale") {
        query.staleHours = staleThreshold
      }

      const [ticketData, summaryData] = await Promise.all([
        fetchTickets(query),
        fetchTicketSummary({ staleHours: staleThreshold }),
      ])
      setTickets(Array.isArray(ticketData.results) ? ticketData.results : [])
      setSummary(summaryData ?? emptySummary)
      setSelectedTicketIds(new Set())
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "加载工单失败")
    } finally {
      setLoading(false)
    }
  }, [assigneeId, keyword, quickView, staleHours, tagId])

  useEffect(() => {
    void loadData()
  }, [loadData])

  useEffect(() => {
    let active = true
    Promise.all([fetchAgentProfilesAll(), fetchTagsAll()])
      .then(([agents, tags]) => {
        if (!active) {
          return
        }
        setAssigneeOptions([
          assigneeAllOption,
          ...(Array.isArray(agents) ? agents : []).map((agent: AdminAgentProfile) => ({
            value: String(agent.userId),
            label:
              agent.displayName ||
              agent.nickname ||
              agent.username ||
              `客服#${agent.userId}`,
          })),
        ])
        setTagOptions([tagAllOption, ...buildTagOptions(Array.isArray(tags) ? tags : [])])
      })
      .catch((error) => {
        toast.error(error instanceof Error ? error.message : "加载筛选项失败")
      })
    return () => {
      active = false
    }
  }, [])

  function toggleTicket(ticketId: number, checked: boolean) {
    setSelectedTicketIds((current) => {
      const next = new Set(current)
      if (checked) {
        next.add(ticketId)
      } else {
        next.delete(ticketId)
      }
      return next
    })
  }

  function togglePage(checked: boolean) {
    setSelectedTicketIds((current) => {
      const next = new Set(current)
      tickets.forEach((ticket) => {
        if (checked) {
          next.add(ticket.id)
        } else {
          next.delete(ticket.id)
        }
      })
      return next
    })
  }

  function resetFilters() {
    setQuickView("all")
    setKeyword("")
    setAssigneeId("0")
    setTagId("0")
    setStaleHours("24")
  }

  async function handleCreateTicket(payload: CreateTicketPayload) {
    setSavingCreate(true)
    try {
      await createTicket(payload)
      toast.success("工单已创建")
      setCreateOpen(false)
      await loadData()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "创建工单失败")
    } finally {
      setSavingCreate(false)
    }
  }

  return (
    <div className="flex flex-col gap-4 p-6">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl font-semibold tracking-normal">工单</h1>
          <p className="text-sm text-muted-foreground">按状态、负责人和标签快速处理轻量工单</p>
        </div>
        <div className="flex items-center gap-2">
          <Button type="button" variant="outline" onClick={() => void loadData()} disabled={loading}>
            <RefreshCcwIcon className={cn("size-4", loading ? "animate-spin" : "")} />
            刷新
          </Button>
          <Button type="button" onClick={() => setCreateOpen(true)}>
            <PlusIcon className="size-4" />
            新建工单
          </Button>
        </div>
      </div>

      <div className="rounded-lg border bg-background/80 p-2">
        <div className="flex flex-wrap gap-2">
          {quickViews.map((view) => (
            <button
              key={view.key}
              type="button"
              className={cn(
                "inline-flex min-w-0 items-center gap-2 rounded-md border px-3 py-1.5 text-sm transition",
                quickView === view.key
                  ? "border-primary bg-primary text-primary-foreground"
                  : "border-border bg-background hover:bg-muted",
              )}
              onClick={() => setQuickView(view.key)}
            >
              <span className="truncate font-medium">{view.label}</span>
              <span
                className={cn(
                  "rounded-full px-1.5 py-0.5 text-[11px] font-semibold tabular-nums",
                  quickView === view.key ? "bg-primary-foreground/20" : "bg-muted text-muted-foreground",
                )}
              >
                {view.count}
              </span>
            </button>
          ))}
        </div>
      </div>

      <div className="flex flex-wrap items-center gap-2 rounded-lg border bg-background/80 p-3">
        <Input
          className="w-full sm:w-72"
          placeholder="搜索编号、标题或描述"
          value={keyword}
          onChange={(event) => setKeyword(event.target.value)}
          onKeyDown={(event) => {
            if (event.key === "Enter") {
              void loadData()
            }
          }}
        />
        <div className="w-full sm:w-44">
          <OptionCombobox
            value={assigneeId}
            onChange={setAssigneeId}
            placeholder="全部负责人"
            options={assigneeOptions}
          />
        </div>
        <div className="w-full sm:w-44">
          <OptionCombobox
            value={tagId}
            onChange={setTagId}
            placeholder="全部标签"
            options={tagOptions}
          />
        </div>
        <div className="w-full sm:w-40">
          <OptionCombobox
            value={staleHours}
            onChange={setStaleHours}
            placeholder="未更新阈值"
            options={staleHourOptions}
          />
        </div>
        <Button type="button" variant="outline" onClick={resetFilters}>
          <SearchXIcon className="size-4" />
          重置
        </Button>
      </div>

      <div className="flex min-h-8 flex-wrap items-center justify-between gap-2">
        <div className="text-sm text-muted-foreground">
          已选择 <span className="font-medium text-foreground">{selectedTicketIds.size}</span> 张工单
        </div>
        <Button
          type="button"
          size="sm"
          variant="outline"
          disabled={selectedTicketIds.size === 0}
          onClick={() => setBatchAssignOpen(true)}
        >
          批量指派
        </Button>
      </div>

      <div className="overflow-hidden rounded-lg border border-border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-10">
                <Checkbox
                  checked={allPageSelected}
                  onCheckedChange={(checked) => togglePage(!!checked)}
                  aria-label="选择当前页工单"
                />
              </TableHead>
              <TableHead>工单</TableHead>
              <TableHead className="w-28">状态</TableHead>
              <TableHead className="w-36">负责人</TableHead>
              <TableHead className="w-40">最后更新</TableHead>
              <TableHead className="w-24 text-right">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {tickets.length === 0 ? (
              <TableRow>
                <TableCell colSpan={6} className="py-12 text-center text-sm text-muted-foreground">
                  {loading ? "加载中..." : "暂无工单"}
                </TableCell>
              </TableRow>
            ) : (
              tickets.map((ticket) => (
                <TableRow key={ticket.id}>
                  <TableCell>
                    <Checkbox
                      checked={selectedTicketIds.has(ticket.id)}
                      onCheckedChange={(checked) => toggleTicket(ticket.id, !!checked)}
                      aria-label={`选择工单 ${ticket.ticketNo}`}
                    />
                  </TableCell>
                  <TableCell>
                    <div className="min-w-0 space-y-1">
                      <div className="truncate text-sm font-medium">{ticket.title}</div>
                      <div className="flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
                        <span className="font-mono">{ticket.ticketNo}</span>
                        <span>{ticket.customer?.name || (ticket.customerId ? `客户#${ticket.customerId}` : "无客户")}</span>
                        <span>{sourceLabel(ticket.source)}</span>
                        {ticket.channel ? <span>{ticket.channel}</span> : null}
                      </div>
                      {ticket.tags && ticket.tags.length > 0 ? (
                        <div className="flex flex-wrap gap-1">
                          {ticket.tags.slice(0, 3).map((tag) => (
                            <Badge key={tag.id} variant="outline" className="px-1.5 py-0 text-[11px]">
                              {tag.name}
                            </Badge>
                          ))}
                          {ticket.tags.length > 3 ? (
                            <Badge variant="outline" className="px-1.5 py-0 text-[11px]">
                              +{ticket.tags.length - 3}
                            </Badge>
                          ) : null}
                        </div>
                      ) : null}
                    </div>
                  </TableCell>
                  <TableCell>
                    <TicketStatusBadge status={ticket.status} />
                  </TableCell>
                  <TableCell className="max-w-36 truncate text-sm text-muted-foreground">
                    {assigneeLabel(ticket)}
                  </TableCell>
                  <TableCell className="text-sm text-muted-foreground">
                    {ticket.updatedAt ? formatDateTime(ticket.updatedAt) : "-"}
                  </TableCell>
                  <TableCell className="text-right">
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      onClick={() => {
                        setSelectedTicketId(ticket.id)
                        setDetailOpen(true)
                      }}
                    >
                      详情
                    </Button>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      <TicketAssignDialog
        open={batchAssignOpen}
        ticketId={null}
        ticketIds={selectedIds}
        onOpenChange={setBatchAssignOpen}
        onSuccess={loadData}
      />
      <EditDialog
        open={createOpen}
        saving={savingCreate}
        itemId={null}
        onOpenChange={setCreateOpen}
        onSubmit={handleCreateTicket}
      />
      <TicketDetailDialog
        open={detailOpen}
        ticketId={selectedTicketId}
        onOpenChange={(open) => {
          setDetailOpen(open)
          if (!open) {
            setSelectedTicketId(null)
          }
        }}
        onChanged={loadData}
      />
    </div>
  )
}
