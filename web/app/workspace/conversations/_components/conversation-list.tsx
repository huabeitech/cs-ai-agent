"use client"

import { UserIcon } from "lucide-react"

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { ScrollArea } from "@/components/ui/scroll-area"
import { formatDateTime } from "@/lib/utils"
import { useAgentConversationsStore } from "@/lib/stores/agent-conversations"
import {
  IMConversationStatus,
  IMConversationStatusLabels,
} from "@/lib/generated/enums"
import { getEnumLabel } from "@/lib/enums"

function getStatusVariant(status: number) {
  switch (status) {
    case IMConversationStatus.Pending:
      return "bg-blue-100 text-blue-700"
    case IMConversationStatus.Active:
      return "bg-green-100 text-green-700"
    case IMConversationStatus.Closed:
      return "bg-gray-100 text-gray-700"
    default:
      return "bg-gray-100 text-gray-700"
  }
}

export function ConversationList() {
  const conversations = useAgentConversationsStore((state) => state.conversations)
  const loading = useAgentConversationsStore((state) => state.conversationsLoading)
  const selectedId = useAgentConversationsStore((state) => state.selectedConversationId)
  const selectConversation = useAgentConversationsStore((state) => state.selectConversation)

  return (
    <ScrollArea className="flex-1">
      <div className="divide-y">
        {loading ? (
          <div className="p-6 text-center text-sm text-muted-foreground">
            加载中...
          </div>
        ) : conversations.length > 0 ? (
          conversations.map((conversation) => {
            const isSelected = selectedId === conversation.id
            return (
              <div
                key={conversation.id}
                className={`cursor-pointer px-2.5 py-1.5 transition-colors hover:bg-muted/50 ${
                  isSelected ? "bg-muted/80" : ""
                }`}
                onClick={() => void selectConversation(conversation.id)}
              >
                <div className="overflow-hidden">
                  <div className="flex items-center gap-2">
                    <Avatar className="size-7 shrink-0">
                      <AvatarImage src="" />
                      <AvatarFallback className="bg-primary/10">
                        <UserIcon className="size-3.5 text-primary" />
                      </AvatarFallback>
                    </Avatar>
                    <div className="min-w-0 flex-1">
                      <div className="flex items-center gap-1.5">
                        <span className="min-w-0 flex-1 truncate font-medium text-sm leading-4">
                          {conversation.subject}
                        </span>
                        {conversation.agentUnreadCount > 0 ? (
                          <div className="flex size-4.5 shrink-0 items-center justify-center rounded-full bg-primary text-[10px] text-primary-foreground">
                            {conversation.agentUnreadCount > 99
                              ? "99+"
                              : conversation.agentUnreadCount}
                          </div>
                        ) : null}
                      </div>
                      <div className="mt-0.5 text-[11px] text-muted-foreground">
                        {conversation.lastMessageAt
                          ? formatDateTime(conversation.lastMessageAt)
                          : "暂无时间"}
                      </div>
                    </div>
                  </div>
                  <div className="mt-0.5 truncate text-xs leading-4 text-muted-foreground">
                    {conversation.lastMessageSummary || "暂无最新消息"}
                  </div>
                  <div className="mt-1 flex items-center gap-1 text-[10px] text-muted-foreground">
                    <span
                      className={`rounded px-1 py-0.5 ${getStatusVariant(
                        conversation.status
                      )}`}
                    >
                      {getEnumLabel(IMConversationStatusLabels, conversation.status)}
                    </span>
                    {conversation.externalSource ? (
                      <>
                        <span className="opacity-40">·</span>
                        <span className="truncate">{conversation.externalSource}</span>
                      </>
                    ) : null}
                  </div>
                </div>
              </div>
            )
          })
        ) : (
          <div className="p-6 text-center text-sm text-muted-foreground">
            暂无会话
          </div>
        )}
      </div>
    </ScrollArea>
  )
}
