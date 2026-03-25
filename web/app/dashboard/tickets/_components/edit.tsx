"use client"

import { useEffect, useState } from "react"
import { zodResolver } from "@hookform/resolvers/zod"
import { Controller, Resolver, useForm } from "react-hook-form"
import { z } from "zod/v4"

import { fetchTicketCategories, type TicketCategory } from "@/lib/api/admin"
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
import { type CreateTicketPayload } from "@/lib/api/admin"

type EditDialogProps = {
  open: boolean
  saving: boolean
  onOpenChange: (open: boolean) => void
  onSubmit: (payload: CreateTicketPayload) => Promise<void>
}

const emptyForm: EditForm = {
  title: "",
  content: "",
  categoryId: 0,
  priority: "0",
  externalUserName: "",
  externalUserEmail: "",
  externalUserMobile: "",
  tags: "",
  remark: "",
}

const ticketPriorityOptions = [
  { value: "0", label: "普通" },
  { value: "1", label: "低" },
  { value: "2", label: "中" },
  { value: "3", label: "高" },
  { value: "4", label: "紧急" },
] as const

const ticketFormSchema = z.object({
  title: z.string().trim().min(1, "标题不能为空"),
  content: z.string().trim(),
  categoryId: z.number(),
  priority: z.enum(["0", "1", "2", "3", "4"], { message: "请选择优先级" }),
  externalUserName: z.string().trim(),
  externalUserEmail: z.string().trim().email("邮箱格式不正确").or(z.string().trim().max(0)),
  externalUserMobile: z.string().trim(),
  tags: z.string().trim(),
  remark: z.string().trim(),
})

type EditForm = z.infer<typeof ticketFormSchema>
const editFormResolver = zodResolver(ticketFormSchema as never) as Resolver<
  z.input<typeof ticketFormSchema>,
  undefined,
  z.output<typeof ticketFormSchema>
>

function getPriorityLabel(value: string) {
  return (
    ticketPriorityOptions.find((option) => option.value === value)?.label ??
    "请选择优先级"
  )
}

function buildForm(): EditForm {
  return emptyForm
}

function buildPayload(form: EditForm): CreateTicketPayload {
  return {
    title: form.title.trim(),
    content: form.content.trim(),
    channelType: "admin",
    channelId: "",
    categoryId: form.categoryId,
    priority: Number(form.priority),
    sourceUserId: 0,
    externalUserId: "",
    externalUserName: form.externalUserName.trim(),
    externalUserEmail: form.externalUserEmail.trim(),
    externalUserMobile: form.externalUserMobile.trim(),
    conversationId: 0,
    tags: form.tags.trim(),
    remark: form.remark.trim(),
  }
}

export function EditDialog({
  open,
  saving,
  onOpenChange,
  onSubmit,
}: EditDialogProps) {
  if (!open) {
    return null
  }

  return (
    <EditDialogBody
      key="create"
      open={open}
      saving={saving}
      onOpenChange={onOpenChange}
      onSubmit={onSubmit}
    />
  )
}

type EditDialogBodyProps = {
  open: boolean
  saving: boolean
  onOpenChange: (open: boolean) => void
  onSubmit: (payload: CreateTicketPayload) => Promise<void>
}

