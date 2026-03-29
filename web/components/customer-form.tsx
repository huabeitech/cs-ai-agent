"use client"

import { useEffect, useMemo, useRef, useState } from "react"
import { zodResolver } from "@hookform/resolvers/zod"
import {
  Controller,
  useFieldArray,
  useForm,
  type Resolver,
  type UseFormReturn,
} from "react-hook-form"
import { PlusIcon, Trash2Icon } from "lucide-react"
import { z } from "zod/v4"

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
import {
  createCustomerContact,
  deleteCustomerContact,
  fetchCustomerContacts,
  updateCustomerContact,
  type AdminCustomerContact,
} from "@/lib/api/customer-contact"
import { fetchCompanies } from "@/lib/api/company"
import {
  fetchCustomer,
  type AdminCustomer,
  type CreateAdminCustomerPayload,
} from "@/lib/api/customer"
import { getEnumLabel, getEnumOptions } from "@/lib/enums"
import { ContactType, ContactTypeLabels, Gender, GenderLabels } from "@/lib/generated/enums"
import { cn } from "@/lib/utils"

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

const contactTypeValues = [
  ContactType.Mobile,
  ContactType.Email,
  ContactType.WeChat,
  ContactType.Other,
] as const

const contactRowSchema = z.object({
  id: z.number().optional(),
  contactType: z.enum(["mobile", "email", "wechat", "other"]),
  contactValue: z.string(),
  remark: z.string(),
  isPrimary: z.boolean(),
})

const customerFormSchema = z.object({
  name: z.string().trim().min(1, "客户名称不能为空"),
  gender: z.enum(genderValueOptions, { message: "请选择性别" }),
  companyId: z.string().trim().regex(/^\d+$/, "请选择所属公司"),
  remark: z.string().trim(),
  contacts: z.array(contactRowSchema),
})

export type CustomerFormValues = z.infer<typeof customerFormSchema>

export type CustomerContactFormRow = {
  id?: number
  contactType: (typeof contactTypeValues)[number]
  contactValue: string
  remark: string
  isPrimary: boolean
}

const customerFormResolver = zodResolver(customerFormSchema as never) as Resolver<
  z.input<typeof customerFormSchema>,
  undefined,
  z.output<typeof customerFormSchema>
>

function defaultContactRow(isPrimary: boolean): CustomerContactFormRow {
  return {
    contactType: ContactType.Mobile,
    contactValue: "",
    remark: "",
    isPrimary,
  }
}

const emptyCustomerForm: CustomerFormValues = {
  name: "",
  gender: "0",
  companyId: "0",
  remark: "",
  contacts: [defaultContactRow(true)],
}

function buildCustomerMainFromAdmin(item: AdminCustomer | null): Omit<CustomerFormValues, "contacts"> {
  if (!item) {
    return {
      name: "",
      gender: "0",
      companyId: "0",
      remark: "",
    }
  }
  return {
    name: item.name,
    gender: String(item.gender) as "0" | "1" | "2",
    companyId: String(item.companyId ?? 0),
    remark: item.remark ?? "",
  }
}

function buildContactsFromApi(list: AdminCustomerContact[]): CustomerContactFormRow[] {
  if (list.length === 0) {
    return [defaultContactRow(true)]
  }
  return list.map((c) => ({
    id: c.id,
    contactType: c.contactType as CustomerContactFormRow["contactType"],
    contactValue: c.contactValue ?? "",
    remark: c.remark ?? "",
    isPrimary: c.isPrimary,
  }))
}

/** 过滤空行并保证至多一条主联系方式（有一条有值时至少一条主） */
export function normalizeContactsForSubmit(rows: CustomerContactFormRow[]): CustomerContactFormRow[] {
  const withValue = rows.filter((r) => r.contactValue.trim() !== "")
  if (withValue.length === 0) {
    return []
  }
  const primaryIdx = withValue.findIndex((r) => r.isPrimary)
  if (primaryIdx < 0) {
    return withValue.map((r, i) => ({ ...r, isPrimary: i === 0 }))
  }
  return withValue.map((r, i) => ({
    ...r,
    isPrimary: i === primaryIdx,
  }))
}

