"use client";

import {
  Building2Icon,
  Link2Icon,
  MailIcon,
  MessageCircleIcon,
  PencilIcon,
  PhoneIcon,
  UserRoundIcon,
} from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { toast } from "sonner";

import { CustomerFormDialog } from "@/components/customer-form-dialog";
import { type CustomerFormSavePayload } from "@/components/customer-form";
import { CustomerLinkOrCreateDialog } from "@/components/customer-link-or-create-dialog";
import type { AgentConversation } from "@/lib/api/agent";
import { useAgentConversationsStore } from "@/lib/stores/agent-conversations";
import { fetchCustomerContacts, type AdminCustomerContact } from "@/lib/api/customer-contact";
import { fetchCustomer, saveCustomerProfile, type AdminCustomer } from "@/lib/api/customer";
import { updateCompany, type AdminCompany } from "@/lib/api/company";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Field,
  FieldContent,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import {
  ContactType,
  ContactTypeLabels,
  Gender,
  GenderLabels,
  Status,
  StatusLabels,
} from "@/lib/generated/enums";
import { cn, formatDateTime } from "@/lib/utils";

function contactTypeLabel(contactType: string) {
  return ContactTypeLabels[contactType as ContactType] ?? contactType;
}

function customerStatusBadgeVariant(status: number): "secondary" | "destructive" | "outline" {
  if (status === Status.Disabled) {
    return "destructive";
  }
  if (status === Status.Deleted) {
    return "destructive";
  }
  return "secondary";
}

function ContactTypeIcon({ contactType }: { contactType: string }) {
  const cls = "size-3.5 shrink-0 text-muted-foreground";
  switch (contactType) {
    case ContactType.Mobile:
      return <PhoneIcon className={cls} aria-hidden />;
    case ContactType.Email:
      return <MailIcon className={cls} aria-hidden />;
    case ContactType.WeChat:
      return <MessageCircleIcon className={cls} aria-hidden />;
    default:
      return <Link2Icon className={cls} aria-hidden />;
  }
}

/** 侧边栏窄屏：左标签右内容，便于扫读 */
function DetailRow({
  label,
  value,
  valueClassName,
}: {
  label: string;
  value: string;
  valueClassName?: string;
}) {
  const empty = !value.trim();
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
  );
}

function SectionHeading({
  children,
  action,
}: {
  children: React.ReactNode;
  action?: React.ReactNode;
}) {
  return (
    <div className="flex items-center justify-between gap-2">
      <h3 className="text-xs font-medium text-muted-foreground">{children}</h3>
      {action}
    </div>
  );
}

function UnlinkedCustomerEmpty({ conversation }: { conversation: AgentConversation }) {
  const [linkDialogOpen, setLinkDialogOpen] = useState(false);
  const loadConversations = useAgentConversationsStore((s) => s.loadConversations);

  return (
    <div className="space-y-6 pt-2">
      <div className="flex flex-col items-center justify-center rounded-xl bg-muted/35 px-4 py-8 text-center">
        <UserRoundIcon className="mb-2 size-10 text-muted-foreground" aria-hidden />
        <p className="text-sm font-medium text-foreground">尚未关联 CRM 客户</p>
        <p className="mt-1 max-w-xs text-xs leading-relaxed text-muted-foreground">
          当前会话未绑定客户主档。绑定后可在此维护公司与联系方式。
        </p>
        <Button
          type="button"
          className="mt-4 gap-2"
          onClick={() => setLinkDialogOpen(true)}
        >
          <Link2Icon className="size-4" />
          关联或创建客户
        </Button>
      </div>
      <div className="space-y-2">
        <SectionHeading>访客标识</SectionHeading>
        <div className="space-y-2">
          <DetailRow label="外部来源" value={conversation.externalSource} />
          <DetailRow label="外部标识" value={conversation.externalId} />
        </div>
      </div>
      <CustomerLinkOrCreateDialog
        open={linkDialogOpen}
        onOpenChange={setLinkDialogOpen}
        conversationId={conversation.id}
        onSuccess={() => void loadConversations()}
      />
    </div>
  );
}

type CustomerTabPanelProps = {
  conversation: AgentConversation;
};

export function CustomerTabPanel({ conversation }: CustomerTabPanelProps) {
  const customerId = conversation.customerId ?? 0;

  if (customerId <= 0) {
    return <UnlinkedCustomerEmpty conversation={conversation} />;
  }

  return <CustomerLinkedBody conversation={conversation} customerId={customerId} />;
}

type CustomerLinkedBodyProps = {
  conversation: AgentConversation;
  customerId: number;
};

