"use client";

import {
  closestCenter,
  DndContext,
  KeyboardSensor,
  MouseSensor,
  TouchSensor,
  useSensor,
  useSensors,
  type DragEndEvent,
} from "@dnd-kit/core";
import {
  arrayMove,
  SortableContext,
  sortableKeyboardCoordinates,
  useSortable,
  verticalListSortingStrategy,
} from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import {
  GripVerticalIcon,
  MoreHorizontalIcon,
  PlusIcon,
  RefreshCwIcon,
  SearchIcon,
  Trash2Icon
} from "lucide-react";
import { useCallback, useEffect, useState, type CSSProperties } from "react";
import { toast } from "sonner";

import { ListPagination } from "@/components/list-pagination";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { ButtonGroup } from "@/components/ui/button-group";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  createAIConfig,
  deleteAIConfig,
  fetchAIConfigs,
  updateAIConfig,
  updateAIConfigSort,
  updateAIConfigStatus,
  type AIConfig,
  type CreateAIConfigPayload,
  type PageResult,
} from "@/lib/api/admin";
import {
  AIModelType,
  AIModelTypeLabels,
  AIProvider,
  AIProviderLabels,
  Status,
  StatusLabels
} from "@/lib/generated/enums";
import { getEnumLabel, getEnumOptions } from "@/lib/enums";
import { cn } from "@/lib/utils";
import { EditDialog } from "./_components/edit";
import { OptionCombobox } from "./_components/option-combobox";

const listStatusOptions = [
  { value: "all", label: "全部状态" },
  ...getEnumOptions(StatusLabels).map((option) => ({
    value: String(option.value),
    label: option.label,
  })),
];

const providerFilterOptions = [
  { value: "all", label: "全部供应商" },
  ...getEnumOptions(AIProviderLabels).map((option) => ({
    value: String(option.value),
    label: option.label,
  })),
];

const modelTypeFilterOptions = [
  { value: "all", label: "全部类型" },
  ...getEnumOptions(AIModelTypeLabels).map((option) => ({
    value: String(option.value),
    label: option.label,
  })),
];

function maskAPIKey(value: string) {
  const text = value.trim();
  if (!text) {
    return "-";
  }
  if (text.length <= 8) {
    return "****";
  }
  return `${text.slice(0, 4)}****${text.slice(-4)}`;
}

type SortableAIConfigRowProps = {
  item: AIConfig;
  disabled: boolean;
  actionLoadingId: number | null;
  openEditDialog: (item: AIConfig) => void;
  handleToggleStatus: (item: AIConfig) => void;
  handleDelete: (item: AIConfig) => void;
};

