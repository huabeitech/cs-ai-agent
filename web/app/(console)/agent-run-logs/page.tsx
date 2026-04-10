"use client"

import { useCallback, useEffect, useMemo, useState, type ReactNode } from "react"
import {
  BotMessageSquareIcon,
  RefreshCwIcon,
  SearchIcon,
  WorkflowIcon,
} from "lucide-react"
import { toast } from "sonner"

import { ListPagination } from "@/components/list-pagination"
import { OptionCombobox } from "@/components/option-combobox"
import { Badge } from "@/components/ui/badge"
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
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import {
  fetchAgentRunLog,
  fetchAgentRunLogs,
  fetchAIAgentsAll,
  type AgentRunLog,
  type AIAgent,
  type PageResult,
} from "@/lib/api/admin"
import { formatDateTime } from "@/lib/utils"

const actionOptions = [
  { value: "all", label: "全部动作" },
  { value: "rag", label: "RAG" },
  { value: "skill", label: "Skill" },
  { value: "tool", label: "Tool" },
  { value: "graph", label: "Graph" },
  { value: "handoff", label: "转人工" },
  { value: "reply", label: "回复" },
  { value: "fallback", label: "兜底" },
]

const finalStatusOptions = [
  { value: "all", label: "全部状态" },
  { value: "completed", label: "completed" },
  { value: "interrupted", label: "interrupted" },
  { value: "expired", label: "expired" },
  { value: "error", label: "error" },
  { value: "fallback", label: "fallback" },
]

function actionBadgeVariant(action: string) {
  switch (action) {
    case "handoff":
      return "destructive" as const
    case "skill":
      return "default" as const
    case "tool":
      return "default" as const
    case "graph":
      return "default" as const
    case "rag":
      return "secondary" as const
    case "fallback":
      return "outline" as const
    default:
      return "secondary" as const
  }
}

