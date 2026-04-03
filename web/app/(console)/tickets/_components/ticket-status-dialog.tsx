"use client"

import Link from "next/link"
import { useEffect, useState } from "react"
import { zodResolver } from "@hookform/resolvers/zod"
import { Settings2Icon } from "lucide-react"
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
import {
  fetchTicketResolutionCodesAll,
  type TicketResolutionCode,
} from "@/lib/api/ticket-config"
import { batchChangeTicketStatus, changeTicketStatus } from "@/lib/api/ticket"

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
  ticketIds?: number[]
  currentStatus?: string
  onOpenChange: (open: boolean) => void
  onSuccess?: () => Promise<void> | void
}

export function TicketStatusDialog({
  open,
  ticketId,
  ticketIds,
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
  const [resolutionCodes, setResolutionCodes] = useState<TicketResolutionCode[]>([])

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

  useEffect(() => {
    if (!open) {
      return
    }
    void (async () => {
      try {
        const data = await fetchTicketResolutionCodesAll()
        setResolutionCodes(Array.isArray(data) ? data : [])
      } catch (error) {
        toast.error(error instanceof Error ? error.message : "加载解决码失败")
      }
    })()
  }, [open])

  const resolutionCodeOptions = resolutionCodes.map((item) => ({
    value: item.code,
    label: item.name,
  }))

  async function onFormSubmit(values: FormValues) {
    const validTicketIds = (ticketIds ?? []).filter((item) => item > 0)
    if (!ticketId && validTicketIds.length === 0) {
      toast.error("请选择工单")
      return
    }
    try {
      if (validTicketIds.length > 0) {
        await batchChangeTicketStatus({
          ticketIds: validTicketIds,
          status: values.status,
          pendingReason: values.pendingReason || undefined,
          closeReason: values.status === "closed" ? values.closeReason || undefined : undefined,
          resolutionCode: values.resolutionCode || undefined,
          resolutionSummary: values.resolutionSummary || undefined,
          reason: values.reason || undefined,
        })
        toast.success(`已批量更新 ${validTicketIds.length} 张工单`)
      } else {
        await changeTicketStatus({
          ticketId: ticketId!,
          status: values.status,
          pendingReason: values.pendingReason || undefined,
          closeReason: values.status === "closed" ? values.closeReason || undefined : undefined,
          resolutionCode: values.resolutionCode || undefined,
          resolutionSummary: values.resolutionSummary || undefined,
          reason: values.reason || undefined,
        })
        toast.success("状态已更新")
      }
      onOpenChange(false)
      await onSuccess?.()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "更新状态失败")
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg gap-0 p-0 sm:max-w-lg">
        <DialogHeader className="px-6 pt-6">
          <DialogTitle>{ticketIds?.length ? `批量变更状态（${ticketIds.length}）` : "变更工单状态"}</DialogTitle>
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
                      options={[
                        { value: "new", label: "新建" },
                        { value: "open", label: "处理中" },
                        { value: "pending_customer", label: "待客户反馈" },
                        { value: "pending_internal", label: "待内部处理" },
                        { value: "resolved", label: "已解决" },
                        { value: "closed", label: "已关闭" },
                        { value: "cancelled", label: "已取消" },
                      ]}
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

            {targetStatus === "resolved" && (
              <>
                <Field>
                  <div className="flex items-center justify-between gap-3">
                    <FieldLabel>解决编码</FieldLabel>
                    <Link href="/ticket-resolution-codes" target="_blank" rel="noreferrer">
                      <Button variant="ghost" size="sm" type="button">
                        <Settings2Icon className="size-4" />
                        管理解决码
                      </Button>
                    </Link>
                  </div>
                  <FieldContent>
                    <Controller
                      control={control}
                      name="resolutionCode"
                      render={({ field }) => (
                        <OptionCombobox
                          value={field.value}
                          onChange={field.onChange}
                          placeholder="请选择解决编码"
                          options={resolutionCodeOptions}
                          emptyText="暂无可选解决码"
                        />
                      )}
                    />
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

            {targetStatus === "closed" && (
              <Field>
                <FieldLabel>关闭原因</FieldLabel>
                <FieldContent>
                  <Textarea rows={3} placeholder="请输入关闭原因" {...register("closeReason")} />
                </FieldContent>
              </Field>
            )}

            <Field>
              <FieldLabel>操作说明</FieldLabel>
              <FieldContent>
                <Textarea
                  rows={3}
                  placeholder={targetStatus === "closed" ? "可补充本次批量关闭说明" : "填写本次状态变更说明"}
                  {...register("reason")}
                />
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
