"use client"

import { Resolver, useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { z } from "zod/v4"

import { type CreateAdminRolePayload } from "@/lib/api/admin"
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
import { Textarea } from "@/components/ui/textarea"

type CreateRoleDrawerProps = {
  open: boolean
  saving: boolean
  onOpenChange: (open: boolean) => void
  onSubmit: (payload: CreateAdminRolePayload) => Promise<void>
}

const createFormSchema = z.object({
  name: z.string().trim().min(1, "角色名称不能为空"),
  code: z
    .string()
    .trim()
    .min(1, "角色编码不能为空")
    .regex(/^[A-Za-z][A-Za-z0-9:_-]*$/, "角色编码需以字母开头，仅支持字母、数字、冒号、下划线和短横线"),
  remark: z.string().trim(),
})

type CreateForm = z.infer<typeof createFormSchema>

const createFormResolver = zodResolver(createFormSchema as never) as Resolver<
  z.input<typeof createFormSchema>,
  undefined,
  z.output<typeof createFormSchema>
>

const emptyForm: CreateForm = {
  name: "",
  code: "",
  remark: "",
}

function buildEmptyForm(): CreateForm {
  return {
    ...emptyForm,
  }
}

function buildPayload(form: CreateForm): CreateAdminRolePayload {
  return {
    name: form.name.trim(),
    code: form.code.trim(),
    remark: form.remark.trim(),
  }
}

export function CreateRoleDrawer({
  open,
  saving,
  onOpenChange,
  onSubmit,
}: CreateRoleDrawerProps) {
  return (
    <Drawer open={open} onOpenChange={onOpenChange} direction="right">
      {open ? (
        <CreateRoleDrawerBody
          key="create-role"
          saving={saving}
          onOpenChange={onOpenChange}
          onSubmit={onSubmit}
        />
      ) : null}
    </Drawer>
  )
}

type CreateRoleDrawerBodyProps = {
  saving: boolean
  onOpenChange: (open: boolean) => void
  onSubmit: (payload: CreateAdminRolePayload) => Promise<void>
}

function CreateRoleDrawerBody({
  saving,
  onOpenChange,
  onSubmit,
}: CreateRoleDrawerBodyProps) {
  const form = useForm<
    z.input<typeof createFormSchema>,
    undefined,
    z.output<typeof createFormSchema>
  >({
    resolver: createFormResolver,
    defaultValues: buildEmptyForm(),
  })
  const {
    handleSubmit,
    register,
    reset,
    formState: { errors },
  } = form

  async function onFormSubmit(values: CreateForm) {
    await onSubmit(buildPayload(values))
    reset(buildEmptyForm())
  }

  return (
    <DrawerContent className="min-w-2xl">
      <DrawerHeader>
        <DrawerTitle>添加角色</DrawerTitle>
        <DrawerDescription>创建后可在列表中分配权限和调整排序。</DrawerDescription>
      </DrawerHeader>
      <form
        className="flex h-full flex-col"
        onSubmit={handleSubmit(onFormSubmit)}
      >
        <div className="space-y-4 overflow-y-auto px-4 pb-4">
          <Field data-invalid={!!errors.name}>
            <FieldLabel htmlFor="create-role-name">角色名称</FieldLabel>
            <FieldContent>
              <Input
                id="create-role-name"
                placeholder="例如：客服主管"
                autoComplete="off"
                aria-invalid={!!errors.name}
                {...register("name")}
              />
              <FieldError errors={[errors.name]} />
            </FieldContent>
          </Field>
          <Field data-invalid={!!errors.code}>
            <FieldLabel htmlFor="create-role-code">角色编码</FieldLabel>
            <FieldContent>
              <Input
                id="create-role-code"
                placeholder="例如：support_manager"
                autoComplete="off"
                aria-invalid={!!errors.code}
                {...register("code")}
              />
              <FieldError errors={[errors.code]} />
            </FieldContent>
          </Field>
          <Field data-invalid={!!errors.remark}>
            <FieldLabel htmlFor="create-role-remark">备注</FieldLabel>
            <FieldContent>
              <Textarea
                id="create-role-remark"
                placeholder="可选"
                aria-invalid={!!errors.remark}
                {...register("remark")}
              />
              <FieldError errors={[errors.remark]} />
            </FieldContent>
          </Field>
        </div>
        <DrawerFooter className="border-t">
          <Button type="submit" disabled={saving}>
            {saving ? "创建中..." : "创建角色"}
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
