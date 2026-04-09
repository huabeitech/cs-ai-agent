"use client"

import { useEffect, useMemo, useState } from "react"
import { Controller, Resolver, useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { SearchIcon, ShieldAlertIcon, ShieldCheckIcon, ShieldIcon } from "lucide-react"
import { z } from "zod/v4"

import type { AdminRole, AdminUser } from "@/lib/api/admin"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import {
  Drawer,
  DrawerContent,
  DrawerFooter,
  DrawerHeader,
  DrawerTitle,
} from "@/components/ui/drawer"
import {
  Field,
  FieldContent,
  FieldError,
  FieldLabel,
} from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Status } from "@/lib/generated/enums"
import { cn } from "@/lib/utils"

type AssignRolesDrawerProps = {
  open: boolean
  saving: boolean
  loading: boolean
  item: AdminUser | null
  roles: AdminRole[]
  selectedRoleIds: number[]
  onOpenChange: (open: boolean) => void
  onSubmit: (roleIds: number[]) => Promise<void>
}

const assignRolesSchema = z.object({
  roleIds: z.array(z.number().int().positive()),
})

type AssignRolesForm = z.infer<typeof assignRolesSchema>

const assignRolesResolver = zodResolver(assignRolesSchema as never) as Resolver<
  z.input<typeof assignRolesSchema>,
  undefined,
  z.output<typeof assignRolesSchema>
>

function buildForm(selectedRoleIds: number[]): AssignRolesForm {
  return {
    roleIds: selectedRoleIds,
  }
}

export function AssignRolesDrawer({
  open,
  saving,
  loading,
  item,
  roles,
  selectedRoleIds,
  onOpenChange,
  onSubmit,
}: AssignRolesDrawerProps) {
  return (
    <Drawer open={open} onOpenChange={onOpenChange} direction="right">
      {open ? (
        <AssignRolesDrawerBody
          key={item ? `assign-roles-${item.id}` : "assign-roles"}
          saving={saving}
          loading={loading}
          item={item}
          roles={roles}
          selectedRoleIds={selectedRoleIds}
          onOpenChange={onOpenChange}
          onSubmit={onSubmit}
        />
      ) : null}
    </Drawer>
  )
}

type AssignRolesDrawerBodyProps = {
  saving: boolean
  loading: boolean
  item: AdminUser | null
  roles: AdminRole[]
  selectedRoleIds: number[]
  onOpenChange: (open: boolean) => void
  onSubmit: (roleIds: number[]) => Promise<void>
}

