"use client"

import { useCallback, useEffect, useState } from "react"
import {
  MoreHorizontalIcon,
  PlusIcon,
  RefreshCwIcon,
  SearchIcon,
  Trash2Icon,
} from "lucide-react"
import { toast } from "sonner"

import {
  closeTicket,
  createTicket,
  deleteTicket,
  fetchTickets,
  reopenTicket,
  type Ticket,
  type CreateTicketPayload,
  type PageResult,
} from "@/lib/api/admin"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { ButtonGroup } from "@/components/ui/button-group"
import { EditDialog } from "./_components/edit"
import { TicketDetailDrawer } from "./_components/detail"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Input } from "@/components/ui/input"
import { ListPagination } from "@/components/list-pagination"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"

const ticketStatusOptions = [
  { value: "all", label: "全部状态" },
  { value: "1", label: "待处理" },
  { value: "2", label: "处理中" },
  { value: "3", label: "待确认" },
  { value: "4", label: "已解决" },
  { value: "5", label: "已关闭" },
  { value: "6", label: "已取消" },
] as const

const ticketPriorityOptions = [
  { value: "all", label: "全部优先级" },
  { value: "0", label: "普通" },
  { value: "1", label: "低" },
  { value: "2", label: "中" },
  { value: "3", label: "高" },
  { value: "4", label: "紧急" },
] as const

function getStatusLabel(value: string, options: ReadonlyArray<{ value: string; label: string }>) {
  return options.find((item) => item.value === value)?.label ?? "请选择状态"
}

function getPriorityLabel(value: number) {
  const labels = ["普通", "低", "中", "高", "紧急"]
  return labels[value] ?? "普通"
}

function getStatusLabelText(value: number) {
  const labels: Record<number, string> = {
    1: "待处理",
    2: "处理中",
    3: "待确认",
    4: "已解决",
    5: "已关闭",
    6: "已取消",
  }
  return labels[value] ?? "未知"
}

function formatTime(timestamp: number) {
  if (!timestamp) return "-"
  return new Date(timestamp * 1000).toLocaleString("zh-CN", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  })
}

