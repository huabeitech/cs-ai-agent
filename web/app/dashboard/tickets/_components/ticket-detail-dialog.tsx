"use client"

import { useCallback, useEffect, useState } from "react"
import { MessageSquareTextIcon, RefreshCcwIcon, SendIcon, UserRoundIcon } from "lucide-react"
import { toast } from "sonner"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Separator } from "@/components/ui/separator"
import { Textarea } from "@/components/ui/textarea"
import {
  changeTicketStatus,
  createTicketProgress,
  fetchTicketDetail,
  type CreateTicketPayload,
  type TicketDetail,
  type TicketStatus,
  type UpdateTicketPayload,
  updateTicket,
} from "@/lib/api/ticket"
import { cn, formatDateTime } from "@/lib/utils"
import { EditDialog } from "./edit"
import { TicketAssignDialog } from "./ticket-assign-dialog"
import { TicketStatusBadge } from "./ticket-status-badge"

type TicketDetailDialogProps = {
  ticketId: number | null
  open: boolean
  onOpenChange: (open: boolean) => void
  onChanged: () => void
}

const statusOptions: Array<{ value: TicketStatus; label: string }> = [
  { value: "pending", label: "待处理" },
  { value: "in_progress", label: "处理中" },
  { value: "done", label: "已处理" },
]

function sourceLabel(source: string) {
  switch (source) {
    case "manual":
      return "手动创建"
    case "conversation":
      return "会话生成"
    default:
      return source || "-"
  }
}

function metadataValue(value?: string | number | null) {
  if (value === undefined || value === null || value === "") {
    return "-"
  }
  return String(value)
}

