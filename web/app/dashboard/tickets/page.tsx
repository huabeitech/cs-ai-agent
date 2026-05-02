"use client"

import { PlusIcon, RefreshCcwIcon } from "lucide-react"
import { useCallback, useEffect, useState } from "react"
import { toast } from "sonner"

import { OptionCombobox } from "@/components/option-combobox"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import {
  createTicket,
  fetchTicketSummary,
  fetchTickets,
  updateTicket,
  type CreateTicketPayload,
  type TicketItem,
  type TicketListQuery,
  type TicketStatus,
  type TicketSummary,
  type UpdateTicketPayload,
} from "@/lib/api/ticket"
import { formatDateTime } from "@/lib/utils"
import { EditDialog } from "./_components/edit"
import { TicketAssignDialog } from "./_components/ticket-assign-dialog"
import { TicketStatusBadge } from "./_components/ticket-status-badge"
import { TicketStatusDialog } from "./_components/ticket-status-dialog"

const emptySummary: TicketSummary = {
  all: 0,
  pending: 0,
  inProgress: 0,
  done: 0,
  unassigned: 0,
  mine: 0,
  stale: 0,
}

const statusOptions = [
  { value: "", label: "全部状态" },
  { value: "pending", label: "待处理" },
  { value: "in_progress", label: "处理中" },
  { value: "done", label: "已处理" },
]

function isUpdatePayload(payload: CreateTicketPayload | UpdateTicketPayload): payload is UpdateTicketPayload {
  return "ticketId" in payload
}

export default function TicketsPage() {
  const [tickets, setTickets] = useState<TicketItem[]>([])
  const [summary, setSummary] = useState<TicketSummary>(emptySummary)
  const [keyword, setKeyword] = useState("")
  const [status, setStatus] = useState("")
  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState(false)
  const [editOpen, setEditOpen] = useState(false)
  const [editingId, setEditingId] = useState<number | null>(null)
  const [assignTicketItem, setAssignTicketItem] = useState<TicketItem | null>(null)
  const [statusTicketItem, setStatusTicketItem] = useState<TicketItem | null>(null)

  const loadData = useCallback(async () => {
    setLoading(true)
    try {
      const query: TicketListQuery = {
        page: 1,
        limit: 50,
        keyword: keyword.trim() || undefined,
        status: status ? (status as TicketStatus) : undefined,
      }
      const [ticketData, summaryData] = await Promise.all([
        fetchTickets(query),
        fetchTicketSummary(),
      ])
      setTickets(Array.isArray(ticketData.results) ? ticketData.results : [])
      setSummary(summaryData ?? emptySummary)
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "加载工单失败")
    } finally {
      setLoading(false)
    }
  }, [keyword, status])

  useEffect(() => {
    void loadData()
  }, [loadData])

  async function handleEditSubmit(payload: CreateTicketPayload | UpdateTicketPayload) {
    setSaving(true)
    try {
      if (isUpdatePayload(payload)) {
        await updateTicket(payload)
        toast.success("工单已更新")
      } else {
        await createTicket(payload)
        toast.success("工单已创建")
      }
      setEditOpen(false)
      setEditingId(null)
      await loadData()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "保存工单失败")
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="flex flex-col gap-4 p-6">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl font-semibold tracking-normal">工单</h1>
          <p className="text-sm text-muted-foreground">轻量工单列表临时入口</p>
        </div>
        <div className="flex items-center gap-2">
          <Button type="button" variant="outline" onClick={() => void loadData()} disabled={loading}>
            <RefreshCcwIcon className="size-4" />
            刷新
          </Button>
          <Button
            type="button"
            onClick={() => {
              setEditingId(null)
              setEditOpen(true)
            }}
          >
            <PlusIcon className="size-4" />
            新建工单
          </Button>
        </div>
      </div>

      <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-7">
        {[
          ["全部", summary.all],
          ["待处理", summary.pending],
          ["处理中", summary.inProgress],
          ["已处理", summary.done],
          ["未分配", summary.unassigned],
          ["我的", summary.mine],
          ["超时未动", summary.stale],
        ].map(([label, value]) => (
          <div key={label} className="rounded-lg border border-border bg-background p-3">
            <div className="text-xs text-muted-foreground">{label}</div>
            <div className="mt-1 text-xl font-semibold">{value}</div>
          </div>
        ))}
      </div>

      <div className="flex flex-wrap items-center gap-2">
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
            value={status}
            onChange={setStatus}
            placeholder="全部状态"
            options={statusOptions}
          />
        </div>
      </div>

      <div className="overflow-hidden rounded-lg border border-border">
        <div className="grid grid-cols-[minmax(0,1fr)_120px_140px_180px] gap-3 border-b bg-muted/40 px-4 py-2 text-xs font-medium text-muted-foreground">
          <div>工单</div>
          <div>状态</div>
          <div>处理人</div>
          <div className="text-right">操作</div>
        </div>
        {tickets.length === 0 ? (
          <div className="px-4 py-10 text-center text-sm text-muted-foreground">
            {loading ? "加载中..." : "暂无工单"}
          </div>
        ) : (
          tickets.map((ticket) => (
            <div
              key={ticket.id}
              className="grid grid-cols-[minmax(0,1fr)_120px_140px_180px] gap-3 border-b px-4 py-3 last:border-b-0"
            >
              <div className="min-w-0">
                <div className="truncate text-sm font-medium">{ticket.title}</div>
                <div className="mt-1 flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
                  <span>{ticket.ticketNo}</span>
                  <span>{ticket.updatedAt ? formatDateTime(ticket.updatedAt) : "-"}</span>
                </div>
              </div>
              <div>
                <TicketStatusBadge status={ticket.status} />
              </div>
              <div className="truncate text-sm text-muted-foreground">
                {ticket.currentAssigneeName || "未分配"}
              </div>
              <div className="flex justify-end gap-2">
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => {
                    setEditingId(ticket.id)
                    setEditOpen(true)
                  }}
                >
                  编辑
                </Button>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => setAssignTicketItem(ticket)}
                >
                  指派
                </Button>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => setStatusTicketItem(ticket)}
                >
                  状态
                </Button>
              </div>
            </div>
          ))
        )}
      </div>

      <EditDialog
        open={editOpen}
        saving={saving}
        itemId={editingId}
        onOpenChange={(open) => {
          setEditOpen(open)
          if (!open) {
            setEditingId(null)
          }
        }}
        onSubmit={handleEditSubmit}
      />
      <TicketAssignDialog
        open={!!assignTicketItem}
        ticketId={assignTicketItem?.id ?? null}
        currentAssigneeId={assignTicketItem?.currentAssigneeId}
        onOpenChange={(open) => {
          if (!open) {
            setAssignTicketItem(null)
          }
        }}
        onSuccess={loadData}
      />
      <TicketStatusDialog
        open={!!statusTicketItem}
        ticketId={statusTicketItem?.id ?? null}
        currentStatus={statusTicketItem?.status}
        onOpenChange={(open) => {
          if (!open) {
            setStatusTicketItem(null)
          }
        }}
        onSuccess={loadData}
      />
    </div>
  )
}
