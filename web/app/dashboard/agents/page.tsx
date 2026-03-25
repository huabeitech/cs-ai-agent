"use client";

import {
  MoreHorizontalIcon,
  PanelLeftCloseIcon,
  PanelLeftOpenIcon,
  PlusIcon,
  RefreshCwIcon,
  SearchIcon,
  Trash2Icon,
  UserCogIcon,
} from "lucide-react";
import { useCallback, useEffect, useState, type KeyboardEvent } from "react";
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
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  createAgentProfile,
  deleteAgentProfile,
  fetchAgentProfiles,
  updateAgentProfile,
  type AdminAgentProfile,
  type AdminAgentTeam,
  type CreateAdminAgentProfilePayload,
  type PageResult,
} from "@/lib/api/admin";
import {
  ServiceStatus,
  ServiceStatusLabels,
} from "@/lib/generated/enums";
import { getEnumLabel, getEnumOptions } from "@/lib/enums";
import { formatDateTime } from "@/lib/utils";
import { EditDialog } from "./_components/edit";
import { AgentTeamSidebar } from "./_components/team-sidebar";

const serviceStatusOptions = [
  { value: "all", label: "全部状态" },
  ...getEnumOptions(ServiceStatusLabels),
];

function getStatusLabel(value: number) {
  return getEnumLabel(ServiceStatusLabels, value as ServiceStatus);
}

