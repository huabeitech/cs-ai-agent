import { request } from "@/lib/api/client"
import type { PageResult } from "@/lib/api/admin"

export type AdminCustomer = {
  id: number
  name: string
  gender: number
  companyId: number
  lastActiveAt?: string
  primaryMobile: string
  primaryEmail: string
  status: number
  remark: string
  createdAt: string
  updatedAt: string
}

export type CreateAdminCustomerPayload = {
  name: string
  gender: number
  companyId: number
  primaryMobile: string
  primaryEmail: string
  remark: string
}

export type UpdateAdminCustomerPayload = CreateAdminCustomerPayload & {
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

export function fetchCustomers(query?: Record<string, string | number | undefined>) {
  return request<PageResult<AdminCustomer>>(
    `/api/console/customer/list${toQueryString(query)}`
  )
}

export function fetchCustomer(id: number) {
  return request<AdminCustomer>(`/api/console/customer/${id}`)
}

export function createCustomer(payload: CreateAdminCustomerPayload) {
  return request<AdminCustomer>("/api/console/customer/create", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function updateCustomer(payload: UpdateAdminCustomerPayload) {
  return request<void>("/api/console/customer/update", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function updateCustomerStatus(id: number, status: number) {
  return request<void>("/api/console/customer/update_status", {
    method: "POST",
    body: JSON.stringify({ id, status }),
  })
}

export function deleteCustomer(id: number) {
  return request<void>("/api/console/customer/delete", {
    method: "POST",
    body: JSON.stringify({ id }),
  })
}

