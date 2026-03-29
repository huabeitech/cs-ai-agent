"use client"

import { useCallback, useEffect, useMemo, useState } from "react"
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
  ChevronRightIcon,
  GripVerticalIcon,
  MoreHorizontalIcon,
  PlusIcon,
  RefreshCwIcon,
  SearchIcon,
  TagIcon,
  Trash2Icon,
} from "lucide-react"
import { toast } from "sonner"

import {
  createTag,
  deleteTag,
  fetchTags,
  fetchTagsAll,
  updateTag,
  updateTagSort,
  updateTagStatus,
  type CreateTagPayload,
  type Tag,
  type TagTree,
} from "@/lib/api/admin"
import { EditDialog } from "./_components/edit"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { ButtonGroup } from "@/components/ui/button-group"
import { Switch } from "@/components/ui/switch"
import { cn } from "@/lib/utils"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Input } from "@/components/ui/input"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"

type TagNode = TagTree & {
  children: TagNode[]
  depth: number
}

function withDepth(nodes: TagTree[], depth = 0): TagNode[] {
  return nodes.map((node) => ({
    ...node,
    depth,
    children: withDepth(node.children, depth + 1),
  }))
}

function collectParentIds(nodes: TagNode[]): Set<number> {
  const ids = new Set<number>()
  const walk = (items: TagNode[]) => {
    items.forEach((item) => {
      if (item.children.length > 0) {
        ids.add(item.id)
        walk(item.children)
      }
    })
  }
  walk(nodes)
  return ids
}

function filterTree(nodes: TagNode[], keyword: string, status?: number): TagNode[] {
  if (!keyword && status === undefined) {
    return nodes
  }

  const result: TagNode[] = []

  function matchesFilter(node: TagNode): boolean {
    const nameMatch = !keyword || node.name.toLowerCase().includes(keyword.toLowerCase())
    const statusMatch = status === undefined || node.status === status
    return nameMatch && statusMatch
  }

  function hasMatchingDescendant(node: TagNode): boolean {
    if (matchesFilter(node)) {
      return true
    }
    return node.children.some(hasMatchingDescendant)
  }

  function filterNode(node: TagNode): TagNode | null {
    if (!hasMatchingDescendant(node)) {
      return null
    }

    const filteredChildren = node.children
      .map(filterNode)
      .filter((child): child is TagNode => child !== null)

    if (matchesFilter(node)) {
      return { ...node, children: filteredChildren }
    }

    if (filteredChildren.length > 0) {
      return { ...node, children: filteredChildren }
    }

    return null
  }

  nodes.forEach((node) => {
    const filtered = filterNode(node)
    if (filtered) {
      result.push(filtered)
    }
  })

  return result
}

const listStatusOptions = [
  { value: "all", label: "全部状态" },
  { value: "0", label: "启用" },
  { value: "1", label: "禁用" },
] as const

type SortableRowProps = {
  item: TagNode & { hasChildren: boolean }
  disabled: boolean
  expanded: boolean
  onToggleExpand: () => void
  onEdit: (item: TagNode) => void
  onToggleStatus: (item: TagNode) => void
  onDelete: (item: TagNode) => void
  actionLoadingId: number | null
}

function SortableRow({
  item,
  disabled,
  expanded,
  onToggleExpand,
  onEdit,
  onToggleStatus,
  onDelete,
  actionLoadingId,
}: SortableRowProps) {
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
        <div
          className="flex items-center gap-2"
          style={{ paddingLeft: item.depth * 24 }}
        >
          {item.hasChildren ? (
            <button
              type="button"
              onClick={onToggleExpand}
              className="flex size-6 items-center justify-center rounded hover:bg-muted"
            >
              <ChevronRightIcon
                className={`size-4 transition-transform ${
                  expanded ? "rotate-90" : ""
                }`}
              />
            </button>
          ) : (
            <span className="w-6" />
          )}
          <div className="flex size-8 items-center justify-center rounded-lg bg-muted text-muted-foreground">
            <TagIcon className="size-4" />
          </div>
          <span className="font-medium">{item.name}</span>
        </div>
      </TableCell>
      <TableCell>
        <div className="flex items-center gap-3">
          <Switch
            checked={item.status === 0}
            disabled={actionLoadingId === item.id}
            onCheckedChange={() => void onToggleStatus(item)}
            aria-label={`${item.name} 状态切换`}
          />
          <Badge variant={item.status === 0 ? "default" : "outline"}>
            {item.status === 0 ? "启用" : "禁用"}
          </Badge>
        </div>
      </TableCell>
      <TableCell>
        <span className="line-clamp-2 text-sm text-muted-foreground">
          {item.remark || "-"}
        </span>
      </TableCell>
      <TableCell className="text-sm text-muted-foreground">
        {item.createdAt}
      </TableCell>
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
              <DropdownMenuItem
                onClick={() => void onDelete(item)}
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
  )
}

