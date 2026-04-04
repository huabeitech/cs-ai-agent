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

export type TicketCategory = {
  id: number
  name: string
  code: string
  parentId: number
  parentName?: string
  sortNo: number
  status: number
  remark: string
}

export type TicketResolutionCode = {
  id: number
  name: string
  code: string
  sortNo: number
  status: number
  remark: string
}

export type TicketSLAConfig = {
  id: number
  name: string
  priority: number
  firstResponseMinutes: number
  resolutionMinutes: number
  status: number
  remark: string
}

export type CreateTicketCategoryPayload = {
  name: string
  code: string
  parentId?: number
  sortNo: number
  status: number
  remark: string
}

export type UpdateTicketCategoryPayload = CreateTicketCategoryPayload & {
  id: number
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

export type CreateTicketSLAConfigPayload = {
  name: string
  priority: number
  firstResponseMinutes: number
  resolutionMinutes: number
  status: number
  remark: string
}

export type UpdateTicketSLAConfigPayload = CreateTicketSLAConfigPayload & {
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

export function fetchTicketCategories(query?: Record<string, string | number | undefined>) {
  return request<PageResult<TicketCategory>>(
    `/api/console/ticket-category/list${toQueryString(query)}`
  )
}

export function fetchTicketCategoriesAll() {
  return request<TicketCategory[]>("/api/console/ticket-category/list_all")
}

export function createTicketCategory(payload: CreateTicketCategoryPayload) {
  return request<TicketCategory>("/api/console/ticket-category/create", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function updateTicketCategory(payload: UpdateTicketCategoryPayload) {
  return request<void>("/api/console/ticket-category/update", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function deleteTicketCategory(id: number) {
  return request<void>("/api/console/ticket-category/delete", {
    method: "POST",
    body: JSON.stringify({ id }),
  })
}

export function fetchTicketResolutionCodes(query?: Record<string, string | number | undefined>) {
  return request<PageResult<TicketResolutionCode>>(
    `/api/console/ticket-resolution-code/list${toQueryString(query)}`
  )
}

export function fetchTicketResolutionCodesAll() {
  return request<TicketResolutionCode[]>("/api/console/ticket-resolution-code/list_all")
}

export function createTicketResolutionCode(payload: CreateTicketResolutionCodePayload) {
  return request<TicketResolutionCode>("/api/console/ticket-resolution-code/create", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function updateTicketResolutionCode(payload: UpdateTicketResolutionCodePayload) {
  return request<void>("/api/console/ticket-resolution-code/update", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function deleteTicketResolutionCode(id: number) {
  return request<void>("/api/console/ticket-resolution-code/delete", {
    method: "POST",
    body: JSON.stringify({ id }),
  })
}

export function fetchTicketSLAConfigs(query?: Record<string, string | number | undefined>) {
  return request<PageResult<TicketSLAConfig>>(
    `/api/console/ticket-sla-config/list${toQueryString(query)}`
  )
}

export function createTicketSLAConfig(payload: CreateTicketSLAConfigPayload) {
  return request<TicketSLAConfig>("/api/console/ticket-sla-config/create", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function updateTicketSLAConfig(payload: UpdateTicketSLAConfigPayload) {
  return request<void>("/api/console/ticket-sla-config/update", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function deleteTicketSLAConfig(id: number) {
  return request<void>("/api/console/ticket-sla-config/delete", {
    method: "POST",
    body: JSON.stringify({ id }),
  })
}
