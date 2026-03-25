"use client"

import { useCallback, useEffect, useMemo, useState } from "react"
import { RefreshCwIcon, SearchIcon } from "lucide-react"
import { toast } from "sonner"

import {
  fetchKnowledgeRetrieveLogs,
  type KnowledgeRetrieveLog,
  type PageResult,
} from "@/lib/api/admin"
import {
  KnowledgeAnswerStatusLabels,
  KnowledgeChunkProviderLabels,
  KnowledgeRetrieveChannelLabels,
  KnowledgeRetrieveSceneLabels,
} from "@/lib/generated/enums"
import { getEnumOptions } from "@/lib/enums"
import { formatDateTime } from "@/lib/utils"
import { ListPagination } from "@/components/list-pagination"
import { OptionCombobox } from "@/components/option-combobox"
import { Badge } from "@/components/ui/badge"
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
import { RetrieveLogDetailDrawer } from "./retrieve-log-detail"

type RetrieveLogListProps = {
  knowledgeBaseId: number | null
}

const channelOptions = [
  { value: "all", label: "全部渠道" },
  ...getEnumOptions(KnowledgeRetrieveChannelLabels).map((item) => ({
    value: String(item.value),
    label: item.label,
  })),
]

const sceneOptions = [
  { value: "all", label: "全部场景" },
  ...getEnumOptions(KnowledgeRetrieveSceneLabels).map((item) => ({
    value: String(item.value),
    label: item.label,
  })),
]

const answerStatusOptions = [
  { value: "all", label: "全部回答状态" },
  ...getEnumOptions(KnowledgeAnswerStatusLabels).map((item) => ({
    value: String(item.value),
    label: item.label,
  })),
]

const providerOptions = [
  { value: "all", label: "全部切分策略" },
  ...getEnumOptions(KnowledgeChunkProviderLabels).map((item) => ({
    value: String(item.value),
    label: item.label,
  })),
]

const rerankOptions = [
  { value: "all", label: "全部 Rerank" },
  { value: "1", label: "已启用 Rerank" },
  { value: "0", label: "未启用 Rerank" },
]

function getAnswerStatusVariant(status: number): "default" | "secondary" | "outline" | "destructive" {
  switch (status) {
    case 1:
      return "default"
    case 2:
      return "secondary"
    case 3:
      return "outline"
    case 4:
      return "destructive"
    default:
      return "outline"
  }
}

