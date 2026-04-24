"use client"

import { useEffect, useState } from "react"
import { zodResolver } from "@hookform/resolvers/zod"
import { Resolver, useForm } from "react-hook-form"
import { z } from "zod/v4"
import { CircleXIcon } from "lucide-react"
import { toast } from "sonner"

import { closeConversation } from "@/lib/api/admin"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import {
  Field,
  FieldContent,
  FieldError,
  FieldLabel,
} from "@/components/ui/field"
import { Textarea } from "@/components/ui/textarea"

type ConversationCloseDialogProps = {
  open: boolean
  conversationId: number | null
  onOpenChange: (open: boolean) => void
  onSuccess?: () => Promise<void> | void
}

const closeSchema = z.object({
  closeReason: z.string().trim().min(1, "请输入关闭原因"),
})

type CloseForm = z.infer<typeof closeSchema>

const closeResolver = zodResolver(closeSchema as never) as Resolver<
  z.input<typeof closeSchema>,
  undefined,
  z.output<typeof closeSchema>
>

const emptyForm: CloseForm = {
  closeReason: "",
}

export function ConversationCloseDialog({
  open,
  conversationId,
  onOpenChange,
  onSuccess,
}: ConversationCloseDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      {open ? (
        <ConversationCloseDialogBody
          key={conversationId ? `close-${conversationId}` : "close"}
          conversationId={conversationId}
          onOpenChange={onOpenChange}
          onSuccess={onSuccess}
        />
      ) : null}
    </Dialog>
  )
}

type ConversationCloseDialogBodyProps = {
  conversationId: number | null
  onOpenChange: (open: boolean) => void
  onSuccess?: () => Promise<void> | void
}

function ConversationCloseDialogBody({
  conversationId,
  onOpenChange,
  onSuccess,
}: ConversationCloseDialogBodyProps) {
  const [saving, setSaving] = useState(false)
  const form = useForm<
    z.input<typeof closeSchema>,
    undefined,
    z.output<typeof closeSchema>
  >({
    resolver: closeResolver,
    defaultValues: emptyForm,
  })
  const {
    handleSubmit,
    reset,
    register,
    formState: { errors },
  } = form

  useEffect(() => {
    reset(emptyForm)
  }, [conversationId, reset])

  async function onFormSubmit(values: CloseForm) {
    if (!conversationId) {
      toast.error("会话不存在")
      return
    }

    setSaving(true)
    try {
      await closeConversation(conversationId, values.closeReason.trim())
      toast.success(`已关闭会话：#${conversationId}`)
      reset(emptyForm)
      onOpenChange(false)
      await onSuccess?.()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "关闭会话失败")
    } finally {
      setSaving(false)
    }
  }

  return (
    <DialogContent className="max-w-lg gap-0 p-0 sm:max-w-lg">
      <DialogHeader className="px-6 pt-6">
        <DialogTitle>关闭会话</DialogTitle>
        {/* <DialogDescription>
          当前会话：{conversationId ? `#${conversationId}` : "-"}
        </DialogDescription> */}
      </DialogHeader>
      <form onSubmit={handleSubmit(onFormSubmit)}>
        <div className="space-y-4 p-6">
          <Field data-invalid={!!errors.closeReason}>
            <FieldLabel htmlFor="conversation-close-reason">关闭原因</FieldLabel>
            <FieldContent>
              <Textarea
                id="conversation-close-reason"
                rows={4}
                placeholder="填写关闭原因，关闭后会写入操作记录"
                aria-invalid={!!errors.closeReason}
                {...register("closeReason")}
              />
              <FieldError errors={[errors.closeReason]} />
            </FieldContent>
          </Field>
        </div>
        <DialogFooter className="mx-0 mb-0 px-6 py-4">
          <Button
            type="button"
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={saving}
          >
            取消
          </Button>
          <Button type="submit" disabled={saving}>
            <CircleXIcon />
            {saving ? "关闭中..." : "确认关闭"}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  )
}
