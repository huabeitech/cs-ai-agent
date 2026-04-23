"use client"

import { useEffect, useState } from "react"
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
  PlusIcon,
  RefreshCwIcon,
  ShieldCheckIcon,
  ShieldIcon,
} from "lucide-react"
import { toast } from "sonner"

import {
  assignRolePermissions,
  createRole,
  fetchPermissions,
  fetchRoleDetail,
  fetchRoles,
  type AdminPermission,
  type AdminRole,
  type CreateAdminRolePayload,
  type PageResult,
  updateRoleSort,
} from "@/lib/api/admin"
import { cn } from "@/lib/utils"
import { AssignPermissionsDrawer } from "./_components/assign-permissions"
import { CreateRoleDrawer } from "./_components/create"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { Status } from "@/lib/generated/enums"

type SortableRoleRowProps = {
  item: AdminRole
  disabled: boolean
  actionLoading: boolean
  onAssignPermissions: (item: AdminRole) => void
}

function SortableRoleRow({
  item,
  disabled,
  actionLoading,
  onAssignPermissions,
}: SortableRoleRowProps) {
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
        <div className="flex items-center gap-3">
          <div className="flex size-10 items-center justify-center rounded-2xl bg-muted text-muted-foreground">
            <ShieldCheckIcon className="size-4" />
          </div>
          <div className="font-medium">{item.name}</div>
        </div>
      </TableCell>
      <TableCell>
        <Badge variant="outline">
          <ShieldIcon className="size-3" />
          {item.code}
        </Badge>
      </TableCell>
      <TableCell>
        <Badge variant={item.status === Status.Ok ? "secondary" : "outline"}>
          {item.status === Status.Ok ? "启用" : "禁用"}
        </Badge>
        {item.isSystem ? (
          <Badge variant="outline" className="ml-2">
            系统
          </Badge>
        ) : null}
      </TableCell>
      <TableCell>{item.sortNo}</TableCell>
      <TableCell className="text-right">
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={() => onAssignPermissions(item)}
          disabled={disabled || actionLoading}
        >
          {actionLoading ? "处理中..." : "分配权限"}
        </Button>
      </TableCell>
    </TableRow>
  )
}

