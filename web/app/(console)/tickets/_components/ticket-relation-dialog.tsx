"use client"

import { useEffect } from "react"
import { zodResolver } from "@hookform/resolvers/zod"
import { Resolver, useForm } from "react-hook-form"
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
import { Input } from "@/components/ui/input"
import { addTicketRelation } from "@/lib/api/ticket"

const relationOptions = [
  { value: "duplicate", label: "重复工单" },
  { value: "related", label: "相关工单" },
  { value: "parent", label: "父工单" },
  { value: "child", label: "子工单" },
]

const schema = z.object({
  relationType: z.string().trim().min(1, "请选择关联类型"),
  relatedTicketNo: z.string().trim().min(1, "请输入关联工单号"),
})

type FormValues = z.infer<typeof schema>

const resolver = zodResolver(schema as never) as Resolver<
  z.input<typeof schema>,
  undefined,
  z.output<typeof schema>
>

type TicketRelationDialogProps = {
  open: boolean
  ticketId: number | null
  onOpenChange: (open: boolean) => void
  onSuccess?: () => Promise<void> | void
}

export function TicketRelationDialog({
  open,
  ticketId,
  onOpenChange,
  onSuccess,
}: TicketRelationDialogProps) {
  const form = useForm<
    z.input<typeof schema>,
    undefined,
    z.output<typeof schema>
  >({
    resolver,
    defaultValues: {
      relationType: "related",
      relatedTicketNo: "",
    },
  })

  const {
    handleSubmit,
    setValue,
    register,
    reset,
    watch,
    formState: { errors, isSubmitting },
  } = form

  useEffect(() => {
    if (open) {
      reset({ relationType: "related", relatedTicketNo: "" })
    }
  }, [open, reset])

  async function onFormSubmit(values: FormValues) {
    if (!ticketId) {
      toast.error("工单不存在")
      return
    }
    try {
      await addTicketRelation({
        ticketId,
        relationType: values.relationType,
        relatedTicketNo: values.relatedTicketNo.trim(),
      })
      toast.success("关联工单已添加")
      onOpenChange(false)
      await onSuccess?.()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "添加关联工单失败")
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg gap-0 p-0 sm:max-w-lg">
        <DialogHeader className="px-6 pt-6">
          <DialogTitle>新增关联工单</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit(onFormSubmit)}>
          <div className="space-y-4 p-6">
            <Field data-invalid={!!errors.relationType}>
              <FieldLabel>关联类型</FieldLabel>
              <FieldContent>
                <OptionCombobox
                  value={watch("relationType")}
                  options={relationOptions}
                  placeholder="请选择关联类型"
                  onChange={(value) => setValue("relationType", value, { shouldValidate: true })}
                />
                <FieldError errors={[errors.relationType]} />
              </FieldContent>
            </Field>
            <Field data-invalid={!!errors.relatedTicketNo}>
              <FieldLabel>关联工单号</FieldLabel>
              <FieldContent>
                <Input placeholder="例如 TK20260403120000123" {...register("relatedTicketNo")} />
                <FieldError errors={[errors.relatedTicketNo]} />
              </FieldContent>
            </Field>
          </div>
          <DialogFooter className="mx-0 mb-0 px-6 py-4">
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              取消
            </Button>
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting ? "提交中..." : "确认添加"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
