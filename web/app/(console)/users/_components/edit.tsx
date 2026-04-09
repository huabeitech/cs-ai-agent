"use client"

import { useEffect, useState } from "react"
import { zodResolver } from "@hookform/resolvers/zod"
import { Resolver, useForm } from "react-hook-form"
import { z } from "zod/v4"

import {
  type AdminUser,
  type UpdateAdminUserPayload,
  fetchUserDetail,
} from "@/lib/api/admin"
import { Button } from "@/components/ui/button"
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

type UserEditDrawerProps = {
  open: boolean
  saving: boolean
  itemId: number | null
  onOpenChange: (open: boolean) => void
  onSubmit: (payload: UpdateAdminUserPayload) => Promise<void>
}

const emptyForm: EditForm = {
  nickname: "",
  avatar: "",
  mobile: "",
  email: "",
}

const editFormSchema = z.object({
  nickname: z.string().trim().min(1, "昵称不能为空"),
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
})

type EditForm = z.infer<typeof editFormSchema>
const editFormResolver = zodResolver(editFormSchema as never) as Resolver<
  z.input<typeof editFormSchema>,
  undefined,
  z.output<typeof editFormSchema>
>

function toNullableString(value: string) {
  const output = value.trim()
  return output ? output : null
}

function buildForm(item: AdminUser | null): EditForm {
  if (!item) {
    return emptyForm
  }

  return {
    nickname: item.nickname || "",
    avatar: item.avatar || "",
    mobile: item.mobile || "",
    email: item.email || "",
  }
}

function buildPayload(userId: number, form: EditForm): UpdateAdminUserPayload {
  return {
    id: userId,
    nickname: form.nickname.trim(),
    avatar: form.avatar.trim(),
    mobile: toNullableString(form.mobile),
    email: toNullableString(form.email),
    remark: "",
  }
}

export function EditDrawer({
  open,
  saving,
  itemId,
  onOpenChange,
  onSubmit,
}: UserEditDrawerProps) {
  return (
    <Drawer open={open} onOpenChange={onOpenChange} direction="right">
      {open ? (
        <UserEditDrawerBody
          key={itemId ? `edit-${itemId}` : "edit"}
          itemId={itemId}
          saving={saving}
          onOpenChange={onOpenChange}
          onSubmit={onSubmit}
        />
      ) : null}
    </Drawer>
  )
}

type UserEditDrawerBodyProps = {
  saving: boolean
  itemId: number | null
  onOpenChange: (open: boolean) => void
  onSubmit: (payload: UpdateAdminUserPayload) => Promise<void>
}

function UserEditDrawerBody({
  saving,
  itemId,
  onOpenChange,
  onSubmit,
}: UserEditDrawerBodyProps) {
  const [loading, setLoading] = useState(false)
  const [item, setItem] = useState<AdminUser | null>(null)
  const form = useForm<
    z.input<typeof editFormSchema>,
    undefined,
    z.output<typeof editFormSchema>
  >({
    resolver: editFormResolver,
    defaultValues: emptyForm,
  })
  const {
    handleSubmit,
    reset,
    register,
    formState: { errors },
  } = form

  useEffect(() => {
    async function loadDetail() {
      if (!itemId) {
        setItem(null)
        reset(emptyForm)
        return
      }
      setLoading(true)
      try {
        const data = await fetchUserDetail(itemId)
        setItem(data)
        reset(buildForm(data))
      } catch (error) {
        console.error("Failed to load user:", error)
      } finally {
        setLoading(false)
      }
    }
    void loadDetail()
  }, [itemId, reset])

  async function onFormSubmit(values: EditForm) {
    if (!itemId) {
      return
    }

    await onSubmit(buildPayload(itemId, values))
  }

  return (
    <DrawerContent className="min-w-2xl">
      <DrawerHeader>
        <DrawerTitle>修改用户</DrawerTitle>
        <DrawerDescription>当前用户：{item?.username || "-"}</DrawerDescription>
      </DrawerHeader>
      {loading ? (
        <div className="flex items-center justify-center py-12">
          <div className="text-muted-foreground">加载中...</div>
        </div>
      ) : (
        <form
          className="flex h-full flex-col"
          onSubmit={handleSubmit(onFormSubmit)}
        >
          <div className="space-y-4 px-4 pb-4">
            <Field data-invalid={!!errors.nickname}>
              <FieldLabel htmlFor="user-nickname">昵称</FieldLabel>
              <FieldContent>
                <Input
                  id="user-nickname"
                  placeholder="请输入昵称"
                  aria-invalid={!!errors.nickname}
                  {...register("nickname")}
                />
                <FieldError errors={[errors.nickname]} />
              </FieldContent>
            </Field>
            <Field data-invalid={!!errors.avatar}>
              <FieldLabel htmlFor="user-avatar">头像地址</FieldLabel>
              <FieldContent>
                <Input
                  id="user-avatar"
                  placeholder="请输入头像 URL"
                  aria-invalid={!!errors.avatar}
                  {...register("avatar")}
                />
                <FieldError errors={[errors.avatar]} />
              </FieldContent>
            </Field>
            <Field data-invalid={!!errors.mobile}>
              <FieldLabel htmlFor="user-mobile">手机号</FieldLabel>
              <FieldContent>
                <Input
                  id="user-mobile"
                  placeholder="请输入手机号"
                  aria-invalid={!!errors.mobile}
                  {...register("mobile")}
                />
                <FieldError errors={[errors.mobile]} />
              </FieldContent>
            </Field>
            <Field data-invalid={!!errors.email}>
              <FieldLabel htmlFor="user-email">邮箱</FieldLabel>
              <FieldContent>
                <Input
                  id="user-email"
                  placeholder="请输入邮箱"
                  aria-invalid={!!errors.email}
                  {...register("email")}
                />
                <FieldError errors={[errors.email]} />
              </FieldContent>
            </Field>
          </div>
          <DrawerFooter className="border-t">
            <Button type="submit" disabled={saving || loading}>
              {saving ? "保存中..." : "保存修改"}
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
      )}
    </DrawerContent>
  );
}
