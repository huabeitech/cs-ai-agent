"use client"

import { useCallback, useEffect, useState } from "react"
import { PlusIcon, RefreshCwIcon, SearchIcon, Trash2Icon } from "lucide-react"
import { zodResolver } from "@hookform/resolvers/zod"
import { Controller, type Resolver, useForm } from "react-hook-form"
import { toast } from "sonner"
import { z } from "zod/v4"

import { ListPagination } from "@/components/list-pagination"
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
  type PageResult,
  type TicketPriorityConfig,
  updateTicketPriorityConfig,
} from "@/lib/api/ticket-config"
import { getEnumOptions } from "@/lib/enums"
import { Status, StatusLabels } from "@/lib/generated/enums"

const listStatusOptions = [
  { value: "all", label: "全部状态" },
  ...getEnumOptions(StatusLabels)
    .filter((item) => Number(item.value) !== Status.Deleted)
    .map((item) => ({ value: String(item.value), label: item.label })),
] as const

const formSchema = z.object({
  name: z.string().trim().min(1, "优先级名称不能为空"),
  sortNo: z.string().trim().min(1, "排序号不能为空").regex(/^-?\d+$/, "请输入整数"),
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
  sortNo: "0",
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
    sortNo: String(item.sortNo),
    firstResponseMinutes: String(item.firstResponseMinutes),
    resolutionMinutes: String(item.resolutionMinutes),
    status: String(item.status) as EditForm["status"],
    remark: item.remark || "",
  }
}

function buildPayload(form: EditForm): CreateTicketPriorityConfigPayload {
  return {
    name: form.name.trim(),
    sortNo: Number(form.sortNo),
    firstResponseMinutes: Number(form.firstResponseMinutes),
    resolutionMinutes: Number(form.resolutionMinutes),
    status: Number(form.status),
    remark: form.remark.trim(),
  }
}

export default function TicketPrioritiesPage() {
  const [keywordInput, setKeywordInput] = useState("")
  const [statusFilterInput, setStatusFilterInput] = useState("all")
  const [keyword, setKeyword] = useState("")
  const [statusFilter, setStatusFilter] = useState("all")
  const [page, setPage] = useState(1)
  const [limit, setLimit] = useState(20)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editingItem, setEditingItem] = useState<TicketPriorityConfig | null>(null)
  const [result, setResult] = useState<PageResult<TicketPriorityConfig>>({
    results: [],
    page: { page: 1, limit: 20, total: 0 },
  })

  const loadData = useCallback(async () => {
    setLoading(true)
    try {
      const data = await fetchTicketPriorityConfigs({
        name: keyword.trim() || undefined,
        status: statusFilter === "all" ? undefined : statusFilter,
        page,
        limit,
      })
      setResult(data)
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "加载工单优先级失败")
    } finally {
      setLoading(false)
    }
  }, [keyword, statusFilter, page, limit])

  useEffect(() => {
    void loadData()
  }, [loadData])

  function applyFilters() {
    setKeyword(keywordInput)
    setStatusFilter(statusFilterInput)
    setPage(1)
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
          <table className="w-full text-sm">
            <thead className="bg-muted/35">
              <tr>
                <th className="px-4 py-3 text-left font-medium">名称</th>
                <th className="px-4 py-3 text-left font-medium">排序</th>
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
              ) : result.results.length > 0 ? (
                result.results.map((item) => (
                  <tr key={item.id} className="border-t">
                    <td className="px-4 py-3">{item.name}</td>
                    <td className="px-4 py-3">{item.sortNo}</td>
                    <td className="px-4 py-3">{item.firstResponseMinutes} 分钟</td>
                    <td className="px-4 py-3">{item.resolutionMinutes} 分钟</td>
                    <td className="px-4 py-3">
                      <Badge variant={item.status === Status.Ok ? "default" : "secondary"}>
                        {item.status === Status.Ok ? "启用" : "停用"}
                      </Badge>
                    </td>
                    <td className="px-4 py-3 text-right">
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" size="sm">
                            操作
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem
                            onClick={() => {
                              setEditingItem(item)
                              setDialogOpen(true)
                            }}
                          >
                            编辑
                          </DropdownMenuItem>
                          <DropdownMenuItem
                            className="text-destructive focus:text-destructive"
                            onClick={() => void handleDelete(item)}
                          >
                            <Trash2Icon className="size-4" />
                            删除
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </td>
                  </tr>
                ))
              ) : (
                <tr>
                  <td colSpan={6} className="h-32 text-center text-muted-foreground">
                    暂无工单优先级
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>

        <ListPagination
          page={result.page.page}
          pageSize={result.page.limit}
          total={result.page.total}
          onPageChange={setPage}
          onPageSizeChange={(nextLimit) => {
            setLimit(nextLimit)
            setPage(1)
          }}
        />
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
      description="优先级同时承载首响与解决时长配置。"
      onConfirm={() => void handleSubmit(async (values) => onSubmit(buildPayload(values)))()}
      confirmLoading={saving}
      confirmText="保存"
      contentClassName="sm:max-w-xl"
    >
      <form id={formId} className="space-y-4" onSubmit={(event) => void handleSubmit(async (values) => onSubmit(buildPayload(values)))(event)}>
        <Field data-invalid={Boolean(errors.name)}>
          <FieldLabel htmlFor="ticket-priority-name">名称</FieldLabel>
          <FieldContent>
            <Input id="ticket-priority-name" placeholder="请输入优先级名称" {...register("name")} />
            {errors.name ? <FieldError errors={[errors.name]} /> : null}
          </FieldContent>
        </Field>

        <Field data-invalid={Boolean(errors.sortNo)}>
          <FieldLabel htmlFor="ticket-priority-sort-no">排序号</FieldLabel>
          <FieldContent>
            <Input id="ticket-priority-sort-no" placeholder="数值越小越靠前" {...register("sortNo")} />
            {errors.sortNo ? <FieldError errors={[errors.sortNo]} /> : null}
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
