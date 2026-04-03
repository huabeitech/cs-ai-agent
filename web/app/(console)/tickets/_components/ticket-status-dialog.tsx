"use client"

import { useEffect } from "react"
import { zodResolver } from "@hookform/resolvers/zod"
import { Controller, Resolver, useForm } from "react-hook-form"
import { toast } from "sonner"
import { z } from "zod/v4"

import { OptionCombobox } from "@/components/option-combobox"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Field, FieldContent, FieldError, FieldLabel } from "@/components/ui/field"
import { Textarea } from "@/components/ui/textarea"
import { changeTicketStatus } from "@/lib/api/ticket"

const schema = z.object({
  status: z.string().trim().min(1, "请选择状态"),
  pendingReason: z.string().trim(),
  closeReason: z.string().trim(),
  resolutionCode: z.string().trim(),
  resolutionSummary: z.string().trim(),
  reason: z.string().trim(),
})

type FormValues = z.infer<typeof schema>

const resolver = zodResolver(schema as never) as Resolver<
  z.input<typeof schema>,
  undefined,
  z.output<typeof schema>
>

type TicketStatusDialogProps = {
  open: boolean
  ticketId: number | null
  currentStatus?: string
  onOpenChange: (open: boolean) => void
  onSuccess?: () => Promise<void> | void
}

export function TicketStatusDialog({
  open,
  ticketId,
  currentStatus,
  onOpenChange,
  onSuccess,
}: TicketStatusDialogProps) {
  const form = useForm<
    z.input<typeof schema>,
    undefined,
    z.output<typeof schema>
  >({
    resolver,
    defaultValues: {
      status: "",
      pendingReason: "",
      closeReason: "",
      resolutionCode: "",
      resolutionSummary: "",
      reason: "",
    },
  })

  const {
    control,
    watch,
    register,
    reset,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = form

  const targetStatus = watch("status")

  useEffect(() => {
    reset({
      status: currentStatus || "",
      pendingReason: "",
      closeReason: "",
      resolutionCode: "",
      resolutionSummary: "",
      reason: "",
    })
  }, [currentStatus, reset, ticketId])

  async function onFormSubmit(values: FormValues) {
    if (!ticketId) {
      toast.error("工单不存在")
      return
    }
    try {
      await changeTicketStatus({
        ticketId,
        status: values.status,
        pendingReason: values.pendingReason || undefined,
        closeReason: values.closeReason || undefined,
        resolutionCode: values.resolutionCode || undefined,
        resolutionSummary: values.resolutionSummary || undefined,
        reason: values.reason || undefined,
      })
      toast.success("状态已更新")
      onOpenChange(false)
      await onSuccess?.()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "更新状态失败")
    }
  }

  const statusOptions = buildAllowedStatusOptions(currentStatus)

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg gap-0 p-0 sm:max-w-lg">
        <DialogHeader className="px-6 pt-6">
          <DialogTitle>变更工单状态</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit(onFormSubmit)}>
          <div className="space-y-4 p-6">
            <Field data-invalid={!!errors.status}>
              <FieldLabel>目标状态</FieldLabel>
              <FieldContent>
                <Controller
                  control={control}
                  name="status"
                  render={({ field }) => (
                    <OptionCombobox
                      value={field.value}
                      onChange={field.onChange}
                      placeholder="请选择状态"
                      options={statusOptions}
                    />
                  )}
                />
                <FieldError errors={[errors.status]} />
              </FieldContent>
            </Field>

            {(targetStatus === "pending_customer" ||
              targetStatus === "pending_internal") && (
              <Field data-invalid={!!errors.pendingReason}>
                <FieldLabel>挂起原因</FieldLabel>
                <FieldContent>
                  <Textarea rows={3} placeholder="请输入待处理原因" {...register("pendingReason")} />
                  <FieldError errors={[errors.pendingReason]} />
                </FieldContent>
              </Field>
            )}

            {targetStatus === "closed" && (
              <Field data-invalid={!!errors.closeReason}>
                <FieldLabel>关闭原因</FieldLabel>
                <FieldContent>
                  <Textarea rows={3} placeholder="请输入关闭原因" {...register("closeReason")} />
                  <FieldError errors={[errors.closeReason]} />
                </FieldContent>
              </Field>
            )}

            {targetStatus === "resolved" && (
              <>
                <Field>
                  <FieldLabel>解决编码</FieldLabel>
                  <FieldContent>
                    <Textarea rows={2} placeholder="可选：填写解决编码" {...register("resolutionCode")} />
                  </FieldContent>
                </Field>
                <Field>
                  <FieldLabel>解决说明</FieldLabel>
                  <FieldContent>
                    <Textarea rows={3} placeholder="请输入解决说明" {...register("resolutionSummary")} />
                  </FieldContent>
                </Field>
              </>
            )}

            <Field>
              <FieldLabel>操作说明</FieldLabel>
              <FieldContent>
                <Textarea rows={3} placeholder="填写本次状态变更说明" {...register("reason")} />
              </FieldContent>
            </Field>
          </div>
          <DialogFooter className="mx-0 mb-0 px-6 py-4">
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              取消
            </Button>
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting ? "提交中..." : "确认变更"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

function buildAllowedStatusOptions(currentStatus?: string) {
  const statusLabels: Record<string, string> = {
    new: "新建",
    open: "处理中",
    pending_customer: "待客户反馈",
    pending_internal: "待内部处理",
    resolved: "已解决",
    closed: "已关闭",
    cancelled: "已取消",
  }
  const transitionMap: Record<string, string[]> = {
    new: ["open", "pending_internal", "cancelled"],
    open: ["pending_customer", "pending_internal", "resolved", "closed", "cancelled"],
    pending_customer: ["open", "resolved", "closed", "cancelled"],
    pending_internal: ["open", "resolved", "closed", "cancelled"],
    resolved: ["open", "closed"],
    closed: ["open"],
  }
  const candidates = currentStatus ? [currentStatus].concat(transitionMap[currentStatus] ?? []) : []
  const seen = new Set<string>()
  return candidates
    .filter((status) => {
      if (!statusLabels[status] || seen.has(status)) {
        return false
      }
      seen.add(status)
      return true
    })
    .map((status) => ({
      value: status,
      label: statusLabels[status],
    }))
}
