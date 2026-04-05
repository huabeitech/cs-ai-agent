"use client"

import {
  closestCenter,
  DndContext,
  KeyboardSensor,
  MouseSensor,
  TouchSensor,
  useSensor,
  useSensors,
  type DragEndEvent,
} from "@dnd-kit/core"
import {
  arrayMove,
  SortableContext,
  sortableKeyboardCoordinates,
  useSortable,
  verticalListSortingStrategy,
} from "@dnd-kit/sortable"
import { CSS } from "@dnd-kit/utilities"
import { zodResolver } from "@hookform/resolvers/zod"
import { GripVerticalIcon, PlusIcon, RefreshCwIcon, SearchIcon, Trash2Icon } from "lucide-react"
import { useCallback, useEffect, useState, type CSSProperties } from "react"
import { Controller, type Resolver, useForm } from "react-hook-form"
import { toast } from "sonner"
import { z } from "zod/v4"

import { OptionCombobox } from "@/components/option-combobox"
import { ProjectDialog } from "@/components/project-dialog"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Field, FieldContent, FieldError, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"
import {
  createTicketPriorityConfig,
  deleteTicketPriorityConfig,
  fetchTicketPriorityConfigs,
  type CreateTicketPriorityConfigPayload,
  type TicketPriorityConfig,
  updateTicketPriorityConfig,
  updateTicketPriorityConfigSort,
} from "@/lib/api/ticket-config"
import { getEnumOptions } from "@/lib/enums"
import { Status, StatusLabels } from "@/lib/generated/enums"
import { cn } from "@/lib/utils"

const listStatusOptions = [
  { value: "all", label: "全部状态" },
  ...getEnumOptions(StatusLabels)
    .filter((item) => Number(item.value) !== Status.Deleted)
    .map((item) => ({ value: String(item.value), label: item.label })),
] as const

const formSchema = z.object({
  name: z.string().trim().min(1, "优先级名称不能为空"),
  firstResponseMinutes: z.string().trim().min(1, "首响时长不能为空").regex(/^\d+$/, "请输入正整数"),
  resolutionMinutes: z.string().trim().min(1, "解决时长不能为空").regex(/^\d+$/, "请输入正整数"),
  status: z.enum([String(Status.Ok), String(Status.Disabled)], {
    message: "请选择状态",
  }),
  remark: z.string().trim(),
})

type EditForm = z.infer<typeof formSchema>

const resolver = zodResolver(formSchema as never) as Resolver<
  z.input<typeof formSchema>,
  undefined,
  z.output<typeof formSchema>
>

const emptyForm: EditForm = {
  name: "",
  firstResponseMinutes: "30",
  resolutionMinutes: "1440",
  status: String(Status.Ok),
  remark: "",
}

function buildForm(item: TicketPriorityConfig | null): EditForm {
  if (!item) {
    return emptyForm
  }
  return {
    name: item.name,
    firstResponseMinutes: String(item.firstResponseMinutes),
    resolutionMinutes: String(item.resolutionMinutes),
    status: String(item.status) as EditForm["status"],
    remark: item.remark || "",
  }
}

function buildPayload(form: EditForm): CreateTicketPriorityConfigPayload {
  return {
    name: form.name.trim(),
    firstResponseMinutes: Number(form.firstResponseMinutes),
    resolutionMinutes: Number(form.resolutionMinutes),
    status: Number(form.status),
    remark: form.remark.trim(),
  }
}

type SortablePriorityRowProps = {
  item: TicketPriorityConfig
  disabled: boolean
  onEdit: (item: TicketPriorityConfig) => void
  onDelete: (item: TicketPriorityConfig) => void
}

