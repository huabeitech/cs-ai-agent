"use client"

import { toast } from "sonner"

import { createTicketFromConversation } from "@/lib/api/ticket"
import { TicketEditDialog } from "./ticket-edit-dialog"

type ConversationSeed = {
  id: number
  subject: string
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
  return (
    <TicketEditDialog
      open={open}
      onOpenChange={onOpenChange}
      fixedConversationId={conversation?.id}
      fixedCustomerId={conversation?.customerId}
      item={
        conversation
          ? {
              id: 0,
              ticketNo: "",
              title: conversation.subject || "",
              description: conversation.lastMessageSummary || "",
              source: "conversation",
              channel: "",
              customerId: conversation.customerId || 0,
              conversationId: conversation.id,
              categoryId: 0,
              type: "",
              priority: 2,
              severity: 1,
              status: "new",
              currentTeamId: 0,
              currentAssigneeId: conversation.currentAssigneeId || 0,
              watchedByMe: false,
              reopenedCount: 0,
            }
          : null
      }
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
          priority: payload.priority,
          severity: payload.severity,
          currentTeamId: payload.currentTeamId,
          currentAssigneeId: payload.currentAssigneeId,
          syncToConversation: true,
        })
        toast.success("工单创建成功")
        onSuccess?.()
      }}
    />
  )
}
