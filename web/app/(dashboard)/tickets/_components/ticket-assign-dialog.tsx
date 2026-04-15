"use client"

import { useEffect, useState } from "react"
import { zodResolver } from "@hookform/resolvers/zod"
import { Controller, Resolver, useForm } from "react-hook-form"
import { toast } from "sonner"
import { z } from "zod/v4"

import { OptionCombobox } from "@/components/option-combobox"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Field, FieldContent, FieldError, FieldLabel } from "@/components/ui/field"
import { Textarea } from "@/components/ui/textarea"
import {
  fetchAgentProfilesAll,
  fetchAgentTeamsAll,
  type AdminAgentProfile,
  type AdminAgentTeam,
} from "@/lib/api/admin"
import { assignTicket, batchAssignTickets } from "@/lib/api/ticket"

const schema = z.object({
  toUserId: z.string().trim().min(1, "请选择处理人"),
  toTeamId: z.string().trim(),
  reason: z.string().trim(),
})

type FormValues = z.infer<typeof schema>

const resolver = zodResolver(schema as never) as Resolver<
  z.input<typeof schema>,
  undefined,
  z.output<typeof schema>
>

const emptyForm: FormValues = {
  toUserId: "",
  toTeamId: "",
  reason: "",
}

type TicketAssignDialogProps = {
  open: boolean
  ticketId: number | null
  ticketIds?: number[]
  currentTeamId?: number
  currentAssigneeId?: number
  onOpenChange: (open: boolean) => void
  onSuccess?: () => Promise<void> | void
}

export function TicketAssignDialog({
  open,
  ticketId,
  ticketIds,
  currentTeamId,
  currentAssigneeId,
  onOpenChange,
  onSuccess,
}: TicketAssignDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      {open ? (
        <TicketAssignDialogBody
          key={ticketId ?? "ticket-assign"}
          ticketId={ticketId}
          ticketIds={ticketIds}
          currentTeamId={currentTeamId}
          currentAssigneeId={currentAssigneeId}
          onOpenChange={onOpenChange}
          onSuccess={onSuccess}
        />
      ) : null}
    </Dialog>
  )
}

function TicketAssignDialogBody({
  ticketId,
  ticketIds,
  currentTeamId,
  currentAssigneeId,
  onOpenChange,
  onSuccess,
}: Omit<TicketAssignDialogProps, "open">) {
  const [saving, setSaving] = useState(false)
  const [loading, setLoading] = useState(false)
  const [teams, setTeams] = useState<AdminAgentTeam[]>([])
  const [agents, setAgents] = useState<AdminAgentProfile[]>([])

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
    reset({
      toUserId: currentAssigneeId ? String(currentAssigneeId) : "",
      toTeamId: currentTeamId ? String(currentTeamId) : "",
      reason: "",
    })
  }, [currentAssigneeId, currentTeamId, reset, ticketId])

  useEffect(() => {
    setLoading(true)
    Promise.all([fetchAgentTeamsAll(), fetchAgentProfilesAll()])
      .then(([teamData, agentData]) => {
        setTeams(Array.isArray(teamData) ? teamData : [])
        setAgents(Array.isArray(agentData) ? agentData : [])
      })
      .catch((error) => {
        toast.error(error instanceof Error ? error.message : "加载处理人失败")
      })
      .finally(() => setLoading(false))
  }, [])

  async function onFormSubmit(values: FormValues) {
    const validTicketIds = (ticketIds ?? []).filter((item) => item > 0)
    if (!ticketId && validTicketIds.length === 0) {
      toast.error("请选择工单")
      return
    }
    setSaving(true)
    try {
      if (validTicketIds.length > 0) {
        await batchAssignTickets({
          ticketIds: validTicketIds,
          toUserId: Number(values.toUserId),
          toTeamId: values.toTeamId ? Number(values.toTeamId) : undefined,
          reason: values.reason.trim() || undefined,
        })
        toast.success(`已批量指派 ${validTicketIds.length} 张工单`)
      } else {
        await assignTicket({
          ticketId: ticketId!,
          toUserId: Number(values.toUserId),
          toTeamId: values.toTeamId ? Number(values.toTeamId) : undefined,
          reason: values.reason.trim() || undefined,
        })
        toast.success("处理人已更新")
      }
      onOpenChange(false)
      await onSuccess?.()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "指派工单失败")
    } finally {
      setSaving(false)
    }
  }

  return (
    <DialogContent className="max-w-lg gap-0 p-0 sm:max-w-lg">
      <DialogHeader className="px-6 pt-6">
        <DialogTitle>{ticketIds?.length ? `批量指派工单（${ticketIds.length}）` : "指派工单"}</DialogTitle>
      </DialogHeader>
      <form onSubmit={handleSubmit(onFormSubmit)}>
        <div className="space-y-4 p-6">
          <Field>
            <FieldLabel>处理团队</FieldLabel>
            <FieldContent>
              <Controller
                control={control}
                name="toTeamId"
                render={({ field }) => (
                  <OptionCombobox
                    value={field.value}
                    onChange={field.onChange}
                    placeholder={loading ? "加载中..." : "选择处理团队"}
                    options={[
                      { value: "", label: "不指定团队" },
                      ...teams.map((team) => ({
                        value: String(team.id),
                        label: team.name,
                      })),
                    ]}
                  />
                )}
              />
            </FieldContent>
          </Field>
          <Field data-invalid={!!errors.toUserId}>
            <FieldLabel>处理人</FieldLabel>
            <FieldContent>
              <Controller
                control={control}
                name="toUserId"
                render={({ field }) => (
                  <OptionCombobox
                    value={field.value}
                    onChange={field.onChange}
                    placeholder={loading ? "加载中..." : "选择处理人"}
                    options={agents.map((agent) => ({
                      value: String(agent.userId),
                      label:
                        agent.displayName ||
                        agent.nickname ||
                        agent.username ||
                        `客服#${agent.userId}`,
                    }))}
                  />
                )}
              />
              <FieldError errors={[errors.toUserId]} />
            </FieldContent>
          </Field>
          <Field data-invalid={!!errors.reason}>
            <FieldLabel>说明</FieldLabel>
            <FieldContent>
              <Textarea rows={4} placeholder="填写指派说明" {...register("reason")} />
              <FieldError errors={[errors.reason]} />
            </FieldContent>
          </Field>
        </div>
        <DialogFooter className="mx-0 mb-0 px-6 py-4">
          <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
            取消
          </Button>
          <Button type="submit" disabled={saving}>
            {saving ? "提交中..." : "确认指派"}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  )
}
