"use client"

import { useEffect, useMemo, useState } from "react"
import { Controller, Resolver, useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { SearchIcon, ShieldAlertIcon, ShieldCheckIcon } from "lucide-react"
import { z } from "zod/v4"

import {
  fetchRoleListAll,
  type AdminRole,
  type CreateAdminUserPayload,
} from "@/lib/api/admin"
import { Status } from "@/lib/generated/enums"
import { cn } from "@/lib/utils"
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

type CreateUserDrawerProps = {
  open: boolean
  saving: boolean
  onOpenChange: (open: boolean) => void
  onSubmit: (payload: CreateAdminUserPayload) => Promise<void>
}

const createFormSchema = z.object({
  username: z.string().trim().min(1, "用户名不能为空"),
  nickname: z.string().trim(),
  avatar: z
    .string()
    .trim()
    .refine(
      (value) => value.length === 0 || /^https?:\/\/\S+$/i.test(value),
      "头像地址必须是 http 或 https 链接"
    ),
  mobile: z
    .string()
    .trim()
    .refine(
      (value) => value.length === 0 || /^[0-9+\-\s]{6,20}$/.test(value),
      "手机号格式不正确"
    ),
  email: z
    .string()
    .trim()
    .refine(
      (value) =>
        value.length === 0 || /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value),
      "邮箱格式不正确"
    ),
  remark: z.string().trim(),
  roleIds: z.array(z.number().int().positive()),
})

type CreateForm = z.infer<typeof createFormSchema>

const emptyForm: CreateForm = {
  username: "",
  nickname: "",
  avatar: "",
  mobile: "",
  email: "",
  remark: "",
  roleIds: [],
}

const createFormResolver = zodResolver(createFormSchema as never) as Resolver<
  z.input<typeof createFormSchema>,
  undefined,
  z.output<typeof createFormSchema>
>

function toNullableString(value: string) {
  const output = value.trim()
  return output ? output : null
}

function buildPayload(form: CreateForm): CreateAdminUserPayload {
  return {
    username: form.username.trim(),
    nickname: form.nickname.trim(),
    avatar: form.avatar.trim(),
    mobile: toNullableString(form.mobile),
    email: toNullableString(form.email),
    remark: form.remark.trim(),
    roleIds: form.roleIds,
  }
}

export function CreateUserDrawer({
  open,
  saving,
  onOpenChange,
  onSubmit,
}: CreateUserDrawerProps) {
  return (
    <Drawer open={open} onOpenChange={onOpenChange} direction="right">
      {open ? (
        <CreateUserDrawerBody
          key="create-user"
          saving={saving}
          onOpenChange={onOpenChange}
          onSubmit={onSubmit}
        />
      ) : null}
    </Drawer>
  )
}

type CreateUserDrawerBodyProps = {
  saving: boolean
  onOpenChange: (open: boolean) => void
  onSubmit: (payload: CreateAdminUserPayload) => Promise<void>
}

