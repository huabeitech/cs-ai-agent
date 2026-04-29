"use client"

import { useCallback, useEffect, useMemo, useState } from "react"
import {
  CalendarClockIcon,
  CalendarDaysIcon,
  ChevronLeftIcon,
  ChevronRightIcon,
  ListIcon,
  MoreHorizontalIcon,
  PlusIcon,
  RefreshCwIcon,
  SearchIcon,
  Trash2Icon,
} from "lucide-react"
import { toast } from "sonner"

import { ListPagination } from "@/components/list-pagination"
import { Button } from "@/components/ui/button"
import { ButtonGroup } from "@/components/ui/button-group"
import { OptionCombobox } from "@/components/option-combobox"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import {
  createAgentTeamSchedule,
  deleteAgentTeamSchedule,
  fetchAgentTeamScheduleCalendar,
  fetchAgentTeamSchedules,
  fetchAgentTeamsAll,
  updateAgentTeamSchedule,
  type AdminAgentTeam,
  type AdminAgentTeamSchedule,
  type CreateAdminAgentTeamSchedulePayload,
  type PageResult,
  type UpdateAdminAgentTeamSchedulePayload,
} from "@/lib/api/admin"
import { formatDateTime } from "@/lib/utils"
import { ScheduleCalendar } from "./_components/calendar"
import { EditDialog } from "./_components/edit"

type ViewMode = "calendar" | "list"

function startOfDay(date: Date) {
  const ret = new Date(date)
  ret.setHours(0, 0, 0, 0)
  return ret
}

function startOfWeek(date: Date) {
  const ret = startOfDay(date)
  const day = ret.getDay()
  const offset = day === 0 ? -6 : 1 - day
  ret.setDate(ret.getDate() + offset)
  return ret
}

function startOfMonth(date: Date) {
  const ret = startOfDay(date)
  ret.setDate(1)
  return ret
}

function startOfMonthCalendar(date: Date) {
  return startOfWeek(startOfMonth(date))
}

function endOfMonthCalendar(date: Date) {
  const monthEnd = startOfMonth(date)
  monthEnd.setMonth(monthEnd.getMonth() + 1)
  const ret = startOfWeek(monthEnd)
  if (ret.getTime() < monthEnd.getTime()) {
    ret.setDate(ret.getDate() + 7)
  }
  return ret
}

function addDays(date: Date, days: number) {
  const ret = new Date(date)
  ret.setDate(ret.getDate() + days)
  return ret
}

function formatDateTimeValue(date: Date) {
  const month = String(date.getMonth() + 1).padStart(2, "0")
  const day = String(date.getDate()).padStart(2, "0")
  const hour = String(date.getHours()).padStart(2, "0")
  const minute = String(date.getMinutes()).padStart(2, "0")
  const second = String(date.getSeconds()).padStart(2, "0")
  return `${date.getFullYear()}-${month}-${day} ${hour}:${minute}:${second}`
}

function addMonths(date: Date, months: number) {
  const ret = startOfMonth(date)
  ret.setMonth(ret.getMonth() + months)
  return ret
}

function formatMonthTitle(monthStart: Date) {
  return `${monthStart.getFullYear()}年${String(monthStart.getMonth() + 1).padStart(2, "0")}月`
}

