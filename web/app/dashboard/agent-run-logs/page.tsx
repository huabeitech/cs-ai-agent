"use client"

import { useCallback, useEffect, useMemo, useState } from "react"
import { RefreshCwIcon, SearchIcon } from "lucide-react"
import { toast } from "sonner"

import { ListPagination } from "@/components/list-pagination"
import { OptionCombobox } from "@/components/option-combobox"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { AgentRunLogDetailDialog } from "./_components/detail"
import {
  fetchAgentRunLogs,
  fetchAIAgentsAll,
  type AIAgent,
  type AgentRunLog,
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

const hitlStatusOptions = [
  { value: "all", label: "全部 HITL" },
  { value: "pending", label: "等待确认" },
  { value: "confirmed", label: "已确认" },
  { value: "cancelled", label: "已取消" },
  { value: "expired", label: "已过期" },
  { value: "triggered", label: "已触发" },
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

function hitlBadgeVariant(status: string) {
  switch (status) {
    case "pending":
      return "secondary" as const
    case "confirmed":
      return "default" as const
    case "cancelled":
      return "outline" as const
    case "expired":
      return "destructive" as const
    default:
      return "secondary" as const
  }
}

export default function DashboardAgentRunLogsPage() {
  const [keywordInput, setKeywordInput] = useState("")
  const [plannedActionInput, setPlannedActionInput] = useState("all")
  const [finalActionInput, setFinalActionInput] = useState("all")
  const [finalStatusInput, setFinalStatusInput] = useState("all")
  const [hitlStatusInput, setHitlStatusInput] = useState("all")
  const [aiAgentIdInput, setAiAgentIdInput] = useState("all")
  const [keyword, setKeyword] = useState("")
  const [plannedAction, setPlannedAction] = useState("all")
  const [finalAction, setFinalAction] = useState("all")
  const [finalStatus, setFinalStatus] = useState("all")
  const [hitlStatus, setHitlStatus] = useState("all")
  const [aiAgentId, setAiAgentId] = useState("all")
  const [page, setPage] = useState(1)
  const [limit, setLimit] = useState(20)
  const [loading, setLoading] = useState(true)
  const [detailOpen, setDetailOpen] = useState(false)
  const [activeLogId, setActiveLogId] = useState<number | null>(null)
  const [result, setResult] = useState<PageResult<AgentRunLog>>({
    results: [],
    page: { page: 1, limit: 20, total: 0 },
  })
  const [aiAgents, setAiAgents] = useState<AIAgent[]>([])

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
        hitlStatus: hitlStatus === "all" ? undefined : hitlStatus,
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
  }, [aiAgentId, finalAction, finalStatus, hitlStatus, keyword, limit, page, plannedAction])

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
    setHitlStatus(hitlStatusInput)
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

  return (
    <>
      <div className="flex flex-1 flex-col gap-6 p-4 lg:p-6">
        <div className="grid grid-cols-1 gap-2 md:grid-cols-2 xl:grid-cols-[minmax(0,1.6fr)_repeat(5,minmax(0,0.7fr))_auto_auto] xl:items-center">
          <div className="relative min-w-0">
            <SearchIcon className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              value={keywordInput}
              onChange={(event) => setKeywordInput(event.target.value)}
              onKeyDown={handleFilterKeyDown}
              placeholder="按用户问题筛选"
              className="pl-9"
            />
          </div>
          <div className="min-w-0">
            <OptionCombobox
              value={plannedActionInput}
              options={actionOptions}
              placeholder="规划动作"
              searchPlaceholder="搜索动作"
              emptyText="未找到动作"
              onChange={(value) => setPlannedActionInput(value || "all")}
            />
          </div>
          <div className="min-w-0">
            <OptionCombobox
              value={finalActionInput}
              options={actionOptions}
              placeholder="最终动作"
              searchPlaceholder="搜索动作"
              emptyText="未找到动作"
              onChange={(value) => setFinalActionInput(value || "all")}
            />
          </div>
          <div className="min-w-0">
            <OptionCombobox
              value={finalStatusInput}
              options={finalStatusOptions}
              placeholder="最终状态"
              searchPlaceholder="搜索状态"
              emptyText="未找到状态"
              onChange={(value) => setFinalStatusInput(value || "all")}
            />
          </div>
          <div className="min-w-0">
            <OptionCombobox
              value={hitlStatusInput}
              options={hitlStatusOptions}
              placeholder="HITL 状态"
              searchPlaceholder="搜索 HITL 状态"
              emptyText="未找到状态"
              onChange={(value) => setHitlStatusInput(value || "all")}
            />
          </div>
          <div className="min-w-0">
            <OptionCombobox
              value={aiAgentIdInput}
              options={aiAgentOptions}
              placeholder="选择 Agent"
              searchPlaceholder="搜索 Agent"
              emptyText="未找到 Agent"
              onChange={(value) => setAiAgentIdInput(value || "all")}
            />
          </div>
          <Button variant="outline" onClick={applyFilters} disabled={loading} className="w-full xl:w-auto">
            <SearchIcon />
            查询
          </Button>
          <Button
            variant="outline"
            onClick={() => void loadData()}
            disabled={loading}
            className="w-full xl:w-auto"
          >
            <RefreshCwIcon className={loading ? "animate-spin" : ""} />
            刷新列表
          </Button>
        </div>

        <div className="overflow-hidden rounded-lg border bg-background">
          {!loading && result.results.length === 0 ? (
            <div className="py-14 text-center text-sm text-muted-foreground">
              暂无 Agent 运行日志
            </div>
          ) : (
            <div>
              <div className="hidden grid-cols-[160px_minmax(0,1.8fr)_110px_minmax(0,1.2fr)_130px_90px_76px] gap-3 border-b bg-muted/40 px-4 py-3 text-sm text-muted-foreground lg:grid">
                <div>时间</div>
                <div>用户问题</div>
                <div>规划动作</div>
                <div>Skill / Tool</div>
                <div>最终状态</div>
                <div className="text-right">耗时</div>
                <div className="text-right">操作</div>
              </div>

              <div className="divide-y">
                {result.results.map((item) => (
                  <article
                    key={item.id}
                    className="grid grid-cols-1 gap-2 px-4 py-3 lg:grid-cols-[160px_minmax(0,1.8fr)_110px_minmax(0,1.2fr)_130px_90px_76px] lg:items-center lg:gap-3"
                  >
                    <div className="min-w-0 text-sm text-muted-foreground">
                      {formatDateTime(item.createdAt)}
                    </div>

                    <div className="min-w-0">
                      <UserMessagePreview value={item.userMessage} />
                      {item.errorMessage ? (
                        <div className="truncate text-xs text-destructive">{item.errorMessage}</div>
                      ) : null}
                    </div>

                    <div className="min-w-0">
                      <Badge variant={actionBadgeVariant(item.plannedAction)}>
                        {item.plannedAction || "-"}
                      </Badge>
                    </div>

                    <div className="min-w-0 text-sm">
                      {item.plannedSkillCode || item.graphToolCode || item.plannedToolCode ? (
                        <div className="min-w-0 space-y-1">
                          <div className="truncate font-medium">
                            {item.plannedSkillCode || item.graphToolCode || item.plannedToolCode}
                          </div>
                          {item.plannedSkillName ? (
                            <div className="truncate text-xs text-muted-foreground">
                              {item.plannedSkillName}
                            </div>
                          ) : item.handoffReason ? (
                            <div className="truncate text-xs text-muted-foreground">
                              转人工原因：{item.handoffReason}
                            </div>
                          ) : item.recommendedAction ? (
                            <div className="truncate text-xs text-muted-foreground">
                              分流建议：{item.recommendedAction}
                              {item.riskLevel ? ` / ${item.riskLevel} risk` : ""}
                              {item.ticketDraftReady ? " / 草稿已就绪" : ""}
                            </div>
                          ) : null}
                        </div>
                      ) : (
                        <span className="text-muted-foreground">-</span>
                      )}
                    </div>

                    <div className="min-w-0">
                      <div className="space-y-1">
                        <Badge variant={actionBadgeVariant(item.finalAction)}>
                          {item.finalAction || "-"}
                        </Badge>
                        <div className="truncate text-xs text-muted-foreground">
                          {item.hitlStatusName
                            ? `${item.hitlStatusName} / ${item.finalStatus || "-"}`
                            : item.finalStatus || "-"}
                        </div>
                      </div>
                    </div>

                    <div className="text-sm text-muted-foreground lg:text-right">
                      {item.latencyMs} ms
                    </div>

                    <div className="lg:text-right">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => {
                          setActiveLogId(item.id)
                          setDetailOpen(true)
                        }}
                      >
                        详情
                      </Button>
                    </div>
                  </article>
                ))}
              </div>
            </div>
          )}
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
      <AgentRunLogDetailDialog
        open={detailOpen}
        logId={activeLogId}
        onOpenChange={(open) => {
          setDetailOpen(open)
          if (!open) {
            setActiveLogId(null)
          }
        }}
      />
    </>
  )
}

function UserMessagePreview({ value }: { value?: string }) {
  const preview = useMemo(() => summarizeUserMessage(value), [value])

  return (
    <div className="truncate text-sm text-foreground">
      {preview}
    </div>
  )
}

function summarizeUserMessage(value?: string) {
  const normalized = value?.trim()
  if (!normalized) {
    return "-"
  }
  const text = extractTextFromHTML(normalized).replace(/\s+/g, " ").trim()
  if (text) {
    return text
  }
  if (containsHTML(normalized)) {
    if (/<img[\s>]/i.test(normalized)) {
      return "[图片]"
    }
    return "[富文本消息]"
  }
  return normalized
}

function containsHTML(value: string) {
  return /<[^>]+>/.test(value)
}

function extractTextFromHTML(value: string) {
  if (typeof window === "undefined") {
    return value
  }
  const doc = new DOMParser().parseFromString(value, "text/html")
  return doc.body.textContent || ""
}
