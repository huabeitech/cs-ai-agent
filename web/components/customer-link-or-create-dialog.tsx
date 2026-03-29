"use client"

import { useCallback, useEffect, useState } from "react"
import { toast } from "sonner"

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
  FieldLabel,
} from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"
import { OptionCombobox, type ComboboxOption } from "@/components/option-combobox"
import { linkConversationToCustomer } from "@/lib/api/agent"
import { fetchCompanies, type AdminCompany } from "@/lib/api/company"
import {
  createCustomer,
  fetchCustomers,
  type AdminCustomer,
  type CreateAdminCustomerPayload,
} from "@/lib/api/customer"
import { getEnumOptions } from "@/lib/enums"
import { GenderLabels } from "@/lib/generated/enums"

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

const genderOptions: ComboboxOption[] = [
  ...getEnumOptions(GenderLabels).map((item) => ({
    value: String(item.value),
    label: item.label,
  })),
]

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

  const [companyOptions, setCompanyOptions] = useState<ComboboxOption[]>([
    { value: "0", label: "无" },
  ])

  const [formName, setFormName] = useState("")
  const [formGender, setFormGender] = useState("0")
  const [formCompanyId, setFormCompanyId] = useState("0")
  const [formMobile, setFormMobile] = useState("")
  const [formEmail, setFormEmail] = useState("")
  const [formRemark, setFormRemark] = useState("")

  const resetForm = useCallback(() => {
    setFormName("")
    setFormGender("0")
    setFormCompanyId("0")
    setFormMobile("")
    setFormEmail("")
    setFormRemark("")
  }, [])

  useEffect(() => {
    if (!open) {
      return
    }
    setSearchText("")
    setResults([])
    setShowCreate(false)
    setLinkingId(null)
    resetForm()
    let cancelled = false
    void (async () => {
      try {
        const data = await fetchCompanies({ status: 0, page: 1, limit: 500 })
        if (cancelled) {
          return
        }
        setCompanyOptions([
          { value: "0", label: "无" },
          ...data.results.map((c: AdminCompany) => ({
            value: String(c.id),
            label: c.name || `公司 #${c.id}`,
          })),
        ])
      } catch {
        if (!cancelled) {
          setCompanyOptions([{ value: "0", label: "无" }])
        }
      }
    })()
    return () => {
      cancelled = true
    }
  }, [open, resetForm])

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

  const handleCreateSubmit = async () => {
    const name = formName.trim()
    if (!name) {
      toast.error("客户名称不能为空")
      return
    }
    const payload: CreateAdminCustomerPayload = {
      name,
      gender: Number.parseInt(formGender, 10) || 0,
      companyId: Number.parseInt(formCompanyId, 10) || 0,
      primaryMobile: formMobile.trim(),
      primaryEmail: formEmail.trim(),
      remark: formRemark.trim(),
    }
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

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        className="max-h-[min(90vh,640px)] overflow-y-auto sm:max-w-lg"
        showCloseButton
      >
        <DialogHeader>
          <DialogTitle>关联或创建客户</DialogTitle>
          <DialogDescription>
            先搜索已有客户；{conversationId ? "选中即可关联当前会话。" : "未接入会话时仅创建或定位客户。"}
            若无结果，可填写下方新客户
            {conversationId ? "，保存后将自动关联会话。" : "。"}
          </DialogDescription>
        </DialogHeader>

        <div className="flex flex-col gap-4 py-1">
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
            <div className="flex flex-col gap-3 rounded-lg border border-border bg-muted/10 p-3">
              <Field orientation="vertical">
                <FieldLabel htmlFor="nl-name">客户名称</FieldLabel>
                <FieldContent>
                  <Input
                    id="nl-name"
                    value={formName}
                    onChange={(e) => setFormName(e.target.value)}
                    autoComplete="off"
                  />
                </FieldContent>
              </Field>
              <Field orientation="vertical">
                <FieldLabel>性别</FieldLabel>
                <FieldContent>
                  <OptionCombobox
                    value={formGender}
                    options={genderOptions}
                    placeholder="性别"
                    onChange={setFormGender}
                  />
                </FieldContent>
              </Field>
              <Field orientation="vertical">
                <FieldLabel>所属公司</FieldLabel>
                <FieldContent>
                  <OptionCombobox
                    value={formCompanyId}
                    options={companyOptions}
                    placeholder="选择公司"
                    onChange={setFormCompanyId}
                  />
                </FieldContent>
              </Field>
              <Field orientation="vertical">
                <FieldLabel htmlFor="nl-mobile">主手机号</FieldLabel>
                <FieldContent>
                  <Input
                    id="nl-mobile"
                    value={formMobile}
                    onChange={(e) => setFormMobile(e.target.value)}
                  />
                </FieldContent>
              </Field>
              <Field orientation="vertical">
                <FieldLabel htmlFor="nl-email">主邮箱</FieldLabel>
                <FieldContent>
                  <Input
                    id="nl-email"
                    value={formEmail}
                    onChange={(e) => setFormEmail(e.target.value)}
                  />
                </FieldContent>
              </Field>
              <Field orientation="vertical">
                <FieldLabel htmlFor="nl-remark">备注</FieldLabel>
                <FieldContent>
                  <Textarea
                    id="nl-remark"
                    rows={2}
                    value={formRemark}
                    onChange={(e) => setFormRemark(e.target.value)}
                  />
                </FieldContent>
              </Field>
            </div>
          ) : null}
        </div>

        <DialogFooter className="gap-2 sm:justify-between">
          <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
            关闭
          </Button>
          {showCreate ? (
            <Button
              type="button"
              disabled={saving}
              onClick={() => void handleCreateSubmit()}
            >
              {saving ? "提交中…" : conversationId ? "创建并关联会话" : "创建客户"}
            </Button>
          ) : null}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