function EditDialogBody({
  open,
  saving,
  onOpenChange,
  onSubmit,
}: EditDialogBodyProps) {
  const [categories, setCategories] = useState<TicketCategory[]>([])
  const [loadingCategories, setLoadingCategories] = useState(true)

  const formId = "ticket-edit-form"
  const form = useForm<
    z.input<typeof ticketFormSchema>,
    undefined,
    z.output<typeof ticketFormSchema>
  >({
    resolver: editFormResolver,
    defaultValues: buildForm(),
  })
  const {
    control,
    handleSubmit,
    reset,
    register,
    formState: { errors },
  } = form

  useEffect(() => {
    reset(buildForm())
    let ignore = false
    setLoadingCategories(true)
    fetchTicketCategories({ limit: 100 })
      .then((data) => {
        if (!ignore) {
          setCategories(data.results)
        }
      })
      .finally(() => {
        if (!ignore) {
          setLoadingCategories(false)
        }
      })
    return () => {
      ignore = true
    }
  }, [reset])

  async function onFormSubmit(values: EditForm) {
    const payload = buildPayload(values)
    await onSubmit(payload)
  }

  return (
    <ProjectDialog
      open={open}
      onOpenChange={onOpenChange}
      title="新建工单"
      size="lg"
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
          <Button type="submit" form={formId} disabled={saving}>
            {saving ? "保存中..." : "创建"}
          </Button>
        </>
      }
    >
      <form id={formId} onSubmit={handleSubmit(onFormSubmit)} className="space-y-4">
        <Field data-invalid={!!errors.title}>
          <FieldLabel htmlFor="ticket-title">工单标题</FieldLabel>
          <FieldContent>
            <Input
              id="ticket-title"
              placeholder="请输入工单标题"
              aria-invalid={!!errors.title}
              {...register("title")}
            />
            <FieldError errors={[errors.title]} />
          </FieldContent>
        </Field>
        <Field data-invalid={!!errors.content}>
          <FieldLabel htmlFor="ticket-content">工单内容</FieldLabel>
          <FieldContent>
            <Textarea
              id="ticket-content"
              placeholder="请输入工单内容"
              rows={4}
              aria-invalid={!!errors.content}
              {...register("content")}
            />
            <FieldError errors={[errors.content]} />
          </FieldContent>
        </Field>
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <Field data-invalid={!!errors.categoryId}>
            <FieldLabel htmlFor="ticket-category">工单分类</FieldLabel>
            <FieldContent>
              <Controller
                control={control}
                name="categoryId"
                render={({ field }) => (
                  <Select
                    value={String(field.value)}
                    onValueChange={(value) => field.onChange(Number(value))}
                    modal={false}
                    disabled={loadingCategories}
                  >
                    <SelectTrigger
                      id="ticket-category"
                      className="w-full"
                      aria-invalid={!!errors.categoryId}
                    >
                      <SelectValue placeholder="请选择分类" />
                    </SelectTrigger>
                    <SelectContent>
                      {categories.map((item) => (
                        <SelectItem key={item.id} value={String(item.id)}>
                          {item.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                )}
              />
              <FieldError errors={[errors.categoryId]} />
            </FieldContent>
          </Field>
          <Field data-invalid={!!errors.priority}>
            <FieldLabel htmlFor="ticket-priority">优先级</FieldLabel>
            <FieldContent>
              <Controller
                control={control}
                name="priority"
                render={({ field }) => (
                  <Select
                    value={field.value}
                    onValueChange={field.onChange}
                    modal={false}
                  >
                    <SelectTrigger
                      id="ticket-priority"
                      className="w-full"
                      aria-invalid={!!errors.priority}
                    >
                      <SelectValue>{getPriorityLabel(field.value)}</SelectValue>
                    </SelectTrigger>
                    <SelectContent>
                      {ticketPriorityOptions.map((option) => (
                        <SelectItem key={option.value} value={option.value}>
                          {option.label}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                )}
              />
              <FieldError errors={[errors.priority]} />
            </FieldContent>
          </Field>
        </div>
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
          <Field data-invalid={!!errors.externalUserName}>
            <FieldLabel htmlFor="ticket-user-name">客户名称</FieldLabel>
            <FieldContent>
              <Input
                id="ticket-user-name"
                placeholder="请输入客户名称"
                aria-invalid={!!errors.externalUserName}
                {...register("externalUserName")}
              />
              <FieldError errors={[errors.externalUserName]} />
            </FieldContent>
          </Field>
          <Field data-invalid={!!errors.externalUserEmail}>
            <FieldLabel htmlFor="ticket-user-email">客户邮箱</FieldLabel>
            <FieldContent>
              <Input
                id="ticket-user-email"
                type="email"
                placeholder="请输入客户邮箱"
                aria-invalid={!!errors.externalUserEmail}
                {...register("externalUserEmail")}
              />
              <FieldError errors={[errors.externalUserEmail]} />
            </FieldContent>
          </Field>
          <Field data-invalid={!!errors.externalUserMobile}>
            <FieldLabel htmlFor="ticket-user-mobile">客户手机</FieldLabel>
            <FieldContent>
              <Input
                id="ticket-user-mobile"
                placeholder="请输入客户手机"
                aria-invalid={!!errors.externalUserMobile}
                {...register("externalUserMobile")}
              />
              <FieldError errors={[errors.externalUserMobile]} />
            </FieldContent>
          </Field>
        </div>
        <Field data-invalid={!!errors.tags}>
          <FieldLabel htmlFor="ticket-tags">标签</FieldLabel>
          <FieldContent>
            <Input
              id="ticket-tags"
              placeholder="多个标签用逗号分隔"
              aria-invalid={!!errors.tags}
              {...register("tags")}
            />
            <FieldError errors={[errors.tags]} />
          </FieldContent>
        </Field>
        <Field data-invalid={!!errors.remark}>
          <FieldLabel htmlFor="ticket-remark">备注</FieldLabel>
          <FieldContent>
            <Textarea
              id="ticket-remark"
              placeholder="请输入备注信息"
              rows={2}
              aria-invalid={!!errors.remark}
              {...register("remark")}
            />
            <FieldError errors={[errors.remark]} />
          </FieldContent>
        </Field>
      </form>
    </ProjectDialog>
  )
}
