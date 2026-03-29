"use client";

import { PencilIcon, PlusIcon, UserRoundIcon } from "lucide-react";
import { useCallback, useEffect, useMemo, useState } from "react";
import { toast } from "sonner";

import type { AgentConversation } from "@/lib/api/agent";
import {
  createCustomerContact,
  fetchCustomerContacts,
  updateCustomerContact,
  type AdminCustomerContact,
} from "@/lib/api/customer-contact";
import { fetchCustomer, updateCustomer, type AdminCustomer } from "@/lib/api/customer";
import {
  fetchCompanies,
  fetchCompany,
  updateCompany,
  type AdminCompany,
} from "@/lib/api/company";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
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
import { OptionCombobox, type ComboboxOption } from "@/components/option-combobox";
import { formatDateTime } from "@/lib/utils";

const GENDER_LABELS: Record<number, string> = {
  0: "未知",
  1: "男",
  2: "女",
};

const STATUS_LABELS: Record<number, string> = {
  0: "启用",
  1: "禁用",
  2: "已删除",
};

const CONTACT_TYPE_OPTIONS: ComboboxOption[] = [
  { value: "mobile", label: "手机号" },
  { value: "email", label: "邮箱" },
  { value: "wechat", label: "微信" },
  { value: "other", label: "其他" },
];

const CONTACT_STATUS_OPTIONS: ComboboxOption[] = [
  { value: "0", label: "启用" },
  { value: "1", label: "禁用" },
];

function InfoRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex flex-col gap-0.5 py-2">
      <span className="text-xs text-muted-foreground">{label}</span>
      <span className="break-all text-sm text-foreground">{value || "-"}</span>
    </div>
  );
}

function SectionTitle({
  children,
  action,
}: {
  children: React.ReactNode;
  action?: React.ReactNode;
}) {
  return (
    <div className="flex items-center justify-between gap-2 pt-3 pb-1">
      <h3 className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
        {children}
      </h3>
      {action}
    </div>
  );
}