function AssignRolesDrawerBody({
  saving,
  loading,
  item,
  roles,
  selectedRoleIds,
  onOpenChange,
  onSubmit,
}: AssignRolesDrawerBodyProps) {
  const [keyword, setKeyword] = useState("")
  const form = useForm<
    z.input<typeof assignRolesSchema>,
    undefined,
    z.output<typeof assignRolesSchema>
  >({
    resolver: assignRolesResolver,
    defaultValues: buildForm(selectedRoleIds),
  })
  const {
    control,
    handleSubmit,
    reset,
    formState: { errors },
  } = form

  useEffect(() => {
    reset(buildForm(selectedRoleIds))
  }, [reset, selectedRoleIds])

  const roleMap = useMemo(
    () => new Map(roles.map((role) => [role.id, role])),
    [roles]
  )

  async function onFormSubmit(values: AssignRolesForm) {
    await onSubmit(values.roleIds)
  }

  return (
    <DrawerContent className="flex min-w-2xl flex-col overflow-hidden">
      <DrawerHeader>
        <DrawerTitle>分配角色</DrawerTitle>
      </DrawerHeader>
      <form
        className="flex min-h-0 flex-1 flex-col overflow-hidden"
        onSubmit={handleSubmit(onFormSubmit)}
      >
        <div className="min-h-0 flex-1 overflow-y-auto">
          <Controller
            control={control}
            name="roleIds"
            render={({ field }) => {
              const value = field.value || []
              const selectedRoleSet = new Set(value)
              const initiallySelectedSet = new Set(selectedRoleIds)
              const selectedRoles = roles.filter((role) => selectedRoleSet.has(role.id))
              const removedRoles = selectedRoleIds
                .map((roleId) => roleMap.get(roleId))
                .filter((role): role is AdminRole => !!role && !selectedRoleSet.has(role.id))
              const addedRoles = value
                .map((roleId) => roleMap.get(roleId))
                .filter((role): role is AdminRole => !!role && !initiallySelectedSet.has(role.id))
              const filteredRoles = roles.filter((role) => {
                const output = keyword.trim().toLowerCase()
                if (!output) {
                  return true
                }
                return `${role.name} ${role.code}`.toLowerCase().includes(output)
              })

              return (
                <div className="space-y-4 px-4 pb-4">
                  <Field>
                    <FieldLabel>当前已分配</FieldLabel>
                    <FieldContent>
                      <div className="rounded-lg border p-3">
                        {selectedRoles.length > 0 ? (
                          <div className="flex flex-wrap gap-2">
                            {selectedRoles.map((role) => (
                              <Badge
                                key={role.id}
                                variant={role.status === Status.Ok ? "secondary" : "outline"}
                                className="gap-1"
                              >
                                {role.status === Status.Ok ? (
                                  <ShieldCheckIcon className="size-3" />
                                ) : (
                                  <ShieldAlertIcon className="size-3" />
                                )}
                                {role.name}
                                {role.status !== Status.Ok ? "（已禁用）" : ""}
                              </Badge>
                            ))}
                          </div>
                        ) : (
                          <div className="text-sm text-muted-foreground">当前未分配角色</div>
                        )}
                      </div>
                    </FieldContent>
                  </Field>

                  <Field data-invalid={!!errors.roleIds}>
                    <FieldLabel>角色列表</FieldLabel>
                    <FieldContent>
                      <div className="relative">
                        <SearchIcon className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
                        <Input
                          value={keyword}
                          onChange={(event) => setKeyword(event.target.value)}
                          placeholder="搜索角色名称或编码"
                          className="pl-9"
                          disabled={loading}
                        />
                      </div>
                      <div className="mt-2 max-h-[360px] space-y-1 overflow-y-auto rounded-lg border p-2">
                        {loading ? (
                          <div className="py-8 text-center text-sm text-muted-foreground">
                            正在加载角色列表...
                          </div>
                        ) : filteredRoles.length > 0 ? (
                          filteredRoles.map((role) => {
                            const checked = selectedRoleSet.has(role.id)
                            const disabled = role.status !== Status.Ok && !checked

                            return (
                              <label
                                key={role.id}
                                className={cn(
                                  "flex items-center gap-2 rounded-md border px-2.5 py-2 text-sm transition-colors",
                                  disabled
                                    ? "cursor-not-allowed border-dashed bg-muted/20 opacity-70"
                                    : "cursor-pointer hover:bg-muted/50",
                                  checked && "border-primary/40 bg-primary/5"
                                )}
                              >
                                <Checkbox
                                  checked={checked}
                                  disabled={disabled}
                                  onCheckedChange={(nextChecked) => {
                                    if (nextChecked) {
                                      field.onChange([...value, role.id])
                                      return
                                    }
                                    field.onChange(
                                      value.filter((currentId) => currentId !== role.id)
                                    )
                                  }}
                                />
                                <div className="min-w-0 flex-1">
                                  <div className="flex items-center gap-2 whitespace-nowrap">
                                    <span className="truncate font-medium">{role.name}</span>
                                    <span className="truncate text-muted-foreground">
                                      {role.code}
                                    </span>
                                  </div>
                                </div>
                                {role.isSystem ? (
                                  <Badge variant="outline" className="shrink-0">
                                    系统
                                  </Badge>
                                ) : null}
                                <Badge
                                  variant={role.status === Status.Ok ? "secondary" : "outline"}
                                  className="shrink-0"
                                >
                                  {role.status === Status.Ok ? "启用" : "禁用"}
                                </Badge>
                              </label>
                            )
                          })
                        ) : (
                          <div className="py-8 text-center text-sm text-muted-foreground">
                            没有匹配的角色
                          </div>
                        )}
                      </div>
                      <FieldError errors={[errors.roleIds]} />
                    </FieldContent>
                  </Field>

                  <Field>
                    <FieldLabel>本次变更</FieldLabel>
                    <FieldContent>
                      <div className="space-y-3 rounded-lg border p-3">
                        <div>
                          <div className="mb-2 text-sm font-medium">新增角色</div>
                          {addedRoles.length > 0 ? (
                            <div className="flex flex-wrap gap-2">
                              {addedRoles.map((role) => (
                                <Badge key={role.id} variant="secondary" className="gap-1">
                                  <ShieldIcon className="size-3" />
                                  {role.name}
                                </Badge>
                              ))}
                            </div>
                          ) : (
                            <div className="text-sm text-muted-foreground">无新增</div>
                          )}
                        </div>
                        <div>
                          <div className="mb-2 text-sm font-medium">移除角色</div>
                          {removedRoles.length > 0 ? (
                            <div className="flex flex-wrap gap-2">
                              {removedRoles.map((role) => (
                                <Badge key={role.id} variant="outline" className="gap-1">
                                  <ShieldIcon className="size-3" />
                                  {role.name}
                                </Badge>
                              ))}
                            </div>
                          ) : (
                            <div className="text-sm text-muted-foreground">无移除</div>
                          )}
                        </div>
                      </div>
                    </FieldContent>
                  </Field>
                </div>
              )
            }}
          />
        </div>
        <DrawerFooter className="border-t">
          <Button type="submit" disabled={saving || loading || !item}>
            {saving ? "保存中..." : "确认分配"}
          </Button>
          <Button
            type="button"
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={saving}
          >
            取消
          </Button>
        </DrawerFooter>
      </form>
    </DrawerContent>
  )
}
