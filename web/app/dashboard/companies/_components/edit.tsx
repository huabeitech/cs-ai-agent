"use client"

import { useEffect, useState } from "react"
import { zodResolver } from "@hookform/resolvers/zod"
import type { Resolver } from "react-hook-form"
import { useForm } from "react-hook-form"
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
import { Textarea } from "@/components/ui/textarea"
import {
  fetchCompany,
  type AdminCompany,
  type CreateAdminCompanyPayload,
} from "@/lib/api/company"

type CompanyEditDialogProps = {
  open: boolean
  saving: boolean
  itemId: number | null
  initialValues?: Partial<CreateAdminCompanyPayload>
  onOpenChange: (open: boolean) => void
  onSubmit: (payload: CreateAdminCompanyPayload) => Promise<void>
}

const companyFormSchema = z.object({
  name: z.string().trim().min(1, "公司名称不能为空"),
  code: z.string().trim(),
  remark: z.string().trim(),
})

type EditForm = z.infer<typeof companyFormSchema>

const editFormResolver = zodResolver(companyFormSchema as never) as Resolver<
  z.input<typeof companyFormSchema>,
  undefined,
  z.output<typeof companyFormSchema>
>

const emptyForm: EditForm = {
  name: "",
  code: "",
  remark: "",
}

function buildForm(item: AdminCompany | null): EditForm {
  if (!item) {
    return emptyForm
  }
  return {
    name: item.name,
    code: item.code,
    remark: item.remark,
  }
}

function buildPayload(form: EditForm): CreateAdminCompanyPayload {
  return {
    name: form.name.trim(),
    code: form.code.trim(),
    remark: form.remark.trim(),
  }
}

function buildInitialForm(initialValues?: Partial<CreateAdminCompanyPayload>): EditForm {
  return {
    name: initialValues?.name?.trim() ?? "",
    code: initialValues?.code?.trim() ?? "",
    remark: initialValues?.remark?.trim() ?? "",
  }
}

export function EditDialog({
  open,
  saving,
  itemId,
  initialValues,
  onOpenChange,
  onSubmit,
}: CompanyEditDialogProps) {
  if (!open) {
    return null
  }
  return (
    <CompanyEditDialogBody
      key={itemId ? `edit-${itemId}` : "create"}
      open={open}
      saving={saving}
      itemId={itemId}
      initialValues={initialValues}
      onOpenChange={onOpenChange}
      onSubmit={onSubmit}
    />
  )
}

type CompanyEditDialogBodyProps = CompanyEditDialogProps

function CompanyEditDialogBody({
  open,
  saving,
  itemId,
  initialValues,
  onOpenChange,
  onSubmit,
}: CompanyEditDialogBodyProps) {
  const formId = "company-edit-form"
  const [loading, setLoading] = useState(false)
  const form = useForm<
    z.input<typeof companyFormSchema>,
    undefined,
    z.output<typeof companyFormSchema>
  >({
    resolver: editFormResolver,
    defaultValues: emptyForm,
  })
  const {
    handleSubmit,
    register,
    reset,
    formState: { errors },
  } = form

  useEffect(() => {
    async function loadDetail() {
      if (!itemId) {
        reset(buildInitialForm(initialValues))
        return
      }
      setLoading(true)
      try {
        const data = await fetchCompany(itemId)
        reset(buildForm(data))
      } finally {
        setLoading(false)
      }
    }
    void loadDetail()
  }, [initialValues, itemId, reset])

  async function onFormSubmit(values: EditForm) {
    await onSubmit(buildPayload(values))
  }

  return (
    <ProjectDialog
      open={open}
      onOpenChange={onOpenChange}
      title={itemId ? "编辑公司" : "新建公司"}
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
            <FieldLabel htmlFor="company-name">公司名称</FieldLabel>
            <FieldContent>
              <Input
                id="company-name"
                placeholder="请输入公司名称"
                aria-invalid={!!errors.name}
                {...register("name")}
              />
              <FieldError errors={[errors.name]} />
            </FieldContent>
          </Field>
          <Field data-invalid={!!errors.code}>
            <FieldLabel htmlFor="company-code">公司编码</FieldLabel>
            <FieldContent>
              <Input
                id="company-code"
                placeholder="可选"
                aria-invalid={!!errors.code}
                {...register("code")}
              />
              <FieldError errors={[errors.code]} />
            </FieldContent>
          </Field>
          <Field data-invalid={!!errors.remark}>
            <FieldLabel htmlFor="company-remark">备注</FieldLabel>
            <FieldContent>
              <Textarea
                id="company-remark"
                placeholder="可选"
                rows={4}
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
