"use client"

import { zodResolver } from "@hookform/resolvers/zod"
import { Controller, type Resolver, useForm } from "react-hook-form"
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
import { changeTicketStatus, type TicketStatus } from "@/lib/api/ticket"

const ticketStatuses = [
  { value: "pending", label: "待处理" },
  { value: "in_progress", label: "处理中" },
  { value: "done", label: "已处理" },
] satisfies Array<{ value: TicketStatus; label: string }>

const schema = z.object({
  status: z.enum(["pending", "in_progress", "done"], { message: "请选择状态" }),
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
      status: isTicketStatus(currentStatus) ? currentStatus : "pending",
    },
  })

  const {
    control,
    reset,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = form

  function handleOpenChange(nextOpen: boolean) {
    if (nextOpen) {
      reset({ status: isTicketStatus(currentStatus) ? currentStatus : "pending" })
    }
    onOpenChange(nextOpen)
  }

  async function onFormSubmit(values: FormValues) {
    const validTicketIds = (ticketIds ?? []).filter((item) => item > 0)
    if (!ticketId && validTicketIds.length === 0) {
      toast.error("请选择工单")
      return
    }
    try {
      if (validTicketIds.length > 0) {
        await Promise.all(
          validTicketIds.map((id) => changeTicketStatus({ ticketId: id, status: values.status })),
        )
        toast.success(`已批量更新 ${validTicketIds.length} 张工单`)
      } else {
        await changeTicketStatus({
          ticketId: ticketId!,
          status: values.status,
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
    <Dialog open={open} onOpenChange={handleOpenChange}>
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
                      options={ticketStatuses}
                    />
                  )}
                />
                <FieldError errors={[errors.status]} />
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

function isTicketStatus(status: string | undefined): status is TicketStatus {
  return status === "pending" || status === "in_progress" || status === "done"
}