export default function DashboardAgentsPage() {
  const [selectedTeam, setSelectedTeam] = useState<AdminAgentTeam | null>(null);
  const [teams, setTeams] = useState<AdminAgentTeam[]>([]);
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
  const [agentCodeInput, setAgentCodeInput] = useState("");
  const [displayNameInput, setDisplayNameInput] = useState("");
  const [statusFilterInput, setStatusFilterInput] = useState("all");
  const [agentCode, setAgentCode] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [statusFilter, setStatusFilter] = useState("all");
  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState(20);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [actionLoadingId, setActionLoadingId] = useState<number | null>(null);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingItem, setEditingItem] = useState<AdminAgentProfile | null>(
    null,
  );
  const [result, setResult] = useState<PageResult<AdminAgentProfile>>({
    results: [],
    page: { page: 1, limit: 20, total: 0 },
  });

  const loadData = useCallback(async () => {
    setLoading(true);
    try {
      const data = await fetchAgentProfiles({
        teamId: selectedTeam?.id,
        agentCode: agentCode.trim() || undefined,
        displayName: displayName.trim() || undefined,
        serviceStatus: statusFilter === "all" ? undefined : statusFilter,
        page,
        limit,
      });
      setResult(data);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "加载客服档案失败");
    } finally {
      setLoading(false);
    }
  }, [agentCode, displayName, limit, page, selectedTeam?.id, statusFilter]);

  useEffect(() => {
    void loadData();
  }, [loadData]);

  useEffect(() => {
    setPage(1);
  }, [selectedTeam?.id]);

  useEffect(() => {
    if (teams.length === 0) {
      if (selectedTeam) {
        setSelectedTeam(null);
      }
      return;
    }
    if (!selectedTeam) {
      setSelectedTeam(teams[0]);
      return;
    }
    const matchedTeam = teams.find((item) => item.id === selectedTeam.id);
    if (!matchedTeam) {
      setSelectedTeam(teams[0]);
      return;
    }
    if (matchedTeam !== selectedTeam) {
      setSelectedTeam(matchedTeam);
    }
  }, [selectedTeam, teams]);

  function applyFilters() {
    setAgentCode(agentCodeInput);
    setDisplayName(displayNameInput);
    setStatusFilter(statusFilterInput);
    setPage(1);
  }

  function handleFilterKeyDown(event: KeyboardEvent<HTMLInputElement>) {
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

  function openCreateDialog() {
    setEditingItem(null);
    setDialogOpen(true);
  }

  function openEditDialog(item: AdminAgentProfile) {
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

  async function handleSubmit(payload: CreateAdminAgentProfilePayload) {
    if (saving) {
      return;
    }
    setSaving(true);
    try {
      if (editingItem) {
        await updateAgentProfile({ id: editingItem.id, ...payload });
        toast.success(`已更新客服档案：${editingItem.displayName}`);
      } else {
        await createAgentProfile(payload);
        toast.success(`已创建客服档案：${payload.displayName}`);
      }
      setDialogOpen(false);
      setEditingItem(null);
      await loadData();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "保存客服档案失败");
    } finally {
      setSaving(false);
    }
  }

  async function handleDelete(item: AdminAgentProfile) {
    setActionLoadingId(item.id);
    try {
      await deleteAgentProfile(item.id);
      toast.success(`已删除客服档案：${item.displayName}`);
      await loadData();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "删除客服档案失败");
    } finally {
      setActionLoadingId(null);
    }
  }

  return (
    <>
      <div className="flex h-[calc(100vh-4rem)]">
        <div
          className={`shrink-0 overflow-hidden transition-[width] duration-200 ${
            sidebarCollapsed ? "w-0" : "w-80"
          }`}
        >
          <AgentTeamSidebar
            selectedTeamId={selectedTeam?.id ?? null}
            onSelectTeam={setSelectedTeam}
            onTeamsChange={setTeams}
          />
        </div>
        <div className="relative shrink-0 bg-background">
          <Button
            variant="outline"
            size="icon"
            className="absolute top-4 left-1/2 z-10 size-7 -translate-x-1/2 rounded-full shadow-sm"
            onClick={() => setSidebarCollapsed((value) => !value)}
            aria-label={sidebarCollapsed ? "展开客服组列表" : "折叠客服组列表"}
          >
            {sidebarCollapsed ? (
              <PanelLeftOpenIcon className="size-3.5" />
            ) : (
              <PanelLeftCloseIcon className="size-3.5" />
            )}
          </Button>
        </div>
        <div className="min-w-0 flex-1 p-4 lg:p-6">
          <div className="flex h-full flex-col gap-6">
            <div className="flex items-start justify-between gap-4">
              <div className="min-w-0">
                <div className="text-lg font-semibold">
                  {selectedTeam ? selectedTeam.name : "客服档案"}
                </div>
              </div>
              <div className="flex flex-col gap-2 xl:flex-row xl:items-center xl:justify-end">
                <div className="relative min-w-72">
                  <SearchIcon className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
                  <Input
                    value={displayNameInput}
                    onChange={(event) =>
                      setDisplayNameInput(event.target.value)
                    }
                    onKeyDown={handleFilterKeyDown}
                    placeholder="按展示名筛选"
                    className="pl-9"
                  />
                </div>
                <Input
                  value={agentCodeInput}
                  onChange={(event) => setAgentCodeInput(event.target.value)}
                  onKeyDown={handleFilterKeyDown}
                  placeholder="按客服工号筛选"
                  className="w-full xl:w-48"
                />
                <Select
                  value={statusFilterInput}
                  onValueChange={(value) =>
                    setStatusFilterInput(value ?? "all")
                  }
                >
                  <SelectTrigger className="w-full xl:w-36">
                    <SelectValue>
                      {serviceStatusOptions.find(
                        (item) => item.value === statusFilterInput,
                      )?.label ?? "全部状态"}
                    </SelectValue>
                  </SelectTrigger>
                  <SelectContent>
                    {serviceStatusOptions.map((item) => (
                      <SelectItem key={item.value} value={item.value}>
                        {item.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <Button
                  variant="outline"
                  onClick={applyFilters}
                  disabled={loading}
                >
                  <SearchIcon />
                  查询
                </Button>
                <Button onClick={openCreateDialog}>
                  <PlusIcon />
                  新建
                </Button>
              </div>
            </div>
            <div className="min-h-0 space-y-4">
              <div className="overflow-hidden rounded-2xl border bg-background">
                <Table>
                  <TableHeader className="bg-muted/40">
                    <TableRow>
                      <TableHead>客服</TableHead>
                      <TableHead>服务规则</TableHead>
                      <TableHead>分配策略</TableHead>
                      <TableHead>最近时间</TableHead>
                      <TableHead className="w-[92px] text-right">
                        操作
                      </TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {result.results.map((item) => (
                      <TableRow key={item.id}>
                        <TableCell>
                          <div className="flex items-start gap-3">
                            <div className="mt-0.5 flex size-10 items-center justify-center overflow-hidden rounded-2xl bg-muted">
                              {item.avatar ? (
                                <img
                                  src={item.avatar}
                                  alt={item.displayName}
                                  className="size-full object-cover"
                                />
                              ) : (
                                <UserCogIcon className="size-4 text-muted-foreground" />
                              )}
                            </div>
                            <div className="min-w-0">
                              <div className="font-medium">
                                {item.displayName}
                              </div>
                              <div className="text-xs text-muted-foreground">
                                {item.nickname ||
                                  item.username ||
                                  `用户#${item.userId}`}
                              </div>
                              <div className="mt-1 text-xs text-muted-foreground">
                                工号：{item.agentCode}
                              </div>
                            </div>
                          </div>
                        </TableCell>
                        <TableCell>
                          <Badge variant="outline">
                            {getStatusLabel(item.serviceStatus)}
                          </Badge>
                          <div className="mt-2 text-sm text-muted-foreground">
                            最大并发 {item.maxConcurrentCount} / 优先级{" "}
                            {item.priorityLevel}
                          </div>
                        </TableCell>
                        <TableCell>
                          <div className="flex flex-wrap gap-1.5">
                            <Badge
                              variant={
                                item.autoAssignEnabled ? "secondary" : "outline"
                              }
                            >
                              {item.autoAssignEnabled
                                ? "自动分配"
                                : "不自动分配"}
                            </Badge>
                            <Badge
                              variant={
                                item.receiveOfflineMessage
                                  ? "secondary"
                                  : "outline"
                              }
                            >
                              {item.receiveOfflineMessage
                                ? "离线接收"
                                : "离线不接收"}
                            </Badge>
                          </div>
                        </TableCell>
                        <TableCell>
                          <div className="text-sm">
                            在线：{formatDateTime(item.lastOnlineAt)}
                          </div>
                          <div className="text-sm text-muted-foreground">
                            状态：{formatDateTime(item.lastStatusAt)}
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
                                  <Button variant="outline" size="icon-sm" />
                                }
                                aria-label={`更多操作 ${item.displayName}`}
                              >
                                <MoreHorizontalIcon />
                              </DropdownMenuTrigger>
                              <DropdownMenuContent
                                align="end"
                                className="w-40 min-w-40"
                              >
                                <DropdownMenuItem
                                  onClick={() => void handleDelete(item)}
                                  className="text-destructive focus:text-destructive"
                                >
                                  <Trash2Icon />
                                  {actionLoadingId === item.id
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
                          className="py-12 text-center text-muted-foreground"
                        >
                          {selectedTeam
                            ? "当前客服组下没有匹配的客服档案"
                            : "没有匹配的客服档案"}
                        </TableCell>
                      </TableRow>
                    ) : null}
                  </TableBody>
                </Table>
              </div>
              <ListPagination
                page={result.page.page}
                total={result.page.total}
                limit={limit}
                loading={loading}
                onPageChange={handlePageChange}
                onLimitChange={(nextLimit) => {
                  setLimit(nextLimit);
                  setPage(1);
                }}
              />
            </div>
          </div>
        </div>
      </div>
      <EditDialog
        open={dialogOpen}
        saving={saving}
        itemId={editingItem?.id ?? null}
        defaultTeamId={selectedTeam?.id ?? null}
        onOpenChange={handleDialogOpenChange}
        onSubmit={handleSubmit}
      />
    </>
  );
}
