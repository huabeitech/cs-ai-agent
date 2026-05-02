"use client"

import { CheckIcon, TagIcon } from "lucide-react"
import { useEffect, useMemo, useState } from "react"
import { zodResolver } from "@hookform/resolvers/zod"
import type { Resolver } from "react-hook-form"
import { Controller, useForm } from "react-hook-form"
import { z } from "zod/v4"

import { OptionCombobox } from "@/components/option-combobox"
import { ProjectDialog } from "@/components/project-dialog"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command"
import {
  Field,
  FieldContent,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { Textarea } from "@/components/ui/textarea"
import {
  fetchAgentProfilesAll,
  fetchTagsAll,
  type AdminAgentProfile,
  type TagTree,
} from "@/lib/api/admin"
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

const schema = z.object({
  title: z.string().trim().min(1, "请输入工单标题"),
  description: z.string().trim().min(1, "请输入问题描述"),
  currentAssigneeId: z.coerce.number().int().min(0).optional(),
  tagIds: z.array(z.number().int().positive()).default([]),
})

type EditForm = z.infer<typeof schema>

const editFormResolver = zodResolver(schema as never) as Resolver<
  z.input<typeof schema>,
  undefined,
  z.output<typeof schema>
>

const emptyForm: EditForm = {
  title: "",
  description: "",
  currentAssigneeId: 0,
  tagIds: [],
}

function buildForm(item: TicketItem | null): EditForm {
  if (!item) {
    return emptyForm
  }
  return {
    title: item.title ?? "",
    description: item.description ?? "",
    currentAssigneeId: item.currentAssigneeId ?? 0,
    tagIds: (item.tags ?? []).map((tag) => tag.id),
  }
}

function buildInitialForm(initialValues?: Partial<CreateTicketPayload>): EditForm {
  return {
    title: initialValues?.title?.trim() ?? "",
    description: initialValues?.description?.trim() ?? "",
    currentAssigneeId: initialValues?.currentAssigneeId ?? 0,
    tagIds: initialValues?.tagIds ?? [],
  }
}

function buildPayload(form: EditForm): CreateTicketPayload {
  const currentAssigneeId = form.currentAssigneeId ?? 0
  return {
    title: form.title.trim(),
    description: form.description.trim(),
    currentAssigneeId,
    tagIds: form.tagIds,
  }
}

type FlatTagNode = TagTree & {
  depth: number
  path: string
}

function flattenTagTree(nodes: TagTree[], depth = 0, parentPath = ""): FlatTagNode[] {
  const result: FlatTagNode[] = []
  nodes.forEach((item) => {
    const path = parentPath ? `${parentPath} / ${item.name}` : item.name
    result.push({ ...item, depth, path })
    if (item.children.length > 0) {
      result.push(...flattenTagTree(item.children, depth + 1, path))
    }
  })
  return result
}

type TicketTagSelectorProps = {
  value?: number[]
  onChange: (value: number[]) => void
  availableTags: TagTree[]
}

function TicketTagSelector({ value, onChange, availableTags }: TicketTagSelectorProps) {
  const selectedValues = value ?? []
  const flatTags = useMemo(() => flattenTagTree(availableTags), [availableTags])
  const selectedTagIDs = useMemo(() => new Set(selectedValues), [selectedValues])
  const selectedTags = useMemo(
    () => flatTags.filter((tag) => selectedTagIDs.has(tag.id)),
    [flatTags, selectedTagIDs],
  )

  function handleToggle(tagID: number) {
    if (selectedTagIDs.has(tagID)) {
      onChange(selectedValues.filter((item) => item !== tagID))
      return
    }
    onChange(selectedValues.concat(tagID))
  }

  return (
    <div className="space-y-2">
      <Popover>
        <PopoverTrigger
          render={
            <Button type="button" variant="outline" className="w-full justify-start" />
          }
        >
          <TagIcon className="size-4" />
          {selectedTags.length > 0 ? `已选择 ${selectedTags.length} 个标签` : "请选择工单标签"}
        </PopoverTrigger>
        <PopoverContent align="start" className="w-[320px] p-0">
          <Command>
            <CommandInput placeholder="搜索标签" />
            <CommandList>
              <CommandEmpty>暂无可用标签</CommandEmpty>
              <CommandGroup heading="标签">
                {flatTags.map((tag) => {
                  const checked = selectedTagIDs.has(tag.id)
                  return (
                    <CommandItem
                      key={tag.id}
                      value={`${tag.id} ${tag.path} ${tag.remark}`}
                      onSelect={() => handleToggle(tag.id)}
                    >
                      <CheckIcon className={`mr-2 size-4 ${checked ? "opacity-100" : "opacity-0"}`} />
                      <span className="truncate" style={{ paddingLeft: `${tag.depth * 12}px` }}>
                        {tag.name}
                      </span>
                    </CommandItem>
                  )
                })}
              </CommandGroup>
            </CommandList>
          </Command>
        </PopoverContent>
      </Popover>
      {selectedTags.length > 0 ? (
        <div className="flex flex-wrap gap-1">
          {selectedTags.map((tag) => (
            <Badge key={tag.id} variant="outline">
              {tag.path}
            </Badge>
          ))}
        </div>
      ) : null}
    </div>
  )
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
  const [tags, setTags] = useState<TagTree[]>([])
  const [agents, setAgents] = useState<AdminAgentProfile[]>([])
  const form = useForm<
    z.input<typeof schema>,
    undefined,
    z.output<typeof schema>
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
      const [tagData, agentData] = await Promise.all([
        fetchTagsAll(),
        fetchAgentProfilesAll(),
      ])
      setTags(Array.isArray(tagData) ? tagData : [])
      setAgents(Array.isArray(agentData) ? agentData : [])
    })()
  }, [open])

  const agentOptions = [{ value: "0", label: "不指定处理人" }].concat(
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
              <FieldLabel>处理人</FieldLabel>
              <FieldContent>
                <Controller
                  control={control}
                  name="currentAssigneeId"
                  render={({ field }) => (
                    <OptionCombobox
                      value={String(field.value ?? 0)}
                      onChange={(value) => field.onChange(Number(value))}
                      placeholder="请选择处理人"
                      options={agentOptions}
                    />
                  )}
                />
              </FieldContent>
            </Field>

            <Field>
              <FieldLabel>工单标签</FieldLabel>
              <FieldContent>
                <Controller
                  control={control}
                  name="tagIds"
                  render={({ field }) => (
                    <TicketTagSelector
                      value={field.value}
                      onChange={field.onChange}
                      availableTags={tags}
                    />
                  )}
                />
              </FieldContent>
            </Field>
          </FieldGroup>
        </form>
      )}
    </ProjectDialog>
  )
}
