"use client"

import { useCallback, useEffect, useState } from "react"
import type { CSSProperties } from "react"
import {
  closestCenter,
  DndContext,
  type DragEndEvent,
  KeyboardSensor,
  MouseSensor,
  TouchSensor,
  useSensor,
  useSensors,
} from "@dnd-kit/core"
import {
  arrayMove,
  SortableContext,
  sortableKeyboardCoordinates,
  useSortable,
  verticalListSortingStrategy,
} from "@dnd-kit/sortable"
import { CSS } from "@dnd-kit/utilities"
import {
  GripVerticalIcon,
  MoreHorizontalIcon,
  PlusIcon,
  RefreshCwIcon,
  SearchIcon,
  Trash2Icon,
} from "lucide-react"
import { toast } from "sonner"

import {
  createTicketCategory,
  deleteTicketCategory,
  fetchTicketCategories,
  updateTicketCategory,
  updateTicketCategorySort,
  type TicketCategory,
  type CreateTicketCategoryPayload,
  type PageResult,
} from "@/lib/api/admin"
import { EditDialog } from "./_components/edit"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { ButtonGroup } from "@/components/ui/button-group"
import { cn } from "@/lib/utils"
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
  { value: "0", label: "停用" },
] as const

function getStatusLabel(
  value: string,
  options: ReadonlyArray<{ value: string; label: string }>
) {
  return options.find((item) => item.value === value)?.label ?? "请选择状态"
}

type SortableCategoryRowProps = {
  item: TicketCategory
  disabled: boolean
  onEdit: (item: TicketCategory) => void
  onToggleStatus: (item: TicketCategory) => void
  onDelete: (item: TicketCategory) => void
}