function derivePrimaryFromContacts(contacts: CustomerContactFormRow[]): {
  primaryMobile: string
  primaryEmail: string
} {
  const primary = contacts.find((c) => c.isPrimary)
  if (!primary) {
    return { primaryMobile: "", primaryEmail: "" }
  }
  const v = primary.contactValue.trim()
  if (primary.contactType === ContactType.Email) {
    return { primaryMobile: "", primaryEmail: v }
  }
  return { primaryMobile: v, primaryEmail: "" }
}

function buildCustomerPayload(values: CustomerFormValues, contacts: CustomerContactFormRow[]): CreateAdminCustomerPayload {
  const { primaryMobile, primaryEmail } = derivePrimaryFromContacts(contacts)
  return {
    name: values.name.trim(),
    gender: Number(values.gender),
    companyId: Number(values.companyId),
    primaryMobile,
    primaryEmail,
    remark: values.remark.trim(),
  }
}

/**
 * 创建/更新联系方式记录（删除已移除、按「非主→主」顺序写入，便于后端主联系方式互斥逻辑）。
 */
export async function persistCustomerContacts(
  customerId: number,
  contacts: CustomerContactFormRow[],
  previous: AdminCustomerContact[] | null
): Promise<void> {
  const prevList = previous ?? []
  const nextIds = new Set(contacts.filter((c) => c.id).map((c) => c.id!))
  for (const p of prevList) {
    if (!nextIds.has(p.id)) {
      await deleteCustomerContact(p.id)
    }
  }
  const sorted = [...contacts].sort((a, b) => Number(a.isPrimary) - Number(b.isPrimary))
  for (const row of sorted) {
    const base = {
      contactType: row.contactType,
      contactValue: row.contactValue.trim(),
      isPrimary: row.isPrimary,
      isVerified: false,
      source: "manual",
      status: 0,
      remark: row.remark.trim(),
    }
    if (row.id) {
      await updateCustomerContact({ id: row.id, ...base })
    } else {
      await createCustomerContact({ customerId, ...base })
    }
  }
}

export type CustomerFormSavePayload = {
  customerPayload: CreateAdminCustomerPayload
  customerIdForUpdate: number | null
  contacts: CustomerContactFormRow[]
  previousContactRecords: AdminCustomerContact[] | null
}

function getGenderLabel(value: string) {
  return getEnumLabel(GenderLabels, Number(value) as Gender)
}

function getContactTypeLabel(value: string) {
  return ContactTypeLabels[value as ContactType] ?? value
}

type CustomerFormFieldsProps = {
  form: UseFormReturn<CustomerFormValues>
  companyOptions: ComboboxOption[]
  fieldIdPrefix?: string
  remarkRows?: number
}