function CustomerLinkedBody({ conversation, customerId }: CustomerLinkedBodyProps) {
  const [loading, setLoading] = useState(true);
  const [customer, setCustomer] = useState<AdminCustomer | null>(null);
  const [contacts, setContacts] = useState<AdminCustomerContact[]>([]);

  const [customerEditOpen, setCustomerEditOpen] = useState(false);
  const [customerEditSaving, setCustomerEditSaving] = useState(false);
  const [companyEditOpen, setCompanyEditOpen] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const c = await fetchCustomer(customerId);
      setCustomer(c);
      const list = await fetchCustomerContacts(customerId);
      setContacts(Array.isArray(list) ? list : []);
    } catch (e) {
      const msg = e instanceof Error ? e.message : "加载客户信息失败";
      toast.error(msg);
      setCustomer(null);
      setContacts([]);
    } finally {
      setLoading(false);
    }
  }, [customerId]);

  useEffect(() => {
    void load();
  }, [load]);

  /** 客户主档存在但姓名等均为空时的弱空态 */
  const isProfileEmpty =
    customer &&
    !customer.name.trim() &&
    !customer.primaryMobile.trim() &&
    !customer.primaryEmail.trim() &&
    customer.companyId === 0 &&
    !customer.remark.trim();

  if (loading && !customer) {
    return (
      <p className="pt-4 text-sm text-muted-foreground">加载客户信息…</p>
    );
  }

  if (!customer) {
    return (
      <div className="space-y-3 pt-2">
        <p className="text-sm text-muted-foreground">
          无法加载客户（可能已被删除）。会话仍绑定客户 ID {customerId}。
        </p>
        <p className="text-xs text-muted-foreground">
          外部来源 {conversation.externalSource} / {conversation.externalId}
        </p>
      </div>
    );
  }

  const displayName = customer.name.trim() || "未填写姓名";
  const company = customer.company ?? null;
  const companyLine =
    customer.companyId > 0 ? company?.name || `公司 ID ${customer.companyId}` : "";
  const genderLabel =
    customer.gender === Gender.Male || customer.gender === Gender.Female
      ? GenderLabels[customer.gender as Gender] ?? String(customer.gender)
      : null;

  return (
    <div className="space-y-7 py-3">
      {isProfileEmpty ? (
        <div className="rounded-lg bg-amber-500/10 px-3 py-2.5 text-xs leading-relaxed text-amber-950 dark:text-amber-100">
          客户主档已关联，但基础信息尚未填写。请点击「编辑」补全资料。
        </div>
      ) : null}

      <section className="space-y-4">
        <div className="space-y-1.5">
          <div className="flex items-start justify-between gap-2">
            <p className="min-w-0 flex-1 line-clamp-2 leading-snug">
              <span className="text-base font-semibold text-foreground">
                {displayName}
              </span>
              {genderLabel ? (
                <span className="text-sm font-normal text-muted-foreground">
                  {" "}
                  · {genderLabel}
                </span>
              ) : null}
            </p>
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
          </div>
          {companyLine ? (
            <p className="flex min-w-0 items-start gap-1.5 text-xs text-muted-foreground">
              <Building2Icon
                className="mt-0.5 size-3.5 shrink-0 opacity-80"
                aria-hidden
              />
              <span className="min-w-0 break-all">{companyLine}</span>
            </p>
          ) : null}
        </div>

        {/* <div className="space-y-2 text-sm">
          {customer.primaryMobile ? (
            <div className="flex min-w-0 items-baseline gap-2">
              <PhoneIcon className="size-3.5 shrink-0 translate-y-0.5 text-muted-foreground" aria-hidden />
              <span className="min-w-0 break-all tabular-nums text-foreground">{customer.primaryMobile}</span>
              <span className="shrink-0 text-xs text-muted-foreground">主</span>
            </div>
          ) : null}
          {customer.primaryEmail ? (
            <div className="flex min-w-0 items-baseline gap-2">
              <MailIcon className="size-3.5 shrink-0 translate-y-0.5 text-muted-foreground" aria-hidden />
              <span className="min-w-0 break-all text-foreground">{customer.primaryEmail}</span>
              <span className="shrink-0 text-xs text-muted-foreground">主</span>
            </div>
          ) : null}
        </div> */}

        <div className="space-y-2">
          <DetailRow
            label="最近活跃"
            value={
              customer.lastActiveAt ? formatDateTime(customer.lastActiveAt) : ""
            }
          />
          <DetailRow
            label="备注"
            value={customer.remark.trim() ? customer.remark : ""}
            valueClassName="whitespace-pre-wrap"
          />
          <DetailRow
            label="创建时间"
            value={formatDateTime(customer.createdAt)}
            valueClassName="whitespace-pre-wrap"
          />
          <DetailRow
            label="更新时间"
            value={formatDateTime(customer.updatedAt)}
            valueClassName="whitespace-pre-wrap"
          />
        </div>
      </section>

      {customer.companyId > 0 ? (
        <section className="space-y-2 border-t">
          <SectionHeading
            action={
              company ? (
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  className="h-7 gap-1 px-2 text-xs"
                  onClick={() => setCompanyEditOpen(true)}
                >
                  <PencilIcon className="size-3.5" />
                  编辑
                </Button>
              ) : null
            }
          >
            公司
          </SectionHeading>
          {company ? (
            <div className="space-y-2">
              <div className="flex items-start gap-2 text-sm">
                <Building2Icon
                  className="mt-0.5 size-4 shrink-0 text-muted-foreground"
                  aria-hidden
                />
                <div className="min-w-0 flex-1 space-y-0.5">
                  <p className="font-medium leading-snug text-foreground">
                    {company.name}
                  </p>
                  {company.code ? (
                    <p className="font-mono text-xs text-muted-foreground">
                      {company.code}
                    </p>
                  ) : null}
                </div>
              </div>
              <div className="space-y-2 pt-1">
                <DetailRow
                  label="创建"
                  value={formatDateTime(company.createdAt)}
                />
                <DetailRow
                  label="更新"
                  value={formatDateTime(company.updatedAt)}
                />
              </div>
              <DetailRow
                label="备注"
                value={company.remark.trim() ? company.remark : ""}
                valueClassName="whitespace-pre-wrap"
              />
            </div>
          ) : (
            <p className="text-sm text-muted-foreground">
              公司信息加载失败或公司已删除。
            </p>
          )}
        </section>
      ) : null}

      <section className="space-y-2">
        <SectionHeading>联系方式</SectionHeading>
        {contacts.length === 0 ? (
          <p className="text-sm text-muted-foreground">暂无</p>
        ) : (
          <ul className="space-y-3">
            {contacts.map((row) => {
              const tags: string[] = [];
              if (row.isPrimary) {
                tags.push("主");
              }
              if (row.isVerified) {
                tags.push("已验证");
              }
              return (
                <li key={row.id} className="text-sm">
                  <div className="flex gap-2 items-center">
                    <ContactTypeIcon contactType={row.contactType} />
                    <div className="min-w-0 flex-1">
                      <p className="break-all font-medium leading-snug text-foreground">
                        {row.contactValue}
                        <span className="ml-2 font-normal text-xs text-muted-foreground">
                          {contactTypeLabel(row.contactType)}
                        </span>
                        {tags.length > 0 ? (
                          <span className="ml-2 text-xs text-muted-foreground">
                            {tags.join(" · ")}
                          </span>
                        ) : null}
                      </p>
                      {row.remark ? (
                        <p className="mt-1 line-clamp-3 text-xs leading-relaxed text-muted-foreground break-all">
                          {row.remark}
                        </p>
                      ) : null}
                    </div>
                  </div>
                </li>
              );
            })}
          </ul>
        )}
      </section>

      <CustomerFormDialog
        open={customerEditOpen}
        onOpenChange={setCustomerEditOpen}
        saving={customerEditSaving}
        itemId={customer.id}
        onSave={async (payload: CustomerFormSavePayload) => {
          if (customerEditSaving) {
            return;
          }
          setCustomerEditSaving(true);
          try {
            await saveCustomerProfile({ ...payload, id: customer.id });
            toast.success("已保存");
            void load();
            setCustomerEditOpen(false);
          } catch (e) {
            toast.error(e instanceof Error ? e.message : "保存失败");
          } finally {
            setCustomerEditSaving(false);
          }
        }}
      />
      {company ? (
        <CompanyEditDialog
          open={companyEditOpen}
          onOpenChange={setCompanyEditOpen}
          company={company}
          onSaved={() => {
            void load();
          }}
        />
      ) : null}
    </div>
  );
}

type CompanyEditDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  company: AdminCompany;
  onSaved: () => void;
};

function CompanyEditDialog({
  open,
  onOpenChange,
  company,
  onSaved,
}: CompanyEditDialogProps) {
  const [name, setName] = useState("");
  const [code, setCode] = useState("");
  const [remark, setRemark] = useState("");
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (!open) {
      return;
    }
    setName(company.name);
    setCode(company.code);
    setRemark(company.remark);
  }, [open, company]);

  const handleSubmit = async () => {
    const trimmedName = name.trim();
    if (!trimmedName) {
      toast.error("公司名称不能为空");
      return;
    }
    setSaving(true);
    try {
      await updateCompany({
        id: company.id,
        name: trimmedName,
        code: code.trim(),
        remark: remark.trim(),
      });
      toast.success("已保存");
      onSaved();
      onOpenChange(false);
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "保存失败");
    } finally {
      setSaving(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md" showCloseButton>
        <DialogHeader>
          <DialogTitle>编辑公司</DialogTitle>
        </DialogHeader>
        <div className="flex flex-col gap-4 py-1">
          <Field orientation="vertical">
            <FieldLabel htmlFor="co-name">公司名称</FieldLabel>
            <FieldContent>
              <Input id="co-name" value={name} onChange={(e) => setName(e.target.value)} />
            </FieldContent>
          </Field>
          <Field orientation="vertical">
            <FieldLabel htmlFor="co-code">公司编码</FieldLabel>
            <FieldContent>
              <Input id="co-code" value={code} onChange={(e) => setCode(e.target.value)} />
            </FieldContent>
          </Field>
          <Field orientation="vertical">
            <FieldLabel htmlFor="co-remark">备注</FieldLabel>
            <FieldContent>
              <Textarea id="co-remark" value={remark} onChange={(e) => setRemark(e.target.value)} rows={3} />
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
  );
}
