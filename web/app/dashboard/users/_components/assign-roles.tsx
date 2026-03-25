"use client"

import { useEffect, useMemo, useState } from "react"
import { Controller, Resolver, useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { SearchIcon, ShieldIcon } from "lucide-react"
import { z } from "zod/v4"

import type { AdminRole, AdminUser } from "@/lib/api/admin"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import {
  Drawer,
  DrawerContent,
  DrawerDescription,
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

  const filteredRoles = useMemo(() => {
    const output = keyword.trim().toLowerCase()
    if (!output) {
      return roles
    }
    return roles.filter((role) =>
      `${role.name} ${role.code}`.toLowerCase().includes(output)
    )
  }, [keyword, roles])

  async function onFormSubmit(values: AssignRolesForm) {
    await onSubmit(values.roleIds)
  }

  return (
    <DrawerContent className="max-w-xl">
      <DrawerHeader>
        <DrawerTitle>分配角色</DrawerTitle>
        <DrawerDescription>
          当前用户：{item?.nickname || item?.username || "-"}
        </DrawerDescription>
      </DrawerHeader>
      <form
        className="flex h-full min-h-0 flex-col"
        onSubmit={handleSubmit(onFormSubmit)}
      >
        <div className="space-y-4 px-4 pb-4">
          <Field>
            <FieldLabel>用户账号</FieldLabel>
            <FieldContent>
              <Input value={item?.username || "-"} disabled />
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
              <Controller
                control={control}
                name="roleIds"
                render={({ field }) => {
                  const value = field.value || []
                  const selectedRoles = roles.filter((role) =>
                    value.includes(role.id)
                  )

                  return (
                    <div className="space-y-3">
                      <div className="flex flex-wrap gap-2">
                        {selectedRoles.length > 0 ? (
                          selectedRoles.map((role) => (
                            <Badge key={role.id} variant="secondary">
                              <ShieldIcon className="size-3" />
                              {role.name}
                            </Badge>
                          ))
                        ) : (
                          <span className="text-sm text-muted-foreground">
                            当前未分配角色
                          </span>
                        )}
                      </div>
                      <div className="max-h-80 space-y-2 overflow-y-auto rounded-xl border p-3">
                        {loading ? (
                          <div className="py-8 text-center text-sm text-muted-foreground">
                            正在加载角色列表...
                          </div>
                        ) : filteredRoles.length > 0 ? (
                          filteredRoles.map((role) => {
                            const checked = value.includes(role.id)
                            return (
                              <label
                                key={role.id}
                                className="flex cursor-pointer items-start gap-3 rounded-lg border border-transparent px-3 py-2 hover:bg-muted/50"
                              >
                                <Checkbox
                                  checked={checked}
                                  onCheckedChange={(nextChecked) => {
                                    if (nextChecked) {
                                      field.onChange([...value, role.id])
                                      return
                                    }
                                    field.onChange(
                                      value.filter(
                                        (currentId) => currentId !== role.id
                                      )
                                    )
                                  }}
                                />
                                <div className="min-w-0 flex-1">
                                  <div className="font-medium">{role.name}</div>
                                  <div className="text-sm text-muted-foreground">
                                    {role.code}
                                  </div>
                                </div>
                                <Badge variant={role.status === 1 ? "secondary" : "outline"}>
                                  {role.status === 1 ? "启用" : "禁用"}
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
                    </div>
                  )
                }}
              />
              <FieldError errors={[errors.roleIds]} />
            </FieldContent>
          </Field>
        </div>
        <DrawerFooter className="border-t">
          <Button type="submit" disabled={saving || loading || !item}>
            {saving ? "保存中..." : "保存角色"}
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
