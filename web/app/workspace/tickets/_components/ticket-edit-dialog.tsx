"use client"

import { useEffect, useState } from "react"
import { zodResolver } from "@hookform/resolvers/zod"
import { Controller, useForm } from "react-hook-form"
import { z } from "zod"

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
import type {
  CreateTicketPayload,
  TicketItem,
  UpdateTicketPayload,
} from "@/lib/api/ticket"

const schema = z.object({
  title: z.string().trim().min(1, "标题不能为空"),
  description: z.string().trim(),
  priority: z.string().trim().min(1, "请选择优先级"),
  severity: z.string().trim().min(1, "请选择严重度"),
  currentTeamId: z.string().trim(),
  currentAssigneeId: z.string().trim(),
  dueAt: z.string().trim(),
})

type FormValues = z.infer<typeof schema>

const emptyForm: FormValues = {
  title: "",
  description: "",
  priority: "2",
  severity: "1",
  currentTeamId: "",
  currentAssigneeId: "",
  dueAt: "",
}

function buildForm(item?: TicketItem | null): FormValues {
  if (!item) {
    return emptyForm
  }
  return {
    title: item.title ?? "",
    description: item.description ?? "",
    priority: String(item.priority || 2),
    severity: String(item.severity || 1),
    currentTeamId: item.currentTeamId ? String(item.currentTeamId) : "",
    currentAssigneeId: item.currentAssigneeId ? String(item.currentAssigneeId) : "",
    dueAt: item.dueAt ? item.dueAt.replace(" ", "T").slice(0, 16) : "",
  }
}

type TicketEditDialogProps = {
  open: boolean
  item?: TicketItem | null
  fixedConversationId?: number
  fixedCustomerId?: number
  titleOverride?: string
  descriptionOverride?: string
  onOpenChange: (open: boolean) => void
  onSubmit: (payload: CreateTicketPayload | UpdateTicketPayload) => Promise<void>
}

export function TicketEditDialog({
  open,
  item,
  fixedConversationId,
  fixedCustomerId,
  titleOverride,
  descriptionOverride,
  onOpenChange,
  onSubmit,
}: TicketEditDialogProps) {
  const [submitting, setSubmitting] = useState(false)
  const [teams, setTeams] = useState<AdminAgentTeam[]>([])
  const [agents, setAgents] = useState<AdminAgentProfile[]>([])

  const {
    register,
    control,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: buildForm(item),
  })

  useEffect(() => {
    reset(buildForm(item))
  }, [item, reset, open])

  useEffect(() => {
    if (!open) {
      return
    }
    void (async () => {
      const [teamData, agentData] = await Promise.all([
        fetchAgentTeamsAll(),
        fetchAgentProfilesAll(),
      ])
      setTeams(Array.isArray(teamData) ? teamData : [])
      setAgents(Array.isArray(agentData) ? agentData : [])
    })()
  }, [open])

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

  async function submitForm(form: FormValues) {
    setSubmitting(true)
    try {
      const basePayload = {
        title: form.title.trim(),
        description: form.description.trim(),
        priority: Number(form.priority),
        severity: Number(form.severity),
        currentTeamId: form.currentTeamId ? Number(form.currentTeamId) : undefined,
        currentAssigneeId: form.currentAssigneeId
          ? Number(form.currentAssigneeId)
          : undefined,
        dueAt: form.dueAt ? `${form.dueAt.replace("T", " ")}:00` : undefined,
      }
      if (item?.id) {
        await onSubmit({
          ticketId: item.id,
          ...basePayload,
        })
      } else {
        await onSubmit({
          ...basePayload,
          source: fixedConversationId ? "conversation" : "manual",
          conversationId: fixedConversationId,
          customerId: fixedCustomerId,
        })
      }
      onOpenChange(false)
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <ProjectDialog
      open={open}
      onOpenChange={onOpenChange}
      title={titleOverride || (item?.id ? "编辑工单" : "新建工单")}
      description={descriptionOverride || "填写工单基础信息"}
      size="lg"
      footer={
        <>
          <Button
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={submitting}
          >
            取消
          </Button>
          <Button onClick={handleSubmit(submitForm)} disabled={submitting}>
            {submitting ? "提交中..." : item?.id ? "保存" : "创建"}
          </Button>
        </>
      }
    >
      <FieldGroup>
        <Field data-invalid={!!errors.title}>
          <FieldLabel htmlFor="ticket-title">标题</FieldLabel>
          <FieldContent>
            <Input id="ticket-title" placeholder="请输入工单标题" {...register("title")} />
            <FieldError errors={[errors.title]} />
          </FieldContent>
        </Field>

        <Field>
          <FieldLabel htmlFor="ticket-description">描述</FieldLabel>
          <FieldContent>
            <Textarea
              id="ticket-description"
              rows={5}
              placeholder="请输入问题描述"
              {...register("description")}
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

        <Field>
          <FieldLabel htmlFor="ticket-due-at">截止时间</FieldLabel>
          <FieldContent>
            <Input id="ticket-due-at" type="datetime-local" {...register("dueAt")} />
          </FieldContent>
        </Field>
      </FieldGroup>
    </ProjectDialog>
  )
}
