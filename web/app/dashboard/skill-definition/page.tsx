"use client"

import { useCallback, useEffect, useState } from "react"
import {
  BrainCircuitIcon,
  BugIcon,
  MoreHorizontalIcon,
  PlusIcon,
  RefreshCwIcon,
  RotateCcwIcon,
  SearchIcon,
  Trash2Icon,
} from "lucide-react"
import { toast } from "sonner"

import { ListPagination } from "@/components/list-pagination"
import { OptionCombobox } from "@/components/option-combobox"
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
import { Switch } from "@/components/ui/switch"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import {
  createSkillDefinition,
  deleteSkillDefinition,
  fetchSkillDefinitions,
  restoreSkillDefinition,
  updateSkillDefinition,
  updateSkillDefinitionStatus,
  type CreateSkillDefinitionPayload,
  type PageResult,
  type SkillDefinition,
} from "@/lib/api/admin"
import { Status, StatusLabels } from "@/lib/generated/enums"
import { getEnumLabel, getEnumOptions } from "@/lib/enums"
import { formatDateTime } from "@/lib/utils"
import { EditDialog } from "./_components/edit"
import { DebugDialog } from "./_components/debug-dialog"

const statusFilterOptions = [
  { value: "all", label: "全部状态" },
  ...getEnumOptions(StatusLabels).map((option) => ({
    value: String(option.value),
    label: option.label,
  })),
]

type SkillRowProps = {
  item: SkillDefinition
  actionLoadingId: number | null
  openEditDialog: (item: SkillDefinition) => void
  openDebugDialog: (item: SkillDefinition) => void
  handleToggleStatus: (item: SkillDefinition) => void
  handleDelete: (item: SkillDefinition) => void
  handleRestore: (item: SkillDefinition) => void
}

function SkillRow({
  item,
  actionLoadingId,
  openEditDialog,
  openDebugDialog,
  handleToggleStatus,
  handleDelete,
  handleRestore,
}: SkillRowProps) {
  const isDeleted = item.status === Status.Deleted
  const statusBadgeVariant = isDeleted
    ? "destructive"
    : item.status === Status.Ok
      ? "default"
      : "outline"

  return (
    <TableRow className={isDeleted ? "bg-destructive/5" : undefined}>
      <TableCell>
        <div className="flex items-start gap-3">
          <div className="mt-0.5 flex size-10 items-center justify-center rounded-2xl bg-muted text-muted-foreground">
            <BrainCircuitIcon className="size-4" />
          </div>
          <div className="min-w-0">
            <div className="flex flex-wrap items-center gap-2">
              <div className="font-medium">{item.name}</div>
              <Badge variant="outline">{item.code}</Badge>
              <Badge variant="secondary">白名单 {item.toolWhitelist.length}</Badge>
              <Badge variant="secondary">示例 {item.examples.length}</Badge>
            </div>
            <div className="mt-2 space-y-2">
              <div className="line-clamp-2 text-sm leading-6 text-muted-foreground">
                {item.description || "暂无描述"}
              </div>
            </div>
            {item.toolWhitelist.length > 0 ? (
              <div className="mt-2 flex flex-wrap gap-2">
                {item.toolWhitelist.slice(0, 3).map((toolCode) => (
                  <Badge key={toolCode} variant="outline">
                    {toolCode}
                  </Badge>
                ))}
                {item.toolWhitelist.length > 3 ? (
                  <Badge variant="outline">+{item.toolWhitelist.length - 3}</Badge>
                ) : null}
              </div>
            ) : null}
          </div>
        </div>
      </TableCell>
      <TableCell>
        <div className="flex items-center gap-3">
          <Switch
            checked={item.status === Status.Ok}
            disabled={actionLoadingId === item.id || isDeleted}
            onCheckedChange={() => void handleToggleStatus(item)}
            aria-label={`${item.name} 状态切换`}
          />
          <Badge variant={statusBadgeVariant}>
            {getEnumLabel(StatusLabels, item.status as keyof typeof StatusLabels)}
          </Badge>
        </div>
      </TableCell>
      <TableCell>
        <div className="space-y-1 text-sm">
          <div>{formatDateTime(item.updatedAt)}</div>
          <div className="text-xs text-muted-foreground">
            {item.updateUserName || "-"}
          </div>
        </div>
      </TableCell>
      <TableCell className="text-right">
        <ButtonGroup className="ml-auto">
          <Button variant="outline" size="sm" onClick={() => openDebugDialog(item)}>
            <BugIcon />
            调试
          </Button>
          <Button variant="outline" size="sm" onClick={() => openEditDialog(item)}>
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
              {isDeleted ? (
                <DropdownMenuItem
                  disabled={actionLoadingId === item.id}
                  onClick={() => void handleRestore(item)}
                >
                  <RotateCcwIcon />
                  {actionLoadingId === item.id ? "恢复中..." : "恢复"}
                </DropdownMenuItem>
              ) : (
                <DropdownMenuItem
                  disabled={actionLoadingId === item.id}
                  onClick={() => void handleDelete(item)}
                  className="text-destructive focus:text-destructive"
                >
                  <Trash2Icon />
                  {actionLoadingId === item.id ? "删除中..." : "删除"}
                </DropdownMenuItem>
              )}
            </DropdownMenuContent>
          </DropdownMenu>
        </ButtonGroup>
      </TableCell>
    </TableRow>
  )
}

