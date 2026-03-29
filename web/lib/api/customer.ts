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

/** 与 POST /customer/list JSON Body 一致 */
export type CustomerListRequest = {
  page: number
  limit: number
  status?: number
  gender?: number
  companyId?: number
  /** 模糊匹配：客户名、主手机、主邮箱、联系方式、公司名称 */
  keyword?: string
}

export function fetchCustomers(body: CustomerListRequest) {
  return request<PageResult<AdminCustomer>>("/api/console/customer/list", {
    method: "POST",
    body: JSON.stringify(body),
  })
}

export function fetchCustomer(id: number) {
  return request<AdminCustomer | null>(`/api/console/customer/${id}`)
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
