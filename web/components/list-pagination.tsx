"use client"

import { ChevronLeftIcon, ChevronRightIcon } from "lucide-react"

import { Button } from "@/components/ui/button"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"

type ListPaginationProps = {
  page: number
  total: number
  limit: number
  loading?: boolean
  pageSizeOptions?: number[]
  onPageChange: (page: number) => void
  onLimitChange: (limit: number) => void
}

export function ListPagination({
  page,
  total,
  limit,
  loading = false,
  pageSizeOptions = [10, 20, 50, 100],
  onPageChange,
  onLimitChange,
}: ListPaginationProps) {
  const totalPages = Math.max(1, Math.ceil(total / limit))
  const canGoPreviousPage = page > 1
  const canGoNextPage = page < totalPages

  function handleLimitChange(value: string | null) {
    if (!value) {
      return
    }
    const nextLimit = Number(value)
    if (!Number.isInteger(nextLimit) || nextLimit <= 0 || nextLimit === limit) {
      return
    }
    onLimitChange(nextLimit)
  }

  return (
    <div className="flex flex-col gap-3 text-sm text-muted-foreground sm:flex-row sm:items-center sm:justify-between">
      <div className="flex items-center gap-3">
        <span>
          第 {page} 页 / 共 {totalPages} 页
        </span>
        <span>共 {total} 条记录</span>
      </div>
      <div className="flex flex-wrap items-center gap-2">
        <Select value={String(limit)} onValueChange={handleLimitChange}>
          <SelectTrigger className="w-20">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {pageSizeOptions.map((pageSize) => (
              <SelectItem key={pageSize} value={String(pageSize)}>
                每页 {pageSize} 条
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        <Button
          variant="outline"
          onClick={() => onPageChange(page - 1)}
          disabled={loading || !canGoPreviousPage}
        >
          <ChevronLeftIcon />
          上一页
        </Button>
        <Button
          variant="outline"
          onClick={() => onPageChange(page + 1)}
          disabled={loading || !canGoNextPage}
        >
          下一页
          <ChevronRightIcon />
        </Button>
      </div>
    </div>
  )
}
