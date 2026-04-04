"use client"

import { useEffect, useState } from "react"
import { Controller, Resolver, useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
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
import { fetchAgentProfilesAll, type AdminAgentProfile } from "@/lib/api/admin"
import { addTicketCollaborator } from "@/lib/api/ticket"

const schema = z.object({
  userId: z.string().trim().min(1, "请选择协作人"),
})

type FormValues = z.infer<typeof schema>

const resolver = zodResolver(schema as never) as Resolver<
  z.input<typeof schema>,
  undefined,
  z.output<typeof schema>
>

type TicketCollaboratorDialogProps = {
  open: boolean
  ticketId: number | null
  onOpenChange: (open: boolean) => void
  onSuccess?: () => Promise<void> | void
}

export function TicketCollaboratorDialog({
  open,
  ticketId,
  onOpenChange,
  onSuccess,
}: TicketCollaboratorDialogProps) {
  const [loadingAgents, setLoadingAgents] = useState(false)
  const [agents, setAgents] = useState<AdminAgentProfile[]>([])
  const userOptions = agents.map((agent) => ({
    value: String(agent.userId),
    label: agent.displayName || agent.nickname || agent.username || `客服 #${agent.userId}`,
  }))

  const form = useForm<
    z.input<typeof schema>,
    undefined,
    z.output<typeof schema>
  >({
    resolver,
    defaultValues: { userId: "" },
  })
  const {
    control,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
  } = form

  useEffect(() => {
    if (open) {
      reset({ userId: "" })
    }
  }, [open, reset])

  useEffect(() => {
    if (!open) {
      return
    }
    setLoadingAgents(true)
    fetchAgentProfilesAll()
      .then((data) => {
        setAgents(Array.isArray(data) ? data : [])
      })
      .catch((error) => {
        toast.error(error instanceof Error ? error.message : "加载客服列表失败")
      })
      .finally(() => {
        setLoadingAgents(false)
      })
  }, [open])

  async function onFormSubmit(values: FormValues) {
    if (!ticketId) {
      toast.error("工单不存在")
      return
    }
    try {
      await addTicketCollaborator({ ticketId, userId: Number(values.userId) })
      toast.success("协作人已添加")
      onOpenChange(false)
      await onSuccess?.()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "添加协作人失败")
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg gap-0 p-0 sm:max-w-lg">
        <DialogHeader className="px-6 pt-6">
          <DialogTitle>新增协作人</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit(onFormSubmit)}>
          <div className="space-y-4 p-6">
            <Field data-invalid={!!errors.userId}>
              <FieldLabel>协作人</FieldLabel>
              <FieldContent>
                <Controller
                  control={control}
                  name="userId"
                  render={({ field }) => (
                    <OptionCombobox
                      value={field.value}
                      options={userOptions}
                      placeholder={loadingAgents ? "加载中..." : "选择协作人"}
                      searchPlaceholder="搜索客服"
                      emptyText="暂无可选客服"
                      disabled={isSubmitting || loadingAgents}
                      onChange={field.onChange}
                    />
                  )}
                />
                <FieldError errors={[errors.userId]} />
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
