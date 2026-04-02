"use client"
import { zodResolver } from "@hookform/resolvers/zod"
import { useCallback, useEffect, useState } from "react"
import { Controller, Resolver, useForm } from "react-hook-form"
import { toast } from "sonner"
import { z } from "zod/v4"

import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import {
  Field,
  FieldContent,
  FieldError,
  FieldLabel,
} from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { OptionCombobox } from "@/components/option-combobox"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Textarea } from "@/components/ui/textarea"
import {
  type AdminAgentTeam,
  type AdminAgentTeamSchedule,
  type CreateAdminAgentTeamSchedulePayload,
  fetchAgentTeamSchedule,
  fetchAgentTeamsAll
} from "@/lib/api/admin"

type ScheduleEditDialogProps = {
  open: boolean
  saving: boolean
  itemId: number | null
  onOpenChange: (open: boolean) => void
  onSubmit: (payload: CreateAdminAgentTeamSchedulePayload) => Promise<void>
}

const sourceTypeOptions = [
  { value: "manual", label: "手工录入" },
  { value: "batch_import", label: "批量导入" },
  { value: "template_generate", label: "模板生成" },
] as const

const emptyForm: EditForm = {
  teamId: "",
  startAt: "",
  endAt: "",
  sourceType: "manual",
  remark: "",
}

const editFormSchema = z.object({
  teamId: z.string().trim().regex(/^\d+$/, "请选择客服组"),
  startAt: z.string().trim().min(1, "开始时间不能为空"),
  endAt: z.string().trim().min(1, "结束时间不能为空"),
  sourceType: z.enum(["manual", "batch_import", "template_generate"], { message: "请选择排班来源" }),
  remark: z.string().trim(),
})

type EditForm = z.infer<typeof editFormSchema>
const editFormResolver = zodResolver(editFormSchema as never) as Resolver<
  z.input<typeof editFormSchema>,
  undefined,
  z.output<typeof editFormSchema>
>

function toDateTimeLocal(value?: string) {
  if (!value) {
    return ""
  }
  return value.replace(" ", "T").slice(0, 16)
}

function buildForm(item: AdminAgentTeamSchedule | null): EditForm {
  if (!item) {
    return emptyForm
  }
  return {
    teamId: String(item.teamId),
    startAt: toDateTimeLocal(item.startAt),
    endAt: toDateTimeLocal(item.endAt),
    sourceType: item.sourceType as EditForm["sourceType"],
    remark: item.remark || "",
  }
}

function buildPayload(form: EditForm): CreateAdminAgentTeamSchedulePayload {
  return {
    teamId: Number(form.teamId),
    startAt: form.startAt.trim(),
    endAt: form.endAt.trim(),
    sourceType: form.sourceType,
    remark: form.remark.trim(),
  }
}

export function EditDialog({
  open,
  saving,
  itemId,
  onOpenChange,
  onSubmit,
}: ScheduleEditDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      {open ? (
        <ScheduleEditDialogBody
          key={itemId ? `edit-${itemId}` : "create"}
          itemId={itemId}
          saving={saving}
          onOpenChange={onOpenChange}
          onSubmit={onSubmit}
        />
      ) : null}
    </Dialog>
  )
}

type ScheduleEditDialogBodyProps = Omit<ScheduleEditDialogProps, "open">

