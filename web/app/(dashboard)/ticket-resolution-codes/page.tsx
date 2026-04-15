"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { PlusIcon, RefreshCwIcon, SearchIcon, Trash2Icon } from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { Controller, useForm, type Resolver } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod/v4";

import { ListPagination } from "@/components/list-pagination";
import { OptionCombobox } from "@/components/option-combobox";
import { ProjectDialog } from "@/components/project-dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Field,
  FieldContent,
  FieldError,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import {
  createTicketResolutionCode,
  deleteTicketResolutionCode,
  fetchTicketResolutionCodes,
  updateTicketResolutionCode,
  type CreateTicketResolutionCodePayload,
  type PageResult,
  type TicketResolutionCode,
} from "@/lib/api/ticket-config";
import { getEnumLabel, getEnumOptions } from "@/lib/enums";
import { Status, StatusLabels } from "@/lib/generated/enums";

const listStatusOptions = [
  { value: "all", label: "全部状态" },
  ...getEnumOptions(StatusLabels)
    .filter((item) => Number(item.value) !== Status.Deleted)
    .map((item) => ({ value: String(item.value), label: item.label })),
] as const;

const formSchema = z.object({
  name: z.string().trim().min(1, "解决码名称不能为空"),
  code: z.string().trim().min(1, "解决码编码不能为空"),
  sortNo: z
    .string()
    .trim()
    .min(1, "排序不能为空")
    .regex(/^\d+$/, "排序值必须是大于等于 0 的整数"),
  status: z.enum([String(Status.Ok), String(Status.Disabled)], {
    message: "请选择状态",
  }),
  remark: z.string().trim(),
});

type EditForm = z.infer<typeof formSchema>;

const resolver = zodResolver(formSchema as never) as Resolver<
  z.input<typeof formSchema>,
  undefined,
  z.output<typeof formSchema>
>;

const emptyForm: EditForm = {
  name: "",
  code: "",
  sortNo: "0",
  status: String(Status.Ok),
  remark: "",
};

function buildForm(item: TicketResolutionCode | null): EditForm {
  if (!item) {
    return emptyForm;
  }
  return {
    name: item.name,
    code: item.code,
    sortNo: String(item.sortNo),
    status: String(item.status) as EditForm["status"],
    remark: item.remark || "",
  };
}

function buildPayload(form: EditForm): CreateTicketResolutionCodePayload {
  return {
    name: form.name.trim(),
    code: form.code.trim(),
    sortNo: Number(form.sortNo),
    status: Number(form.status),
    remark: form.remark.trim(),
  };
}

