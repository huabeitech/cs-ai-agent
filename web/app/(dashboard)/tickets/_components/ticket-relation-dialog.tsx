"use client"

import { useEffect, useState } from "react"
import { zodResolver } from "@hookform/resolvers/zod"
import { Resolver, useForm } from "react-hook-form"
import { toast } from "sonner"
import { z } from "zod/v4"
import { SearchIcon } from "lucide-react"

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
import { Input } from "@/components/ui/input"
import { addTicketRelation, fetchTickets, type TicketItem } from "@/lib/api/ticket"
import { cn } from "@/lib/utils"

const relationOptions = [
  { value: "duplicate", label: "重复工单" },
  { value: "related", label: "相关工单" },
  { value: "parent", label: "父工单" },
  { value: "child", label: "子工单" },
]

const schema = z.object({
  relationType: z.string().trim().min(1, "请选择关联类型"),
  relatedTicketId: z.number().int().positive("请选择关联工单"),
})

type FormValues = z.infer<typeof schema>

const resolver = zodResolver(schema as never) as Resolver<
  z.input<typeof schema>,
  undefined,
  z.output<typeof schema>
>

type TicketRelationDialogProps = {
  open: boolean
  ticketId: number | null
  onOpenChange: (open: boolean) => void
  onSuccess?: () => Promise<void> | void
}

export function TicketRelationDialog({
  open,
  ticketId,
  onOpenChange,
  onSuccess,
}: TicketRelationDialogProps) {
  const [keyword, setKeyword] = useState("")
  const [searching, setSearching] = useState(false)
  const [searchResults, setSearchResults] = useState<TicketItem[]>([])

  const form = useForm<
    z.input<typeof schema>,
    undefined,
    z.output<typeof schema>
  >({
    resolver,
    defaultValues: {
      relationType: "related",
      relatedTicketId: 0,
    },
  })

  const {
    handleSubmit,
    setValue,
    reset,
    watch,
    formState: { errors, isSubmitting },
  } = form

  useEffect(() => {
    if (open) {
      reset({ relationType: "related", relatedTicketId: 0 })
      setKeyword("")
      setSearchResults([])
    }
  }, [open, reset])

  useEffect(() => {
    if (!open) {
      return
    }
    const trimmedKeyword = keyword.trim()
    if (trimmedKeyword.length < 2) {
      setSearchResults([])
      return
    }
    const timer = window.setTimeout(async () => {
      setSearching(true)
      try {
        const data = await fetchTickets({
          keyword: trimmedKeyword,
          page: 1,
          limit: 8,
        })
        const results = Array.isArray(data.results) ? data.results : []
        setSearchResults(results.filter((item) => item.id !== ticketId))
      } catch (error) {
        toast.error(error instanceof Error ? error.message : "搜索工单失败")
      } finally {
        setSearching(false)
      }
    }, 250)
    return () => window.clearTimeout(timer)
  }, [keyword, open, ticketId])

  const selectedTicketId = watch("relatedTicketId")
  const selectedTicket =
    searchResults.find((item) => item.id === selectedTicketId) ?? null

  async function onFormSubmit(values: FormValues) {
    if (!ticketId) {
      toast.error("工单不存在")
      return
    }
    try {
      await addTicketRelation({
        ticketId,
        relationType: values.relationType,
        relatedTicketId: values.relatedTicketId,
      })
      toast.success("关联工单已添加")
      onOpenChange(false)
      await onSuccess?.()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "添加关联工单失败")
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg gap-0 p-0 sm:max-w-lg">
        <DialogHeader className="px-6 pt-6">
          <DialogTitle>新增关联工单</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit(onFormSubmit)}>
          <div className="space-y-4 p-6">
            <Field data-invalid={!!errors.relationType}>
              <FieldLabel>关联类型</FieldLabel>
              <FieldContent>
                <OptionCombobox
                  value={watch("relationType")}
                  options={relationOptions}
                  placeholder="请选择关联类型"
                  onChange={(value) => setValue("relationType", value, { shouldValidate: true })}
                />
                <FieldError errors={[errors.relationType]} />
              </FieldContent>
            </Field>
            <Field data-invalid={!!errors.relatedTicketId}>
              <FieldLabel>搜索并选择工单</FieldLabel>
              <FieldContent>
                <div className="space-y-3">
                  <div className="relative">
                    <SearchIcon className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
                    <Input
                      className="pl-9"
                      value={keyword}
                      placeholder="输入工单号或标题，至少 2 个字"
                      onChange={(event) => {
                        setKeyword(event.target.value)
                        setValue("relatedTicketId", 0, { shouldValidate: true })
                      }}
                    />
                  </div>
                  <div className="max-h-64 overflow-y-auto rounded-lg border">
                    {searching ? (
                      <div className="p-3 text-sm text-muted-foreground">搜索中...</div>
                    ) : searchResults.length > 0 ? (
                      searchResults.map((item) => {
                        const active = item.id === selectedTicketId
                        return (
                          <button
                            key={item.id}
                            type="button"
                            className={cn(
                              "flex w-full flex-col items-start gap-1 border-b px-3 py-3 text-left last:border-b-0",
                              active ? "bg-accent text-accent-foreground" : "hover:bg-muted/40",
                            )}
                            onClick={() => setValue("relatedTicketId", item.id, { shouldValidate: true })}
                          >
                            <div className="text-xs text-muted-foreground">{item.ticketNo}</div>
                            <div className="line-clamp-1 text-sm font-medium">{item.title}</div>
                            <div className="flex flex-wrap gap-2 text-xs text-muted-foreground">
                              <span>状态：{item.status}</span>
                              <span>处理人：{item.currentAssigneeName || "未指派"}</span>
                            </div>
                          </button>
                        )
                      })
                    ) : (
                      <div className="p-3 text-sm text-muted-foreground">
                        {keyword.trim().length < 2 ? "输入至少 2 个字开始搜索" : "未找到匹配工单"}
                      </div>
                    )}
                  </div>
                  {selectedTicket ? (
                    <div className="rounded-lg border bg-muted/20 p-3 text-sm">
                      已选中：{selectedTicket.ticketNo} / {selectedTicket.title}
                    </div>
                  ) : null}
                </div>
                <FieldError errors={[errors.relatedTicketId]} />
              </FieldContent>
            </Field>
          </div>
          <DialogFooter className="mx-0 mb-0 px-6 py-4">
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              取消
            </Button>
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting ? "提交中..." : "确认添加"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