function SortableAIConfigRow({
  item,
  disabled,
  actionLoadingId,
  openEditDialog,
  handleToggleStatus,
  handleDelete,
}: SortableAIConfigRowProps) {
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
    <TableRow
      ref={setNodeRef}
      style={style}
      className={cn(
        isDragging && "relative z-10 bg-muted/60 shadow-sm",
        !disabled && "cursor-move",
      )}
    >
      <TableCell className="w-14">
        <Button
          type="button"
          variant="ghost"
          size="icon"
          className="size-8 cursor-grab active:cursor-grabbing"
          disabled={disabled}
          aria-label={`拖拽排序 ${item.name}`}
          {...attributes}
          {...listeners}
        >
          <GripVerticalIcon className="size-4 text-muted-foreground" />
        </Button>
      </TableCell>
      <TableCell>
        <div className="space-y-1 text-sm font-medium">{item.name}</div>
      </TableCell>
      <TableCell>
        <Badge variant="outline">
          {getEnumLabel(
            AIProviderLabels,
            item.provider as AIProvider,
          )}
        </Badge>
      </TableCell>
      <TableCell>
        <div className="space-y-1">
          <Badge variant="secondary">
            {getEnumLabel(
              AIModelTypeLabels,
              item.modelType as AIModelType,
            )}
          </Badge>
          <div className="text-sm">{item.modelName}</div>
          {item.dimension > 0 && (
            <div className="text-xs text-muted-foreground">
              {item.dimension} 维
            </div>
          )}
        </div>
      </TableCell>
      <TableCell>
        <div className="space-y-1 text-sm">
          <div className="line-clamp-1">{item.baseUrl}</div>
          <div className="text-xs text-muted-foreground">
            Key: {maskAPIKey(item.apiKey)}
          </div>
        </div>
      </TableCell>
      <TableCell>
        <div className="space-y-1 text-xs text-muted-foreground">
          <div>上下文 {item.maxContextTokens || 0}</div>
          <div>输出 {item.maxOutputTokens || 0}</div>
          <div>
            超时 {item.timeoutMs}ms / 重试 {item.maxRetryCount}
          </div>
          <div>
            RPM {item.rpmLimit || 0} / TPM {item.tpmLimit || 0}
          </div>
        </div>
      </TableCell>
      <TableCell>
        <div className="flex items-center gap-3">
          <Switch
            checked={item.status === Status.Ok}
            disabled={actionLoadingId === item.id}
            onCheckedChange={() => void handleToggleStatus(item)}
            aria-label={`${item.name} 状态切换`}
          />
          <Badge
            variant={
              item.status === Status.Ok ? "default" : "outline"
            }
          >
            {getEnumLabel(
              StatusLabels,
              item.status as keyof typeof StatusLabels,
            )}
          </Badge>
        </div>
      </TableCell>
      <TableCell className="text-right">
        <ButtonGroup className="ml-auto">
          <Button
            variant="outline"
            size="sm"
            onClick={() => openEditDialog(item)}
          >
            编辑
          </Button>
          <DropdownMenu>
            <DropdownMenuTrigger
              render={<Button variant="outline" size="icon-sm" />}
              aria-label={`更多操作 ${item.name}`}
            >
              <MoreHorizontalIcon />
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-40 min-w-40">
              <DropdownMenuItem
                disabled={
                  item.status === Status.Ok ||
                  actionLoadingId === item.id
                }
                onClick={() => void handleDelete(item)}
                className="text-destructive focus:text-destructive"
              >
                <Trash2Icon />
                {item.status === Status.Ok
                  ? "启用中不可删"
                  : actionLoadingId === item.id
                    ? "删除中..."
                    : "删除"}
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </ButtonGroup>
      </TableCell>
    </TableRow>
  );
}

