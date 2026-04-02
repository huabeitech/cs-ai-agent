"use client"

import Link from "next/link"
import { PlusIcon, RefreshCcwIcon, SearchIcon } from "lucide-react"
import { useCallback, useEffect, useState } from "react"
import { toast } from "sonner"

import { ListPagination } from "@/components/list-pagination"
import { OptionCombobox } from "@/components/option-combobox"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import {
  fetchAgentProfilesAll,
  fetchAgentTeamsAll,
  type AdminAgentProfile,
  type AdminAgentTeam,
} from "@/lib/api/admin"
import {
  createTicket,
  fetchTickets,
  updateTicket,
  type Paging,
  type TicketItem,
} from "@/lib/api/ticket"
import { formatDateTime } from "@/lib/utils"
import { TicketEditDialog } from "./_components/ticket-edit-dialog"
import { TicketPriorityBadge } from "./_components/ticket-priority-badge"
import { TicketStatusBadge } from "./_components/ticket-status-badge"

const emptyPaging: Paging = { page: 1, limit: 20, total: 0 }

export default function TicketsPage() {
  const [items, setItems] = useState<TicketItem[]>([])
  const [paging, setPaging] = useState<Paging>(emptyPaging)
  const [loading, setLoading] = useState(false)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editingItem, setEditingItem] = useState<TicketItem | null>(null)
  const [keywordInput, setKeywordInput] = useState("")
  const [keyword, setKeyword] = useState("")
  const [statusFilter, setStatusFilter] = useState("all")
  const [priorityFilter, setPriorityFilter] = useState("all")
  const [teamFilter, setTeamFilter] = useState("all")
  const [assigneeFilter, setAssigneeFilter] = useState("all")
  const [teams, setTeams] = useState<AdminAgentTeam[]>([])
  const [agents, setAgents] = useState<AdminAgentProfile[]>([])

  const loadTickets = useCallback(async () => {
    setLoading(true)
    try {
      const data = await fetchTickets({
        page: paging.page,
        limit: paging.limit,
        keyword: keyword.trim() || undefined,
        status: statusFilter === "all" ? undefined : statusFilter,
        priority: priorityFilter === "all" ? undefined : Number(priorityFilter),
        currentTeamId: teamFilter === "all" ? undefined : Number(teamFilter),
        currentAssigneeId:
          assigneeFilter === "all" ? undefined : Number(assigneeFilter),
      })
      setItems(Array.isArray(data.results) ? data.results : [])
      setPaging(data.page ?? emptyPaging)
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "加载工单失败")
    } finally {
      setLoading(false)
    }
  }, [
    assigneeFilter,
    keyword,
    paging.limit,
    paging.page,
    priorityFilter,
    statusFilter,
    teamFilter,
  ])

  useEffect(() => {
    void loadTickets()
  }, [loadTickets])

  useEffect(() => {
    void (async () => {
      const [teamData, agentData] = await Promise.all([
        fetchAgentTeamsAll(),
        fetchAgentProfilesAll(),
      ])
      setTeams(Array.isArray(teamData) ? teamData : [])
      setAgents(Array.isArray(agentData) ? agentData : [])
    })()
  }, [])

  async function handleSubmit(payload: Parameters<typeof createTicket>[0] | Parameters<typeof updateTicket>[0]) {
    if ("ticketId" in payload) {
      await updateTicket(payload)
      toast.success("工单已更新")
    } else {
      await createTicket(payload)
      toast.success("工单已创建")
    }
    await loadTickets()
  }

  const teamOptions = [{ value: "all", label: "全部团队" }].concat(
    teams.map((team) => ({ value: String(team.id), label: team.name })),
  )
  const agentOptions = [{ value: "all", label: "全部处理人" }].concat(
    agents.map((agent) => ({
      value: String(agent.userId),
      label:
        agent.displayName ||
        agent.nickname ||
        agent.username ||
        `客服#${agent.userId}`,
    })),
  )

  return (
    <div className="min-h-0 flex-1 overflow-auto bg-muted/20 p-4 md:p-6">
      <div className="mx-auto flex max-w-7xl flex-col gap-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between gap-4">
            <div>
              <CardTitle>工单</CardTitle>
              <p className="mt-1 text-sm text-muted-foreground">
                集中处理异步问题、转派、回复与关闭
              </p>
            </div>
            <Button
              onClick={() => {
                setEditingItem(null)
                setDialogOpen(true)
              }}
            >
              <PlusIcon className="size-4" />
              新建工单
            </Button>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid gap-3 lg:grid-cols-[minmax(0,1.4fr)_repeat(4,minmax(0,1fr))_auto]">
              <div className="relative">
                <SearchIcon className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
                <Input
                  className="pl-9"
                  value={keywordInput}
                  placeholder="搜索工单号、标题或描述"
                  onChange={(event) => setKeywordInput(event.target.value)}
                  onKeyDown={(event) => {
                    if (event.key === "Enter") {
                      setPaging((current) => ({ ...current, page: 1 }))
                      setKeyword(keywordInput)
                    }
                  }}
                />
              </div>
              <OptionCombobox
                value={statusFilter}
                onChange={(value) => {
                  setStatusFilter(value)
                  setPaging((current) => ({ ...current, page: 1 }))
                }}
                placeholder="全部状态"
                options={[
                  { value: "all", label: "全部状态" },
                  { value: "new", label: "新建" },
                  { value: "open", label: "处理中" },
                  { value: "pending_customer", label: "待客户反馈" },
                  { value: "pending_internal", label: "待内部处理" },
                  { value: "resolved", label: "已解决" },
                  { value: "closed", label: "已关闭" },
                  { value: "cancelled", label: "已取消" },
                ]}
              />
              <OptionCombobox
                value={priorityFilter}
                onChange={(value) => {
                  setPriorityFilter(value)
                  setPaging((current) => ({ ...current, page: 1 }))
                }}
                placeholder="全部优先级"
                options={[
                  { value: "all", label: "全部优先级" },
                  { value: "1", label: "低" },
                  { value: "2", label: "普通" },
                  { value: "3", label: "高" },
                  { value: "4", label: "紧急" },
                ]}
              />
              <OptionCombobox
                value={teamFilter}
                onChange={(value) => {
                  setTeamFilter(value)
                  setPaging((current) => ({ ...current, page: 1 }))
                }}
                placeholder="全部团队"
                options={teamOptions}
              />
              <OptionCombobox
                value={assigneeFilter}
                onChange={(value) => {
                  setAssigneeFilter(value)
                  setPaging((current) => ({ ...current, page: 1 }))
                }}
                placeholder="全部处理人"
                options={agentOptions}
              />
              <div className="flex gap-2">
                <Button
                  variant="outline"
                  onClick={() => {
                    setPaging((current) => ({ ...current, page: 1 }))
                    setKeyword(keywordInput)
                  }}
                >
                  查询
                </Button>
                <Button variant="outline" onClick={() => void loadTickets()}>
                  <RefreshCcwIcon className="size-4" />
                </Button>
              </div>
            </div>

            <div className="rounded-lg border bg-background">
              <Table>
                <TableHeader className="bg-muted/35">
                  <TableRow>
                    <TableHead>工单</TableHead>
                    <TableHead>客户</TableHead>
                    <TableHead>优先级</TableHead>
                    <TableHead>状态</TableHead>
                    <TableHead>处理人</TableHead>
                    <TableHead>团队</TableHead>
                    <TableHead>更新时间</TableHead>
                    <TableHead className="text-right">操作</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {loading ? (
                    <TableRow>
                      <TableCell colSpan={8} className="h-32 text-center text-muted-foreground">
                        加载中...
                      </TableCell>
                    </TableRow>
                  ) : items.length > 0 ? (
                    items.map((item) => (
                      <TableRow key={item.id}>
                        <TableCell>
                          <div className="space-y-1">
                            <div className="font-medium">{item.title}</div>
                            <div className="text-xs text-muted-foreground">{item.ticketNo}</div>
                          </div>
                        </TableCell>
                        <TableCell>{item.customer?.name || "未绑定客户"}</TableCell>
                        <TableCell>
                          <TicketPriorityBadge priority={item.priority} />
                        </TableCell>
                        <TableCell>
                          <TicketStatusBadge status={item.status} />
                        </TableCell>
                        <TableCell>{item.currentAssigneeName || "未指派"}</TableCell>
                        <TableCell>{item.currentTeamName || "未分组"}</TableCell>
                        <TableCell>
                          {item.updatedAt ? formatDateTime(item.updatedAt) : "—"}
                        </TableCell>
                        <TableCell className="text-right">
                          <div className="flex justify-end gap-2">
                            <Link href={`/workspace/tickets/${item.id}`}>
                              <Button variant="outline" size="sm">详情</Button>
                            </Link>
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() => {
                                setEditingItem(item)
                                setDialogOpen(true)
                              }}
                            >
                              编辑
                            </Button>
                          </div>
                        </TableCell>
                      </TableRow>
                    ))
                  ) : (
                    <TableRow>
                      <TableCell colSpan={8} className="h-32 text-center text-muted-foreground">
                        暂无工单
                      </TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
            </div>

            <ListPagination
              page={paging.page}
              total={paging.total}
              limit={paging.limit}
              loading={loading}
              onPageChange={(page) => setPaging((current) => ({ ...current, page }))}
              onLimitChange={(limit) =>
                setPaging((current) => ({ ...current, page: 1, limit }))
              }
            />
          </CardContent>
        </Card>
      </div>

      <TicketEditDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        item={editingItem}
        onSubmit={handleSubmit}
      />
    </div>
  )
}