export default function TicketResolutionCodesPage() {
  const [keywordInput, setKeywordInput] = useState("");
  const [statusFilterInput, setStatusFilterInput] = useState("all");
  const [keyword, setKeyword] = useState("");
  const [statusFilter, setStatusFilter] = useState("all");
  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState(20);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingItem, setEditingItem] = useState<TicketResolutionCode | null>(
    null,
  );
  const [result, setResult] = useState<PageResult<TicketResolutionCode>>({
    results: [],
    page: { page: 1, limit: 20, total: 0 },
  });

  const loadData = useCallback(async () => {
    setLoading(true);
    try {
      const data = await fetchTicketResolutionCodes({
        name: keyword.trim() || undefined,
        status: statusFilter === "all" ? undefined : statusFilter,
        page,
        limit,
      });
      setResult(data);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "加载解决码失败");
    } finally {
      setLoading(false);
    }
  }, [keyword, statusFilter, page, limit]);

  useEffect(() => {
    void loadData();
  }, [loadData]);

  function applyFilters() {
    setKeyword(keywordInput);
    setStatusFilter(statusFilterInput);
    setPage(1);
  }

  async function handleSubmit(payload: CreateTicketResolutionCodePayload) {
    if (saving) {
      return;
    }
    setSaving(true);
    try {
      if (editingItem) {
        await updateTicketResolutionCode({ id: editingItem.id, ...payload });
        toast.success(`已更新解决码：${payload.name}`);
      } else {
        await createTicketResolutionCode(payload);
        toast.success(`已创建解决码：${payload.name}`);
      }
      setDialogOpen(false);
      setEditingItem(null);
      await loadData();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "保存解决码失败");
    } finally {
      setSaving(false);
    }
  }

  async function handleDelete(item: TicketResolutionCode) {
    try {
      await deleteTicketResolutionCode(item.id);
      toast.success(`已删除解决码：${item.name}`);
      await loadData();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "删除解决码失败");
    }
  }

  return (
    <>
      <div className="flex flex-1 flex-col gap-6 p-4 lg:p-6">
        <div className="flex flex-col gap-3 xl:flex-row xl:items-center">
          <div className="relative min-w-72">
            <SearchIcon className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              value={keywordInput}
              onChange={(event) => setKeywordInput(event.target.value)}
              onKeyDown={(event) => {
                if (event.key === "Enter") {
                  event.preventDefault();
                  applyFilters();
                }
              }}
              placeholder="按解决码名称筛选"
              className="pl-9"
            />
          </div>
          <div className="w-full xl:w-40">
            <OptionCombobox
              value={statusFilterInput}
              onChange={setStatusFilterInput}
              placeholder="全部状态"
              options={listStatusOptions.map((item) => ({
                value: item.value,
                label: item.label,
              }))}
            />
          </div>
          <Button variant="outline" onClick={applyFilters} disabled={loading}>
            <SearchIcon className="size-4" />
            查询
          </Button>
          <Button
            variant="outline"
            onClick={() => void loadData()}
            disabled={loading}
          >
            <RefreshCwIcon className="size-4" />
          </Button>
          <Button
            onClick={() => {
              setEditingItem(null);
              setDialogOpen(true);
            }}
          >
            <PlusIcon className="size-4" />
            新建解决码
          </Button>
        </div>

        <div className="overflow-hidden rounded-lg border bg-background">
          <table className="w-full text-sm">
            <thead className="bg-muted/35">
              <tr>
                <th className="px-4 py-3 text-left font-medium">名称</th>
                <th className="px-4 py-3 text-left font-medium">编码</th>
                <th className="px-4 py-3 text-left font-medium">状态</th>
                <th className="px-4 py-3 text-left font-medium">排序</th>
                <th className="px-4 py-3 text-right font-medium">操作</th>
              </tr>
            </thead>
            <tbody>
              {loading ? (
                <tr>
                  <td
                    colSpan={5}
                    className="h-32 text-center text-muted-foreground"
                  >
                    加载中...
                  </td>
                </tr>
              ) : result.results.length > 0 ? (
                result.results.map((item) => (
                  <tr key={item.id} className="border-t">
                    <td className="px-4 py-3">{item.name}</td>
                    <td className="px-4 py-3 font-mono text-xs">{item.code}</td>
                    <td className="px-4 py-3">
                      <Badge
                        variant={
                          item.status === Status.Ok ? "default" : "secondary"
                        }
                      >
                        {getEnumLabel(StatusLabels, item.status as Status)}
                      </Badge>
                    </td>
                    <td className="px-4 py-3">{item.sortNo}</td>
                    <td className="px-4 py-3 text-right">
                      <DropdownMenu>
                        <DropdownMenuTrigger
                          render={<Button variant="ghost" size="sm" />}
                        >
                          操作
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem
                            onClick={() => {
                              setEditingItem(item);
                              setDialogOpen(true);
                            }}
                          >
                            编辑
                          </DropdownMenuItem>
                          <DropdownMenuItem
                            className="text-destructive"
                            onClick={() => void handleDelete(item)}
                          >
                            <Trash2Icon className="size-4" />
                            删除
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </td>
                  </tr>
                ))
              ) : (
                <tr>
                  <td
                    colSpan={5}
                    className="h-32 text-center text-muted-foreground"
                  >
                    暂无解决码
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>

        <ListPagination
          page={result.page.page}
          total={result.page.total}
          limit={result.page.limit}
          loading={loading}
          onPageChange={setPage}
          onLimitChange={(value) => {
            setLimit(value);
            setPage(1);
          }}
        />
      </div>

      <TicketResolutionCodeEditDialog
        open={dialogOpen}
        saving={saving}
        item={editingItem}
        onOpenChange={(open) => {
          if (!saving) {
            setDialogOpen(open);
            if (!open) {
              setEditingItem(null);
            }
          }
        }}
        onSubmit={handleSubmit}
      />
    </>
  );
}

type TicketResolutionCodeEditDialogProps = {
  open: boolean;
  saving: boolean;
  item: TicketResolutionCode | null;
  onOpenChange: (open: boolean) => void;
  onSubmit: (payload: CreateTicketResolutionCodePayload) => Promise<void>;
};

function TicketResolutionCodeEditDialog({
  open,
  saving,
  item,
  onOpenChange,
  onSubmit,
}: TicketResolutionCodeEditDialogProps) {
  const formId = "ticket-resolution-code-edit-form";
  const form = useForm<
    z.input<typeof formSchema>,
    undefined,
    z.output<typeof formSchema>
  >({
    resolver,
    defaultValues: emptyForm,
  });
  const {
    control,
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = form;

  useEffect(() => {
    reset(buildForm(item));
  }, [item, reset]);

  return (
    <ProjectDialog
      open={open}
      onOpenChange={onOpenChange}
      title={item ? "编辑解决码" : "新建解决码"}
      size="md"
      allowFullscreen
      footer={
        <>
          <Button
            type="button"
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={saving}
          >
            取消
          </Button>
          <Button type="submit" form={formId} disabled={saving}>
            {saving ? "保存中..." : item ? "保存" : "创建"}
          </Button>
        </>
      }
    >
      <form
        id={formId}
        onSubmit={handleSubmit(async (values) =>
          onSubmit(buildPayload(values)),
        )}
        className="space-y-4"
      >
        <Field data-invalid={!!errors.name}>
          <FieldLabel htmlFor="ticket-resolution-code-name">
            解决码名称
          </FieldLabel>
          <FieldContent>
            <Input
              id="ticket-resolution-code-name"
              placeholder="请输入解决码名称"
              {...register("name")}
            />
            <FieldError errors={[errors.name]} />
          </FieldContent>
        </Field>
        <Field data-invalid={!!errors.code}>
          <FieldLabel htmlFor="ticket-resolution-code-code">
            解决码编码
          </FieldLabel>
          <FieldContent>
            <Input
              id="ticket-resolution-code-code"
              placeholder="请输入解决码编码"
              {...register("code")}
            />
            <FieldError errors={[errors.code]} />
          </FieldContent>
        </Field>
        <div className="grid gap-4 md:grid-cols-2">
          <Field data-invalid={!!errors.sortNo}>
            <FieldLabel htmlFor="ticket-resolution-code-sort">排序</FieldLabel>
            <FieldContent>
              <Input id="ticket-resolution-code-sort" {...register("sortNo")} />
              <FieldError errors={[errors.sortNo]} />
            </FieldContent>
          </Field>
          <Field data-invalid={!!errors.status}>
            <FieldLabel>状态</FieldLabel>
            <FieldContent>
              <Controller
                control={control}
                name="status"
                render={({ field }) => (
                  <OptionCombobox
                    value={field.value}
                    onChange={field.onChange}
                    placeholder="请选择状态"
                    options={listStatusOptions
                      .filter((item) => item.value !== "all")
                      .map((item) => ({
                        value: item.value,
                        label: item.label,
                      }))}
                  />
                )}
              />
              <FieldError errors={[errors.status]} />
            </FieldContent>
          </Field>
        </div>
        <Field>
          <FieldLabel htmlFor="ticket-resolution-code-remark">备注</FieldLabel>
          <FieldContent>
            <Textarea
              id="ticket-resolution-code-remark"
              rows={4}
              {...register("remark")}
            />
          </FieldContent>
        </Field>
      </form>
    </ProjectDialog>
  );
}
