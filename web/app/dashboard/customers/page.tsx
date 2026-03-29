"use client"

import { useCallback, useEffect, useMemo, useState } from "react"
import {
  Link2Icon,
  MoreHorizontalIcon,
  PlusIcon,
  RefreshCwIcon,
  SearchIcon,
  Trash2Icon,
} from "lucide-react"
import { toast } from "sonner"

import {
  deleteCustomer,
  fetchCustomers,
  saveCustomerProfile,
  updateCustomerStatus,
  type AdminCustomer,
} from "@/lib/api/customer"
import { fetchCompanies, type AdminCompany } from "@/lib/api/company"
import { type PageResult } from "@/lib/api/admin"
import { type CustomerFormSavePayload } from "@/components/customer-form"
import { CustomerLinkOrCreateDialog } from "@/components/customer-link-or-create-dialog"
import { EditDialog } from "./_components/edit"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { ButtonGroup } from "@/components/ui/button-group"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Input } from "@/components/ui/input"
import { ListPagination } from "@/components/list-pagination"
import { OptionCombobox, type ComboboxOption } from "@/components/option-combobox"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { getEnumLabel, getEnumOptions } from "@/lib/enums"
import { Gender, GenderLabels, Status, StatusLabels } from "@/lib/generated/enums"

const listStatusOptions = [
  { value: "all", label: "全部状态" },
  ...getEnumOptions(StatusLabels)
    .filter((item) => Number(item.value) !== Status.Deleted)
    .map((item) => ({
      value: String(item.value),
      label: item.label,
    })),
] as const

const genderOptions = [
  { value: "all", label: "全部性别" },
  ...getEnumOptions(GenderLabels).map((item) => ({
    value: String(item.value),
    label: item.label,
  })),
] as const

function getLabel(
  value: string,
  options: ReadonlyArray<{ value: string; label: string }>
) {
  return options.find((item) => item.value === value)?.label ?? "请选择"
}