export default function DashboardSkillsPage() {
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
  const [debugDialogOpen, setDebugDialogOpen] = useState(false)
  const [editingItem, setEditingItem] = useState<SkillDefinition | null>(null)
  const [debuggingItem, setDebuggingItem] = useState<SkillDefinition | null>(null)
  const [result, setResult] = useState<PageResult<SkillDefinition>>({
    results: [],
    page: { page: 1, limit: 20, total: 0 },
  })

  const loadData = useCallback(async () => {
    setLoading(true)
    try {
      const data = await fetchSkillDefinitions({
        name: name.trim() || undefined,
        code: code.trim() || undefined,
        status: statusFilter === "all" ? undefined : Number(statusFilter),
        page,
        limit,
      })
      setResult(data)
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "加载 Skills 失败")
    } finally {
      setLoading(false)
    }
  }, [name, code, statusFilter, page, limit])

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

  function openEditDialog(item: SkillDefinition) {
    setEditingItem(item)
    setDialogOpen(true)
  }

  function openDebugDialog(item: SkillDefinition) {
    setDebuggingItem(item)
    setDebugDialogOpen(true)
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

  function handleDebugDialogOpenChange(open: boolean) {
    if (!open) {
      setDebuggingItem(null)
    }
    setDebugDialogOpen(open)
  }

  async function handleSubmit(payload: CreateSkillDefinitionPayload) {
    if (saving) {
      return
    }

    setSaving(true)
    try {
      if (editingItem) {
        await updateSkillDefinition({
          id: editingItem.id,
          ...payload,
        })
        toast.success(`已更新 Skill：${editingItem.name}`)
      } else {
        await createSkillDefinition(payload)
        toast.success(`已创建 Skill：${payload.name}`)
      }
      setDialogOpen(false)
      setEditingItem(null)
      await loadData()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "保存 Skill 失败")
    } finally {
      setSaving(false)
    }
  }

  async function handleToggleStatus(item: SkillDefinition) {
    if (item.status === Status.Deleted) {
      return
    }

    const nextStatus = item.status === Status.Ok ? Status.Disabled : Status.Ok

    setActionLoadingId(item.id)
    try {
      await updateSkillDefinitionStatus(item.id, nextStatus)
      toast.success(`已${nextStatus === Status.Ok ? "启用" : "停用"}：${item.name}`)
      await loadData()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "更新状态失败")
    } finally {
      setActionLoadingId(null)
    }
  }

  async function handleDelete(item: SkillDefinition) {
    if (item.status === Status.Deleted) {
      return
    }

    setActionLoadingId(item.id)
    try {
      await deleteSkillDefinition(item.id)
      toast.success(`已删除 Skill：${item.name}`)
      await loadData()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "删除 Skill 失败")
    } finally {
      setActionLoadingId(null)
    }
  }

  async function handleRestore(item: SkillDefinition) {
    if (item.status !== Status.Deleted) {
      return
    }

    setActionLoadingId(item.id)
    try {
      await restoreSkillDefinition(item.id)
      toast.success(`已恢复 Skill：${item.name}`)
      await loadData()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "恢复 Skill 失败")
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
              placeholder="按名称筛选"
              className="pl-9"
            />
          </div>
          <Input
            value={codeInput}
            onChange={(event) => setCodeInput(event.target.value)}
            onKeyDown={handleFilterKeyDown}
            placeholder="按编码筛选"
            className="w-full xl:w-56"
          />
          <div className="w-full xl:w-36">
            <OptionCombobox
              value={statusFilterInput}
              options={statusFilterOptions}
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
                  <TableHead>Skill</TableHead>
                  <TableHead>状态</TableHead>
                  <TableHead>最近更新</TableHead>
                  <TableHead className="w-[168px] text-right">操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {!loading && result.results.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={4} className="py-12 text-center text-muted-foreground">
                      没有匹配的 Skill
                    </TableCell>
                  </TableRow>
                ) : null}
                {result.results.map((item) => (
                  <SkillRow
                    key={item.id}
                    item={item}
                    actionLoadingId={actionLoadingId}
                    openEditDialog={openEditDialog}
                    openDebugDialog={openDebugDialog}
                    handleToggleStatus={handleToggleStatus}
                    handleDelete={handleDelete}
                    handleRestore={handleRestore}
                  />
                ))}
              </TableBody>
            </Table>
            <div className="border-t px-4 py-3">
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
        </div>
      </div>

      <EditDialog
        open={dialogOpen}
        saving={saving}
        itemId={editingItem?.id ?? null}
        onOpenChange={handleDialogOpenChange}
        onSubmit={handleSubmit}
      />
      <DebugDialog
        open={debugDialogOpen}
        skillCode={debuggingItem?.code ?? ""}
        skillName={debuggingItem?.name ?? ""}
        onOpenChange={handleDebugDialogOpenChange}
      />
    </>
  )
}