function ScheduleEditDialogBody({
  saving,
  itemId,
  onOpenChange,
  onSubmit,
}: ScheduleEditDialogBodyProps) {
  const [teams, setTeams] = useState<AdminAgentTeam[]>([])
  const [loading, setLoading] = useState(false)
  const loadOptions = useCallback(async () => {
    try {
      const teamsData = await fetchAgentTeamsAll()
      setTeams(teamsData)
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "加载选项失败")
    }
  }, [])
  const form = useForm<
    z.input<typeof editFormSchema>,
    undefined,
    z.output<typeof editFormSchema>
  >({
    resolver: editFormResolver,
    defaultValues: emptyForm,
  })
  const {
    control,
    handleSubmit,
    reset,
    register,
    formState: { errors },
  } = form

  useEffect(() => {
    async function loadDetail() {
      if (!itemId) {
        reset(emptyForm)
        return
      }
      setLoading(true)
      try {
        const data = await fetchAgentTeamSchedule(itemId)
        reset(buildForm(data))
      } catch (error) {
        toast.error(error instanceof Error ? error.message : "加载客服组排班详情失败")
      } finally {
        setLoading(false)
      }
    }
    void loadDetail()
  }, [itemId, reset])

  useEffect(() => {
    void loadOptions()
  }, [loadOptions])

  async function onFormSubmit(values: EditForm) {
    await onSubmit(buildPayload(values))
  }

  return (
    <DialogContent className="max-w-xl gap-0 p-0 sm:max-w-xl">
      <DialogHeader className="px-6 pt-6">
        <DialogTitle>{itemId ? "编辑客服组排班" : "新建客服组排班"}</DialogTitle>
      </DialogHeader>
      {loading ? (
        <div className="flex items-center justify-center py-12">
          <div className="text-muted-foreground">加载中...</div>
        </div>
      ) : (
        <form onSubmit={handleSubmit(onFormSubmit)}>
          <div className="space-y-4 p-6">
            <div className="grid grid-cols-1 gap-4">
              <Field data-invalid={!!errors.teamId}>
                <FieldLabel>客服组</FieldLabel>
                <FieldContent>
                  <Controller
                    control={control}
                    name="teamId"
                    render={({ field }) => (
                      <OptionCombobox
                        value={field.value}
                        options={teams.map((team) => ({
                          value: String(team.id),
                          label: team.name,
                        }))}
                        placeholder="请选择客服组"
                        searchPlaceholder="搜索客服组"
                        onChange={field.onChange}
                      />
                    )}
                  />
                  <FieldError errors={[errors.teamId]} />
                </FieldContent>
              </Field>
            </div>
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <Field data-invalid={!!errors.startAt}>
                <FieldLabel htmlFor="agent-team-schedule-start-at">开始时间</FieldLabel>
                <FieldContent>
                  <Input id="agent-team-schedule-start-at" type="datetime-local" {...register("startAt")} />
                  <FieldError errors={[errors.startAt]} />
                </FieldContent>
              </Field>
              <Field data-invalid={!!errors.endAt}>
                <FieldLabel htmlFor="agent-team-schedule-end-at">结束时间</FieldLabel>
                <FieldContent>
                  <Input id="agent-team-schedule-end-at" type="datetime-local" {...register("endAt")} />
                  <FieldError errors={[errors.endAt]} />
                </FieldContent>
              </Field>
            </div>
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <Field data-invalid={!!errors.sourceType}>
                <FieldLabel>排班来源</FieldLabel>
                <FieldContent>
                  <Controller
                    control={control}
                    name="sourceType"
                    render={({ field }) => (
                      <Select value={field.value} onValueChange={field.onChange} modal={false}>
                        <SelectTrigger className="w-full">
                          <SelectValue>
                            {sourceTypeOptions.find((item) => item.value === field.value)?.label ?? "请选择来源"}
                          </SelectValue>
                        </SelectTrigger>
                        <SelectContent>
                          {sourceTypeOptions.map((option) => (
                            <SelectItem key={option.value} value={option.value}>
                              {option.label}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    )}
                  />
                  <FieldError errors={[errors.sourceType]} />
                </FieldContent>
              </Field>
            </div>
            <Field>
              <FieldLabel htmlFor="agent-team-schedule-remark">备注</FieldLabel>
              <FieldContent>
                <Textarea id="agent-team-schedule-remark" rows={4} placeholder="请输入备注" {...register("remark")} />
              </FieldContent>
            </Field>
          </div>
          <DialogFooter className="mx-0 mb-0 px-6 py-4">
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)} disabled={saving}>
              取消
            </Button>
            <Button type="submit" disabled={saving || loading}>
              {saving ? "保存中..." : "保存"}
            </Button>
          </DialogFooter>
        </form>
      )}
    </DialogContent>
  )
}