export function RetrieveLogList({
  knowledgeBaseId,
}: RetrieveLogListProps) {
  const [questionInput, setQuestionInput] = useState("")
  const [question, setQuestion] = useState("")
  const [channel, setChannel] = useState("all")
  const [scene, setScene] = useState("all")
  const [answerStatus, setAnswerStatus] = useState("all")
  const [chunkProvider, setChunkProvider] = useState("all")
  const [rerankEnabled, setRerankEnabled] = useState("all")
  const [page, setPage] = useState(1)
  const [limit, setLimit] = useState(20)
  const [loading, setLoading] = useState(false)
  const [detailOpen, setDetailOpen] = useState(false)
  const [selectedLogId, setSelectedLogId] = useState<number | null>(null)
  const [result, setResult] = useState<PageResult<KnowledgeRetrieveLog>>({
    results: [],
    page: { page: 1, limit: 20, total: 0 },
  })

  const loadData = useCallback(async () => {
    if (!knowledgeBaseId) {
      setResult({ results: [], page: { page: 1, limit: 20, total: 0 } })
      setLoading(false)
      return
    }

    setLoading(true)
    try {
      const data = await fetchKnowledgeRetrieveLogs({
        knowledgeBaseId,
        question: question.trim() || undefined,
        channel: channel === "all" ? undefined : channel,
        scene: scene === "all" ? undefined : scene,
        answerStatus: answerStatus === "all" ? undefined : Number(answerStatus),
        chunkProvider: chunkProvider === "all" ? undefined : chunkProvider,
        rerankEnabled: rerankEnabled === "all" ? undefined : Number(rerankEnabled),
        page,
        limit,
      })
      setResult(data)
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "加载检索日志失败")
    } finally {
      setLoading(false)
    }
  }, [answerStatus, channel, chunkProvider, knowledgeBaseId, limit, page, question, rerankEnabled, scene])

  useEffect(() => {
    void loadData()
  }, [loadData])

  useEffect(() => {
    setPage(1)
    setSelectedLogId(null)
    setDetailOpen(false)
  }, [knowledgeBaseId])

  const emptyStateText = useMemo(() => {
    if (!knowledgeBaseId) {
      return "请选择一个知识库查看检索日志"
    }
    if (loading) {
      return "正在加载检索日志..."
    }
    return "当前知识库还没有检索日志"
  }, [knowledgeBaseId, loading])

  function applyFilters() {
    setQuestion(questionInput)
    setPage(1)
  }

  function handleQuestionKeyDown(event: React.KeyboardEvent<HTMLInputElement>) {
    if (event.key !== "Enter") {
      return
    }
    event.preventDefault()
    applyFilters()
  }

  function handleOpenDetail(logId: number) {
    setSelectedLogId(logId)
    setDetailOpen(true)
  }

  if (!knowledgeBaseId) {
    return (
      <div className="flex h-full items-center justify-center text-sm text-muted-foreground">
        {emptyStateText}
      </div>
    )
  }

  return (
    <>
      <div className="flex h-full flex-col">
        <div className="flex flex-col gap-3 border-b bg-background px-6 py-2">
          <div className="grid gap-3 xl:grid-cols-[minmax(0,1.8fr)_repeat(5,minmax(0,0.8fr))_auto]">
            <div className="relative">
              <SearchIcon className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                value={questionInput}
                onChange={(event) => setQuestionInput(event.target.value)}
                onKeyDown={handleQuestionKeyDown}
                placeholder="按问题关键字筛选"
                className="pl-9"
              />
            </div>
            <OptionCombobox value={channel} options={channelOptions} placeholder="选择渠道" onChange={setChannel} />
            <OptionCombobox value={scene} options={sceneOptions} placeholder="选择场景" onChange={setScene} />
            <OptionCombobox value={answerStatus} options={answerStatusOptions} placeholder="回答状态" onChange={setAnswerStatus} />
            <OptionCombobox value={chunkProvider} options={providerOptions} placeholder="切分策略" onChange={setChunkProvider} />
            <OptionCombobox value={rerankEnabled} options={rerankOptions} placeholder="Rerank" onChange={setRerankEnabled} />
            <Button onClick={applyFilters}>筛选</Button>
          </div>
        </div>

        <div className="min-h-0 flex-1 overflow-auto px-6 py-4">
          <div className="overflow-hidden rounded-xl border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-42">时间</TableHead>
                  <TableHead>问题</TableHead>
                  <TableHead className="w-28">回答状态</TableHead>
                  <TableHead className="w-24 text-right">命中数</TableHead>
                  <TableHead className="w-24 text-right">TopScore</TableHead>
                  <TableHead className="w-28">Provider</TableHead>
                  <TableHead className="w-24">Rerank</TableHead>
                  <TableHead className="w-24 text-right">引用</TableHead>
                  <TableHead className="w-28 text-right">耗时</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {result.results.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={9} className="h-32 text-center text-muted-foreground">
                      {emptyStateText}
                    </TableCell>
                  </TableRow>
                ) : (
                  result.results.map((item) => (
                    <TableRow
                      key={item.id}
                      className="cursor-pointer"
                      onClick={() => handleOpenDetail(item.id)}
                    >
                      <TableCell className="text-xs text-muted-foreground">{formatDateTime(item.createdAt)}</TableCell>
                      <TableCell>
                        <div className="space-y-1">
                          <div className="line-clamp-2 font-medium">{item.question || "-"}</div>
                          <div className="flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
                            <span>{item.channelName}</span>
                            <span>{item.sceneName}</span>
                            {item.knowledgeBaseName ? <span>{item.knowledgeBaseName}</span> : null}
                          </div>
                        </div>
                      </TableCell>
                      <TableCell>
                        <Badge variant={getAnswerStatusVariant(item.answerStatus)}>
                          {item.answerStatusName}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-right">{item.hitCount}</TableCell>
                      <TableCell className="text-right font-mono text-xs">{item.topScore.toFixed(4)}</TableCell>
                      <TableCell>{item.chunkProvider || "-"}</TableCell>
                      <TableCell>{item.rerankEnabled ? `是 (${item.rerankLimit})` : "否"}</TableCell>
                      <TableCell className="text-right">{item.citationCount}</TableCell>
                      <TableCell className="text-right">{item.latencyMs} ms</TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </div>
        </div>

        <div className="border-t px-6 py-4">
          <ListPagination
            page={result.page.page}
            total={result.page.total}
            limit={result.page.limit}
            loading={loading}
            onPageChange={setPage}
            onLimitChange={(nextLimit) => {
              setLimit(nextLimit)
              setPage(1)
            }}
          />
        </div>
      </div>

      <RetrieveLogDetailDrawer
        open={detailOpen}
        retrieveLogId={selectedLogId}
        onOpenChange={setDetailOpen}
      />
    </>
  )
}
