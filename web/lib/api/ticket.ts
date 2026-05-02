import { request } from "@/lib/api/client"
import type { Tag } from "@/lib/api/admin"

export type Paging = {
  page: number
  limit: number
  total: number
}

export type PageResult<T> = {
  results: T[]
  page: Paging
}

export type TicketStatus = "pending" | "in_progress" | "done"
export type TicketSource = "manual" | "conversation"

export type TicketCustomer = {
  id: number
  name: string
  companyId?: number
  company?: {
    id: number
    name: string
    code?: string
    remark?: string
    createdAt?: string
    updatedAt?: string
  }
  primaryMobile?: string
  primaryEmail?: string
}

export type TicketProgress = {
  id: number
  ticketId: number
  content: string
  authorId: number
  authorName?: string
  createdAt?: string
}

export type TicketItem = {
  id: number
  ticketNo: string
  title: string
  description: string
  source: TicketSource
  channel: string
  customerId: number
  conversationId: number
  tags?: Tag[]
  status: TicketStatus
  currentAssigneeId: number
  currentAssigneeName?: string
  createdBy: number
  createdByName?: string
  handledAt?: string
  createdAt?: string
  updatedAt?: string
  customer?: TicketCustomer
}

export type TicketDetail = {
  ticket: TicketItem
  progresses?: TicketProgress[]
}

export type TicketSummary = {
  all: number
  pending: number
  inProgress: number
  done: number
  unassigned: number
  mine: number
  stale: number
}

export type TicketSavedView = {
  id: number
  name: string
  filters?: Record<string, unknown>
  sortNo: number
}

export type TicketListQuery = {
  page?: number
  limit?: number
  keyword?: string
  status?: TicketStatus
  tagId?: number
  currentAssigneeId?: number
  customerId?: number
  conversationId?: number
  source?: TicketSource
  channel?: string
  mine?: number | boolean
  unassigned?: number | boolean
}

export type CreateTicketPayload = {
  title: string
  description: string
  source?: TicketSource
  channel?: string
  customerId?: number
  conversationId?: number
  tagIds?: number[]
  currentAssigneeId?: number
}

export type CreateTicketFromConversationPayload = {
  conversationId: number
  title: string
  description: string
  tagIds?: number[]
  currentAssigneeId?: number
}

export type UpdateTicketPayload = {
  ticketId: number
  title: string
  description: string
  tagIds?: number[]
  currentAssigneeId?: number
}

function toQueryString(query?: Record<string, string | number | boolean | undefined>) {
  if (!query) {
    return ""
  }

  const params = new URLSearchParams()
  Object.entries(query).forEach(([key, value]) => {
    if (value === undefined || value === "") {
      return
    }
    params.set(key, String(value))
  })
  const output = params.toString()
  return output ? `?${output}` : ""
}

export function fetchTickets(query?: TicketListQuery) {
  return request<PageResult<TicketItem>>(`/api/dashboard/ticket/list${toQueryString(query)}`)
}

export function fetchTicketSummary(query?: { staleHours?: number }) {
  return request<TicketSummary>(`/api/dashboard/ticket/summary${toQueryString(query)}`)
}

export function fetchTicketDetail(id: number) {
  return request<TicketDetail>(`/api/dashboard/ticket/${id}`)
}

export function createTicket(payload: CreateTicketPayload) {
  return request<TicketItem>("/api/dashboard/ticket/create", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function createTicketFromConversation(payload: CreateTicketFromConversationPayload) {
  return request<TicketItem>("/api/dashboard/ticket/create_from_conversation", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function updateTicket(payload: UpdateTicketPayload) {
  return request<void>("/api/dashboard/ticket/update", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function assignTicket(payload: {
  ticketId: number
  toUserId: number
  reason?: string
}) {
  return request<void>("/api/dashboard/ticket/assign", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function batchAssignTickets(payload: {
  ticketIds: number[]
  toUserId: number
  reason?: string
}) {
  return request<void>("/api/dashboard/ticket/batch_assign", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function changeTicketStatus(payload: {
  ticketId: number
  status: TicketStatus
}) {
  return request<void>("/api/dashboard/ticket/change_status", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function fetchTicketProgresses(query: {
  ticketId: number
  page?: number
  limit?: number
}) {
  return request<PageResult<TicketProgress>>(
    `/api/dashboard/ticket/progress/list${toQueryString(query)}`,
  )
}

export function createTicketProgress(payload: {
  ticketId: number
  content: string
}) {
  return request<TicketProgress>("/api/dashboard/ticket/progress/create", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function fetchTicketViews() {
  return request<TicketSavedView[]>("/api/dashboard/ticket/view_list")
}

export function saveTicketView(payload: {
  id?: number
  name: string
  filters: Record<string, unknown>
}) {
  return request<TicketSavedView>("/api/dashboard/ticket/save_view", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function deleteTicketView(id: number) {
  return request<void>("/api/dashboard/ticket/delete_view", {
    method: "POST",
    body: JSON.stringify({ id }),
  })
}
