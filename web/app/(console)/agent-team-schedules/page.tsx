"use client"

import { useCallback, useEffect, useState } from "react"
import {
  CalendarClockIcon,
  MoreHorizontalIcon,
  PlusIcon,
  RefreshCwIcon,
  SearchIcon,
  Trash2Icon,
} from "lucide-react"
import { toast } from "sonner"

import {
  createAgentTeamSchedule,
  deleteAgentTeamSchedule,
  fetchAgentTeamSchedules,
  fetchAgentTeams,
  updateAgentTeamSchedule,
  type AdminAgentTeam,
  type AdminAgentTeamSchedule,
  type CreateAdminAgentTeamSchedulePayload,
  type PageResult,
} from "@/lib/api/admin"
import { formatDateTime } from "@/lib/utils"
import { EditDialog } from "./_components/edit"
import { Button } from "@/components/ui/button"
import { ButtonGroup } from "@/components/ui/button-group"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Input } from "@/components/ui/input"
import { ListPagination } from "@/components/list-pagination"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"

export default function DashboardAgentTeamSchedulesPage() {
  const [teamFilterInput, setTeamFilterInput] = useState("all")
  const [teamFilter, setTeamFilter] = useState("all")
  const [page, setPage] = useState(1)
  const [limit, setLimit] = useState(20)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [actionLoadingId, setActionLoadingId] = useState<number | null>(null)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editingItem, setEditingItem] = useState<AdminAgentTeamSchedule | null>(null)
  const [teams, setTeams] = useState<AdminAgentTeam[]>([])
  const [result, setResult] = useState<PageResult<AdminAgentTeamSchedule>>({
    results: [],
    page: { page: 1, limit: 20, total: 0 },
  })

  const loadData = useCallback(async () => {
    setLoading(true)
    try {
      const data = await fetchAgentTeamSchedules({
        teamId: teamFilter === "all" ? undefined : teamFilter,
        page,
        limit,
      })
      setResult(data)
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "加载客服组排班失败")
    } finally {
      setLoading(false)
    }
  }, [limit, page, teamFilter])

  const loadTeams = useCallback(async () => {
    try {
      const data = await fetchAgentTeams()
      setTeams(data)
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "加载客服组选项失败")
    }
  }, [])

  useEffect(() => {
    void loadData()
  }, [loadData])

  useEffect(() => {
    void loadTeams()
  }, [loadTeams])

  function applyFilters() {
    setTeamFilter(teamFilterInput)
    setPage(1)
  }

  function handlePageChange(nextPage: number) {
    if (nextPage < 1 || nextPage === page) {
      return
    }
    setPage(nextPage)
  }

  function openCreateDialog() {
    setEditingItem(null)
    setDialogOpen(true)
  }

  function openEditDialog(item: AdminAgentTeamSchedule) {
    setEditingItem(item)
    setDialogOpen(true)
  }

  function handleDialogOpenChange(open: boolean) {
    if (saving) {
      return
    }
    if (!open) {
      setEditingItem(null)
    }
    setDialogOpen(open)
  }

  async function handleSubmit(payload: CreateAdminAgentTeamSchedulePayload) {
    if (saving) {
      return
    }
    setSaving(true)
    try {
      if (editingItem) {
        await updateAgentTeamSchedule({ id: editingItem.id, ...payload })
        toast.success("已更新客服组排班")
      } else {
        await createAgentTeamSchedule(payload)
        toast.success("已创建客服组排班")
      }
      setDialogOpen(false)
      setEditingItem(null)
      await loadData()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "保存客服组排班失败")
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete(item: AdminAgentTeamSchedule) {
    setActionLoadingId(item.id)
    try {
      await deleteAgentTeamSchedule(item.id)
      toast.success("已删除客服组排班")
      await loadData()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "删除客服组排班失败")
    } finally {
      setActionLoadingId(null)
    }
  }

  return (
    <>
      <div className="flex flex-1 flex-col gap-6 p-4 lg:p-6">
        <div className="flex flex-col gap-2 xl:flex-row xl:items-center xl:justify-end">
          <Select value={teamFilterInput} onValueChange={(value) => setTeamFilterInput(value ?? "all")}>
            <SelectTrigger className="w-full xl:w-48">
              <SelectValue placeholder="筛选客服组" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">全部客服组</SelectItem>
              {teams.map((team) => (
                <SelectItem key={team.id} value={String(team.id)}>
                  {team.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <Button variant="outline" onClick={applyFilters} disabled={loading}>
            <SearchIcon />
            查询
          </Button>
          <Button variant="outline" onClick={() => void loadData()} disabled={loading}>
            <RefreshCwIcon className={loading ? "animate-spin" : ""} />
            刷新列表
          </Button>
          <Button onClick={openCreateDialog}>
            <PlusIcon />
            新建
          </Button>
        </div>
        <div className="space-y-4">
          <div className="overflow-hidden rounded-2xl border bg-background">
            <Table>
              <TableHeader className="bg-muted/40">
                <TableRow>
                  <TableHead>客服组</TableHead>
                  <TableHead>时间范围</TableHead>
                  <TableHead>来源</TableHead>
                  <TableHead className="w-[92px] text-right">操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {result.results.map((item) => (
                  <TableRow key={item.id}>
                    <TableCell>
                      <div className="flex items-start gap-3">
                        <div className="mt-0.5 flex size-10 items-center justify-center rounded-2xl bg-muted text-muted-foreground">
                          <CalendarClockIcon className="size-4" />
                        </div>
                        <div className="min-w-0">
                          <div className="font-medium">{item.teamName || `客服组#${item.teamId}`}</div>
                          <div className="text-xs text-muted-foreground">组ID：{item.teamId}</div>
                        </div>
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="text-sm">{formatDateTime(item.startAt)}</div>
                      <div className="text-sm text-muted-foreground">{formatDateTime(item.endAt)}</div>
                    </TableCell>
                    <TableCell>
                      <div className="text-sm">{item.sourceType}</div>
                    </TableCell>
                    <TableCell className="text-right">
                      <ButtonGroup className="ml-auto">
                        <Button variant="outline" size="sm" onClick={() => openEditDialog(item)}>
                          编辑
                        </Button>
                        <DropdownMenu>
                          <DropdownMenuTrigger
                            render={<Button variant="outline" size="icon-sm" />}
                            aria-label={`更多操作 ${item.startAt}`}
                          >
                            <MoreHorizontalIcon />
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end" className="w-40 min-w-40">
                            <DropdownMenuItem
                              onClick={() => void handleDelete(item)}
                              className="text-destructive focus:text-destructive"
                            >
                              <Trash2Icon />
                              {actionLoadingId === item.id ? "删除中..." : "删除"}
                            </DropdownMenuItem>
                          </DropdownMenuContent>
                        </DropdownMenu>
                      </ButtonGroup>
                    </TableCell>
                  </TableRow>
                ))}
                {!loading && result.results.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={4} className="py-12 text-center text-muted-foreground">
                      没有匹配的客服组排班
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
              setLimit(nextLimit)
              setPage(1)
            }}
          />
        </div>
      </div>
      <EditDialog
        open={dialogOpen}
        saving={saving}
        itemId={editingItem?.id ?? null}
        onOpenChange={handleDialogOpenChange}
        onSubmit={handleSubmit}
      />
    </>
  )
}
