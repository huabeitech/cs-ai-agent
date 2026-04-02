"use client";

import {
  MoreHorizontalIcon,
  SearchIcon,
  Trash2Icon,
  WrenchIcon,
} from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { toast } from "sonner";

import { ListPagination } from "@/components/list-pagination";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { ButtonGroup } from "@/components/ui/button-group";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import {
  buildKnowledgeFAQIndex,
  createKnowledgeFAQ,
  deleteKnowledgeFAQ,
  fetchKnowledgeFAQs,
  updateKnowledgeFAQ,
  type CreateKnowledgeFAQPayload,
  type KnowledgeFAQ,
  type PageResult,
} from "@/lib/api/admin";
import { getEnumLabel, getEnumOptions } from "@/lib/enums";
import {
  KnowledgeDocumentIndexStatus,
  KnowledgeDocumentIndexStatusLabels,
} from "@/lib/generated/enums";
import { formatDateTime } from "@/lib/utils";
import { FAQEditDialog } from "./faq-edit";
import { FAQImportDialog } from "./faq-import-dialog";

type FAQListProps = {
  knowledgeBaseId: number | null;
  onActionStateChange?: (state: FAQListActionState) => void;
};

export type FAQListActionState = {
  onRefresh: () => void;
  onCreate: () => void;
  onImport: () => void;
  loading: boolean;
  importing: boolean;
};

const indexStatusOptions = [
  { value: "all", label: "全部索引状态" },
  ...getEnumOptions(KnowledgeDocumentIndexStatusLabels),
] as const;

function getIndexStatusBadgeVariant(status: string) {
  switch (status) {
    case KnowledgeDocumentIndexStatus.Indexed:
      return "secondary" as const;
    case KnowledgeDocumentIndexStatus.Failed:
      return "destructive" as const;
    default:
      return "outline" as const;
  }
}

function renderIndexStatusBadge(item: KnowledgeFAQ) {
  const badge = (
    <Badge variant={getIndexStatusBadgeVariant(item.indexStatus)}>
      {item.indexStatusName}
    </Badge>
  );

  if (
    item.indexStatus !== KnowledgeDocumentIndexStatus.Failed ||
    !item.indexError
  ) {
    return badge;
  }

  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger>
          <span className="inline-flex">{badge}</span>
        </TooltipTrigger>
        <TooltipContent align="start" className="max-w-sm whitespace-normal">
          {item.indexError}
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}

