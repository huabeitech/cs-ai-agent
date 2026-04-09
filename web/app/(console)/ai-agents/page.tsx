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
  BotMessageSquareIcon,
  GripVerticalIcon,
  MoreHorizontalIcon,
  PlusIcon,
  PowerIcon,
  RefreshCwIcon,
  SearchIcon,
  Trash2Icon,
} from "lucide-react";
import {
  useCallback,
  useEffect,
  useState,
  type CSSProperties,
} from "react";
import { toast } from "sonner";

import { ListPagination } from "@/components/list-pagination";
import { OptionCombobox } from "@/components/option-combobox";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
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
  createAIAgent,
  deleteAIAgent,
  fetchAIAgents,
  updateAIAgent,
  updateAIAgentSort,
  updateAIAgentStatus,
  type AIAgent,
  type CreateAIAgentPayload,
  type PageResult,
} from "@/lib/api/admin";
import {
  IMConversationServiceModeLabels,
  Status,
  StatusLabels,
} from "@/lib/generated/enums";
import { getEnumLabel, getEnumOptions } from "@/lib/enums";
import { cn } from "@/lib/utils";
import { EditDialog } from "./_components/edit";
import { ButtonGroup } from "@/components/ui/button-group";

const statusOptions = [
  { value: "all", label: "全部状态" },
  ...getEnumOptions(StatusLabels).map((option) => ({
    value: String(option.value),
    label: option.label,
  })),
];

function getStatusLabel(value: string) {
  return (
    statusOptions.find((item) => item.value === value)?.label ?? "全部状态"
  );
}

type SortableAIAgentRowProps = {
  item: AIAgent;
  disabled: boolean;
  actionLoadingId: number | null;
  openEditDialog: (item: AIAgent) => void;
  handleToggleStatus: (item: AIAgent) => void;
  handleDelete: (item: AIAgent) => void;
};

function SortableAIAgentRow({
  item,
  disabled,
  actionLoadingId,
  openEditDialog,
  handleToggleStatus,
  handleDelete,
}: SortableAIAgentRowProps) {
  const knowledgeIds = item.knowledgeIds ?? [];
  const knowledgeBaseNames = item.knowledgeBaseNames ?? [];
  const skills = item.skills ?? [];
  const directTools = item.directTools ?? [];
  const directToolServerCodes = Array.from(
    new Set(directTools.map((tool) => tool.serverCode).filter(Boolean)),
  );
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
        <div className="flex items-center gap-3">
          <div className="flex size-10 items-center justify-center rounded-2xl bg-muted">
            <BotMessageSquareIcon className="size-4" />
          </div>
          <div>
            <div className="font-medium">{item.name}</div>
          </div>
        </div>
      </TableCell>
      <TableCell>
        {item.aiConfigName || "-"}
      </TableCell>
      <TableCell>
        {getEnumLabel(
          IMConversationServiceModeLabels,
          item.serviceMode as keyof typeof IMConversationServiceModeLabels,
        )}
      </TableCell>
      <TableCell>
        <div className="flex flex-wrap gap-1">
          {knowledgeIds.length === 0 ? (
            <span className="text-sm text-muted-foreground">未配置</span>
          ) : (
            knowledgeBaseNames.map((name, index) => (
              <Badge key={knowledgeIds[index] ?? `${item.id}-${index}`} variant="secondary">
                {name}
              </Badge>
            ))
          )}
        </div>
      </TableCell>
      <TableCell>
        <div className="flex flex-wrap gap-1">
          {skills.length === 0 ? (
            <span className="text-sm text-muted-foreground">仅RAG</span>
          ) : (
            skills.map((skill) => (
              <Badge key={skill.id} variant="outline">
                {skill.name}
              </Badge>
            ))
          )}
        </div>
      </TableCell>
      <TableCell>
        <div className="space-y-2">
          <div className="flex flex-wrap gap-1">
            <Badge variant="secondary">{skills.length} Skills</Badge>
            <Badge variant="secondary">{directTools.length} Tools</Badge>
          </div>
          <div className="flex flex-wrap gap-1">
            {directToolServerCodes.length === 0 ? (
              <span className="text-sm text-muted-foreground">未绑定 MCP Server</span>
            ) : (
              directToolServerCodes.map((serverCode) => (
                <Badge key={serverCode} variant="outline">
                  {serverCode}
                </Badge>
              ))
            )}
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
            variant={item.status === Status.Ok ? "default" : "secondary"}
          >
            {getStatusLabel(String(item.status))}
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
              render={
                <Button variant="outline" size="icon-sm" className="ml-auto" />
              }
              aria-label={`更多操作 ${item.name}`}
            >
              <MoreHorizontalIcon />
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem
                disabled={actionLoadingId === item.id}
                onClick={() => void handleToggleStatus(item)}
              >
                <PowerIcon className="size-4" />
                {item.status === Status.Ok ? "停用" : "启用"}
              </DropdownMenuItem>
              <DropdownMenuItem
                className="text-destructive"
                disabled={actionLoadingId === item.id}
                onClick={() => void handleDelete(item)}
              >
                <Trash2Icon className="size-4" />
                删除
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </ButtonGroup>
      </TableCell>
    </TableRow>
  );
}

