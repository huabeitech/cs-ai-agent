"use client"

import { useMemo } from "react"
import { zodResolver } from "@hookform/resolvers/zod"
import { Controller, type Resolver, type UseFormReturn } from "react-hook-form"
import { z } from "zod/v4"

import { OptionCombobox, type ComboboxOption } from "@/components/option-combobox"
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
import type { AdminCustomer, CreateAdminCustomerPayload } from "@/lib/api/customer"
import { getEnumLabel, getEnumOptions } from "@/lib/enums"
import { Gender, GenderLabels } from "@/lib/generated/enums"

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

export const customerFormSchema = z.object({
  name: z.string().trim().min(1, "客户名称不能为空"),
  gender: z.enum(genderValueOptions, { message: "请选择性别" }),
  companyId: z.string().trim().regex(/^\d+$/, "请选择所属公司"),
  primaryMobile: z.string().trim(),
  primaryEmail: z.string().trim(),
  remark: z.string().trim(),
})

export type CustomerFormValues = z.infer<typeof customerFormSchema>

export const customerFormResolver = zodResolver(customerFormSchema as never) as Resolver<
  z.input<typeof customerFormSchema>,
  undefined,
  z.output<typeof customerFormSchema>
>

export const emptyCustomerForm: CustomerFormValues = {
  name: "",
  gender: "0",
  companyId: "0",
  primaryMobile: "",
  primaryEmail: "",
  remark: "",
}

export function buildCustomerFormFromAdmin(item: AdminCustomer | null): CustomerFormValues {
  if (!item) {
    return emptyCustomerForm
  }
  return {
    name: item.name,
    gender: String(item.gender) as "0" | "1" | "2",
    companyId: String(item.companyId ?? 0),
    primaryMobile: item.primaryMobile ?? "",
    primaryEmail: item.primaryEmail ?? "",
    remark: item.remark ?? "",
  }
}

export function customerFormToPayload(values: CustomerFormValues): CreateAdminCustomerPayload {
  return {
    name: values.name.trim(),
    gender: Number(values.gender),
    companyId: Number(values.companyId),
    primaryMobile: values.primaryMobile.trim(),
    primaryEmail: values.primaryEmail.trim(),
    remark: values.remark.trim(),
  }
}

function getGenderLabel(value: string) {
  return getEnumLabel(GenderLabels, Number(value) as Gender)
}

export type CustomerFormFieldsProps = {
  form: UseFormReturn<CustomerFormValues>
  companyOptions: ComboboxOption[]
  /** 用于 input/label 的 id 前缀，避免同页多个表单冲突 */
  fieldIdPrefix?: string
  /** 备注行数 */
  remarkRows?: number
}

/** 客户档案表单字段（与 CustomerFormDialog 内联校验一致），供弹窗或嵌入式区域复用。 */
export function CustomerFormFields({
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
  } = form
  const companyId = watch("companyId")
  const selectedCompanyLabel = useMemo(() => {
    return companyOptions.find((item) => item.value === companyId)?.label ?? "请选择公司"
  }, [companyOptions, companyId])

  const id = (suffix: string) => `${fieldIdPrefix}-${suffix}`

  return (
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

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        <Field data-invalid={!!errors.primaryMobile}>
          <FieldLabel htmlFor={id("mobile")}>手机号</FieldLabel>
          <FieldContent>
            <Input
              id={id("mobile")}
              placeholder="可选"
              aria-invalid={!!errors.primaryMobile}
              {...register("primaryMobile")}
            />
            <FieldError errors={[errors.primaryMobile]} />
          </FieldContent>
        </Field>
        <Field data-invalid={!!errors.primaryEmail}>
          <FieldLabel htmlFor={id("email")}>邮箱</FieldLabel>
          <FieldContent>
            <Input
              id={id("email")}
              placeholder="可选"
              aria-invalid={!!errors.primaryEmail}
              {...register("primaryEmail")}
            />
            <FieldError errors={[errors.primaryEmail]} />
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
  )
}
