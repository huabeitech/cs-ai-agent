"use client"

import { useEffect, useState } from "react"
import { zodResolver } from "@hookform/resolvers/zod"
import { Controller, Resolver, useForm, useWatch } from "react-hook-form"
import { z } from "zod/v4"

import { OptionCombobox } from "@/components/option-combobox"
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
import {
  type AIAgent,
  type AdminChannel,
  type CreateAdminChannelPayload,
  fetchAIAgentsAll,
  fetchChannel,
} from "@/lib/api/admin"

type ChannelFormDialogProps = {
  open: boolean
  saving: boolean
  itemId: number | null
  onOpenChange: (open: boolean) => void
  onSubmit: (payload: CreateAdminChannelPayload) => Promise<void>
}

const channelTypeOptions = [
  { value: "web", label: "Web 站点" },
  { value: "wxwork_kf", label: "企业微信客服" },
] as const

const schema = z.object({
  channelType: z.enum(["web", "wxwork_kf"], "请选择渠道类型"),
  channelCode: z.string().trim().min(1, "渠道编码不能为空"),
  aiAgentId: z.string().trim().regex(/^\d+$/, "请选择 AI Agent"),
  name: z.string().trim().min(1, "渠道名称不能为空"),
  appId: z.string().trim(),
  openKfId: z.string().trim(),
  sortNo: z.string().trim(),
  remark: z.string().trim(),
})

type EditForm = z.infer<typeof schema>

const resolver = zodResolver(schema as never) as Resolver<
  z.input<typeof schema>,
  undefined,
  z.output<typeof schema>
>

const emptyForm: EditForm = {
  channelType: "web",
  channelCode: "",
  aiAgentId: "",
  name: "",
  appId: "",
  openKfId: "",
  sortNo: "0",
  remark: "",
}

function parseOpenKfId(configJson: string): string {
  if (!configJson.trim()) {
    return ""
  }
  try {
    const parsed = JSON.parse(configJson) as { openKfId?: string }
    return typeof parsed.openKfId === "string" ? parsed.openKfId.trim() : ""
  } catch {
    return ""
  }
}

function buildForm(item: AdminChannel | null): EditForm {
  if (!item) {
    return emptyForm
  }
  return {
    channelType: item.channelType === "wxwork_kf" ? "wxwork_kf" : "web",
    channelCode: item.channelCode,
    aiAgentId: item.aiAgentId > 0 ? String(item.aiAgentId) : "",
    name: item.name,
    appId: item.appId || "",
    openKfId: parseOpenKfId(item.configJson),
    sortNo: String(item.sortNo ?? 0),
    remark: item.remark || "",
  }
}

function buildPayload(form: EditForm, status: number): CreateAdminChannelPayload {
  const channelType = form.channelType
  const appId = channelType === "web" ? form.appId.trim() : ""
  const configJson =
    channelType === "wxwork_kf"
      ? JSON.stringify({ openKfId: form.openKfId.trim() })
      : ""
  return {
    channelType,
    channelCode: form.channelCode.trim(),
    aiAgentId: Number(form.aiAgentId),
    name: form.name.trim(),
    appId,
    configJson,
    sortNo: Number(form.sortNo || "0"),
    status,
    remark: form.remark.trim(),
  }
}

type ChannelFormBodyProps = Omit<ChannelFormDialogProps, "open">

export function EditDialog({
  open,
  saving,
  itemId,
  onOpenChange,
  onSubmit,
}: ChannelFormDialogProps) {
  if (!open) {
    return null
  }

  return (
    <ChannelFormBody
      key={itemId ? `edit-${itemId}` : "create"}
      itemId={itemId}
      saving={saving}
      onOpenChange={onOpenChange}
      onSubmit={onSubmit}
    />
  )
}

