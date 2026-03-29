"use client"

import { useState } from "react"

import { CustomerForm } from "@/components/customer-form"
import { ProjectDialog } from "@/components/project-dialog"
import { Button } from "@/components/ui/button"
import type { CreateAdminCustomerPayload } from "@/lib/api/customer"

export type CustomerFormDialogProps = {
  open: boolean
  saving: boolean
  itemId: number | null
  onOpenChange: (open: boolean) => void
  onSubmit: (payload: CreateAdminCustomerPayload) => Promise<void>
}

/** 客户新建/编辑表单弹窗（ProjectDialog + CustomerForm），供客户管理页与会话工作台等复用。 */
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
      saving={saving}
      itemId={itemId}
      onOpenChange={onOpenChange}
      onSubmit={onSubmit}
    />
  )
}

type CustomerFormDialogBodyProps = Omit<CustomerFormDialogProps, "open">

function CustomerFormDialogBody({
  saving,
  itemId,
  onOpenChange,
  onSubmit,
}: CustomerFormDialogBodyProps) {
  const formId = "customer-form-dialog"
  const [loadingDetail, setLoadingDetail] = useState(() => Boolean(itemId))

  return (
    <ProjectDialog
      open
      onOpenChange={(next) => onOpenChange(next)}
      title={itemId ? "编辑客户" : "新建客户"}
      allowFullscreen
      size="xl"
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
          <Button
            type="submit"
            form={formId}
            disabled={saving || loadingDetail}
          >
            {saving ? "保存中..." : itemId ? "保存" : "创建"}
          </Button>
        </>
      }
    >
      <CustomerForm
        formId={formId}
        itemId={itemId}
        onSubmit={onSubmit}
        fieldIdPrefix="customer"
        className="space-y-4"
        onLoadingDetailChange={setLoadingDetail}
      />
    </ProjectDialog>
  );
}
