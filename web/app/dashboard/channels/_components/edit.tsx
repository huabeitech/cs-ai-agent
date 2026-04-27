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
  type WxWorkKFAccount,
  fetchAIAgentsAll,
  fetchChannel,
  fetchWxWorkKFAccounts,
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

const widgetPositionOptions = [
  { value: "right", label: "右下角" },
  { value: "left", label: "左下角" },
] as const

type WebChannelConfig = {
  title?: string
  subtitle?: string
  themeColor?: string
  position?: "left" | "right"
  width?: string
}

const defaultWebChannelConfig: Required<WebChannelConfig> = {
  title: "在线客服",
  subtitle: "欢迎咨询",
  themeColor: "#2563eb",
  position: "right",
  width: "380px",
}

const schema = z
  .object({
    channelType: z.enum(["web", "wxwork_kf"], "请选择渠道类型"),
    aiAgentId: z.string().trim().regex(/^\d+$/, "请选择 AI Agent"),
    name: z.string().trim().min(1, "渠道名称不能为空"),
    openKfId: z.string().trim(),
    widgetTitle: z.string().trim(),
    widgetSubtitle: z.string().trim(),
    widgetThemeColor: z.string().trim(),
    widgetPosition: z.enum(["left", "right"]),
    widgetWidth: z.string().trim(),
    remark: z.string().trim(),
  })
  .superRefine((values, ctx) => {
    if (values.channelType === "wxwork_kf" && !values.openKfId.trim()) {
      ctx.addIssue({
        code: "custom",
        path: ["openKfId"],
        message: "请选择企业微信客服账号",
      })
    }
  })

type EditForm = z.infer<typeof schema>

const resolver = zodResolver(schema as never) as Resolver<
  z.input<typeof schema>,
  undefined,
  z.output<typeof schema>
>

