"use client"

import { useEffect, useState } from "react"
import { zodResolver } from "@hookform/resolvers/zod"
import { Controller, Resolver, useForm } from "react-hook-form"
import { z } from "zod/v4"

import { ProjectDialog } from "@/components/project-dialog"
import { Button } from "@/components/ui/button"
import {
  Field,
  FieldContent,
  FieldError,
  FieldLabel,
} from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Textarea } from "@/components/ui/textarea"
import {
  type TicketCategory,
  type CreateTicketCategoryPayload,
  fetchTicketCategory,
} from "@/lib/api/admin"

type EditDialogProps = {
  open: boolean
  saving: boolean
  itemId: number | null
  onOpenChange: (open: boolean) => void
  onSubmit: (payload: CreateTicketCategoryPayload) => Promise<void>
}

const emptyForm: EditForm = {
  parentId: 0,
  name: "",
  code: "",
  description: "",
  status: "1",
  remark: "",
}

const formStatusOptions = [
  { value: "1", label: "启用" },
  { value: "0", label: "停用" },
] as const

const ticketCategoryFormSchema = z.object({
  parentId: z.number(),
  name: z.string().trim().min(1, "分类名称不能为空"),
  code: z.string().trim().min(1, "分类编码不能为空"),
  description: z.string().trim(),
  status: z.enum(["0", "1"], { message: "请选择状态" }),
  remark: z.string().trim(),
})

type EditForm = z.infer<typeof ticketCategoryFormSchema>
const editFormResolver = zodResolver(ticketCategoryFormSchema as never) as Resolver<
  z.input<typeof ticketCategoryFormSchema>,
  undefined,
  z.output<typeof ticketCategoryFormSchema>
>

function getStatusLabel(value: string) {
  return (
    formStatusOptions.find((option) => option.value === value)?.label ??
    "请选择状态"
  )
}

function buildForm(item: TicketCategory | null): EditForm {
  if (!item) {
    return emptyForm
  }

  return {
    parentId: item.parentId,
    name: item.name,
    code: item.code,
    description: item.description || "",
    status: item.status === 1 ? "1" : "0",
    remark: item.remark || "",
  }
}

function buildPayload(form: EditForm): CreateTicketCategoryPayload {
  return {
    parentId: form.parentId,
    name: form.name.trim(),
    code: form.code.trim(),
    description: form.description.trim(),
    status: Number(form.status) === 1 ? 1 : 0,
    remark: form.remark.trim(),
  }
}

export function EditDialog({
  open,
  saving,
  itemId,
  onOpenChange,
  onSubmit,
}: EditDialogProps) {
  if (!open) {
    return null
  }

  return (
    <EditDialogBody
      key={itemId ? `edit-${itemId}` : "create"}
      open={open}
      itemId={itemId}
      saving={saving}
      onOpenChange={onOpenChange}
      onSubmit={onSubmit}
    />
  )
}

type EditDialogBodyProps = {
  open: boolean
  saving: boolean
  itemId: number | null
  onOpenChange: (open: boolean) => void
  onSubmit: (payload: CreateTicketCategoryPayload) => Promise<void>
}

function EditDialogBody({
  open,
  saving,
  itemId,
  onOpenChange,
  onSubmit,
}: EditDialogBodyProps) {
  const formId = "ticket-category-edit-form"
  const [loading, setLoading] = useState(false)
  const form = useForm<
    z.input<typeof ticketCategoryFormSchema>,
    undefined,
    z.output<typeof ticketCategoryFormSchema>
  >({
    resolver: editFormResolver,
    defaultValues: emptyForm,
  })
  const {
    control,
    handleSubmit,
    reset,
    register,
    formState: { errors },
  } = form

  useEffect(() => {
    async function loadDetail() {
      if (!itemId) {
        reset(emptyForm)
        return
      }
      setLoading(true)
      try {
        const data = await fetchTicketCategory(itemId)
        reset(buildForm(data))
      } catch (error) {
        console.error("Failed to load ticket category:", error)
      } finally {
        setLoading(false)
      }
    }
    void loadDetail()
  }, [itemId, reset])

  async function onFormSubmit(values: EditForm) {
    const payload = buildPayload(values)
    await onSubmit(payload)
  }

  return (
    <ProjectDialog
      open={open}
      onOpenChange={onOpenChange}
      title={itemId ? "编辑分类" : "新建分类"}
      size="md"
      allowFullscreen
      footer={
        <>
          <Button
            type="button"
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={saving}
          >
            取消
          </Button>
          <Button type="submit" form={formId} disabled={saving || loading}>
            {saving ? "保存中..." : itemId ? "保存" : "创建"}
          </Button>
        </>
      }
    >
      {loading ? (
        <div className="flex items-center justify-center py-12">
          <div className="text-muted-foreground">加载中...</div>
        </div>
      ) : (
        <form id={formId} onSubmit={handleSubmit(onFormSubmit)} className="space-y-4">
          <Field data-invalid={!!errors.name}>
            <FieldLabel htmlFor="ticket-category-name">分类名称</FieldLabel>
            <FieldContent>
              <Input
                id="ticket-category-name"
                placeholder="例如：咨询、投诉、售后"
                aria-invalid={!!errors.name}
                {...register("name")}
              />
              <FieldError errors={[errors.name]} />
            </FieldContent>
          </Field>
          <Field data-invalid={!!errors.code}>
            <FieldLabel htmlFor="ticket-category-code">分类编码</FieldLabel>
            <FieldContent>
              <Input
                id="ticket-category-code"
                placeholder="例如：consult、complaint、after-sale"
                aria-invalid={!!errors.code}
                {...register("code")}
              />
              <FieldError errors={[errors.code]} />
            </FieldContent>
          </Field>
          <Field data-invalid={!!errors.description}>
            <FieldLabel htmlFor="ticket-category-description">描述</FieldLabel>
            <FieldContent>
              <Textarea
                id="ticket-category-description"
                placeholder="请输入分类描述"
                rows={3}
                aria-invalid={!!errors.description}
                {...register("description")}
              />
              <FieldError errors={[errors.description]} />
            </FieldContent>
          </Field>
          <Field data-invalid={!!errors.status}>
            <FieldLabel htmlFor="ticket-category-status">状态</FieldLabel>
            <FieldContent>
              <Controller
                control={control}
                name="status"
                render={({ field }) => (
                  <Select
                    value={field.value}
                    onValueChange={field.onChange}
                    modal={false}
                  >
                    <SelectTrigger
                      id="ticket-category-status"
                      className="w-full"
                      aria-invalid={!!errors.status}
                    >
                      <SelectValue>{getStatusLabel(field.value)}</SelectValue>
                    </SelectTrigger>
                    <SelectContent>
                      {formStatusOptions.map((option) => (
                        <SelectItem key={option.value} value={option.value}>
                          {option.label}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                )}
              />
              <FieldError errors={[errors.status]} />
            </FieldContent>
          </Field>
          <Field data-invalid={!!errors.remark}>
            <FieldLabel htmlFor="ticket-category-remark">备注</FieldLabel>
            <FieldContent>
              <Textarea
                id="ticket-category-remark"
                placeholder="请输入备注信息"
                rows={2}
                aria-invalid={!!errors.remark}
                {...register("remark")}
              />
              <FieldError errors={[errors.remark]} />
            </FieldContent>
          </Field>
        </form>
      )}
    </ProjectDialog>
  )
}
