"use client"

import { useEffect, useRef, useState } from "react"
import { useForm } from "react-hook-form"

import {
  buildCustomerFormFromAdmin,
  customerFormResolver,
  CustomerFormFields,
  customerFormToPayload,
  emptyCustomerForm,
  type CustomerFormValues,
} from "@/components/customer-form-fields"
import type { ComboboxOption } from "@/components/option-combobox"
import { fetchCompanies } from "@/lib/api/company"
import { fetchCustomer, type CreateAdminCustomerPayload } from "@/lib/api/customer"

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
  /** 与外部提交按钮 `form` 属性对应 */
  formId: string
  onSubmit: (payload: CreateAdminCustomerPayload) => Promise<void> | void
  /** 编辑时传入客户 ID；新建不传或传 null */
  itemId?: number | null
  fieldIdPrefix?: string
  remarkRows?: number
  className?: string
  /** 编辑拉取详情时的加载状态，用于禁用外层提交按钮等 */
  onLoadingDetailChange?: (loading: boolean) => void
}

/**
 * 客户档案完整表单：react-hook-form、校验、公司列表加载、编辑时拉取详情。
 * 弹窗或页面仅包一层壳子并挂提交按钮即可。
 */
export function CustomerForm({
  formId,
  onSubmit,
  itemId,
  fieldIdPrefix = "customer",
  remarkRows = 4,
  className,
  onLoadingDetailChange,
}: CustomerFormProps) {
  const [companyOptions, setCompanyOptions] = useState<ComboboxOption[]>(defaultCompanyOptions)
  /** 有 itemId 时首帧即进入加载，避免先闪空表单 */
  const [loadingDetail, setLoadingDetail] = useState(() => Boolean(itemId))

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
        reset(emptyCustomerForm)
        return
      }
      setLoadingDetail(true)
      notify(true)
      try {
        const data = await fetchCustomer(itemId)
        reset(buildCustomerFormFromAdmin(data))
      } finally {
        setLoadingDetail(false)
        notify(false)
      }
    }
    void loadDetail()
  }, [itemId, reset])

  async function onFormSubmit(values: CustomerFormValues) {
    await onSubmit(customerFormToPayload(values))
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