function SortablePriorityRow({ item, disabled, onEdit, onDelete }: SortablePriorityRowProps) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({
    id: item.id,
    disabled,
  })

  const style: CSSProperties = {
    transform: CSS.Transform.toString(transform),
    transition,
  }

  return (
    <tr
      ref={setNodeRef}
      style={style}
      className={cn("border-t", isDragging && "relative z-10 bg-muted/60 shadow-sm")}
    >
      <td className="w-14 px-4 py-3">
        <Button
          type="button"
          variant="ghost"
          size="icon"
          className="size-8 cursor-grab active:cursor-grabbing"
          disabled={disabled}
          aria-label={`拖拽排序 ${item.name}`}
          {...attributes}
          {...listeners}
        >
          <GripVerticalIcon className="size-4 text-muted-foreground" />
        </Button>
      </td>
      <td className="px-4 py-3">{item.name}</td>
      <td className="px-4 py-3">{item.firstResponseMinutes} 分钟</td>
      <td className="px-4 py-3">{item.resolutionMinutes} 分钟</td>
      <td className="px-4 py-3">
        <Badge variant={item.status === Status.Ok ? "default" : "secondary"}>
          {item.status === Status.Ok ? "启用" : "停用"}
        </Badge>
      </td>
      <td className="px-4 py-3 text-right">
        <DropdownMenu>
          <DropdownMenuTrigger>操作</DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem onClick={() => onEdit(item)}>编辑</DropdownMenuItem>
            <DropdownMenuItem className="text-destructive focus:text-destructive" onClick={() => void onDelete(item)}>
              <Trash2Icon className="size-4" />
              删除
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </td>
    </tr>
  )
}