function UnlinkedCustomerEmpty({ conversation }: { conversation: AgentConversation }) {
  return (
    <div className="space-y-4 pt-2">
      <div className="flex flex-col items-center justify-center rounded-lg border border-dashed border-border bg-muted/20 px-4 py-8 text-center">
        <UserRoundIcon className="mb-2 size-10 text-muted-foreground" aria-hidden />
        <p className="text-sm font-medium text-foreground">尚未关联 CRM 客户</p>
        <p className="mt-1 max-w-xs text-xs leading-relaxed text-muted-foreground">
          当前会话未绑定客户主档（CustomerID 为 0）。访客仅通过外部身份接待；绑定后可在此维护公司与联系方式。
        </p>
      </div>
      <div className="divide-y divide-border rounded-lg border border-border">
        <section className="px-1">
          <SectionTitle>会话侧访客标识</SectionTitle>
          <InfoRow label="外部来源" value={conversation.externalSource} />
          <InfoRow label="外部用户标识" value={conversation.externalId} />
        </section>
      </div>
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
  const [company, setCompany] = useState<AdminCompany | null>(null);
  const [contacts, setContacts] = useState<AdminCustomerContact[]>([]);

  const [customerEditOpen, setCustomerEditOpen] = useState(false);
  const [companyEditOpen, setCompanyEditOpen] = useState(false);
  const [contactDialogOpen, setContactDialogOpen] = useState(false);
  const [editingContact, setEditingContact] = useState<AdminCustomerContact | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const c = await fetchCustomer(customerId);
      setCustomer(c);
      if (c.companyId > 0) {
        const co = await fetchCompany(c.companyId);
        setCompany(co);
      } else {
        setCompany(null);
      }
      const list = await fetchCustomerContacts(customerId);
      setContacts(Array.isArray(list) ? list : []);
    } catch (e) {
      const msg = e instanceof Error ? e.message : "加载客户信息失败";
      toast.error(msg);
      setCustomer(null);
      setCompany(null);
      setContacts([]);
    } finally {
      setLoading(false);
    }
  }, [customerId]);

  useEffect(() => {
    void load();
  }, [load]);

  const openCreateContact = () => {
    setEditingContact(null);
    setContactDialogOpen(true);
  };

  const openEditContact = (row: AdminCustomerContact) => {
    setEditingContact(row);
    setContactDialogOpen(true);
  };

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

  return (
    <div className="space-y-1">
      {isProfileEmpty ? (
        <div className="mb-3 rounded-lg border border-amber-500/30 bg-amber-500/5 px-3 py-2 text-xs text-amber-950 dark:text-amber-100">
          客户主档已关联，但基础信息尚未填写。请点击下方「编辑客户」补全资料。
        </div>
      ) : null}

      <div className="divide-y divide-border">
        <section>
          <SectionTitle
            action={
              <Button
                type="button"
                variant="ghost"
                size="sm"
                className="h-7 gap-1 px-2 text-xs"
                onClick={() => setCustomerEditOpen(true)}
              >
                <PencilIcon className="size-3.5" />
                编辑客户
              </Button>
            }
          >
            客户信息
          </SectionTitle>
          <InfoRow label="姓名" value={customer.name} />
          <InfoRow label="性别" value={GENDER_LABELS[customer.gender] ?? String(customer.gender)} />
          <InfoRow
            label="所属公司"
            value={
              customer.companyId > 0
                ? company?.name || `公司 ID ${customer.companyId}`
                : "无"
            }
          />
          <InfoRow
            label="最近活跃"
            value={customer.lastActiveAt ? formatDateTime(customer.lastActiveAt) : "-"}
          />
          <InfoRow label="主手机号" value={customer.primaryMobile} />
          <InfoRow label="主邮箱" value={customer.primaryEmail} />
          <InfoRow label="状态" value={STATUS_LABELS[customer.status] ?? String(customer.status)} />
          <InfoRow label="创建时间" value={formatDateTime(customer.createdAt)} />
          <InfoRow label="更新时间" value={formatDateTime(customer.updatedAt)} />
          <div className="flex flex-col gap-0.5 py-2">
            <span className="text-xs text-muted-foreground">备注</span>
            <p className="whitespace-pre-wrap break-all text-sm text-foreground">
              {customer.remark || "-"}
            </p>
          </div>
        </section>

        <section>
          <SectionTitle
            action={
              customer.companyId > 0 && company ? (
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  className="h-7 gap-1 px-2 text-xs"
                  onClick={() => setCompanyEditOpen(true)}
                >
                  <PencilIcon className="size-3.5" />
                  编辑公司
                </Button>
              ) : null
            }
          >
            客户公司
          </SectionTitle>
          {customer.companyId <= 0 ? (
            <p className="py-2 text-sm text-muted-foreground">
              未关联公司。可在「编辑客户」中选择所属公司。
            </p>
          ) : company ? (
            <>
              <InfoRow label="公司名称" value={company.name} />
              <InfoRow label="公司编码" value={company.code} />
              <InfoRow label="状态" value={STATUS_LABELS[company.status] ?? String(company.status)} />
              <InfoRow label="创建时间" value={formatDateTime(company.createdAt)} />
              <InfoRow label="更新时间" value={formatDateTime(company.updatedAt)} />
              <div className="flex flex-col gap-0.5 py-2">
                <span className="text-xs text-muted-foreground">备注</span>
                <p className="whitespace-pre-wrap break-all text-sm text-foreground">
                  {company.remark || "-"}
                </p>
              </div>
            </>
          ) : (
            <p className="py-2 text-sm text-muted-foreground">公司信息加载失败或公司已删除。</p>
          )}
        </section>

        <section>
          <SectionTitle
            action={
              <Button
                type="button"
                variant="ghost"
                size="sm"
                className="h-7 gap-1 px-2 text-xs"
                onClick={openCreateContact}
              >
                <PlusIcon className="size-3.5" />
                新增联系方式
              </Button>
            }
          >
            联系方式
          </SectionTitle>
          {contacts.length === 0 ? (
            <p className="py-2 text-sm text-muted-foreground">暂无联系方式记录。</p>
          ) : (
            <ul className="space-y-2 py-2">
              {contacts.map((row) => (
                <li
                  key={row.id}
                  className="flex items-start justify-between gap-2 rounded-md border border-border bg-muted/20 px-2 py-2"
                >
                  <div className="min-w-0 flex-1 text-sm">
                    <div className="flex flex-wrap items-center gap-1.5">
                      <span className="font-medium text-foreground">
                        {CONTACT_TYPE_OPTIONS.find((o) => o.value === row.contactType)?.label ??
                          row.contactType}
                      </span>
                      {row.isPrimary ? (
                        <span className="rounded bg-primary/10 px-1.5 py-px text-[10px] font-medium text-primary">
                          主
                        </span>
                      ) : null}
                      {row.isVerified ? (
                        <span className="text-[10px] text-muted-foreground">已验证</span>
                      ) : null}
                    </div>
                    <div className="mt-0.5 break-all text-foreground">{row.contactValue}</div>
                    <div className="mt-1 text-xs text-muted-foreground">
                      来源 {row.source || "-"} · 状态{" "}
                      {STATUS_LABELS[row.status] ?? row.status}
                      {row.verifiedAt ? ` · 验证于 ${formatDateTime(row.verifiedAt)}` : ""}
                      {row.remark ? ` · ${row.remark}` : ""}
                    </div>
                    <div className="mt-0.5 text-[11px] text-muted-foreground/80">
                      创建 {formatDateTime(row.createdAt)} · 更新 {formatDateTime(row.updatedAt)}
                    </div>
                  </div>
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    className="h-7 shrink-0 px-2 text-xs"
                    onClick={() => openEditContact(row)}
                  >
                    编辑
                  </Button>
                </li>
              ))}
            </ul>
          )}
        </section>
      </div>

      <CustomerEditDialog
        open={customerEditOpen}
        onOpenChange={setCustomerEditOpen}
        customer={customer}
        onSaved={() => {
          void load();
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
      <ContactEditDialog
        open={contactDialogOpen}
        onOpenChange={setContactDialogOpen}
        customerId={customerId}
        editing={editingContact}
        onSaved={() => {
          void load();
        }}
      />
    </div>
  );
}

type CustomerEditDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  customer: AdminCustomer;
  onSaved: () => void;
};

