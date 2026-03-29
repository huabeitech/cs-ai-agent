"use client"

import { useEffect, useMemo, useState } from "react"
import { zodResolver } from "@hookform/resolvers/zod"
import { Controller, type Resolver, useForm } from "react-hook-form"
import { z } from "zod/v4"

import { ProjectDialog } from "@/components/project-dialog"
import { OptionCombobox, type ComboboxOption } from "@/components/option-combobox"
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
import { fetchCompanies } from "@/lib/api/company"
import {
  fetchCustomer,
  type AdminCustomer,
  type CreateAdminCustomerPayload,
} from "@/lib/api/customer"
import { getEnumLabel, getEnumOptions } from "@/lib/enums"
import { Gender, GenderLabels } from "@/lib/generated/enums"

export type CustomerFormDialogProps = {
  open: boolean
  saving: boolean
  itemId: number | null
  onOpenChange: (open: boolean) => void
  onSubmit: (payload: CreateAdminCustomerPayload) => Promise<void>
}

const genderOptions = [
  ...getEnumOptions(GenderLabels).map((item) => ({
    value: String(item.value),
    label: item.label,
  })),
] as const

const genderValueOptions = [
  String(Gender.Unknown),
  String(Gender.Male),
  String(Gender.Female),
] as const

const customerFormSchema = z.object({
  name: z.string().trim().min(1, "客户名称不能为空"),
  gender: z.enum(genderValueOptions, { message: "请选择性别" }),
  companyId: z.string().trim().regex(/^\d+$/, "请选择所属公司"),
  primaryMobile: z.string().trim(),
  primaryEmail: z.string().trim(),
  remark: z.string().trim(),
})

type EditForm = z.infer<typeof customerFormSchema>

const editFormResolver = zodResolver(customerFormSchema as never) as Resolver<
  z.input<typeof customerFormSchema>,
  undefined,
  z.output<typeof customerFormSchema>
>

const emptyForm: EditForm = {
  name: "",
  gender: "0",
  companyId: "0",
  primaryMobile: "",
  primaryEmail: "",
  remark: "",
}

function buildForm(item: AdminCustomer | null): EditForm {
  if (!item) return emptyForm
  return {
    name: item.name,
    gender: String(item.gender) as "0" | "1" | "2",
    companyId: String(item.companyId ?? 0),
    primaryMobile: item.primaryMobile ?? "",
    primaryEmail: item.primaryEmail ?? "",
    remark: item.remark ?? "",
  }
}

function buildPayload(form: EditForm): CreateAdminCustomerPayload {
  return {
    name: form.name.trim(),
    gender: Number(form.gender),
    companyId: Number(form.companyId),
    primaryMobile: form.primaryMobile.trim(),
    primaryEmail: form.primaryEmail.trim(),
    remark: form.remark.trim(),
  }
}

function getGenderLabel(value: string) {
  return getEnumLabel(GenderLabels, Number(value) as Gender)
}

/** 客户新建/编辑表单弹窗（ProjectDialog + react-hook-form），供客户管理页与会话工作台等复用。 */
export function CustomerFormDialog({
  open,
  saving,
  itemId,
  onOpenChange,
  onSubmit,
}: CustomerFormDialogProps) {
  if (!open) return null
  return (
    <CustomerFormDialogBody
      key={itemId ? `edit-${itemId}` : "create"}
      open={open}
      saving={saving}
      itemId={itemId}
      onOpenChange={onOpenChange}
      onSubmit={onSubmit}
    />
  )
}

type CustomerFormDialogBodyProps = CustomerFormDialogProps

