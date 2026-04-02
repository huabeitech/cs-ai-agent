"use client";

import {
  DndContext,
  KeyboardSensor,
  MouseSensor,
  TouchSensor,
  closestCenter,
  useSensor,
  useSensors,
  type DragEndEvent,
} from "@dnd-kit/core";
import {
  SortableContext,
  arrayMove,
  sortableKeyboardCoordinates,
  useSortable,
  verticalListSortingStrategy,
} from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import {
  ChevronRightIcon,
  FolderIcon,
  MoreHorizontalIcon,
  PencilIcon,
  PlusIcon,
  RefreshCwIcon,
  SearchIcon,
  Trash2Icon,
} from "lucide-react";
import type { CSSProperties } from "react";
import { useCallback, useEffect, useState } from "react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
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
  createKnowledgeBase,
  deleteKnowledgeBase,
  fetchKnowledgeBases,
  rebuildKnowledgeBaseIndex,
  updateKnowledgeBase,
  updateKnowledgeBaseSort,
  type CreateKnowledgeBasePayload,
  type KnowledgeBase,
} from "@/lib/api/admin";
import { StatusLabels } from "@/lib/generated/enums";
import { getEnumLabel, getEnumOptions } from "@/lib/enums";
import { cn } from "@/lib/utils";
import { EditDialog } from "./knowledge-base-edit";

type KnowledgeBaseListProps = {
  selectedKnowledgeBaseId: number | null;
  onSelectKnowledgeBase: (knowledgeBase: KnowledgeBase | null) => void;
};

const statusOptions = [
  { value: "all", label: "全部状态" },
  ...getEnumOptions(StatusLabels),
] as const;

type SortableKnowledgeBaseCardProps = {
  item: KnowledgeBase;
  isSelected: boolean;
  disabled: boolean;
  onSelect: () => void;
  onEdit: () => void;
  onDelete: () => void;
  onRebuildIndex: () => void;
  deleteLoadingId: number | null;
  rebuildIndexLoadingId: number | null;
};

function SortableKnowledgeBaseCard({
  item,
  isSelected,
  disabled,
  onSelect,
  onEdit,
  onDelete,
  onRebuildIndex,
  deleteLoadingId,
  rebuildIndexLoadingId,
}: SortableKnowledgeBaseCardProps) {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({
    id: item.id,
    disabled,
  });

  const style: CSSProperties = {
    transform: CSS.Transform.toString(transform),
    transition,
  };

  return (
    <ContextMenu>
      <ContextMenuTrigger
        ref={setNodeRef}
        style={style}
        className={cn(
          "group flex items-center gap-1 rounded mx-2 px-2 py-1.5 text-sm transition-colors hover:bg-accent cursor-pointer",
          isSelected && "bg-accent text-accent-foreground",
          isDragging && "bg-muted/60 shadow-sm opacity-80",
        )}
        onClick={onSelect}
        onKeyDown={(e) => {
          if (e.key === "Enter" || e.key === " ") {
            e.preventDefault();
            onSelect();
          }
        }}
        {...attributes}
        {...listeners}
      >
        <ChevronRightIcon className="size-4 shrink-0 text-muted-foreground opacity-0 group-hover:opacity-100 transition-opacity" />
        <FolderIcon className="size-4 shrink-0 text-muted-foreground" />
        <span className="min-w-0 flex-1 truncate">{item.name}</span>
        <span className="shrink-0 text-xs text-muted-foreground">
          {item.knowledgeType === "faq" ? item.faqCount : item.documentCount}
        </span>
        <DropdownMenu>
          <DropdownMenuTrigger
            render={
              <Button
                variant="ghost"
                size="icon"
                className="size-6 opacity-0 group-hover:opacity-100"
              />
            }
            aria-label={`更多操作 ${item.name}`}
          >
            <MoreHorizontalIcon className="size-3.5" />
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-40 min-w-40">
            <DropdownMenuItem
              onClick={(e) => {
                e.stopPropagation();
                onEdit();
              }}
            >
              <PencilIcon className="mr-2 size-3.5" />
              编辑
            </DropdownMenuItem>
            <DropdownMenuItem
              onClick={(e) => {
                e.stopPropagation();
                onRebuildIndex();
              }}
            >
              <RefreshCwIcon className="mr-2 size-3.5" />
              {rebuildIndexLoadingId === item.id ? "重建中..." : "重建索引"}
            </DropdownMenuItem>
            <DropdownMenuItem
              onClick={(e) => {
                e.stopPropagation();
                onDelete();
              }}
              className="text-destructive focus:text-destructive"
            >
              <Trash2Icon className="mr-2 size-3.5" />
              {deleteLoadingId === item.id ? "删除中..." : "删除"}
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </ContextMenuTrigger>
      <ContextMenuContent>
        <ContextMenuItem
          onClick={(e) => {
            e.stopPropagation();
            onEdit();
          }}
        >
          <PencilIcon className="mr-2 size-3.5" />
          编辑
        </ContextMenuItem>
        <ContextMenuItem
          onClick={(e) => {
            e.stopPropagation();
            onRebuildIndex();
          }}
        >
          <RefreshCwIcon className="mr-2 size-3.5" />
          {rebuildIndexLoadingId === item.id ? "重建中..." : "重建索引"}
        </ContextMenuItem>
        <ContextMenuItem
          onClick={(e) => {
            e.stopPropagation();
            onDelete();
          }}
          variant="destructive"
        >
          <Trash2Icon className="mr-2 size-3.5" />
          {deleteLoadingId === item.id ? "删除中..." : "删除"}
        </ContextMenuItem>
      </ContextMenuContent>
    </ContextMenu>
  );
}

