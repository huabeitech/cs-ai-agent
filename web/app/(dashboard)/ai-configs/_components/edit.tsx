"use client"

import { useEffect, useState } from "react"
import { zodResolver } from "@hookform/resolvers/zod"
import { Controller, Resolver, useForm } from "react-hook-form"
import { z } from "zod/v4"

import { ProjectDialog } from "@/components/project-dialog"
import { Button } from "@/components/ui/button"
import {
  Field,
  FieldContent,
  FieldError,
  FieldLabel,
} from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"
import { type AIConfig, type CreateAIConfigPayload, fetchAIConfig } from "@/lib/api/admin"
import {
  AIModelType,
  AIModelTypeLabels,
  AIProvider,
  AIProviderLabels,
} from "@/lib/generated/enums"
import { getEnumOptions } from "@/lib/enums"
import { OptionCombobox } from "./option-combobox"

type AIConfigEditDialogProps = {
  open: boolean
  saving: boolean
  itemId: number | null
  onOpenChange: (open: boolean) => void
  onSubmit: (payload: CreateAIConfigPayload) => Promise<void>
}

const providerOptions = getEnumOptions(AIProviderLabels).map((option) => ({
  value: String(option.value),
  label: option.label,
}))

const modelTypeOptions = getEnumOptions(AIModelTypeLabels).map((option) => ({
  value: String(option.value),
  label: option.label,
}))

const emptyForm: EditForm = {
  name: "",
  provider: AIProvider.OpenAI,
  baseUrl: "",
  apiKey: "",
  modelType: AIModelType.LLM,
  modelName: "",
  dimension: "0",
  maxContextTokens: "0",
  maxOutputTokens: "0",
  timeoutMs: "120000",
  maxRetryCount: "0",
  rpmLimit: "0",
  tpmLimit: "0",
  remark: "",
}

const aiConfigFormSchema = z.object({
  name: z.string().trim().min(1, "配置名称不能为空"),
  provider: z.string().trim().min(1, "供应商不能为空"),
  baseUrl: z.string().trim().min(1, "基础地址不能为空"),
  apiKey: z.string().trim(),
  modelType: z.string().trim().min(1, "模型类型不能为空"),
  modelName: z.string().trim().min(1, "模型名称不能为空"),
  dimension: z.string().trim().regex(/^\d+$/, "向量维度必须是大于等于 0 的整数"),
  maxContextTokens: z.string().trim().regex(/^\d+$/, "最大上下文 Token 必须是大于等于 0 的整数"),
  maxOutputTokens: z.string().trim().regex(/^\d+$/, "最大输出 Token 必须是大于等于 0 的整数"),
  timeoutMs: z.string().trim().regex(/^\d+$/, "超时时间必须是大于等于 0 的整数"),
  maxRetryCount: z.string().trim().regex(/^\d+$/, "最大重试次数必须是大于等于 0 的整数"),
  rpmLimit: z.string().trim().regex(/^\d+$/, "RPM 限制必须是大于等于 0 的整数"),
  tpmLimit: z.string().trim().regex(/^\d+$/, "TPM 限制必须是大于等于 0 的整数"),
  remark: z.string().trim(),
})

type EditForm = z.infer<typeof aiConfigFormSchema>
const editFormResolver = zodResolver(aiConfigFormSchema as never) as Resolver<
  z.input<typeof aiConfigFormSchema>,
  undefined,
  z.output<typeof aiConfigFormSchema>
>

function buildForm(item: AIConfig | null): EditForm {
  if (!item) {
    return emptyForm
  }

  return {
    name: item.name,
    provider: item.provider,
    baseUrl: item.baseUrl,
    apiKey: item.apiKey,
    modelType: item.modelType,
    modelName: item.modelName,
    dimension: String(item.dimension),
    maxContextTokens: String(item.maxContextTokens),
    maxOutputTokens: String(item.maxOutputTokens),
    timeoutMs: String(item.timeoutMs),
    maxRetryCount: String(item.maxRetryCount),
    rpmLimit: String(item.rpmLimit),
    tpmLimit: String(item.tpmLimit),
    remark: item.remark ?? "",
  }
}

function buildPayload(form: EditForm): CreateAIConfigPayload {
  return {
    name: form.name.trim(),
    provider: form.provider,
    baseUrl: form.baseUrl.trim(),
    apiKey: form.apiKey.trim(),
    modelType: form.modelType,
    modelName: form.modelName.trim(),
    dimension: Number(form.dimension),
    maxContextTokens: Number(form.maxContextTokens),
    maxOutputTokens: Number(form.maxOutputTokens),
    timeoutMs: Number(form.timeoutMs),
    maxRetryCount: Number(form.maxRetryCount),
    rpmLimit: Number(form.rpmLimit),
    tpmLimit: Number(form.tpmLimit),
    remark: form.remark.trim(),
  }
}

