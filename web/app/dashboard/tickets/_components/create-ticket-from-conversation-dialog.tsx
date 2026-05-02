"use client"

import { toast } from "sonner"

import { createTicketFromConversation } from "@/lib/api/ticket"
import { EditDialog } from "./edit"

type ConversationSeed = {
  id: number
  customerName: string
  customerId?: number
  lastMessageSummary?: string
  currentAssigneeId?: number
}

type CreateTicketFromConversationDialogProps = {
  open: boolean
  conversation: ConversationSeed | null
  onOpenChange: (open: boolean) => void
  onSuccess?: () => void
}

export function CreateTicketFromConversationDialog({
  open,
  conversation,
  onOpenChange,
  onSuccess,
}: CreateTicketFromConversationDialogProps) {
  const initialValues = conversation
    ? {
        title: conversation.customerName || "",
        description: conversation.lastMessageSummary || "",
        currentAssigneeId: conversation.currentAssigneeId || undefined,
      }
    : undefined

  return (
    <EditDialog
      open={open}
      saving={false}
      itemId={null}
      onOpenChange={onOpenChange}
      fixedConversationId={conversation?.id}
      fixedCustomerId={conversation?.customerId}
      initialValues={initialValues}
      titleOverride="会话转工单"
      descriptionOverride="从当前会话上下文创建正式工单"
      onSubmit={async (payload) => {
        if (!conversation?.id) {
          throw new Error("会话不存在")
        }
        await createTicketFromConversation({
          conversationId: conversation.id,
          title: payload.title,
          description: payload.description,
          currentAssigneeId: payload.currentAssigneeId,
          tagIds: payload.tagIds,
        })
        toast.success("工单创建成功")
        onSuccess?.()
      }}
    />
  )
}
