"use client";

import { useCallback, useEffect, useState } from "react";
import {
  GlobeIcon,
  MoreHorizontalIcon,
  PencilIcon,
  PlusIcon,
  RefreshCwIcon,
  SearchIcon,
  Trash2Icon,
} from "lucide-react";
import { toast } from "sonner";

import {
  createWidgetSite,
  deleteWidgetSite,
  fetchWidgetSites,
  updateWidgetSite,
  updateWidgetSiteStatus,
  type AdminWidgetSite,
  type CreateAdminWidgetSitePayload,
  type PageResult,
} from "@/lib/api/admin";
import { OptionCombobox } from "@/components/option-combobox";
import { EditDialog } from "./_components/edit";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import { ListPagination } from "@/components/list-pagination";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Status,
  StatusLabels,
} from "@/lib/generated/enums";
import { getEnumLabel, getEnumOptions } from "@/lib/enums";
import { ButtonGroup } from "@/components/ui/button-group";
import { Switch } from "@/components/ui/switch";

const statusOptions = [
  { value: "all", label: "全部状态" },
  ...getEnumOptions(StatusLabels).map((option) => ({
    value: String(option.value),
    label: option.label,
  })),
] as const;

export default function DashboardWidgetSitesPage() {
  const [nameInput, setNameInput] = useState("");
  const [appIdInput, setAppIdInput] = useState("");
  const [statusInput, setStatusInput] = useState("all");
  const [name, setName] = useState("");
  const [appId, setAppId] = useState("");
  const [status, setStatus] = useState("all");
  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState(20);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [actionLoadingId, setActionLoadingId] = useState<number | null>(null);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingItem, setEditingItem] = useState<AdminWidgetSite | null>(null);
  const [result, setResult] = useState<PageResult<AdminWidgetSite>>({
    results: [],
    page: { page: 1, limit: 20, total: 0 },
  });

  const loadData = useCallback(async () => {
    setLoading(true);
    try {
      const data = await fetchWidgetSites({
        name: name.trim() || undefined,
        appId: appId.trim() || undefined,
        status: status === "all" ? undefined : status,
        page,
        limit,
      });
      setResult(data);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "加载嵌入站点失败");
    } finally {
      setLoading(false);
    }
  }, [appId, limit, name, page, status]);

  useEffect(() => {
    void loadData();
  }, [loadData]);

  function applyFilters() {
    setName(nameInput);
    setAppId(appIdInput);
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
    setEditingItem(null);
    setDialogOpen(true);
  }

  function openEditDialog(item: AdminWidgetSite) {
    setEditingItem(item);
    setDialogOpen(true);
  }

  async function handleSubmit(payload: CreateAdminWidgetSitePayload) {
    if (saving) {
      return;
    }
    setSaving(true);
    try {
      if (editingItem) {
        await updateWidgetSite({ id: editingItem.id, ...payload });
        toast.success(`已更新嵌入站点：${payload.name}`);
      } else {
        const created = await createWidgetSite(payload);
        toast.success(`已创建嵌入站点：${created.name}`);
      }
      setDialogOpen(false);
      setEditingItem(null);
      await loadData();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "保存嵌入站点失败");
    } finally {
      setSaving(false);
    }
  }

  async function handleToggleStatus(item: AdminWidgetSite) {
    setActionLoadingId(item.id);
    try {
      const nextStatus =
        item.status === Status.Ok ? Status.Disabled : Status.Ok;
      await updateWidgetSiteStatus(item.id, nextStatus);
      toast.success(
        `已${nextStatus === Status.Ok ? "启用" : "禁用"}：${item.name}`,
      );
      await loadData();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "更新站点状态失败");
    } finally {
      setActionLoadingId(null);
    }
  }

  async function handleDelete(item: AdminWidgetSite) {
    setActionLoadingId(item.id);
    try {
      await deleteWidgetSite(item.id);
      toast.success(`已删除嵌入站点：${item.name}`);
      await loadData();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "删除嵌入站点失败");
    } finally {
      setActionLoadingId(null);
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
            placeholder="按站点名称筛选"
            className="w-full xl:w-56"
          />
          <div className="relative min-w-72">
            <SearchIcon className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              value={appIdInput}
              onChange={(event) => setAppIdInput(event.target.value)}
              onKeyDown={handleFilterKeyDown}
              placeholder="按 appId 筛选"
              className="pl-9"
            />
          </div>
          <div className="w-full xl:w-36">
            <OptionCombobox
              value={statusInput}
              options={[...statusOptions]}
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
            新建站点
          </Button>
        </div>

        <div className="rounded-xl border bg-card">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>站点</TableHead>
                <TableHead>appId</TableHead>
                <TableHead>接待 Agent</TableHead>
                <TableHead>状态</TableHead>
                <TableHead className="w-[88px] text-right">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {result.results.length === 0 ? (
                <TableRow>
                  <TableCell
                    colSpan={5}
                    className="py-12 text-center text-muted-foreground"
                  >
                    {loading ? "正在加载嵌入站点..." : "暂无嵌入站点"}
                  </TableCell>
                </TableRow>
              ) : null}
              {result.results.map((item) => (
                <TableRow key={item.id}>
                  <TableCell>
                    <div className="flex items-center gap-3">
                      <div className="flex size-10 items-center justify-center rounded-2xl bg-muted">
                        <GlobeIcon className="size-4" />
                      </div>
                      <div>
                        <div className="font-medium">{item.name}</div>
                        <div className="text-xs text-muted-foreground">
                          AppId：{item.appId}
                        </div>
                      </div>
                    </div>
                  </TableCell>
                  <TableCell className="font-mono text-xs">
                    {item.appId}
                  </TableCell>
                  <TableCell>{item.aiAgentName || "-"}</TableCell>
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
                          item.status === Status.Ok
                            ? "default"
                            : "outline"
                        }
                      >
                        {getEnumLabel(StatusLabels, item.status as Status)}
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
                            <Button
                              variant="outline"
                              size="icon-sm"
                              className="ml-auto"
                            />
                          }
                          aria-label={`更多操作 ${item.name}`}
                        >
                          <MoreHorizontalIcon />
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
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
              ))}
            </TableBody>
          </Table>
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
        itemId={editingItem?.id ?? null}
        onOpenChange={setDialogOpen}
        onSubmit={handleSubmit}
      />
    </>
  );
}
