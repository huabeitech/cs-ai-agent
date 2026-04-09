"use client"

import { useCallback, useEffect, useMemo, useState } from "react"
import { toast } from "sonner"

import {
  fetchKnowledgeRetrieveLog,
  type KnowledgeRetrieveHit,
  type KnowledgeRetrieveLogDetail,
} from "@/lib/api/admin"
import { formatDateTime } from "@/lib/utils"
import { Badge } from "@/components/ui/badge"
import {
  Drawer,
  DrawerContent,
  DrawerDescription,
  DrawerHeader,
  DrawerTitle,
} from "@/components/ui/drawer"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Separator } from "@/components/ui/separator"

type RetrieveLogDetailDrawerProps = {
  open: boolean
  retrieveLogId: number | null
  onOpenChange: (open: boolean) => void
}

function safeParseJSON(value: string) {
  if (!value) {
    return null
  }
  try {
    return JSON.parse(value) as Record<string, unknown>
  } catch {
    return null
  }
}

function CitationList({ hits }: { hits: KnowledgeRetrieveHit[] }) {
  const citations = hits.filter((item) => item.isCitation)
  if (citations.length === 0) {
    return <div className="text-sm text-muted-foreground">暂无引用来源</div>
  }
  return (
    <div className="space-y-3">
      {citations.map((item) => (
        <div key={item.id} className="rounded-lg border p-3">
          <div className="flex items-center gap-2">
            <span className="font-medium">{getHitSourceLabel(item)}</span>
            <Badge variant="outline">Chunk #{item.chunkNo}</Badge>
          </div>
          <div className="mt-1 text-xs text-muted-foreground">
            {item.sectionPath || item.title || "未记录章节"}
          </div>
          <div className="mt-2 text-sm leading-6 whitespace-pre-wrap text-foreground/90">
            {item.snippet || "-"}
          </div>
        </div>
      ))}
    </div>
  )
}