function CreateUserDrawerBody({
  saving,
  onOpenChange,
  onSubmit,
}: CreateUserDrawerBodyProps) {
  const [rolesLoading, setRolesLoading] = useState(true)
  const [roles, setRoles] = useState<AdminRole[]>([])
  const [roleKeyword, setRoleKeyword] = useState("")
  const form = useForm<
    z.input<typeof createFormSchema>,
    undefined,
    z.output<typeof createFormSchema>
  >({
    resolver: createFormResolver,
    defaultValues: emptyForm,
  })
  const {
    control,
    handleSubmit,
    register,
    reset,
    formState: { errors },
  } = form

  useEffect(() => {
    async function loadRoles() {
      setRolesLoading(true)
      try {
        const list = await fetchRoleListAll()
        setRoles(list)
      } catch {
        setRoles([])
      } finally {
        setRolesLoading(false)
      }
    }
    void loadRoles()
  }, [])

  const filteredRoles = useMemo(() => {
    const q = roleKeyword.trim().toLowerCase()
    if (!q) {
      return roles
    }
    return roles.filter((role) =>
      `${role.name} ${role.code}`.toLowerCase().includes(q)
    )
  }, [roleKeyword, roles])

  async function onFormSubmit(values: CreateForm) {
    await onSubmit(buildPayload(values))
    reset(emptyForm)
    setRoleKeyword("")
  }

  return (
    <DrawerContent className="max-w-md">
      <DrawerHeader>
        <DrawerTitle>添加用户</DrawerTitle>
        <DrawerDescription>
          提交后由系统生成初始密码，并仅展示一次，请妥善保存。
        </DrawerDescription>
      </DrawerHeader>
      <form
        className="flex h-full flex-col"
        onSubmit={handleSubmit(onFormSubmit)}
      >
        <div className="space-y-4 overflow-y-auto px-4 pb-4">
          <Field data-invalid={!!errors.username}>
            <FieldLabel htmlFor="create-username">用户名</FieldLabel>
            <FieldContent>
              <Input
                id="create-username"
                placeholder="登录名，必填"
                autoComplete="off"
                aria-invalid={!!errors.username}
                {...register("username")}
              />
              <FieldError errors={[errors.username]} />
            </FieldContent>
          </Field>
          <Field data-invalid={!!errors.nickname}>
            <FieldLabel htmlFor="create-nickname">昵称</FieldLabel>
            <FieldContent>
              <Input
                id="create-nickname"
                placeholder="可选，默认同用户名"
                aria-invalid={!!errors.nickname}
                {...register("nickname")}
              />
              <FieldError errors={[errors.nickname]} />
            </FieldContent>
          </Field>
          <Field data-invalid={!!errors.avatar}>
            <FieldLabel htmlFor="create-avatar">头像地址</FieldLabel>
            <FieldContent>
              <Input
                id="create-avatar"
                placeholder="可选，http(s) 链接"
                aria-invalid={!!errors.avatar}
                {...register("avatar")}
              />
              <FieldError errors={[errors.avatar]} />
            </FieldContent>
          </Field>
          <Field data-invalid={!!errors.mobile}>
            <FieldLabel htmlFor="create-mobile">手机号</FieldLabel>
            <FieldContent>
              <Input
                id="create-mobile"
                placeholder="可选"
                aria-invalid={!!errors.mobile}
                {...register("mobile")}
              />
              <FieldError errors={[errors.mobile]} />
            </FieldContent>
          </Field>
          <Field data-invalid={!!errors.email}>
            <FieldLabel htmlFor="create-email">邮箱</FieldLabel>
            <FieldContent>
              <Input
                id="create-email"
                placeholder="可选"
                aria-invalid={!!errors.email}
                {...register("email")}
              />
              <FieldError errors={[errors.email]} />
            </FieldContent>
          </Field>
          <Field data-invalid={!!errors.remark}>
            <FieldLabel htmlFor="create-remark">备注</FieldLabel>
            <FieldContent>
              <Input
                id="create-remark"
                placeholder="可选"
                aria-invalid={!!errors.remark}
                {...register("remark")}
              />
              <FieldError errors={[errors.remark]} />
            </FieldContent>
          </Field>

          <Field data-invalid={!!errors.roleIds}>
            <FieldLabel>角色（可选）</FieldLabel>
            <FieldContent>
              <div className="relative">
                <SearchIcon className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
                <Input
                  value={roleKeyword}
                  onChange={(event) => setRoleKeyword(event.target.value)}
                  placeholder="搜索角色"
                  className="pl-9"
                  disabled={rolesLoading}
                />
              </div>
              <Controller
                control={control}
                name="roleIds"
                render={({ field }) => {
                  const value = field.value || []
                  const selectedSet = new Set(value)
                  return (
                    <div className="mt-2 max-h-[240px] space-y-1 overflow-y-auto rounded-lg border p-2">
                      {rolesLoading ? (
                        <div className="py-6 text-center text-sm text-muted-foreground">
                          正在加载角色...
                        </div>
                      ) : filteredRoles.length > 0 ? (
                        filteredRoles.map((role) => {
                          const checked = selectedSet.has(role.id)
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
                                    value.filter(
                                      (id: number) => id !== role.id
                                    )
                                  )
                                }}
                              />
                              <span className="flex min-w-0 flex-1 items-center gap-2">
                                {role.status === Status.Ok ? (
                                  <ShieldCheckIcon className="size-3.5 shrink-0 text-muted-foreground" />
                                ) : (
                                  <ShieldAlertIcon className="size-3.5 shrink-0 text-muted-foreground" />
                                )}
                                <span className="truncate">{role.name}</span>
                                {role.status !== Status.Ok ? (
                                  <Badge variant="outline" className="text-xs">
                                    已禁用
                                  </Badge>
                                ) : null}
                              </span>
                            </label>
                          )
                        })
                      ) : (
                        <div className="py-6 text-center text-sm text-muted-foreground">
                          暂无角色
                        </div>
                      )}
                    </div>
                  )
                }}
              />
              <FieldError errors={[errors.roleIds]} />
            </FieldContent>
          </Field>
        </div>
        <DrawerFooter className="border-t">
          <Button type="submit" disabled={saving || rolesLoading}>
            {saving ? "创建中..." : "创建用户"}
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
