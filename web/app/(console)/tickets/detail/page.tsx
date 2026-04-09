"use client";

import {
  ChevronDownIcon,
  ChevronRightIcon,
  Cog,
  ExternalLinkIcon,
  Heart,
  MessageSquarePlusIcon,
  PanelRightCloseIcon,
  PanelRightOpenIcon,
  PencilIcon,
  PlusIcon,
  RefreshCcwIcon,
  RotateCcwIcon,
  UserRoundPlusIcon,
  XIcon,
} from "lucide-react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import {
  useCallback,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";
import { toast } from "sonner";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { ButtonGroup } from "@/components/ui/button-group";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  fetchAgentProfilesAll,
  fetchConversationDetail,
  type AdminAgentProfile,
  type AdminConversationDetail,
} from "@/lib/api/admin";
import {
  deleteTicketCollaborator,
  deleteTicketRelation,
  fetchTicketDetail,
  unwatchTicket,
  updateTicket,
  watchTicket,
  type CreateTicketPayload,
  type TicketDetail,
  type TicketItem,
  type UpdateTicketPayload,
} from "@/lib/api/ticket";
import { readSession } from "@/lib/auth";
import { getEnumLabel } from "@/lib/enums";
import {
  IMConversationStatus,
  IMConversationStatusLabels,
} from "@/lib/generated/enums";
import { formatDateTime } from "@/lib/utils";
import { EditDialog } from "../_components/edit";
import { TicketAssignDialog } from "../_components/ticket-assign-dialog";
import { TicketCollaboratorDialog } from "../_components/ticket-collaborator-dialog";
import { TicketCustomerPanel } from "../_components/ticket-customer-panel";
import { TicketPriorityBadge } from "../_components/ticket-priority-badge";
import { TicketReasonDialog } from "../_components/ticket-reason-dialog";
import { TicketRelationDialog } from "../_components/ticket-relation-dialog";
import { TicketReplyDialog } from "../_components/ticket-reply-dialog";
import { TicketSLABadge } from "../_components/ticket-sla-badge";
import {
  TicketStatusBadge,
  ticketStatusLabel,
} from "../_components/ticket-status-badge";
import { TicketStatusDialog } from "../_components/ticket-status-dialog";

function formatTicketSource(source?: string) {
  switch (source) {
    case "conversation":
      return "会话转工单";
    case "manual":
      return "手动创建";
    case "email":
      return "邮件";
    case "api":
      return "API";
    case "system":
      return "系统";
    default:
      return source || "—";
  }
}

function formatSLAStatus(status: string) {
  switch (status) {
    case "running":
      return "进行中";
    case "paused":
      return "已暂停";
    case "completed":
      return "已完成";
    case "breached":
      return "已超时";
    default:
      return status;
  }
}

function ticketSeverityLabel(severity: number) {
  switch (severity) {
    case 1:
      return "轻微";
    case 2:
      return "一般";
    case 3:
      return "严重";
    default:
      return String(severity);
  }
}

function ticketRelationLabel(relationType?: string) {
  switch (relationType) {
    case "duplicate":
      return "重复工单";
    case "related":
      return "相关工单";
    case "parent":
      return "父工单";
    case "child":
      return "子工单";
    default:
      return relationType || "关联工单";
  }
}

function ticketEventLabel(eventType?: string) {
  switch (eventType) {
    case "mentioned":
      return "提及协作人";
    case "internal_noted":
      return "内部备注";
    case "replied":
      return "回复客户";
    case "status_changed":
      return "状态变更";
    case "assigned":
      return "指派工单";
    case "transferred":
      return "转派工单";
    default:
      return eventType || "事件";
  }
}

function getMainSLA(ticket: TicketItem | null) {
  return (
    ticket?.sla?.find((item) => item.slaType === "resolution") ??
    ticket?.sla?.[0] ??
    null
  );
}

function parseMentionUserIds(payload?: string) {
  if (!payload) {
    return [];
  }
  try {
    const data = JSON.parse(payload) as { mentionUserIds?: number[] };
    return Array.isArray(data.mentionUserIds) ? data.mentionUserIds : [];
  } catch {
    return [];
  }
}

function isDoneTicketStatus(status?: string) {
  return status === "resolved" || status === "closed" || status === "cancelled";
}

