"use client";

import {
  AlertCircleIcon,
  FileTextIcon,
  MoreHorizontalIcon,
  PencilIcon,
  SearchIcon,
  Trash2Icon,
  WrenchIcon,
} from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  ContextMenu,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuTrigger,
} from "@/components/ui/context-menu";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import {
  buildKnowledgeDocumentIndex,
  createKnowledgeDocument,
  deleteKnowledgeDocument,
  fetchKnowledgeDocuments,
  updateKnowledgeDocument,
  type CreateKnowledgeDocumentPayload,
  type KnowledgeDocument,
  type PageResult,
} from "@/lib/api/admin";
import { getEnumLabel, getEnumOptions } from "@/lib/enums";
import {
  KnowledgeDocumentIndexStatus,
  KnowledgeDocumentIndexStatusLabels,
  StatusLabels
} from "@/lib/generated/enums";
import { cn, formatDateTime } from "@/lib/utils";
import { DocumentEditDialog } from "./document-edit";

type DocumentListProps = {
  knowledgeBaseId: number | null;
  onActionStateChange?: (state: DocumentListActionState) => void;
};

export type DocumentListActionState = {
  onRefresh: () => void;
  onChangeViewMode: (mode: "list" | "grid") => void;
  onCreate: () => void;
  viewMode: "list" | "grid";
  loading: boolean;
};

const statusOptions = [
  { value: "all", label: "全部状态" },
  ...getEnumOptions(StatusLabels),
] as const;

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

