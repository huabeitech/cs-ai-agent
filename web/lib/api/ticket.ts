import { request } from "@/lib/api/client"

export type Paging = {
  page: number
  limit: number
  total: number
}

export type PageResult<T> = {
  results: T[]
  page: Paging
}

export type TicketCustomer = {
  id: number
  name: string
  primaryMobile?: string
  primaryEmail?: string
}

export type TicketSLA = {
  slaType: string
  targetMinutes: number
  status: string
  startedAt?: string
  pausedAt?: string
  stoppedAt?: string
  breachedAt?: string
  elapsedMin: number
}

export type TicketComment = {
  id: number
  ticketId: number
  commentType: string
  authorType: string
  authorId: number
  authorName?: string
  contentType: string
  content: string
  payload?: string
  createdAt?: string
}

export type TicketEvent = {
  id: number
  ticketId: number
  eventType: string
  operatorType: string
  operatorId: number
  operatorName?: string
  oldValue?: string
  newValue?: string
  content?: string
  payload?: string
  createdAt?: string
}

export type TicketItem = {
  id: number
  ticketNo: string
  title: string
  description: string
  source: string
  channel: string
  customerId: number
  conversationId: number
  categoryId: number
  categoryName?: string
  type: string
  priority: number
  severity: number
  status: string
  currentTeamId: number
  currentTeamName?: string
  currentAssigneeId: number
  currentAssigneeName?: string
  watchedByMe: boolean
  pendingReason?: string
  closeReason?: string
  resolutionCode?: string
  resolutionCodeName?: string
  resolutionSummary?: string
  firstResponseAt?: string
  resolvedAt?: string
  closedAt?: string
  dueAt?: string
  nextReplyDeadlineAt?: string
  resolveDeadlineAt?: string
  reopenedCount: number
  createdAt?: string
  updatedAt?: string
  customer?: TicketCustomer
  sla?: TicketSLA[]
}

export type TicketDetail = {
  ticket: TicketItem
  watchers?: Array<{
    id: number
    userId: number
    userName?: string
  }>
  collaborators?: TicketCollaborator[]
  comments?: TicketComment[]
  events?: TicketEvent[]
  relatedTickets?: TicketRelation[]
}

export type TicketCollaborator = {
  id: number
  userId: number
  userName?: string
  teamName?: string
}

export type TicketRelation = {
  id: number
  ticketId: number
  relatedTicketId: number
  relationType: string
  relatedTicketNo?: string
  relatedTicketTitle?: string
  relatedTicketStatus?: string
  currentTeamName?: string
  currentAssigneeName?: string
  updatedAt?: string
}

export type TicketSummary = {
  all: number
  mine: number
  watching: number
  participating: number
  unassigned: number
  pendingCustomer: number
  pendingInternal: number
  overdue: number
}

export type TicketRiskReason = {
  code: string
  title: string
  description: string
  count: number
}

export type TicketRiskOverview = {
  overdue: number
  highRisk: number
  unassigned: number
  pendingInternal: number
  pendingCustomer: number
  riskWindowMins: number
  reasons?: TicketRiskReason[]
}

export type TicketListQuery = {
  page?: number
  limit?: number
  keyword?: string
  status?: string
  priority?: number
  severity?: number
  categoryId?: number
  currentTeamId?: number
  currentAssigneeId?: number
  customerId?: number
  conversationId?: number
  source?: string
  watching?: number
  collaborating?: number
  mine?: number
  unassigned?: number
  overdue?: number
}

export type CreateTicketPayload = {
  title: string
  description: string
  source?: string
  channel?: string
  customerId?: number
  conversationId?: number
  categoryId?: number
  type?: string
  priority: number
  severity: number
  currentTeamId?: number
  currentAssigneeId?: number
  dueAt?: string
  customFields?: Record<string, unknown>
}

export type CreateTicketFromConversationPayload = {
  conversationId: number
  title: string
  description: string
  categoryId?: number
  priority: number
  severity: number
  currentTeamId?: number
  currentAssigneeId?: number
  syncToConversation: boolean
  customFields?: Record<string, unknown>
}