export default function DashboardCustomersPage() {
  const [nameInput, setNameInput] = useState("")
  const [mobileInput, setMobileInput] = useState("")
  const [emailInput, setEmailInput] = useState("")
  const [statusFilterInput, setStatusFilterInput] = useState("all")
  const [genderFilterInput, setGenderFilterInput] = useState("all")
  const [companyFilterInput, setCompanyFilterInput] = useState("0")

  const [name, setName] = useState("")
  const [mobile, setMobile] = useState("")
  const [email, setEmail] = useState("")
  const [statusFilter, setStatusFilter] = useState("all")
  const [genderFilter, setGenderFilter] = useState("all")
  const [companyFilter, setCompanyFilter] = useState("0")

  const [companyOptions, setCompanyOptions] = useState<ComboboxOption[]>([
    { value: "0", label: "全部公司" },
  ])
  const [companyNameMap, setCompanyNameMap] = useState<Record<number, string>>({})

  const [page, setPage] = useState(1)
  const [limit, setLimit] = useState(20)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [actionLoadingId, setActionLoadingId] = useState<number | null>(null)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [linkOrCreateOpen, setLinkOrCreateOpen] = useState(false)
  const [editingItem, setEditingItem] = useState<AdminCustomer | null>(null)
  const [result, setResult] = useState<PageResult<AdminCustomer>>({
    results: [],
    page: { page: 1, limit: 20, total: 0 },
  })

  useEffect(() => {
    async function loadCompanies() {
      try {
        const data = await fetchCompanies({ status: 0, page: 1, limit: 500 })
        const opts: ComboboxOption[] = [
          { value: "0", label: "全部公司" },
          ...data.results.map((item) => ({
            value: String(item.id),
            label: item.name,
          })),
        ]
        setCompanyOptions(opts)
        const map: Record<number, string> = {}
        data.results.forEach((item: AdminCompany) => {
          map[item.id] = item.name
        })
        setCompanyNameMap(map)
      } catch {
        // ignore
      }
    }
    void loadCompanies()
  }, [])

  const loadData = useCallback(async () => {
    setLoading(true)
    try {
      const data = await fetchCustomers({
        name: name.trim() || undefined,
        primaryMobile: mobile.trim() || undefined,
        primaryEmail: email.trim() || undefined,
        status:
          statusFilter === "all" ? undefined : Number(statusFilter),
        gender:
          genderFilter === "all" ? undefined : Number(genderFilter),
        companyId:
          companyFilter === "0" ? undefined : Number(companyFilter),
        page,
        limit,
      })
      setResult(data)
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "加载客户列表失败")
    } finally {
      setLoading(false)
    }
  }, [companyFilter, email, genderFilter, limit, mobile, name, page, statusFilter])

  useEffect(() => {
    void loadData()
  }, [loadData])

  const companyFilterLabel = useMemo(() => {
    return (
      companyOptions.find((item) => item.value === companyFilterInput)?.label ??
      "全部公司"
    )
  }, [companyFilterInput, companyOptions])

  function applyFilters() {
    setName(nameInput)
    setMobile(mobileInput)
    setEmail(emailInput)
    setStatusFilter(statusFilterInput)
    setGenderFilter(genderFilterInput)
    setCompanyFilter(companyFilterInput)
    setPage(1)
  }

  function handleFilterKeyDown(event: React.KeyboardEvent<HTMLInputElement>) {
    if (event.key !== "Enter") return
    event.preventDefault()
    applyFilters()
  }

  function handlePageChange(nextPage: number) {
    if (nextPage < 1 || nextPage === page) return
    setPage(nextPage)
  }

  function openCreateDialog() {
    setEditingItem(null)
    setDialogOpen(true)
  }

  function openEditDialog(item: AdminCustomer) {
    setEditingItem(item)
    setDialogOpen(true)
  }

  function handleDialogOpenChange(open: boolean) {
    if (saving) return
    if (!open) setEditingItem(null)
    setDialogOpen(open)
  }

  async function handleSave(payload: CustomerFormSavePayload) {
    if (saving) return
    setSaving(true)
    try {
      await saveCustomerProfile(payload)
      toast.success(
        editingItem ? `已更新客户：${editingItem.name}` : `已创建客户：${payload.name}`
      )
      setDialogOpen(false)
      setEditingItem(null)
      await loadData()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "保存客户失败")
    } finally {
      setSaving(false)
    }
  }

  async function handleToggleStatus(item: AdminCustomer) {
    setActionLoadingId(item.id)
    try {
      const nextStatus = item.status === 0 ? 1 : 0
      await updateCustomerStatus(item.id, nextStatus)
      toast.success(`已${nextStatus === 0 ? "启用" : "禁用"}：${item.name}`)
      await loadData()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "更新状态失败")
    } finally {
      setActionLoadingId(null)
    }
  }

  async function handleDelete(item: AdminCustomer) {
    setActionLoadingId(item.id)
    try {
      await deleteCustomer(item.id)
      toast.success(`已删除客户：${item.name}`)
      await loadData()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "删除客户失败")
    } finally {
      setActionLoadingId(null)
    }
  }

  function getGenderText(gender: number) {
    return getEnumLabel(GenderLabels, gender as Gender)
  }

  return (
    <>
      <div className="flex flex-1 flex-col gap-6 p-4 lg:p-6">
        <div className="flex flex-col gap-2 xl:flex-row xl:items-center xl:justify-end">
          <div className="relative min-w-72">
            <SearchIcon className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              value={nameInput}
              onChange={(event) => setNameInput(event.target.value)}
              onKeyDown={handleFilterKeyDown}
              placeholder="按客户名称筛选"
              className="pl-9"
            />
          </div>
          <Input
            value={mobileInput}
            onChange={(event) => setMobileInput(event.target.value)}
            onKeyDown={handleFilterKeyDown}
            placeholder="按手机号筛选"
            className="w-full xl:w-40"
          />
          <Input
            value={emailInput}
            onChange={(event) => setEmailInput(event.target.value)}
            onKeyDown={handleFilterKeyDown}
            placeholder="按邮箱筛选"
            className="w-full xl:w-56"
          />

          <Select value={genderFilterInput} onValueChange={(v) => setGenderFilterInput(v ?? "all")}>
            <SelectTrigger className="w-full xl:w-28">
              <SelectValue>{getLabel(genderFilterInput, genderOptions)}</SelectValue>
            </SelectTrigger>
            <SelectContent>
              {genderOptions.map((item) => (
                <SelectItem key={item.value} value={item.value}>
                  {item.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>

          <div className="w-full xl:w-56">
            <OptionCombobox
              value={companyFilterInput}
              options={companyOptions}
              placeholder={companyFilterLabel}
              searchPlaceholder="搜索公司名称"
              onChange={(v) => setCompanyFilterInput(v)}
            />
          </div>

          <Select value={statusFilterInput} onValueChange={(v) => setStatusFilterInput(v ?? "all")}>
            <SelectTrigger className="w-full xl:w-28">
              <SelectValue>{getLabel(statusFilterInput, listStatusOptions)}</SelectValue>
            </SelectTrigger>
            <SelectContent>
              {listStatusOptions.map((item) => (
                <SelectItem key={item.value} value={item.value}>
                  {item.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>

          <Button variant="outline" onClick={applyFilters} disabled={loading}>
            <SearchIcon />
            查询
          </Button>
          <Button variant="outline" onClick={() => void loadData()} disabled={loading}>
            <RefreshCwIcon className={loading ? "animate-spin" : ""} />
            刷新列表
          </Button>
          <Button variant="outline" onClick={() => setLinkOrCreateOpen(true)}>
            <Link2Icon />
            搜索或新建
          </Button>
          <Button onClick={openCreateDialog}>
            <PlusIcon />
            新建
          </Button>
        </div>

        <div className="overflow-hidden rounded-lg border bg-card">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-20">ID</TableHead>
                <TableHead>客户名称</TableHead>
                <TableHead className="w-20">性别</TableHead>
                <TableHead>所属公司</TableHead>
                <TableHead>手机号</TableHead>
                <TableHead>邮箱</TableHead>
                <TableHead className="w-24">状态</TableHead>
                <TableHead className="w-40">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {result.results.length === 0 && !loading ? (
                <TableRow>
                  <TableCell colSpan={8} className="py-10 text-center text-muted-foreground">
                    暂无客户数据
                  </TableCell>
                </TableRow>
              ) : (
                result.results.map((item) => {
                  const actionLoading = actionLoadingId === item.id
                  return (
                    <TableRow key={item.id}>
                      <TableCell>{item.id}</TableCell>
                      <TableCell className="font-medium">{item.name}</TableCell>
                      <TableCell className="text-muted-foreground">
                        {getGenderText(item.gender)}
                      </TableCell>
                      <TableCell className="text-muted-foreground">
                        {item.companyId > 0 ? companyNameMap[item.companyId] ?? String(item.companyId) : "-"}
                      </TableCell>
                      <TableCell className="text-muted-foreground">
                        {item.primaryMobile || "-"}
                      </TableCell>
                      <TableCell className="text-muted-foreground">
                        {item.primaryEmail || "-"}
                      </TableCell>
                      <TableCell>
                        <Badge variant={item.status === 0 ? "default" : "secondary"}>
                          {item.status === 0 ? "启用" : "禁用"}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        <ButtonGroup className="w-full justify-end">
                          <Button variant="outline" size="sm" onClick={() => openEditDialog(item)}>
                            编辑
                          </Button>
                          <DropdownMenu>
                            <DropdownMenuTrigger
                              render={
                                <Button variant="outline" size="sm" disabled={actionLoading} />
                              }
                              aria-label={`更多操作 ${item.name}`}
                            >
                              <MoreHorizontalIcon />
                            </DropdownMenuTrigger>
                            <DropdownMenuContent align="end" className="w-40 min-w-40">
                              <DropdownMenuItem onClick={() => void handleToggleStatus(item)}>
                                {item.status === 0 ? "禁用" : "启用"}
                              </DropdownMenuItem>
                              <DropdownMenuItem
                                className="text-destructive focus:text-destructive"
                                onClick={() => void handleDelete(item)}
                              >
                                <Trash2Icon className="mr-2 size-4" />
                                删除
                              </DropdownMenuItem>
                            </DropdownMenuContent>
                          </DropdownMenu>
                        </ButtonGroup>
                      </TableCell>
                    </TableRow>
                  )
                })
              )}
            </TableBody>
          </Table>
        </div>

        <ListPagination
          page={result.page.page}
          total={result.page.total}
          limit={result.page.limit}
          loading={loading}
          onPageChange={handlePageChange}
          onLimitChange={(nextLimit) => {
            setLimit(nextLimit)
            setPage(1)
          }}
        />
      </div>

      <CustomerLinkOrCreateDialog
        open={linkOrCreateOpen}
        onOpenChange={setLinkOrCreateOpen}
        onSuccess={() => void loadData()}
      />

      <EditDialog
        open={dialogOpen}
        saving={saving}
        itemId={editingItem?.id ?? null}
        onOpenChange={handleDialogOpenChange}
        onSave={handleSave}
      />
    </>
  )
}