export function EditDialog({
  open,
  saving,
  itemId,
  onOpenChange,
  onSubmit,
}: AIConfigEditDialogProps) {
  if (!open) {
    return null
  }

  return (
    <AIConfigEditDialogBody
      key={itemId ? `edit-${itemId}` : "create"}
      open={open}
      saving={saving}
      itemId={itemId}
      onOpenChange={onOpenChange}
      onSubmit={onSubmit}
    />
  )
}

type AIConfigEditDialogBodyProps = AIConfigEditDialogProps

function AIConfigEditDialogBody({
  open,
  saving,
  itemId,
  onOpenChange,
  onSubmit,
}: AIConfigEditDialogBodyProps) {
  const formId = "ai-config-edit-form"
  const [loading, setLoading] = useState(false)
  const form = useForm<
    z.input<typeof aiConfigFormSchema>,
    undefined,
    z.output<typeof aiConfigFormSchema>
  >({
    resolver: editFormResolver,
    defaultValues: emptyForm,
  })
  const {
    control,
    handleSubmit,
    reset,
    register,
    watch,
    formState: { errors },
  } = form

  const modelType = watch("modelType")

  useEffect(() => {
    async function loadDetail() {
      if (!itemId) {
        reset(emptyForm)
        return
      }
      setLoading(true)
      try {
        const data = await fetchAIConfig(itemId)
        reset(buildForm(data))
      } catch (error) {
        console.error("Failed to load AI config:", error)
      } finally {
        setLoading(false)
      }
    }
    void loadDetail()
  }, [itemId, reset])

  async function onFormSubmit(values: EditForm) {
    await onSubmit(buildPayload(values))
  }

  return (
    <ProjectDialog
      open={open}
      onOpenChange={onOpenChange}
      title={itemId ? "编辑 AI 配置" : "新建 AI 配置"}
      size="xl"
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
          <Field data-invalid={!!errors.name}>
            <FieldLabel htmlFor="ai-config-name">配置名称</FieldLabel>
            <FieldContent>
              <Input
                id="ai-config-name"
                placeholder="例如：OpenAI 主回答模型"
                aria-invalid={!!errors.name}
                {...register("name")}
              />
              <FieldError errors={[errors.name]} />
            </FieldContent>
          </Field>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <Field data-invalid={!!errors.provider}>
              <FieldLabel>供应商</FieldLabel>
              <FieldContent>
                <Controller
                  control={control}
                  name="provider"
                  render={({ field }) => (
                    <OptionCombobox
                      value={field.value}
                      options={providerOptions}
                      placeholder="请选择供应商"
                      searchPlaceholder="搜索供应商"
                      emptyText="未找到供应商"
                      onChange={field.onChange}
                    />
                  )}
                />
                <FieldError errors={[errors.provider]} />
              </FieldContent>
            </Field>

            <Field data-invalid={!!errors.modelType}>
              <FieldLabel>模型类型</FieldLabel>
              <FieldContent>
                <Controller
                  control={control}
                  name="modelType"
                  render={({ field }) => (
                    <OptionCombobox
                      value={field.value}
                      options={modelTypeOptions}
                      placeholder="请选择模型类型"
                      searchPlaceholder="搜索模型类型"
                      emptyText="未找到模型类型"
                      onChange={field.onChange}
                    />
                  )}
                />
                <FieldError errors={[errors.modelType]} />
              </FieldContent>
            </Field>
          </div>

          <Field data-invalid={!!errors.baseUrl}>
            <FieldLabel htmlFor="ai-config-base-url">Base URL</FieldLabel>
            <FieldContent>
              <Input
                id="ai-config-base-url"
                placeholder="例如：https://api.openai.com/v1"
                aria-invalid={!!errors.baseUrl}
                {...register("baseUrl")}
              />
              <FieldError errors={[errors.baseUrl]} />
            </FieldContent>
          </Field>

          <Field data-invalid={!!errors.apiKey}>
            <FieldLabel htmlFor="ai-config-api-key">API Key</FieldLabel>
            <FieldContent>
              <Input
                id="ai-config-api-key"
                type="password"
                placeholder="请输入 API Key"
                aria-invalid={!!errors.apiKey}
                {...register("apiKey")}
              />
              <FieldError errors={[errors.apiKey]} />
            </FieldContent>
          </Field>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <Field data-invalid={!!errors.modelName}>
              <FieldLabel htmlFor="ai-config-model-name">模型名称</FieldLabel>
              <FieldContent>
                <Input
                  id="ai-config-model-name"
                  placeholder="例如：gpt-4o-mini"
                  aria-invalid={!!errors.modelName}
                  {...register("modelName")}
                />
                <FieldError errors={[errors.modelName]} />
              </FieldContent>
            </Field>

            <Field data-invalid={!!errors.dimension}>
              <FieldLabel htmlFor="ai-config-dimension">向量维度</FieldLabel>
              <FieldContent>
                <Input
                  id="ai-config-dimension"
                  type="number"
                  min={0}
                  step={1}
                  disabled={modelType !== AIModelType.Embedding}
                  aria-invalid={!!errors.dimension}
                  {...register("dimension")}
                />
                <FieldError errors={[errors.dimension]} />
              </FieldContent>
            </Field>
          </div>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <Field data-invalid={!!errors.maxContextTokens}>
              <FieldLabel htmlFor="ai-config-max-context">最大上下文 Token</FieldLabel>
              <FieldContent>
                <Input
                  id="ai-config-max-context"
                  type="number"
                  min={0}
                  step={1}
                  aria-invalid={!!errors.maxContextTokens}
                  {...register("maxContextTokens")}
                />
                <FieldError errors={[errors.maxContextTokens]} />
              </FieldContent>
            </Field>

            <Field data-invalid={!!errors.maxOutputTokens}>
              <FieldLabel htmlFor="ai-config-max-output">最大输出 Token</FieldLabel>
              <FieldContent>
                <Input
                  id="ai-config-max-output"
                  type="number"
                  min={0}
                  step={1}
                  aria-invalid={!!errors.maxOutputTokens}
                  {...register("maxOutputTokens")}
                />
                <FieldError errors={[errors.maxOutputTokens]} />
              </FieldContent>
            </Field>
          </div>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <Field data-invalid={!!errors.timeoutMs}>
              <FieldLabel htmlFor="ai-config-timeout">超时时间 (ms)</FieldLabel>
              <FieldContent>
                <Input
                  id="ai-config-timeout"
                  type="number"
                  min={0}
                  step={1}
                  aria-invalid={!!errors.timeoutMs}
                  {...register("timeoutMs")}
                />
                <FieldError errors={[errors.timeoutMs]} />
              </FieldContent>
            </Field>

            <Field data-invalid={!!errors.maxRetryCount}>
              <FieldLabel htmlFor="ai-config-retry">最大重试次数</FieldLabel>
              <FieldContent>
                <Input
                  id="ai-config-retry"
                  type="number"
                  min={0}
                  step={1}
                  aria-invalid={!!errors.maxRetryCount}
                  {...register("maxRetryCount")}
                />
                <FieldError errors={[errors.maxRetryCount]} />
              </FieldContent>
            </Field>
          </div>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
            <Field data-invalid={!!errors.rpmLimit}>
              <FieldLabel htmlFor="ai-config-rpm">RPM 限制</FieldLabel>
              <FieldContent>
                <Input
                  id="ai-config-rpm"
                  type="number"
                  min={0}
                  step={1}
                  aria-invalid={!!errors.rpmLimit}
                  {...register("rpmLimit")}
                />
                <FieldError errors={[errors.rpmLimit]} />
              </FieldContent>
            </Field>

            <Field data-invalid={!!errors.tpmLimit}>
              <FieldLabel htmlFor="ai-config-tpm">TPM 限制</FieldLabel>
              <FieldContent>
                <Input
                  id="ai-config-tpm"
                  type="number"
                  min={0}
                  step={1}
                  aria-invalid={!!errors.tpmLimit}
                  {...register("tpmLimit")}
                />
                <FieldError errors={[errors.tpmLimit]} />
              </FieldContent>
            </Field>
          </div>

          <Field data-invalid={!!errors.remark}>
            <FieldLabel htmlFor="ai-config-remark">备注</FieldLabel>
            <FieldContent>
              <Textarea
                id="ai-config-remark"
                placeholder="记录用途、费用、限制说明等"
                rows={3}
                aria-invalid={!!errors.remark}
                {...register("remark")}
              />
              <FieldError errors={[errors.remark]} />
            </FieldContent>
          </Field>
        </form>
      )}
    </ProjectDialog>
  )
}