const emptyForm: EditForm = {
  channelType: "web",
  aiAgentId: "",
  name: "",
  openKfId: "",
  widgetTitle: defaultWebChannelConfig.title,
  widgetSubtitle: defaultWebChannelConfig.subtitle,
  widgetThemeColor: defaultWebChannelConfig.themeColor,
  widgetPosition: defaultWebChannelConfig.position,
  widgetWidth: defaultWebChannelConfig.width,
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

function parseWebChannelConfig(configJson: string): Required<WebChannelConfig> {
  if (!configJson.trim()) {
    return defaultWebChannelConfig
  }
  try {
    const parsed = JSON.parse(configJson) as WebChannelConfig
    const position = parsed.position === "left" ? "left" : "right"
    return {
      title: parsed.title?.trim() || defaultWebChannelConfig.title,
      subtitle: parsed.subtitle?.trim() ?? defaultWebChannelConfig.subtitle,
      themeColor:
        parsed.themeColor?.trim() || defaultWebChannelConfig.themeColor,
      position,
      width: parsed.width?.trim() || defaultWebChannelConfig.width,
    }
  } catch {
    return defaultWebChannelConfig
  }
}

function buildForm(item: AdminChannel | null): EditForm {
  if (!item) {
    return emptyForm
  }
  const widgetConfig = parseWebChannelConfig(item.configJson)
  return {
    channelType: item.channelType === "wxwork_kf" ? "wxwork_kf" : "web",
    aiAgentId: item.aiAgentId > 0 ? String(item.aiAgentId) : "",
    name: item.name,
    openKfId: parseOpenKfId(item.configJson),
    widgetTitle: widgetConfig.title,
    widgetSubtitle: widgetConfig.subtitle,
    widgetThemeColor: widgetConfig.themeColor,
    widgetPosition: widgetConfig.position,
    widgetWidth: widgetConfig.width,
    remark: item.remark || "",
  }
}

function buildPayload(form: EditForm, status: number): CreateAdminChannelPayload {
  const channelType = form.channelType
  const configJson =
    channelType === "wxwork_kf"
      ? JSON.stringify({ openKfId: form.openKfId.trim() })
      : JSON.stringify({
          title: form.widgetTitle.trim() || defaultWebChannelConfig.title,
          subtitle: form.widgetSubtitle.trim(),
          themeColor:
            form.widgetThemeColor.trim() || defaultWebChannelConfig.themeColor,
          position: form.widgetPosition || defaultWebChannelConfig.position,
          width: form.widgetWidth.trim() || defaultWebChannelConfig.width,
        })
  return {
    channelType,
    aiAgentId: Number(form.aiAgentId),
    name: form.name.trim(),
    configJson,
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
  const [wxWorkKFAccounts, setWxWorkKFAccounts] = useState<WxWorkKFAccount[]>([])
  const [wxWorkKFAccountsLoading, setWxWorkKFAccountsLoading] = useState(false)
  const [wxWorkKFAccountsError, setWxWorkKFAccountsError] = useState("")
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
  const openKfId = useWatch({ control, name: "openKfId" })

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

  useEffect(() => {
    if (
      channelType !== "wxwork_kf" ||
      wxWorkKFAccounts.length > 0 ||
      wxWorkKFAccountsLoading ||
      wxWorkKFAccountsError
    ) {
      return
    }
    async function loadWxWorkKFAccounts() {
      setWxWorkKFAccountsLoading(true)
      setWxWorkKFAccountsError("")
      try {
        const data = await fetchWxWorkKFAccounts()
        setWxWorkKFAccounts(data)
      } catch (error) {
        console.error("Failed to load WeCom KF accounts:", error)
        setWxWorkKFAccountsError("企业微信客服账号加载失败")
      } finally {
        setWxWorkKFAccountsLoading(false)
      }
    }
    void loadWxWorkKFAccounts()
  }, [
    channelType,
    wxWorkKFAccounts.length,
    wxWorkKFAccountsError,
    wxWorkKFAccountsLoading,
  ])

  const aiAgentOptions = aiAgents.map((item) => ({
    value: String(item.id),
    label: item.name,
  }))
  const wxWorkKFAccountOptions = wxWorkKFAccounts.map((item) => ({
    value: item.openKfId,
    label: item.name ? `${item.name} (${item.openKfId})` : item.openKfId,
  }))
  if (
    channelType === "wxwork_kf" &&
    openKfId &&
    !wxWorkKFAccountOptions.some((item) => item.value === openKfId)
  ) {
    wxWorkKFAccountOptions.unshift({
      value: openKfId,
      label: openKfId,
    })
  }

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
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-1">
            <Field data-invalid={!!errors.name}>
              <FieldLabel htmlFor="channel-name">渠道名称</FieldLabel>
              <FieldContent>
                <Input id="channel-name" {...register("name")} />
                <FieldError errors={[errors.name]} />
              </FieldContent>
            </Field>

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

            {channelType === "wxwork_kf" ? (
              <Field data-invalid={!!errors.openKfId}>
                <FieldLabel>企业微信客服账号</FieldLabel>
                <FieldContent>
                  <Controller
                    control={control}
                    name="openKfId"
                    render={({ field }) => (
                      <OptionCombobox
                        value={field.value}
                        options={wxWorkKFAccountOptions}
                        placeholder={
                          wxWorkKFAccountsLoading ? "正在加载客服账号" : "请选择客服账号"
                        }
                        searchPlaceholder="搜索客服账号"
                        emptyText={
                          wxWorkKFAccountsError || "未找到企业微信客服账号"
                        }
                        disabled={wxWorkKFAccountsLoading}
                        onChange={field.onChange}
                      />
                    )}
                  />
                  <FieldError errors={[errors.openKfId]} />
                </FieldContent>
              </Field>
            ) : null}
          </div>

          {channelType === "web" ? (
            <div className="grid grid-cols-1 gap-4 rounded-md border p-4 sm:grid-cols-2">
              <Field data-invalid={!!errors.widgetTitle}>
                <FieldLabel htmlFor="channel-widget-title">窗口标题</FieldLabel>
                <FieldContent>
                  <Input id="channel-widget-title" {...register("widgetTitle")} />
                  <FieldError errors={[errors.widgetTitle]} />
                </FieldContent>
              </Field>

              <Field data-invalid={!!errors.widgetSubtitle}>
                <FieldLabel htmlFor="channel-widget-subtitle">窗口副标题</FieldLabel>
                <FieldContent>
                  <Input
                    id="channel-widget-subtitle"
                    {...register("widgetSubtitle")}
                  />
                  <FieldError errors={[errors.widgetSubtitle]} />
                </FieldContent>
              </Field>

              <Field data-invalid={!!errors.widgetThemeColor}>
                <FieldLabel htmlFor="channel-widget-theme-color">主题色</FieldLabel>
                <FieldContent>
                  <Input
                    id="channel-widget-theme-color"
                    placeholder="#2563eb"
                    {...register("widgetThemeColor")}
                  />
                  <FieldError errors={[errors.widgetThemeColor]} />
                </FieldContent>
              </Field>

              <Field data-invalid={!!errors.widgetPosition}>
                <FieldLabel>挂载位置</FieldLabel>
                <FieldContent>
                  <Controller
                    control={control}
                    name="widgetPosition"
                    render={({ field }) => (
                      <OptionCombobox
                        value={field.value}
                        options={[...widgetPositionOptions]}
                        placeholder="请选择挂载位置"
                        searchPlaceholder="搜索挂载位置"
                        emptyText="未找到挂载位置"
                        onChange={field.onChange}
                      />
                    )}
                  />
                  <FieldError errors={[errors.widgetPosition]} />
                </FieldContent>
              </Field>

              <Field data-invalid={!!errors.widgetWidth}>
                <FieldLabel htmlFor="channel-widget-width">窗口宽度</FieldLabel>
                <FieldContent>
                  <Input
                    id="channel-widget-width"
                    placeholder="380px"
                    {...register("widgetWidth")}
                  />
                  <FieldError errors={[errors.widgetWidth]} />
                </FieldContent>
              </Field>

            </div>
          ) : null}

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
