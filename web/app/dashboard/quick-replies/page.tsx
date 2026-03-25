"use client"

import { useCallback, useEffect, useState } from "react"
import {
  FileTextIcon,
  MoreHorizontalIcon,
  PlusIcon,
  RefreshCwIcon,
  SearchIcon,
  Trash2Icon,
} from "lucide-react"
import { toast } from "sonner"

import {
  createQuickReply,
  deleteQuickReply,
  fetchQuickReplies,
  updateQuickReply,
  type AdminQuickReply,
  type CreateAdminQuickReplyPayload,
  type PageResult,
} from "@/lib/api/admin"
import { EditDialog } from "./_components/edit"
import { Badge } from "@/components/ui/badge"
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

const listStatusOptions = [
  { value: "all", label: "全部状态" },
  { value: "1", label: "启用" },
  { value: "0", label: "禁用" },
] as const

function getStatusLabel(
  value: string,
  options: ReadonlyArray<{ value: string; label: string }>
) {
  return options.find((item) => item.value === value)?.label ?? "请选择状态"
}

export default function DashboardQuickRepliesPage() {
  const [keywordInput, setKeywordInput] = useState("")
  const [groupNameInput, setGroupNameInput] = useState("")
  const [statusFilterInput, setStatusFilterInput] = useState("all")
  const [keyword, setKeyword] = useState("")
  const [groupName, setGroupName] = useState("")
  const [statusFilter, setStatusFilter] = useState("all")
  const [page, setPage] = useState(1)
  const [limit, setLimit] = useState(20)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [actionLoadingId, setActionLoadingId] = useState<number | null>(null)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editingItem, setEditingItem] = useState<AdminQuickReply | null>(null)
  const [result, setResult] = useState<PageResult<AdminQuickReply>>({
    results: [],
    page: { page: 1, limit: 20, total: 0 },
  })

  const loadData = useCallback(async () => {
    setLoading(true)
    try {
      const data = await fetchQuickReplies({
        title: keyword.trim() || undefined,
        groupName: groupName.trim() || undefined,
        status: statusFilter === "all" ? undefined : statusFilter,
        page,
        limit,
      })
      setResult(data)
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "加载快捷回复失败")
    } finally {
      setLoading(false)
    }
  }, [groupName, keyword, limit, page, statusFilter])

  useEffect(() => {
    void loadData()
  }, [loadData])

  function handleStatusFilterChange(value: string | null) {
    setStatusFilterInput(value ?? "all")
  }

  function applyFilters() {
    setKeyword(keywordInput)
    setGroupName(groupNameInput)
    setStatusFilter(statusFilterInput)
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

  function openCreateDialog() {
    setEditingItem(null)
    setDialogOpen(true)
  }

  function openEditDialog(item: AdminQuickReply) {
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

  async function handleSubmit(payload: CreateAdminQuickReplyPayload) {
    if (saving) {
      return
    }

    setSaving(true)
    try {
      if (editingItem) {
        await updateQuickReply({
          id: editingItem.id,
          ...payload,
        })
        toast.success(`已更新快捷回复：${editingItem.title}`)
      } else {
        await createQuickReply(payload)
        toast.success(`已创建快捷回复：${payload.title}`)
      }
      setDialogOpen(false)
      setEditingItem(null)
      await loadData()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "保存快捷回复失败")
    } finally {
      setSaving(false)
    }
  }

  async function handleToggleStatus(item: AdminQuickReply) {
    setActionLoadingId(item.id)
    try {
      const nextStatus = item.status === 1 ? 0 : 1
      await updateQuickReply({
        id: item.id,
        groupName: item.groupName,
        title: item.title,
        content: item.content,
        sortNo: item.sortNo,
        status: nextStatus,
      })
      toast.success(`已${nextStatus === 1 ? "启用" : "禁用"}：${item.title}`)
      await loadData()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "更新状态失败")
    } finally {
      setActionLoadingId(null)
    }
  }

  async function handleDelete(item: AdminQuickReply) {
    setActionLoadingId(item.id)
    try {
      await deleteQuickReply(item.id)
      toast.success(`已删除快捷回复：${item.title}`)
      await loadData()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "删除快捷回复失败")
    } finally {
      setActionLoadingId(null)
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
              placeholder="按标题筛选"
              className="pl-9"
            />
          </div>
          <Input
            value={groupNameInput}
            onChange={(event) => setGroupNameInput(event.target.value)}
            onKeyDown={handleFilterKeyDown}
            placeholder="按分组筛选"
            className="w-full xl:w-48"
          />
          <Select
            value={statusFilterInput}
            onValueChange={handleStatusFilterChange}
          >
            <SelectTrigger className="w-full xl:w-36">
              <SelectValue>{getStatusLabel(statusFilterInput, listStatusOptions)}</SelectValue>
            </SelectTrigger>
            <SelectContent>
              {listStatusOptions.map((item) => (
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
                  <TableHead>快捷回复</TableHead>
                  <TableHead>分组</TableHead>
                  <TableHead>状态</TableHead>
                  <TableHead>排序</TableHead>
                  <TableHead>创建人</TableHead>
                  <TableHead className="w-[92px] text-right">操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {result.results.map((item) => (
                  <TableRow key={item.id}>
                    <TableCell>
                      <div className="flex items-start gap-3">
                        <div className="mt-0.5 flex size-10 items-center justify-center rounded-2xl bg-muted text-muted-foreground">
                          <FileTextIcon className="size-4" />
                        </div>
                        <div className="min-w-0">
                          <div className="font-medium">{item.title}</div>
                          <div className="mt-1 line-clamp-2 text-sm text-muted-foreground">
                            {item.content}
                          </div>
                        </div>
                      </div>
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline">{item.groupName}</Badge>
                    </TableCell>
                    <TableCell>
                      <Badge variant={item.status === 1 ? "secondary" : "outline"}>
                        {item.status === 1 ? "启用" : "禁用"}
                      </Badge>
                    </TableCell>
                    <TableCell>{item.sortNo}</TableCell>
                    <TableCell>{item.createdBy || "-"}</TableCell>
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
                            aria-label={`更多操作 ${item.title}`}
                          >
                            <MoreHorizontalIcon />
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end" className="w-40 min-w-40">
                            <DropdownMenuItem onClick={() => void handleToggleStatus(item)}>
                              <RefreshCwIcon />
                              {actionLoadingId === item.id
                                ? "处理中..."
                                : item.status === 1
                                  ? "禁用"
                                  : "启用"}
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
                      </ButtonGroup>
                    </TableCell>
                  </TableRow>
                ))}
                {!loading && result.results.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={6} className="py-12 text-center text-muted-foreground">
                      没有匹配的快捷回复
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