function getDocumentPreview(content: string, contentType: string) {
  const preview =
    contentType === "markdown"
      ? content
          .replace(/[`*_>#-]/g, " ")
          .replace(/\[(.*?)\]\((.*?)\)/g, "$1")
          .replace(/\s+/g, " ")
          .trim()
      : content
          .replace(/<[^>]+>/g, " ")
          .replace(/\s+/g, " ")
          .trim();
  return preview || "暂无内容";
}

const VIEW_MODE_STORAGE_KEY = "knowledge-document-view-mode";

export function DocumentList({ knowledgeBaseId, onActionStateChange }: DocumentListProps) {
  const [keywordInput, setKeywordInput] = useState("");
  const [statusFilterInput, setStatusFilterInput] = useState("all");
  const [indexStatusFilterInput, setIndexStatusFilterInput] = useState("all");
  const [keyword, setKeyword] = useState("");
  const [statusFilter, setStatusFilter] = useState("all");
  const [indexStatusFilter, setIndexStatusFilter] = useState("all");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [actionLoadingMap, setActionLoadingMap] = useState<Record<number, { rebuildIndex: boolean; delete: boolean }>>({});
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingItem, setEditingItem] = useState<KnowledgeDocument | null>(
    null,
  );
  const [viewMode, setViewMode] = useState<"list" | "grid">(() => {
    if (typeof window === "undefined") return "grid";
    const saved = localStorage.getItem(VIEW_MODE_STORAGE_KEY);
    return saved === "list" || saved === "grid" ? saved : "grid";
  });
  const [documents, setDocuments] = useState<PageResult<KnowledgeDocument>>({
    results: [],
    page: { page: 1, limit: 20, total: 0 },
  });

  const loadData = useCallback(async () => {
    if (!knowledgeBaseId) {
      setDocuments({ results: [], page: { page: 1, limit: 20, total: 0 } });
      setLoading(false);
      return;
    }

    setLoading(true);
    try {
      const data = await fetchKnowledgeDocuments({
        title: keyword.trim() || undefined,
        status: statusFilter === "all" ? undefined : statusFilter,
        indexStatus: indexStatusFilter === "all" ? undefined : indexStatusFilter,
        knowledgeBaseId,
        limit: 1000,
      });
      setDocuments(data);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "加载文档失败");
    } finally {
      setLoading(false);
    }
  }, [indexStatusFilter, keyword, statusFilter, knowledgeBaseId]);

  useEffect(() => {
    void loadData();
  }, [loadData]);

  useEffect(() => {
    localStorage.setItem(VIEW_MODE_STORAGE_KEY, viewMode);
  }, [viewMode]);

  function handleStatusFilterChange(value: string | null) {
    const newValue = value ?? "all";
    setStatusFilterInput(newValue);
    setStatusFilter(newValue);
  }

  function handleIndexStatusFilterChange(value: string | null) {
    const newValue = value ?? "all";
    setIndexStatusFilterInput(newValue);
    setIndexStatusFilter(newValue);
  }

  function applyFilters() {
    setKeyword(keywordInput);
    setStatusFilter(statusFilterInput);
    setIndexStatusFilter(indexStatusFilterInput);
  }

  function handleFilterKeyDown(event: React.KeyboardEvent<HTMLInputElement>) {
    if (event.key !== "Enter") {
      return;
    }
    event.preventDefault();
    applyFilters();
  }

  const openCreateDialog = useCallback(() => {
    setEditingItem(null);
    setDialogOpen(true);
  }, []);

  useEffect(() => {
    if (!onActionStateChange) {
      return;
    }

    onActionStateChange({
      onRefresh: () => void loadData(),
      onChangeViewMode: setViewMode,
      onCreate: openCreateDialog,
      viewMode,
      loading,
    });
  }, [onActionStateChange, loadData, openCreateDialog, viewMode, loading]);

  function openEditDialog(item: KnowledgeDocument) {
    setEditingItem(item);
    setDialogOpen(true);
  }

  function handleDialogOpenChange(open: boolean) {
    if (saving) {
      return;
    }
    if (!open) {
      setEditingItem(null);
    }
    setDialogOpen(open);
  }

  async function handleSubmit(payload: CreateKnowledgeDocumentPayload) {
    if (saving) {
      return;
    }

    setSaving(true);
    try {
      if (editingItem) {
        await updateKnowledgeDocument({
          id: editingItem.id,
          ...payload,
        });
        toast.success(`已更新文档：${editingItem.title}`);
      } else {
        await createKnowledgeDocument(payload);
        toast.success(`已创建文档：${payload.title}`);
      }
      setDialogOpen(false);
      setEditingItem(null);
      await loadData();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "保存文档失败");
    } finally {
      setSaving(false);
    }
  }

  async function handleDelete(item: KnowledgeDocument) {
    setActionLoadingMap((prev) => ({ ...prev, [item.id]: { ...prev[item.id], delete: true } }));
    try {
      await deleteKnowledgeDocument(item.id);
      toast.success(`已删除文档：${item.title}`);
      await loadData();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "删除文档失败");
    } finally {
      setActionLoadingMap((prev) => ({ ...prev, [item.id]: { ...prev[item.id], delete: false } }));
    }
  }

  async function handleBuildIndex(item: KnowledgeDocument) {
    setActionLoadingMap((prev) => ({ ...prev, [item.id]: { ...prev[item.id], rebuildIndex: true } }));
    try {
      await buildKnowledgeDocumentIndex(item.id);
      toast.success(`已重建索引：${item.title}`);
      await loadData();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "重建索引失败");
    } finally {
      setActionLoadingMap((prev) => ({ ...prev, [item.id]: { ...prev[item.id], rebuildIndex: false } }));
    }
  }

  if (!knowledgeBaseId) {
    return (
      <div className="flex h-full flex-col items-center justify-center text-muted-foreground">
        <FileTextIcon className="mb-2 size-12 opacity-50" />
        <p>请选择一个知识库查看文档</p>
      </div>
    );
  }

  return (
    <>
      <div className="flex h-full min-h-0 flex-col">
        <div className="flex flex-col gap-2 border-b bg-background px-6 py-2">
          <div className="flex gap-2">
            <div className="relative flex-1">
              <SearchIcon className="pointer-events-none absolute top-1/2 left-2 size-3.5 -translate-y-1/2 text-muted-foreground" />
              <Input
                value={keywordInput}
                onChange={(event) => setKeywordInput(event.target.value)}
                onKeyDown={handleFilterKeyDown}
                placeholder="搜索文档标题"
                className="h-8 pl-8 text-xs"
              />
            </div>
            <Select
              value={statusFilterInput}
              onValueChange={handleStatusFilterChange}
            >
              <SelectTrigger className="h-8 w-32 text-xs">
                <SelectValue>
                  {statusFilterInput === "all" ? "全部状态" : getEnumLabel(StatusLabels, Number(statusFilterInput))}
                </SelectValue>
              </SelectTrigger>
              <SelectContent>
                {statusOptions.map((item) => (
                  <SelectItem
                    key={item.value}
                    value={item.value}
                    className="text-xs"
                  >
                    {item.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Select
              value={indexStatusFilterInput}
              onValueChange={handleIndexStatusFilterChange}
            >
              <SelectTrigger className="h-8 w-36 text-xs">
                <SelectValue>
                  {indexStatusFilterInput === "all"
                    ? "全部索引状态"
                    : getEnumLabel(
                        KnowledgeDocumentIndexStatusLabels,
                        indexStatusFilterInput as KnowledgeDocumentIndexStatus
                      )}
                </SelectValue>
              </SelectTrigger>
              <SelectContent>
                {indexStatusOptions.map((item) => (
                  <SelectItem
                    key={item.value}
                    value={item.value}
                    className="text-xs"
                  >
                    {item.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </div>
        <div className="min-h-0 flex-1">
          <ScrollArea className="h-full">
            <div className={viewMode === "grid" ? "p-2 space-y-1" : "p-2 space-y-0.5"}>
            {documents.results.map((item) => (
              viewMode === "grid" ? (
                <ContextMenu key={item.id}>
                  <ContextMenuTrigger className="w-full">
                    <div
                      className="bg-background p-3 transition-colors hover:bg-accent w-full"
                    >
                      <div className="flex items-start justify-between gap-2">
                        <div className="flex min-w-0 flex-1 items-start gap-2">
                          {/* <FileTextIcon className="mt-0.5 size-4 shrink-0 text-muted-foreground" /> */}
                          <div className="min-w-0 flex-1">
                            <div className="flex items-center gap-2">
                              <div className="text-sm font-medium">{item.title}</div>
                              <TooltipProvider>
                                <Tooltip>
                                  <TooltipTrigger>
                                    <Badge variant={getIndexStatusBadgeVariant(item.indexStatus)}>
                                      {item.indexStatusName}
                                    </Badge>
                                  </TooltipTrigger>
                                  <TooltipContent align="start">
                                    <div className="space-y-1">
                                      <div>索引状态：{item.indexStatusName}</div>
                                      <div>索引时间：{formatDateTime(item.indexedAt)}</div>
                                      {item.indexError ? <div>失败原因：{item.indexError}</div> : null}
                                    </div>
                                  </TooltipContent>
                                </Tooltip>
                              </TooltipProvider>
                              {item.indexError ? (
                                <TooltipProvider>
                                  <Tooltip>
                                    <TooltipTrigger>
                                      <AlertCircleIcon className="size-3.5 text-destructive" />
                                    </TooltipTrigger>
                                    <TooltipContent align="start">
                                      {item.indexError}
                                    </TooltipContent>
                                  </Tooltip>
                                </TooltipProvider>
                              ) : null}
                            </div>
                            <div className="mt-1 text-xs text-muted-foreground line-clamp-2">
                              {getDocumentPreview(item.content, item.contentType)}
                            </div>
                            <div className="mt-2 flex flex-wrap items-center gap-x-2 gap-y-1 text-xs text-muted-foreground">
                              <span>{item.createUserName || "-"}</span>
                              <span>{formatDateTime(item.createdAt)}</span>
                              <span className={cn(item.indexStatus === KnowledgeDocumentIndexStatus.Failed && "text-destructive")}>
                                {item.indexStatus === KnowledgeDocumentIndexStatus.Indexed
                                  ? `已索引 ${formatDateTime(item.indexedAt)}`
                                  : item.indexStatusName}
                              </span>
                            </div>
                          </div>
                        </div>
                        <DropdownMenu>
                          <DropdownMenuTrigger
                            render={
                              <Button
                                variant="ghost"
                                size="icon"
                                className="size-6"
                              />
                            }
                            aria-label={`更多操作 ${item.title}`}
                          >
                            <MoreHorizontalIcon className="size-3.5" />
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end" className="w-32 min-w-32">
                            <DropdownMenuItem onClick={() => openEditDialog(item)}>
                              <PencilIcon className="mr-2 size-3.5" />
                              编辑
                            </DropdownMenuItem>
                            <DropdownMenuItem onClick={() => void handleBuildIndex(item)}>
                              <WrenchIcon className="mr-2 size-3.5" />
                              {actionLoadingMap[item.id]?.rebuildIndex ? "执行中..." : "重建索引"}
                            </DropdownMenuItem>
                            <DropdownMenuItem
                              onClick={() => void handleDelete(item)}
                              className="text-destructive focus:text-destructive"
                            >
                              <Trash2Icon className="mr-2 size-3.5" />
                              {actionLoadingMap[item.id]?.delete ? "删除中..." : "删除"}
                            </DropdownMenuItem>
                          </DropdownMenuContent>
                        </DropdownMenu>
                      </div>
                    </div>
                  </ContextMenuTrigger>
                  <ContextMenuContent className="w-40">
                    <ContextMenuItem onClick={() => openEditDialog(item)}>
                      <PencilIcon className="mr-2 size-3.5" />
                      编辑
                    </ContextMenuItem>
                    <ContextMenuItem onClick={() => void handleBuildIndex(item)} disabled={actionLoadingMap[item.id]?.rebuildIndex}>
                      <WrenchIcon className="mr-2 size-3.5" />
                      {actionLoadingMap[item.id]?.rebuildIndex ? "执行中..." : "重建索引"}
                    </ContextMenuItem>
                    <ContextMenuItem
                      onClick={() => void handleDelete(item)}
                      variant="destructive"
                      disabled={actionLoadingMap[item.id]?.delete}
                    >
                      <Trash2Icon className="mr-2 size-3.5" />
                      {actionLoadingMap[item.id]?.delete ? "删除中..." : "删除"}
                    </ContextMenuItem>
                  </ContextMenuContent>
                </ContextMenu>
              ) : (
                <ContextMenu key={item.id}>
                  <ContextMenuTrigger className="w-full">
                    <div
                      className="flex items-center gap-3 bg-background p-2 transition-colors hover:bg-accent w-full"
                    >
                      {/* <FileTextIcon className="size-4 shrink-0 text-muted-foreground" /> */}
                      <div className="min-w-0 flex-1">
                        <div className="flex items-center gap-2">
                          <div className="truncate text-sm font-medium">{item.title}</div>
                          <Badge variant={getIndexStatusBadgeVariant(item.indexStatus)}>
                            {item.indexStatusName}
                          </Badge>
                          {item.indexError ? (
                            <TooltipProvider>
                              <Tooltip>
                                <TooltipTrigger>
                                  <AlertCircleIcon className="size-3.5 text-destructive" />
                                </TooltipTrigger>
                                <TooltipContent align="start">
                                  <div className="space-y-1">
                                    <div>索引时间：{formatDateTime(item.indexedAt)}</div>
                                    <div>失败原因：{item.indexError}</div>
                                  </div>
                                </TooltipContent>
                              </Tooltip>
                            </TooltipProvider>
                          ) : null}
                        </div>
                        <div className="mt-0.5 truncate text-xs text-muted-foreground">
                          {getDocumentPreview(item.content, item.contentType)}
                        </div>
                        <div className="mt-1 text-xs text-muted-foreground">
                          {item.indexStatus === KnowledgeDocumentIndexStatus.Indexed
                            ? `索引时间：${formatDateTime(item.indexedAt)}`
                            : item.indexError || item.indexStatusName}
                        </div>
                      </div>
                      <div className="shrink-0 text-xs text-muted-foreground">
                        {formatDateTime(item.createdAt)}
                      </div>
                      <DropdownMenu>
                        <DropdownMenuTrigger
                          render={
                            <Button
                              variant="ghost"
                              size="icon"
                              className="size-6"
                            />
                          }
                          aria-label={`更多操作 ${item.title}`}
                        >
                          <MoreHorizontalIcon className="size-3.5" />
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end" className="w-32 min-w-32">
                          <DropdownMenuItem onClick={() => openEditDialog(item)}>
                            <PencilIcon className="mr-2 size-3.5" />
                            编辑
                          </DropdownMenuItem>
                          <DropdownMenuItem onClick={() => void handleBuildIndex(item)}>
                            <WrenchIcon className="mr-2 size-3.5" />
                            {actionLoadingMap[item.id]?.rebuildIndex ? "执行中..." : "重建索引"}
                          </DropdownMenuItem>
                          <DropdownMenuItem
                            onClick={() => void handleDelete(item)}
                            className="text-destructive focus:text-destructive"
                          >
                            <Trash2Icon className="mr-2 size-3.5" />
                            {actionLoadingMap[item.id]?.delete ? "删除中..." : "删除"}
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </div>
                  </ContextMenuTrigger>
                  <ContextMenuContent className="w-40">
                    <ContextMenuItem onClick={() => openEditDialog(item)}>
                      <PencilIcon className="mr-2 size-3.5" />
                      编辑
                    </ContextMenuItem>
                    <ContextMenuItem onClick={() => void handleBuildIndex(item)} disabled={actionLoadingMap[item.id]?.rebuildIndex}>
                      <WrenchIcon className="mr-2 size-3.5" />
                      {actionLoadingMap[item.id]?.rebuildIndex ? "执行中..." : "重建索引"}
                    </ContextMenuItem>
                    <ContextMenuItem
                      onClick={() => void handleDelete(item)}
                      variant="destructive"
                      disabled={actionLoadingMap[item.id]?.delete}
                    >
                      <Trash2Icon className="mr-2 size-3.5" />
                      {actionLoadingMap[item.id]?.delete ? "删除中..." : "删除"}
                    </ContextMenuItem>
                  </ContextMenuContent>
                </ContextMenu>
              )
            ))}
            {!loading && documents.results.length === 0 ? (
              <div className="py-8 text-center text-sm text-muted-foreground">
                没有匹配的文档
              </div>
            ) : null}
            </div>
          </ScrollArea>
        </div>
      </div>
      <DocumentEditDialog
        open={dialogOpen}
        saving={saving}
        itemId={editingItem?.id ?? null}
        knowledgeBaseId={knowledgeBaseId}
        onOpenChange={handleDialogOpenChange}
        onSubmit={handleSubmit}
      />
    </>
  );
}