function CustomerFormDialogBody({
  open,
  saving,
  itemId,
  onOpenChange,
  onSubmit,
}: CustomerFormDialogBodyProps) {
  const formId = "customer-form-dialog"
  const [loading, setLoading] = useState(false)
  const [companyOptions, setCompanyOptions] = useState<ComboboxOption[]>([
    { value: "0", label: "无所属公司" },
  ])
  const form = useForm<
    z.input<typeof customerFormSchema>,
    undefined,
    z.output<typeof customerFormSchema>
  >({
    resolver: editFormResolver,
    defaultValues: emptyForm,
  })
  const {
    control,
    handleSubmit,
    register,
    reset,
    formState: { errors },
  } = form

  const selectedCompanyLabel = useMemo(() => {
    const current = companyOptions.find((item) => item.value === form.getValues("companyId"))
    return current?.label ?? "请选择公司"
  }, [companyOptions, form])

  useEffect(() => {
    async function loadCompanies() {
      try {
        const data = await fetchCompanies({ status: 0, page: 1, limit: 200 })
        const options: ComboboxOption[] = [
          { value: "0", label: "无所属公司" },
          ...data.results.map((item) => ({
            value: String(item.id),
            label: item.name,
          })),
        ]
        setCompanyOptions(options)
      } catch {
        // ignore, keep minimal options
      }
    }
    void loadCompanies()
  }, [])

  useEffect(() => {
    async function loadDetail() {
      if (!itemId) {
        reset(emptyForm)
        return
      }
      setLoading(true)
      try {
        const data = await fetchCustomer(itemId)
        reset(buildForm(data))
      } finally {
        setLoading(false)
      }
    }
    void loadDetail()
  }, [itemId, reset])

  async function onFormSubmit(values: EditForm) {
    await onSubmit(buildPayload(values))
  }

  return (
    <ProjectDialog
      open={open}
      onOpenChange={(next) => onOpenChange(next)}
      title={itemId ? "编辑客户" : "新建客户"}
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
            <FieldLabel htmlFor="customer-name">客户名称</FieldLabel>
            <FieldContent>
              <Input
                id="customer-name"
                placeholder="请输入客户名称"
                aria-invalid={!!errors.name}
                {...register("name")}
              />
              <FieldError errors={[errors.name]} />
            </FieldContent>
          </Field>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <Field data-invalid={!!errors.gender}>
              <FieldLabel htmlFor="customer-gender">性别</FieldLabel>
              <FieldContent>
                <Controller
                  control={control}
                  name="gender"
                  render={({ field }) => (
                    <Select value={field.value} onValueChange={field.onChange} modal={false}>
                      <SelectTrigger id="customer-gender">
                        <SelectValue>{getGenderLabel(field.value)}</SelectValue>
                      </SelectTrigger>
                      <SelectContent>
                        {genderOptions.map((option) => (
                          <SelectItem key={option.value} value={option.value}>
                            {option.label}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  )}
                />
                <FieldError errors={[errors.gender]} />
              </FieldContent>
            </Field>

            <Field data-invalid={!!errors.companyId}>
              <FieldLabel htmlFor="customer-company">所属公司</FieldLabel>
              <FieldContent>
                <Controller
                  control={control}
                  name="companyId"
                  render={({ field }) => (
                    <div className="w-full">
                      <OptionCombobox
                        value={field.value}
                        options={companyOptions}
                        placeholder={selectedCompanyLabel}
                        searchPlaceholder="搜索公司名称"
                        onChange={field.onChange}
                      />
                    </div>
                  )}
                />
                <FieldError errors={[errors.companyId]} />
              </FieldContent>
            </Field>
          </div>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <Field data-invalid={!!errors.primaryMobile}>
              <FieldLabel htmlFor="customer-mobile">手机号</FieldLabel>
              <FieldContent>
                <Input
                  id="customer-mobile"
                  placeholder="可选"
                  aria-invalid={!!errors.primaryMobile}
                  {...register("primaryMobile")}
                />
                <FieldError errors={[errors.primaryMobile]} />
              </FieldContent>
            </Field>
            <Field data-invalid={!!errors.primaryEmail}>
              <FieldLabel htmlFor="customer-email">邮箱</FieldLabel>
              <FieldContent>
                <Input
                  id="customer-email"
                  placeholder="可选"
                  aria-invalid={!!errors.primaryEmail}
                  {...register("primaryEmail")}
                />
                <FieldError errors={[errors.primaryEmail]} />
              </FieldContent>
            </Field>
          </div>

          <Field data-invalid={!!errors.remark}>
            <FieldLabel htmlFor="customer-remark">备注</FieldLabel>
            <FieldContent>
              <Textarea
                id="customer-remark"
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
