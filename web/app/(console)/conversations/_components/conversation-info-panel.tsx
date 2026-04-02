"use client";

import Link from "next/link";
import {
  Building2Icon,
  Link2Icon,
  MailIcon,
  PencilIcon,
  PhoneIcon,
  UserRoundIcon,
} from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { toast } from "sonner";

import { type CustomerFormSavePayload } from "@/components/customer-form";
import { CustomerFormDialog } from "@/components/customer-form-dialog";
import { CustomerLinkOrCreateDialog } from "@/components/customer-link-or-create-dialog";
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
import type { AgentConversation } from "@/lib/api/agent";
import { type TagTree, fetchTagsAll } from "@/lib/api/admin";
import { updateCompany, type AdminCompany } from "@/lib/api/company";
import { fetchTickets, type TicketItem } from "@/lib/api/ticket";
import {
  fetchCustomer,
  saveCustomerProfile,
  type AdminCustomer,
} from "@/lib/api/customer";
import {
  fetchCustomerContacts,
  type AdminCustomerContact,
} from "@/lib/api/customer-contact";
import {
  ContactType,
  ContactTypeLabels,
  Gender,
  GenderLabels,
} from "@/lib/generated/enums";
import { useAgentConversationsStore } from "@/lib/stores/agent-conversations";
import { cn, formatDateTime } from "@/lib/utils";
import {
  ConversationTagBadges,
  ConversationTagPicker,
} from "./conversation-tag-picker";
import { TicketPriorityBadge } from "../../tickets/_components/ticket-priority-badge";
import { TicketStatusBadge } from "../../tickets/_components/ticket-status-badge";

function contactTypeLabel(contactType: ContactType | string) {
  return ContactTypeLabels[contactType as ContactType] ?? contactType;
}

function ContactTypeIcon({ contactType }: { contactType: ContactType | string }) {
  const cls = "size-3.5 shrink-0 text-muted-foreground";
  switch (contactType) {
    case ContactType.Mobile:
      return <PhoneIcon className={cls} aria-hidden />;
    case ContactType.Email:
      return <MailIcon className={cls} aria-hidden />;
    default:
      return <Link2Icon className={cls} aria-hidden />;
  }
}

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
      <CustomerLinkOrCreateDialog
        open={linkDialogOpen}
        onOpenChange={setLinkDialogOpen}
        conversationId={conversation.id}
        onSuccess={() => void loadConversations()}
      />
    </div>
  );
}

