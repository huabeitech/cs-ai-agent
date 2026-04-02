"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useEffect, useState } from "react";
import { Resolver, useForm } from "react-hook-form";
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
import { Textarea } from "@/components/ui/textarea";
import {
  fetchSkillDefinition,
  type CreateSkillDefinitionPayload,
  type SkillDefinition,
} from "@/lib/api/admin";
 

type SkillEditDialogProps = {
  open: boolean;
  saving: boolean;
  itemId: number | null;
  onOpenChange: (open: boolean) => void;
  onSubmit: (payload: CreateSkillDefinitionPayload) => Promise<void>;
};

const emptyForm: EditForm = {
  code: "",
  name: "",
  description: "",
  prompt: "",
  remark: "",
};

const skillFormSchema = z.object({
  code: z
    .string()
    .trim()
    .min(1, "Skill 编码不能为空")
    .regex(/^[a-zA-Z0-9_-]+$/, "Skill 编码仅支持字母、数字、下划线和中划线"),
  name: z.string().trim().min(1, "Skill 名称不能为空"),
  description: z.string().trim(),
  prompt: z.string().trim().min(1, "Prompt 不能为空"),
  remark: z.string().trim(),
});

type EditForm = z.infer<typeof skillFormSchema>;
const editFormResolver = zodResolver(skillFormSchema as never) as Resolver<
  z.input<typeof skillFormSchema>,
  undefined,
  z.output<typeof skillFormSchema>
>;

function buildForm(item: SkillDefinition | null): EditForm {
  if (!item) {
    return emptyForm;
  }

  return {
    code: item.code,
    name: item.name,
    description: item.description ?? "",
    prompt: item.prompt ?? "",
    remark: item.remark ?? "",
  };
}

function buildPayload(form: EditForm): CreateSkillDefinitionPayload {
  return {
    code: form.code.trim(),
    name: form.name.trim(),
    description: form.description.trim(),
    prompt: form.prompt.trim(),
    remark: form.remark.trim(),
  };
}

export function EditDialog({
  open,
  saving,
  itemId,
  onOpenChange,
  onSubmit,
}: SkillEditDialogProps) {
  if (!open) {
    return null;
  }

  return (
    <SkillEditDialogBody
      key={itemId ? `edit-${itemId}` : "create"}
      open={open}
      saving={saving}
      itemId={itemId}
      onOpenChange={onOpenChange}
      onSubmit={onSubmit}
    />
  );
}

type SkillEditDialogBodyProps = SkillEditDialogProps;

function SkillEditDialogBody({
  open,
  saving,
  itemId,
  onOpenChange,
  onSubmit,
}: SkillEditDialogBodyProps) {
  const formId = "skill-definition-edit-form";
  const [loading, setLoading] = useState(false);
  const form = useForm<
    z.input<typeof skillFormSchema>,
    undefined,
    z.output<typeof skillFormSchema>
  >({
    resolver: editFormResolver,
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
        const data = await fetchSkillDefinition(itemId);
        reset(buildForm(data));
      } catch (error) {
        console.error("Failed to load skill definition:", error);
      } finally {
        setLoading(false);
      }
    }

    void loadDetail();
  }, [itemId, reset]);

  async function onFormSubmit(values: EditForm) {
    await onSubmit(buildPayload(values));
  }

  return (
    <ProjectDialog
      open={open}
      onOpenChange={onOpenChange}
      title={itemId ? "编辑" : "新建"}
      size="xl"
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
        <form
          id={formId}
          onSubmit={handleSubmit(onFormSubmit)}
          className="space-y-4"
        >
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <Field data-invalid={!!errors.code}>
              <FieldLabel htmlFor="skill-code">编码</FieldLabel>
              <FieldContent>
                <Input
                  id="skill-code"
                  placeholder="例如：refund_skill"
                  aria-invalid={!!errors.code}
                  {...register("code")}
                />
                <FieldError errors={[errors.code]} />
              </FieldContent>
            </Field>
            <Field data-invalid={!!errors.name}>
              <FieldLabel htmlFor="skill-name">名称</FieldLabel>
              <FieldContent>
                <Input
                  id="skill-name"
                  placeholder="例如：退款处理"
                  aria-invalid={!!errors.name}
                  {...register("name")}
                />
                <FieldError errors={[errors.name]} />
              </FieldContent>
            </Field>
          </div>

          <Field data-invalid={!!errors.description}>
            <FieldLabel htmlFor="skill-description">描述</FieldLabel>
            <FieldContent>
              <Textarea
                id="skill-description"
                rows={3}
                placeholder="描述这个 Skill 的用途、边界和适用场景"
                aria-invalid={!!errors.description}
                {...register("description")}
              />
              <FieldError errors={[errors.description]} />
            </FieldContent>
          </Field>

          <Field data-invalid={!!errors.prompt}>
            <FieldLabel htmlFor="skill-prompt">Prompt</FieldLabel>
            <FieldContent>
              <Textarea
                id="skill-prompt"
                rows={10}
                placeholder="请输入 Skill 的核心提示词模板"
                aria-invalid={!!errors.prompt}
                {...register("prompt")}
              />
              <FieldError errors={[errors.prompt]} />
            </FieldContent>
          </Field>

          <Field data-invalid={!!errors.remark}>
            <FieldLabel htmlFor="skill-remark">备注</FieldLabel>
            <FieldContent>
              <Textarea
                id="skill-remark"
                rows={3}
                placeholder="记录内部备注或维护说明"
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