export default function TicketDetailPage() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const ticketId = Number(searchParams.get("id") || 0);
  const [detail, setDetail] = useState<TicketDetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [replyDialogOpen, setReplyDialogOpen] = useState(false);
  const [assignDialogOpen, setAssignDialogOpen] = useState(false);
  const [statusDialogOpen, setStatusDialogOpen] = useState(false);
  const [closeDialogOpen, setCloseDialogOpen] = useState(false);
  const [reopenDialogOpen, setReopenDialogOpen] = useState(false);
  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const [relationDialogOpen, setRelationDialogOpen] = useState(false);
  const [collaboratorDialogOpen, setCollaboratorDialogOpen] = useState(false);
  const [collapsedSections, setCollapsedSections] = useState<
    Record<string, boolean>
  >({});
  const [rightPanelCollapsed, setRightPanelCollapsed] = useState(false);
  const [sourceConversation, setSourceConversation] =
    useState<AdminConversationDetail | null>(null);
  const [agents, setAgents] = useState<AdminAgentProfile[]>([]);

  const ticket = detail?.ticket ?? null;
  const currentUserId = readSession()?.user?.id ?? 0;
  const isWatching = Boolean(
    detail?.watchers?.some((item) => item.userId === currentUserId),
  );
  const resolutionSLA = useMemo(() => getMainSLA(ticket), [ticket]);
  const childRelations = useMemo(
    () =>
      (detail?.relatedTickets || []).filter(
        (item) => item.relationType === "child",
      ),
    [detail?.relatedTickets],
  );
  const childProgress = useMemo(() => {
    const total = childRelations.length;
    const completed = childRelations.filter((item) =>
      isDoneTicketStatus(item.relatedTicketStatus),
    ).length;
    return {
      total,
      completed,
      active: total - completed,
    };
  }, [childRelations]);

  const toggleSection = useCallback((sectionKey: string) => {
    setCollapsedSections((current) => ({
      ...current,
      [sectionKey]: !current[sectionKey],
    }));
  }, []);

  const loadDetail = useCallback(async () => {
    if (!ticketId) {
      setDetail(null);
      setLoading(false);
      return;
    }
    setLoading(true);
    try {
      const data = await fetchTicketDetail(ticketId);
      setDetail(data);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "加载工单详情失败");
    } finally {
      setLoading(false);
    }
  }, [ticketId]);

  useEffect(() => {
    void loadDetail();
  }, [loadDetail]);

  useEffect(() => {
    void (async () => {
      try {
        const data = await fetchAgentProfilesAll();
        setAgents(Array.isArray(data) ? data : []);
      } catch {
        setAgents([]);
      }
    })();
  }, []);

  useEffect(() => {
    if (!ticket?.conversationId) {
      setSourceConversation(null);
      return;
    }
    void (async () => {
      try {
        const data = await fetchConversationDetail(ticket.conversationId);
        setSourceConversation(data);
      } catch {
        setSourceConversation(null);
      }
    })();
  }, [ticket?.conversationId]);

  async function handleWatchToggle() {
    if (!ticket) {
      return;
    }
    setSaving(true);
    try {
      if (isWatching) {
        await unwatchTicket(ticket.id);
        toast.success("已取消关注");
      } else {
        await watchTicket(ticket.id);
        toast.success("已关注工单");
      }
      await loadDetail();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "更新关注状态失败");
    } finally {
      setSaving(false);
    }
  }

  async function handleEditSubmit(
    payload: CreateTicketPayload | UpdateTicketPayload,
  ) {
    if (!("ticketId" in payload)) {
      toast.error("详情页仅支持编辑现有工单");
      return;
    }
    setSaving(true);
    try {
      await updateTicket(payload);
      toast.success("工单基础信息已更新");
      setEditDialogOpen(false);
      await loadDetail();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "更新工单失败");
    } finally {
      setSaving(false);
    }
  }

  async function handleDeleteRelation(relationId: number) {
    if (!ticket) {
      return;
    }
    setSaving(true);
    try {
      await deleteTicketRelation({ ticketId: ticket.id, relationId });
      toast.success("关联工单已移除");
      await loadDetail();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "移除关联工单失败");
    } finally {
      setSaving(false);
    }
  }

  async function handleDeleteCollaborator(collaboratorId: number) {
    if (!ticket) {
      return;
    }
    setSaving(true);
    try {
      await deleteTicketCollaborator({ ticketId: ticket.id, collaboratorId });
      toast.success("协作人已移除");
      await loadDetail();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "移除协作人失败");
    } finally {
      setSaving(false);
    }
  }

  return (
    <div className="flex h-[calc(100vh-4rem)] min-w-0 overflow-hidden">
      {loading ? (
        <div className="min-w-0 flex-1 p-4 lg:p-6">
          <div className="py-20 text-center text-muted-foreground">
            加载中...
          </div>
        </div>
      ) : ticket ? (
        <>
          <div className="min-w-0 flex-1">
            <div className="flex h-full flex-col">
              {/* 工单header */}
              <div className="border-b border-border/70 bg-muted/10 flex flex-col gap-2.5 px-8 py-2">
                <div className="flex flex-col gap-2.5 xl:flex-row xl:items-start xl:justify-between">
                  <div className="min-w-0 space-y-2">
                    <div className="text-xs font-medium tracking-wide text-muted-foreground">
                      {ticket.ticketNo}
                    </div>
                    <div className="flex min-w-0 flex-wrap items-center gap-2">
                      <TicketStatusBadge status={ticket.status} />
                      <TicketPriorityBadge priority={ticket.priority} priorityName={ticket.priorityName} />
                      <TicketSLABadge ticket={ticket} />
                    </div>
                  </div>
                  <div className="flex flex-wrap justify-end gap-2">
                    <ButtonGroup>
                      <Button
                        size="sm"
                        variant={isWatching ? "default" : "outline"}
                        onClick={() => void handleWatchToggle()}
                        disabled={saving || !ticket}
                      >
                        <Heart className="size-4" />
                        {isWatching ? "已关注" : "关注"}
                      </Button>
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => setReplyDialogOpen(true)}
                      >
                        <MessageSquarePlusIcon className="size-4" />
                        回复/备注
                      </Button>
                      {ticket.status === "closed" ? (
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => setReopenDialogOpen(true)}
                          disabled={saving}
                        >
                          <RotateCcwIcon className="size-4" />
                          重开工单
                        </Button>
                      ) : (
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => setCloseDialogOpen(true)}
                          disabled={saving || !ticket}
                        >
                          <XIcon className="size-4" />
                          关闭工单
                        </Button>
                      )}
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => setAssignDialogOpen(true)}
                      >
                        <UserRoundPlusIcon className="size-4" />
                        分配处理人
                      </Button>
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => setStatusDialogOpen(true)}
                      >
                        <Cog className="size-4" />
                        变更状态
                      </Button>
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => setEditDialogOpen(true)}
                      >
                        <PencilIcon className="size-4" />
                        {/* 编辑 */}
                      </Button>
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => void loadDetail()}
                        disabled={loading || saving}
                      >
                        <RefreshCcwIcon className="size-4" />
                        {/* 刷新 */}
                      </Button>
                    </ButtonGroup>
                  </div>
                </div>
              </div>

              <div className="min-h-0 overflow-y-auto">
                <div className="space-y-0">
                  <DetailSection className="px-4 lg:px-6">
                    <div className="space-y-2.5">
                      <h1 className="text-xl font-semibold tracking-tight text-foreground">
                        {ticket.title}
                      </h1>
                      <div className="border-l-2 pl-3 border-border/70 ">
                        <div className="whitespace-pre-wrap break-words text-sm leading-6 text-foreground/80">
                          {ticket.description || "暂无工单描述"}
                        </div>
                      </div>
                    </div>
                  </DetailSection>

                  <DetailSection className="px-4 pt-4 lg:px-6">
                    <Tabs defaultValue="comments" className="gap-3">
                      <TabsList>
                        <TabsTrigger value="comments">回复&备注</TabsTrigger>
                        <TabsTrigger value="events">事件记录</TabsTrigger>
                      </TabsList>
                      <TabsContent value="comments" className="space-y-3">
                        {(detail?.comments?.length || 0) > 0 ? (
                          <div className="relative pl-6">
                            <div className="absolute top-0 bottom-0 left-[0.55rem] w-px bg-border/70" />
                            <div className="space-y-4">
                              {detail?.comments?.map((comment) => (
                                <div
                                  key={`comment-${comment.id}`}
                                  className="relative"
                                >
                                  <div
                                    className={`absolute top-2 -left-6 size-3 rounded-full ring-4 ring-background ${
                                      comment.commentType === "public_reply"
                                        ? "bg-emerald-500"
                                        : "bg-amber-500"
                                    }`}
                                  />
                                  <div className="rounded-lg border border-border/60 bg-muted/15 px-3.5 py-3">
                                    <div className="mb-2 flex items-start justify-between gap-3">
                                      <div className="min-w-0 space-y-1">
                                        <div className="flex flex-wrap items-center gap-2">
                                          <div className="text-sm font-medium">
                                            {comment.authorName ||
                                              `用户#${comment.authorId}`}
                                          </div>
                                          <Badge
                                            variant="outline"
                                            className={
                                              comment.commentType ===
                                              "public_reply"
                                                ? "border-emerald-200 bg-emerald-50 text-emerald-700"
                                                : "border-amber-200 bg-amber-50 text-amber-700"
                                            }
                                          >
                                            {comment.commentType ===
                                            "public_reply"
                                              ? "客户回复"
                                              : "内部备注"}
                                          </Badge>
                                        </div>
                                        <div className="text-xs text-muted-foreground">
                                          {comment.createdAt
                                            ? formatDateTime(comment.createdAt)
                                            : "—"}
                                        </div>
                                      </div>
                                    </div>
                                    {comment.commentType === "internal_note" &&
                                    parseMentionUserIds(comment.payload).length ? (
                                      <div className="mb-2 flex flex-wrap gap-2">
                                        {parseMentionUserIds(comment.payload).map(
                                          (userId) => {
                                            const user = agents.find(
                                              (item) => item.userId === userId,
                                            );
                                            return (
                                              <span
                                                key={`${comment.id}-${userId}`}
                                                className="rounded-full border px-2 py-1 text-xs text-muted-foreground"
                                              >
                                                @
                                                {user?.displayName ||
                                                  user?.nickname ||
                                                  user?.username ||
                                                  `用户#${userId}`}
                                              </span>
                                            );
                                          },
                                        )}
                                      </div>
                                    ) : null}
                                    <div className="whitespace-pre-wrap break-words text-sm leading-6 text-foreground/90">
                                      {comment.content}
                                    </div>
                                  </div>
                                </div>
                              ))}
                            </div>
                          </div>
                        ) : (
                          <div className="text-sm text-muted-foreground">
                            暂无评论记录
                          </div>
                        )}
                      </TabsContent>
                      <TabsContent value="events" className="space-y-2">
                        {(detail?.events?.length || 0) > 0 ? (
                          detail?.events?.map((event) => (
                            <div
                              key={`event-${event.id}`}
                              className={`rounded-lg border px-3 py-2.5 ${
                                event.eventType === "mentioned"
                                  ? "border-amber-200 bg-amber-50/60"
                                  : "border-border/60 bg-muted/20"
                              }`}
                            >
                              <div className="flex items-center justify-between gap-3">
                                <div className="flex items-center gap-2">
                                  {event.eventType === "mentioned" ? (
                                    <Badge
                                      variant="outline"
                                      className="border-amber-300 bg-amber-100 text-amber-800"
                                    >
                                      提及
                                    </Badge>
                                  ) : null}
                                  <div className="text-sm font-medium">
                                    {event.content ||
                                      ticketEventLabel(event.eventType)}
                                  </div>
                                </div>
                                <div className="text-xs text-muted-foreground">
                                  {event.createdAt
                                    ? formatDateTime(event.createdAt)
                                    : "—"}
                                </div>
                              </div>
                              <div className="mt-1 text-xs text-muted-foreground">
                                {event.operatorName ||
                                  `用户#${event.operatorId}`}
                              </div>
                              {event.payload &&
                              event.eventType === "mentioned" &&
                              parseMentionUserIds(event.payload).length ? (
                                <div className="mt-2 flex flex-wrap gap-2">
                                  {parseMentionUserIds(event.payload).map(
                                    (userId) => {
                                      const user = agents.find(
                                        (item) => item.userId === userId,
                                      );
                                      return (
                                        <span
                                          key={`${event.id}-${userId}`}
                                          className="rounded-full border border-amber-200 px-2 py-1 text-xs text-amber-800"
                                        >
                                          @
                                          {user?.displayName ||
                                            user?.nickname ||
                                            user?.username ||
                                            `用户#${userId}`}
                                        </span>
                                      );
                                    },
                                  )}
                                </div>
                              ) : null}
                            </div>
                          ))
                        ) : (
                          <div className="text-sm text-muted-foreground">
                            暂无事件记录
                          </div>
                        )}
                      </TabsContent>
                    </Tabs>
                  </DetailSection>
                </div>
              </div>
            </div>
          </div>

          <div className="relative hidden shrink-0 border-l border-border/70 bg-background lg:block">
            <Button
              variant="outline"
              size="icon"
              className="absolute top-2.5 left-1/2 z-10 size-6.5 -translate-x-1/2 rounded-full shadow-sm"
              onClick={() => setRightPanelCollapsed((value) => !value)}
              aria-label={
                rightPanelCollapsed ? "展开右侧信息面板" : "折叠右侧信息面板"
              }
            >
              {rightPanelCollapsed ? (
                <PanelRightOpenIcon className="size-3.5" />
              ) : (
                <PanelRightCloseIcon className="size-3.5" />
              )}
            </Button>
          </div>

          <div
            className={`min-w-0 shrink-0 overflow-hidden bg-background transition-[width] duration-200 ${
              rightPanelCollapsed ? "w-0" : "w-full lg:w-[360px]"
            }`}
          >
            {rightPanelCollapsed ? null : (
              <div className="h-full">
                <div className="flex h-full flex-col gap-4">
                  <div className="min-h-0 overflow-y-auto">
                    <div className="space-y-0">
                      <DetailSection
                        title="基础信息"
                        collapsible
                        collapsed={Boolean(collapsedSections.basics)}
                        onToggle={() => toggleSection("basics")}
                        className="px-4 pt-3 lg:px-6 lg:pt-4"
                        action={
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => setEditDialogOpen(true)}
                          >
                            <PencilIcon className="size-3.5" />
                            编辑
                          </Button>
                        }
                        contentClassName="space-y-0 text-sm"
                      >
                        <InfoRow label="工单号" value={ticket.ticketNo} />
                        <InfoRow
                          label="状态"
                          value={ticketStatusLabel(ticket.status)}
                        />
                        <InfoRow
                          label="标签"
                          value={
                            ticket.tags && ticket.tags.length > 0
                              ? ticket.tags.map((tag) => tag.name).join(" / ")
                              : "未打标签"
                          }
                        />
                        <InfoRow
                          label="优先级"
                          value={ticket.priorityName || String(ticket.priority)}
                        />
                        <InfoRow
                          label="严重度"
                          value={ticketSeverityLabel(ticket.severity)}
                        />
                        <InfoRow
                          label="处理人"
                          value={ticket.currentAssigneeName || "未指派"}
                        />
                        <InfoRow
                          label="处理团队"
                          value={ticket.currentTeamName || "未分组"}
                        />
                        <InfoRow
                          label="解决时限"
                          value={
                            resolutionSLA
                              ? `${resolutionSLA.targetMinutes} 分钟 / ${formatSLAStatus(resolutionSLA.status)}`
                              : "未设置"
                          }
                        />
                        <InfoRow
                          label="来源"
                          value={formatTicketSource(ticket.source)}
                        />
                        <InfoRow label="渠道" value={ticket.channel || "—"} />
                        <InfoRow
                          label="重开次数"
                          value={String(ticket.reopenedCount)}
                        />
                        <InfoRow
                          label="创建时间"
                          value={
                            ticket.createdAt
                              ? formatDateTime(ticket.createdAt)
                              : "—"
                          }
                        />
                        <InfoRow
                          label="更新时间"
                          value={
                            ticket.updatedAt
                              ? formatDateTime(ticket.updatedAt)
                              : "—"
                          }
                        />
                        <InfoRow
                          label="截止时间"
                          value={
                            ticket.dueAt ? formatDateTime(ticket.dueAt) : "—"
                          }
                        />
                        <InfoRow
                          label="首次响应"
                          value={
                            ticket.firstResponseAt
                              ? formatDateTime(ticket.firstResponseAt)
                              : "—"
                          }
                        />
                      </DetailSection>

                      <DetailSection
                        title="客户信息"
                        collapsible
                        collapsed={Boolean(collapsedSections.customer)}
                        onToggle={() => toggleSection("customer")}
                        className="px-4 pt-4 lg:px-6"
                        contentClassName="text-sm"
                      >
                        <TicketCustomerPanel
                          ticketId={ticket.id}
                          customerId={ticket.customerId}
                          onRefresh={loadDetail}
                        />
                      </DetailSection>

                      <DetailSection
                        title="SLA 信息"
                        collapsible
                        collapsed={Boolean(collapsedSections.sla)}
                        onToggle={() => toggleSection("sla")}
                        className="px-4 pt-4 lg:px-6"
                        contentClassName="space-y-3 text-sm"
                      >
                      {ticket.sla?.length ? (
                        ticket.sla.map((sla) => (
                          <SurfacePanel key={sla.slaType} className="p-2.5">
                            <div className="font-medium">
                              {sla.slaType === "first_response"
                                ? "首次响应"
                                : "解决时效"}
                            </div>
                            <div className="mt-1 text-muted-foreground">
                              目标：{sla.targetMinutes} 分钟
                            </div>
                            <div className="mt-1 text-muted-foreground">
                              状态：{formatSLAStatus(sla.status)}
                            </div>
                            <div className="mt-1 text-muted-foreground">
                              已耗时：{sla.elapsedMin} 分钟
                            </div>
                            {sla.breachedAt ? (
                              <div className="mt-1 text-muted-foreground">
                                超时于：{formatDateTime(sla.breachedAt)}
                              </div>
                            ) : null}
                          </SurfacePanel>
                        ))
                      ) : (
                        <div className="rounded-lg border border-amber-200 bg-amber-50/70 p-3 text-sm text-amber-900">
                          当前工单未生成 SLA 记录，请检查工单 SLA
                          配置；如果系统仍在使用默认策略，也需要确认配置是否已补齐。
                            <Link
                              href="/ticket-priorities"
                              className="ml-1 font-medium underline underline-offset-4"
                            >
                              前往配置优先级
                            </Link>
                        </div>
                      )}
                    </DetailSection>

                      <DetailSection
                        title="解决信息"
                        collapsible
                        collapsed={Boolean(collapsedSections.resolution)}
                        onToggle={() => toggleSection("resolution")}
                        className="px-4 pt-4 lg:px-6"
                        contentClassName="space-y-0 text-sm"
                      >
                      <InfoRow
                        label="解决码"
                        value={
                          ticket.resolutionCodeName ||
                          ticket.resolutionCode ||
                          "—"
                        }
                      />
                      <InfoRow
                        label="解决说明"
                        value={ticket.resolutionSummary || "—"}
                      />
                      <InfoRow
                        label="关闭原因"
                        value={ticket.closeReason || "—"}
                      />
                      <InfoRow
                        label="解决时间"
                        value={
                          ticket.resolvedAt
                            ? formatDateTime(ticket.resolvedAt)
                            : "—"
                        }
                      />
                      <InfoRow
                        label="关闭时间"
                        value={
                          ticket.closedAt
                            ? formatDateTime(ticket.closedAt)
                            : "—"
                        }
                      />
                    </DetailSection>

                      <DetailSection
                        title="来源关联"
                        collapsible
                        collapsed={Boolean(collapsedSections.source)}
                        onToggle={() => toggleSection("source")}
                        className="px-4 pt-4 lg:px-6"
                        contentClassName="space-y-3 text-sm"
                      >
                      <InfoRow
                        label="关联会话"
                        value={
                          ticket.conversationId
                            ? `#${ticket.conversationId}`
                            : "未关联"
                        }
                      />
                      {sourceConversation ? (
                        <SurfacePanel className="p-2.5">
                          <div className="text-sm font-medium">
                            {sourceConversation.subject || "未命名会话"}
                          </div>
                          <div className="mt-2 flex flex-wrap gap-2 text-xs text-muted-foreground">
                            <span>
                              状态：
                              {getEnumLabel(
                                IMConversationStatusLabels,
                                sourceConversation.status as IMConversationStatus,
                              )}
                            </span>
                            <span>
                              处理人：
                              {sourceConversation.currentAssigneeName ||
                                "未指派"}
                            </span>
                          </div>
                          <div className="mt-2 text-xs leading-6 text-muted-foreground">
                            {sourceConversation.lastMessageSummary ||
                              "暂无最近消息摘要"}
                          </div>
                          <div className="mt-2 text-xs text-muted-foreground">
                            最近活跃：
                            {sourceConversation.lastActiveAt
                              ? formatDateTime(sourceConversation.lastActiveAt)
                              : "—"}
                          </div>
                        </SurfacePanel>
                      ) : null}
                      {ticket.conversationId ? (
                        <Button
                          variant="outline"
                          className="w-full"
                          onClick={() => router.push("/conversations")}
                        >
                          <RotateCcwIcon className="size-4" />
                          前往会话工作台
                        </Button>
                      ) : null}
                    </DetailSection>

                      <DetailSection
                        title="关联工单"
                        collapsible
                        collapsed={Boolean(collapsedSections.related)}
                        onToggle={() => toggleSection("related")}
                        className="px-4 pt-4 lg:px-6"
                        action={
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => setRelationDialogOpen(true)}
                        >
                          <PlusIcon className="size-4" />
                          新增关联
                        </Button>
                        }
                        contentClassName="space-y-3 text-sm"
                      >
                      {childProgress.total > 0 ? (
                        <div className="rounded-lg border border-blue-200 bg-blue-50/70 p-3">
                          <div className="text-sm font-medium text-blue-900">
                            子工单进展
                          </div>
                          <div className="mt-2 flex flex-wrap gap-2 text-xs text-blue-900">
                            <span>总数：{childProgress.total}</span>
                            <span>已完成：{childProgress.completed}</span>
                            <span>进行中：{childProgress.active}</span>
                          </div>
                        </div>
                      ) : null}
                      {detail?.relatedTickets?.length ? (
                        detail.relatedTickets.map((relation) => (
                          <SurfacePanel key={relation.id} className="p-2.5">
                            <div className="flex items-start justify-between gap-3">
                              <div className="min-w-0 space-y-1">
                                <div className="text-xs text-muted-foreground">
                                  {ticketRelationLabel(relation.relationType)}
                                </div>
                                <div className="truncate font-medium">
                                  {relation.relatedTicketNo ||
                                    `#${relation.relatedTicketId}`}
                                </div>
                                <div className="text-sm text-muted-foreground">
                                  {relation.relatedTicketTitle || "未命名工单"}
                                </div>
                              </div>
                              <div className="flex shrink-0 gap-1">
                                <Button
                                  variant="ghost"
                                  size="icon"
                                  onClick={() =>
                                    router.push(
                                      `/tickets/detail?id=${relation.relatedTicketId}`,
                                    )
                                  }
                                >
                                  <ExternalLinkIcon className="size-4" />
                                </Button>
                                <Button
                                  variant="ghost"
                                  size="icon"
                                  disabled={saving}
                                  onClick={() =>
                                    void handleDeleteRelation(relation.id)
                                  }
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
                                    isDoneTicketStatus(
                                      relation.relatedTicketStatus,
                                    )
                                      ? "border-emerald-300 bg-emerald-50 text-emerald-800"
                                      : "border-orange-300 bg-orange-50 text-orange-800"
                                  }
                                >
                                  {isDoneTicketStatus(
                                    relation.relatedTicketStatus,
                                  )
                                    ? "子单已完成"
                                    : "子单处理中"}
                                </Badge>
                              ) : null}
                              <span>
                                状态：
                                {ticketStatusLabel(
                                  relation.relatedTicketStatus || "",
                                )}
                              </span>
                              <span>
                                团队：{relation.currentTeamName || "未分组"}
                              </span>
                              <span>
                                处理人：
                                {relation.currentAssigneeName || "未指派"}
                              </span>
                            </div>
                            <div className="mt-2 text-xs text-muted-foreground">
                              最近更新：
                              {relation.updatedAt
                                ? formatDateTime(relation.updatedAt)
                                : "—"}
                            </div>
                          </SurfacePanel>
                        ))
                      ) : (
                        <div className="text-muted-foreground">
                          暂无关联工单
                        </div>
                      )}
                    </DetailSection>

                      <DetailSection
                        title="协作人"
                        collapsible
                        collapsed={Boolean(collapsedSections.collaborators)}
                        onToggle={() => toggleSection("collaborators")}
                        className="px-4 pt-4 lg:px-6"
                        action={
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => setCollaboratorDialogOpen(true)}
                        >
                          <PlusIcon className="size-4" />
                          新增协作人
                        </Button>
                        }
                        contentClassName="space-y-3 text-sm"
                      >
                      {detail?.collaborators?.length ? (
                        detail.collaborators.map((collaborator) => (
                          <SurfacePanel
                            key={collaborator.id}
                            className="flex items-center justify-between gap-3 p-2.5"
                          >
                            <div className="min-w-0">
                              <div className="font-medium">
                                {collaborator.userName ||
                                  `用户#${collaborator.userId}`}
                              </div>
                              <div className="text-xs text-muted-foreground">
                                {collaborator.teamName || "未分组"}
                              </div>
                            </div>
                            <Button
                              variant="ghost"
                              size="icon"
                              disabled={saving}
                              onClick={() =>
                                void handleDeleteCollaborator(collaborator.id)
                              }
                            >
                              <XIcon className="size-4" />
                            </Button>
                          </SurfacePanel>
                        ))
                      ) : (
                        <div className="text-muted-foreground">暂无协作人</div>
                      )}
                    </DetailSection>

                      <DetailSection
                        title="关注人"
                        collapsible
                        collapsed={Boolean(collapsedSections.watchers)}
                        onToggle={() => toggleSection("watchers")}
                        className="px-4 pt-4 lg:px-6"
                        contentClassName="space-y-3 text-sm"
                      >
                      {detail?.watchers?.length ? (
                        detail.watchers.map((watcher) => (
                          <InfoRow
                            key={watcher.id}
                            label={`用户#${watcher.userId}`}
                            value={watcher.userName || "未命名"}
                          />
                        ))
                      ) : (
                        <div className="text-muted-foreground">暂无关注人</div>
                      )}
                      </DetailSection>
                    </div>
                  </div>
                </div>
              </div>
            )}
          </div>
        </>
      ) : (
        <div className="min-w-0 flex-1 p-4 lg:p-6">
          <div className="py-20 text-center text-muted-foreground">
            工单不存在
          </div>
        </div>
      )}
      <TicketReplyDialog
        open={replyDialogOpen}
        ticketId={ticket?.id ?? null}
        onOpenChange={setReplyDialogOpen}
        onSuccess={loadDetail}
      />
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
  );
}

function InfoRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="grid grid-cols-[78px_minmax(0,1fr)] items-start gap-3 py-1">
      <span className="text-sm text-muted-foreground">{label}</span>
      <span className="text-right text-sm text-foreground wrap-break-word">
        {value}
      </span>
    </div>
  );
}

function DetailSection({
  title,
  description,
  action,
  collapsible,
  collapsed,
  onToggle,
  className,
  contentClassName,
  children,
}: {
  title?: string;
  description?: string;
  action?: ReactNode;
  collapsible?: boolean;
  collapsed?: boolean;
  onToggle?: () => void;
  className?: string;
  contentClassName?: string;
  children: ReactNode;
}) {
  return (
    <section
      className={`border-b border-border/70 py-4 last:border-b-0 ${className ?? ""}`}
    >
      {title || description || action ? (
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div className="min-w-0 space-y-1">
            <div className="flex items-center gap-2">
              {collapsible ? (
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  className="size-6 shrink-0"
                  onClick={onToggle}
                  aria-label={collapsed ? "展开面板" : "折叠面板"}
                >
                  {collapsed ? (
                    <ChevronRightIcon className="size-4" />
                  ) : (
                    <ChevronDownIcon className="size-4" />
                  )}
                </Button>
              ) : null}
              {title ? <h2 className="text-base font-medium">{title}</h2> : null}
            </div>
            {description ? (
              <p className="text-sm leading-5 text-muted-foreground">
                {description}
              </p>
            ) : null}
          </div>
          {action ? <div className="shrink-0">{action}</div> : null}
        </div>
      ) : null}
      {collapsed ? null :<div className="pt-3"><div className={contentClassName}>{children}</div></div>}
    </section>
  );
}

function SurfacePanel({
  className,
  children,
}: {
  className?: string;
  children: ReactNode;
}) {
  return (
    <div
      className={`rounded-lg border border-border/60 bg-muted/20 ${className ?? ""}`}
    >
      {children}
    </div>
  );
}