export function FAQList({
  knowledgeBaseId,
  onActionStateChange,
}: FAQListProps) {
  const [keywordInput, setKeywordInput] = useState("");
  const [indexStatusFilterInput, setIndexStatusFilterInput] = useState("all");
  const [keyword, setKeyword] = useState("");
  const [indexStatusFilter, setIndexStatusFilter] = useState("all");
  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState(20);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [importing, setImporting] = useState(false);
  const [actionLoadingMap, setActionLoadingMap] = useState<
    Record<number, { rebuildIndex: boolean; delete: boolean }>
  >({});
  const [dialogOpen, setDialogOpen] = useState(false);
  const [importDialogOpen, setImportDialogOpen] = useState(false);
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
        indexStatus:
          indexStatusFilter === "all" ? undefined : indexStatusFilter,
        page,
        limit,
      });
      setResult(data);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "加载FAQ失败");
    } finally {
      setLoading(false);
    }
  }, [indexStatusFilter, keyword, knowledgeBaseId, limit, page]);

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
      onImport: () => setImportDialogOpen(true),
      loading,
      importing,
    });
  }, [importing, loadData, loading, onActionStateChange]);

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
    setActionLoadingMap((prev) => ({
      ...prev,
      [item.id]: { ...prev[item.id], delete: true },
    }));
    try {
      await deleteKnowledgeFAQ(item.id);
      toast.success("FAQ已删除");
      await loadData();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "删除FAQ失败");
    } finally {
      setActionLoadingMap((prev) => ({
        ...prev,
        [item.id]: { ...prev[item.id], delete: false },
      }));
    }
  }

  async function handleBuildIndex(item: KnowledgeFAQ) {
    setActionLoadingMap((prev) => ({
      ...prev,
      [item.id]: { ...prev[item.id], rebuildIndex: true },
    }));
    try {
      await buildKnowledgeFAQIndex(item.id);
      toast.success("FAQ索引已重建");
      await loadData();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "重建FAQ索引失败");
    } finally {
      setActionLoadingMap((prev) => ({
        ...prev,
        [item.id]: { ...prev[item.id], rebuildIndex: false },
      }));
    }
  }

  if (!knowledgeBaseId) {
    return (
      <div className="flex h-full items-center justify-center text-sm text-muted-foreground">
        请选择一个FAQ知识库查看FAQ
      </div>
    );
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
          <Select
            value={indexStatusFilterInput}
            onValueChange={(value) => setIndexStatusFilterInput(value ?? "all")}
          >
            <SelectTrigger className="w-40">
              <SelectValue>
                {indexStatusFilterInput === "all"
                  ? "全部索引状态"
                  : getEnumLabel(
                      KnowledgeDocumentIndexStatusLabels,
                      indexStatusFilterInput as KnowledgeDocumentIndexStatus,
                    )}
              </SelectValue>
            </SelectTrigger>
            <SelectContent>
              {indexStatusOptions.map((item) => (
                <SelectItem key={item.value} value={item.value}>
                  {item.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <Button
            variant="outline"
            onClick={() => {
              setKeyword(keywordInput);
              setIndexStatusFilter(indexStatusFilterInput);
              setPage(1);
            }}
            disabled={loading}
          >
            查询
          </Button>
        </div>

        <div className="min-h-0 flex-1 overflow-hidden rounded-md border">
          <div className="h-full overflow-auto">
            <table className="w-full min-w-max caption-bottom text-sm">
              <TableHeader>
                <TableRow>
                  <TableHead>问题</TableHead>
                  <TableHead>索引状态</TableHead>
                  <TableHead>相似问题</TableHead>
                  <TableHead>更新时间</TableHead>
                  <TableHead className="w-20 text-right">操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {result.results.map((item) => (
                  <TableRow key={item.id}>
                    <TableCell className="max-w-sm">
                      <div className="font-medium">{item.question}</div>
                      <div className="mt-1 line-clamp-2 text-sm text-muted-foreground">
                        {item.answer}
                      </div>
                    </TableCell>
                    <TableCell>{renderIndexStatusBadge(item)}</TableCell>
                    <TableCell>
                      {Array.isArray(item.similarQuestions)
                        ? item.similarQuestions.length
                        : 0}
                    </TableCell>
                    <TableCell>{formatDateTime(item.updatedAt)}</TableCell>
                    <TableCell className="w-20 text-right">
                      <ButtonGroup className="ml-auto">
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => {
                            setEditingItem(item);
                            setDialogOpen(true);
                          }}
                        >
                          编辑
                        </Button>
                        <DropdownMenu>
                          <DropdownMenuTrigger
                            render={<Button variant="outline" size="icon-sm" />}
                            aria-label={`更多操作 ${item.question}`}
                          >
                            <MoreHorizontalIcon className="size-4" />
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end">
                            <DropdownMenuItem
                              onClick={() => void handleBuildIndex(item)}
                            >
                              <WrenchIcon className="mr-2 size-4" />
                              {actionLoadingMap[item.id]?.rebuildIndex
                                ? "重建中..."
                                : "重建索引"}
                            </DropdownMenuItem>
                            <DropdownMenuItem
                              className="text-destructive focus:text-destructive"
                              onClick={() => void handleDelete(item)}
                            >
                              <Trash2Icon className="mr-2 size-4" />
                              {actionLoadingMap[item.id]?.delete
                                ? "删除中..."
                                : "删除"}
                            </DropdownMenuItem>
                          </DropdownMenuContent>
                        </DropdownMenu>
                      </ButtonGroup>
                    </TableCell>
                  </TableRow>
                ))}
                {!loading && result.results.length === 0 ? (
                  <TableRow>
                    <TableCell
                      colSpan={5}
                      className="py-12 text-center text-sm text-muted-foreground"
                    >
                      当前知识库还没有FAQ
                    </TableCell>
                  </TableRow>
                ) : null}
              </TableBody>
            </table>
          </div>
        </div>

        <ListPagination
          page={result.page.page}
          limit={result.page.limit}
          total={result.page.total}
          onPageChange={setPage}
          onLimitChange={(next: number) => {
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

      <FAQImportDialog
        open={importDialogOpen}
        knowledgeBaseId={knowledgeBaseId}
        importing={importing}
        onOpenChange={setImportDialogOpen}
        onImportingChange={setImporting}
        onImported={loadData}
      />
    </>
  );
}