function MissingCustomerEmpty({ conversation }: { conversation: AgentConversation }) {
  const [linkDialogOpen, setLinkDialogOpen] = useState(false);
  const loadConversations = useAgentConversationsStore((s) => s.loadConversations);

  return (
    <div className="space-y-6 pt-2">
      <div className="flex flex-col items-center justify-center rounded-xl bg-muted/35 px-4 py-8 text-center">
        <UserRoundIcon className="mb-2 size-10 text-muted-foreground" aria-hidden />
        <p className="text-sm font-medium text-foreground">客户已删除或不存在</p>
        <p className="mt-1 max-w-xs text-xs leading-relaxed text-muted-foreground">
          当前会话绑定的客户主档已不可用。你可以重新关联已有客户，或直接新建一个客户并绑定到当前会话。
        </p>
        <Button
          type="button"
          className="mt-4 gap-2"
          onClick={() => setLinkDialogOpen(true)}
        >
          <Link2Icon className="size-4" />
          重新关联或创建客户
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

type ConversationInfoPanelProps = {
  conversation: AgentConversation | null;
  className?: string;
  variant?: "default" | "embedded";
};

export function ConversationInfoPanel({
  conversation,
  className,
  variant = "default",
}: ConversationInfoPanelProps) {
  const embedded = variant === "embedded";

  return (
    <div
      className={cn(
        "flex h-full min-h-0 flex-col overflow-hidden",
        embedded
          ? "bg-background text-foreground"
          : "border-l border-border bg-card text-card-foreground",
        className,
      )}
    >
      <div className="flex h-12.5 shrink-0 items-center border-b border-border px-3">
        <h2 className="text-sm font-medium text-foreground">会话信息</h2>
      </div>

      <div
        className={cn(
          "min-h-0 flex-1 overflow-y-auto px-3 pb-4",
          embedded && "pb-[max(1rem,env(safe-area-inset-bottom))] pt-1",
        )}
      >
        {!conversation ? (
          <p className="pt-4 text-sm text-muted-foreground">
            {embedded
              ? "请选择会话以查看会话信息"
              : "请选择左侧会话以查看会话信息"}
          </p>
        ) : (
          <div className="space-y-4 py-3">
            <CustomerBody conversation={conversation} />
          </div>
        )}
      </div>
    </div>
  );
}

function ConversationTagSection({
  conversation,
}: {
  conversation: AgentConversation;
}) {
  const setConversationTags = useAgentConversationsStore(
    (state) => state.setConversationTags,
  );
  const [availableTags, setAvailableTags] = useState<TagTree[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    let cancelled = false;

    async function loadTags() {
      setLoading(true);
      try {
        const data = await fetchTagsAll();
        if (!cancelled) {
          setAvailableTags(Array.isArray(data) ? data : []);
        }
      } catch (error) {
        if (!cancelled) {
          toast.error(error instanceof Error ? error.message : "加载标签失败");
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    }

    void loadTags();

    return () => {
      cancelled = true;
    };
  }, []);

  return (
    <section className="space-y-2 border-t pt-2">
      <SectionHeading
        action={
          <ConversationTagPicker
            conversation={conversation}
            availableTags={availableTags}
            loading={loading}
            onTagsChange={(tags) => {
              setConversationTags(conversation.id, tags);
            }}
          />
        }
      >
        会话标签
      </SectionHeading>
      <ConversationTagBadges
        tags={conversation.tags}
        availableTags={availableTags}
      />
      {!conversation.tags || conversation.tags.length === 0 ? (
        <p className="text-sm text-muted-foreground">暂未设置会话标签</p>
      ) : null}
    </section>
  );
}

function CustomerBody({ conversation }: { conversation: AgentConversation }) {
  const customerId = conversation.customerId ?? 0;

  if (customerId <= 0) {
    return (
      <div className="space-y-4">
        <UnlinkedCustomerEmpty conversation={conversation} />
        <ConversationTagSection conversation={conversation} />
      </div>
    );
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
      if (!c) {
        setContacts([]);
        return;
      }
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
      <div className="space-y-4">
        <MissingCustomerEmpty conversation={conversation} />
        <ConversationTagSection conversation={conversation} />
      </div>
    );
  }

  const displayName = customer.name.trim() || "未填写姓名";
  const company = customer.company ?? null;
  const genderLabel =
    customer.gender === Gender.Male || customer.gender === Gender.Female
      ? GenderLabels[customer.gender as Gender] ?? String(customer.gender)
      : null;

  return (
    <div className="space-y-4">
      {isProfileEmpty ? (
        <div className="rounded-lg bg-amber-500/10 px-3 py-2.5 text-xs leading-relaxed text-amber-950 dark:text-amber-100">
          客户主档已关联，但基础信息尚未填写。请点击「编辑」补全资料。
        </div>
      ) : null}

      <section className="space-y-2">
        <div className="flex items-start justify-between gap-2">
          <div className="flex min-w-0 flex-1 items-start gap-2 text-sm">
            <UserRoundIcon
              className="mt-0.5 size-4 shrink-0 text-muted-foreground"
              aria-hidden
            />
            <div className="min-w-0 flex-1 space-y-0.5">
              <p className="line-clamp-2 leading-snug text-foreground">
                <span className="font-medium">{displayName}</span>
                {genderLabel ? (
                  <span className="font-normal text-muted-foreground">
                    {" "}
                    · {genderLabel}
                  </span>
                ) : null}
              </p>
            </div>
          </div>
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

      <section className="space-y-2">
        {contacts.length === 0 ? (
          <p className="text-sm text-muted-foreground">暂无联系方式</p>
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
              );
            })}
          </ul>
        )}
      </section>

      {customer.companyId > 0 ? (
        <section className="border-t pt-2">
          {company ? (
            <div className="space-y-2">
              <div className="flex items-start justify-between gap-2">
                <div className="flex min-w-0 flex-1 items-start gap-2 text-sm">
                  <Building2Icon
                    className="mt-0.5 size-4 shrink-0 text-muted-foreground"
                    aria-hidden
                  />
                  <div className="min-w-0 flex-1 space-y-0.5">
                    <p className="line-clamp-2 font-medium leading-snug text-foreground">
                      {company.name}
                    </p>
                    {company.code ? (
                      <p className="font-mono text-xs text-muted-foreground">
                        {company.code}
                      </p>
                    ) : null}
                  </div>
                </div>
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

      <RelatedTicketsSection conversation={conversation} />

      <ConversationTagSection conversation={conversation} />

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

function RelatedTicketsSection({ conversation }: { conversation: AgentConversation }) {
  const [tickets, setTickets] = useState<TicketItem[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    let cancelled = false;
    async function loadTickets() {
      setLoading(true);
      try {
        const data = await fetchTickets({
          conversationId: conversation.id,
          page: 1,
          limit: 5,
        });
        if (!cancelled) {
          setTickets(Array.isArray(data.results) ? data.results : []);
        }
      } catch (error) {
        if (!cancelled) {
          toast.error(error instanceof Error ? error.message : "加载关联工单失败");
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    }
    void loadTickets();
    return () => {
      cancelled = true;
    };
  }, [conversation.id]);

  return (
    <section className="space-y-2 border-t pt-2">
      <SectionHeading>关联工单</SectionHeading>
      {loading ? (
        <p className="text-sm text-muted-foreground">加载工单中…</p>
      ) : tickets.length > 0 ? (
        <div className="space-y-2">
          {tickets.map((ticket) => (
            <Link
              key={ticket.id}
              href={`/tickets/detail?id=${ticket.id}`}
              target="_blank"
              rel="noreferrer"
              className="block rounded-lg border border-border bg-background px-3 py-2 transition-colors hover:bg-muted/40"
            >
              <div className="flex items-start justify-between gap-3">
                <div className="min-w-0 flex-1">
                  <div className="truncate text-sm font-medium text-foreground">
                    {ticket.title}
                  </div>
                  <div className="mt-0.5 text-xs text-muted-foreground">
                    {ticket.ticketNo}
                  </div>
                </div>
                <TicketPriorityBadge priority={ticket.priority} />
              </div>
              <div className="mt-2 flex items-center justify-between gap-3">
                <TicketStatusBadge status={ticket.status} />
                <span className="text-xs text-muted-foreground">
                  {ticket.updatedAt ? formatDateTime(ticket.updatedAt) : "—"}
                </span>
              </div>
            </Link>
          ))}
        </div>
      ) : (
        <p className="text-sm text-muted-foreground">当前会话暂无关联工单</p>
      )}
    </section>
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
              <Textarea
                id="co-remark"
                value={remark}
                onChange={(e) => setRemark(e.target.value)}
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
  );
}