export default function DashboardAgentRunLogsPage() {
  const [keywordInput, setKeywordInput] = useState("")
  const [plannedActionInput, setPlannedActionInput] = useState("all")
  const [finalActionInput, setFinalActionInput] = useState("all")
  const [finalStatusInput, setFinalStatusInput] = useState("all")
  const [aiAgentIdInput, setAiAgentIdInput] = useState("all")
  const [keyword, setKeyword] = useState("")
  const [plannedAction, setPlannedAction] = useState("all")
  const [finalAction, setFinalAction] = useState("all")
  const [finalStatus, setFinalStatus] = useState("all")
  const [aiAgentId, setAiAgentId] = useState("all")
  const [page, setPage] = useState(1)
  const [limit, setLimit] = useState(20)
  const [loading, setLoading] = useState(true)
  const [detailLoading, setDetailLoading] = useState(false)
  const [detailOpen, setDetailOpen] = useState(false)
  const [activeLog, setActiveLog] = useState<AgentRunLog | null>(null)
  const [result, setResult] = useState<PageResult<AgentRunLog>>({
    results: [],
    page: { page: 1, limit: 20, total: 0 },
  })
  const [aiAgents, setAiAgents] = useState<AIAgent[]>([])
  const activeTraceData = useMemo(
    () => safeParseJSON(activeLog?.traceData ?? ""),
    [activeLog?.traceData]
  )
  const activeToolSearchTrace = useMemo(
    () => safeParseJSON(activeLog?.toolSearchTrace ?? ""),
    [activeLog?.toolSearchTrace]
  )
  const activeGraphToolTrace = useMemo(
    () => safeParseJSON(activeLog?.graphToolTrace ?? ""),
    [activeLog?.graphToolTrace]
  )

  const aiAgentOptions = useMemo(
    () => [
      { value: "all", label: "全部 Agent" },
      ...aiAgents.map((item) => ({
        value: String(item.id),
        label: item.name,
      })),
    ],
    [aiAgents]
  )

  const loadData = useCallback(async () => {
    setLoading(true)
    try {
      const data = await fetchAgentRunLogs({
        userMessage: keyword.trim() || undefined,
        plannedAction: plannedAction === "all" ? undefined : plannedAction,
        finalAction: finalAction === "all" ? undefined : finalAction,
        finalStatus: finalStatus === "all" ? undefined : finalStatus,
        aiAgentId: aiAgentId === "all" ? undefined : aiAgentId,
        page,
        limit,
      })
      setResult(data)
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "加载 Agent 运行日志失败")
    } finally {
      setLoading(false)
    }
  }, [aiAgentId, finalAction, finalStatus, keyword, limit, page, plannedAction])

  useEffect(() => {
    void loadData()
  }, [loadData])

  useEffect(() => {
    async function loadAIAgents() {
      try {
        const data = await fetchAIAgentsAll()
        setAiAgents(data)
      } catch (error) {
        toast.error(error instanceof Error ? error.message : "加载 AI Agent 列表失败")
      }
    }
    void loadAIAgents()
  }, [])

  function applyFilters() {
    setKeyword(keywordInput)
    setPlannedAction(plannedActionInput)
    setFinalAction(finalActionInput)
    setFinalStatus(finalStatusInput)
    setAiAgentId(aiAgentIdInput)
    setPage(1)
  }

  function handleFilterKeyDown(event: React.KeyboardEvent<HTMLInputElement>) {
    if (event.key !== "Enter") {
      return
    }
    event.preventDefault()
    applyFilters()
  }

  async function openDetail(id: number) {
    setDetailLoading(true)
    setDetailOpen(true)
    try {
      const data = await fetchAgentRunLog(id)
      setActiveLog(data)
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "加载日志详情失败")
      setDetailOpen(false)
    } finally {
      setDetailLoading(false)
    }
  }

  return (
    <>
      <div className="flex flex-1 flex-col gap-6 p-4 lg:p-6">
        <div className="flex flex-col gap-2 xl:flex-row xl:items-center xl:justify-end">
          <div className="relative min-w-72">
            <SearchIcon className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              value={keywordInput}
              onChange={(event) => setKeywordInput(event.target.value)}
              onKeyDown={handleFilterKeyDown}
              placeholder="按用户问题筛选"
              className="pl-9"
            />
          </div>
          <div className="w-full xl:w-40">
            <OptionCombobox
              value={plannedActionInput}
              options={actionOptions}
              placeholder="规划动作"
              searchPlaceholder="搜索动作"
              emptyText="未找到动作"
              onChange={(value) => setPlannedActionInput(value || "all")}
            />
          </div>
          <div className="w-full xl:w-40">
            <OptionCombobox
              value={finalActionInput}
              options={actionOptions}
              placeholder="最终动作"
              searchPlaceholder="搜索动作"
              emptyText="未找到动作"
              onChange={(value) => setFinalActionInput(value || "all")}
            />
          </div>
          <div className="w-full xl:w-40">
            <OptionCombobox
              value={finalStatusInput}
              options={finalStatusOptions}
              placeholder="最终状态"
              searchPlaceholder="搜索状态"
              emptyText="未找到状态"
              onChange={(value) => setFinalStatusInput(value || "all")}
            />
          </div>
          <div className="w-full xl:w-52">
            <OptionCombobox
              value={aiAgentIdInput}
              options={aiAgentOptions}
              placeholder="选择 Agent"
              searchPlaceholder="搜索 Agent"
              emptyText="未找到 Agent"
              onChange={(value) => setAiAgentIdInput(value || "all")}
            />
          </div>
          <Button variant="outline" onClick={applyFilters} disabled={loading}>
            <SearchIcon />
            查询
          </Button>
          <Button variant="outline" onClick={() => void loadData()} disabled={loading}>
            <RefreshCwIcon className={loading ? "animate-spin" : ""} />
            刷新列表
          </Button>
        </div>

        <div className="overflow-hidden rounded-lg border bg-background">
          <Table>
            <TableHeader className="bg-muted/40">
              <TableRow>
                <TableHead className="w-[180px]">时间</TableHead>
                <TableHead>用户问题</TableHead>
                <TableHead className="w-[120px]">规划动作</TableHead>
                <TableHead className="w-[220px]">Skill / Tool</TableHead>
                <TableHead className="w-[140px]">最终状态</TableHead>
                <TableHead className="w-[110px] text-right">耗时</TableHead>
                <TableHead className="w-[96px] text-right">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {!loading && result.results.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={7} className="py-14 text-center text-muted-foreground">
                    暂无 Agent 运行日志
                  </TableCell>
                </TableRow>
              ) : null}
              {result.results.map((item) => (
                <TableRow key={item.id}>
                  <TableCell className="text-sm text-muted-foreground">
                    {formatDateTime(item.createdAt)}
                  </TableCell>
                  <TableCell>
                    <div className="line-clamp-2 max-w-[620px] text-sm">
                      {item.userMessage || "-"}
                    </div>
                    {item.errorMessage ? (
                      <div className="mt-1 line-clamp-1 text-xs text-destructive">
                        {item.errorMessage}
                      </div>
                    ) : null}
                  </TableCell>
                  <TableCell>
                    <Badge variant={actionBadgeVariant(item.plannedAction)}>
                      {item.plannedAction || "-"}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-sm">
                    {item.plannedSkillCode || item.plannedToolCode ? (
                      <div className="space-y-1">
                        <Badge variant="outline">
                          {item.plannedSkillCode || item.graphToolCode || item.plannedToolCode}
                        </Badge>
                        {item.plannedSkillName ? (
                          <div className="line-clamp-1 text-xs text-muted-foreground">
                            {item.plannedSkillName}
                          </div>
                        ) : null}
                        {item.handoffReason ? (
                          <div className="line-clamp-1 text-xs text-muted-foreground">
                            转人工原因：{item.handoffReason}
                          </div>
                        ) : null}
                      </div>
                    ) : (
                      "-"
                    )}
                  </TableCell>
                  <TableCell>
                    <div className="space-y-1">
                      <Badge variant={actionBadgeVariant(item.finalAction)}>
                        {item.finalAction || "-"}
                      </Badge>
                      <div className="text-xs text-muted-foreground">
                        {item.finalStatus || "-"}
                      </div>
                    </div>
                  </TableCell>
                  <TableCell className="text-right text-sm text-muted-foreground">
                    {item.latencyMs} ms
                  </TableCell>
                  <TableCell className="text-right">
                    <Button variant="outline" size="sm" onClick={() => void openDetail(item.id)}>
                      详情
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
          <div className="border-t px-4 py-3">
            <ListPagination
              page={result.page.page}
              total={result.page.total}
              limit={limit}
              loading={loading}
              onPageChange={setPage}
              onLimitChange={(nextLimit) => {
                setLimit(nextLimit)
                setPage(1)
              }}
            />
          </div>
        </div>
      </div>

      <Drawer
        open={detailOpen}
        direction="right"
        onOpenChange={(open) => {
          setDetailOpen(open)
          if (!open) {
            setActiveLog(null)
          }
        }}
      >
        <DrawerContent className="min-w-180">
          <DrawerHeader>
            <DrawerTitle className="flex items-center gap-2">
              <WorkflowIcon className="size-4" />
              Agent 运行详情
            </DrawerTitle>
            <DrawerDescription>
              查看 planner 选择、最终动作、回复内容与错误信息。
            </DrawerDescription>
          </DrawerHeader>
          <div className="space-y-6 px-6 pb-6 overflow-auto">
            {detailLoading ? (
              <div className="py-10 text-sm text-muted-foreground">加载中...</div>
            ) : activeLog ? (
              <>
                <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
                  <MetricCard label="日志ID" value={String(activeLog.id)} />
                  <MetricCard label="会话ID" value={String(activeLog.conversationId || "-")} />
                  <MetricCard label="消息ID" value={String(activeLog.messageId || "-")} />
                  <MetricCard label="AI Agent" value={String(activeLog.aiAgentId || "-")} />
                </div>

                <InfoBlock
                  title="规划阶段"
                  lines={[
                    `plannedAction: ${activeLog.plannedAction || "-"}`,
                    `plannedSkillCode: ${activeLog.plannedSkillCode || "-"}`,
                    `plannedSkillName: ${activeLog.plannedSkillName || "-"}`,
                    `graphToolCode: ${activeLog.graphToolCode || "-"}`,
                    `plannedToolCode: ${activeLog.plannedToolCode || "-"}`,
                    `planReason: ${activeLog.planReason || "-"}`,
                    `handoffReason: ${activeLog.handoffReason || "-"}`,
                    `skillRouteTrace: ${activeLog.skillRouteTrace || "-"}`,
                  ]}
                />
                <InfoBlock
                  title="执行结果"
                  lines={[
                    `finalAction: ${activeLog.finalAction || "-"}`,
                    `finalStatus: ${activeLog.finalStatus || "-"}`,
                    `interruptType: ${activeLog.interruptType || "-"}`,
                    `resumeSource: ${activeLog.resumeSource || "-"}`,
                    `latencyMs: ${activeLog.latencyMs} ms`,
                    `createdAt: ${formatDateTime(activeLog.createdAt)}`,
                  ]}
                />

                <TextBlock
                  title="动态工具选择"
                  value={
                    activeToolSearchTrace
                      ? JSON.stringify(activeToolSearchTrace, null, 2)
                      : activeLog.toolSearchTrace
                  }
                />
                <TextBlock
                  title="Graph Tool 调用"
                  value={
                    activeGraphToolTrace
                      ? JSON.stringify(activeGraphToolTrace, null, 2)
                      : activeLog.graphToolTrace
                  }
                />
                <TextBlock
                  icon={<BotMessageSquareIcon className="size-4" />}
                  title="用户问题"
                  value={activeLog.userMessage}
                />
                <TextBlock
                  icon={<WorkflowIcon className="size-4" />}
                  title="机器人回复"
                  value={activeLog.replyText}
                />
                <TextBlock
                  title="错误信息"
                  value={activeLog.errorMessage}
                  tone="danger"
                />
                <TextBlock
                  title="链路 Trace"
                  value={
                    activeTraceData
                      ? JSON.stringify(activeTraceData, null, 2)
                      : activeLog.traceData
                  }
                />
              </>
            ) : (
              <div className="py-10 text-sm text-muted-foreground">未找到详情数据</div>
            )}
          </div>
          <DrawerFooter>
            <Button variant="outline" onClick={() => setDetailOpen(false)}>
              关闭
            </Button>
          </DrawerFooter>
        </DrawerContent>
      </Drawer>
    </>
  )
}

function safeParseJSON(value: string) {
  if (!value.trim()) {
    return null
  }
  try {
    return JSON.parse(value)
  } catch {
    return null
  }
}

function MetricCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg border bg-muted/20 p-4">
      <div className="text-xs text-muted-foreground">{label}</div>
      <div className="mt-2 text-lg font-semibold">{value}</div>
    </div>
  )
}

function InfoBlock({ title, lines }: { title: string; lines: string[] }) {
  return (
    <div className="rounded-lg border p-4">
      <div className="text-sm font-medium">{title}</div>
      <div className="mt-3 space-y-2 text-sm text-muted-foreground">
        {lines.map((line) => (
          <div key={line}>{line}</div>
        ))}
      </div>
    </div>
  )
}

function TextBlock({
  title,
  value,
  icon,
  tone = "default",
}: {
  title: string
  value?: string
  icon?: ReactNode
  tone?: "default" | "danger"
}) {
  return (
    <div className="rounded-lg border p-4">
      <div className="flex items-center gap-2 text-sm font-medium">
        {icon}
        {title}
      </div>
      <div
        className={
          tone === "danger"
            ? "mt-3 select-text whitespace-pre-wrap wrap-break-word text-sm text-destructive"
            : "mt-3 select-text whitespace-pre-wrap wrap-break-word text-sm text-muted-foreground"
        }
      >
        {value?.trim() || "-"}
      </div>
    </div>
  )
}
