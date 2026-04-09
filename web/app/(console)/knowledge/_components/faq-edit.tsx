"use client";

import { useEffect, useState } from "react";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm, type Resolver } from "react-hook-form";
import { z } from "zod/v4";

import { ProjectDialog } from "@/components/project-dialog";
import { Button } from "@/components/ui/button";
import { Field, FieldContent, FieldError, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import {
  fetchKnowledgeFAQ,
  type CreateKnowledgeFAQPayload,
  type KnowledgeFAQ,
} from "@/lib/api/admin";

type FAQEditDialogProps = {
  open: boolean;
  saving: boolean;
  itemId: number | null;
  knowledgeBaseId: number | null;
  onOpenChange: (open: boolean) => void;
  onSubmit: (payload: CreateKnowledgeFAQPayload) => Promise<void>;
};

const formSchema = z.object({
  question: z.string().trim().min(1, "问题不能为空").max(500, "问题最多500个字符"),
  answer: z.string().trim().min(1, "答案不能为空"),
  similarQuestionsText: z.string(),
  remark: z.string().trim().max(500, "备注最多500个字符"),
});

type EditForm = z.infer<typeof formSchema>;

const resolver = zodResolver(formSchema as never) as Resolver<
  z.input<typeof formSchema>,
  undefined,
  z.output<typeof formSchema>
>;

const emptyForm: EditForm = {
  question: "",
  answer: "",
  similarQuestionsText: "",
  remark: "",
};

function buildForm(item: KnowledgeFAQ | null): EditForm {
  if (!item) {
    return emptyForm;
  }
  return {
    question: item.question,
    answer: item.answer,
    similarQuestionsText: (item.similarQuestions ?? []).join("\n"),
    remark: item.remark ?? "",
  };
}

function buildPayload(form: EditForm, knowledgeBaseId: number): CreateKnowledgeFAQPayload {
  return {
    knowledgeBaseId,
    question: form.question.trim(),
    answer: form.answer.trim(),
    similarQuestions: form.similarQuestionsText
      .split("\n")
      .map((item) => item.trim())
      .filter(Boolean),
    remark: form.remark.trim(),
  };
}

export function FAQEditDialog({
  open,
  saving,
  itemId,
  knowledgeBaseId,
  onOpenChange,
  onSubmit,
}: FAQEditDialogProps) {
  if (!open || !knowledgeBaseId) {
    return null;
  }
  return (
    <FAQEditDialogBody
      key={itemId ? `edit-${itemId}` : "create"}
      open={open}
      saving={saving}
      itemId={itemId}
      knowledgeBaseId={knowledgeBaseId}
      onOpenChange={onOpenChange}
      onSubmit={onSubmit}
    />
  );
}

type FAQEditDialogBodyProps = {
  open: boolean;
  saving: boolean;
  itemId: number | null;
  knowledgeBaseId: number;
  onOpenChange: (open: boolean) => void;
  onSubmit: (payload: CreateKnowledgeFAQPayload) => Promise<void>;
};

function FAQEditDialogBody({
  open,
  saving,
  itemId,
  knowledgeBaseId,
  onOpenChange,
  onSubmit,
}: FAQEditDialogBodyProps) {
  const [loading, setLoading] = useState(false);
  const formId = "knowledge-faq-edit-form";
  const form = useForm<
    z.input<typeof formSchema>,
    undefined,
    z.output<typeof formSchema>
  >({
    resolver,
    defaultValues: emptyForm,
  });
  const {
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
        const data = await fetchKnowledgeFAQ(itemId);
        reset(buildForm(data));
      } finally {
        setLoading(false);
      }
    }
    if (open) {
      void loadDetail();
    }
  }, [itemId, open, reset]);

  async function onFormSubmit(values: EditForm) {
    await onSubmit(buildPayload(values, knowledgeBaseId));
  }

  return (
    <ProjectDialog
      open={open}
      onOpenChange={onOpenChange}
      title={itemId ? "编辑FAQ" : "新建FAQ"}
      allowFullscreen
      size="xl"
      footer={
        <>
          <Button type="button" variant="outline" onClick={() => onOpenChange(false)} disabled={saving}>
            取消
          </Button>
          <Button type="submit" form={formId} disabled={saving || loading}>
            {saving ? "保存中..." : itemId ? "保存" : "创建"}
          </Button>
        </>
      }
    >
      {loading ? (
        <div className="flex items-center justify-center py-12 text-muted-foreground">加载中...</div>
      ) : (
        <form id={formId} onSubmit={handleSubmit(onFormSubmit)} className="space-y-4">
          <Field data-invalid={!!errors.question}>
            <FieldLabel htmlFor="faq-question">标准问题</FieldLabel>
            <FieldContent>
              <Input id="faq-question" placeholder="请输入标准问题" {...register("question")} />
              <FieldError errors={[errors.question]} />
            </FieldContent>
          </Field>

          <Field data-invalid={!!errors.answer}>
            <FieldLabel htmlFor="faq-answer">答案</FieldLabel>
            <FieldContent>
              <Textarea id="faq-answer" rows={8} placeholder="请输入FAQ答案" {...register("answer")} />
              <FieldError errors={[errors.answer]} />
            </FieldContent>
          </Field>

          <Field>
            <FieldLabel htmlFor="faq-similar-questions">相似问题</FieldLabel>
            <FieldContent>
              <Textarea
                id="faq-similar-questions"
                rows={5}
                placeholder={"一行一个相似问题"}
                {...register("similarQuestionsText")}
              />
            </FieldContent>
          </Field>

          <Field data-invalid={!!errors.remark}>
            <FieldLabel htmlFor="faq-remark">备注</FieldLabel>
            <FieldContent>
              <Textarea id="faq-remark" rows={3} placeholder="备注" {...register("remark")} />
              <FieldError errors={[errors.remark]} />
            </FieldContent>
          </Field>
        </form>
      )}
    </ProjectDialog>
  );
}

