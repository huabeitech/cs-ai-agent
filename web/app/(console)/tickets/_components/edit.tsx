"use client"

import { useEffect, useState } from "react"
import { zodResolver } from "@hookform/resolvers/zod"
import type { Resolver } from "react-hook-form"
import { Controller, useForm } from "react-hook-form"
import { z } from "zod/v4"

import { OptionCombobox } from "@/components/option-combobox"
import { ProjectDialog } from "@/components/project-dialog"
import { Button } from "@/components/ui/button"
import {
  Field,
  FieldContent,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"
import {
  fetchAgentProfilesAll,
  fetchAgentTeamsAll,
  type AdminAgentProfile,
  type AdminAgentTeam,
} from "@/lib/api/admin"
import {
  fetchTicketCategoriesAll,
  type TicketCategory,
} from "@/lib/api/ticket-config"
import {
  fetchTicketDetail,
  type CreateTicketPayload,
  type TicketItem,
  type UpdateTicketPayload,
} from "@/lib/api/ticket"

type EditDialogProps = {
  open: boolean
  saving: boolean
  itemId: number | null
  initialValues?: Partial<CreateTicketPayload>
  fixedConversationId?: number
  fixedCustomerId?: number
  titleOverride?: string
  descriptionOverride?: string
  onOpenChange: (open: boolean) => void
  onSubmit: (payload: CreateTicketPayload | UpdateTicketPayload) => Promise<void>
}

const ticketFormSchema = z.object({
  title: z.string().trim().min(1, "标题不能为空"),
  description: z.string().trim(),
  categoryId: z.string().trim(),
  priority: z.enum(["1", "2", "3", "4"], { message: "请选择优先级" }),
  severity: z.enum(["1", "2", "3"], { message: "请选择严重度" }),
  currentTeamId: z.string().trim(),
  currentAssigneeId: z.string().trim(),
  dueAt: z.string().trim(),
})

type EditForm = z.infer<typeof ticketFormSchema>

const editFormResolver = zodResolver(ticketFormSchema as never) as Resolver<
  z.input<typeof ticketFormSchema>,
  undefined,
  z.output<typeof ticketFormSchema>
>

const emptyForm: EditForm = {
  title: "",
  description: "",
  categoryId: "",
  priority: "2",
  severity: "1",
  currentTeamId: "",
  currentAssigneeId: "",
  dueAt: "",
}

function buildForm(item: TicketItem | null): EditForm {
  if (!item) {
    return emptyForm
  }
  return {
    title: item.title ?? "",
    description: item.description ?? "",
    categoryId: item.categoryId ? String(item.categoryId) : "",
    priority: String(item.priority || 2) as EditForm["priority"],
    severity: String(item.severity || 1) as EditForm["severity"],
    currentTeamId: item.currentTeamId ? String(item.currentTeamId) : "",
    currentAssigneeId: item.currentAssigneeId ? String(item.currentAssigneeId) : "",
    dueAt: item.dueAt ? item.dueAt.replace(" ", "T").slice(0, 16) : "",
  }
}

function buildInitialForm(initialValues?: Partial<CreateTicketPayload>): EditForm {
  return {
    title: initialValues?.title?.trim() ?? "",
    description: initialValues?.description?.trim() ?? "",
    categoryId: initialValues?.categoryId ? String(initialValues.categoryId) : "",
    priority: String(initialValues?.priority ?? 2) as EditForm["priority"],
    severity: String(initialValues?.severity ?? 1) as EditForm["severity"],
    currentTeamId: initialValues?.currentTeamId ? String(initialValues.currentTeamId) : "",
    currentAssigneeId: initialValues?.currentAssigneeId
      ? String(initialValues.currentAssigneeId)
      : "",
    dueAt: initialValues?.dueAt ? initialValues.dueAt.replace(" ", "T").slice(0, 16) : "",
  }
}

function buildPayload(form: EditForm): CreateTicketPayload {
  return {
    title: form.title.trim(),
    description: form.description.trim(),
    categoryId: form.categoryId ? Number(form.categoryId) : undefined,
    priority: Number(form.priority),
    severity: Number(form.severity),
    currentTeamId: form.currentTeamId ? Number(form.currentTeamId) : undefined,
    currentAssigneeId: form.currentAssigneeId ? Number(form.currentAssigneeId) : undefined,
    dueAt: form.dueAt ? `${form.dueAt.replace("T", " ")}:00` : undefined,
  }
}

export function EditDialog({
  open,
  saving,
  itemId,
  initialValues,
  fixedConversationId,
  fixedCustomerId,
  titleOverride,
  descriptionOverride,
  onOpenChange,
  onSubmit,
}: EditDialogProps) {
  if (!open) {
    return null
  }
  return (
    <TicketEditDialogBody
      key={itemId ? `edit-${itemId}` : "create"}
      open={open}
      saving={saving}
      itemId={itemId}
      initialValues={initialValues}
      fixedConversationId={fixedConversationId}
      fixedCustomerId={fixedCustomerId}
      titleOverride={titleOverride}
      descriptionOverride={descriptionOverride}
      onOpenChange={onOpenChange}
      onSubmit={onSubmit}
    />
  )
}

type TicketEditDialogBodyProps = EditDialogProps

function TicketEditDialogBody({
  open,
  saving,
  itemId,
  initialValues,
  fixedConversationId,
  fixedCustomerId,
  titleOverride,
  descriptionOverride,
  onOpenChange,
  onSubmit,
}: TicketEditDialogBodyProps) {
  const formId = "ticket-edit-form"
  const [loading, setLoading] = useState(false)
  const [categories, setCategories] = useState<TicketCategory[]>([])
  const [teams, setTeams] = useState<AdminAgentTeam[]>([])
  const [agents, setAgents] = useState<AdminAgentProfile[]>([])
  const form = useForm<
    z.input<typeof ticketFormSchema>,
    undefined,
    z.output<typeof ticketFormSchema>
  >({
    resolver: editFormResolver,
    defaultValues: emptyForm,
  })
  const {
    register,
    control,
    handleSubmit,
    reset,
    formState: { errors },
  } = form

  useEffect(() => {
    async function loadDetail() {
      if (!itemId) {
        reset(buildInitialForm(initialValues))
        return
      }
      setLoading(true)
      try {
        const data = await fetchTicketDetail(itemId)
        reset(buildForm(data.ticket))
      } finally {
        setLoading(false)
      }
    }
    void loadDetail()
  }, [initialValues, itemId, reset])

  useEffect(() => {
    if (!open) {
      return
    }
    void (async () => {
      const [categoryData, teamData, agentData] = await Promise.all([
        fetchTicketCategoriesAll(),
        fetchAgentTeamsAll(),
        fetchAgentProfilesAll(),
      ])
      setCategories(Array.isArray(categoryData) ? categoryData : [])
      setTeams(Array.isArray(teamData) ? teamData : [])
      setAgents(Array.isArray(agentData) ? agentData : [])
    })()
  }, [open])

  const categoryOptions = [{ value: "", label: "不指定分类" }].concat(
    categories.map((category) => ({
      value: String(category.id),
      label: category.parentName
        ? `${category.parentName} / ${category.name}`
        : category.name,
    })),
  )

  const teamOptions = [{ value: "", label: "不指定团队" }].concat(
    teams.map((team) => ({
      value: String(team.id),
      label: team.name,
    })),
  )
  const agentOptions = [{ value: "", label: "不指定处理人" }].concat(
    agents.map((agent) => ({
      value: String(agent.userId),
      label:
        agent.displayName ||
        agent.nickname ||
        agent.username ||
        `客服#${agent.userId}`,
    })),
  )

  async function onFormSubmit(values: EditForm) {
    const payload = buildPayload(values)
    if (itemId) {
      await onSubmit({
        ticketId: itemId,
        ...payload,
      })
      return
    }
    await onSubmit({
      ...payload,
      source: fixedConversationId ? "conversation" : "manual",
      conversationId: fixedConversationId,
      customerId: fixedCustomerId,
    })
  }

  return (
    <ProjectDialog
      open={open}
      onOpenChange={onOpenChange}
      title={titleOverride || (itemId ? "编辑工单" : "新建工单")}
      description={descriptionOverride || "填写工单基础信息"}
      size="lg"
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
          <FieldGroup>
            <Field data-invalid={!!errors.title}>
              <FieldLabel htmlFor="ticket-title">标题</FieldLabel>
              <FieldContent>
                <Input
                  id="ticket-title"
                  placeholder="请输入工单标题"
                  aria-invalid={!!errors.title}
                  {...register("title")}
                />
                <FieldError errors={[errors.title]} />
              </FieldContent>
            </Field>

            <Field data-invalid={!!errors.description}>
              <FieldLabel htmlFor="ticket-description">描述</FieldLabel>
              <FieldContent>
                <Textarea
                  id="ticket-description"
                  rows={5}
                  placeholder="请输入问题描述"
                  aria-invalid={!!errors.description}
                  {...register("description")}
                />
                <FieldError errors={[errors.description]} />
              </FieldContent>
            </Field>

            <Field>
              <FieldLabel>工单分类</FieldLabel>
              <FieldContent>
                <Controller
                  control={control}
                  name="categoryId"
                  render={({ field }) => (
                    <OptionCombobox
                      value={field.value}
                      onChange={field.onChange}
                      placeholder="请选择工单分类"
                      options={categoryOptions}
                    />
                  )}
                />
              </FieldContent>
            </Field>

            <div className="grid gap-4 md:grid-cols-2">
              <Field data-invalid={!!errors.priority}>
                <FieldLabel>优先级</FieldLabel>
                <FieldContent>
                  <Controller
                    control={control}
                    name="priority"
                    render={({ field }) => (
                      <OptionCombobox
                        value={field.value}
                        onChange={field.onChange}
                        placeholder="请选择优先级"
                        options={[
                          { value: "1", label: "低" },
                          { value: "2", label: "普通" },
                          { value: "3", label: "高" },
                          { value: "4", label: "紧急" },
                        ]}
                      />
                    )}
                  />
                  <FieldError errors={[errors.priority]} />
                </FieldContent>
              </Field>

              <Field data-invalid={!!errors.severity}>
                <FieldLabel>严重度</FieldLabel>
                <FieldContent>
                  <Controller
                    control={control}
                    name="severity"
                    render={({ field }) => (
                      <OptionCombobox
                        value={field.value}
                        onChange={field.onChange}
                        placeholder="请选择严重度"
                        options={[
                          { value: "1", label: "轻微" },
                          { value: "2", label: "严重" },
                          { value: "3", label: "致命" },
                        ]}
                      />
                    )}
                  />
                  <FieldError errors={[errors.severity]} />
                </FieldContent>
              </Field>
            </div>

            <div className="grid gap-4 md:grid-cols-2">
              <Field>
                <FieldLabel>处理团队</FieldLabel>
                <FieldContent>
                  <Controller
                    control={control}
                    name="currentTeamId"
                    render={({ field }) => (
                      <OptionCombobox
                        value={field.value}
                        onChange={field.onChange}
                        placeholder="请选择团队"
                        options={teamOptions}
                      />
                    )}
                  />
                </FieldContent>
              </Field>

              <Field>
                <FieldLabel>处理人</FieldLabel>
                <FieldContent>
                  <Controller
                    control={control}
                    name="currentAssigneeId"
                    render={({ field }) => (
                      <OptionCombobox
                        value={field.value}
                        onChange={field.onChange}
                        placeholder="请选择处理人"
                        options={agentOptions}
                      />
                    )}
                  />
                </FieldContent>
              </Field>
            </div>

            <Field data-invalid={!!errors.dueAt}>
              <FieldLabel htmlFor="ticket-due-at">截止时间</FieldLabel>
              <FieldContent>
                <Input
                  id="ticket-due-at"
                  type="datetime-local"
                  aria-invalid={!!errors.dueAt}
                  {...register("dueAt")}
                />
                <FieldError errors={[errors.dueAt]} />
              </FieldContent>
            </Field>
          </FieldGroup>
        </form>
      )}
    </ProjectDialog>
  )
}