function SortableCategoryRow({
  item,
  disabled,
  onEdit,
  onToggleStatus,
  onDelete,
}: SortableCategoryRowProps) {
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
  })

  const style: CSSProperties = {
    transform: CSS.Transform.toString(transform),
    transition,
  }

  return (
    <TableRow
      ref={setNodeRef}
      style={style}
      className={cn(
        isDragging && "relative z-10 bg-muted/60 shadow-sm",
        !disabled && "cursor-move"
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
        <div className="font-medium">{item.name}</div>
      </TableCell>
      <TableCell>
        <Badge variant="outline">{item.code}</Badge>
      </TableCell>
      <TableCell>
        <div className="max-w-xs truncate text-muted-foreground">
          {item.description || "-"}
        </div>
      </TableCell>
      <TableCell>
        <Badge variant={item.status === 1 ? "secondary" : "outline"}>
          {item.status === 1 ? "启用" : "停用"}
        </Badge>
      </TableCell>
      <TableCell>{item.sortNo}</TableCell>
      <TableCell className="text-right">
        <ButtonGroup className="ml-auto">
          <Button
            variant="outline"
            size="sm"
            onClick={() => onEdit(item)}
          >
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
              <DropdownMenuItem onClick={() => onToggleStatus(item)}>
                <RefreshCwIcon />
                {item.status === 1 ? "停用" : "启用"}
              </DropdownMenuItem>
              <DropdownMenuItem
                onClick={() => onDelete(item)}
                className="text-destructive focus:text-destructive"
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
}

export default function DashboardTicketCategoriesPage() {
  const [keywordInput, setKeywordInput] = useState("")
  const [statusFilterInput, setStatusFilterInput] = useState("all")
  const [keyword, setKeyword] = useState("")
  const [statusFilter, setStatusFilter] = useState("all")
  const [page, setPage] = useState(1)
  const [limit, setLimit] = useState(20)
  const [loading, setLoading] = useState(true)
  const [sorting, setSorting] = useState(false)
  const [saving, setSaving] = useState(false)
  const [actionLoadingId, setActionLoadingId] = useState<number | null>(null)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editingItem, setEditingItem] = useState<TicketCategory | null>(null)
  const [result, setResult] = useState<PageResult<TicketCategory>>({
    results: [],
    page: { page: 1, limit: 20, total: 0 },
  })

  const sensors = useSensors(
    useSensor(MouseSensor, {
      activationConstraint: { distance: 8 },
    }),
    useSensor(TouchSensor, {
      activationConstraint: { delay: 150, tolerance: 8 },
    }),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    })
  )

  const loadData = useCallback(async () => {
    setLoading(true)
    try {
      const data = await fetchTicketCategories({
        name: keyword.trim() || undefined,
        status: statusFilter === "all" ? undefined : statusFilter,
        page,
        limit,
      })
      setResult(data)
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "加载工单分类失败")
    } finally {
      setLoading(false)
    }
  }, [keyword, limit, page, statusFilter])

  useEffect(() => {
    void loadData()
  }, [loadData])

  async function handleDragEnd(event: DragEndEvent) {
    const { active, over } = event
    if (!over || active.id === over.id || sorting) {
      return
    }

    const previousResults = result.results
    const oldIndex = previousResults.findIndex((item) => item.id === active.id)
    const newIndex = previousResults.findIndex((item) => item.id === over.id)
    if (oldIndex < 0 || newIndex < 0) {
      return
    }

    const nextResults = arrayMove(previousResults, oldIndex, newIndex)
    setResult((current) => ({
      ...current,
      results: nextResults,
    }))
    setSorting(true)

    try {
      await updateTicketCategorySort(nextResults.map((item) => item.id))
      toast.success("分类排序已更新")
      await loadData()
    } catch (error) {
      setResult((current) => ({
        ...current,
        results: previousResults,
      }))
      toast.error(error instanceof Error ? error.message : "更新分类排序失败")
    } finally {
      setSorting(false)
    }
  }

  function handleStatusFilterChange(value: string | null) {
    setStatusFilterInput(value ?? "all")
  }

  function applyFilters() {
    setKeyword(keywordInput)
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

  function openEditDialog(item: TicketCategory) {
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

  async function handleSubmit(payload: CreateTicketCategoryPayload) {
    if (saving) {
      return
    }

    setSaving(true)
    try {
      if (editingItem) {
        await updateTicketCategory({
          id: editingItem.id,
          ...payload,
        })
        toast.success(`已更新工单分类：${editingItem.name}`)
      } else {
        await createTicketCategory(payload)
        toast.success(`已创建工单分类：${payload.name}`)
      }
      setDialogOpen(false)
      setEditingItem(null)
      await loadData()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "保存工单分类失败")
    } finally {
      setSaving(false)
    }
  }

  async function handleToggleStatus(item: TicketCategory) {
    setActionLoadingId(item.id)
    try {
      const nextStatus = item.status === 1 ? 0 : 1
      await updateTicketCategory({
        id: item.id,
        parentId: item.parentId,
        name: item.name,
        code: item.code,
        description: item.description,
        status: nextStatus,
        remark: item.remark,
      })
      toast.success(`已${nextStatus === 1 ? "启用" : "停用"}：${item.name}`)
      await loadData()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "更新状态失败")
    } finally {
      setActionLoadingId(null)
    }
  }

  async function handleDelete(item: TicketCategory) {
    setActionLoadingId(item.id)
    try {
      await deleteTicketCategory(item.id)
      toast.success(`已删除工单分类：${item.name}`)
      await loadData()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "删除工单分类失败")
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
              placeholder="按名称筛选"
              className="pl-9"
            />
          </div>
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
          <Button variant="outline" onClick={() => void loadData()} disabled={loading || sorting}>
            <RefreshCwIcon className={cn((loading || sorting) && "animate-spin")} />
            刷新列表
          </Button>
          <Button onClick={openCreateDialog}>
            <PlusIcon />
            新建
          </Button>
        </div>
        <div className="space-y-4">
          <div className="overflow-hidden rounded-2xl border bg-background">
            <DndContext
              sensors={sensors}
              collisionDetection={closestCenter}
              onDragEnd={(event) => void handleDragEnd(event)}
            >
              <Table>
                <TableHeader className="bg-muted/40">
                  <TableRow>
                    <TableHead className="w-14"></TableHead>
                    <TableHead>分类名称</TableHead>
                    <TableHead>分类编码</TableHead>
                    <TableHead>描述</TableHead>
                    <TableHead>状态</TableHead>
                    <TableHead>排序</TableHead>
                    <TableHead className="w-[92px] text-right">操作</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  <SortableContext
                    items={result.results.map((item) => item.id)}
                    strategy={verticalListSortingStrategy}
                  >
                    {result.results.map((item) => (
                      <SortableCategoryRow
                        key={item.id}
                        item={item}
                        disabled={loading || sorting}
                        onEdit={openEditDialog}
                        onToggleStatus={handleToggleStatus}
                        onDelete={handleDelete}
                      />
                    ))}
                  </SortableContext>
                  {!loading && result.results.length === 0 ? (
                    <TableRow>
                      <TableCell colSpan={7} className="py-12 text-center text-muted-foreground">
                        没有匹配的工单分类
                      </TableCell>
                    </TableRow>
                  ) : null}
                </TableBody>
              </Table>
            </DndContext>
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
