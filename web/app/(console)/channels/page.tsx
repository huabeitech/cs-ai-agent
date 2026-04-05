"use client"

import { useCallback, useEffect, useState } from "react"
import {
  Building2Icon,
  MessageSquareMoreIcon,
  MoreHorizontalIcon,
  PlusIcon,
  RefreshCwIcon,
  SearchIcon,
  Trash2Icon,
} from "lucide-react"
import { toast } from "sonner"

import {
  createChannel,
  deleteChannel,
  fetchChannels,
  updateChannel,
  updateChannelStatus,
  type AdminChannel,
  type CreateAdminChannelPayload,
  type PageResult,
} from "@/lib/api/admin"
import { OptionCombobox } from "@/components/option-combobox"
import { EditDialog } from "./_components/edit"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Input } from "@/components/ui/input"
import { ListPagination } from "@/components/list-pagination"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { Status, StatusLabels } from "@/lib/generated/enums"
import { getEnumLabel, getEnumOptions } from "@/lib/enums"
import { ButtonGroup } from "@/components/ui/button-group"
import { Switch } from "@/components/ui/switch"

const statusOptions = [
  { value: "all", label: "全部状态" },
  ...getEnumOptions(StatusLabels).map((option) => ({
    value: String(option.value),
    label: option.label,
  })),
] as const

const channelTypeOptions = [
  { value: "all", label: "全部类型" },
  { value: "web", label: "Web 站点" },
  { value: "wxwork_kf", label: "企业微信客服" },
] as const

function getChannelTypeLabel(channelType: string) {
  if (channelType === "wxwork_kf") {
    return "企业微信客服"
  }
  return "Web 站点"
}

function ChannelIcon({ channelType }: { channelType: string }) {
  if (channelType === "wxwork_kf") {
    return <MessageSquareMoreIcon className="size-4" />
  }
  return <Building2Icon className="size-4" />
}