export default function DashboardAIAgentsPage() {
  const [nameInput, setNameInput] = useState("");
  const [statusInput, setStatusInput] = useState("all");
  const [name, setName] = useState("");
  const [status, setStatus] = useState("all");
  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState(20);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [actionLoadingId, setActionLoadingId] = useState<number | null>(null);
  const [sorting, setSorting] = useState(false);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingItemId, setEditingItemId] = useState<number | null>(null);
  const [result, setResult] = useState<PageResult<AIAgent>>({
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
      const data = await fetchAIAgents({
        name: name.trim() || undefined,
        status: status === "all" ? undefined : status,
        page,
        limit,
      });
      setResult(data);
    } catch (error) {
      toast.error(
        error instanceof Error ? error.message : "加载 AI Agent 失败",
      );
    } finally {
      setLoading(false);
    }
  }, [limit, name, page, status]);

  useEffect(() => {
    void loadData();
  }, [loadData]);

  function applyFilters() {
    setName(nameInput);
    setStatus(statusInput);
    setPage(1);
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

  function openEditDialog(item: AIAgent) {
    setEditingItemId(item.id);
    setDialogOpen(true);
  }

  async function handleSubmit(payload: CreateAIAgentPayload) {
    if (saving) {
      return;
    }
    setSaving(true);
    try {
      if (editingItemId) {
        await updateAIAgent({ id: editingItemId, ...payload });
        toast.success(`已更新 AI Agent：${payload.name}`);
      } else {
        const created = await createAIAgent(payload);
        toast.success(`已创建 AI Agent：${created.name}`);
      }
      setDialogOpen(false);
      setEditingItemId(null);
      await loadData();
    } catch (error) {
      toast.error(
        error instanceof Error ? error.message : "保存 AI Agent 失败",
      );
    } finally {
      setSaving(false);
    }
  }

  async function handleToggleStatus(item: AIAgent) {
    setActionLoadingId(item.id);
    try {
      const nextStatus =
        item.status === Status.Ok ? Status.Disabled : Status.Ok;
      await updateAIAgentStatus(item.id, nextStatus);
      toast.success(
        `已${nextStatus === Status.Ok ? "启用" : "停用"}：${item.name}`,
      );
      await loadData();
    } catch (error) {
      toast.error(
        error instanceof Error ? error.message : "更新 AI Agent 状态失败",
      );
    } finally {
      setActionLoadingId(null);
    }
  }

  async function handleDelete(item: AIAgent) {
    setActionLoadingId(item.id);
    try {
      await deleteAIAgent(item.id);
      toast.success(`已删除 AI Agent：${item.name}`);
      await loadData();
    } catch (error) {
      toast.error(
        error instanceof Error ? error.message : "删除 AI Agent 失败",
      );
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
      await updateAIAgentSort(nextResults.map((item) => item.id));
      toast.success("AI Agent 排序已更新");
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
          <Input
            value={nameInput}
            onChange={(event) => setNameInput(event.target.value)}
            onKeyDown={handleFilterKeyDown}
            placeholder="按名称筛选"
            className="w-full xl:w-56"
          />
          <div className="w-full xl:w-52">
            <OptionCombobox
              value={statusInput}
              options={statusOptions}
              placeholder="全部状态"
              searchPlaceholder="搜索状态"
              emptyText="未找到状态"
              onChange={setStatusInput}
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
            新建 AI Agent
          </Button>
        </div>

        <div className="rounded-xl border bg-card">
          <DndContext
            sensors={sensors}
            collisionDetection={closestCenter}
            onDragEnd={handleDragEnd}
          >
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-14"></TableHead>
                  <TableHead>Agent</TableHead>
                  <TableHead>AI配置</TableHead>
                  <TableHead>服务模式</TableHead>
                  <TableHead>知识库</TableHead>
                  <TableHead>Skills</TableHead>
                  <TableHead>能力概览</TableHead>
                  <TableHead>状态</TableHead>
                  <TableHead className="w-[88px] text-right">操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {result.results.length === 0 ? (
                  <TableRow>
                    <TableCell
                      colSpan={9}
                      className="py-12 text-center text-muted-foreground"
                    >
                      {loading ? "正在加载 AI Agent..." : "暂无 AI Agent"}
                    </TableCell>
                  </TableRow>
                ) : null}
                <SortableContext
                  items={result.results.map((item) => item.id)}
                  strategy={verticalListSortingStrategy}
                >
                  {result.results.map((item) => (
                    <SortableAIAgentRow
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
              </TableBody>
            </Table>
          </DndContext>
          <div className="border-t px-4 py-3">
            <ListPagination
              page={result.page.page}
              limit={result.page.limit}
              total={result.page.total}
              onPageChange={(nextPage) => setPage(nextPage)}
              onLimitChange={(nextLimit) => {
                setLimit(nextLimit);
                setPage(1);
              }}
            />
          </div>
        </div>
      </div>

      <EditDialog
        open={dialogOpen}
        saving={saving}
        itemId={editingItemId}
        onOpenChange={setDialogOpen}
        onSubmit={handleSubmit}
      />
    </>
  );
}