export default function TicketPrioritiesPage() {
  const [keywordInput, setKeywordInput] = useState("")
  const [statusFilterInput, setStatusFilterInput] = useState("all")
  const [keyword, setKeyword] = useState("")
  const [statusFilter, setStatusFilter] = useState("all")
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [sorting, setSorting] = useState(false)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editingItem, setEditingItem] = useState<TicketPriorityConfig | null>(null)
  const [items, setItems] = useState<TicketPriorityConfig[]>([])

  const sensors = useSensors(
    useSensor(MouseSensor, { activationConstraint: { distance: 6 } }),
    useSensor(TouchSensor, { activationConstraint: { delay: 120, tolerance: 8 } }),
    useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates }),
  )

  const loadData = useCallback(async () => {
    setLoading(true)
    try {
      const data = await fetchTicketPriorityConfigs({
        name: keyword.trim() || undefined,
        status: statusFilter === "all" ? undefined : statusFilter,
      })
      setItems(Array.isArray(data) ? data : [])
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "加载工单优先级失败")
    } finally {
      setLoading(false)
    }
  }, [keyword, statusFilter])

  useEffect(() => {
    void loadData()
  }, [loadData])

  function applyFilters() {
    setKeyword(keywordInput)
    setStatusFilter(statusFilterInput)
  }

  async function handleSubmit(payload: CreateTicketPriorityConfigPayload) {
    if (saving) {
      return
    }
    setSaving(true)
    try {
      if (editingItem) {
        await updateTicketPriorityConfig({ id: editingItem.id, ...payload })
        toast.success(`已更新工单优先级：${payload.name}`)
      } else {
        await createTicketPriorityConfig(payload)
        toast.success(`已创建工单优先级：${payload.name}`)
      }
      setDialogOpen(false)
      setEditingItem(null)
      await loadData()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "保存工单优先级失败")
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete(item: TicketPriorityConfig) {
    try {
      await deleteTicketPriorityConfig(item.id)
      toast.success(`已删除工单优先级：${item.name}`)
      await loadData()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "删除工单优先级失败")
    }
  }

  async function handleDragEnd(event: DragEndEvent) {
    const { active, over } = event
    if (!over || active.id === over.id || sorting || loading) {
      return
    }
    const previousResults = items
    const oldIndex = previousResults.findIndex((item) => item.id === active.id)
    const newIndex = previousResults.findIndex((item) => item.id === over.id)
    if (oldIndex < 0 || newIndex < 0) {
      return
    }
    const nextResults = arrayMove(previousResults, oldIndex, newIndex)
    setItems(nextResults)
    setSorting(true)
    try {
      await updateTicketPriorityConfigSort(nextResults.map((item) => item.id))
      toast.success("工单优先级排序已更新")
      await loadData()
    } catch (error) {
      setItems(previousResults)
      toast.error(error instanceof Error ? error.message : "更新排序失败")
    } finally {
      setSorting(false)
    }
  }

  return (
    <>
      <div className="flex flex-1 flex-col gap-6 p-4 lg:p-6">
        <div className="flex flex-col gap-3 xl:flex-row xl:items-center xl:justify-between">
          <div>
            <h1 className="text-xl font-semibold">工单优先级</h1>
            <p className="mt-1 text-sm text-muted-foreground">统一维护优先级名称、排序和 SLA 时长</p>
          </div>
          <Button
            onClick={() => {
              setEditingItem(null)
              setDialogOpen(true)
            }}
          >
            <PlusIcon className="size-4" />
            新建优先级
          </Button>
        </div>

        <div className="flex flex-col gap-3 xl:flex-row xl:items-center">
          <div className="relative min-w-72">
            <SearchIcon className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              value={keywordInput}
              onChange={(event) => setKeywordInput(event.target.value)}
              onKeyDown={(event) => {
                if (event.key === "Enter") {
                  event.preventDefault()
                  applyFilters()
                }
              }}
              placeholder="按优先级名称筛选"
              className="pl-9"
            />
          </div>
          <div className="w-full xl:w-40">
            <OptionCombobox
              value={statusFilterInput}
              onChange={setStatusFilterInput}
              placeholder="全部状态"
              options={listStatusOptions.map((item) => ({
                value: item.value,
                label: item.label,
              }))}
            />
          </div>
          <Button variant="outline" onClick={applyFilters} disabled={loading}>
            <SearchIcon className="size-4" />
            查询
          </Button>
          <Button variant="outline" onClick={() => void loadData()} disabled={loading}>
            <RefreshCwIcon className="size-4" />
          </Button>
        </div>

        <div className="overflow-hidden rounded-lg border bg-background">
          <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={handleDragEnd}>
            <table className="w-full text-sm">
              <thead className="bg-muted/35">
                <tr>
                  <th className="w-14 px-4 py-3 text-left font-medium"></th>
                  <th className="px-4 py-3 text-left font-medium">名称</th>
                  <th className="px-4 py-3 text-left font-medium">首响时长</th>
                  <th className="px-4 py-3 text-left font-medium">解决时长</th>
                  <th className="px-4 py-3 text-left font-medium">状态</th>
                  <th className="px-4 py-3 text-right font-medium">操作</th>
                </tr>
              </thead>
              <tbody>
                {loading ? (
                  <tr>
                    <td colSpan={6} className="h-32 text-center text-muted-foreground">
                      加载中...
                    </td>
                  </tr>
                ) : items.length > 0 ? (
                  <SortableContext items={items.map((item) => item.id)} strategy={verticalListSortingStrategy}>
                    {items.map((item) => (
                      <SortablePriorityRow
                        key={item.id}
                        item={item}
                        disabled={sorting}
                        onEdit={(current) => {
                          setEditingItem(current)
                          setDialogOpen(true)
                        }}
                        onDelete={handleDelete}
                      />
                    ))}
                  </SortableContext>
                ) : (
                  <tr>
                    <td colSpan={6} className="h-32 text-center text-muted-foreground">
                      暂无工单优先级
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </DndContext>
        </div>

      </div>

      <TicketPriorityEditDialog
        open={dialogOpen}
        saving={saving}
        item={editingItem}
        onOpenChange={(nextOpen) => {
          setDialogOpen(nextOpen)
          if (!nextOpen) {
            setEditingItem(null)
          }
        }}
        onSubmit={handleSubmit}
      />
    </>
  )
}

type TicketPriorityEditDialogProps = {
  open: boolean
  saving: boolean
  item: TicketPriorityConfig | null
  onOpenChange: (open: boolean) => void
  onSubmit: (payload: CreateTicketPriorityConfigPayload) => Promise<void>
}

function TicketPriorityEditDialog({
  open,
  saving,
  item,
  onOpenChange,
  onSubmit,
}: TicketPriorityEditDialogProps) {
  const formId = "ticket-priority-edit-form"
  const form = useForm<
    z.input<typeof formSchema>,
    undefined,
    z.output<typeof formSchema>
  >({
    resolver,
    defaultValues: buildForm(item),
  })
  const {
    register,
    control,
    handleSubmit,
    reset,
    formState: { errors },
  } = form

  useEffect(() => {
    reset(buildForm(item))
  }, [item, reset])

  return (
    <ProjectDialog
      open={open}
      onOpenChange={onOpenChange}
      title={item ? "编辑工单优先级" : "新建工单优先级"}
      description="优先级同时承载首响与解决时长配置。排序请在列表中拖动调整。"
      size="md"
      footer={
        <>
          <Button type="button" variant="outline" onClick={() => onOpenChange(false)} disabled={saving}>
            取消
          </Button>
          <Button type="submit" form={formId} disabled={saving}>
            {saving ? "保存中..." : item ? "保存" : "创建"}
          </Button>
        </>
      }
    >
      <form id={formId} className="space-y-4" onSubmit={handleSubmit(async (values) => onSubmit(buildPayload(values)))}>
        <Field data-invalid={Boolean(errors.name)}>
          <FieldLabel htmlFor="ticket-priority-name">名称</FieldLabel>
          <FieldContent>
            <Input id="ticket-priority-name" placeholder="请输入优先级名称" {...register("name")} />
            {errors.name ? <FieldError errors={[errors.name]} /> : null}
          </FieldContent>
        </Field>

        <div className="grid gap-4 md:grid-cols-2">
          <Field data-invalid={Boolean(errors.firstResponseMinutes)}>
            <FieldLabel htmlFor="ticket-priority-first-response">首响时长</FieldLabel>
            <FieldContent>
              <Input id="ticket-priority-first-response" placeholder="分钟" {...register("firstResponseMinutes")} />
              {errors.firstResponseMinutes ? <FieldError errors={[errors.firstResponseMinutes]} /> : null}
            </FieldContent>
          </Field>

          <Field data-invalid={Boolean(errors.resolutionMinutes)}>
            <FieldLabel htmlFor="ticket-priority-resolution">解决时长</FieldLabel>
            <FieldContent>
              <Input id="ticket-priority-resolution" placeholder="分钟" {...register("resolutionMinutes")} />
              {errors.resolutionMinutes ? <FieldError errors={[errors.resolutionMinutes]} /> : null}
            </FieldContent>
          </Field>
        </div>

        <Field data-invalid={Boolean(errors.status)}>
          <FieldLabel>状态</FieldLabel>
          <FieldContent>
            <Controller
              control={control}
              name="status"
              render={({ field }) => (
                <OptionCombobox
                  value={field.value}
                  onChange={field.onChange}
                  placeholder="请选择状态"
                  options={[
                    { value: String(Status.Ok), label: "启用" },
                    { value: String(Status.Disabled), label: "停用" },
                  ]}
                />
              )}
            />
            {errors.status ? <FieldError errors={[errors.status]} /> : null}
          </FieldContent>
        </Field>

        <Field data-invalid={Boolean(errors.remark)}>
          <FieldLabel htmlFor="ticket-priority-remark">备注</FieldLabel>
          <FieldContent>
            <Textarea id="ticket-priority-remark" rows={4} placeholder="可选" {...register("remark")} />
            {errors.remark ? <FieldError errors={[errors.remark]} /> : null}
          </FieldContent>
        </Field>
      </form>
    </ProjectDialog>
  )
}
