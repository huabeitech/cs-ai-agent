"use client";

import {
  MoreHorizontalIcon,
  Pencil,
  PlusIcon,
  RefreshCwIcon,
  SearchIcon,
  Trash2Icon,
  UsersRoundIcon,
} from "lucide-react";
import { useCallback, useEffect, useMemo, useState } from "react";
import { toast } from "sonner";

import { EditDialog } from "./team-edit";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  createAgentTeam,
  deleteAgentTeam,
  fetchAgentTeams,
  updateAgentTeam,
  type AdminAgentTeam,
  type CreateAdminAgentTeamPayload,
} from "@/lib/api/admin";
import { Status, StatusLabels } from "@/lib/generated/enums";
import { getEnumLabel } from "@/lib/enums";
import { cn } from "@/lib/utils";

type AgentTeamSidebarProps = {
  selectedTeamId: number | null;
  onSelectTeam: (team: AdminAgentTeam | null) => void;
  onTeamsChange?: (teams: AdminAgentTeam[]) => void;
};

const statusTabs = [
  { value: "all", label: "全部" },
  { value: String(Status.Ok), label: StatusLabels[Status.Ok] },
  { value: String(Status.Disabled), label: StatusLabels[Status.Disabled] },
] as const;

export function AgentTeamSidebar({
  selectedTeamId,
  onSelectTeam,
  onTeamsChange,
}: AgentTeamSidebarProps) {
  const [keyword, setKeyword] = useState("");
  const [statusFilter, setStatusFilter] =
    useState<(typeof statusTabs)[number]["value"]>("all");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [actionLoadingId, setActionLoadingId] = useState<number | null>(null);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingItem, setEditingItem] = useState<AdminAgentTeam | null>(null);
  const [teams, setTeams] = useState<AdminAgentTeam[]>([]);

  const loadData = useCallback(async () => {
    setLoading(true);
    try {
      const data = await fetchAgentTeams({ page: 1, limit: 200 });
      setTeams(data);
      onTeamsChange?.(data);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "加载客服组失败");
    } finally {
      setLoading(false);
    }
  }, [onTeamsChange]);

  useEffect(() => {
    void loadData();
  }, [loadData]);

  useEffect(() => {
    if (selectedTeamId == null) {
      return;
    }
    const matchedTeam =
      teams.find((item) => item.id === selectedTeamId) ?? null;
    if (matchedTeam) {
      onSelectTeam(matchedTeam);
      return;
    }
    if (!loading && teams.length > 0) {
      onSelectTeam(teams[0]);
    }
  }, [loading, onSelectTeam, selectedTeamId, teams]);

  const filteredTeams = useMemo(() => {
    const output = keyword.trim().toLowerCase();
    return teams.filter((item) => {
      const matchedKeyword =
        output.length === 0 ||
        item.name.toLowerCase().includes(output) ||
        item.description.toLowerCase().includes(output);
      const matchedStatus =
        statusFilter === "all" || String(item.status) === statusFilter;
      return matchedKeyword && matchedStatus;
    });
  }, [keyword, statusFilter, teams]);

  function openCreateDialog() {
    setEditingItem(null);
    setDialogOpen(true);
  }

  function openEditDialog(item: AdminAgentTeam) {
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

  async function handleSubmit(payload: CreateAdminAgentTeamPayload) {
    if (saving) {
      return;
    }
    setSaving(true);
    try {
      if (editingItem) {
        await updateAgentTeam({ id: editingItem.id, ...payload });
        toast.success(`已更新客服组：${editingItem.name}`);
      } else {
        await createAgentTeam(payload);
        toast.success(`已创建客服组：${payload.name}`);
      }
      setDialogOpen(false);
      setEditingItem(null);
      await loadData();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "保存客服组失败");
    } finally {
      setSaving(false);
    }
  }

  async function handleDelete(item: AdminAgentTeam) {
    setActionLoadingId(item.id);
    try {
      await deleteAgentTeam(item.id);
      toast.success(`已删除客服组：${item.name}`);
      await loadData();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "删除客服组失败");
    } finally {
      setActionLoadingId(null);
    }
  }

  return (
    <>
      <div className="flex h-full flex-col border-r bg-muted/10">
        <div className="border-b px-3 py-3">
          <div className="flex items-center justify-between gap-2">
            <div>
              <div className="text-sm font-medium">客服组</div>
            </div>
          </div>
          <div className="relative mt-3">
            <SearchIcon className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              value={keyword}
              onChange={(event) => setKeyword(event.target.value)}
              placeholder="搜索客服组"
              className="pl-9"
            />
          </div>
          <div className="mt-3 flex flex-wrap gap-2">
            {statusTabs.map((item) => (
              <Button
                key={item.value}
                variant={statusFilter === item.value ? "default" : "outline"}
                size="sm"
                onClick={() => setStatusFilter(item.value)}
              >
                {item.label}
              </Button>
            ))}
            <Button
              variant="outline"
              size="sm"
              onClick={() => void loadData()}
              disabled={loading}
            >
              <RefreshCwIcon
                className={cn("size-4", loading && "animate-spin")}
              />
            </Button>

            <Button size="icon-sm" onClick={openCreateDialog}>
              <PlusIcon />
            </Button>
          </div>
        </div>
        <ScrollArea className="min-h-0 flex-1">
          <div className="px-2 py-2">
            {filteredTeams.map((item) => (
              <div
                key={item.id}
                className={cn(
                  "group mt-1 flex items-center gap-2 rounded-lg px-2 py-2 text-sm transition-colors hover:bg-accent",
                  selectedTeamId === item.id &&
                    "bg-accent text-accent-foreground",
                )}
              >
                <button
                  type="button"
                  className="flex min-w-0 flex-1 items-center gap-2 text-left"
                  onClick={() => onSelectTeam(item)}
                >
                  <UsersRoundIcon className="size-4 shrink-0 text-muted-foreground" />
                  <span className="min-w-0 flex-1">
                    <span className="block truncate font-medium">
                      {item.name}
                    </span>
                  </span>
                  <Badge
                    variant={
                      item.status === Status.Ok ? "secondary" : "outline"
                    }
                  >
                    {getEnumLabel(StatusLabels, item.status as Status)}
                  </Badge>
                </button>
                <DropdownMenu>
                  <DropdownMenuTrigger
                    render={
                      <Button
                        variant="ghost"
                        size="icon-sm"
                        className="opacity-0 group-hover:opacity-100"
                      />
                    }
                    aria-label={`更多操作 ${item.name}`}
                  >
                    <MoreHorizontalIcon />
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end" className="w-40 min-w-40">
                    <DropdownMenuItem onClick={() => openEditDialog(item)}>
                      <Pencil />
                      编辑
                    </DropdownMenuItem>
                    <DropdownMenuItem
                      onClick={() => void handleDelete(item)}
                      className="text-destructive focus:text-destructive"
                    >
                      <Trash2Icon />
                      {actionLoadingId === item.id ? "删除中..." : "删除"}
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </div>
            ))}
            {!loading && filteredTeams.length === 0 ? (
              <div className="px-2 py-10 text-center text-sm text-muted-foreground">
                没有匹配的客服组
              </div>
            ) : null}
          </div>
        </ScrollArea>
      </div>
      <EditDialog
        open={dialogOpen}
        saving={saving}
        itemId={editingItem?.id ?? null}
        onOpenChange={handleDialogOpenChange}
        onSubmit={handleSubmit}
      />
    </>
  );
}
