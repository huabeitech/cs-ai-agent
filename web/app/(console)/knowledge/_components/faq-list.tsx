"use client";

import { MoreHorizontalIcon, PencilIcon, SearchIcon, Trash2Icon } from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { toast } from "sonner";

import { ListPagination } from "@/components/list-pagination";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import {
  createKnowledgeFAQ,
  deleteKnowledgeFAQ,
  fetchKnowledgeFAQs,
  updateKnowledgeFAQ,
  type CreateKnowledgeFAQPayload,
  type KnowledgeFAQ,
  type PageResult,
} from "@/lib/api/admin";
import { formatDateTime } from "@/lib/utils";
import { FAQEditDialog } from "./faq-edit";

type FAQListProps = {
  knowledgeBaseId: number | null;
  onActionStateChange?: (state: FAQListActionState) => void;
};

export type FAQListActionState = {
  onRefresh: () => void;
  onCreate: () => void;
  loading: boolean;
};

export function FAQList({ knowledgeBaseId, onActionStateChange }: FAQListProps) {
  const [keywordInput, setKeywordInput] = useState("");
  const [keyword, setKeyword] = useState("");
  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState(20);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingItem, setEditingItem] = useState<KnowledgeFAQ | null>(null);
  const [result, setResult] = useState<PageResult<KnowledgeFAQ>>({
    results: [],
    page: { page: 1, limit: 20, total: 0 },
  });

  const loadData = useCallback(async () => {
    if (!knowledgeBaseId) {
      setResult({
        results: [],
        page: { page: 1, limit: 20, total: 0 },
      });
      return;
    }
    setLoading(true);
    try {
      const data = await fetchKnowledgeFAQs({
        knowledgeBaseId,
        question: keyword.trim() || undefined,
        page,
        limit,
      });
      setResult(data);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "加载FAQ失败");
    } finally {
      setLoading(false);
    }
  }, [keyword, knowledgeBaseId, limit, page]);

  useEffect(() => {
    void loadData();
  }, [loadData]);

  useEffect(() => {
    onActionStateChange?.({
      onRefresh: () => void loadData(),
      onCreate: () => {
        setEditingItem(null);
        setDialogOpen(true);
      },
      loading,
    });
  }, [loadData, loading, onActionStateChange]);

  async function handleSubmit(payload: CreateKnowledgeFAQPayload) {
    setSaving(true);
    try {
      if (editingItem) {
        await updateKnowledgeFAQ({ id: editingItem.id, ...payload });
      } else {
        await createKnowledgeFAQ(payload);
      }
      setDialogOpen(false);
      setEditingItem(null);
      await loadData();
      toast.success("FAQ已保存");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "保存FAQ失败");
    } finally {
      setSaving(false);
    }
  }

  async function handleDelete(item: KnowledgeFAQ) {
    try {
      await deleteKnowledgeFAQ(item.id);
      toast.success("FAQ已删除");
      await loadData();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "删除FAQ失败");
    }
  }

  if (!knowledgeBaseId) {
    return <div className="flex h-full items-center justify-center text-sm text-muted-foreground">请选择一个FAQ知识库查看FAQ</div>;
  }

  return (
    <>
      <div className="flex h-full flex-col gap-4 p-4">
        <div className="flex items-center gap-2">
          <div className="relative max-w-md flex-1">
            <SearchIcon className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              value={keywordInput}
              onChange={(event) => setKeywordInput(event.target.value)}
              onKeyDown={(event) => {
                if (event.key === "Enter") {
                  setKeyword(keywordInput);
                  setPage(1);
                }
              }}
              placeholder="按问题搜索FAQ"
              className="pl-9"
            />
          </div>
          <Button
            variant="outline"
            onClick={() => {
              setKeyword(keywordInput);
              setPage(1);
            }}
            disabled={loading}
          >
            查询
          </Button>
        </div>

        <div className="min-h-0 flex-1 overflow-auto rounded-md border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>问题</TableHead>
                <TableHead>索引状态</TableHead>
                <TableHead>相似问</TableHead>
                <TableHead>更新时间</TableHead>
                <TableHead className="w-[80px] text-right">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {result.results.map((item) => (
                <TableRow key={item.id}>
                  <TableCell>
                    <div className="font-medium">{item.question}</div>
                    <div className="mt-1 line-clamp-2 text-sm text-muted-foreground">{item.answer}</div>
                  </TableCell>
                  <TableCell>
                    <Badge variant={item.indexStatus === "failed" ? "destructive" : item.indexStatus === "indexed" ? "secondary" : "outline"}>
                      {item.indexStatusName}
                    </Badge>
                  </TableCell>
                  <TableCell>{item.similarQuestions.length}</TableCell>
                  <TableCell>{formatDateTime(item.updatedAt)}</TableCell>
                  <TableCell className="text-right">
                    <DropdownMenu>
                      <DropdownMenuTrigger
                        render={<Button variant="ghost" size="icon" />}
                        aria-label={`更多操作 ${item.question}`}
                      >
                        <MoreHorizontalIcon className="size-4" />
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuItem
                          onClick={() => {
                            setEditingItem(item);
                            setDialogOpen(true);
                          }}
                        >
                          <PencilIcon className="mr-2 size-4" />
                          编辑
                        </DropdownMenuItem>
                        <DropdownMenuItem
                          className="text-destructive focus:text-destructive"
                          onClick={() => void handleDelete(item)}
                        >
                          <Trash2Icon className="mr-2 size-4" />
                          删除
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </TableCell>
                </TableRow>
              ))}
              {!loading && result.results.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={5} className="py-12 text-center text-sm text-muted-foreground">
                    当前知识库还没有FAQ
                  </TableCell>
                </TableRow>
              ) : null}
            </TableBody>
          </Table>
        </div>

        <ListPagination
          page={result.page.page}
          pageSize={result.page.limit}
          total={result.page.total}
          onPageChange={setPage}
          onPageSizeChange={(next) => {
            setLimit(next);
            setPage(1);
          }}
        />
      </div>

      <FAQEditDialog
        open={dialogOpen}
        saving={saving}
        itemId={editingItem?.id ?? null}
        knowledgeBaseId={knowledgeBaseId}
        onOpenChange={(open) => {
          if (!open) {
            setEditingItem(null);
          }
          setDialogOpen(open);
        }}
        onSubmit={handleSubmit}
      />
    </>
  );
}
