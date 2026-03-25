"use client"

import { useEffect, useState } from "react"
import { zodResolver } from "@hookform/resolvers/zod"
import { Controller, Resolver, useForm } from "react-hook-form"
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
  type AdminWidgetSite,
  type CreateAdminWidgetSitePayload,
  fetchAIAgentsAll,
  fetchWidgetSite,
} from "@/lib/api/admin"

type WidgetSiteFormDialogProps = {
  open: boolean
  saving: boolean
  itemId: number | null
  onOpenChange: (open: boolean) => void
  onSubmit: (payload: CreateAdminWidgetSitePayload) => Promise<void>
}

const schema = z.object({
  aiAgentId: z.string().trim().regex(/^\d+$/, "请选择 AI Agent"),
  name: z.string().trim().min(1, "站点名称不能为空"),
  remark: z.string().trim(),
})

type EditForm = z.infer<typeof schema>

const resolver = zodResolver(schema as never) as Resolver<
  z.input<typeof schema>,
  undefined,
  z.output<typeof schema>
>

const emptyForm: EditForm = {
  aiAgentId: "",
  name: "",
  remark: "",
}

function buildForm(item: AdminWidgetSite | null): EditForm {
  if (!item) {
    return emptyForm
  }

  return {
    aiAgentId: item.aiAgentId > 0 ? String(item.aiAgentId) : "",
    name: item.name,
    remark: item.remark || "",
  }
}

function buildPayload(form: EditForm, status: number): CreateAdminWidgetSitePayload {
  return {
    aiAgentId: Number(form.aiAgentId),
    name: form.name.trim(),
    status,
    remark: form.remark.trim(),
  }
}

type WidgetSiteFormBodyProps = Omit<WidgetSiteFormDialogProps, "open">

export function EditDialog({
  open,
  saving,
  itemId,
  onOpenChange,
  onSubmit,
}: WidgetSiteFormDialogProps) {
  if (!open) {
    return null
  }

  return (
    <WidgetSiteFormBody
      key={itemId ? `edit-${itemId}` : "create"}
      itemId={itemId}
      saving={saving}
      onOpenChange={onOpenChange}
      onSubmit={onSubmit}
    />
  )
}

function WidgetSiteFormBody({
  saving,
  itemId,
  onOpenChange,
  onSubmit,
}: WidgetSiteFormBodyProps) {
  const formId = "widget-site-edit-form"
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
        const data = await fetchWidgetSite(itemId)
        setCurrentStatus(data.status)
        reset(buildForm(data))
      } catch (error) {
        console.error("Failed to load widget site:", error)
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
      title={itemId ? "编辑" : "新建"}
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
              <FieldLabel htmlFor="widget-site-name">站点名称</FieldLabel>
              <FieldContent>
                <Input id="widget-site-name" {...register("name")} />
                <FieldError errors={[errors.name]} />
              </FieldContent>
            </Field>
          </div>

          <Field data-invalid={!!errors.remark}>
            <FieldLabel htmlFor="widget-site-remark">备注</FieldLabel>
            <FieldContent>
              <Textarea id="widget-site-remark" rows={3} {...register("remark")} />
              <FieldError errors={[errors.remark]} />
            </FieldContent>
          </Field>
        </form>
      )}
    </ProjectDialog>
  )
}