export function KnowledgeBaseList({
  selectedKnowledgeBaseId,
  onSelectKnowledgeBase,
}: KnowledgeBaseListProps) {
  const [keywordInput, setKeywordInput] = useState("");
  const [statusFilterInput, setStatusFilterInput] = useState("all");
  const [keyword, setKeyword] = useState("");
  const [statusFilter, setStatusFilter] = useState("all");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [sorting, setSorting] = useState(false);
  const [deleteLoadingId, setDeleteLoadingId] = useState<number | null>(null);
  const [rebuildIndexLoadingId, setRebuildIndexLoadingId] = useState<
    number | null
  >(null);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingItemId, setEditingItemId] = useState<number | null>(null);
  const [knowledgeBases, setKnowledgeBases] = useState<KnowledgeBase[]>([]);

  const sensors = useSensors(
    useSensor(MouseSensor, {
      activationConstraint: { distance: 8 },
    }),
    useSensor(TouchSensor, {
      activationConstraint: { delay: 150, tolerance: 8 },
    }),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    }),
  );

  const loadData = useCallback(async () => {
    setLoading(true);
    try {
      const data = await fetchKnowledgeBases({
        name: keyword.trim() || undefined,
        status: statusFilter === "all" ? undefined : statusFilter,
        limit: 1000,
      });
      setKnowledgeBases(data.results);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "加载知识库失败");
    } finally {
      setLoading(false);
    }
  }, [keyword, statusFilter]);

  useEffect(() => {
    void loadData();
  }, [loadData]);

  useEffect(() => {
    if (
      selectedKnowledgeBaseId === null &&
      knowledgeBases.length > 0 &&
      !loading
    ) {
      onSelectKnowledgeBase(knowledgeBases[0]);
    }
  }, [selectedKnowledgeBaseId, knowledgeBases, loading, onSelectKnowledgeBase]);

  function handleStatusFilterChange(value: string | null) {
    setStatusFilterInput(value ?? "all");
  }

  function applyFilters() {
    setKeyword(keywordInput);
    setStatusFilter(statusFilterInput);
  }

  function handleFilterKeyDown(event: React.KeyboardEvent<HTMLInputElement>) {
    if (event.key !== "Enter") {
      return;
    }
    event.preventDefault();
    applyFilters();
  }

  function openCreateDialog() {
    setEditingItemId(null);
    setDialogOpen(true);
  }

  function openEditDialog(item: KnowledgeBase) {
    setEditingItemId(item.id);
    setDialogOpen(true);
  }

  function handleDialogOpenChange(open: boolean) {
    if (saving) {
      return;
    }
    if (!open) {
      setEditingItemId(null);
    }
    setDialogOpen(open);
  }

  async function handleSubmit(payload: CreateKnowledgeBasePayload) {
    if (saving) {
      return;
    }

    setSaving(true);
    try {
      if (editingItemId) {
        await updateKnowledgeBase({
          id: editingItemId,
          ...payload,
        });
        const editingItem = knowledgeBases.find(
          (item) => item.id === editingItemId,
        );
        toast.success(`已更新知识库：${editingItem?.name || payload.name}`);
      } else {
        await createKnowledgeBase(payload);
        toast.success(`已创建知识库：${payload.name}`);
      }
      setDialogOpen(false);
      setEditingItemId(null);
      await loadData();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "保存知识库失败");
    } finally {
      setSaving(false);
    }
  }

  async function handleDelete(item: KnowledgeBase) {
    setDeleteLoadingId(item.id);
    try {
      await deleteKnowledgeBase(item.id);
      toast.success(`已删除知识库：${item.name}`);
      if (selectedKnowledgeBaseId === item.id) {
        onSelectKnowledgeBase(null);
      }
      await loadData();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "删除知识库失败");
    } finally {
      setDeleteLoadingId(null);
    }
  }

  async function handleRebuildIndex(item: KnowledgeBase) {
    setRebuildIndexLoadingId(item.id);
    try {
      await rebuildKnowledgeBaseIndex(item.id);
      toast.success(`已开始重建知识库索引：${item.name}`);
    } catch (error) {
      toast.error(
        error instanceof Error ? error.message : "重建知识库索引失败",
      );
    } finally {
      setRebuildIndexLoadingId(null);
    }
  }

  async function handleDragEnd(event: DragEndEvent) {
    const { active, over } = event;
    if (!over || active.id === over.id || sorting) {
      return;
    }

    const previousResults = knowledgeBases;
    const oldIndex = previousResults.findIndex((item) => item.id === active.id);
    const newIndex = previousResults.findIndex((item) => item.id === over.id);
    if (oldIndex < 0 || newIndex < 0) {
      return;
    }

    const nextResults = arrayMove(previousResults, oldIndex, newIndex);
    setKnowledgeBases(nextResults);
    setSorting(true);

    try {
      await updateKnowledgeBaseSort(nextResults.map((item) => item.id));
      toast.success("知识库排序已更新");
      await loadData();
    } catch (error) {
      setKnowledgeBases(previousResults);
      toast.error(
        error instanceof Error ? error.message : "更新知识库排序失败",
      );
    } finally {
      setSorting(false);
    }
  }

  return (
    <>
      <div className="flex h-full flex-col border-r bg-muted/30">
        <div className="flex flex-col gap-2 border-b bg-background p-4">
          <div className="flex items-center justify-between">
            <h2 className="text-sm font-semibold">知识库</h2>
            <div className="flex items-center gap-1">
              <Button
                variant="ghost"
                size="icon"
                className="size-7"
                onClick={() => void loadData()}
                disabled={loading || sorting}
              >
                <RefreshCwIcon
                  className={loading || sorting ? "animate-spin" : "size-4"}
                />
              </Button>
              <Button
                variant="ghost"
                size="icon"
                className="size-7"
                onClick={openCreateDialog}
              >
                <PlusIcon className="size-4" />
              </Button>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <div className="relative flex-1">
              <SearchIcon className="pointer-events-none absolute top-1/2 left-2 size-3.5 -translate-y-1/2 text-muted-foreground" />
              <Input
                value={keywordInput}
                onChange={(event) => setKeywordInput(event.target.value)}
                onKeyDown={handleFilterKeyDown}
                placeholder="搜索知识库"
                className="h-8 pl-8 text-xs"
              />
            </div>
            <Select
              value={statusFilterInput}
              onValueChange={handleStatusFilterChange}
            >
              <SelectTrigger className="h-8 w-28 text-xs">
                <SelectValue>
                  {statusFilterInput === "all"
                    ? "全部状态"
                    : getEnumLabel(StatusLabels, Number(statusFilterInput))}
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
          </div>
        </div>
        <ScrollArea className="flex-1">
          <div className="py-1 space-y-0.5">
            <DndContext
              sensors={sensors}
              collisionDetection={closestCenter}
              onDragEnd={(event) => void handleDragEnd(event)}
            >
              <SortableContext
                items={knowledgeBases.map((item) => item.id)}
                strategy={verticalListSortingStrategy}
              >
                {knowledgeBases.map((item) => (
                  <SortableKnowledgeBaseCard
                    key={item.id}
                    item={item}
                    isSelected={selectedKnowledgeBaseId === item.id}
                    disabled={loading || sorting}
                    onSelect={() => onSelectKnowledgeBase(item)}
                    onEdit={() => openEditDialog(item)}
                    onDelete={() => void handleDelete(item)}
                    onRebuildIndex={() => void handleRebuildIndex(item)}
                    deleteLoadingId={deleteLoadingId}
                    rebuildIndexLoadingId={rebuildIndexLoadingId}
                  />
                ))}
              </SortableContext>
            </DndContext>
            {!loading && knowledgeBases.length === 0 ? (
              <div className="py-8 text-center text-sm text-muted-foreground">
                没有匹配的知识库
              </div>
            ) : null}
          </div>
        </ScrollArea>
      </div>
      <EditDialog
        open={dialogOpen}
        saving={saving}
        itemId={editingItemId}
        onOpenChange={handleDialogOpenChange}
        onSubmit={handleSubmit}
      />
    </>
  );
}
