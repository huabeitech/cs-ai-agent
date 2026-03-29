"use client"

import { type KeyboardEvent, useCallback, useEffect, useState } from "react"
import {
  KeyRoundIcon,
  MoreHorizontalIcon,
  PlusIcon,
  RefreshCwIcon,
  SearchIcon,
  ShieldIcon,
  UserRoundIcon,
} from "lucide-react"
import { toast } from "sonner"

import {
  assignUserRoles,
  createUser,
  fetchRoleListAll,
  fetchUserDetail,
  fetchUsers,
  resetUserPassword,
  updateUser,
  updateUserStatus,
  type AdminRole,
  type AdminUser,
  type CreateAdminUserPayload,
  type PageResult,
  type ResetPasswordResult,
  type UpdateAdminUserPayload,
} from "@/lib/api/admin"
import { Status } from "@/lib/generated/enums"
import { formatDateTime } from "@/lib/utils"
import { AssignRolesDrawer } from "./_components/assign-roles"
import { CreateUserDrawer } from "./_components/create"
import { EditDrawer } from "./_components/edit"
import { InitialPasswordDialog } from "./_components/initial-password-dialog"
import { ResetPasswordDialogs } from "./_components/reset-password"
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
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"

export default function DashboardUsersPage() {
  const [keywordInput, setKeywordInput] = useState("")
  const [keyword, setKeyword] = useState("")
  const [page, setPage] = useState(1)
  const [limit, setLimit] = useState(20)
  const [loading, setLoading] = useState(true)
  const [creatingOpen, setCreatingOpen] = useState(false)
  const [savingCreate, setSavingCreate] = useState(false)
  const [initialPassword, setInitialPassword] = useState<{
    username: string
    password: string
  } | null>(null)
  const [savingEdit, setSavingEdit] = useState(false)
  const [savingPassword, setSavingPassword] = useState(false)
  const [savingRoles, setSavingRoles] = useState(false)
  const [editingUser, setEditingUser] = useState<AdminUser | null>(null)
  const [resettingUser, setResettingUser] = useState<AdminUser | null>(null)
  const [assigningRolesUser, setAssigningRolesUser] = useState<AdminUser | null>(null)
  const [assignRoleOptions, setAssignRoleOptions] = useState<AdminRole[]>([])
  const [assignRoleIds, setAssignRoleIds] = useState<number[]>([])
  const [assignRolesLoading, setAssignRolesLoading] = useState(false)
  const [resetPasswordResult, setResetPasswordResult] =
    useState<ResetPasswordResult | null>(null)
  const [actionLoadingId, setActionLoadingId] = useState<number | null>(null)
  const [result, setResult] = useState<PageResult<AdminUser>>({
    results: [],
    page: { page: 1, limit: 20, total: 0 },
  })

  const loadUsers = useCallback(async () => {
    setLoading(true)
    try {
      const data = await fetchUsers({
        username: keyword.trim() || undefined,
        page,
        limit,
      })
      setResult(data)
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "加载用户失败")
    } finally {
      setLoading(false)
    }
  }, [keyword, limit, page])

  useEffect(() => {
    void loadUsers()
  }, [loadUsers])

  function applyFilters() {
    setKeyword(keywordInput)
    setPage(1)
  }

  function handleFilterKeyDown(event: KeyboardEvent<HTMLInputElement>) {
    if (event.key !== "Enter") {
      return
    }

    event.preventDefault()
    applyFilters()
  }

  function openEditDrawer(user: AdminUser) {
    setEditingUser(user)
  }

  async function openAssignRolesDrawer(user: AdminUser) {
    setActionLoadingId(user.id)
    setAssigningRolesUser(user)
    setAssignRolesLoading(true)
    try {
      const [roles, userDetail] = await Promise.all([
        fetchRoleListAll(),
        fetchUserDetail(user.id),
      ])
      setAssignRoleOptions(roles)
      setAssignRoleIds((userDetail.roles || []).map((role) => role.id))
    } catch (error) {
      setAssigningRolesUser(null)
      toast.error(error instanceof Error ? error.message : "加载角色分配数据失败")
    } finally {
      setAssignRolesLoading(false)
      setActionLoadingId(null)
    }
  }

  function handlePageChange(nextPage: number) {
    if (nextPage < 1 || nextPage === page) {
      return
    }
    setPage(nextPage)
  }

  function handleLimitChange(nextLimit: number) {
    if (nextLimit <= 0 || nextLimit === limit) {
      return
    }
    setLimit(nextLimit)
    setPage(1)
  }

  function handleEditDrawerOpenChange(open: boolean) {
    if (savingEdit) {
      return
    }
    if (!open) {
      setEditingUser(null)
    }
  }

  function handleCreateDrawerOpenChange(open: boolean) {
    if (savingCreate) {
      return
    }
    if (!open) {
      setCreatingOpen(false)
    }
  }

  async function handleCreateUser(payload: CreateAdminUserPayload) {
    if (savingCreate) {
      return
    }

    setSavingCreate(true)
    try {
      const result = await createUser(payload)
      toast.success(`已创建用户 ${result.user.username}`)
      setCreatingOpen(false)
      setInitialPassword({
        username: result.user.username,
        password: result.password,
      })
      await loadUsers()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "创建用户失败")
    } finally {
      setSavingCreate(false)
    }
  }

  function handleAssignRolesOpenChange(open: boolean) {
    if (savingRoles) {
      return
    }
    if (!open) {
      setAssigningRolesUser(null)
      setAssignRoleOptions([])
      setAssignRoleIds([])
    }
  }

  async function handleSaveUser(payload: UpdateAdminUserPayload) {
    if (savingEdit) {
      return
    }

    setSavingEdit(true)
    try {
      await updateUser(payload)
      toast.success(`已更新 ${editingUser?.username || "用户"}`)
      setEditingUser(null)
      await loadUsers()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "更新用户失败")
    } finally {
      setSavingEdit(false)
    }
  }

  async function handleAssignRoles(roleIds: number[]) {
    if (!assigningRolesUser || savingRoles) {
      return
    }

    setSavingRoles(true)
    try {
      await assignUserRoles(assigningRolesUser.id, roleIds)
      toast.success(`已更新 ${assigningRolesUser.username} 的角色`)
      setAssigningRolesUser(null)
      setAssignRoleOptions([])
      setAssignRoleIds([])
      await loadUsers()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "保存角色分配失败")
    } finally {
      setSavingRoles(false)
    }
  }

  function openResetDrawer(user: AdminUser) {
    setResetPasswordResult(null)
    setResettingUser(user)
  }

  function handleResetDrawerOpenChange(open: boolean) {
    if (savingPassword) {
      return
    }
    if (!open) {
      setResetPasswordResult(null)
      setResettingUser(null)
    }
  }

  async function handleResetPassword() {
    if (!resettingUser || savingPassword) {
      return
    }

    setSavingPassword(true)
    try {
      const result = await resetUserPassword(resettingUser.id)
      setResetPasswordResult(result)
      toast.success(`已重置 ${resettingUser.username} 的密码`)
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "重置密码失败")
    } finally {
      setSavingPassword(false)
    }
  }

  async function handleToggleStatus(user: AdminUser) {
    setActionLoadingId(user.id)
    try {
      const nextStatus = user.status === Status.Ok ? Status.Disabled : Status.Ok
      await updateUserStatus(user.id, nextStatus)
      toast.success(`${user.username} 已${nextStatus === Status.Ok ? "启用" : "禁用"}`)
      await loadUsers()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "更新状态失败")
    } finally {
      setActionLoadingId(null)
    }
  }

  return (
    <>
      <div className="flex flex-1 flex-col gap-6 p-4 lg:p-6">
        <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-end">
          <Button onClick={() => setCreatingOpen(true)} disabled={loading}>
            <PlusIcon />
            添加用户
          </Button>
          <div className="relative min-w-72">
            <SearchIcon className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              value={keywordInput}
              onChange={(event) => setKeywordInput(event.target.value)}
              onKeyDown={handleFilterKeyDown}
              placeholder="按用户名筛选"
              className="pl-9"
            />
          </div>
          <Button variant="outline" onClick={applyFilters} disabled={loading}>
            查询
          </Button>
          <Button onClick={() => void loadUsers()} disabled={loading}>
            <RefreshCwIcon className={loading ? "animate-spin" : ""} />
            刷新列表
          </Button>
        </div>
        <div className="space-y-4">
          <div className="overflow-hidden rounded-2xl border bg-background">
            <Table>
              <TableHeader className="bg-muted/40">
                <TableRow>
                  <TableHead>用户</TableHead>
                  <TableHead>角色</TableHead>
                  <TableHead>状态</TableHead>
                  <TableHead>最后登录</TableHead>
                  <TableHead>联系方式</TableHead>
                  <TableHead className="w-[92px] text-right">操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {result.results.map((item) => (
                  <TableRow key={item.id}>
                    <TableCell>
                      <div className="flex items-center gap-3">
                        <div className="flex size-10 items-center justify-center rounded-2xl bg-muted text-muted-foreground">
                          <UserRoundIcon className="size-4" />
                        </div>
                        <div>
                          <div className="font-medium">{item.nickname || item.username}</div>
                          <div className="text-xs text-muted-foreground">{item.username}</div>
                        </div>
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="flex flex-wrap gap-1.5">
                        {(item.roles || []).length > 0 ? (
                          item.roles?.map((role) => (
                            <Badge key={role.id} variant="outline">
                              <ShieldIcon className="size-3" />
                              {role.name}
                            </Badge>
                          ))
                        ) : (
                          <span className="text-sm text-muted-foreground">未分配</span>
                        )}
                      </div>
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
                    <TableCell>
                      <div className="text-sm">{formatDateTime(item.lastLoginAt)}</div>
                      <div className="text-xs text-muted-foreground">
                        {item.lastLoginIp || "-"}
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="text-sm">{item.mobile || "-"}</div>
                      <div className="text-xs text-muted-foreground">
                        {item.email || "-"}
                      </div>
                    </TableCell>
                    <TableCell className="text-right">
                      <ButtonGroup className="ml-auto">
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => openEditDrawer(item)}
                          disabled={actionLoadingId === item.id}
                        >
                          编辑
                        </Button>
                        <DropdownMenu>
                          <DropdownMenuTrigger
                            render={<Button variant="outline" size="icon-sm" />}
                            aria-label={`更多操作 ${item.username}`}
                          >
                            <MoreHorizontalIcon />
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end" className="w-40 min-w-40">
                            <DropdownMenuItem
                              onClick={() => void openAssignRolesDrawer(item)}
                              disabled={actionLoadingId === item.id}
                            >
                              <ShieldIcon />
                              {actionLoadingId === item.id
                                ? "处理中..."
                                : "分配角色"}
                            </DropdownMenuItem>
                            <DropdownMenuItem
                              onClick={() => openResetDrawer(item)}
                              disabled={actionLoadingId === item.id}
                            >
                              <KeyRoundIcon />
                              重置密码
                            </DropdownMenuItem>
                            <DropdownMenuItem
                              onClick={() => handleToggleStatus(item)}
                              disabled={actionLoadingId === item.id}
                            >
                              <ShieldIcon />
                              {actionLoadingId === item.id
                                ? "处理中..."
                                : item.status === Status.Ok
                                  ? "禁用"
                                  : "启用"}
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
                      没有匹配的用户数据
                    </TableCell>
                  </TableRow>
                ) : null}
              </TableBody>
            </Table>
          </div>
          <ListPagination
            page={result.page.page}
            total={result.page.total}
            limit={result.page.limit}
            loading={loading}
            onPageChange={handlePageChange}
            onLimitChange={handleLimitChange}
          />
        </div>
      </div>
      <CreateUserDrawer
        open={creatingOpen}
        saving={savingCreate}
        onOpenChange={handleCreateDrawerOpenChange}
        onSubmit={handleCreateUser}
      />
      <InitialPasswordDialog
        open={!!initialPassword}
        username={initialPassword?.username ?? ""}
        password={initialPassword?.password ?? ""}
        onOpenChange={(open) => {
          if (!open) {
            setInitialPassword(null)
          }
        }}
      />
      <EditDrawer
        open={!!editingUser}
        saving={savingEdit}
        itemId={editingUser?.id ?? null}
        onOpenChange={handleEditDrawerOpenChange}
        onSubmit={handleSaveUser}
      />
      <ResetPasswordDialogs
        open={!!resettingUser}
        saving={savingPassword}
        item={resettingUser}
        password={resetPasswordResult?.password || ""}
        onOpenChange={handleResetDrawerOpenChange}
        onConfirm={handleResetPassword}
      />
      <AssignRolesDrawer
        open={!!assigningRolesUser}
        saving={savingRoles}
        loading={assignRolesLoading}
        item={assigningRolesUser}
        roles={assignRoleOptions}
        selectedRoleIds={assignRoleIds}
        onOpenChange={handleAssignRolesOpenChange}
        onSubmit={handleAssignRoles}
      />
    </>
  )
}