export default function DashboardTagsPage() {
  const [keywordInput, setKeywordInput] = useState("")
  const [statusFilterInput, setStatusFilterInput] = useState("all")
  const [keyword, setKeyword] = useState("")
  const [statusFilter, setStatusFilter] = useState("all")
  const [loading, setLoading] = useState(true)
  const [sorting, setSorting] = useState(false)
  const [saving, setSaving] = useState(false)
  const [actionLoadingId, setActionLoadingId] = useState<number | null>(null)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editingItem, setEditingItem] = useState<{ id: number; name: string } | null>(null)
  const [allTags, setAllTags] = useState<Tag[]>([])
  const [tree, setTree] = useState<TagNode[]>([])
  const [expandedIds, setExpandedIds] = useState<Set<number>>(new Set())

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
      const [treeData, listData] = await Promise.all([
        fetchTagsAll(),
        fetchTags({ page: 1, limit: 10000 }),
      ])
      const nextTree = withDepth(treeData)
      setTree(nextTree)
      setAllTags(listData.results)
      setExpandedIds(collectParentIds(nextTree))
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "加载标签失败")
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void loadData()
  }, [loadData])

  function handleStatusFilterChange(value: string | null) {
    setStatusFilterInput(value ?? "all")
  }

  function applyFilters() {
    setKeyword(keywordInput)
    setStatusFilter(statusFilterInput)
  }

  function handleFilterKeyDown(event: React.KeyboardEvent<HTMLInputElement>) {
    if (event.key !== "Enter") {
      return
    }
    event.preventDefault()
    applyFilters()
  }

  function toggleExpanded(id: number) {
    setExpandedIds((prev) => {
      const next = new Set(prev)
      if (next.has(id)) {
        next.delete(id)
      } else {
        next.add(id)
      }
      return next
    })
  }

  function expandAll() {
    setExpandedIds(collectParentIds(tree))
  }

  function collapseAll() {
    setExpandedIds(new Set())
  }

  function openCreateDialog() {
    setEditingItem(null)
    setDialogOpen(true)
  }

  function openEditDialog(item: TagNode) {
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

  async function handleSubmit(payload: CreateTagPayload) {
    if (saving) {
      return
    }

    setSaving(true)
    try {
      if (editingItem) {
        await updateTag({
          id: editingItem.id,
          ...payload,
        })
        toast.success(`已更新标签：${editingItem.name}`)
      } else {
        await createTag(payload)
        toast.success(`已创建标签：${payload.name}`)
      }
      setDialogOpen(false)
      setEditingItem(null)
      await loadData()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "保存标签失败")
    } finally {
      setSaving(false)
    }
  }

  async function handleToggleStatus(item: TagNode) {
    setActionLoadingId(item.id)
    try {
      const nextStatus = item.status === 0 ? 1 : 0
      await updateTagStatus(item.id, nextStatus)
      toast.success(`已${nextStatus === 0 ? "启用" : "禁用"}：${item.name}`)
      await loadData()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "更新状态失败")
    } finally {
      setActionLoadingId(null)
    }
  }

  async function handleDelete(item: TagNode) {
    setActionLoadingId(item.id)
    try {
      await deleteTag(item.id)
      toast.success(`已删除标签：${item.name}`)
      await loadData()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "删除标签失败")
    } finally {
      setActionLoadingId(null)
    }
  }

  const filteredTree = useMemo(
    () =>
      filterTree(
        tree,
        keyword.trim(),
        statusFilter === "all" ? undefined : Number(statusFilter)
      ),
    [keyword, statusFilter, tree]
  )

  type FlatItem = TagNode & { hasChildren: boolean }
  const [flatList, setFlatList] = useState<FlatItem[]>([])

  useEffect(() => {
    const items: FlatItem[] = []
    function collectVisible(nodes: TagNode[]) {
      nodes.forEach((node) => {
        const hasChildren = node.children.length > 0
        items.push({ ...node, hasChildren })
        if (expandedIds.has(node.id)) {
          collectVisible(node.children)
        }
      })
    }
    collectVisible(filteredTree)
    setFlatList(items)
  }, [filteredTree, expandedIds])

  async function handleDragEnd(event: DragEndEvent) {
    const { active, over } = event
    if (!over || active.id === over.id || sorting) {
      return
    }

    const activeItem = flatList.find((item) => item.id === active.id)
    const overItem = flatList.find((item) => item.id === over.id)
    if (!activeItem || !overItem) {
      return
    }

    if (activeItem.parentId !== overItem.parentId) {
      toast.error("只能在同一父标签下拖拽排序")
      return
    }

    const oldIndex = flatList.findIndex((item) => item.id === active.id)
    const newIndex = flatList.findIndex((item) => item.id === over.id)
    if (oldIndex < 0 || newIndex < 0) {
      return
    }

    const previousList = [...flatList]
    const nextList = arrayMove(flatList, oldIndex, newIndex)
    setFlatList(nextList)
    setSorting(true)

    try {
      const siblings = allTags.filter((t) => t.parentId === activeItem.parentId)
      const siblingIds = siblings.map((t) => t.id)
      const movedId = active.id as number
      const targetId = over.id as number
      const movedIndex = siblingIds.indexOf(movedId)
      const targetIndex = siblingIds.indexOf(targetId)
      if (movedIndex < 0 || targetIndex < 0) {
        throw new Error("找不到标签")
      }
      const newSiblingIds = arrayMove(siblingIds, movedIndex, targetIndex)
      await updateTagSort(newSiblingIds)
      toast.success("标签排序已更新")
      await loadData()
    } catch (error) {
      setFlatList(previousList)
      toast.error(error instanceof Error ? error.message : "更新标签排序失败")
    } finally {
      setSorting(false)
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
          <select
            value={statusFilterInput}
            onChange={(e) => handleStatusFilterChange(e.target.value)}
            className="h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-xs outline-none transition-[color,box-shadow] focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50 xl:w-36"
          >
            {listStatusOptions.map((item) => (
              <option key={item.value} value={item.value}>
                {item.label}
              </option>
            ))}
          </select>
          <Button variant="outline" onClick={applyFilters} disabled={loading}>
            <SearchIcon />
            查询
          </Button>
          <Button variant="outline" onClick={() => void loadData()} disabled={loading}>
            <RefreshCwIcon className={loading ? "animate-spin" : ""} />
            刷新
          </Button>
          <Button variant="outline" onClick={expandAll} disabled={loading}>
            展开全部
          </Button>
          <Button variant="outline" onClick={collapseAll} disabled={loading}>
            折叠全部
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
              onDragEnd={handleDragEnd}
            >
              <Table>
                <TableHeader className="bg-muted/40">
                  <TableRow>
                    <TableHead className="w-14" />
                    <TableHead className="min-w-[260px]">标签名称</TableHead>
                    <TableHead>状态</TableHead>
                    <TableHead>备注</TableHead>
                    <TableHead>创建时间</TableHead>
                    <TableHead className="w-[92px] text-right">操作</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  <SortableContext
                    items={flatList.map((item) => item.id)}
                    strategy={verticalListSortingStrategy}
                  >
                    {flatList.map((item) => (
                      <SortableRow
                        key={item.id}
                        item={item}
                        disabled={loading || sorting}
                        expanded={expandedIds.has(item.id)}
                        onToggleExpand={() => toggleExpanded(item.id)}
                        onEdit={openEditDialog}
                        onToggleStatus={handleToggleStatus}
                        onDelete={handleDelete}
                        actionLoadingId={actionLoadingId}
                      />
                    ))}
                  </SortableContext>
                  {!loading && flatList.length === 0 ? (
                    <TableRow>
                      <TableCell colSpan={6} className="py-12 text-center text-muted-foreground">
                        没有匹配的标签
                      </TableCell>
                    </TableRow>
                  ) : null}
                </TableBody>
              </Table>
            </DndContext>
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
    </>
  )
}
