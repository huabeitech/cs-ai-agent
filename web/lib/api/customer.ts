import { request } from "@/lib/api/client"
import type { PageResult } from "@/lib/api/admin"
import { AdminCompany } from "./company"

export type AdminCustomer = {
  id: number
  name: string
  gender: number
  companyId: number
  company?: AdminCompany
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

/** 与 POST /customer/save/profile 请求体一致 */
export type SaveCustomerProfileContactLine = {
  id?: number
  contactType: string
  contactValue: string
  remark: string
  isPrimary: boolean
}

export type SaveCustomerProfilePayload = {
  id?: number
  name: string
  gender: number
  companyId: number
  remark: string
  contacts: SaveCustomerProfileContactLine[]
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

/** 单请求 + 单事务保存客户主信息与联系方式全量 */
export function saveCustomerProfile(payload: SaveCustomerProfilePayload) {
  return request<AdminCustomer>("/api/console/customer/save_profile", {
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

