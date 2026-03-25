"use client"

import { useCallback, useEffect, useState } from "react"
import {
  CalendarClockIcon,
  ClockIcon,
  RefreshCwIcon,
  SendIcon,
  TagIcon,
  TicketIcon,
  Trash2Icon,
  UserIcon,
} from "lucide-react"
import { toast } from "sonner"

import {
  closeTicket,
  createTicketComment,
  createTicketReply,
  deleteTicket,
  deleteTicketComment,
  fetchTicketAssignments,
  fetchTicketComments,
  fetchTicketDetail,
  fetchTicketReplies,
  reopenTicket,
  type Ticket,
  type TicketAssignment,
  type TicketComment,
  type TicketReply,
} from "@/lib/api/admin"
import { formatDateTime } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import {
  Drawer,
  DrawerContent,
  DrawerDescription,
  DrawerFooter,
  DrawerHeader,
  DrawerTitle,
} from "@/components/ui/drawer"
import { Input } from "@/components/ui/input"
import { Separator } from "@/components/ui/separator"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"

type TicketDetailDrawerProps = {
  open: boolean
  ticketId: number | null
  onOpenChange: (open: boolean) => void
  onRefresh: () => void
}

function getStatusMeta(status: number) {
  const map: Record<number, { label: string; variant: "secondary" | "outline" | "default" | "destructive" }> = {
    1: { label: "待处理", variant: "outline" },
    2: { label: "处理中", variant: "secondary" },
    3: { label: "待确认", variant: "secondary" },
    4: { label: "已解决", variant: "default" },
    5: { label: "已关闭", variant: "outline" },
    6: { label: "已取消", variant: "outline" },
  }
  return map[status] ?? { label: "未知", variant: "outline" }
}

function getPriorityLabel(priority: number) {
  const labels = ["普通", "低", "中", "高", "紧急"]
  return labels[priority] ?? "普通"
}

function formatTime(timestamp: number) {
  if (!timestamp) return "-"
  return formatDateTime(timestamp)
}

