"use client"

import {
  Building2Icon,
  Link2Icon,
  MailIcon,
  PencilIcon,
  PhoneIcon,
  UserRoundIcon,
} from "lucide-react"
import { useCallback, useEffect, useState } from "react"
import { toast } from "sonner"

import { type CustomerFormSavePayload } from "@/components/customer-form"
import { CustomerFormDialog } from "@/components/customer-form-dialog"
import { CustomerLinkOrCreateDialog } from "@/components/customer-link-or-create-dialog"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Field, FieldContent, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"
import { updateCompany, type AdminCompany } from "@/lib/api/company"
import {
  fetchCustomer,
  saveCustomerProfile,
  type AdminCustomer,
} from "@/lib/api/customer"
import {
  fetchCustomerContacts,
  type AdminCustomerContact,
} from "@/lib/api/customer-contact"
import { Gender, GenderLabels, ContactType, ContactTypeLabels } from "@/lib/generated/enums"
import { cn, formatDateTime } from "@/lib/utils"

function contactTypeLabel(contactType: ContactType | string) {
  return ContactTypeLabels[contactType as ContactType] ?? contactType
}

function ContactTypeIcon({ contactType }: { contactType: ContactType | string }) {
  const cls = "size-3.5 shrink-0 text-muted-foreground"
  switch (contactType) {
    case ContactType.Mobile:
      return <PhoneIcon className={cls} aria-hidden />
    case ContactType.Email:
      return <MailIcon className={cls} aria-hidden />
    default:
      return <Link2Icon className={cls} aria-hidden />
  }
}

function DetailRow({
  label,
  value,
  valueClassName,
}: {
  label: string
  value: string
  valueClassName?: string
}) {
  const empty = !value.trim()
  return (
    <div className="flex gap-2.5 text-sm leading-snug">
      <span className="w-17 shrink-0 pt-px text-xs text-muted-foreground">{label}</span>
      <span
        className={cn(
          "min-w-0 flex-1 break-all text-foreground",
          empty && "text-muted-foreground",
          valueClassName,
        )}
      >
        {empty ? "—" : value}
      </span>
    </div>
  )
}

function SectionHeading({
  children,
  action,
}: {
  children: React.ReactNode
  action?: React.ReactNode
}) {
  return (
    <div className="flex items-center justify-between gap-2">
      <h3 className="text-xs font-medium text-muted-foreground">{children}</h3>
      {action}
    </div>
  )
}

function UnlinkedCustomerEmpty({
  ticketId,
  onSuccess,
}: {
  ticketId: number
  onSuccess: () => void | Promise<void>
}) {
  const [linkDialogOpen, setLinkDialogOpen] = useState(false)

  return (
    <div className="space-y-4 pt-2">
      <div className="flex flex-col items-center justify-center rounded-xl bg-muted/35 px-4 py-8 text-center">
        <UserRoundIcon className="mb-2 size-10 text-muted-foreground" aria-hidden />
        <p className="text-sm font-medium text-foreground">尚未关联 CRM 客户</p>
        <p className="mt-1 max-w-xs text-xs leading-relaxed text-muted-foreground">
          当前工单未绑定客户主档。绑定后可在此查看客户资料、公司信息与联系方式。
        </p>
        <Button type="button" className="mt-4 gap-2" onClick={() => setLinkDialogOpen(true)}>
          <Link2Icon className="size-4" />
          关联或创建客户
        </Button>
      </div>
      <CustomerLinkOrCreateDialog
        open={linkDialogOpen}
        onOpenChange={setLinkDialogOpen}
        ticketId={ticketId}
        onSuccess={onSuccess}
      />
    </div>
  )
}

function MissingCustomerEmpty({
  ticketId,
  onSuccess,
}: {
  ticketId: number
  onSuccess: () => void | Promise<void>
}) {
  const [linkDialogOpen, setLinkDialogOpen] = useState(false)

  return (
    <div className="space-y-4 pt-2">
      <div className="flex flex-col items-center justify-center rounded-xl bg-muted/35 px-4 py-8 text-center">
        <UserRoundIcon className="mb-2 size-10 text-muted-foreground" aria-hidden />
        <p className="text-sm font-medium text-foreground">客户已删除或不存在</p>
        <p className="mt-1 max-w-xs text-xs leading-relaxed text-muted-foreground">
          当前工单绑定的客户主档已不可用。你可以重新关联已有客户，或直接新建一个客户并绑定到当前工单。
        </p>
        <Button type="button" className="mt-4 gap-2" onClick={() => setLinkDialogOpen(true)}>
          <Link2Icon className="size-4" />
          重新关联或创建客户
        </Button>
      </div>
      <CustomerLinkOrCreateDialog
        open={linkDialogOpen}
        onOpenChange={setLinkDialogOpen}
        ticketId={ticketId}
        onSuccess={onSuccess}
      />
    </div>
  )
}

type TicketCustomerPanelProps = {
  ticketId: number
  customerId?: number
  onRefresh: () => void | Promise<void>
}

