"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useEffect, useMemo, useState } from "react";
import { Resolver, useForm } from "react-hook-form";
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
import { Textarea } from "@/components/ui/textarea";
import {
  fetchMCPCatalog,
  fetchSkillDefinition,
  type CreateSkillDefinitionPayload,
  type MCPToolCatalogItem,
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
  instruction: "",
  examplesText: "",
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
  instruction: z.string().trim().min(1, "技能说明不能为空"),
  examplesText: z.string().trim(),
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
    instruction: item.instruction ?? "",
    examplesText: (item.examples ?? []).join("\n"),
    remark: item.remark ?? "",
  };
}

function buildPayload(
  form: EditForm,
  allowedToolCodes: string[],
): CreateSkillDefinitionPayload {
  return {
    code: form.code.trim(),
    name: form.name.trim(),
    description: form.description.trim(),
    instruction: form.instruction.trim(),
    examples: form.examplesText
      .split("\n")
      .map((item) => item.trim())
      .filter(Boolean),
    allowedToolCodes,
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
  const [toolCatalog, setToolCatalog] = useState<MCPToolCatalogItem[]>([]);
  const [selectedAllowedToolCodes, setSelectedAllowedToolCodes] = useState<
    string[]
  >([]);
  const [toolCodeToAdd, setToolCodeToAdd] = useState("");
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
        setSelectedAllowedToolCodes([]);
        setToolCodeToAdd("");
        return;
      }

      setLoading(true);
      try {
        const data = await fetchSkillDefinition(itemId);
        reset(buildForm(data));
        setSelectedAllowedToolCodes(data.allowedToolCodes ?? []);
        setToolCodeToAdd("");
      } catch (error) {
        console.error("Failed to load skill definition:", error);
      } finally {
        setLoading(false);
      }
    }

    void loadDetail();
  }, [itemId, reset]);

  useEffect(() => {
    async function loadToolCatalog() {
      try {
        const data = await fetchMCPCatalog();
        setToolCatalog(data);
      } catch (error) {
        console.error("Failed to load MCP tool catalog:", error);
      }
    }

    void loadToolCatalog();
  }, []);

  async function onFormSubmit(values: EditForm) {
    await onSubmit(buildPayload(values, selectedAllowedToolCodes));
  }

  const toolOptions = useMemo(
    () =>
      toolCatalog.map((item) => ({
        value: item.toolCode,
        label: `${item.title || item.toolName} · ${item.toolCode}`,
      })),
    [toolCatalog],
  );

  const addableToolOptions = useMemo(
    () =>
      toolOptions.filter(
        (option) => !selectedAllowedToolCodes.includes(option.value),
      ),
    [selectedAllowedToolCodes, toolOptions],
  );

  const selectedToolOptions = useMemo(
    () =>
      selectedAllowedToolCodes
        .map((toolCode) => toolOptions.find((option) => option.value === toolCode))
        .filter(
          (option): option is { value: string; label: string } => !!option,
        ),
    [selectedAllowedToolCodes, toolOptions],
  );

  function handleAddAllowedTool(toolCode: string) {
    if (!toolCode || selectedAllowedToolCodes.includes(toolCode)) {
      return;
    }
    setSelectedAllowedToolCodes((prev) => [...prev, toolCode]);
    setToolCodeToAdd("");
  }

  function handleRemoveAllowedTool(toolCode: string) {
    setSelectedAllowedToolCodes((prev) =>
      prev.filter((item) => item !== toolCode),
    );
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

          <Field data-invalid={!!errors.instruction}>
            <FieldLabel htmlFor="skill-instruction">技能说明</FieldLabel>
            <FieldContent>
              <Textarea
                id="skill-instruction"
                rows={12}
                placeholder="请输入 Skill 文档内容，描述目标、步骤、工具使用规则和边界。"
                aria-invalid={!!errors.instruction}
                {...register("instruction")}
              />
              <FieldError errors={[errors.instruction]} />
            </FieldContent>
          </Field>

          <Field data-invalid={!!errors.examplesText}>
            <FieldLabel htmlFor="skill-examples">示例问法</FieldLabel>
            <FieldContent>
              <Textarea
                id="skill-examples"
                rows={5}
                placeholder={"每行一个典型用户问法，例如：\n我要申请退款\n帮我查下订单"}
                aria-invalid={!!errors.examplesText}
                {...register("examplesText")}
              />
              <FieldError errors={[errors.examplesText]} />
            </FieldContent>
          </Field>

          <Field>
            <FieldLabel>Allowed Tools</FieldLabel>
            <FieldContent className="space-y-3">
              <div className="flex items-center gap-2">
                <div className="flex-1">
                  <OptionCombobox
                    value={toolCodeToAdd}
                    options={addableToolOptions}
                    placeholder="选择允许该 Skill 使用的工具"
                    searchPlaceholder="搜索 toolCode 或工具名"
                    emptyText="没有可添加的工具"
                    onChange={handleAddAllowedTool}
                  />
                </div>
                <Button
                  type="button"
                  variant="outline"
                  disabled={!toolCodeToAdd}
                  onClick={() => handleAddAllowedTool(toolCodeToAdd)}
                >
                  添加
                </Button>
              </div>
              <div className="flex flex-wrap gap-2">
                {selectedToolOptions.length === 0 ? (
                  <span className="text-sm text-muted-foreground">
                    不限制时，Skill 会继承 Agent 的可用工具范围。
                  </span>
                ) : (
                  selectedToolOptions.map((option) => (
                    <Button
                      key={option.value}
                      type="button"
                      variant="outline"
                      size="sm"
                      onClick={() => handleRemoveAllowedTool(option.value)}
                      className="justify-start"
                    >
                      {option.label}
                    </Button>
                  ))
                )}
              </div>
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