export type UpdateTicketPayload = {
  ticketId: number
  title: string
  description: string
  categoryId?: number
  type?: string
  priority: number
  severity: number
  currentTeamId?: number
  currentAssigneeId?: number
  dueAt?: string
  customFields?: Record<string, unknown>
}

function toQueryString(query?: Record<string, string | number | undefined>) {
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
  return request<PageResult<TicketItem>>(`/api/console/ticket/list${toQueryString(query)}`)
}

export function fetchTicketDetail(id: number) {
  return request<TicketDetail>(`/api/console/ticket/${id}`)
}

export function fetchTicketSummary() {
  return request<TicketSummary>("/api/console/ticket/summary")
}

export function fetchTicketRiskOverview(query?: {
  currentTeamId?: number
  riskWindowMins?: number
}) {
  return request<TicketRiskOverview>(`/api/console/ticket/risk_overview${toQueryString(query)}`)
}

export function createTicket(payload: CreateTicketPayload) {
  return request<TicketItem>("/api/console/ticket/create", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function createTicketFromConversation(payload: CreateTicketFromConversationPayload) {
  return request<TicketItem>("/api/console/ticket/create_from_conversation", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function updateTicket(payload: UpdateTicketPayload) {
  return request<void>("/api/console/ticket/update", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function assignTicket(payload: {
  ticketId: number
  toUserId: number
  toTeamId?: number
  reason?: string
}) {
  return request<void>("/api/console/ticket/assign", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function batchAssignTickets(payload: {
  ticketIds: number[]
  toUserId: number
  toTeamId?: number
  reason?: string
}) {
  return request<void>("/api/console/ticket/batch_assign", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function changeTicketStatus(payload: {
  ticketId: number
  status: string
  pendingReason?: string
  closeReason?: string
  resolutionCode?: string
  resolutionSummary?: string
  reason?: string
}) {
  return request<void>("/api/console/ticket/change_status", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function batchChangeTicketStatus(payload: {
  ticketIds: number[]
  status: string
  pendingReason?: string
  closeReason?: string
  resolutionCode?: string
  resolutionSummary?: string
  reason?: string
}) {
  return request<void>("/api/console/ticket/batch_change_status", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function replyTicket(payload: {
  ticketId: number
  contentType?: string
  content: string
  payload?: string
}) {
  return request<TicketComment>("/api/console/ticket/reply", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function addTicketInternalNote(payload: {
  ticketId: number
  contentType?: string
  content: string
  payload?: string
}) {
  return request<TicketComment>("/api/console/ticket/internal_note", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function closeTicket(payload: { ticketId: number; closeReason: string }) {
  return request<void>("/api/console/ticket/close", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function reopenTicket(payload: { ticketId: number; reason: string }) {
  return request<void>("/api/console/ticket/reopen", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function watchTicket(ticketId: number) {
  return request<void>("/api/console/ticket/watch", {
    method: "POST",
    body: JSON.stringify({ ticketId }),
  })
}

export function unwatchTicket(ticketId: number) {
  return request<void>("/api/console/ticket/unwatch", {
    method: "POST",
    body: JSON.stringify({ ticketId }),
  })
}

export function batchWatchTickets(payload: { ticketIds: number[]; watched: boolean }) {
  return request<void>("/api/console/ticket/batch_watch", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function addTicketRelation(payload: {
  ticketId: number
  relatedTicketId?: number
  relatedTicketNo?: string
  relationType: string
}) {
  return request<void>("/api/console/ticket/add_relation", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function deleteTicketRelation(payload: { ticketId: number; relationId: number }) {
  return request<void>("/api/console/ticket/delete_relation", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function addTicketCollaborator(payload: { ticketId: number; userId: number }) {
  return request<void>("/api/console/ticket/add_collaborator", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function deleteTicketCollaborator(payload: { ticketId: number; collaboratorId: number }) {
  return request<void>("/api/console/ticket/delete_collaborator", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}