export default function DashboardChannelsPage() {
  const [nameInput, setNameInput] = useState("")
  const [appIdInput, setAppIdInput] = useState("")
  const [channelTypeInput, setChannelTypeInput] = useState("all")
  const [statusInput, setStatusInput] = useState("all")
  const [name, setName] = useState("")
  const [appId, setAppId] = useState("")
  const [channelType, setChannelType] = useState("all")
  const [status, setStatus] = useState("all")
  const [page, setPage] = useState(1)
  const [limit, setLimit] = useState(20)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [actionLoadingId, setActionLoadingId] = useState<number | null>(null)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editingItem, setEditingItem] = useState<AdminChannel | null>(null)
  const [result, setResult] = useState<PageResult<AdminChannel>>({
    results: [],
    page: { page: 1, limit: 20, total: 0 },
  })

  const loadData = useCallback(async () => {
    setLoading(true)
    try {
      const data = await fetchChannels({
        name: name.trim() || undefined,
        appId: appId.trim() || undefined,
        channelType: channelType === "all" ? undefined : channelType,
        status: status === "all" ? undefined : status,
        page,
        limit,
      })
      setResult(data)
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "加载接入渠道失败")
    } finally {
      setLoading(false)
    }
  }, [appId, channelType, limit, name, page, status])

  useEffect(() => {
    void loadData()
  }, [loadData])

  function applyFilters() {
    setName(nameInput)
    setAppId(appIdInput)
    setChannelType(channelTypeInput)
    setStatus(statusInput)
    setPage(1)
  }

  function handleFilterKeyDown(event: React.KeyboardEvent<HTMLInputElement>) {
    if (event.key !== "Enter") {
      return
    }
    event.preventDefault()
    applyFilters()
  }

  function openCreateDialog() {
    setEditingItem(null)
    setDialogOpen(true)
  }

  function openEditDialog(item: AdminChannel) {
    setEditingItem(item)
    setDialogOpen(true)
  }

  async function handleSubmit(payload: CreateAdminChannelPayload) {
    if (saving) {
      return
    }
    setSaving(true)
    try {
      if (editingItem) {
        await updateChannel({ id: editingItem.id, ...payload })
        toast.success(`已更新接入渠道：${payload.name}`)
      } else {
        const created = await createChannel(payload)
        toast.success(`已创建接入渠道：${created.name}`)
      }
      setDialogOpen(false)
      setEditingItem(null)
      await loadData()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "保存接入渠道失败")
    } finally {
      setSaving(false)
    }
  }

  async function handleToggleStatus(item: AdminChannel) {
    setActionLoadingId(item.id)
    try {
      const nextStatus = item.status === Status.Ok ? Status.Disabled : Status.Ok
      await updateChannelStatus(item.id, nextStatus)
      toast.success(`已${nextStatus === Status.Ok ? "启用" : "禁用"}：${item.name}`)
      await loadData()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "更新渠道状态失败")
    } finally {
      setActionLoadingId(null)
    }
  }

  async function handleDelete(item: AdminChannel) {
    setActionLoadingId(item.id)
    try {
      await deleteChannel(item.id)
      toast.success(`已删除接入渠道：${item.name}`)
      await loadData()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "删除接入渠道失败")
    } finally {
      setActionLoadingId(null)
    }
  }

  return (
    <>
      <div className="flex flex-1 flex-col gap-6 p-4 lg:p-6">
        <div className="flex flex-col gap-2 xl:flex-row xl:flex-wrap xl:items-center xl:justify-end">
          <Input
            value={nameInput}
            onChange={(event) => setNameInput(event.target.value)}
            onKeyDown={handleFilterKeyDown}
            placeholder="按渠道名称筛选"
            className="w-full xl:w-56"
          />
          <div className="relative min-w-72">
            <SearchIcon className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              value={appIdInput}
              onChange={(event) => setAppIdInput(event.target.value)}
              onKeyDown={handleFilterKeyDown}
              placeholder="按 appId 筛选"
              className="pl-9"
            />
          </div>
          <div className="w-full xl:w-40">
            <OptionCombobox
              value={channelTypeInput}
              options={[...channelTypeOptions]}
              placeholder="全部类型"
              searchPlaceholder="搜索渠道类型"
              emptyText="未找到渠道类型"
              onChange={setChannelTypeInput}
            />
          </div>
          <div className="w-full xl:w-36">
            <OptionCombobox
              value={statusInput}
              options={[...statusOptions]}
              placeholder="全部状态"
              searchPlaceholder="搜索状态"
              emptyText="未找到状态"
              onChange={setStatusInput}
            />
          </div>
          <Button variant="outline" onClick={applyFilters} disabled={loading}>
            <SearchIcon />
            查询
          </Button>
          <Button variant="outline" onClick={() => void loadData()} disabled={loading}>
            <RefreshCwIcon className={loading ? "animate-spin" : ""} />
            刷新列表
          </Button>
          <Button onClick={openCreateDialog}>
            <PlusIcon />
            新建渠道
          </Button>
        </div>

        <div className="rounded-xl border bg-card">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>渠道</TableHead>
                <TableHead>类型</TableHead>
                <TableHead>入口标识</TableHead>
                <TableHead>接待 Agent</TableHead>
                <TableHead>状态</TableHead>
                <TableHead className="w-[88px] text-right">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {result.results.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={6} className="py-12 text-center text-muted-foreground">
                    {loading ? "正在加载接入渠道..." : "暂无接入渠道"}
                  </TableCell>
                </TableRow>
              ) : null}
              {result.results.map((item) => (
                <TableRow key={item.id}>
                  <TableCell>
                    <div className="flex items-center gap-3">
                      <div className="flex size-10 items-center justify-center rounded-2xl bg-muted">
                        <ChannelIcon channelType={item.channelType} />
                      </div>
                      <div>
                        <div className="font-medium">{item.name}</div>
                        <div className="text-xs text-muted-foreground">{getChannelTypeLabel(item.channelType)}</div>
                      </div>
                    </div>
                  </TableCell>
                  <TableCell>
                    <Badge variant="outline">{getChannelTypeLabel(item.channelType)}</Badge>
                  </TableCell>
                  <TableCell className="font-mono text-xs">{item.appId || item.configJson || "-"}</TableCell>
                  <TableCell>{item.aiAgentName || "-"}</TableCell>
                  <TableCell>
                    <div className="flex items-center gap-3">
                      <Switch
                        checked={item.status === Status.Ok}
                        disabled={actionLoadingId === item.id}
                        onCheckedChange={() => void handleToggleStatus(item)}
                        aria-label={`${item.name} 状态切换`}
                      />
                      <Badge variant={item.status === Status.Ok ? "default" : "outline"}>
                        {getEnumLabel(StatusLabels, item.status as Status)}
                      </Badge>
                    </div>
                  </TableCell>
                  <TableCell className="text-right">
                    <ButtonGroup className="ml-auto">
                      <Button variant="outline" size="sm" onClick={() => openEditDialog(item)}>
                        编辑
                      </Button>
                      <DropdownMenu>
                        <DropdownMenuTrigger
                          render={<Button variant="outline" size="icon-sm" className="ml-auto" />}
                          aria-label={`更多操作 ${item.name}`}
                        >
                          <MoreHorizontalIcon />
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem
                            className="text-destructive"
                            disabled={actionLoadingId === item.id}
                            onClick={() => void handleDelete(item)}
                          >
                            <Trash2Icon className="size-4" />
                            删除
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </ButtonGroup>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
          <div className="border-t px-4 py-3">
            <ListPagination
              page={result.page.page}
              limit={result.page.limit}
              total={result.page.total}
              onPageChange={(nextPage) => setPage(nextPage)}
              onLimitChange={(nextLimit) => {
                setLimit(nextLimit)
                setPage(1)
              }}
            />
          </div>
        </div>
      </div>

      <EditDialog
        open={dialogOpen}
        saving={saving}
        itemId={editingItem?.id ?? null}
        onOpenChange={setDialogOpen}
        onSubmit={handleSubmit}
      />
    </>
  )
}
