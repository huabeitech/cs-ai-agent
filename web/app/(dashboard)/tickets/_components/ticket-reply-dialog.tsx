"use client"

import { useEffect, useMemo, useState } from "react"
import { MessageSquarePlusIcon } from "lucide-react"
import { toast } from "sonner"

import { OptionCombobox } from "@/components/option-combobox"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Textarea } from "@/components/ui/textarea"
import { fetchAgentProfilesAll, type AdminAgentProfile } from "@/lib/api/admin"
import { addTicketInternalNote, replyTicket } from "@/lib/api/ticket"

type TicketReplyDialogProps = {
  open: boolean
  ticketId: number | null
  onOpenChange: (open: boolean) => void
  onSuccess?: () => Promise<void> | void
}

export function TicketReplyDialog({
  open,
  ticketId,
  onOpenChange,
  onSuccess,
}: TicketReplyDialogProps) {
  const [replyMode, setReplyMode] = useState<"public" | "internal">("public")
  const [replyContent, setReplyContent] = useState("")
  const [mentionUserId, setMentionUserId] = useState("")
  const [mentionedUsers, setMentionedUsers] = useState<AdminAgentProfile[]>([])
  const [agents, setAgents] = useState<AdminAgentProfile[]>([])
  const [loadingAgents, setLoadingAgents] = useState(false)
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    if (!open) {
      return
    }
    setReplyMode("public")
    setReplyContent("")
    setMentionUserId("")
    setMentionedUsers([])
  }, [open])

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

  const mentionOptions = useMemo(
    () =>
      agents.map((agent) => ({
        value: String(agent.userId),
        label:
          agent.displayName ||
          agent.nickname ||
          agent.username ||
          `客服 #${agent.userId}`,
      })),
    [agents],
  )

  function handleAddMentionUser() {
    const userId = Number(mentionUserId)
    if (!userId) {
      return
    }
    const user = agents.find((item) => item.userId === userId)
    if (!user) {
      return
    }
    setMentionedUsers((current) => {
      if (current.some((item) => item.userId === user.userId)) {
        return current
      }
      return [...current, user]
    })
    setMentionUserId("")
  }

  async function handleSubmit() {
    if (!ticketId) {
      toast.error("工单不存在")
      return
    }
    if (!replyContent.trim()) {
      toast.error(replyMode === "public" ? "回复内容不能为空" : "备注内容不能为空")
      return
    }

    setSubmitting(true)
    try {
      if (replyMode === "public") {
        await replyTicket({
          ticketId,
          contentType: "text",
          content: replyContent.trim(),
        })
        toast.success("已回复客户")
      } else {
        const payload =
          mentionedUsers.length > 0
            ? JSON.stringify({
                mentionUserIds: mentionedUsers.map((item) => item.userId),
              })
            : undefined
        await addTicketInternalNote({
          ticketId,
          contentType: "text",
          content: replyContent.trim(),
          payload,
        })
        toast.success("已添加内部备注")
      }

      onOpenChange(false)
      await onSuccess?.()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "提交失败")
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl gap-0 p-0 sm:max-w-2xl">
        <DialogHeader className="px-6 pt-6">
          <DialogTitle>回复与备注</DialogTitle>
        </DialogHeader>
        <div className="space-y-4 p-6">
          <div className="flex gap-2">
            <Button
              variant={replyMode === "public" ? "default" : "outline"}
              onClick={() => setReplyMode("public")}
              disabled={submitting}
            >
              回复客户
            </Button>
            <Button
              variant={replyMode === "internal" ? "default" : "outline"}
              onClick={() => setReplyMode("internal")}
              disabled={submitting}
            >
              内部备注
            </Button>
          </div>

          <Textarea
            rows={8}
            value={replyContent}
            placeholder={replyMode === "public" ? "输入给客户的回复内容" : "输入内部备注"}
            disabled={submitting}
            onChange={(event) => setReplyContent(event.target.value)}
          />

          {replyMode === "internal" ? (
            <div className="space-y-3 rounded-lg border border-border/60 bg-muted/20 p-4">
              <div className="text-sm font-medium">@提及协作人</div>
              <div className="flex gap-2">
                <div className="flex-1">
                  <OptionCombobox
                    value={mentionUserId}
                    options={mentionOptions}
                    placeholder={loadingAgents ? "加载中..." : "选择要提及的客服"}
                    searchPlaceholder="搜索客服"
                    emptyText="暂无可选客服"
                    disabled={submitting || loadingAgents}
                    onChange={setMentionUserId}
                  />
                </div>
                <Button
                  type="button"
                  variant="outline"
                  disabled={submitting || loadingAgents}
                  onClick={handleAddMentionUser}
                >
                  添加
                </Button>
              </div>
              {mentionedUsers.length ? (
                <div className="flex flex-wrap gap-2">
                  {mentionedUsers.map((user) => (
                    <button
                      key={user.userId}
                      type="button"
                      className="rounded-full border px-3 py-1 text-xs"
                      onClick={() =>
                        setMentionedUsers((current) =>
                          current.filter((item) => item.userId !== user.userId),
                        )
                      }
                    >
                      @
                      {user.displayName ||
                        user.nickname ||
                        user.username ||
                        `客服#${user.userId}`}{" "}
                      ×
                    </button>
                  ))}
                </div>
              ) : (
                <div className="text-xs text-muted-foreground">未添加提及对象</div>
              )}
            </div>
          ) : null}
        </div>
        <DialogFooter className="mx-0 mb-0 px-6 py-4">
          <Button
            type="button"
            variant="outline"
            disabled={submitting}
            onClick={() => onOpenChange(false)}
          >
            取消
          </Button>
          <Button type="button" disabled={submitting} onClick={() => void handleSubmit()}>
            <MessageSquarePlusIcon className="size-4" />
            {submitting ? "提交中..." : replyMode === "public" ? "发送回复" : "保存备注"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