export default function DashboardTicketsPage() {
  const [keywordInput, setKeywordInput] = useState("")
  const [statusFilterInput, setStatusFilterInput] = useState("all")
  const [priorityFilterInput, setPriorityFilterInput] = useState("all")
  const [keyword, setKeyword] = useState("")
  const [statusFilter, setStatusFilter] = useState("all")
  const [priorityFilter, setPriorityFilter] = useState("all")
  const [page, setPage] = useState(1)
  const [limit, setLimit] = useState(20)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [actionLoadingId, setActionLoadingId] = useState<number | null>(null)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [detailDrawerOpen, setDetailDrawerOpen] = useState(false)
  const [detailTicketId, setDetailTicketId] = useState<number | null>(null)
  const [result, setResult] = useState<PageResult<Ticket>>({
    results: [],
    page: { page: 1, limit: 20, total: 0 },
  })

  const loadData = useCallback(async () => {
    setLoading(true)
    try {
      const data = await fetchTickets({
        title: keyword.trim() || undefined,
        status: statusFilter === "all" ? undefined : statusFilter,
        priority: priorityFilter === "all" ? undefined : priorityFilter,
        page,
        limit,
      })
      setResult(data)
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "加载工单失败")
    } finally {
      setLoading(false)
    }
  }, [keyword, limit, page, priorityFilter, statusFilter])

  useEffect(() => {
    void loadData()
  }, [loadData])

  function handleStatusFilterChange(value: string | null) {
    setStatusFilterInput(value ?? "all")
  }

  function handlePriorityFilterChange(value: string | null) {
    setPriorityFilterInput(value ?? "all")
  }

  function applyFilters() {
    setKeyword(keywordInput)
    setStatusFilter(statusFilterInput)
    setPriorityFilter(priorityFilterInput)
    setPage(1)
  }

  function handleFilterKeyDown(event: React.KeyboardEvent<HTMLInputElement>) {
    if (event.key !== "Enter") {
      return
    }
    event.preventDefault()
    applyFilters()
  }

  function handlePageChange(nextPage: number) {
    if (nextPage < 1 || nextPage === page) {
      return
    }
    setPage(nextPage)
  }

  async function handleSubmit(payload: CreateTicketPayload) {
    if (saving) {
      return
    }
    setSaving(true)
    try {
      await createTicket(payload)
      toast.success(`已创建工单：${payload.title}`)
      setDialogOpen(false)
      await loadData()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "创建工单失败")
    } finally {
      setSaving(false)
    }
  }

  async function handleCloseTicket(item: Ticket) {
    setActionLoadingId(item.id)
    try {
      await closeTicket(item.id)
      toast.success(`已关闭工单：${item.title}`)
      await loadData()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "关闭工单失败")
    } finally {
      setActionLoadingId(null)
    }
  }

  async function handleReopenTicket(item: Ticket) {
    setActionLoadingId(item.id)
    try {
      await reopenTicket(item.id)
      toast.success(`已重开工单：${item.title}`)
      await loadData()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "重开工单失败")
    } finally {
      setActionLoadingId(null)
    }
  }

  async function handleDeleteTicket(item: Ticket) {
    setActionLoadingId(item.id)
    try {
      await deleteTicket(item.id)
      toast.success(`已删除工单：${item.title}`)
      await loadData()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "删除工单失败")
    } finally {
      setActionLoadingId(null)
    }
  }

  return (
    <>
      <div className="flex flex-1 flex-col gap-6 p-4 lg:p-6">
        <div className="flex flex-col gap-2 xl:flex-row xl:items-center xl:justify-between">
          <h1 className="text-2xl font-semibold">工单管理</h1>
          <div className="flex flex-col sm:flex-row gap-2">
            <Button onClick={() => setDialogOpen(true)} disabled={saving}>
              <PlusIcon className="mr-2 h-4 w-4" />
              新建工单
            </Button>
            <div className="relative min-w-72">
              <SearchIcon className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                value={keywordInput}
                onChange={(event) => setKeywordInput(event.target.value)}
                onKeyDown={handleFilterKeyDown}
                placeholder="按标题筛选"
                className="pl-9"
              />
            </div>
            <Select
              value={statusFilterInput}
              onValueChange={handleStatusFilterChange}
            >
              <SelectTrigger className="w-full xl:w-36">
                <SelectValue>{getStatusLabel(statusFilterInput, ticketStatusOptions)}</SelectValue>
              </SelectTrigger>
              <SelectContent>
                {ticketStatusOptions.map((item) => (
                  <SelectItem key={item.value} value={item.value}>
                    {item.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Select
              value={priorityFilterInput}
              onValueChange={handlePriorityFilterChange}
            >
              <SelectTrigger className="w-full xl:w-36">
                <SelectValue>{getStatusLabel(priorityFilterInput, ticketPriorityOptions)}</SelectValue>
              </SelectTrigger>
              <SelectContent>
                {ticketPriorityOptions.map((item) => (
                  <SelectItem key={item.value} value={item.value}>
                    {item.label}
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
          </div>
        </div>
        <div className="space-y-4">
          <div className="overflow-hidden rounded-2xl border bg-background">
            <Table>
              <TableHeader className="bg-muted/40">
                <TableRow>
                  <TableHead>工单编号</TableHead>
                  <TableHead>标题</TableHead>
                  <TableHead>客户</TableHead>
                  <TableHead>优先级</TableHead>
                  <TableHead>状态</TableHead>
                  <TableHead>回复数</TableHead>
                  <TableHead>创建时间</TableHead>
                  <TableHead className="w-[92px] text-right">操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {result.results.map((item) => (
                  <TableRow key={item.id}>
                    <TableCell>
                      <Badge variant="outline">{item.ticketNo}</Badge>
                    </TableCell>
                    <TableCell>
                      <div 
                        className="font-medium cursor-pointer hover:underline"
                        onClick={() => {
                          setDetailTicketId(item.id)
                          setDetailDrawerOpen(true)
                        }}
                      >
                        {item.title}
                      </div>
                      <div className="mt-1 line-clamp-1 text-sm text-muted-foreground">
                        {item.content}
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="text-sm">{item.externalUserName || "-"}</div>
                      {item.externalUserEmail && (
                        <div className="text-xs text-muted-foreground">{item.externalUserEmail}</div>
                      )}
                    </TableCell>
                    <TableCell>
                      <Badge variant={item.priority >= 3 ? "destructive" : "secondary"}>
                        {getPriorityLabel(item.priority)}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <Badge
                        variant={
                          item.status === 4
                            ? "default"
                            : item.status === 5
                              ? "outline"
                              : "secondary"
                        }
                      >
                        {getStatusLabelText(item.status)}
                      </Badge>
                    </TableCell>
                    <TableCell>{item.replyCount}</TableCell>
                    <TableCell className="text-muted-foreground">
                      {formatTime(item.createdAt)}
                    </TableCell>
                    <TableCell className="text-right">
                      <ButtonGroup className="ml-auto">
                        <DropdownMenu>
                          <DropdownMenuTrigger
                            render={<Button variant="outline" size="icon-sm" />}
                            aria-label={`更多操作 ${item.title}`}
                          >
                            <MoreHorizontalIcon />
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end" className="w-40 min-w-40">
                            {item.status === 5 && (
                              <DropdownMenuItem onClick={() => void handleReopenTicket(item)}>
                                <RefreshCwIcon />
                                {actionLoadingId === item.id ? "处理中..." : "重开"}
                              </DropdownMenuItem>
                            )}
                            {item.status !== 5 && (
                              <DropdownMenuItem onClick={() => void handleCloseTicket(item)}>
                                <RefreshCwIcon />
                                {actionLoadingId === item.id ? "处理中..." : "关闭"}
                              </DropdownMenuItem>
                            )}
                            <DropdownMenuItem
                              onClick={() => void handleDeleteTicket(item)}
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
                    <TableCell colSpan={8} className="py-12 text-center text-muted-foreground">
                      没有匹配的工单
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
        onOpenChange={setDialogOpen}
        onSubmit={handleSubmit}
      />
      <TicketDetailDrawer
        open={detailDrawerOpen}
        ticketId={detailTicketId}
        onOpenChange={setDetailDrawerOpen}
        onRefresh={loadData}
      />
    </>
  )
}
