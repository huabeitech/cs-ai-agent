"use client"

import {
  MoreHorizontalIcon,
  PlusIcon,
  RefreshCwIcon,
  SearchIcon,
  Trash2Icon
} from "lucide-react"
import { useCallback, useEffect, useState } from "react"
import { toast } from "sonner"

import { ListPagination } from "@/components/list-pagination"
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
import { type PageResult } from "@/lib/api/admin"
import {
  createCompany,
  deleteCompany,
  fetchCompanies,
  updateCompany,
  updateCompanyStatus,
  type AdminCompany,
  type CreateAdminCompanyPayload,
} from "@/lib/api/company"
import { getEnumOptions } from "@/lib/enums"
import { Status, StatusLabels } from "@/lib/generated/enums"
import { EditDialog } from "./_components/edit"

const listStatusOptions = [
  { value: "all", label: "全部状态" },
  ...getEnumOptions(StatusLabels)
    .filter((item) => Number(item.value) !== Status.Deleted)
    .map((item) => ({
      value: String(item.value),
      label: item.label,
    })),
] as const

function getStatusLabel(
  value: string,
  options: ReadonlyArray<{ value: string; label: string }>
) {
  return options.find((item) => item.value === value)?.label ?? "请选择状态"
}

export default function DashboardCompaniesPage() {
  const [nameInput, setNameInput] = useState("")
  const [codeInput, setCodeInput] = useState("")
  const [statusFilterInput, setStatusFilterInput] = useState("all")
  const [name, setName] = useState("")
  const [code, setCode] = useState("")
  const [statusFilter, setStatusFilter] = useState("all")
  const [page, setPage] = useState(1)
  const [limit, setLimit] = useState(20)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [actionLoadingId, setActionLoadingId] = useState<number | null>(null)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editingItem, setEditingItem] = useState<AdminCompany | null>(null)
  const [result, setResult] = useState<PageResult<AdminCompany>>({
    results: [],
    page: { page: 1, limit: 20, total: 0 },
  })

  const loadData = useCallback(async () => {
    setLoading(true)
    try {
      const data = await fetchCompanies({
        name: name.trim() || undefined,
        code: code.trim() || undefined,
        status: statusFilter === "all" ? undefined : statusFilter,
        page,
        limit,
      })
      setResult(data)
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "加载公司列表失败")
    } finally {
      setLoading(false)
    }
  }, [code, limit, name, page, statusFilter])

  useEffect(() => {
    void loadData()
  }, [loadData])

  function applyFilters() {
    setName(nameInput)
    setCode(codeInput)
    setStatusFilter(statusFilterInput)
    setPage(1)
  }

  function handleFilterKeyDown(event: React.KeyboardEvent<HTMLInputElement>) {
    if (event.key !== "Enter") return
    event.preventDefault()
    applyFilters()
  }

  function handlePageChange(nextPage: number) {
    if (nextPage < 1 || nextPage === page) return
    setPage(nextPage)
  }

  function openCreateDialog() {
    setEditingItem(null)
    setDialogOpen(true)
  }

  function openEditDialog(item: AdminCompany) {
    setEditingItem(item)
    setDialogOpen(true)
  }

  function handleDialogOpenChange(open: boolean) {
    if (saving) return
    if (!open) setEditingItem(null)
    setDialogOpen(open)
  }

  async function handleSubmit(payload: CreateAdminCompanyPayload) {
    if (saving) return
    setSaving(true)
    try {
      if (editingItem) {
        await updateCompany({ id: editingItem.id, ...payload })
        toast.success(`已更新公司：${editingItem.name}`)
      } else {
        await createCompany(payload)
        toast.success(`已创建公司：${payload.name}`)
      }
      setDialogOpen(false)
      setEditingItem(null)
      await loadData()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "保存公司失败")
    } finally {
      setSaving(false)
    }
  }

  async function handleToggleStatus(item: AdminCompany) {
    setActionLoadingId(item.id)
    try {
      const nextStatus = item.status === 0 ? 1 : 0
      await updateCompanyStatus(item.id, nextStatus)
      toast.success(`已${nextStatus === 0 ? "启用" : "禁用"}：${item.name}`)
      await loadData()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "更新状态失败")
    } finally {
      setActionLoadingId(null)
    }
  }

  async function handleDelete(item: AdminCompany) {
    setActionLoadingId(item.id)
    try {
      await deleteCompany(item.id)
      toast.success(`已删除公司：${item.name}`)
      await loadData()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "删除公司失败")
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
              value={nameInput}
              onChange={(event) => setNameInput(event.target.value)}
              onKeyDown={handleFilterKeyDown}
              placeholder="按公司名称筛选"
              className="pl-9"
            />
          </div>
          <Input
            value={codeInput}
            onChange={(event) => setCodeInput(event.target.value)}
            onKeyDown={handleFilterKeyDown}
            placeholder="按公司编码筛选"
            className="w-full xl:w-48"
          />
          <Select value={statusFilterInput} onValueChange={(v) => setStatusFilterInput(v ?? "all")}>
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

        <div className="overflow-hidden rounded-lg border bg-card">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-20">ID</TableHead>
                <TableHead>公司名称</TableHead>
                <TableHead>公司编码</TableHead>
                <TableHead className="w-28">客户数</TableHead>
                <TableHead className="w-24">状态</TableHead>
                <TableHead>备注</TableHead>
                <TableHead className="w-40">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {result.results.length === 0 && !loading ? (
                <TableRow>
                  <TableCell colSpan={7} className="py-10 text-center text-muted-foreground">
                    暂无公司数据
                  </TableCell>
                </TableRow>
              ) : (
                result.results.map((item) => {
                  const actionLoading = actionLoadingId === item.id
                  return (
                    <TableRow key={item.id}>
                      <TableCell>{item.id}</TableCell>
                      <TableCell className="font-medium">{item.name}</TableCell>
                      <TableCell className="text-muted-foreground">{item.code || "-"}</TableCell>
                      <TableCell>{item.customerCount}</TableCell>
                      <TableCell>
                        <Badge
                          variant={
                            item.status === Status.Ok
                              ? "default"
                              : item.status === Status.Deleted
                                ? "outline"
                                : "secondary"
                          }
                        >
                          {StatusLabels[item.status as Status] ?? "未知"}
                        </Badge>
                      </TableCell>
                      <TableCell className="max-w-[320px]">
                        <div className="line-clamp-2 text-muted-foreground">{item.remark || "-"}</div>
                      </TableCell>
                      <TableCell>
                        <ButtonGroup className="w-full justify-end">
                          <Button variant="outline" size="sm" onClick={() => openEditDialog(item)}>
                            编辑
                          </Button>
                          <DropdownMenu>
                            <DropdownMenuTrigger
                              render={
                                <Button variant="outline" size="sm" disabled={actionLoading} />
                              }
                              aria-label={`更多操作 ${item.name}`}
                            >
                              <MoreHorizontalIcon />
                            </DropdownMenuTrigger>
                            <DropdownMenuContent align="end" className="w-40 min-w-40">
                              <DropdownMenuItem
                                disabled={item.status === Status.Deleted}
                                onClick={() => void handleToggleStatus(item)}
                              >
                                {item.status === 0 ? "禁用" : "启用"}
                              </DropdownMenuItem>
                              <DropdownMenuItem
                                variant="destructive"
                                disabled={item.status === Status.Deleted}
                                onClick={() => void handleDelete(item)}
                              >
                                <Trash2Icon />
                                删除
                              </DropdownMenuItem>
                            </DropdownMenuContent>
                          </DropdownMenu>
                        </ButtonGroup>
                      </TableCell>
                    </TableRow>
                  )
                })
              )}
            </TableBody>
          </Table>
        </div>

        <ListPagination
          page={result.page.page}
          total={result.page.total}
          limit={result.page.limit}
          loading={loading}
          onPageChange={handlePageChange}
          onLimitChange={(nextLimit) => {
            setLimit(nextLimit)
            setPage(1)
          }}
        />
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