export default function DashboardAgentTeamSchedulesPage() {
  const [viewMode, setViewMode] = useState<ViewMode>("calendar")
  const [teamFilterInput, setTeamFilterInput] = useState("all")
  const [teamFilter, setTeamFilter] = useState("all")
  const [monthStart, setMonthStart] = useState(() => startOfMonth(new Date()))
  const [page, setPage] = useState(1)
  const [limit, setLimit] = useState(20)
  const [loading, setLoading] = useState(true)
  const [calendarLoading, setCalendarLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [actionLoadingId, setActionLoadingId] = useState<number | null>(null)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editingItem, setEditingItem] = useState<AdminAgentTeamSchedule | null>(null)
  const [dialogDefaults, setDialogDefaults] = useState<Partial<CreateAdminAgentTeamSchedulePayload> | null>(null)
  const [teams, setTeams] = useState<AdminAgentTeam[]>([])
  const [calendarItems, setCalendarItems] = useState<AdminAgentTeamSchedule[]>([])
  const [result, setResult] = useState<PageResult<AdminAgentTeamSchedule>>({
    results: [],
    page: { page: 1, limit: 20, total: 0 },
  })

  const visibleTeams = useMemo(() => {
    if (teamFilter === "all") {
      return teams
    }
    return teams.filter((team) => String(team.id) === teamFilter)
  }, [teamFilter, teams])

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

  const loadCalendarData = useCallback(async () => {
    setCalendarLoading(true)
    try {
      const data = await fetchAgentTeamScheduleCalendar({
        startAt: formatDateTimeValue(startOfMonthCalendar(monthStart)),
        endAt: formatDateTimeValue(endOfMonthCalendar(monthStart)),
        teamId: teamFilter === "all" ? undefined : teamFilter,
      })
      setCalendarItems(data)
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "加载客服组排班日历失败")
    } finally {
      setCalendarLoading(false)
    }
  }, [monthStart, teamFilter])

  const loadTeams = useCallback(async () => {
    try {
      const data = await fetchAgentTeamsAll()
      setTeams(data)
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "加载客服组选项失败")
    }
  }, [])

  const refreshActiveView = useCallback(async () => {
    await Promise.all([
      loadCalendarData(),
      viewMode === "list" ? loadData() : Promise.resolve(),
    ])
  }, [loadCalendarData, loadData, viewMode])

  useEffect(() => {
    void loadCalendarData()
  }, [loadCalendarData])

  useEffect(() => {
    if (viewMode === "list") {
      void loadData()
    }
  }, [loadData, viewMode])

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

  function openCreateDialog(defaults?: Partial<CreateAdminAgentTeamSchedulePayload>) {
    setEditingItem(null)
    setDialogDefaults(defaults ?? null)
    setDialogOpen(true)
  }

  function openEditDialog(item: AdminAgentTeamSchedule) {
    setDialogDefaults(null)
    setEditingItem(item)
    setDialogOpen(true)
  }

  function handleDialogOpenChange(open: boolean) {
    if (saving) {
      return
    }
    if (!open) {
      setEditingItem(null)
      setDialogDefaults(null)
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
      setDialogDefaults(null)
      await refreshActiveView()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "保存客服组排班失败")
    } finally {
      setSaving(false)
    }
  }

  async function handleDeleteById(id: number) {
    setActionLoadingId(id)
    try {
      await deleteAgentTeamSchedule(id)
      toast.success("已删除客服组排班")
      setDialogOpen(false)
      setEditingItem(null)
      setDialogDefaults(null)
      await refreshActiveView()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "删除客服组排班失败")
    } finally {
      setActionLoadingId(null)
    }
  }

  async function handleDelete(item: AdminAgentTeamSchedule) {
    await handleDeleteById(item.id)
  }

  async function handleCalendarUpdate(payload: UpdateAdminAgentTeamSchedulePayload) {
    setActionLoadingId(payload.id)
    try {
      await updateAgentTeamSchedule(payload)
      toast.success("已更新客服组排班")
      await loadCalendarData()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "更新客服组排班失败")
      await loadCalendarData()
    } finally {
      setActionLoadingId(null)
    }
  }

  return (
    <>
      <div className="flex h-[calc(100vh-var(--header-height))] min-h-0 flex-1 flex-col gap-4 overflow-hidden p-4 lg:p-6">
        <div className="shrink-0 flex flex-col gap-3 xl:flex-row xl:items-center xl:justify-between">
          <div className="flex flex-wrap items-center gap-2">
            <ButtonGroup>
              <Button
                variant={viewMode === "calendar" ? "default" : "outline"}
                size="sm"
                onClick={() => setViewMode("calendar")}
              >
                <CalendarDaysIcon />
                日历
              </Button>
              <Button
                variant={viewMode === "list" ? "default" : "outline"}
                size="sm"
                onClick={() => setViewMode("list")}
              >
                <ListIcon />
                列表
              </Button>
            </ButtonGroup>
            {viewMode === "calendar" ? (
              <ButtonGroup>
                <Button variant="outline" size="icon-sm" onClick={() => setMonthStart(addMonths(monthStart, -1))} aria-label="上一月">
                  <ChevronLeftIcon />
                </Button>
                <Button variant="outline" size="sm" onClick={() => setMonthStart(startOfMonth(new Date()))}>
                  本月
                </Button>
                <Button variant="outline" size="icon-sm" onClick={() => setMonthStart(addMonths(monthStart, 1))} aria-label="下一月">
                  <ChevronRightIcon />
                </Button>
              </ButtonGroup>
            ) : null}
            {viewMode === "calendar" ? (
              <div className="text-sm text-muted-foreground">{formatMonthTitle(monthStart)}</div>
            ) : null}
          </div>

          <div className="flex flex-col gap-2 sm:flex-row sm:items-center xl:justify-end">
            <div className="w-full sm:w-48">
              <OptionCombobox
                value={teamFilterInput}
                options={[
                  { value: "all", label: "全部客服组" },
                  ...teams.map((team) => ({ value: String(team.id), label: team.name })),
                ]}
                placeholder="筛选客服组"
                searchPlaceholder="搜索客服组"
                emptyText="未找到客服组"
                onChange={(value) => setTeamFilterInput(value)}
              />
            </div>
            <Button variant="outline" onClick={applyFilters} disabled={loading || calendarLoading}>
              <SearchIcon />
              查询
            </Button>
            <Button
              variant="outline"
              onClick={() => void refreshActiveView()}
              disabled={loading || calendarLoading}
            >
              <RefreshCwIcon className={loading || calendarLoading ? "animate-spin" : ""} />
              刷新
            </Button>
            <Button onClick={() => openCreateDialog()}>
              <PlusIcon />
              新建
            </Button>
          </div>
        </div>

        {viewMode === "calendar" ? (
          <div className="min-h-0 flex-1 overflow-auto">
            <ScheduleCalendar
              monthStart={monthStart}
              calendarStart={startOfMonthCalendar(monthStart)}
              calendarEnd={endOfMonthCalendar(monthStart)}
              teams={visibleTeams}
              schedules={calendarItems}
              loading={calendarLoading}
              savingId={actionLoadingId}
              onCreate={openCreateDialog}
              onEdit={openEditDialog}
              onMove={handleCalendarUpdate}
              onResize={handleCalendarUpdate}
            />
          </div>
        ) : (
          <div className="min-h-0 flex-1 space-y-4 overflow-auto">
            <div className="min-w-[720px] overflow-hidden rounded-lg border bg-background">
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
                          <div className="mt-0.5 flex size-10 items-center justify-center rounded-md bg-muted text-muted-foreground">
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
        )}
      </div>
      <EditDialog
        open={dialogOpen}
        saving={saving || actionLoadingId === editingItem?.id}
        itemId={editingItem?.id ?? null}
        defaultValues={dialogDefaults}
        onOpenChange={handleDialogOpenChange}
        onSubmit={handleSubmit}
        onDelete={handleDeleteById}
      />
    </>
  )
}
