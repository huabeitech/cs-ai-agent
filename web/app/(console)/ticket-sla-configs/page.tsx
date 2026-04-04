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
  createTicketSLAConfig,
  deleteTicketSLAConfig,
  fetchTicketSLAConfigs,
  updateTicketSLAConfig,
  type CreateTicketSLAConfigPayload,
  type PageResult,
  type TicketSLAConfig,
} from "@/lib/api/ticket-config"
import { getEnumLabel, getEnumOptions } from "@/lib/enums"
import { Status, StatusLabels } from "@/lib/generated/enums"

const listStatusOptions = [
  { value: "all", label: "全部状态" },
  ...getEnumOptions(StatusLabels)
    .filter((item) => Number(item.value) !== Status.Deleted)
    .map((item) => ({ value: String(item.value), label: item.label })),
] as const

const priorityOptions = [
  { value: "1", label: "低" },
  { value: "2", label: "普通" },
  { value: "3", label: "高" },
  { value: "4", label: "紧急" },
]

const formSchema = z.object({
  name: z.string().trim().min(1, "配置名称不能为空"),
  priority: z.enum(["1", "2", "3", "4"], { message: "请选择优先级" }),
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
  priority: "2",
  firstResponseMinutes: "60",
  resolutionMinutes: "720",
  status: String(Status.Ok),
  remark: "",
}

function buildForm(item: TicketSLAConfig | null): EditForm {
  if (!item) {
    return emptyForm
  }
  return {
    name: item.name,
    priority: String(item.priority) as EditForm["priority"],
    firstResponseMinutes: String(item.firstResponseMinutes),
    resolutionMinutes: String(item.resolutionMinutes),
    status: String(item.status) as EditForm["status"],
    remark: item.remark || "",
  }
}

function buildPayload(form: EditForm): CreateTicketSLAConfigPayload {
  return {
    name: form.name.trim(),
    priority: Number(form.priority),
    firstResponseMinutes: Number(form.firstResponseMinutes),
    resolutionMinutes: Number(form.resolutionMinutes),
    status: Number(form.status),
    remark: form.remark.trim(),
  }
}

function priorityLabel(priority: number) {
  return priorityOptions.find((item) => Number(item.value) === priority)?.label ?? String(priority)
}

export default function TicketSLAConfigsPage() {
  const [keywordInput, setKeywordInput] = useState("")
  const [statusFilterInput, setStatusFilterInput] = useState("all")
  const [keyword, setKeyword] = useState("")
  const [statusFilter, setStatusFilter] = useState("all")
  const [page, setPage] = useState(1)
  const [limit, setLimit] = useState(20)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editingItem, setEditingItem] = useState<TicketSLAConfig | null>(null)
  const [result, setResult] = useState<PageResult<TicketSLAConfig>>({
    results: [],
    page: { page: 1, limit: 20, total: 0 },
  })

  const loadData = useCallback(async () => {
    setLoading(true)
    try {
      const data = await fetchTicketSLAConfigs({
        name: keyword.trim() || undefined,
        status: statusFilter === "all" ? undefined : statusFilter,
        page,
        limit,
      })
      setResult(data)
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "加载工单SLA配置失败")
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

  async function handleSubmit(payload: CreateTicketSLAConfigPayload) {
    if (saving) {
      return
    }
    setSaving(true)
    try {
      if (editingItem) {
        await updateTicketSLAConfig({ id: editingItem.id, ...payload })
        toast.success(`已更新工单SLA配置：${payload.name}`)
      } else {
        await createTicketSLAConfig(payload)
        toast.success(`已创建工单SLA配置：${payload.name}`)
      }
      setDialogOpen(false)
      setEditingItem(null)
      await loadData()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "保存工单SLA配置失败")
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete(item: TicketSLAConfig) {
    try {
      await deleteTicketSLAConfig(item.id)
      toast.success(`已删除工单SLA配置：${item.name}`)
      await loadData()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "删除工单SLA配置失败")
    }
  }

  return (
    <>
      <div className="flex flex-1 flex-col gap-6 p-4 lg:p-6">
        <div className="flex flex-col gap-3 xl:flex-row xl:items-center xl:justify-between">
          <div>
            <h1 className="text-xl font-semibold">工单SLA配置</h1>
            <p className="mt-1 text-sm text-muted-foreground">维护不同优先级对应的首响与解决时长</p>
          </div>
          <Button onClick={() => {
            setEditingItem(null)
            setDialogOpen(true)
          }}>
            <PlusIcon className="size-4" />
            新建SLA配置
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
              placeholder="按配置名称筛选"
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
                <th className="px-4 py-3 text-left font-medium">优先级</th>
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
                    <td className="px-4 py-3">{priorityLabel(item.priority)}</td>
                    <td className="px-4 py-3">{item.firstResponseMinutes} 分钟</td>
                    <td className="px-4 py-3">{item.resolutionMinutes} 分钟</td>
                    <td className="px-4 py-3">
                      <Badge variant={item.status === Status.Ok ? "default" : "secondary"}>
                        {getEnumLabel(StatusLabels, item.status as Status)}
                      </Badge>
                    </td>
                    <td className="px-4 py-3 text-right">
                      <DropdownMenu>
                        <DropdownMenuTrigger render={<Button variant="ghost" size="sm" />}>
                          操作
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
                            className="text-destructive"
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
                    暂无工单SLA配置
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>

        <ListPagination
          page={result.page.page}
          total={result.page.total}
          limit={result.page.limit}
          loading={loading}
          onPageChange={setPage}
          onLimitChange={(value) => {
            setLimit(value)
            setPage(1)
          }}
        />
      </div>

      <TicketSLAConfigEditDialog
        open={dialogOpen}
        saving={saving}
        item={editingItem}
        onOpenChange={(open) => {
          if (!saving) {
            setDialogOpen(open)
            if (!open) {
              setEditingItem(null)
            }
          }
        }}
        onSubmit={handleSubmit}
      />
    </>
  )
}

type TicketSLAConfigEditDialogProps = {
  open: boolean
  saving: boolean
  item: TicketSLAConfig | null
  onOpenChange: (open: boolean) => void
  onSubmit: (payload: CreateTicketSLAConfigPayload) => Promise<void>
}

function TicketSLAConfigEditDialog({
  open,
  saving,
  item,
  onOpenChange,
  onSubmit,
}: TicketSLAConfigEditDialogProps) {
  const formId = "ticket-sla-config-edit-form"
  const form = useForm<
    z.input<typeof formSchema>,
    undefined,
    z.output<typeof formSchema>
  >({
    resolver,
    defaultValues: emptyForm,
  })
  const {
    control,
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = form

  useEffect(() => {
    if (!open) {
      return
    }
    reset(buildForm(item))
  }, [open, item, reset])

  return (
    <ProjectDialog
      open={open}
      onOpenChange={onOpenChange}
      title={item ? "编辑工单SLA配置" : "新建工单SLA配置"}
      size="md"
      allowFullscreen
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
      <form id={formId} onSubmit={handleSubmit(async (values) => onSubmit(buildPayload(values)))} className="space-y-4">
        <Field data-invalid={!!errors.name}>
          <FieldLabel htmlFor="ticket-sla-config-name">配置名称</FieldLabel>
          <FieldContent>
            <Input id="ticket-sla-config-name" placeholder="请输入配置名称" {...register("name")} />
            <FieldError errors={[errors.name]} />
          </FieldContent>
        </Field>
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
                  options={priorityOptions}
                />
              )}
            />
            <FieldError errors={[errors.priority]} />
          </FieldContent>
        </Field>
        <div className="grid gap-4 md:grid-cols-2">
          <Field data-invalid={!!errors.firstResponseMinutes}>
            <FieldLabel htmlFor="ticket-sla-first">首响时长(分钟)</FieldLabel>
            <FieldContent>
              <Input id="ticket-sla-first" {...register("firstResponseMinutes")} />
              <FieldError errors={[errors.firstResponseMinutes]} />
            </FieldContent>
          </Field>
          <Field data-invalid={!!errors.resolutionMinutes}>
            <FieldLabel htmlFor="ticket-sla-resolution">解决时长(分钟)</FieldLabel>
            <FieldContent>
              <Input id="ticket-sla-resolution" {...register("resolutionMinutes")} />
              <FieldError errors={[errors.resolutionMinutes]} />
            </FieldContent>
          </Field>
        </div>
        <Field data-invalid={!!errors.status}>
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
                  options={listStatusOptions
                    .filter((item) => item.value !== "all")
                    .map((item) => ({ value: item.value, label: item.label }))}
                />
              )}
            />
            <FieldError errors={[errors.status]} />
          </FieldContent>
        </Field>
        <Field>
          <FieldLabel htmlFor="ticket-sla-remark">备注</FieldLabel>
          <FieldContent>
            <Textarea id="ticket-sla-remark" rows={4} {...register("remark")} />
          </FieldContent>
        </Field>
      </form>
    </ProjectDialog>
  )
}
