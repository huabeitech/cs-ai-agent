"use client"

import { useEffect, useMemo, useState } from "react"
import { ChevronsUpDownIcon, PlusIcon } from "lucide-react"
import { toast } from "sonner"

import { EditDialog as CompanyEditDialog } from "@/app/(dashboard)/companies/_components/edit"
import { Button } from "@/components/ui/button"
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
} from "@/components/ui/command"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import {
  createCompany,
  fetchCompanies,
  fetchCompany,
  type AdminCompany,
  type CreateAdminCompanyPayload,
} from "@/lib/api/company"
import { cn } from "@/lib/utils"

type CompanyPickerProps = {
  value: string
  onChange: (value: string) => void
  disabled?: boolean
  placeholder?: string
}

export function CompanyPicker({
  value,
  onChange,
  disabled = false,
  placeholder = "请选择公司",
}: CompanyPickerProps) {
  const [open, setOpen] = useState(false)
  const [keyword, setKeyword] = useState("")
  const [loading, setLoading] = useState(false)
  const [options, setOptions] = useState<AdminCompany[]>([])
  const [selectedCompany, setSelectedCompany] = useState<AdminCompany | null>(null)
  const [createOpen, setCreateOpen] = useState(false)
  const [createSaving, setCreateSaving] = useState(false)

  const trimmedKeyword = keyword.trim()
  const normalizedKeyword = trimmedKeyword.toLowerCase()

  useEffect(() => {
    let cancelled = false
    if (!open) {
      return
    }
    setLoading(true)
    void (async () => {
      try {
        const data = await fetchCompanies({
          status: 0,
          page: 1,
          limit: 20,
          name: trimmedKeyword || undefined,
        })
        if (cancelled) {
          return
        }
        setOptions(data.results)
      } catch (error) {
        if (!cancelled) {
          setOptions([])
          toast.error(error instanceof Error ? error.message : "加载公司列表失败")
        }
      } finally {
        if (!cancelled) {
          setLoading(false)
        }
      }
    })()

    return () => {
      cancelled = true
    }
  }, [open, trimmedKeyword])

  useEffect(() => {
    let cancelled = false
    const companyId = Number(value)
    if (companyId <= 0) {
      setSelectedCompany(null)
      return
    }
    if (selectedCompany?.id === companyId) {
      return
    }
    void (async () => {
      try {
        const data = await fetchCompany(companyId)
        if (!cancelled) {
          setSelectedCompany(data)
        }
      } catch {
        if (!cancelled) {
          setSelectedCompany(null)
        }
      }
    })()
    return () => {
      cancelled = true
    }
  }, [selectedCompany?.id, value])

  const canCreate = useMemo(() => {
    if (!trimmedKeyword) {
      return false
    }
    return !options.some((item) => item.name.trim().toLowerCase() === normalizedKeyword)
  }, [normalizedKeyword, options, trimmedKeyword])

  const buttonLabel =
    Number(value) > 0 ? selectedCompany?.name || `公司 #${value}` : placeholder

  function handleSelectCompany(company: AdminCompany) {
    setSelectedCompany(company)
    onChange(String(company.id))
    setOpen(false)
    setKeyword("")
  }

  function handleClear() {
    setSelectedCompany(null)
    onChange("0")
    setOpen(false)
    setKeyword("")
  }

  async function handleCreateCompany(payload: CreateAdminCompanyPayload) {
    setCreateSaving(true)
    try {
      const created = await createCompany(payload)
      setSelectedCompany(created)
      onChange(String(created.id))
      setCreateOpen(false)
      setOpen(false)
      setKeyword("")
      toast.success(`已创建公司：${created.name}`)
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "创建公司失败")
      throw error
    } finally {
      setCreateSaving(false)
    }
  }

  return (
    <>
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger
          render={
            <Button
              variant="outline"
              role="combobox"
              className="w-full justify-between font-normal"
              disabled={disabled}
            />
          }
        >
          <span className={cn("truncate", Number(value) > 0 ? "text-foreground" : "text-muted-foreground")}>
            {buttonLabel}
          </span>
          <ChevronsUpDownIcon className="ml-2 size-4 shrink-0 opacity-50" />
        </PopoverTrigger>
        <PopoverContent className="w-(--radix-popover-trigger-width) p-0" align="start">
          <Command shouldFilter={false}>
            <CommandInput
              value={keyword}
              onValueChange={setKeyword}
              placeholder="搜索公司名称"
            />
            <CommandList>
              {loading ? <CommandEmpty>加载中...</CommandEmpty> : null}
              {!loading && options.length === 0 ? <CommandEmpty>未找到匹配公司</CommandEmpty> : null}
              {!loading ? (
                <CommandGroup heading="搜索结果">
                  <CommandItem
                    value="none"
                    data-checked={Number(value) <= 0}
                    onSelect={handleClear}
                  >
                    <span>不关联公司</span>
                  </CommandItem>
                  {options.map((item) => (
                    <CommandItem
                      key={item.id}
                      value={`${item.name} ${item.code}`}
                      data-checked={item.id === Number(value)}
                      onSelect={() => handleSelectCompany(item)}
                    >
                      <div className="flex min-w-0 flex-col">
                        <span className="truncate">{item.name}</span>
                        {item.code ? (
                          <span className="truncate text-xs text-muted-foreground">{item.code}</span>
                        ) : null}
                      </div>
                    </CommandItem>
                  ))}
                </CommandGroup>
              ) : null}
              {canCreate ? (
                <>
                  <CommandSeparator />
                  <CommandGroup heading="操作">
                    <CommandItem
                      value={`create ${trimmedKeyword}`}
                      onSelect={() => setCreateOpen(true)}
                    >
                      <PlusIcon className="size-4" />
                      <span className="truncate">新建公司“{trimmedKeyword}”</span>
                    </CommandItem>
                  </CommandGroup>
                </>
              ) : null}
            </CommandList>
          </Command>
        </PopoverContent>
      </Popover>

      <CompanyEditDialog
        open={createOpen}
        saving={createSaving}
        itemId={null}
        initialValues={{ name: trimmedKeyword }}
        onOpenChange={setCreateOpen}
        onSubmit={handleCreateCompany}
      />
    </>
  )
}
