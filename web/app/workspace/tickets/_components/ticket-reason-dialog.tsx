"use client"

import { useEffect } from "react"
import { zodResolver } from "@hookform/resolvers/zod"
import { Resolver, useForm } from "react-hook-form"
import { toast } from "sonner"
import { z } from "zod/v4"

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
import { closeTicket, reopenTicket } from "@/lib/api/ticket"

const schema = z.object({
  reason: z.string().trim().min(1, "请输入原因"),
})

type FormValues = z.infer<typeof schema>

const resolver = zodResolver(schema as never) as Resolver<
  z.input<typeof schema>,
  undefined,
  z.output<typeof schema>
>

type TicketReasonDialogProps = {
  open: boolean
  mode: "close" | "reopen"
  ticketId: number | null
  defaultReason?: string
  onOpenChange: (open: boolean) => void
  onSuccess?: () => Promise<void> | void
}

export function TicketReasonDialog({
  open,
  mode,
  ticketId,
  defaultReason,
  onOpenChange,
  onSuccess,
}: TicketReasonDialogProps) {
  const form = useForm<
    z.input<typeof schema>,
    undefined,
    z.output<typeof schema>
  >({
    resolver,
    defaultValues: { reason: "" },
  })

  const {
    register,
    reset,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = form

  useEffect(() => {
    reset({ reason: defaultReason || "" })
  }, [defaultReason, reset, ticketId, open])

  async function onFormSubmit(values: FormValues) {
    if (!ticketId) {
      toast.error("工单不存在")
      return
    }
    try {
      if (mode === "close") {
        await closeTicket({ ticketId, closeReason: values.reason })
        toast.success("工单已关闭")
      } else {
        await reopenTicket({ ticketId, reason: values.reason })
        toast.success("工单已重开")
      }
      onOpenChange(false)
      await onSuccess?.()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : mode === "close" ? "关闭工单失败" : "重开工单失败")
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg gap-0 p-0 sm:max-w-lg">
        <DialogHeader className="px-6 pt-6">
          <DialogTitle>{mode === "close" ? "关闭工单" : "重开工单"}</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit(onFormSubmit)}>
          <div className="space-y-4 p-6">
            <Field data-invalid={!!errors.reason}>
              <FieldLabel>{mode === "close" ? "关闭原因" : "重开原因"}</FieldLabel>
              <FieldContent>
                <Textarea
                  rows={4}
                  placeholder={mode === "close" ? "请输入关闭原因" : "请输入重开原因"}
                  {...register("reason")}
                />
                <FieldError errors={[errors.reason]} />
              </FieldContent>
            </Field>
          </div>
          <DialogFooter className="mx-0 mb-0 px-6 py-4">
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              取消
            </Button>
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting ? "提交中..." : mode === "close" ? "确认关闭" : "确认重开"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
