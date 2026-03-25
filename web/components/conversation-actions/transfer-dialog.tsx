"use client"

import { zodResolver } from "@hookform/resolvers/zod"
import { ArrowRightLeftIcon } from "lucide-react"
import { useEffect, useState } from "react"
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
  DialogTitle
} from "@/components/ui/dialog"
import {
  Field,
  FieldContent,
  FieldError,
  FieldLabel,
} from "@/components/ui/field"
import { Textarea } from "@/components/ui/textarea"
import {
  assignConversation,
  transferConversation,
  fetchAgentProfilesAll,
  type AdminAgentProfile,
} from "@/lib/api/admin"

type ConversationTransferDialogProps = {
  open: boolean
  mode: "assign" | "transfer"
  conversationId: number | null
  onOpenChange: (open: boolean) => void
  onSuccess?: () => Promise<void> | void
}

const transferSchema = z.object({
  toUserId: z.string().trim().min(1, "请选择目标客服"),
  reason: z.string().trim(),
})

type TransferForm = z.infer<typeof transferSchema>

const emptyForm: TransferForm = {
  toUserId: "",
  reason: "",
}

const transferResolver = zodResolver(transferSchema as never) as Resolver<
  z.input<typeof transferSchema>,
  undefined,
  z.output<typeof transferSchema>
>

export function ConversationTransferDialog({
  open,
  mode,
  conversationId,
  onOpenChange,
  onSuccess,
}: ConversationTransferDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      {open ? (
        <ConversationTransferDialogBody
          key={conversationId ? `transfer-${conversationId}` : "transfer"}
          mode={mode}
          conversationId={conversationId}
          onOpenChange={onOpenChange}
          onSuccess={onSuccess}
        />
      ) : null}
    </Dialog>
  )
}

type ConversationTransferDialogBodyProps = {
  mode: "assign" | "transfer"
  conversationId: number | null
  onOpenChange: (open: boolean) => void
  onSuccess?: () => Promise<void> | void
}

function ConversationTransferDialogBody({
  mode,
  conversationId,
  onOpenChange,
  onSuccess,
}: ConversationTransferDialogBodyProps) {
  const [saving, setSaving] = useState(false)
  const [loadingAgents, setLoadingAgents] = useState(false)
  const [agents, setAgents] = useState<AdminAgentProfile[]>([])
  const userOptions = agents.map((agent) => ({
    value: String(agent.userId),
    label: agent.displayName || agent.nickname || agent.username || `客服 #${agent.userId}`,
  }))

  const form = useForm<
    z.input<typeof transferSchema>,
    undefined,
    z.output<typeof transferSchema>
  >({
    resolver: transferResolver,
    defaultValues: emptyForm,
  })
  const {
    control,
    handleSubmit,
    reset,
    register,
    formState: { errors },
  } = form

  useEffect(() => {
    reset(emptyForm)
  }, [conversationId, reset])

  useEffect(() => {
    setLoadingAgents(true)
    fetchAgentProfilesAll()
      .then((data) => {
        setAgents(data.filter((item) => item.serviceStatus === 0))
      })
      .catch((error) => {
        toast.error(error instanceof Error ? error.message : "加载客服列表失败")
      })
      .finally(() => {
        setLoadingAgents(false)
      })
  }, [])

  async function onFormSubmit(values: TransferForm) {
    if (!conversationId) {
      toast.error("会话不存在")
      return
    }

    const toUserId = Number(values.toUserId)
    const reason = values.reason.trim()

    setSaving(true)
    try {
      if (mode === "assign") {
        await assignConversation(conversationId, toUserId, reason)
        toast.success(`已分配会话：#${conversationId}`)
      } else {
        await transferConversation(conversationId, toUserId, reason)
        toast.success(`已转接会话：#${conversationId}`)
      }
      reset(emptyForm)
      onOpenChange(false)
      await onSuccess?.()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : mode === "assign" ? "分配会话失败" : "转接会话失败")
    } finally {
      setSaving(false)
    }
  }

  const isAssign = mode === "assign"

  return (
    <DialogContent className="max-w-lg gap-0 p-0 sm:max-w-lg">
      <DialogHeader className="px-6 pt-6">
        <DialogTitle>{isAssign ? "分配会话" : "转接会话"}</DialogTitle>
        {/* <DialogDescription>
          当前会话：{conversationId ? `#${conversationId}` : "-"}
        </DialogDescription> */}
      </DialogHeader>
      <form onSubmit={handleSubmit(onFormSubmit)}>
        <div className="space-y-4 p-6">
          <Field data-invalid={!!errors.toUserId}>
            <FieldLabel htmlFor="conversation-transfer-user">目标客服</FieldLabel>
            <FieldContent>
              <Controller
                control={control}
                name="toUserId"
                render={({ field }) => (
                  <OptionCombobox
                    value={field.value}
                    options={userOptions}
                    placeholder={loadingAgents ? "加载中..." : "选择目标客服"}
                    searchPlaceholder="搜索客服"
                    emptyText="暂无可选客服"
                    disabled={saving || loadingAgents}
                    onChange={field.onChange}
                  />
                )}
              />
              <FieldError errors={[errors.toUserId]} />
            </FieldContent>
          </Field>
          <Field data-invalid={!!errors.reason}>
            <FieldLabel htmlFor="conversation-transfer-reason">
              {isAssign ? "分配说明" : "转接原因"}
            </FieldLabel>
            <FieldContent>
              <Textarea
                id="conversation-transfer-reason"
                rows={4}
                placeholder={isAssign ? "填写分配说明，便于后续追踪" : "填写转接原因，便于后续追踪"}
                aria-invalid={!!errors.reason}
                {...register("reason")}
              />
              <FieldError errors={[errors.reason]} />
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
            <ArrowRightLeftIcon />
            {saving ? (isAssign ? "分配中..." : "转接中...") : isAssign ? "确认分配" : "确认转接"}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  )
}