export function RetrieveLogDetailDrawer({
  open,
  retrieveLogId,
  onOpenChange,
}: RetrieveLogDetailDrawerProps) {
  const [loading, setLoading] = useState(false)
  const [detail, setDetail] = useState<KnowledgeRetrieveLogDetail | null>(null)

  const loadDetail = useCallback(async () => {
    if (!retrieveLogId) {
      setDetail(null)
      return
    }
    setLoading(true)
    try {
      const data = await fetchKnowledgeRetrieveLog(retrieveLogId)
      setDetail(data)
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "加载检索日志详情失败")
    } finally {
      setLoading(false)
    }
  }, [retrieveLogId])

  useEffect(() => {
    if (open && retrieveLogId) {
      void loadDetail()
    }
  }, [open, retrieveLogId, loadDetail])

  const traceData = useMemo(() => safeParseJSON(detail?.log.traceData ?? ""), [detail?.log.traceData])

  if (!open) {
    return null
  }

  return (
    <Drawer open={open} onOpenChange={onOpenChange} direction="right">
      <DrawerContent className="max-w-3xl">
        <DrawerHeader>
          <DrawerTitle>检索日志详情</DrawerTitle>
          <DrawerDescription>
            {detail?.log.question || (loading ? "加载中..." : "未找到检索日志")}
          </DrawerDescription>
        </DrawerHeader>
        <ScrollArea className="h-[calc(100vh-6rem)] px-4 pb-6">
          {!detail ? (
            <div className="py-6 text-sm text-muted-foreground">
              {loading ? "正在加载详情..." : "暂无详情数据"}
            </div>
          ) : (
            <div className="space-y-6 pb-6">
              <section className="space-y-3">
                <h3 className="text-sm font-semibold">请求信息</h3>
                <div className="grid gap-3 rounded-lg border p-4 md:grid-cols-2">
                  <div>
                    <div className="text-xs text-muted-foreground">知识库</div>
                    <div className="mt-1 text-sm">{detail.log.knowledgeBaseName || `#${detail.log.knowledgeBaseId}`}</div>
                  </div>
                  <div>
                    <div className="text-xs text-muted-foreground">创建时间</div>
                    <div className="mt-1 text-sm">{formatDateTime(detail.log.createdAt)}</div>
                  </div>
                  <div>
                    <div className="text-xs text-muted-foreground">渠道 / 场景</div>
                    <div className="mt-1 text-sm">{detail.log.channelName} / {detail.log.sceneName}</div>
                  </div>
                  <div>
                    <div className="text-xs text-muted-foreground">Request ID</div>
                    <div className="mt-1 break-all font-mono text-xs">{detail.log.requestId || "-"}</div>
                  </div>
                  <div className="md:col-span-2">
                    <div className="text-xs text-muted-foreground">原始问题</div>
                    <div className="mt-1 text-sm leading-6 whitespace-pre-wrap">{detail.log.question || "-"}</div>
                  </div>
                  <div className="md:col-span-2">
                    <div className="text-xs text-muted-foreground">改写问题</div>
                    <div className="mt-1 text-sm leading-6 whitespace-pre-wrap">{detail.log.rewriteQuestion || "-"}</div>
                  </div>
                  <div className="md:col-span-2">
                    <div className="text-xs text-muted-foreground">回答内容</div>
                    <div className="mt-1 text-sm leading-6 whitespace-pre-wrap">{detail.log.answer || "-"}</div>
                  </div>
                </div>
              </section>

              <section className="space-y-3">
                <h3 className="text-sm font-semibold">检索策略</h3>
                <div className="grid gap-3 rounded-lg border p-4 md:grid-cols-3">
                  <Metric label="Chunk Provider" value={detail.log.chunkProvider || "-"} mono />
                  <Metric label="Target Tokens" value={detail.log.chunkTargetTokens} />
                  <Metric label="Max Tokens" value={detail.log.chunkMaxTokens} />
                  <Metric label="Overlap Tokens" value={detail.log.chunkOverlapTokens} />
                  <Metric label="Rerank" value={detail.log.rerankEnabled ? "已启用" : "未启用"} />
                  <Metric label="Rerank Limit" value={detail.log.rerankLimit} />
                </div>
              </section>

              <section className="space-y-3">
                <h3 className="text-sm font-semibold">结果概览</h3>
                <div className="grid gap-3 rounded-lg border p-4 md:grid-cols-4">
                  <Metric label="回答状态" value={detail.log.answerStatusName} />
                  <Metric label="命中数" value={detail.log.hitCount} />
                  <Metric label="引用数" value={detail.log.citationCount} />
                  <Metric label="上下文 Chunk" value={detail.log.usedChunkCount} />
                  <Metric label="Top Score" value={detail.log.topScore.toFixed(4)} mono />
                  <Metric label="检索耗时" value={`${detail.log.retrieveMs} ms`} />
                  <Metric label="生成耗时" value={`${detail.log.generateMs} ms`} />
                  <Metric label="总耗时" value={`${detail.log.latencyMs} ms`} />
                  <Metric label="Prompt Tokens" value={detail.log.promptTokens} />
                  <Metric label="Completion Tokens" value={detail.log.completionTokens} />
                  <Metric label="模型" value={detail.log.modelName || "-"} mono />
                  <Metric label="会话 ID" value={detail.log.sessionId || "-"} mono />
                </div>
              </section>

              <section className="space-y-3">
                <h3 className="text-sm font-semibold">引用来源</h3>
                <CitationList hits={detail.hits} />
              </section>

              <section className="space-y-3">
                <div className="flex items-center justify-between">
                  <h3 className="text-sm font-semibold">命中详情</h3>
                  <div className="text-xs text-muted-foreground">{detail.hits.length} 条</div>
                </div>
                <div className="space-y-3">
                  {detail.hits.map((item) => (
                    <div key={item.id} className="rounded-lg border p-3">
                      <div className="flex flex-wrap items-center gap-2">
                        <Badge variant="outline">#{item.rankNo}</Badge>
                        <span className="font-medium">{getHitSourceLabel(item)}</span>
                        <Badge variant={item.usedInAnswer ? "default" : "secondary"}>
                          {item.usedInAnswer ? "已入上下文" : "未入上下文"}
                        </Badge>
                        {item.isCitation ? <Badge>引用</Badge> : null}
                      </div>
                      <div className="mt-2 flex flex-wrap gap-x-4 gap-y-1 text-xs text-muted-foreground">
                        <span>章节：{item.sectionPath || item.title || "-"}</span>
                        <span>Chunk #{item.chunkNo}</span>
                        <span>Provider：{item.provider || "-"}</span>
                        <span>Score：{item.score.toFixed(4)}</span>
                        <span>Rerank：{item.rerankScore ? item.rerankScore.toFixed(4) : "-"}</span>
                      </div>
                      <Separator className="my-3" />
                      <div className="text-sm leading-6 whitespace-pre-wrap text-foreground/90">
                        {item.snippet || "-"}
                      </div>
                    </div>
                  ))}
                </div>
              </section>

              <section className="space-y-3">
                <h3 className="text-sm font-semibold">TraceData</h3>
                <div className="rounded-lg border bg-muted/20 p-4">
                  <pre className="overflow-x-auto text-xs leading-6 text-muted-foreground">
                    {traceData ? JSON.stringify(traceData, null, 2) : detail.log.traceData || "-"}
                  </pre>
                </div>
              </section>
            </div>
          )}
        </ScrollArea>
      </DrawerContent>
    </Drawer>
  )
}

function getHitSourceLabel(item: KnowledgeRetrieveHit) {
  if (item.faqQuestion) {
    return item.faqQuestion
  }
  if (item.documentTitle) {
    return item.documentTitle
  }
  if (item.faqId > 0) {
    return `FAQ #${item.faqId}`
  }
  return `文档 #${item.documentId}`
}

function Metric({
  label,
  value,
  mono = false,
}: {
  label: string
  value: string | number
  mono?: boolean
}) {
  return (
    <div>
      <div className="text-xs text-muted-foreground">{label}</div>
      <div className={`mt-1 text-sm ${mono ? "font-mono" : ""}`}>{String(value || value === 0 ? value : "-")}</div>
    </div>
  )
}
