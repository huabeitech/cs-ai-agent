"use client"

import Link from "next/link"
import {
  ExternalLinkIcon,
  ArrowLeftIcon,
  MessageSquarePlusIcon,
  PlusIcon,
  RefreshCcwIcon,
  RotateCcwIcon,
  SaveIcon,
  Settings2Icon,
  UserRoundPlusIcon,
  XIcon,
} from "lucide-react"
import { useRouter, useSearchParams } from "next/navigation"
import { useCallback, useEffect, useMemo, useState } from "react"
import { toast } from "sonner"

import { OptionCombobox } from "@/components/option-combobox"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Textarea } from "@/components/ui/textarea"
import {
  deleteTicketCollaborator,
  deleteTicketRelation,
  addTicketInternalNote,
  fetchTicketDetail,
  replyTicket,
  unwatchTicket,
  type TicketDetail,
  type TicketItem,
  type CreateTicketPayload,
  type UpdateTicketPayload,
  updateTicket,
  watchTicket,
} from "@/lib/api/ticket"
import {
  fetchAgentProfilesAll,
  fetchConversationDetail,
  type AdminAgentProfile,
  type AdminConversationDetail,
} from "@/lib/api/admin"
import { readSession } from "@/lib/auth"
import { getEnumLabel } from "@/lib/enums"
import { IMConversationStatus, IMConversationStatusLabels } from "@/lib/generated/enums"
import { formatDateTime } from "@/lib/utils"
import { EditDialog } from "../_components/edit"
import { TicketAssignDialog } from "../_components/ticket-assign-dialog"
import { TicketPriorityBadge } from "../_components/ticket-priority-badge"
import { TicketReasonDialog } from "../_components/ticket-reason-dialog"
import { TicketCollaboratorDialog } from "../_components/ticket-collaborator-dialog"
import { TicketRelationDialog } from "../_components/ticket-relation-dialog"
import { TicketSLABadge } from "../_components/ticket-sla-badge"
import { TicketStatusDialog } from "../_components/ticket-status-dialog"
import {
  TicketStatusBadge,
  ticketStatusLabel,
} from "../_components/ticket-status-badge"

function formatTicketSource(source?: string) {
  switch (source) {
    case "conversation":
      return "会话转工单"
    case "manual":
      return "手动创建"
    case "email":
      return "邮件"
    case "api":
      return "API"
    case "system":
      return "系统"
    default:
      return source || "—"
  }
}

function formatSLAStatus(status: string) {
  switch (status) {
    case "running":
      return "进行中"
    case "paused":
      return "已暂停"
    case "completed":
      return "已完成"
    case "breached":
      return "已超时"
    default:
      return status
  }
}

function ticketSeverityLabel(severity: number) {
  switch (severity) {
    case 1:
      return "轻微"
    case 2:
      return "一般"
    case 3:
      return "严重"
    default:
      return String(severity)
  }
}

function isClosedStatus(status: string) {
  return status === "resolved" || status === "closed" || status === "cancelled"
}

function ticketRelationLabel(relationType?: string) {
  switch (relationType) {
    case "duplicate":
      return "重复工单"
    case "related":
      return "相关工单"
    case "parent":
      return "父工单"
    case "child":
      return "子工单"
    default:
      return relationType || "关联工单"
  }
}

function ticketEventLabel(eventType?: string) {
  switch (eventType) {
    case "mentioned":
      return "提及协作人"
    case "internal_noted":
      return "内部备注"
    case "replied":
      return "回复客户"
    case "status_changed":
      return "状态变更"
    case "assigned":
      return "指派工单"
    case "transferred":
      return "转派工单"
    default:
      return eventType || "事件"
  }
}

function getMainSLA(ticket: TicketItem | null) {
  return ticket?.sla?.find((item) => item.slaType === "resolution") ?? ticket?.sla?.[0] ?? null
}

function parseMentionUserIds(payload?: string) {
  if (!payload) {
    return []
  }
  try {
    const data = JSON.parse(payload) as { mentionUserIds?: number[] }
    return Array.isArray(data.mentionUserIds) ? data.mentionUserIds : []
  } catch {
    return []
  }
}