export default function DashboardRolesPage() {
  const [loading, setLoading] = useState(true)
  const [sorting, setSorting] = useState(false)
  const [creatingOpen, setCreatingOpen] = useState(false)
  const [savingCreate, setSavingCreate] = useState(false)
  const [savingPermissions, setSavingPermissions] = useState(false)
  const [assignPermissionsLoading, setAssignPermissionsLoading] = useState(false)
  const [assigningRole, setAssigningRole] = useState<AdminRole | null>(null)
  const [assignPermissionOptions, setAssignPermissionOptions] = useState<
    AdminPermission[]
  >([])
  const [assignPermissionIds, setAssignPermissionIds] = useState<number[]>([])
  const [actionLoadingId, setActionLoadingId] = useState<number | null>(null)
  const [result, setResult] = useState<PageResult<AdminRole>>({
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

  async function loadRoles() {
    setLoading(true)
    try {
      setResult(await fetchRoles({ limit: 200 }))
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "加载角色失败")
    } finally {
      setLoading(false)
    }
  }

  function handleCreateDrawerOpenChange(open: boolean) {
    if (savingCreate) {
      return
    }
    setCreatingOpen(open)
  }

  async function handleCreateRole(payload: CreateAdminRolePayload) {
    if (savingCreate) {
      return
    }

    setSavingCreate(true)
    try {
      const role = await createRole(payload)
      toast.success(`已创建角色 ${role.name}`)
      setCreatingOpen(false)
      await loadRoles()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "创建角色失败")
    } finally {
      setSavingCreate(false)
    }
  }

  async function openAssignPermissionsDrawer(role: AdminRole) {
    setActionLoadingId(role.id)
    setAssigningRole(role)
    setAssignPermissionsLoading(true)
    try {
      const [permissionsResult, roleDetail] = await Promise.all([
        fetchPermissions({ limit: 500 }),
        fetchRoleDetail(role.id),
      ])
      const permissionCodeSet = new Set(roleDetail.permissions || [])
      setAssignPermissionOptions(permissionsResult.results)
      setAssignPermissionIds(
        permissionsResult.results
          .filter((permission) => permissionCodeSet.has(permission.code))
          .map((permission) => permission.id)
      )
    } catch (error) {
      setAssigningRole(null)
      toast.error(error instanceof Error ? error.message : "加载权限分配数据失败")
    } finally {
      setAssignPermissionsLoading(false)
      setActionLoadingId(null)
    }
  }

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
      await updateRoleSort(nextResults.map((item) => item.id))
      toast.success("角色排序已更新")
      await loadRoles()
    } catch (error) {
      setResult((current) => ({
        ...current,
        results: previousResults,
      }))
      toast.error(error instanceof Error ? error.message : "更新角色排序失败")
    } finally {
      setSorting(false)
    }
  }

  function handleAssignPermissionsOpenChange(open: boolean) {
    if (savingPermissions) {
      return
    }
    if (!open) {
      setAssigningRole(null)
      setAssignPermissionOptions([])
      setAssignPermissionIds([])
    }
  }

  async function handleAssignPermissions(permissionIds: number[]) {
    if (!assigningRole || savingPermissions) {
      return
    }

    setSavingPermissions(true)
    try {
      await assignRolePermissions(assigningRole.id, permissionIds)
      toast.success(`已更新角色 ${assigningRole.name} 的权限`)
      setAssigningRole(null)
      setAssignPermissionOptions([])
      setAssignPermissionIds([])
      await loadRoles()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "保存权限分配失败")
    } finally {
      setSavingPermissions(false)
    }
  }

  useEffect(() => {
    void loadRoles()
  }, [])

  return (
    <div className="flex flex-1 flex-col gap-6 p-4 lg:p-6">
      <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-end">
        <Button
          type="button"
          onClick={() => setCreatingOpen(true)}
          disabled={loading || sorting}
        >
          <PlusIcon className="size-4" />
          添加角色
        </Button>
        <Button onClick={() => void loadRoles()} disabled={loading || sorting}>
          <RefreshCwIcon className={cn((loading || sorting) && "animate-spin")} />
          刷新列表
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
                  <TableHead>角色</TableHead>
                  <TableHead>编码</TableHead>
                  <TableHead>状态</TableHead>
                  <TableHead>排序</TableHead>
                  <TableHead className="text-right">操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <SortableContext
                  items={result.results.map((item) => item.id)}
                  strategy={verticalListSortingStrategy}
                >
                  {result.results.map((item) => (
                    <SortableRoleRow
                      key={item.id}
                      item={item}
                      disabled={loading || sorting}
                      actionLoading={actionLoadingId === item.id}
                      onAssignPermissions={(current) =>
                        void openAssignPermissionsDrawer(current)
                      }
                    />
                  ))}
                </SortableContext>
                {!loading && result.results.length === 0 ? (
                  <TableRow>
                    <TableCell
                      colSpan={6}
                      className="py-12 text-center text-muted-foreground"
                    >
                      暂无角色数据
                    </TableCell>
                  </TableRow>
                ) : null}
              </TableBody>
            </Table>
          </DndContext>
        </div>
        <div className="flex items-center justify-between text-sm text-muted-foreground">
          <span>
            第 {result.page.page} 页 / 每页 {result.page.limit} 条
          </span>
          <span>共 {result.page.total} 条记录</span>
        </div>
      </div>
      <AssignPermissionsDrawer
        open={!!assigningRole}
        saving={savingPermissions}
        loading={assignPermissionsLoading}
        item={assigningRole}
        permissions={assignPermissionOptions}
        selectedPermissionIds={assignPermissionIds}
        onOpenChange={handleAssignPermissionsOpenChange}
        onSubmit={handleAssignPermissions}
      />
      <CreateRoleDrawer
        open={creatingOpen}
        saving={savingCreate}
        onOpenChange={handleCreateDrawerOpenChange}
        onSubmit={handleCreateRole}
      />
    </div>
  )
}