type TicketLinkedCustomerPanelProps = {
  ticketId: number
  customerId: number
  onRefresh: () => void | Promise<void>
}

export function TicketCustomerPanel({
  ticketId,
  customerId = 0,
  onRefresh,
}: TicketCustomerPanelProps) {
  if (customerId <= 0) {
    return <UnlinkedCustomerEmpty ticketId={ticketId} onSuccess={onRefresh} />
  }
  return (
    <TicketLinkedCustomerPanel
      ticketId={ticketId}
      customerId={customerId}
      onRefresh={onRefresh}
    />
  )
}

function TicketLinkedCustomerPanel({
  ticketId,
  customerId,
  onRefresh,
}: TicketLinkedCustomerPanelProps) {
  const linkedCustomerId = customerId
  const [loading, setLoading] = useState(true)
  const [customer, setCustomer] = useState<AdminCustomer | null>(null)
  const [contacts, setContacts] = useState<AdminCustomerContact[]>([])

  const [customerEditOpen, setCustomerEditOpen] = useState(false)
  const [customerEditSaving, setCustomerEditSaving] = useState(false)
  const [companyEditOpen, setCompanyEditOpen] = useState(false)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const c = await fetchCustomer(linkedCustomerId)
      setCustomer(c)
      if (!c) {
        setContacts([])
        return
      }
      const list = await fetchCustomerContacts(linkedCustomerId)
      setContacts(Array.isArray(list) ? list : [])
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "加载客户信息失败")
      setCustomer(null)
      setContacts([])
    } finally {
      setLoading(false)
    }
  }, [linkedCustomerId])

  useEffect(() => {
    void load()
  }, [load])

  if (loading && !customer) {
    return <p className="pt-4 text-sm text-muted-foreground">加载客户信息…</p>
  }

  if (!customer) {
    return <MissingCustomerEmpty ticketId={ticketId} onSuccess={onRefresh} />
  }

  const displayName = customer.name.trim() || "未填写姓名"
  const company = customer.company ?? null
  const genderLabel =
    customer.gender === Gender.Male || customer.gender === Gender.Female
      ? GenderLabels[customer.gender as Gender] ?? String(customer.gender)
      : null
  const isProfileEmpty =
    !customer.name.trim() &&
    !customer.primaryMobile.trim() &&
    !customer.primaryEmail.trim() &&
    customer.companyId === 0 &&
    !customer.remark.trim()

  return (
    <div className="space-y-4 pt-2">
      {isProfileEmpty ? (
        <div className="rounded-lg bg-amber-500/10 px-3 py-2.5 text-xs leading-relaxed text-amber-950 dark:text-amber-100">
          客户主档已关联，但基础信息尚未填写。请点击「编辑」补全资料。
        </div>
      ) : null}

      <section className="space-y-2">
        <SectionHeading
          action={
            <Button
              type="button"
              variant="ghost"
              size="sm"
              className="h-7 shrink-0 gap-1 px-2 text-xs"
              onClick={() => setCustomerEditOpen(true)}
            >
              <PencilIcon className="size-3.5" />
              编辑
            </Button>
          }
        >
          客户信息
        </SectionHeading>
        <div className="flex min-w-0 items-start gap-2 text-sm">
          <UserRoundIcon className="mt-0.5 size-4 shrink-0 text-muted-foreground" aria-hidden />
          <div className="min-w-0 flex-1 space-y-0.5">
            <p className="line-clamp-2 leading-snug text-foreground">
              <span className="font-medium">{displayName}</span>
              {genderLabel ? (
                <span className="font-normal text-muted-foreground"> · {genderLabel}</span>
              ) : null}
            </p>
          </div>
        </div>
        <div className="space-y-2">
          <DetailRow label="手机" value={customer.primaryMobile || ""} />
          <DetailRow label="邮箱" value={customer.primaryEmail || ""} />
          <DetailRow
            label="最近活跃"
            value={customer.lastActiveAt ? formatDateTime(customer.lastActiveAt) : ""}
          />
          <DetailRow
            label="备注"
            value={customer.remark.trim() ? customer.remark : ""}
            valueClassName="whitespace-pre-wrap"
          />
        </div>
      </section>

      <section className="space-y-2">
        <SectionHeading>联系方式</SectionHeading>
        {contacts.length === 0 ? (
          <p className="text-sm text-muted-foreground">暂无联系方式</p>
        ) : (
          <ul className="space-y-3">
            {contacts.map((row) => {
              const tags: string[] = []
              if (row.isPrimary) tags.push("主")
              if (row.isVerified) tags.push("已验证")
              return (
                <li key={row.id} className="text-sm">
                  <div className="flex items-center gap-2">
                    <ContactTypeIcon contactType={row.contactType} />
                    <div className="min-w-0 flex-1">
                      <p className="break-all font-medium leading-snug text-foreground">
                        {row.contactValue}
                        <span className="ml-2 text-xs font-normal text-muted-foreground">
                          {contactTypeLabel(row.contactType)}
                        </span>
                        {tags.length > 0 ? (
                          <span className="ml-2 text-xs text-muted-foreground">
                            {tags.join(" · ")}
                          </span>
                        ) : null}
                      </p>
                      {row.remark ? (
                        <p className="mt-1 line-clamp-3 break-all text-xs leading-relaxed text-muted-foreground">
                          {row.remark}
                        </p>
                      ) : null}
                    </div>
                  </div>
                </li>
              )
            })}
          </ul>
        )}
      </section>

      <section className="space-y-2 border-t pt-2">
        <SectionHeading
          action={
            company ? (
              <Button
                type="button"
                variant="ghost"
                size="sm"
                className="h-7 shrink-0 gap-1 px-2 text-xs"
                onClick={() => setCompanyEditOpen(true)}
              >
                <PencilIcon className="size-3.5" />
                编辑
              </Button>
            ) : null
          }
        >
          公司信息
        </SectionHeading>
        {company ? (
          <div className="space-y-2">
            <div className="flex min-w-0 items-start gap-2 text-sm">
              <Building2Icon className="mt-0.5 size-4 shrink-0 text-muted-foreground" aria-hidden />
              <div className="min-w-0 flex-1 space-y-0.5">
                <p className="line-clamp-2 font-medium leading-snug text-foreground">
                  {company.name}
                </p>
                {company.code ? (
                  <p className="font-mono text-xs text-muted-foreground">{company.code}</p>
                ) : null}
              </div>
            </div>
            <div className="space-y-2 pt-1">
              <DetailRow label="创建" value={formatDateTime(company.createdAt)} />
              <DetailRow label="更新" value={formatDateTime(company.updatedAt)} />
              <DetailRow
                label="备注"
                value={company.remark.trim() ? company.remark : ""}
                valueClassName="whitespace-pre-wrap"
              />
            </div>
          </div>
        ) : (
          <p className="text-sm text-muted-foreground">
            未关联公司。可通过编辑客户资料补充公司信息。
          </p>
        )}
      </section>

      <CustomerFormDialog
        open={customerEditOpen}
        onOpenChange={setCustomerEditOpen}
        saving={customerEditSaving}
        itemId={customer.id}
        onSave={async (payload: CustomerFormSavePayload) => {
          if (customerEditSaving) {
            return
          }
          setCustomerEditSaving(true)
          try {
            await saveCustomerProfile({ ...payload, id: customer.id })
            toast.success("已保存")
            await load()
            await onRefresh()
            setCustomerEditOpen(false)
          } catch (error) {
            toast.error(error instanceof Error ? error.message : "保存失败")
          } finally {
            setCustomerEditSaving(false)
          }
        }}
      />
      {company ? (
        <CompanyEditDialog
          open={companyEditOpen}
          onOpenChange={setCompanyEditOpen}
          company={company}
          onSaved={async () => {
            await load()
            await onRefresh()
          }}
        />
      ) : null}
    </div>
  )
}

type CompanyEditDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  company: AdminCompany
  onSaved: () => void | Promise<void>
}

function CompanyEditDialog({
  open,
  onOpenChange,
  company,
  onSaved,
}: CompanyEditDialogProps) {
  const [name, setName] = useState("")
  const [code, setCode] = useState("")
  const [remark, setRemark] = useState("")
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    if (!open) {
      return
    }
    setName(company.name)
    setCode(company.code)
    setRemark(company.remark)
  }, [open, company])

  const handleSubmit = async () => {
    const trimmedName = name.trim()
    if (!trimmedName) {
      toast.error("公司名称不能为空")
      return
    }
    setSaving(true)
    try {
      await updateCompany({
        id: company.id,
        name: trimmedName,
        code: code.trim(),
        remark: remark.trim(),
      })
      toast.success("已保存")
      await onSaved()
      onOpenChange(false)
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "保存失败")
    } finally {
      setSaving(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md" showCloseButton>
        <DialogHeader>
          <DialogTitle>编辑公司</DialogTitle>
        </DialogHeader>
        <div className="flex flex-col gap-4 py-1">
          <Field orientation="vertical">
            <FieldLabel htmlFor="ticket-company-name">公司名称</FieldLabel>
            <FieldContent>
              <Input
                id="ticket-company-name"
                value={name}
                onChange={(event) => setName(event.target.value)}
              />
            </FieldContent>
          </Field>
          <Field orientation="vertical">
            <FieldLabel htmlFor="ticket-company-code">公司编码</FieldLabel>
            <FieldContent>
              <Input
                id="ticket-company-code"
                value={code}
                onChange={(event) => setCode(event.target.value)}
              />
            </FieldContent>
          </Field>
          <Field orientation="vertical">
            <FieldLabel htmlFor="ticket-company-remark">备注</FieldLabel>
            <FieldContent>
              <Textarea
                id="ticket-company-remark"
                value={remark}
                onChange={(event) => setRemark(event.target.value)}
                rows={3}
              />
            </FieldContent>
          </Field>
        </div>
        <DialogFooter>
          <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
            取消
          </Button>
          <Button type="button" disabled={saving} onClick={() => void handleSubmit()}>
            {saving ? "保存中…" : "保存"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
