import { request } from "@/lib/api/client"

export type CustomerContactType = "mobile" | "email" | "wechat" | "other"

export type AdminCustomerContact = {
  id: number
  customerId: number
  contactType: CustomerContactType | string
  contactValue: string
  isPrimary: boolean
  isVerified: boolean
  verifiedAt?: string
  source: string
  status: number
  remark: string
  createdAt: string
  updatedAt: string
}

export type CreateCustomerContactPayload = {
  customerId: number
  contactType: string
  contactValue: string
  isPrimary: boolean
  isVerified: boolean
  source: string
  status: number
  remark: string
}

export type UpdateCustomerContactPayload = Omit<
  CreateCustomerContactPayload,
  "customerId"
> & {
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

export function fetchCustomerContacts(customerId: number) {
  return request<AdminCustomerContact[]>(
    `/api/console/customer-contact/list${toQueryString({ customerId })}`
  )
}

export function createCustomerContact(payload: CreateCustomerContactPayload) {
  return request<AdminCustomerContact>("/api/console/customer-contact/create", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}

export function updateCustomerContact(payload: UpdateCustomerContactPayload) {
  return request<void>("/api/console/customer-contact/update", {
    method: "POST",
    body: JSON.stringify(payload),
  })
}
