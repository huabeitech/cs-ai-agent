"use client";

import { useEffect, useState } from "react";
import { zodResolver } from "@hookform/resolvers/zod";
import { Controller, Resolver, useForm } from "react-hook-form";
import { z } from "zod/v4";

import { ProjectDialog } from "@/components/project-dialog";
import { Button } from "@/components/ui/button";
import {
  Field,
  FieldContent,
  FieldError,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import {
  type AdminQuickReply,
  type CreateAdminQuickReplyPayload,
  fetchQuickReply,
} from "@/lib/api/admin";
import { getEnumLabel, getEnumOptions } from "@/lib/enums";
import { Status, StatusLabels } from "@/lib/generated/enums";

type QuickReplyFormDialogProps = {
  open: boolean;
  saving: boolean;
  itemId: number | null;
  onOpenChange: (open: boolean) => void;
  onSubmit: (payload: CreateAdminQuickReplyPayload) => Promise<void>;
};

const emptyForm: EditForm = {
  groupName: "",
  title: "",
  content: "",
  status: String(Status.Ok),
  sortNo: "0",
};

const formStatusOptions = getEnumOptions(StatusLabels).filter(
  (item) => Number(item.value) !== Status.Deleted,
);

const quickReplyFormSchema = z.object({
  groupName: z.string().trim().min(1, "分组名称不能为空"),
  title: z.string().trim().min(1, "标题不能为空"),
  content: z.string().trim().min(1, "回复内容不能为空"),
  status: z.enum([String(Status.Ok), String(Status.Disabled)], {
    message: "请选择状态",
  }),
  sortNo: z
    .string()
    .trim()
    .min(1, "排序不能为空")
    .regex(/^\d+$/, "排序值必须是大于等于 0 的整数"),
});

type EditForm = z.infer<typeof quickReplyFormSchema>;
const editFormResolver = zodResolver(quickReplyFormSchema as never) as Resolver<
  z.input<typeof quickReplyFormSchema>,
  undefined,
  z.output<typeof quickReplyFormSchema>
>;

function getStatusLabel(value: string) {
  return getEnumLabel(StatusLabels, Number(value) as Status);
}

function buildForm(item: AdminQuickReply | null): EditForm {
  if (!item) {
    return emptyForm;
  }

  return {
    groupName: item.groupName,
    title: item.title,
    content: item.content,
    status: String(item.status) as EditForm["status"],
    sortNo: String(item.sortNo),
  };
}

function buildPayload(form: EditForm): CreateAdminQuickReplyPayload {
  return {
    groupName: form.groupName.trim(),
    title: form.title.trim(),
    content: form.content.trim(),
    status: Number(form.status) as Status,
    sortNo: Number(form.sortNo),
  };
}

export function EditDialog({
  open,
  saving,
  itemId,
  onOpenChange,
  onSubmit,
}: QuickReplyFormDialogProps) {
  if (!open) {
    return null;
  }

  return (
    <QuickReplyFormDialogBody
      key={itemId ? `edit-${itemId}` : "create"}
      open={open}
      itemId={itemId}
      saving={saving}
      onOpenChange={onOpenChange}
      onSubmit={onSubmit}
    />
  );
}

type QuickReplyFormDialogBodyProps = {
  open: boolean;
  saving: boolean;
  itemId: number | null;
  onOpenChange: (open: boolean) => void;
  onSubmit: (payload: CreateAdminQuickReplyPayload) => Promise<void>;
};

function QuickReplyFormDialogBody({
  open,
  saving,
  itemId,
  onOpenChange,
  onSubmit,
}: QuickReplyFormDialogBodyProps) {
  const formId = "quick-reply-edit-form";
  const [loading, setLoading] = useState(false);
  const form = useForm<
    z.input<typeof quickReplyFormSchema>,
    undefined,
    z.output<typeof quickReplyFormSchema>
  >({
    resolver: editFormResolver,
    defaultValues: emptyForm,
  });
  const {
    control,
    handleSubmit,
    reset,
    register,
    formState: { errors },
  } = form;

  useEffect(() => {
    async function loadDetail() {
      if (!itemId) {
        reset(emptyForm);
        return;
      }
      setLoading(true);
      try {
        const data = await fetchQuickReply(itemId);
        reset(buildForm(data));
      } catch (error) {
        console.error("Failed to load quick reply:", error);
      } finally {
        setLoading(false);
      }
    }
    void loadDetail();
  }, [itemId, reset]);

  async function onFormSubmit(values: EditForm) {
    const payload = buildPayload(values);
    await onSubmit(payload);
  }

  return (
    <ProjectDialog
      open={open}
      onOpenChange={onOpenChange}
      title={itemId ? "编辑" : "新建"}
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
          <Button type="submit" form={formId} disabled={saving || loading}>
            {saving ? "保存中..." : itemId ? "保存" : "创建"}
          </Button>
        </>
      }
    >
      {loading ? (
        <div className="flex items-center justify-center py-12">
          <div className="text-muted-foreground">加载中...</div>
        </div>
      ) : (
        <form id={formId} onSubmit={handleSubmit(onFormSubmit)} className="space-y-4">
          <Field data-invalid={!!errors.groupName}>
            <FieldLabel htmlFor="quick-reply-group-name">分组名称</FieldLabel>
            <FieldContent>
              <Input
                id="quick-reply-group-name"
                placeholder="例如：售前、售后、催单"
                aria-invalid={!!errors.groupName}
                {...register("groupName")}
              />
              <FieldError errors={[errors.groupName]} />
            </FieldContent>
          </Field>
          <Field data-invalid={!!errors.title}>
            <FieldLabel htmlFor="quick-reply-title">标题</FieldLabel>
            <FieldContent>
              <Input
                id="quick-reply-title"
                placeholder="请输入快捷回复标题"
                aria-invalid={!!errors.title}
                {...register("title")}
              />
              <FieldError errors={[errors.title]} />
            </FieldContent>
          </Field>
          <Field data-invalid={!!errors.content}>
            <FieldLabel htmlFor="quick-reply-content">回复内容</FieldLabel>
            <FieldContent>
              <Textarea
                id="quick-reply-content"
                placeholder="请输入回复内容"
                rows={6}
                aria-invalid={!!errors.content}
                {...register("content")}
              />
              <FieldError errors={[errors.content]} />
            </FieldContent>
          </Field>
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <Field data-invalid={!!errors.status}>
              <FieldLabel htmlFor="quick-reply-status">状态</FieldLabel>
              <FieldContent>
                <Controller
                  control={control}
                  name="status"
                  render={({ field }) => (
                    <Select
                      value={field.value}
                      onValueChange={field.onChange}
                      modal={false}
                    >
                      <SelectTrigger
                        id="quick-reply-status"
                        className="w-full"
                        aria-invalid={!!errors.status}
                      >
                        <SelectValue>{getStatusLabel(field.value)}</SelectValue>
                      </SelectTrigger>
                      <SelectContent>
                        {formStatusOptions.map((option) => (
                          <SelectItem
                            key={String(option.value)}
                            value={String(option.value)}
                          >
                            {option.label}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  )}
                />
                <FieldError errors={[errors.status]} />
              </FieldContent>
            </Field>
            <Field data-invalid={!!errors.sortNo}>
              <FieldLabel htmlFor="quick-reply-sort-no">排序</FieldLabel>
              <FieldContent>
                <Input
                  id="quick-reply-sort-no"
                  type="number"
                  min={0}
                  step={1}
                  placeholder="数字越大越靠前"
                  aria-invalid={!!errors.sortNo}
                  {...register("sortNo")}
                />
                <FieldError errors={[errors.sortNo]} />
              </FieldContent>
            </Field>
          </div>
        </form>
      )}
    </ProjectDialog>
  );
}