function CustomerEditDialog({
  open,
  onOpenChange,
  customer,
  onSaved,
}: CustomerEditDialogProps) {
  const [name, setName] = useState("");
  const [gender, setGender] = useState("0");
  const [companyId, setCompanyId] = useState("0");
  const [primaryMobile, setPrimaryMobile] = useState("");
  const [primaryEmail, setPrimaryEmail] = useState("");
  const [remark, setRemark] = useState("");
  const [companyOptions, setCompanyOptions] = useState<ComboboxOption[]>([
    { value: "0", label: "无" },
  ]);
  const [saving, setSaving] = useState(false);

  const genderOptions = useMemo<ComboboxOption[]>(
    () => [
      { value: "0", label: GENDER_LABELS[0] },
      { value: "1", label: GENDER_LABELS[1] },
      { value: "2", label: GENDER_LABELS[2] },
    ],
    []
  );

  useEffect(() => {
    if (!open) {
      return;
    }
    setName(customer.name);
    setGender(String(customer.gender));
    setCompanyId(String(customer.companyId || 0));
    setPrimaryMobile(customer.primaryMobile);
    setPrimaryEmail(customer.primaryEmail);
    setRemark(customer.remark);
    let cancelled = false;
    void (async () => {
      try {
        const page = await fetchCompanies({ limit: 200, page: 1 });
        if (cancelled) {
          return;
        }
        const opts: ComboboxOption[] = [
          { value: "0", label: "无" },
          ...page.results.map((c) => ({
            value: String(c.id),
            label: c.name || `公司 #${c.id}`,
          })),
        ];
        setCompanyOptions(opts);
      } catch {
        if (!cancelled) {
          setCompanyOptions([{ value: "0", label: "无" }]);
        }
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [open, customer]);

  const handleSubmit = async () => {
    const trimmed = name.trim();
    if (!trimmed) {
      toast.error("姓名不能为空");
      return;
    }
    setSaving(true);
    try {
      await updateCustomer({
        id: customer.id,
        name: trimmed,
        gender: Number.parseInt(gender, 10) || 0,
        companyId: Number.parseInt(companyId, 10) || 0,
        primaryMobile: primaryMobile.trim(),
        primaryEmail: primaryEmail.trim(),
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
      <DialogContent className="max-h-[min(90vh,560px)] overflow-y-auto sm:max-w-md" showCloseButton>
        <DialogHeader>
          <DialogTitle>编辑客户</DialogTitle>
        </DialogHeader>
        <div className="flex flex-col gap-4 py-1">
          <Field orientation="vertical">
            <FieldLabel htmlFor="cust-name">姓名</FieldLabel>
            <FieldContent>
              <Input
                id="cust-name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                autoComplete="off"
              />
            </FieldContent>
          </Field>
          <Field orientation="vertical">
            <FieldLabel>性别</FieldLabel>
            <FieldContent>
              <OptionCombobox
                value={gender}
                options={genderOptions}
                placeholder="性别"
                onChange={setGender}
              />
            </FieldContent>
          </Field>
          <Field orientation="vertical">
            <FieldLabel>所属公司</FieldLabel>
            <FieldContent>
              <OptionCombobox
                value={companyId}
                options={companyOptions}
                placeholder="选择公司"
                onChange={setCompanyId}
              />
            </FieldContent>
          </Field>
          <Field orientation="vertical">
            <FieldLabel htmlFor="cust-mobile">主手机号</FieldLabel>
            <FieldContent>
              <Input
                id="cust-mobile"
                value={primaryMobile}
                onChange={(e) => setPrimaryMobile(e.target.value)}
                autoComplete="off"
              />
            </FieldContent>
          </Field>
          <Field orientation="vertical">
            <FieldLabel htmlFor="cust-email">主邮箱</FieldLabel>
            <FieldContent>
              <Input
                id="cust-email"
                value={primaryEmail}
                onChange={(e) => setPrimaryEmail(e.target.value)}
                autoComplete="off"
              />
            </FieldContent>
          </Field>
          <Field orientation="vertical">
            <FieldLabel htmlFor="cust-remark">备注</FieldLabel>
            <FieldContent>
              <Textarea
                id="cust-remark"
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

type ContactEditDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  customerId: number;
  editing: AdminCustomerContact | null;
  onSaved: () => void;
};

function ContactEditDialog({
  open,
  onOpenChange,
  customerId,
  editing,
  onSaved,
}: ContactEditDialogProps) {
  const [contactType, setContactType] = useState("mobile");
  const [contactValue, setContactValue] = useState("");
  const [isPrimary, setIsPrimary] = useState(false);
  const [isVerified, setIsVerified] = useState(false);
  const [source, setSource] = useState("manual");
  const [status, setStatus] = useState("0");
  const [remark, setRemark] = useState("");
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (!open) {
      return;
    }
    if (editing) {
      setContactType(editing.contactType || "mobile");
      setContactValue(editing.contactValue);
      setIsPrimary(editing.isPrimary);
      setIsVerified(editing.isVerified);
      setSource(editing.source || "manual");
      setStatus(String(editing.status));
      setRemark(editing.remark);
    } else {
      setContactType("mobile");
      setContactValue("");
      setIsPrimary(false);
      setIsVerified(false);
      setSource("manual");
      setStatus("0");
      setRemark("");
    }
  }, [open, editing]);

  const handleSubmit = async () => {
    const val = contactValue.trim();
    if (!val) {
      toast.error("联系方式不能为空");
      return;
    }
    setSaving(true);
    try {
      if (editing) {
        await updateCustomerContact({
          id: editing.id,
          contactType,
          contactValue: val,
          isPrimary,
          isVerified,
          source: source.trim() || "manual",
          status: Number.parseInt(status, 10) || 0,
          remark: remark.trim(),
        });
      } else {
        await createCustomerContact({
          customerId,
          contactType,
          contactValue: val,
          isPrimary,
          isVerified,
          source: source.trim() || "manual",
          status: Number.parseInt(status, 10) || 0,
          remark: remark.trim(),
        });
      }
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
      <DialogContent className="max-h-[min(90vh,520px)] overflow-y-auto sm:max-w-md" showCloseButton>
        <DialogHeader>
          <DialogTitle>{editing ? "编辑联系方式" : "新增联系方式"}</DialogTitle>
        </DialogHeader>
        <div className="flex flex-col gap-4 py-1">
          <Field orientation="vertical">
            <FieldLabel>类型</FieldLabel>
            <FieldContent>
              <OptionCombobox
                value={contactType}
                options={CONTACT_TYPE_OPTIONS}
                placeholder="类型"
                onChange={setContactType}
              />
            </FieldContent>
          </Field>
          <Field orientation="vertical">
            <FieldLabel htmlFor="cc-value">号码 / 地址</FieldLabel>
            <FieldContent>
              <Input
                id="cc-value"
                value={contactValue}
                onChange={(e) => setContactValue(e.target.value)}
                autoComplete="off"
              />
            </FieldContent>
          </Field>
          <div className="flex flex-col gap-3">
            <label className="flex cursor-pointer items-center gap-2 text-sm">
              <Checkbox checked={isPrimary} onCheckedChange={(v) => setIsPrimary(Boolean(v))} />
              主联系方式
            </label>
            <label className="flex cursor-pointer items-center gap-2 text-sm">
              <Checkbox checked={isVerified} onCheckedChange={(v) => setIsVerified(Boolean(v))} />
              已验证
            </label>
          </div>
          <Field orientation="vertical">
            <FieldLabel htmlFor="cc-source">来源</FieldLabel>
            <FieldContent>
              <Input
                id="cc-source"
                value={source}
                onChange={(e) => setSource(e.target.value)}
                placeholder="manual / import / system"
              />
            </FieldContent>
          </Field>
          <Field orientation="vertical">
            <FieldLabel>状态</FieldLabel>
            <FieldContent>
              <OptionCombobox
                value={status}
                options={CONTACT_STATUS_OPTIONS}
                placeholder="状态"
                onChange={setStatus}
              />
            </FieldContent>
          </Field>
          <Field orientation="vertical">
            <FieldLabel htmlFor="cc-remark">备注</FieldLabel>
            <FieldContent>
              <Textarea id="cc-remark" value={remark} onChange={(e) => setRemark(e.target.value)} rows={2} />
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
