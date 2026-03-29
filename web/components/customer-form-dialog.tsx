"use client"

import { useEffect, useState } from "react"
import { useForm } from "react-hook-form"

import {
  buildCustomerFormFromAdmin,
  customerFormResolver,
  CustomerFormFields,
  customerFormToPayload,
  emptyCustomerForm,
  type CustomerFormValues,
} from "@/components/customer-form-fields"
import { ProjectDialog } from "@/components/project-dialog"
import type { ComboboxOption } from "@/components/option-combobox"
import { Button } from "@/components/ui/button"
import { fetchCompanies } from "@/lib/api/company"
import { fetchCustomer, type CreateAdminCustomerPayload } from "@/lib/api/customer"

export type CustomerFormDialogProps = {
  open: boolean
  saving: boolean
  itemId: number | null
  onOpenChange: (open: boolean) => void
  onSubmit: (payload: CreateAdminCustomerPayload) => Promise<void>
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
  const form = useForm<CustomerFormValues>({
    resolver: customerFormResolver,
    defaultValues: emptyCustomerForm,
  })
  const { handleSubmit, reset } = form

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
        reset(emptyCustomerForm)
        return
      }
      setLoading(true)
      try {
        const data = await fetchCustomer(itemId)
        reset(buildCustomerFormFromAdmin(data))
      } finally {
        setLoading(false)
      }
    }
    void loadDetail()
  }, [itemId, reset])

  async function onFormSubmit(values: CustomerFormValues) {
    await onSubmit(customerFormToPayload(values))
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
          <CustomerFormFields form={form} companyOptions={companyOptions} fieldIdPrefix="customer" />
        </form>
      )}
    </ProjectDialog>
  )
}
