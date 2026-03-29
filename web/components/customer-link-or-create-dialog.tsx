"use client"

import { useEffect, useState } from "react"
import { toast } from "sonner"

import { CustomerForm } from "@/components/customer-form"
import { ProjectDialog } from "@/components/project-dialog"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { linkConversationToCustomer } from "@/lib/api/agent"
import {
  createCustomer,
  fetchCustomers,
  type AdminCustomer,
  type CreateAdminCustomerPayload,
} from "@/lib/api/customer"

export type CustomerLinkOrCreateDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  /** 传入时会话侧：关联已有或新建后绑定该会话 */
  conversationId?: number | null
  /** 绑定成功或仅新建成功后的回调 */
  onSuccess?: () => void | Promise<void>
}

function buildCustomerSearchQuery(raw: string) {
  const q = raw.trim()
  if (!q) {
    return {}
  }
  if (q.includes("@")) {
    return { primaryEmail: q }
  }
  const digits = q.replace(/\s/g, "")
  if (/^\d{5,}$/.test(digits)) {
    return { primaryMobile: digits }
  }
  return { name: q }
}

const createFormId = "customer-link-or-create-form"

export function CustomerLinkOrCreateDialog({
  open,
  onOpenChange,
  conversationId,
  onSuccess,
}: CustomerLinkOrCreateDialogProps) {
  const [searchText, setSearchText] = useState("")
  const [searching, setSearching] = useState(false)
  const [results, setResults] = useState<AdminCustomer[]>([])
  const [showCreate, setShowCreate] = useState(false)
  const [linkingId, setLinkingId] = useState<number | null>(null)
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    if (!open) {
      return
    }
    setSearchText("")
    setResults([])
    setShowCreate(false)
    setLinkingId(null)
  }, [open])

  const runSearch = async () => {
    const q = searchText.trim()
    if (!q) {
      toast.error("请输入姓名、手机或邮箱关键词")
      return
    }
    setSearching(true)
    try {
      const filters = buildCustomerSearchQuery(q)
      const data = await fetchCustomers({
        ...filters,
        page: 1,
        limit: 50,
        status: 0,
      })
      setResults(data.results)
      if (data.results.length === 0) {
        toast.message("未找到匹配客户，可点击下方填写新客户")
      }
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "搜索失败")
    } finally {
      setSearching(false)
    }
  }

  const handleLinkExisting = async (customer: AdminCustomer) => {
    if (!conversationId) {
      toast.success(`已选择客户：${customer.name || `#${customer.id}`}`)
      onOpenChange(false)
      await onSuccess?.()
      return
    }
    setLinkingId(customer.id)
    try {
      await linkConversationToCustomer({
        conversationId,
        customerId: customer.id,
      })
      toast.success("已关联客户")
      onOpenChange(false)
      await onSuccess?.()
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "关联失败")
    } finally {
      setLinkingId(null)
    }
  }

  const onCreateSubmit = async (payload: CreateAdminCustomerPayload) => {
    setSaving(true)
    try {
      const created = await createCustomer(payload)
      if (conversationId) {
        await linkConversationToCustomer({
          conversationId,
          customerId: created.id,
        })
        toast.success("已创建客户并关联当前会话")
      } else {
        toast.success("已创建客户")
      }
      onOpenChange(false)
      await onSuccess?.()
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "保存失败")
    } finally {
      setSaving(false)
    }
  }

  const description = (
    <>
      先搜索已有客户；{conversationId ? "选中即可关联当前会话。" : "未接入会话时仅创建或定位客户。"}
      若无结果，可填写下方新客户
      {conversationId ? "，保存后将自动关联会话。" : "。"}
    </>
  )

  return (
    <ProjectDialog
      open={open}
      onOpenChange={(nextOpen) => onOpenChange(nextOpen)}
      title="关联或创建客户"
      description={description}
      allowFullscreen
      defaultFullscreen
      footer={
        <div className="flex w-full flex-wrap items-center justify-end gap-2">
          <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
            关闭
          </Button>
          {showCreate ? (
            <Button type="submit" form={createFormId} disabled={saving}>
              {saving ? "提交中…" : conversationId ? "创建并关联会话" : "创建客户"}
            </Button>
          ) : null}
        </div>
      }
    >
      <div className="flex flex-col gap-4">
        <div className="flex gap-2">
          <Input
            placeholder="姓名 / 手机 / 邮箱"
            value={searchText}
            onChange={(e) => setSearchText(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter") {
                e.preventDefault()
                void runSearch()
              }
            }}
          />
          <Button
            type="button"
            variant="secondary"
            disabled={searching}
            onClick={() => void runSearch()}
          >
            {searching ? "搜索中…" : "搜索"}
          </Button>
        </div>

        {results.length > 0 ? (
          <ul className="max-h-48 space-y-1.5 overflow-y-auto rounded-md border border-border p-2 text-sm">
            {results.map((row) => (
              <li
                key={row.id}
                className="flex items-center justify-between gap-2 rounded border border-transparent px-2 py-1.5 hover:bg-muted/40"
              >
                <div className="min-w-0">
                  <div className="truncate font-medium">{row.name || `客户 #${row.id}`}</div>
                  <div className="truncate text-xs text-muted-foreground">
                    {row.primaryMobile || "-"} · {row.primaryEmail || "-"}
                  </div>
                </div>
                <Button
                  type="button"
                  size="sm"
                  variant="outline"
                  className="shrink-0"
                  disabled={linkingId !== null}
                  onClick={() => void handleLinkExisting(row)}
                >
                  {linkingId === row.id
                    ? "处理中…"
                    : conversationId
                      ? "关联"
                      : "选用"}
                </Button>
              </li>
            ))}
          </ul>
        ) : null}

        <div className="border-t border-border pt-2">
          <button
            type="button"
            className="text-sm text-primary underline-offset-4 hover:underline"
            onClick={() => setShowCreate((v) => !v)}
          >
            {showCreate ? "收起新建表单" : "未找到？填写新客户"}
          </button>
        </div>

        {showCreate ? (
          <CustomerForm
            formId={createFormId}
            onSubmit={onCreateSubmit}
            fieldIdPrefix="link-or-create"
            remarkRows={2}
            className="flex flex-col gap-3 rounded-lg border border-border bg-muted/10 p-3"
          />
        ) : null}
      </div>
    </ProjectDialog>
  )
}