function CustomerFormFields({
  form,
  companyOptions,
  fieldIdPrefix = "customer",
  remarkRows = 4,
}: CustomerFormFieldsProps) {
  const {
    control,
    register,
    formState: { errors },
    watch,
    setValue,
    getValues,
  } = form
  const { fields, append, remove } = useFieldArray({ control, name: "contacts" })

  const companyId = watch("companyId")
  const selectedCompanyLabel = useMemo(() => {
    return companyOptions.find((item) => item.value === companyId)?.label ?? "请选择公司"
  }, [companyOptions, companyId])

  const id = (suffix: string) => `${fieldIdPrefix}-${suffix}`

  function setPrimaryIndex(index: number) {
    fields.forEach((_, i) => {
      setValue(`contacts.${i}.isPrimary`, i === index)
    })
  }

  function addContactRow() {
    append(defaultContactRow(fields.length === 0))
  }

  function removeContactRow(index: number) {
    const wasPrimary = watch(`contacts.${index}.isPrimary`)
    remove(index)
    if (wasPrimary) {
      requestAnimationFrame(() => {
        const list = getValues("contacts")
        if (list.length > 0) {
          list.forEach((_, i) => setValue(`contacts.${i}.isPrimary`, i === 0))
        }
      })
    }
  }

  return (
    <div className="space-y-8">
      <div className="space-y-3">
        <h3 className="text-sm font-semibold text-muted-foreground">客户信息</h3>
        <div className="space-y-4">
          <Field data-invalid={!!errors.name}>
            <FieldLabel htmlFor={id("name")}>客户名称</FieldLabel>
            <FieldContent>
              <Input
                id={id("name")}
                placeholder="请输入客户名称"
                aria-invalid={!!errors.name}
                autoComplete="off"
                {...register("name")}
              />
              <FieldError errors={[errors.name]} />
            </FieldContent>
          </Field>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <Field data-invalid={!!errors.gender}>
              <FieldLabel htmlFor={id("gender")}>性别</FieldLabel>
              <FieldContent>
                <Controller
                  control={control}
                  name="gender"
                  render={({ field }) => (
                    <Select value={field.value} onValueChange={field.onChange} modal={false}>
                      <SelectTrigger id={id("gender")}>
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
              <FieldLabel htmlFor={id("company")}>所属公司</FieldLabel>
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

          <Field data-invalid={!!errors.remark}>
            <FieldLabel htmlFor={id("remark")}>备注</FieldLabel>
            <FieldContent>
              <Textarea
                id={id("remark")}
                placeholder="可选"
                rows={remarkRows}
                aria-invalid={!!errors.remark}
                {...register("remark")}
              />
              <FieldError errors={[errors.remark]} />
            </FieldContent>
          </Field>
        </div>
      </div>

      <div className="space-y-3">
        <h3 className="text-sm font-semibold text-muted-foreground">联系方式</h3>
        <p className="text-xs text-muted-foreground">
          类型与号码/邮箱；在右侧勾选主联系方式（与手机通讯录一致）。主联系方式会同步到列表检索字段。
        </p>

        <div className="space-y-4">
          {fields.map((field, index) => {
            const err = errors.contacts?.[index]
            return (
              <div
                key={field.id}
                className={cn(
                  "rounded-lg border border-border bg-muted/20 p-3",
                  index > 0 && "mt-2"
                )}
              >
                <div className="grid grid-cols-1 gap-3 sm:grid-cols-[minmax(0,1fr)_auto] sm:items-start">
                  <div className="min-w-0 space-y-3">
                    <div className="flex flex-col gap-2 sm:flex-row sm:items-start">
                      <Controller
                        control={control}
                        name={`contacts.${index}.contactType`}
                        render={({ field: f }) => (
                          <Select value={f.value} onValueChange={f.onChange} modal={false}>
                            <SelectTrigger className="w-full sm:w-[120px]" id={id(`ct-${index}`)}>
                              <SelectValue>{getContactTypeLabel(f.value)}</SelectValue>
                            </SelectTrigger>
                            <SelectContent>
                              {contactTypeValues.map((v) => (
                                <SelectItem key={v} value={v}>
                                  {getContactTypeLabel(v)}
                                </SelectItem>
                              ))}
                            </SelectContent>
                          </Select>
                        )}
                      />
                      <Field data-invalid={!!err?.contactValue} className="min-w-0 flex-1">
                        <FieldContent>
                          <Input
                            className="w-full"
                            placeholder={
                              watch(`contacts.${index}.contactType`) === ContactType.Email
                                ? "邮箱地址"
                                : "号码 / 账号"
                            }
                            aria-invalid={!!err?.contactValue}
                            {...register(`contacts.${index}.contactValue`)}
                          />
                          <FieldError errors={[err?.contactValue]} />
                        </FieldContent>
                      </Field>
                    </div>
                    <Field>
                      <FieldLabel htmlFor={id(`tag-${index}`)} className="text-xs">
                        标签（可选）
                      </FieldLabel>
                      <FieldContent>
                        <Input
                          id={id(`tag-${index}`)}
                          placeholder="如：工作、备用"
                          {...register(`contacts.${index}.remark`)}
                        />
                      </FieldContent>
                    </Field>
                  </div>

                  <div className="flex flex-row items-center justify-between gap-3 sm:flex-col sm:items-end sm:justify-start sm:pt-7">
                    <div className="flex items-center gap-2">
                      <input
                        type="radio"
                        className="size-4 accent-primary"
                        name={id("primary-group")}
                        checked={watch(`contacts.${index}.isPrimary`)}
                        onChange={() => setPrimaryIndex(index)}
                        id={id(`primary-${index}`)}
                      />
                      <label
                        htmlFor={id(`primary-${index}`)}
                        className="cursor-pointer text-sm font-normal"
                      >
                        主联系方式
                      </label>
                    </div>
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon"
                      className="shrink-0 text-muted-foreground hover:text-destructive"
                      onClick={() => removeContactRow(index)}
                      aria-label="删除此条联系方式"
                    >
                      <Trash2Icon className="size-4" />
                    </Button>
                  </div>
                </div>
              </div>
            )
          })}
        </div>

        <Button type="button" variant="outline" size="sm" className="gap-1" onClick={addContactRow}>
          <PlusIcon className="size-4" />
          添加联系方式
        </Button>
      </div>
    </div>
  )
}

/** 客户表单内公司列表请求参数（全项目统一） */
export const CUSTOMER_FORM_COMPANY_LIST_QUERY = {
  status: 0,
  page: 1,
  limit: 200,
} as const

const defaultCompanyOptions: ComboboxOption[] = [
  { value: "0", label: "无所属公司" },
]

export type CustomerFormProps = {
  formId: string
  onSave: (payload: CustomerFormSavePayload) => Promise<void> | void
  itemId?: number | null
  fieldIdPrefix?: string
  remarkRows?: number
  className?: string
  onLoadingDetailChange?: (loading: boolean) => void
}

export function CustomerForm({
  formId,
  onSave,
  itemId,
  fieldIdPrefix = "customer",
  remarkRows = 4,
  className,
  onLoadingDetailChange,
}: CustomerFormProps) {
  const [companyOptions, setCompanyOptions] = useState<ComboboxOption[]>(defaultCompanyOptions)
  const [loadingDetail, setLoadingDetail] = useState(() => Boolean(itemId))
  const loadedContactsRef = useRef<AdminCustomerContact[] | null>(null)

  const form = useForm<CustomerFormValues>({
    resolver: customerFormResolver,
    defaultValues: emptyCustomerForm,
  })
  const { handleSubmit, reset } = form
  const onLoadingDetailChangeRef = useRef(onLoadingDetailChange)
  onLoadingDetailChangeRef.current = onLoadingDetailChange

  useEffect(() => {
    let cancelled = false
    void (async () => {
      try {
        const data = await fetchCompanies({ ...CUSTOMER_FORM_COMPANY_LIST_QUERY })
        if (cancelled) {
          return
        }
        setCompanyOptions([
          ...defaultCompanyOptions,
          ...data.results.map((item) => ({
            value: String(item.id),
            label: item.name || `公司 #${item.id}`,
          })),
        ])
      } catch {
        if (!cancelled) {
          setCompanyOptions(defaultCompanyOptions)
        }
      }
    })()
    return () => {
      cancelled = true
    }
  }, [])

  useEffect(() => {
    async function loadDetail() {
      const notify = (loading: boolean) => {
        onLoadingDetailChangeRef.current?.(loading)
      }
      if (!itemId) {
        setLoadingDetail(false)
        notify(false)
        loadedContactsRef.current = null
        reset(emptyCustomerForm)
        return
      }
      setLoadingDetail(true)
      notify(true)
      try {
        const [customer, contacts] = await Promise.all([
          fetchCustomer(itemId),
          fetchCustomerContacts(itemId),
        ])
        loadedContactsRef.current = contacts
        reset({
          ...buildCustomerMainFromAdmin(customer),
          contacts: buildContactsFromApi(contacts),
        })
      } finally {
        setLoadingDetail(false)
        notify(false)
      }
    }
    void loadDetail()
  }, [itemId, reset])

  async function onFormSubmit(values: CustomerFormValues) {
    const contacts = normalizeContactsForSubmit(values.contacts as CustomerContactFormRow[])
    const customerPayload = buildCustomerPayload(values, contacts)
    await onSave({
      customerPayload,
      customerIdForUpdate: itemId ?? null,
      contacts,
      previousContactRecords: loadedContactsRef.current,
    })
  }

  if (loadingDetail) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="text-muted-foreground">加载中...</div>
      </div>
    )
  }

  return (
    <form id={formId} onSubmit={handleSubmit(onFormSubmit)} className={className}>
      <CustomerFormFields
        form={form}
        companyOptions={companyOptions}
        fieldIdPrefix={fieldIdPrefix}
        remarkRows={remarkRows}
      />
    </form>
  )
}