function ChannelFormBody({
  saving,
  itemId,
  onOpenChange,
  onSubmit,
}: ChannelFormBodyProps) {
  const formId = "channel-edit-form"
  const [loading, setLoading] = useState(false)
  const [aiAgents, setAIAgents] = useState<AIAgent[]>([])
  const [currentStatus, setCurrentStatus] = useState(0)
  const form = useForm<
    z.input<typeof schema>,
    undefined,
    z.output<typeof schema>
  >({
    resolver,
    defaultValues: emptyForm,
  })
  const {
    control,
    handleSubmit,
    register,
    reset,
    formState: { errors },
  } = form
  const channelType = useWatch({ control, name: "channelType" })

  useEffect(() => {
    async function loadAIAgents() {
      try {
        const data = await fetchAIAgentsAll({ status: 1 })
        setAIAgents(data)
      } catch (error) {
        console.error("Failed to load AI agents:", error)
      }
    }
    void loadAIAgents()
  }, [])

  useEffect(() => {
    async function loadDetail() {
      if (!itemId) {
        setCurrentStatus(0)
        reset(emptyForm)
        return
      }
      setLoading(true)
      try {
        const data = await fetchChannel(itemId)
        setCurrentStatus(data.status)
        reset(buildForm(data))
      } catch (error) {
        console.error("Failed to load channel:", error)
      } finally {
        setLoading(false)
      }
    }
    void loadDetail()
  }, [itemId, reset])

  const aiAgentOptions = aiAgents.map((item) => ({
    value: String(item.id),
    label: item.name,
  }))

  async function onFormSubmit(values: EditForm) {
    await onSubmit(buildPayload(values, currentStatus))
  }

  return (
    <ProjectDialog
      open={true}
      onOpenChange={onOpenChange}
      title={itemId ? "编辑渠道" : "新建渠道"}
      size="lg"
      allowFullscreen
      footer={
        <>
          <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
            取消
          </Button>
          <Button type="submit" form={formId} disabled={saving || loading}>
            {saving ? "保存中..." : "保存"}
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
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <Field data-invalid={!!errors.channelType}>
              <FieldLabel>渠道类型</FieldLabel>
              <FieldContent>
                <Controller
                  control={control}
                  name="channelType"
                  render={({ field }) => (
                    <OptionCombobox
                      value={field.value}
                      options={[...channelTypeOptions]}
                      placeholder="请选择渠道类型"
                      searchPlaceholder="搜索渠道类型"
                      emptyText="未找到渠道类型"
                      onChange={field.onChange}
                    />
                  )}
                />
                <FieldError errors={[errors.channelType]} />
              </FieldContent>
            </Field>

            <Field data-invalid={!!errors.channelCode}>
              <FieldLabel htmlFor="channel-code">渠道编码</FieldLabel>
              <FieldContent>
                <Input id="channel-code" {...register("channelCode")} />
                <FieldError errors={[errors.channelCode]} />
              </FieldContent>
            </Field>

            <Field data-invalid={!!errors.aiAgentId}>
              <FieldLabel>接待 Agent</FieldLabel>
              <FieldContent>
                <Controller
                  control={control}
                  name="aiAgentId"
                  render={({ field }) => (
                    <OptionCombobox
                      value={field.value}
                      options={aiAgentOptions}
                      placeholder="请选择 AI Agent"
                      searchPlaceholder="搜索 AI Agent"
                      emptyText="未找到 AI Agent"
                      onChange={field.onChange}
                    />
                  )}
                />
                <FieldError errors={[errors.aiAgentId]} />
              </FieldContent>
            </Field>

            <Field data-invalid={!!errors.name}>
              <FieldLabel htmlFor="channel-name">渠道名称</FieldLabel>
              <FieldContent>
                <Input id="channel-name" {...register("name")} />
                <FieldError errors={[errors.name]} />
              </FieldContent>
            </Field>

            <Field data-invalid={!!errors.sortNo}>
              <FieldLabel htmlFor="channel-sort-no">排序号</FieldLabel>
              <FieldContent>
                <Input id="channel-sort-no" {...register("sortNo")} />
                <FieldError errors={[errors.sortNo]} />
              </FieldContent>
            </Field>

            {channelType === "web" ? (
              <Field data-invalid={!!errors.appId}>
                <FieldLabel htmlFor="channel-app-id">AppID</FieldLabel>
                <FieldContent>
                  <Input id="channel-app-id" {...register("appId")} />
                  <FieldError errors={[errors.appId]} />
                </FieldContent>
              </Field>
            ) : (
              <Field data-invalid={!!errors.openKfId}>
                <FieldLabel htmlFor="channel-open-kf-id">OpenKfID</FieldLabel>
                <FieldContent>
                  <Input id="channel-open-kf-id" {...register("openKfId")} />
                  <FieldError errors={[errors.openKfId]} />
                </FieldContent>
              </Field>
            )}
          </div>

          <Field data-invalid={!!errors.remark}>
            <FieldLabel htmlFor="channel-remark">备注</FieldLabel>
            <FieldContent>
              <Textarea id="channel-remark" rows={3} {...register("remark")} />
              <FieldError errors={[errors.remark]} />
            </FieldContent>
          </Field>
        </form>
      )}
    </ProjectDialog>
  )
}
