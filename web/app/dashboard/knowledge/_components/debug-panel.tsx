"use client";

import { useState } from "react";
import { BotIcon, SearchIcon, SparklesIcon } from "lucide-react";
import { toast } from "sonner";

import {
  debugKnowledgeCompare,
  debugKnowledgeAnswer,
  debugKnowledgeSearch,
  type KnowledgeAnswerResponse,
  type KnowledgeCompareResponse,
  type KnowledgeSearchResponse,
} from "@/lib/api/admin";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Textarea } from "@/components/ui/textarea";

type DebugPanelProps = {
  knowledgeBaseId: number | null;
};

export function DebugPanel({ knowledgeBaseId }: DebugPanelProps) {
  const [question, setQuestion] = useState("");
  const [expectedDocIds, setExpectedDocIds] = useState("");
  const [topK, setTopK] = useState("5");
  const [scoreThreshold, setScoreThreshold] = useState("0.2");
  const [rerankLimit, setRerankLimit] = useState("5");
  const [searching, setSearching] = useState(false);
  const [comparing, setComparing] = useState(false);
  const [answering, setAnswering] = useState(false);
  const [searchResult, setSearchResult] = useState<KnowledgeSearchResponse | null>(null);
  const [compareResult, setCompareResult] = useState<KnowledgeCompareResponse | null>(null);
  const [answerResult, setAnswerResult] = useState<KnowledgeAnswerResponse | null>(null);

  async function handleSearch() {
    if (!knowledgeBaseId) {
      toast.error("请先选择知识库");
      return;
    }
    if (!question.trim()) {
      toast.error("请输入调试问题");
      return;
    }

    setSearching(true);
    try {
      const data = await debugKnowledgeSearch({
        knowledgeBaseId,
        question: question.trim(),
        topK: Number(topK) || undefined,
        scoreThreshold: Number(scoreThreshold) || undefined,
        rerankLimit: Number(rerankLimit) || undefined,
      });
      setSearchResult(data);
      toast.success(`检索完成，命中 ${data.hitCount} 条`);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "检索失败");
    } finally {
      setSearching(false);
    }
  }

  async function handleCompare() {
    if (!knowledgeBaseId) {
      toast.error("请先选择知识库");
      return;
    }
    if (!question.trim()) {
      toast.error("请输入调试问题");
      return;
    }

    setComparing(true);
    try {
      const data = await debugKnowledgeCompare({
        knowledgeBaseId,
        question: question.trim(),
        expectedDocIds: parseExpectedDocIds(expectedDocIds),
        topK: Number(topK) || undefined,
        scoreThreshold: Number(scoreThreshold) || undefined,
      });
      setCompareResult(data);
      toast.success(`对比完成，输出 ${data.providers.length} 组结果`);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "对比失败");
    } finally {
      setComparing(false);
    }
  }

  async function handleAnswer() {
    if (!knowledgeBaseId) {
      toast.error("请先选择知识库");
      return;
    }
    if (!question.trim()) {
      toast.error("请输入调试问题");
      return;
    }

    setAnswering(true);
    try {
      const data = await debugKnowledgeAnswer({
        knowledgeBaseId,
        question: question.trim(),
        topK: Number(topK) || undefined,
        scoreThreshold: Number(scoreThreshold) || undefined,
        rerankLimit: Number(rerankLimit) || undefined,
      });
      setAnswerResult(data);
      toast.success(`问答完成，状态：${data.answerStatusName}`);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "问答失败");
    } finally {
      setAnswering(false);
    }
  }

  return (
    <div className="flex h-full flex-col gap-3 p-3">
      <div className="space-y-3">
        <Textarea
          value={question}
          onChange={(event) => setQuestion(event.target.value)}
          placeholder="输入问题，测试知识库召回和回答效果"
          rows={5}
          className="text-sm"
        />
        <div className="grid grid-cols-3 gap-3">
          <div className="space-y-1.5">
            <Label htmlFor="topk" className="text-xs">TopK</Label>
            <Input id="topk" value={topK} onChange={(event) => setTopK(event.target.value)} placeholder="召回数量" className="h-8" />
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="threshold" className="text-xs">相似度阈值</Label>
            <Input id="threshold" value={scoreThreshold} onChange={(event) => setScoreThreshold(event.target.value)} placeholder="最低分数" className="h-8" />
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="rerank" className="text-xs">重排数量</Label>
            <Input id="rerank" value={rerankLimit} onChange={(event) => setRerankLimit(event.target.value)} placeholder="重排条数" className="h-8" />
          </div>
        </div>
        <div className="space-y-1.5">
          <Label htmlFor="expected-doc-ids" className="text-xs">期望文档ID</Label>
          <Input
            id="expected-doc-ids"
            value={expectedDocIds}
            onChange={(event) => setExpectedDocIds(event.target.value)}
            placeholder="例如：12,35,78"
            className="h-8"
          />
        </div>
        <div className="flex gap-2">
          <Button className="flex-1" variant="outline" onClick={() => void handleSearch()} disabled={searching}>
            <SearchIcon className="mr-2 size-4" />
            {searching ? "检索中..." : "调试检索"}
          </Button>
          <Button className="flex-1" variant="outline" onClick={() => void handleCompare()} disabled={comparing}>
            <SearchIcon className="mr-2 size-4" />
            {comparing ? "对比中..." : "对比 Provider"}
          </Button>
          <Button className="flex-1" onClick={() => void handleAnswer()} disabled={answering}>
            <SparklesIcon className="mr-2 size-4" />
            {answering ? "生成中..." : "调试问答"}
          </Button>
        </div>
      </div>

      <ScrollArea className="min-h-0 flex-1">
        <div className="space-y-3">
          {answerResult ? (
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="flex items-center gap-2 text-sm">
                  <BotIcon className="size-4" />
                  回答结果
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-3 text-sm">
                <div className="flex flex-wrap items-center gap-2">
                  <Badge variant="secondary">{answerResult.answerStatusName}</Badge>
                  <span className="text-xs text-muted-foreground">
                    {answerResult.latencyMs}ms · {answerResult.modelName || "fallback"}
                  </span>
                </div>
                <div className="rounded-md border bg-background p-3 whitespace-pre-wrap">
                  {answerResult.answer}
                </div>
                {answerResult.citations.length > 0 ? (
                  <div className="space-y-2">
                    <div className="text-xs font-medium text-muted-foreground">引用来源</div>
                    {answerResult.citations.map((citation) => (
                      <div
                        key={`${citation.documentId}-${citation.chunkNo}-${citation.sectionPath}`}
                        className="rounded-md border bg-muted/30 p-3"
                      >
                        <div className="flex items-center justify-between gap-2">
                          <div className="truncate text-xs font-medium">
                            {citation.documentTitle || `文档 ${citation.documentId}`}
                          </div>
                          <Badge variant="outline">{citation.score.toFixed(4)}</Badge>
                        </div>
                        <div className="mt-1 text-xs text-muted-foreground">
                          {citation.sectionPath || citation.title || `Chunk #${citation.chunkNo}`}
                        </div>
                        <div className="mt-2 text-xs leading-5 text-muted-foreground whitespace-pre-wrap">
                          {citation.snippet}
                        </div>
                      </div>
                    ))}
                  </div>
                ) : null}
              </CardContent>
            </Card>
          ) : null}

          {searchResult ? (
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm">检索命中</CardTitle>
              </CardHeader>
              <CardContent className="space-y-3">
                <div className="text-xs text-muted-foreground">
                  命中 {searchResult.hitCount} 条 · {searchResult.latencyMs}ms
                </div>
                {searchResult.results.map((item) => (
                  <div key={`${item.chunkId}-${item.documentId}`} className="rounded-md border bg-background p-3">
                    <div className="flex items-center justify-between gap-2">
                      <div className="truncate text-sm font-medium">
                        {item.documentTitle || `文档 ${item.documentId}`}
                      </div>
                      <Badge variant="outline">{item.score.toFixed(4)}</Badge>
                    </div>
                    <div className="mt-1 text-xs text-muted-foreground">
                      {item.sectionPath || item.title || `Chunk #${item.chunkNo}`}
                    </div>
                    <div className="mt-2 text-xs leading-5 text-muted-foreground whitespace-pre-wrap">
                      {item.content}
                    </div>
                  </div>
                ))}
              </CardContent>
            </Card>
          ) : null}

          {compareResult ? (
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm">Provider 对比</CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="text-xs text-muted-foreground">
                  总耗时 {compareResult.latencyMs}ms
                </div>
                {compareResult.providers.map((provider) => (
                  <div key={provider.provider} className="rounded-md border bg-background p-3 space-y-3">
                    <div className="flex items-center justify-between gap-2">
                      <div className="text-sm font-medium">{provider.provider}</div>
                      <div className="flex items-center gap-2 text-xs text-muted-foreground">
                        <span>构建 {provider.buildMs}ms · 检索 {provider.retrieveMs}ms · 命中 {provider.hitCount}</span>
                        {provider.top1Matched ? <Badge variant="secondary">Top1命中</Badge> : null}
                        {provider.top3Matched ? <Badge variant="secondary">Top3命中</Badge> : null}
                      </div>
                    </div>
                    {provider.matchedDocumentIds.length > 0 ? (
                      <div className="text-xs text-muted-foreground">
                        命中文档ID：{provider.matchedDocumentIds.join(", ")}
                      </div>
                    ) : null}
                    <div className="space-y-2">
                      {provider.results.map((item) => (
                        <div key={`${provider.provider}-${item.documentId}-${item.chunkNo}`} className="rounded border p-2">
                          <div className="flex items-center justify-between gap-2">
                            <div className="truncate text-xs font-medium">
                              {item.documentTitle || `文档 ${item.documentId}`}
                            </div>
                            <Badge variant="outline">{item.score.toFixed(4)}</Badge>
                          </div>
                          <div className="mt-1 text-xs text-muted-foreground">
                            {item.sectionPath || item.title || `Chunk #${item.chunkNo}`}
                          </div>
                          <div className="mt-2 text-xs leading-5 text-muted-foreground whitespace-pre-wrap">
                            {item.content}
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                ))}
              </CardContent>
            </Card>
          ) : null}
        </div>
      </ScrollArea>
    </div>
  );
}

function parseExpectedDocIds(value: string): number[] | undefined {
  const ids = value
    .split(",")
    .map((item) => Number(item.trim()))
    .filter((item) => Number.isFinite(item) && item > 0)
  return ids.length > 0 ? ids : undefined
}