export function TicketDetailDrawer({
  open,
  ticketId,
  onOpenChange,
  onRefresh,
}: TicketDetailDrawerProps) {
  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState(false)
  const [ticket, setTicket] = useState<Ticket | null>(null)
  const [replies, setReplies] = useState<TicketReply[]>([])
  const [comments, setComments] = useState<TicketComment[]>([])
  const [assignments, setAssignments] = useState<TicketAssignment[]>([])
  const [replyContent, setReplyContent] = useState("")
  const [commentContent, setCommentContent] = useState("")

  const loadTicket = useCallback(async () => {
    if (!ticketId) return
    setLoading(true)
    try {
      const [detail, repliesData, commentsData, assignmentsData] = await Promise.all([
        fetchTicketDetail(ticketId),
        fetchTicketReplies(ticketId).then(r => r.results),
        fetchTicketComments(ticketId).then(r => r.results),
        fetchTicketAssignments(ticketId).then(r => r.results),
      ])
      setTicket(detail)
      setReplies(repliesData)
      setComments(commentsData)
      setAssignments(assignmentsData)
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "加载工单详情失败")
    } finally {
      setLoading(false)
    }
  }, [ticketId])

  useEffect(() => {
    if (open && ticketId) {
      void loadTicket()
    }
  }, [open, ticketId, loadTicket])

  async function handleCloseTicket() {
    if (!ticket) return
    setSaving(true)
    try {
      await closeTicket(ticket.id)
      toast.success("工单已关闭")
      await loadTicket()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "关闭工单失败")
    } finally {
      setSaving(false)
    }
  }

  async function handleReopenTicket() {
    if (!ticket) return
    setSaving(true)
    try {
      await reopenTicket(ticket.id)
      toast.success("工单已重开")
      await loadTicket()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "重开工单失败")
    } finally {
      setSaving(false)
    }
  }

  async function handleDeleteTicket() {
    if (!ticket) return
    setSaving(true)
    try {
      await deleteTicket(ticket.id)
      toast.success("工单已删除")
      onOpenChange(false)
      onRefresh()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "删除工单失败")
    } finally {
      setSaving(false)
    }
  }

  async function handleSendReply() {
    if (!ticket || !replyContent.trim()) return
    setSaving(true)
    try {
      await createTicketReply({
        ticketId: ticket.id,
        parentId: 0,
        content: replyContent,
        isInternal: false,
        attachmentIds: "",
      })
      toast.success("回复已发送")
      setReplyContent("")
      await loadTicket()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "发送回复失败")
    } finally {
      setSaving(false)
    }
  }

  async function handleAddComment() {
    if (!ticket || !commentContent.trim()) return
    setSaving(true)
    try {
      await createTicketComment({
        ticketId: ticket.id,
        content: commentContent,
      })
      toast.success("备注已添加")
      setCommentContent("")
      await loadTicket()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "添加备注失败")
    } finally {
      setSaving(false)
    }
  }

  async function handleDeleteComment(id: number) {
    setSaving(true)
    try {
      await deleteTicketComment(id)
      toast.success("备注已删除")
      await loadTicket()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "删除备注失败")
    } finally {
      setSaving(false)
    }
  }

  if (!open) return null

  const statusMeta = ticket ? getStatusMeta(ticket.status) : { label: "加载中...", variant: "outline" as const }

  return (
    <Drawer open={open} onOpenChange={onOpenChange} direction="right">
      <DrawerContent className="max-w-xl">
        <DrawerHeader>
          <DrawerTitle className="flex items-center gap-2">
            <TicketIcon className="size-5" />
            {ticket?.title || "工单详情"}
          </DrawerTitle>
          <DrawerDescription>
            工单编号：{ticket?.ticketNo || "-"}
          </DrawerDescription>
        </DrawerHeader>

        <div className="space-y-5 overflow-y-auto px-4 pb-4">
          <div className="rounded-lg border p-4">
            <div className="grid grid-cols-2 gap-4 text-sm">
              <div>
                <div className="text-muted-foreground">状态</div>
                <div className="font-medium">{statusMeta.label}</div>
              </div>
              <div>
                <div className="text-muted-foreground">优先级</div>
                <div className="font-medium">{ticket ? getPriorityLabel(ticket.priority) : "-"}</div>
              </div>
              <div>
                <div className="text-muted-foreground">客户名称</div>
                <div className="font-medium">{ticket?.externalUserName || "-"}</div>
              </div>
              <div>
                <div className="text-muted-foreground">客户邮箱</div>
                <div className="font-medium">{ticket?.externalUserEmail || "-"}</div>
              </div>
              <div>
                <div className="text-muted-foreground">客户手机</div>
                <div className="font-medium">{ticket?.externalUserMobile || "-"}</div>
              </div>
              <div>
                <div className="text-muted-foreground">创建时间</div>
                <div className="font-medium">{ticket ? formatTime(ticket.createdAt) : "-"}</div>
              </div>
            </div>
            {ticket?.content && (
              <>
                <Separator className="my-4" />
                <div>
                  <div className="text-muted-foreground mb-2">工单内容</div>
                  <div className="text-sm whitespace-pre-wrap">{ticket.content}</div>
                </div>
              </>
            )}
            {ticket?.tags && (
              <>
                <Separator className="my-4" />
                <div className="flex items-center gap-2">
                  <TagIcon className="size-4 text-muted-foreground" />
                  <span className="text-sm text-muted-foreground">{ticket.tags}</span>
                </div>
              </>
            )}
          </div>

          <Tabs defaultValue="replies" className="w-full">
            <TabsList className="grid w-full grid-cols-3">
              <TabsTrigger value="replies">
                回复 ({replies.length})
              </TabsTrigger>
              <TabsTrigger value="comments">
                备注 ({comments.length})
              </TabsTrigger>
              <TabsTrigger value="history">
                历史 ({assignments.length})
              </TabsTrigger>
            </TabsList>

            <TabsContent value="replies" className="mt-4 space-y-4">
              <div className="space-y-3">
                {replies.map((reply) => (
                  <div key={reply.id} className="rounded-lg border p-3">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2 text-sm">
                        <UserIcon className="size-4 text-muted-foreground" />
                        <span className="font-medium">{reply.senderName}</span>
                        <span className="text-muted-foreground">
                          ({reply.senderType === "agent" ? "客服" : "客户"})
                        </span>
                      </div>
                      <span className="text-xs text-muted-foreground">
                        {formatTime(reply.createdAt)}
                      </span>
                    </div>
                    <div className="mt-2 text-sm whitespace-pre-wrap">{reply.content}</div>
                  </div>
                ))}
                {replies.length === 0 && (
                  <div className="text-center text-muted-foreground py-8">
                    暂无回复
                  </div>
                )}
              </div>
              <div className="flex gap-2">
                <Input
                  placeholder="输入回复内容..."
                  value={replyContent}
                  onChange={(e) => setReplyContent(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === "Enter" && !e.shiftKey) {
                      e.preventDefault()
                      void handleSendReply()
                    }
                  }}
                />
                <Button
                  onClick={() => void handleSendReply()}
                  disabled={saving || !replyContent.trim()}
                >
                  <SendIcon className="size-4" />
                </Button>
              </div>
            </TabsContent>

            <TabsContent value="comments" className="mt-4 space-y-4">
              <div className="space-y-3">
                {comments.map((comment) => (
                  <div key={comment.id} className="rounded-lg border p-3">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2 text-sm">
                        <UserIcon className="size-4 text-muted-foreground" />
                        <span className="font-medium">{comment.userName}</span>
                      </div>
                      <div className="flex items-center gap-2">
                        <span className="text-xs text-muted-foreground">
                          {formatTime(comment.createdAt)}
                        </span>
                        <Button
                          variant="ghost"
                          size="icon-sm"
                          onClick={() => void handleDeleteComment(comment.id)}
                        >
                          <Trash2Icon className="size-4 text-destructive" />
                        </Button>
                      </div>
                    </div>
                    <div className="mt-2 text-sm whitespace-pre-wrap">{comment.content}</div>
                  </div>
                ))}
                {comments.length === 0 && (
                  <div className="text-center text-muted-foreground py-8">
                    暂无备注
                  </div>
                )}
              </div>
              <div className="flex gap-2">
                <Input
                  placeholder="添加内部备注..."
                  value={commentContent}
                  onChange={(e) => setCommentContent(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === "Enter" && !e.shiftKey) {
                      e.preventDefault()
                      void handleAddComment()
                    }
                  }}
                />
                <Button
                  onClick={() => void handleAddComment()}
                  disabled={saving || !commentContent.trim()}
                  variant="outline"
                >
                  <SendIcon className="size-4" />
                </Button>
              </div>
            </TabsContent>

            <TabsContent value="history" className="mt-4">
              <div className="space-y-3">
                {assignments.map((assignment) => (
                  <div key={assignment.id} className="rounded-lg border p-3">
                    <div className="flex items-center gap-2 text-sm">
                      <CalendarClockIcon className="size-4 text-muted-foreground" />
                      <span className="font-medium">
                        {assignment.assignType === "distribute" ? "分配" : assignment.assignType}
                      </span>
                      <span className="text-muted-foreground">→</span>
                      <span>User:{assignment.toUserId} / Team:{assignment.toTeamId}</span>
                    </div>
                    {assignment.reason && (
                      <div className="mt-1 text-sm text-muted-foreground">
                        原因：{assignment.reason}
                      </div>
                    )}
                    <div className="mt-1 text-xs text-muted-foreground">
                      {formatTime(assignment.createdAt)}
                    </div>
                  </div>
                ))}
                {assignments.length === 0 && (
                  <div className="text-center text-muted-foreground py-8">
                    暂无分配历史
                  </div>
                )}
              </div>
            </TabsContent>
          </Tabs>
        </div>

        <DrawerFooter>
          <div className="flex gap-2">
            {ticket?.status === 5 ? (
              <Button onClick={() => void handleReopenTicket()} disabled={saving}>
                <RefreshCwIcon className="size-4" />
                重开工单
              </Button>
            ) : (
              <Button onClick={() => void handleCloseTicket()} disabled={saving} variant="outline">
                <ClockIcon className="size-4" />
                关闭工单
              </Button>
            )}
            <Button
              onClick={() => void handleDeleteTicket()}
              disabled={saving}
              variant="destructive"
              className="ml-auto"
            >
              <Trash2Icon className="size-4" />
              删除工单
            </Button>
          </div>
        </DrawerFooter>
      </DrawerContent>
    </Drawer>
  )
}
