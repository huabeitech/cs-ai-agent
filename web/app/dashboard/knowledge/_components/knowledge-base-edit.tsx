"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useEffect, useState } from "react";
import { Controller, Resolver, useForm } from "react-hook-form";
import { z } from "zod/v4";

import { OptionCombobox } from "@/components/option-combobox";
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
  fetchKnowledgeBase,
  type CreateKnowledgeBasePayload,
  type KnowledgeBase,
} from "@/lib/api/admin";
import { getEnumLabel, getEnumOptions } from "@/lib/enums";
import {
  KnowledgeAnswerMode,
  KnowledgeAnswerModeLabels,
  KnowledgeBaseType,
  KnowledgeBaseTypeLabels,
  KnowledgeChunkProvider,
  KnowledgeChunkProviderLabels,
} from "@/lib/generated/enums";

type KnowledgeBaseEditDialogProps = {
  open: boolean;
  saving: boolean;
  itemId: number | null;
  onOpenChange: (open: boolean) => void;
  onSubmit: (payload: CreateKnowledgeBasePayload) => Promise<void>;
};

const emptyForm: EditForm = {
  name: "",
  description: "",
  knowledgeType: KnowledgeBaseType.Document,
  defaultTopK: "5",
  defaultScoreThreshold: "0.2",
  defaultRerankLimit: "10",
  chunkProvider: KnowledgeChunkProvider.Structured,
  chunkTargetTokens: "300",
  chunkMaxTokens: "400",
  chunkOverlapTokens: "40",
  answerMode: String(KnowledgeAnswerMode.Strict),
  remark: "",
};

const knowledgeBaseFormSchema = z.object({
  name: z.string().trim().min(1, "名称不能为空").max(100, "名称最多100个字符"),
  description: z.string().trim().max(500, "描述最多500个字符"),
  knowledgeType: z.string().trim().min(1, "请选择知识库类型"),
  defaultTopK: z.string().trim().min(1, "请输入TopK值"),
  defaultScoreThreshold: z.string().trim().min(1, "请输入分数阈值"),
  defaultRerankLimit: z.string().trim().min(1, "请输入重排序限制"),
  chunkProvider: z.string().trim().min(1, "请选择分块策略"),
  chunkTargetTokens: z.string().trim().min(1, "请输入目标 token 数"),
  chunkMaxTokens: z.string().trim().min(1, "请输入最大 token 数"),
  chunkOverlapTokens: z.string().trim().min(1, "请输入重叠 token 数"),
  answerMode: z.string().trim().min(1, "请选择回答模式"),
  remark: z.string().trim().max(500, "备注最多500个字符"),
});

type EditForm = z.infer<typeof knowledgeBaseFormSchema>;
const editFormResolver = zodResolver(
  knowledgeBaseFormSchema as never,
) as Resolver<
  z.input<typeof knowledgeBaseFormSchema>,
  undefined,
  z.output<typeof knowledgeBaseFormSchema>
>;

function buildForm(item: KnowledgeBase | null): EditForm {
  if (!item) {
    return emptyForm;
  }

  return {
    name: item.name,
    description: item.description || "",
    knowledgeType: item.knowledgeType || KnowledgeBaseType.Document,
    defaultTopK: String(item.defaultTopK),
    defaultScoreThreshold: String(item.defaultScoreThreshold),
    defaultRerankLimit: String(item.defaultRerankLimit),
    chunkProvider: item.chunkProvider,
    chunkTargetTokens: String(item.chunkTargetTokens),
    chunkMaxTokens: String(item.chunkMaxTokens),
    chunkOverlapTokens: String(item.chunkOverlapTokens),
    answerMode: String(item.answerMode),
    remark: item.remark || "",
  };
}

function buildPayload(form: EditForm): CreateKnowledgeBasePayload {
  return {
    name: form.name.trim(),
    description: form.description.trim(),
    knowledgeType: form.knowledgeType,
    defaultTopK: Number(form.defaultTopK),
    defaultScoreThreshold: Number(form.defaultScoreThreshold),
    defaultRerankLimit: Number(form.defaultRerankLimit),
    chunkProvider: form.chunkProvider,
    chunkTargetTokens: Number(form.chunkTargetTokens),
    chunkMaxTokens: Number(form.chunkMaxTokens),
    chunkOverlapTokens: Number(form.chunkOverlapTokens),
    answerMode: Number(form.answerMode),
    remark: form.remark.trim(),
  };
}

export function EditDialog({
  open,
  saving,
  itemId,
  onOpenChange,
  onSubmit,
}: KnowledgeBaseEditDialogProps) {
  if (!open) {
    return null;
  }

  return (
    <KnowledgeBaseFormDialogBody
      key={itemId ? `edit-${itemId}` : "create"}
      open={open}
      itemId={itemId}
      saving={saving}
      onOpenChange={onOpenChange}
      onSubmit={onSubmit}
    />
  );
}