export function TicketDetailDialog({
  ticketId,
  open,
  onOpenChange,
  onChanged,
}: TicketDetailDialogProps) {
  const [detail, setDetail] = useState<TicketDetail | null>(null)
  const [loading, setLoading] = useState(false)
  const [statusSaving, setStatusSaving] = useState<TicketStatus | null>(null)
  const [progressSaving, setProgressSaving] = useState(false)
  const [progressContent, setProgressContent] = useState("")
  const [assignOpen, setAssignOpen] = useState(false)
  const [editOpen, setEditOpen] = useState(false)
  const [editSaving, setEditSaving] = useState(false)

  const loadDetail = useCallback(async () => {
    if (!open || !ticketId) {
      setDetail(null)
      return
    }
    setLoading(true)
    try {
      const data = await fetchTicketDetail(ticketId)
      setDetail(data)
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "加载工单详情失败")
    } finally {
      setLoading(false)
    }
  }, [open, ticketId])

  useEffect(() => {
    void loadDetail()
  }, [loadDetail])

  async function handleStatusChange(status: TicketStatus) {
    if (!detail || detail.ticket.status === status) {
      return
    }
    setStatusSaving(status)
    try {
      await changeTicketStatus({ ticketId: detail.ticket.id, status })
      toast.success("工单状态已更新")
      await loadDetail()
      onChanged()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "更新工单状态失败")
    } finally {
      setStatusSaving(null)
    }
  }

  async function handleCreateProgress() {
    if (!detail) {
      return
    }
    const content = progressContent.trim()
    if (!content) {
      toast.error("请填写处理进展")
      return
    }
    setProgressSaving(true)
    try {
      await createTicketProgress({
        ticketId: detail.ticket.id,
        content,
      })
      toast.success("处理进展已记录")
      setProgressContent("")
      await loadDetail()
      onChanged()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "记录处理进展失败")
    } finally {
      setProgressSaving(false)
    }
  }

  async function handleAssigned() {
    await loadDetail()
    onChanged()
  }

  async function handleUpdateTicket(payload: CreateTicketPayload | UpdateTicketPayload) {
    if (!("ticketId" in payload) || payload.ticketId <= 0) {
      toast.error("请选择工单")
      return
    }
    setEditSaving(true)
    try {
      await updateTicket(payload)
      toast.success("工单已更新")
      setEditOpen(false)
      await loadDetail()
      onChanged()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "更新工单失败")
    } finally {
      setEditSaving(false)
    }
  }

  const ticket = detail?.ticket

  return (
    <>
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent className="max-h-[88vh] gap-0 p-0 sm:max-w-3xl">
          <DialogHeader className="border-b px-6 py-4">
            <DialogTitle className="flex min-w-0 items-center gap-2 text-base">
              <span className="truncate">{ticket?.title ?? "工单详情"}</span>
              {ticket ? <TicketStatusBadge status={ticket.status} /> : null}
            </DialogTitle>
            {ticket ? (
              <div className="mt-2 flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
                <span className="font-mono">{ticket.ticketNo}</span>
                <span>{sourceLabel(ticket.source)}</span>
                <span>创建人：{metadataValue(ticket.createdByName || ticket.createdBy)}</span>
              </div>
            ) : null}
          </DialogHeader>

          <ScrollArea className="max-h-[calc(88vh-74px)]">
            {loading && !ticket ? (
              <div className="flex items-center justify-center gap-2 px-6 py-12 text-sm text-muted-foreground">
                <RefreshCcwIcon className="size-4 animate-spin" />
                加载中...
              </div>
            ) : ticket ? (
              <div className="space-y-5 p-6">
                <section className="space-y-2">
                  <div className="text-xs font-medium text-muted-foreground">描述</div>
                  <div className="whitespace-pre-wrap rounded-md border bg-muted/30 px-3 py-2 text-sm leading-6">
                    {ticket.description || "暂无描述"}
                  </div>
                </section>

                <section className="grid gap-4 md:grid-cols-[minmax(0,1fr)_220px]">
                  <div className="space-y-2">
                    <div className="text-xs font-medium text-muted-foreground">状态</div>
                    <div className="flex flex-wrap gap-2">
                      {statusOptions.map((option) => (
                        <Button
                          key={option.value}
                          type="button"
                          size="sm"
                          variant={ticket.status === option.value ? "default" : "outline"}
                          disabled={!!statusSaving}
                          onClick={() => void handleStatusChange(option.value)}
                        >
                          {statusSaving === option.value ? "更新中..." : option.label}
                        </Button>
                      ))}
                    </div>
                  </div>
                  <div className="space-y-2">
                    <div className="text-xs font-medium text-muted-foreground">负责人</div>
                    <div className="flex items-center justify-between gap-2 rounded-md border px-3 py-2">
                      <div className="flex min-w-0 items-center gap-2 text-sm">
                        <UserRoundIcon className="size-4 shrink-0 text-muted-foreground" />
                        <span className="truncate">{ticket.currentAssigneeName || "未分配"}</span>
                      </div>
                      <Button type="button" size="sm" variant="outline" onClick={() => setAssignOpen(true)}>
                        指派
                      </Button>
                    </div>
                  </div>
                </section>

                <section className="space-y-2">
                  <div className="flex items-center justify-between gap-2">
                    <div className="text-xs font-medium text-muted-foreground">标签</div>
                    <Button type="button" size="sm" variant="outline" onClick={() => setEditOpen(true)}>
                      编辑
                    </Button>
                  </div>
                  {ticket.tags && ticket.tags.length > 0 ? (
                    <div className="flex flex-wrap gap-1.5">
                      {ticket.tags.map((tag) => (
                        <Badge key={tag.id} variant="outline">
                          {tag.name}
                        </Badge>
                      ))}
                    </div>
                  ) : (
                    <div className="text-sm text-muted-foreground">暂无标签</div>
                  )}
                </section>

                <section className="grid gap-3 rounded-md border p-3 text-sm sm:grid-cols-2">
                  <MetadataItem label="客户" value={ticket.customer?.name || ticket.customerId} />
                  <MetadataItem label="联系方式" value={ticket.customer?.primaryMobile || ticket.customer?.primaryEmail} />
                  <MetadataItem label="来源" value={sourceLabel(ticket.source)} />
                  <MetadataItem label="渠道" value={ticket.channel} />
                  <MetadataItem label="会话 ID" value={ticket.conversationId || undefined} />
                  <MetadataItem label="最后更新" value={ticket.updatedAt ? formatDateTime(ticket.updatedAt) : undefined} />
                </section>

                <Separator />

                <section className="space-y-3">
                  <div className="flex items-center gap-2 text-sm font-medium">
                    <MessageSquareTextIcon className="size-4 text-muted-foreground" />
                    处理进展
                  </div>
                  <div className="space-y-2">
                    <Textarea
                      rows={3}
                      placeholder="记录本次处理进展"
                      value={progressContent}
                      onChange={(event) => setProgressContent(event.target.value)}
                    />
                    <div className="flex justify-end">
                      <Button type="button" size="sm" disabled={progressSaving} onClick={() => void handleCreateProgress()}>
                        <SendIcon className="size-3.5" />
                        {progressSaving ? "提交中..." : "添加进展"}
                      </Button>
                    </div>
                  </div>

                  {detail.progresses && detail.progresses.length > 0 ? (
                    <div className="space-y-3">
                      {detail.progresses.map((progress, index) => (
                        <div key={progress.id} className="flex gap-3">
                          <div className="flex flex-col items-center">
                            <span className="mt-1 size-2 rounded-full bg-primary" />
                            <span
                              className={cn(
                                "mt-1 w-px flex-1 bg-border",
                                index === detail.progresses!.length - 1 ? "opacity-0" : "opacity-100",
                              )}
                            />
                          </div>
                          <div className="min-w-0 flex-1 pb-3">
                            <div className="flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
                              <span>{progress.authorName || `用户#${progress.authorId}`}</span>
                              <span>{progress.createdAt ? formatDateTime(progress.createdAt) : "-"}</span>
                            </div>
                            <div className="mt-1 whitespace-pre-wrap text-sm leading-6">{progress.content}</div>
                          </div>
                        </div>
                      ))}
                    </div>
                  ) : (
                    <div className="rounded-md border border-dashed px-3 py-6 text-center text-sm text-muted-foreground">
                      暂无处理进展
                    </div>
                  )}
                </section>
              </div>
            ) : (
              <div className="px-6 py-12 text-center text-sm text-muted-foreground">请选择工单</div>
            )}
          </ScrollArea>
        </DialogContent>
      </Dialog>

      <TicketAssignDialog
        open={assignOpen}
        ticketId={ticket?.id ?? null}
        currentAssigneeId={ticket?.currentAssigneeId}
        onOpenChange={setAssignOpen}
        onSuccess={handleAssigned}
      />
      <EditDialog
        open={editOpen}
        saving={editSaving}
        itemId={ticket?.id ?? null}
        onOpenChange={setEditOpen}
        onSubmit={handleUpdateTicket}
      />
    </>
  )
}

function MetadataItem({ label, value }: { label: string; value?: string | number | null }) {
  return (
    <div className="min-w-0">
      <div className="text-xs text-muted-foreground">{label}</div>
      <div className="mt-1 truncate">{metadataValue(value)}</div>
    </div>
  )
}