export default function DashboardAIConfigsPage() {
  const [keywordInput, setKeywordInput] = useState("");
  const [statusFilterInput, setStatusFilterInput] = useState("all");
  const [providerFilterInput, setProviderFilterInput] = useState("all");
  const [modelTypeFilterInput, setModelTypeFilterInput] = useState("all");
  const [keyword, setKeyword] = useState("");
  const [statusFilter, setStatusFilter] = useState("all");
  const [providerFilter, setProviderFilter] = useState("all");
  const [modelTypeFilter, setModelTypeFilter] = useState("all");
  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState(20);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [actionLoadingId, setActionLoadingId] = useState<number | null>(null);
  const [sorting, setSorting] = useState(false);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingItem, setEditingItem] = useState<AIConfig | null>(null);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [deletingItem, setDeletingItem] = useState<AIConfig | null>(null);
  const [result, setResult] = useState<PageResult<AIConfig>>({
    results: [],
    page: { page: 1, limit: 20, total: 0 },
  });

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
      const data = await fetchAIConfigs({
        name: keyword.trim() || undefined,
        status: statusFilter === "all" ? undefined : statusFilter,
        provider: providerFilter === "all" ? undefined : providerFilter,
        modelType: modelTypeFilter === "all" ? undefined : modelTypeFilter,
        page,
        limit,
      });
      setResult(data);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "加载 AI 配置失败");
    } finally {
      setLoading(false);
    }
  }, [keyword, statusFilter, providerFilter, modelTypeFilter, page, limit]);

  useEffect(() => {
    void loadData();
  }, [loadData]);

  function applyFilters() {
    setKeyword(keywordInput);
    setStatusFilter(statusFilterInput);
    setProviderFilter(providerFilterInput);
    setModelTypeFilter(modelTypeFilterInput);
    setPage(1);
  }

  function handleFilterKeyDown(event: React.KeyboardEvent<HTMLInputElement>) {
    if (event.key !== "Enter") {
      return;
    }
    event.preventDefault();
    applyFilters();
  }

  function handlePageChange(nextPage: number) {
    if (nextPage < 1 || nextPage === page) {
      return;
    }
    setPage(nextPage);
  }

  function handleLimitChange(nextLimit: number) {
    if (nextLimit <= 0 || nextLimit === limit) {
      return;
    }
    setLimit(nextLimit);
    setPage(1);
  }

  function openCreateDialog() {
    setEditingItem(null);
    setDialogOpen(true);
  }

  function openEditDialog(item: AIConfig) {
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

  async function handleSubmit(payload: CreateAIConfigPayload) {
    if (saving) {
      return;
    }

    setSaving(true);
    try {
      if (editingItem) {
        await updateAIConfig({ id: editingItem.id, ...payload });
        toast.success(`已更新 AI 配置：${editingItem.name}`);
      } else {
        await createAIConfig(payload);
        toast.success(`已创建 AI 配置：${payload.name}`);
      }
      setDialogOpen(false);
      setEditingItem(null);
      await loadData();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "保存 AI 配置失败");
    } finally {
      setSaving(false);
    }
  }

  async function handleToggleStatus(item: AIConfig) {
    setActionLoadingId(item.id);
    try {
      const nextStatus =
        item.status === Status.Ok
          ? Status.Disabled
          : Status.Ok;
      await updateAIConfigStatus(item.id, nextStatus);
      toast.success(
        `已${nextStatus === Status.Ok ? "启用" : "禁用"}：${item.name}`,
      );
      await loadData();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "更新状态失败");
    } finally {
      setActionLoadingId(null);
    }
  }

  async function handleDelete(item: AIConfig) {
    if (item.status === Status.Ok) {
      toast.error("启用中的 AI 配置不允许删除");
      return;
    }
    setDeletingItem(item);
    setDeleteDialogOpen(true);
  }

  async function handleConfirmDelete() {
    if (!deletingItem) {
      return;
    }
    const item = deletingItem;
    setActionLoadingId(item.id);
    try {
      await deleteAIConfig(item.id);
      toast.success(`已删除 AI 配置：${item.name}`);
      setDeleteDialogOpen(false);
      setDeletingItem(null);
      await loadData();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "删除 AI 配置失败");
    } finally {
      setActionLoadingId(null);
    }
  }

  async function handleDragEnd(event: DragEndEvent) {
    const { active, over } = event;
    if (!over || active.id === over.id || sorting) {
      return;
    }

    const previousResults = result.results;
    const oldIndex = previousResults.findIndex((item) => item.id === active.id);
    const newIndex = previousResults.findIndex((item) => item.id === over.id);
    if (oldIndex < 0 || newIndex < 0) {
      return;
    }

    const nextResults = arrayMove(previousResults, oldIndex, newIndex);
    setResult((current) => ({
      ...current,
      results: nextResults,
    }));
    setSorting(true);

    try {
      await updateAIConfigSort(nextResults.map((item) => item.id));
      toast.success("AI 配置排序已更新");
      await loadData();
    } catch (error) {
      setResult((current) => ({
        ...current,
        results: previousResults,
      }));
      toast.error(error instanceof Error ? error.message : "更新排序失败");
    } finally {
      setSorting(false);
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
              placeholder="按配置名称筛选"
              className="pl-9"
            />
          </div>
          <div className="w-full xl:w-40">
            <OptionCombobox
              value={modelTypeFilterInput}
              options={modelTypeFilterOptions}
              placeholder="全部类型"
              searchPlaceholder="搜索模型类型"
              emptyText="未找到模型类型"
              onChange={setModelTypeFilterInput}
            />
          </div>
          <div className="w-full xl:w-40">
            <OptionCombobox
              value={providerFilterInput}
              options={providerFilterOptions}
              placeholder="全部供应商"
              searchPlaceholder="搜索供应商"
              emptyText="未找到供应商"
              onChange={setProviderFilterInput}
            />
          </div>
          <div className="w-full xl:w-32">
            <OptionCombobox
              value={statusFilterInput}
              options={listStatusOptions}
              placeholder="全部状态"
              searchPlaceholder="搜索状态"
              emptyText="未找到状态"
              onChange={setStatusFilterInput}
            />
          </div>
          <Button variant="outline" onClick={applyFilters} disabled={loading}>
            <SearchIcon />
            查询
          </Button>
          <Button
            variant="outline"
            onClick={() => void loadData()}
            disabled={loading}
          >
            <RefreshCwIcon className={loading ? "animate-spin" : ""} />
            刷新列表
          </Button>
          <Button onClick={openCreateDialog}>
            <PlusIcon />
            新建
          </Button>
        </div>

        <div className="rounded-2xl border bg-card">
          <DndContext
            sensors={sensors}
            collisionDetection={closestCenter}
            onDragEnd={handleDragEnd}
          >
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-14"></TableHead>
                  <TableHead>配置</TableHead>
                  <TableHead>供应商</TableHead>
                  <TableHead>模型</TableHead>
                  <TableHead>接入信息</TableHead>
                  <TableHead>限制</TableHead>
                  <TableHead>状态</TableHead>
                  <TableHead className="text-right">操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {loading ? (
                  <TableRow>
                    <TableCell
                      colSpan={8}
                      className="py-10 text-center text-muted-foreground"
                    >
                      正在加载 AI 配置...
                    </TableCell>
                  </TableRow>
                ) : result.results.length === 0 ? (
                  <TableRow>
                    <TableCell
                      colSpan={8}
                      className="py-10 text-center text-muted-foreground"
                    >
                      暂无 AI 配置数据
                    </TableCell>
                  </TableRow>
                ) : (
                  <SortableContext
                    items={result.results.map((item) => item.id)}
                    strategy={verticalListSortingStrategy}
                  >
                    {result.results.map((item) => (
                      <SortableAIConfigRow
                        key={item.id}
                        item={item}
                        disabled={sorting}
                        actionLoadingId={actionLoadingId}
                        openEditDialog={openEditDialog}
                        handleToggleStatus={handleToggleStatus}
                        handleDelete={handleDelete}
                      />
                    ))}
                  </SortableContext>
                )}
              </TableBody>
            </Table>
          </DndContext>
        </div>

        <ListPagination
          page={result.page.page}
          limit={result.page.limit}
          total={result.page.total}
          onPageChange={handlePageChange}
          onLimitChange={handleLimitChange}
        />
      </div>

      <EditDialog
        open={dialogOpen}
        saving={saving}
        itemId={editingItem?.id ?? null}
        onOpenChange={handleDialogOpenChange}
        onSubmit={handleSubmit}
      />

      <Dialog
        open={deleteDialogOpen}
        onOpenChange={(open) => {
          if (actionLoadingId) {
            return;
          }
          setDeleteDialogOpen(open);
          if (!open) {
            setDeletingItem(null);
          }
        }}
      >
        <DialogContent className="max-w-md" showCloseButton={false}>
          <DialogHeader>
            <DialogTitle>确认删除 AI 配置</DialogTitle>
            <DialogDescription>
              {deletingItem
                ? `确认删除“${deletingItem.name}”吗？此操作不可撤销。`
                : "此操作不可撤销。"}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              disabled={!!actionLoadingId}
              onClick={() => {
                setDeleteDialogOpen(false);
                setDeletingItem(null);
              }}
            >
              取消
            </Button>
            <Button
              type="button"
              variant="destructive"
              disabled={!!actionLoadingId}
              onClick={() => void handleConfirmDelete()}
            >
              {actionLoadingId ? "删除中..." : "确认删除"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