type KnowledgeBaseFormDialogBodyProps = {
  open: boolean;
  saving: boolean;
  itemId: number | null;
  onOpenChange: (open: boolean) => void;
  onSubmit: (payload: CreateKnowledgeBasePayload) => Promise<void>;
};

function KnowledgeBaseFormDialogBody({
  open,
  saving,
  itemId,
  onOpenChange,
  onSubmit,
}: KnowledgeBaseFormDialogBodyProps) {
  const formId = "knowledge-base-edit-form";
  const [loading, setLoading] = useState(false);
  const form = useForm<
    z.input<typeof knowledgeBaseFormSchema>,
    undefined,
    z.output<typeof knowledgeBaseFormSchema>
  >({
    resolver: editFormResolver,
    defaultValues: buildForm(null),
  });
  const {
    control,
    handleSubmit,
    reset,
    register,
    watch,
    formState: { errors },
  } = form;
  const knowledgeType = watch("knowledgeType");
  const isFAQKnowledgeBase = knowledgeType === KnowledgeBaseType.FAQ;

  useEffect(() => {
    if (!open) {
      return;
    }

    if (itemId === null) {
      reset(buildForm(null));
      return;
    }

    let cancelled = false;

    async function loadItem() {
      try {
        setLoading(true);
        const data = await fetchKnowledgeBase(itemId!);
        if (!cancelled) {
          reset(buildForm(data));
        }
      } catch (error) {
        if (!cancelled) {
          console.error("Failed to load knowledge base:", error);
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    }

    void loadItem();

    return () => {
      cancelled = true;
    };
  }, [open, itemId, reset]);

  async function onFormSubmit(values: EditForm) {
    const payload = buildPayload(values);
    await onSubmit(payload);
  }

  return (
    <ProjectDialog
      open={open}
      onOpenChange={onOpenChange}
      title={itemId ? "编辑知识库" : "新建知识库"}
      size="lg"
      footer={
        <>
          <Button
            type="button"
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={saving || loading}
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
        <div className="flex items-center justify-center py-8">
          <div className="text-sm text-muted-foreground">加载中...</div>
        </div>
      ) : (
        <form
          id={formId}
          onSubmit={handleSubmit(onFormSubmit)}
          className="space-y-4"
        >
          <Field data-invalid={!!errors.knowledgeType}>
            <FieldLabel htmlFor="kb-knowledge-type">知识库类型</FieldLabel>
            <FieldContent>
              <Controller
                control={control}
                name="knowledgeType"
                render={({ field }) => (
                  <OptionCombobox
                    value={field.value}
                    options={getEnumOptions(KnowledgeBaseTypeLabels).map((option) => ({
                      value: String(option.value),
                      label: option.label,
                    }))}
                    placeholder="选择知识库类型"
                    searchPlaceholder="搜索知识库类型"
                    emptyText="没有匹配的知识库类型"
                    onChange={field.onChange}
                  />
                )}
              />
              <FieldError errors={[errors.knowledgeType]} />
            </FieldContent>
          </Field>

          <Field data-invalid={!!errors.name}>
            <FieldLabel htmlFor="kb-name">名称</FieldLabel>
            <FieldContent>
              <Input
                id="kb-name"
                placeholder="知识库名称"
                aria-invalid={!!errors.name}
                {...register("name")}
              />
              <FieldError errors={[errors.name]} />
            </FieldContent>
          </Field>

          <Field data-invalid={!!errors.description}>
            <FieldLabel htmlFor="kb-description">描述</FieldLabel>
            <FieldContent>
              <Textarea
                id="kb-description"
                placeholder="知识库描述"
                rows={3}
                aria-invalid={!!errors.description}
                {...register("description")}
              />
              <FieldError errors={[errors.description]} />
            </FieldContent>
          </Field>

          {!isFAQKnowledgeBase ? (
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-4">
            <Field data-invalid={!!errors.chunkProvider}>
              <FieldLabel htmlFor="kb-chunk-provider">分块策略</FieldLabel>
              <FieldContent>
                <Controller
                  control={control}
                  name="chunkProvider"
                  render={({ field }) => (
                    <OptionCombobox
                      value={field.value}
                      options={getEnumOptions(KnowledgeChunkProviderLabels)
                        .filter((option) => option.value !== KnowledgeChunkProvider.FAQ)
                        .map((option) => ({
                          value: String(option.value),
                          label: option.label,
                        }))}
                      placeholder="选择分块策略"
                      searchPlaceholder="搜索分块策略"
                      emptyText="没有匹配的分块策略"
                      onChange={field.onChange}
                    />
                  )}
                />
                <FieldError errors={[errors.chunkProvider]} />
              </FieldContent>
            </Field>
            <Field data-invalid={!!errors.chunkTargetTokens}>
              <FieldLabel htmlFor="kb-chunk-target-tokens">
                目标 Token
              </FieldLabel>
              <FieldContent>
                <Input
                  id="kb-chunk-target-tokens"
                  type="number"
                  min="1"
                  max="2000"
                  aria-invalid={!!errors.chunkTargetTokens}
                  {...register("chunkTargetTokens")}
                />
                <FieldError errors={[errors.chunkTargetTokens]} />
              </FieldContent>
            </Field>
            <Field data-invalid={!!errors.chunkMaxTokens}>
              <FieldLabel htmlFor="kb-chunk-max-tokens">最大 Token</FieldLabel>
              <FieldContent>
                <Input
                  id="kb-chunk-max-tokens"
                  type="number"
                  min="1"
                  max="4000"
                  aria-invalid={!!errors.chunkMaxTokens}
                  {...register("chunkMaxTokens")}
                />
                <FieldError errors={[errors.chunkMaxTokens]} />
              </FieldContent>
            </Field>
            <Field data-invalid={!!errors.chunkOverlapTokens}>
              <FieldLabel htmlFor="kb-chunk-overlap-tokens">
                重叠 Token
              </FieldLabel>
              <FieldContent>
                <Input
                  id="kb-chunk-overlap-tokens"
                  type="number"
                  min="0"
                  max="500"
                  aria-invalid={!!errors.chunkOverlapTokens}
                  {...register("chunkOverlapTokens")}
                />
                <FieldError errors={[errors.chunkOverlapTokens]} />
              </FieldContent>
            </Field>
            </div>
          ) : null}

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-4">
            <Field data-invalid={!!errors.defaultTopK}>
              <FieldLabel htmlFor="kb-default-top-k">默认TopK</FieldLabel>
              <FieldContent>
                <Input
                  id="kb-default-top-k"
                  type="number"
                  min="1"
                  max="100"
                  aria-invalid={!!errors.defaultTopK}
                  {...register("defaultTopK")}
                />
                <FieldError errors={[errors.defaultTopK]} />
              </FieldContent>
            </Field>
            <Field data-invalid={!!errors.defaultScoreThreshold}>
              <FieldLabel htmlFor="kb-default-score-threshold">
                默认分数阈值
              </FieldLabel>
              <FieldContent>
                <Input
                  id="kb-default-score-threshold"
                  type="number"
                  min="0"
                  max="1"
                  step="0.1"
                  aria-invalid={!!errors.defaultScoreThreshold}
                  {...register("defaultScoreThreshold")}
                />
                <FieldError errors={[errors.defaultScoreThreshold]} />
              </FieldContent>
            </Field>
            <Field data-invalid={!!errors.defaultRerankLimit}>
              <FieldLabel htmlFor="kb-default-rerank-limit">
                默认重排序限制
              </FieldLabel>
              <FieldContent>
                <Input
                  id="kb-default-rerank-limit"
                  type="number"
                  min="0"
                  max="100"
                  aria-invalid={!!errors.defaultRerankLimit}
                  {...register("defaultRerankLimit")}
                />
                <FieldError errors={[errors.defaultRerankLimit]} />
              </FieldContent>
            </Field>
            <Field data-invalid={!!errors.answerMode}>
              <FieldLabel htmlFor="kb-answer-mode">回答模式</FieldLabel>
              <FieldContent>
                <Controller
                  control={control}
                  name="answerMode"
                  render={({ field }) => (
                    <Select value={field.value} onValueChange={field.onChange}>
                      <SelectTrigger
                        id="kb-answer-mode"
                        aria-invalid={!!errors.answerMode}
                      >
                        <SelectValue placeholder="选择回答模式">
                          {field.value
                            ? getEnumLabel(
                                KnowledgeAnswerModeLabels,
                                Number(field.value),
                              )
                            : undefined}
                        </SelectValue>
                      </SelectTrigger>
                      <SelectContent>
                        {getEnumOptions(KnowledgeAnswerModeLabels).map((option) => (
                          <SelectItem
                            key={option.value}
                            value={String(option.value)}
                          >
                            {option.label}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  )}
                />
                <FieldError errors={[errors.answerMode]} />
              </FieldContent>
            </Field>
          </div>

          <Field data-invalid={!!errors.remark}>
            <FieldLabel htmlFor="kb-remark">备注</FieldLabel>
            <FieldContent>
              <Textarea
                id="kb-remark"
                placeholder="备注信息"
                rows={2}
                aria-invalid={!!errors.remark}
                {...register("remark")}
              />
              <FieldError errors={[errors.remark]} />
            </FieldContent>
          </Field>
        </form>
      )}
    </ProjectDialog>
  );
}