function isDoneTicketStatus(status?: string) {
  return status === "resolved" || status === "closed" || status === "cancelled"
}

export default function TicketDetailPage() {
  const searchParams = useSearchParams()
  const router = useRouter()
  const ticketId = Number(searchParams.get("id") || 0)
  const [detail, setDetail] = useState<TicketDetail | null>(null)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [replyMode, setReplyMode] = useState<"public" | "internal">("public")
  const [replyContent, setReplyContent] = useState("")
  const [assignDialogOpen, setAssignDialogOpen] = useState(false)
  const [statusDialogOpen, setStatusDialogOpen] = useState(false)
  const [closeDialogOpen, setCloseDialogOpen] = useState(false)
  const [reopenDialogOpen, setReopenDialogOpen] = useState(false)
  const [editDialogOpen, setEditDialogOpen] = useState(false)
  const [relationDialogOpen, setRelationDialogOpen] = useState(false)
  const [collaboratorDialogOpen, setCollaboratorDialogOpen] = useState(false)
  const [sourceConversation, setSourceConversation] = useState<AdminConversationDetail | null>(null)
  const [agents, setAgents] = useState<AdminAgentProfile[]>([])
  const [mentionUserId, setMentionUserId] = useState("")
  const [mentionedUsers, setMentionedUsers] = useState<AdminAgentProfile[]>([])

  const ticket = detail?.ticket ?? null
  const currentUserId = readSession()?.user?.id ?? 0
  const isWatching = Boolean(detail?.watchers?.some((item) => item.userId === currentUserId))
  const resolutionSLA = useMemo(() => getMainSLA(ticket), [ticket])
  const childRelations = useMemo(
    () => (detail?.relatedTickets || []).filter((item) => item.relationType === "child"),
    [detail?.relatedTickets],
  )
  const childProgress = useMemo(() => {
    const total = childRelations.length
    const completed = childRelations.filter((item) => isDoneTicketStatus(item.relatedTicketStatus)).length
    return {
      total,
      completed,
      active: total - completed,
    }
  }, [childRelations])

  const loadDetail = useCallback(async () => {
    if (!ticketId) {
      setDetail(null)
      setLoading(false)
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
  }, [ticketId])

  useEffect(() => {
    void loadDetail()
  }, [loadDetail])

  useEffect(() => {
    void (async () => {
      try {
        const data = await fetchAgentProfilesAll()
        setAgents(Array.isArray(data) ? data : [])
      } catch {
        setAgents([])
      }
    })()
  }, [])

  useEffect(() => {
    if (!ticket?.conversationId) {
      setSourceConversation(null)
      return
    }
    void (async () => {
      try {
        const data = await fetchConversationDetail(ticket.conversationId)
        setSourceConversation(data)
      } catch {
        setSourceConversation(null)
      }
    })()
  }, [ticket?.conversationId])

  async function handleReplySubmit() {
    if (!ticket || !replyContent.trim()) {
      toast.error(replyMode === "public" ? "回复内容不能为空" : "备注内容不能为空")
      return
    }
    setSaving(true)
    try {
      if (replyMode === "public") {
        await replyTicket({
          ticketId: ticket.id,
          contentType: "text",
          content: replyContent.trim(),
        })
        toast.success("已回复客户")
      } else {
        const payload =
          mentionedUsers.length > 0
            ? JSON.stringify({
                mentionUserIds: mentionedUsers.map((item) => item.userId),
              })
            : undefined
        await addTicketInternalNote({
          ticketId: ticket.id,
          contentType: "text",
          content: replyContent.trim(),
          payload,
        })
        toast.success("已添加内部备注")
      }
      setReplyContent("")
      setMentionUserId("")
      setMentionedUsers([])
      await loadDetail()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "提交失败")
    } finally {
      setSaving(false)
    }
  }

  const mentionOptions = useMemo(
    () =>
      agents.map((agent) => ({
        value: String(agent.userId),
        label: agent.displayName || agent.nickname || agent.username || `客服 #${agent.userId}`,
      })),
    [agents],
  )

  function handleAddMentionUser() {
    const userId = Number(mentionUserId)
    if (!userId) {
      return
    }
    const user = agents.find((item) => item.userId === userId)
    if (!user) {
      return
    }
    setMentionedUsers((current) => {
      if (current.some((item) => item.userId === user.userId)) {
        return current
      }
      return [...current, user]
    })
    setMentionUserId("")
  }

  async function handleWatchToggle() {
    if (!ticket) {
      return
    }
    setSaving(true)
    try {
      if (isWatching) {
        await unwatchTicket(ticket.id)
        toast.success("已取消关注")
      } else {
        await watchTicket(ticket.id)
        toast.success("已关注工单")
      }
      await loadDetail()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "更新关注状态失败")
    } finally {
      setSaving(false)
    }
  }

  async function handleEditSubmit(payload: CreateTicketPayload | UpdateTicketPayload) {
    if (!("ticketId" in payload)) {
      toast.error("详情页仅支持编辑现有工单")
      return
    }
    setSaving(true)
    try {
      await updateTicket(payload)
      toast.success("工单基础信息已更新")
      setEditDialogOpen(false)
      await loadDetail()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "更新工单失败")
    } finally {
      setSaving(false)
    }
  }

  async function handleDeleteRelation(relationId: number) {
    if (!ticket) {
      return
    }
    setSaving(true)
    try {
      await deleteTicketRelation({ ticketId: ticket.id, relationId })
      toast.success("关联工单已移除")
      await loadDetail()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "移除关联工单失败")
    } finally {
      setSaving(false)
    }
  }

  async function handleDeleteCollaborator(collaboratorId: number) {
    if (!ticket) {
      return
    }
    setSaving(true)
    try {
      await deleteTicketCollaborator({ ticketId: ticket.id, collaboratorId })
      toast.success("协作人已移除")
      await loadDetail()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "移除协作人失败")
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="min-h-0 flex-1 overflow-auto bg-muted/20 p-4 md:p-6">
      <div className="flex w-full flex-col gap-4">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div className="flex items-center gap-3">
            <Link href="/tickets">
              <Button variant="outline" size="sm">
                <ArrowLeftIcon className="size-4" />
                返回工单列表
              </Button>
            </Link>
            {ticket ? (
              <div>
                <div className="text-xs text-muted-foreground">{ticket.ticketNo}</div>
                <div className="text-lg font-semibold">{ticket.title}</div>
              </div>
            ) : null}
          </div>
          <div className="flex flex-wrap gap-2">
            <Link href="/ticket-categories">
              <Button variant="outline">
                <Settings2Icon className="size-4" />
                工单配置
              </Button>
            </Link>
            <Button
              variant="outline"
              onClick={() => void handleWatchToggle()}
              disabled={saving || !ticket}
            >
              {isWatching ? "取消关注" : "关注工单"}
            </Button>
            <Button variant="outline" onClick={() => void loadDetail()} disabled={loading || saving}>
              <RefreshCcwIcon className="size-4" />
              刷新
            </Button>
            {ticket?.status === "closed" ? (
              <Button onClick={() => setReopenDialogOpen(true)} disabled={saving}>
                重开工单
              </Button>
            ) : (
              <Button variant="outline" onClick={() => setCloseDialogOpen(true)} disabled={saving || !ticket}>
                关闭工单
              </Button>
            )}
          </div>
        </div>

        {loading ? (
          <Card>
            <CardContent className="py-20 text-center text-muted-foreground">加载中...</CardContent>
          </Card>
        ) : ticket ? (
          <div className="grid gap-4 lg:grid-cols-[minmax(0,1.75fr)_380px]">
            <div className="space-y-4">
              <Card>
                <CardHeader>
                  <div className="space-y-3">
                    <div className="flex flex-wrap items-center gap-2">
                      <TicketStatusBadge status={ticket.status} />
                      <TicketPriorityBadge priority={ticket.priority} />
                      <TicketSLABadge ticket={ticket} />
                    </div>
                    <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
                      <SummaryMetric label="分类" value={ticket.categoryName || "未分类"} />
                      <SummaryMetric label="严重度" value={ticketSeverityLabel(ticket.severity)} />
                      <SummaryMetric label="处理人" value={ticket.currentAssigneeName || "未指派"} />
                      <SummaryMetric
                        label="解决时限"
                        value={
                          resolutionSLA
                            ? `${resolutionSLA.targetMinutes} 分钟 / ${formatSLAStatus(resolutionSLA.status)}`
                            : "未设置"
                        }
                      />
                    </div>
                  </div>
                <CardAction className="flex w-full flex-wrap justify-end gap-2 pt-3 sm:w-auto sm:pt-0">
                  <Link href="/ticket-categories">
                    <Button className="min-w-28" variant="ghost">管理分类</Button>
                  </Link>
                  <Button className="min-w-28" variant="outline" onClick={() => setEditDialogOpen(true)}>
                    编辑基础信息
                  </Button>
                    <Button className="min-w-28" variant="outline" onClick={() => setAssignDialogOpen(true)}>
                      <UserRoundPlusIcon className="size-4" />
                      分配处理人
                    </Button>
                    <Button className="min-w-28" variant="outline" onClick={() => setStatusDialogOpen(true)}>
                      <SaveIcon className="size-4" />
                      变更状态
                    </Button>
                  </CardAction>
                </CardHeader>
                <CardContent>
                  <div className="rounded-lg border bg-muted/20 p-4 text-sm leading-7 text-muted-foreground">
                    {ticket.description || "暂无工单描述"}
                  </div>
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle className="text-base">回复与备注</CardTitle>
                  <CardDescription>在详情页内完成客户回复和内部协作记录</CardDescription>
                </CardHeader>
                <CardContent className="space-y-3">
                  <div className="flex gap-2">
                    <Button
                      variant={replyMode === "public" ? "default" : "outline"}
                      onClick={() => setReplyMode("public")}
                    >
                      回复客户
                    </Button>
                    <Button
                      variant={replyMode === "internal" ? "default" : "outline"}
                      onClick={() => setReplyMode("internal")}
                    >
                      内部备注
                    </Button>
                  </div>
                  <Textarea
                    rows={5}
                    value={replyContent}
                    placeholder={replyMode === "public" ? "输入给客户的回复内容" : "输入内部备注"}
                    onChange={(event) => setReplyContent(event.target.value)}
                  />
                  {replyMode === "internal" ? (
                    <div className="space-y-3 rounded-lg border bg-muted/20 p-3">
                      <div className="text-sm font-medium">@提及协作人</div>
                      <div className="flex gap-2">
                        <div className="flex-1">
                          <OptionCombobox
                            value={mentionUserId}
                            options={mentionOptions}
                            placeholder="选择要提及的客服"
                            searchPlaceholder="搜索客服"
                            emptyText="暂无可选客服"
                            onChange={setMentionUserId}
                          />
                        </div>
                        <Button type="button" variant="outline" onClick={handleAddMentionUser}>
                          添加
                        </Button>
                      </div>
                      {mentionedUsers.length ? (
                        <div className="flex flex-wrap gap-2">
                          {mentionedUsers.map((user) => (
                            <button
                              key={user.userId}
                              type="button"
                              className="rounded-full border px-3 py-1 text-xs"
                              onClick={() =>
                                setMentionedUsers((current) => current.filter((item) => item.userId !== user.userId))
                              }
                            >
                              @{user.displayName || user.nickname || user.username || `客服#${user.userId}`} ×
                            </button>
                          ))}
                        </div>
                      ) : (
                        <div className="text-xs text-muted-foreground">未添加提及对象</div>
                      )}
                    </div>
                  ) : null}
                  <div className="flex justify-end">
                    <Button onClick={() => void handleReplySubmit()} disabled={saving}>
                      <MessageSquarePlusIcon className="size-4" />
                      {replyMode === "public" ? "发送回复" : "保存备注"}
                    </Button>
                  </div>
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle className="text-base">处理记录</CardTitle>
                  <CardDescription>沟通记录和状态流转分开展示，便于排查</CardDescription>
                </CardHeader>
                <CardContent>
                  <Tabs defaultValue="comments" className="gap-4">
                    <TabsList variant="line">
                      <TabsTrigger value="comments">回复记录</TabsTrigger>
                      <TabsTrigger value="events">事件记录</TabsTrigger>
                    </TabsList>
                    <TabsContent value="comments" className="space-y-3">
                      {(detail?.comments?.length || 0) > 0 ? (
                        detail?.comments?.map((comment) => (
                          <div key={`comment-${comment.id}`} className="rounded-lg border p-3">
                            <div className="mb-1 flex items-center justify-between gap-3">
                              <div className="text-sm font-medium">
                                {comment.authorName || `用户#${comment.authorId}`}
                              </div>
                              <div className="text-xs text-muted-foreground">
                                {comment.createdAt ? formatDateTime(comment.createdAt) : "—"}
                              </div>
                            </div>
                            <div className="mb-2 text-xs text-muted-foreground">
                              {comment.commentType === "public_reply" ? "客户可见回复" : "内部备注"}
                            </div>
                            {comment.commentType === "internal_note" && parseMentionUserIds(comment.payload).length ? (
                              <div className="mb-2 flex flex-wrap gap-2">
                                {parseMentionUserIds(comment.payload).map((userId) => {
                                  const user = agents.find((item) => item.userId === userId)
                                  return (
                                    <span key={`${comment.id}-${userId}`} className="rounded-full border px-2 py-1 text-xs text-muted-foreground">
                                      @{user?.displayName || user?.nickname || user?.username || `用户#${userId}`}
                                    </span>
                                  )
                                })}
                              </div>
                            ) : null}
                            <div className="whitespace-pre-wrap text-sm leading-6">{comment.content}</div>
                          </div>
                        ))
                      ) : (
                        <div className="text-sm text-muted-foreground">暂无评论记录</div>
                      )}
                    </TabsContent>
                    <TabsContent value="events" className="space-y-2">
                      {(detail?.events?.length || 0) > 0 ? (
                        detail?.events?.map((event) => (
                          <div
                            key={`event-${event.id}`}
                            className={`rounded-lg border px-3 py-2 ${
                              event.eventType === "mentioned" ? "border-amber-200 bg-amber-50/60" : ""
                            }`}
                          >
                            <div className="flex items-center justify-between gap-3">
                              <div className="flex items-center gap-2">
                                {event.eventType === "mentioned" ? (
                                  <Badge variant="outline" className="border-amber-300 bg-amber-100 text-amber-800">
                                    提及
                                  </Badge>
                                ) : null}
                                <div className="text-sm font-medium">
                                  {event.content || ticketEventLabel(event.eventType)}
                                </div>
                              </div>
                              <div className="text-xs text-muted-foreground">
                                {event.createdAt ? formatDateTime(event.createdAt) : "—"}
                              </div>
                            </div>
                            <div className="mt-1 text-xs text-muted-foreground">
                              {event.operatorName || `用户#${event.operatorId}`}
                            </div>
                            {event.payload && event.eventType === "mentioned" && parseMentionUserIds(event.payload).length ? (
                              <div className="mt-2 flex flex-wrap gap-2">
                                {parseMentionUserIds(event.payload).map((userId) => {
                                  const user = agents.find((item) => item.userId === userId)
                                  return (
                                    <span
                                      key={`${event.id}-${userId}`}
                                      className="rounded-full border border-amber-200 px-2 py-1 text-xs text-amber-800"
                                    >
                                      @{user?.displayName || user?.nickname || user?.username || `用户#${userId}`}
                                    </span>
                                  )
                                })}
                              </div>
                            ) : null}
                          </div>
                        ))
                      ) : (
                        <div className="text-sm text-muted-foreground">暂无事件记录</div>
                      )}
                    </TabsContent>
                  </Tabs>
                </CardContent>
              </Card>
            </div>

            <div className="space-y-4">
              <Card>
                <CardHeader>
                  <CardTitle className="text-base">基础信息</CardTitle>
                  <CardAction>
                    <Button variant="ghost" size="sm" onClick={() => setEditDialogOpen(true)}>
                      编辑
                    </Button>
                  </CardAction>
                </CardHeader>
                <CardContent className="space-y-3 text-sm">
                  <InfoRow label="工单号" value={ticket.ticketNo} />
                  <InfoRow label="状态" value={ticketStatusLabel(ticket.status)} />
                  <InfoRow label="分类" value={ticket.categoryName || "未分类"} />
                  <InfoRow label="优先级" value={String(ticket.priority)} />
                  <InfoRow label="严重度" value={ticketSeverityLabel(ticket.severity)} />
                  <InfoRow label="处理人" value={ticket.currentAssigneeName || "未指派"} />
                  <InfoRow label="处理团队" value={ticket.currentTeamName || "未分组"} />
                  <InfoRow label="来源" value={formatTicketSource(ticket.source)} />
                  <InfoRow label="渠道" value={ticket.channel || "—"} />
                  <InfoRow label="重开次数" value={String(ticket.reopenedCount)} />
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle className="text-base">客户信息</CardTitle>
                </CardHeader>
                <CardContent className="space-y-3 text-sm">
                  <InfoRow label="客户" value={ticket.customer?.name || "未绑定客户"} />
                  <InfoRow label="手机号" value={ticket.customer?.primaryMobile || "—"} />
                  <InfoRow label="邮箱" value={ticket.customer?.primaryEmail || "—"} />
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle className="text-base">SLA 信息</CardTitle>
                </CardHeader>
                <CardContent className="space-y-3 text-sm">
                  {ticket.sla?.length ? (
                    ticket.sla.map((sla) => (
                      <div key={sla.slaType} className="rounded-lg border p-3">
                        <div className="font-medium">
                          {sla.slaType === "first_response" ? "首次响应" : "解决时效"}
                        </div>
                        <div className="mt-1 text-muted-foreground">目标：{sla.targetMinutes} 分钟</div>
                        <div className="mt-1 text-muted-foreground">状态：{formatSLAStatus(sla.status)}</div>
                        <div className="mt-1 text-muted-foreground">已耗时：{sla.elapsedMin} 分钟</div>
                        {sla.breachedAt ? (
                          <div className="mt-1 text-muted-foreground">
                            超时于：{formatDateTime(sla.breachedAt)}
                          </div>
                        ) : null}
                      </div>
                    ))
                  ) : (
                    <div className="text-muted-foreground">暂无 SLA 信息</div>
                  )}
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle className="text-base">解决信息</CardTitle>
                </CardHeader>
                <CardContent className="space-y-3 text-sm">
                  <InfoRow label="解决码" value={ticket.resolutionCodeName || ticket.resolutionCode || "—"} />
                  <InfoRow label="解决说明" value={ticket.resolutionSummary || "—"} />
                  <InfoRow label="关闭原因" value={ticket.closeReason || "—"} />
                  <InfoRow
                    label="解决时间"
                    value={ticket.resolvedAt ? formatDateTime(ticket.resolvedAt) : "—"}
                  />
                  <InfoRow
                    label="关闭时间"
                    value={ticket.closedAt ? formatDateTime(ticket.closedAt) : "—"}
                  />
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle className="text-base">时间信息</CardTitle>
                </CardHeader>
                <CardContent className="space-y-3 text-sm">
                  <InfoRow
                    label="创建时间"
                    value={ticket.createdAt ? formatDateTime(ticket.createdAt) : "—"}
                  />
                  <InfoRow
                    label="更新时间"
                    value={ticket.updatedAt ? formatDateTime(ticket.updatedAt) : "—"}
                  />
                  <InfoRow
                    label="截止时间"
                    value={ticket.dueAt ? formatDateTime(ticket.dueAt) : "—"}
                  />
                  <InfoRow
                    label="首次响应"
                    value={ticket.firstResponseAt ? formatDateTime(ticket.firstResponseAt) : "—"}
                  />
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle className="text-base">来源关联</CardTitle>
                </CardHeader>
                <CardContent className="space-y-3 text-sm">
                  <InfoRow
                    label="关联会话"
                    value={ticket.conversationId ? `#${ticket.conversationId}` : "未关联"}
                  />
                  {sourceConversation ? (
                    <div className="rounded-lg border bg-muted/20 p-3">
                      <div className="text-sm font-medium">{sourceConversation.subject || "未命名会话"}</div>
                      <div className="mt-2 flex flex-wrap gap-2 text-xs text-muted-foreground">
                        <span>
                          状态：
                          {getEnumLabel(
                            IMConversationStatusLabels,
                            sourceConversation.status as IMConversationStatus,
                          )}
                        </span>
                        <span>处理人：{sourceConversation.currentAssigneeName || "未指派"}</span>
                      </div>
                      <div className="mt-2 text-xs leading-6 text-muted-foreground">
                        {sourceConversation.lastMessageSummary || "暂无最近消息摘要"}
                      </div>
                      <div className="mt-2 text-xs text-muted-foreground">
                        最近活跃：{sourceConversation.lastActiveAt ? formatDateTime(sourceConversation.lastActiveAt) : "—"}
                      </div>
                    </div>
                  ) : null}
                  {ticket.conversationId ? (
                    <Button variant="outline" className="w-full" onClick={() => router.push("/conversations")}>
                      <RotateCcwIcon className="size-4" />
                      前往会话工作台
                    </Button>
                  ) : null}
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle className="text-base">关联工单</CardTitle>
                  <CardAction>
                    <Button variant="ghost" size="sm" onClick={() => setRelationDialogOpen(true)}>
                      <PlusIcon className="size-4" />
                      新增关联
                    </Button>
                  </CardAction>
                </CardHeader>
                <CardContent className="space-y-3 text-sm">
                  {childProgress.total > 0 ? (
                    <div className="rounded-lg border border-blue-200 bg-blue-50/60 p-3">
                      <div className="text-sm font-medium text-blue-900">子工单进展</div>
                      <div className="mt-2 flex flex-wrap gap-2 text-xs text-blue-900">
                        <span>总数：{childProgress.total}</span>
                        <span>已完成：{childProgress.completed}</span>
                        <span>进行中：{childProgress.active}</span>
                      </div>
                    </div>
                  ) : null}
                  {detail?.relatedTickets?.length ? (
                    detail.relatedTickets.map((relation) => (
                      <div key={relation.id} className="rounded-lg border bg-muted/20 p-3">
                        <div className="flex items-start justify-between gap-3">
                          <div className="min-w-0 space-y-1">
                            <div className="text-xs text-muted-foreground">
                              {ticketRelationLabel(relation.relationType)}
                            </div>
                            <div className="truncate font-medium">
                              {relation.relatedTicketNo || `#${relation.relatedTicketId}`}
                            </div>
                            <div className="text-sm text-muted-foreground">
                              {relation.relatedTicketTitle || "未命名工单"}
                            </div>
                          </div>
                          <div className="flex shrink-0 gap-1">
                            <Button
                              variant="ghost"
                              size="icon"
                              onClick={() => router.push(`/tickets/detail?id=${relation.relatedTicketId}`)}
                            >
                              <ExternalLinkIcon className="size-4" />
                            </Button>
                            <Button
                              variant="ghost"
                              size="icon"
                              disabled={saving}
                              onClick={() => void handleDeleteRelation(relation.id)}
                            >
                              <XIcon className="size-4" />
                            </Button>
                          </div>
                        </div>
                        <div className="mt-2 flex flex-wrap gap-2 text-xs text-muted-foreground">
                          {relation.relationType === "child" ? (
                            <Badge
                              variant="outline"
                              className={
                                isDoneTicketStatus(relation.relatedTicketStatus)
                                  ? "border-emerald-300 bg-emerald-50 text-emerald-800"
                                  : "border-orange-300 bg-orange-50 text-orange-800"
                              }
                            >
                              {isDoneTicketStatus(relation.relatedTicketStatus) ? "子单已完成" : "子单处理中"}
                            </Badge>
                          ) : null}
                          <span>状态：{ticketStatusLabel(relation.relatedTicketStatus || "")}</span>
                          <span>团队：{relation.currentTeamName || "未分组"}</span>
                          <span>处理人：{relation.currentAssigneeName || "未指派"}</span>
                        </div>
                        <div className="mt-2 text-xs text-muted-foreground">
                          最近更新：{relation.updatedAt ? formatDateTime(relation.updatedAt) : "—"}
                        </div>
                      </div>
                    ))
                  ) : (
                    <div className="text-muted-foreground">暂无关联工单</div>
                  )}
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle className="text-base">协作人</CardTitle>
                  <CardAction>
                    <Button variant="ghost" size="sm" onClick={() => setCollaboratorDialogOpen(true)}>
                      <PlusIcon className="size-4" />
                      新增协作人
                    </Button>
                  </CardAction>
                </CardHeader>
                <CardContent className="space-y-3 text-sm">
                  {detail?.collaborators?.length ? (
                    detail.collaborators.map((collaborator) => (
                      <div
                        key={collaborator.id}
                        className="flex items-center justify-between gap-3 rounded-lg border bg-muted/20 p-3"
                      >
                        <div className="min-w-0">
                          <div className="font-medium">{collaborator.userName || `用户#${collaborator.userId}`}</div>
                          <div className="text-xs text-muted-foreground">{collaborator.teamName || "未分组"}</div>
                        </div>
                        <Button
                          variant="ghost"
                          size="icon"
                          disabled={saving}
                          onClick={() => void handleDeleteCollaborator(collaborator.id)}
                        >
                          <XIcon className="size-4" />
                        </Button>
                      </div>
                    ))
                  ) : (
                    <div className="text-muted-foreground">暂无协作人</div>
                  )}
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle className="text-base">关注人</CardTitle>
                </CardHeader>
                <CardContent className="space-y-3 text-sm">
                  {detail?.watchers?.length ? (
                    detail.watchers.map((watcher) => (
                      <InfoRow key={watcher.id} label={`用户#${watcher.userId}`} value={watcher.userName || "未命名"} />
                    ))
                  ) : (
                    <div className="text-muted-foreground">暂无关注人</div>
                  )}
                </CardContent>
              </Card>
            </div>
          </div>
        ) : (
          <Card>
            <CardContent className="py-20 text-center text-muted-foreground">工单不存在</CardContent>
          </Card>
        )}
      </div>
      <TicketAssignDialog
        open={assignDialogOpen}
        ticketId={ticket?.id ?? null}
        currentTeamId={ticket?.currentTeamId}
        currentAssigneeId={ticket?.currentAssigneeId}
        onOpenChange={setAssignDialogOpen}
        onSuccess={loadDetail}
      />
      <TicketStatusDialog
        open={statusDialogOpen}
        ticketId={ticket?.id ?? null}
        currentStatus={ticket?.status}
        onOpenChange={setStatusDialogOpen}
        onSuccess={loadDetail}
      />
      <TicketReasonDialog
        open={closeDialogOpen}
        mode="close"
        ticketId={ticket?.id ?? null}
        defaultReason={ticket?.closeReason || "处理完成"}
        onOpenChange={setCloseDialogOpen}
        onSuccess={loadDetail}
      />
      <TicketReasonDialog
        open={reopenDialogOpen}
        mode="reopen"
        ticketId={ticket?.id ?? null}
        defaultReason="客户有新反馈"
        onOpenChange={setReopenDialogOpen}
        onSuccess={loadDetail}
      />
      <EditDialog
        open={editDialogOpen}
        saving={saving}
        itemId={ticket?.id ?? null}
        onOpenChange={setEditDialogOpen}
        onSubmit={handleEditSubmit}
      />
      <TicketRelationDialog
        open={relationDialogOpen}
        ticketId={ticket?.id ?? null}
        onOpenChange={setRelationDialogOpen}
        onSuccess={loadDetail}
      />
      <TicketCollaboratorDialog
        open={collaboratorDialogOpen}
        ticketId={ticket?.id ?? null}
        onOpenChange={setCollaboratorDialogOpen}
        onSuccess={loadDetail}
      />
    </div>
  )
}

function SummaryMetric({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg border bg-muted/20 p-3">
      <div className="text-xs text-muted-foreground">{label}</div>
      <div className="mt-1 text-sm font-medium">{value}</div>
    </div>
  )
}

function InfoRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-start justify-between gap-4">
      <span className="text-muted-foreground">{label}</span>
      <span className="max-w-[60%] text-right">{value}</span>
    </div>
  )
}
