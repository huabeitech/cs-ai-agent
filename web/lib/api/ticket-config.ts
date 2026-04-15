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

export type TicketResolutionCode = {
  id: number
  name: string
  code: string
  sortNo: number
  status: number
  remark: string
}

export type TicketPriorityConfig = {
  id: number
  name: string
  sortNo: number
  firstResponseMinutes: number
  resolutionMinutes: number
  status: number
  remark: string
}

export type CreateTicketResolutionCodePayload = {
  name: string
  code: string
  sortNo: number
  status: number
  remark: string
}

export type UpdateTicketResolutionCodePayload = CreateTicketResolutionCodePayload & {
  id: number
}

export type CreateTicketPriorityConfigPayload = {
  name: string
  firstResponseMinutes: number
  resolutionMinutes: number
  status: number
  remark: string
}

export type UpdateTicketPriorityConfigPayload = CreateTicketPriorityConfigPayload & {
  id: number
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

export function fetchTicketResolutionCodes(query?: Record<string, string | number | undefined>) {
  return request<PageResult<TicketResolutionCode>>(
    `/api/dashboard/ticket-resolution-code/list${toQueryString(query)}`
  )
}

export function fetchTicketResolutionCodesAll() {
  return request<TicketResolutionCode[]>("/api/dashboard/ticket-resolution-code/list_all")
}

export function createTicketResolutionCode(payload: CreateTicketResolutionCodePayload) {
  return request<TicketResolutionCode>("/api/dashboard/ticket-resolution-code/create", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function updateTicketResolutionCode(payload: UpdateTicketResolutionCodePayload) {
  return request<void>("/api/dashboard/ticket-resolution-code/update", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function deleteTicketResolutionCode(id: number) {
  return request<void>("/api/dashboard/ticket-resolution-code/delete", {
    method: "POST",
    body: JSON.stringify({ id }),
  })
}

export function fetchTicketPriorityConfigs(query?: Record<string, string | number | undefined>) {
  return request<TicketPriorityConfig[]>(
    `/api/dashboard/ticket-priority-config/list${toQueryString(query)}`
  )
}

export function fetchTicketPriorityConfigsAll() {
  return request<TicketPriorityConfig[]>("/api/dashboard/ticket-priority-config/list_all")
}

export function createTicketPriorityConfig(payload: CreateTicketPriorityConfigPayload) {
  return request<TicketPriorityConfig>("/api/dashboard/ticket-priority-config/create", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function updateTicketPriorityConfig(payload: UpdateTicketPriorityConfigPayload) {
  return request<void>("/api/dashboard/ticket-priority-config/update", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function updateTicketPriorityConfigSort(ids: number[]) {
  return request<void>("/api/dashboard/ticket-priority-config/update_sort", {
    method: "POST",
    body: JSON.stringify(ids),
  })
}

export function deleteTicketPriorityConfig(id: number) {
  return request<void>("/api/dashboard/ticket-priority-config/delete", {
    method: "POST",
    body: JSON.stringify({ id }),
  })
}
