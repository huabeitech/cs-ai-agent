import { expireSession, readSession } from "@/lib/auth"

const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL?.trim() || ""

type JsonResult<T> = {
  errorCode: number
  message: string
  data: T
  success: boolean
}

type RequestOptions = RequestInit & {
  skipAuth?: boolean
  baseUrl?: string
  onResponse?: (response: Response) => void
}

async function parseResult<T>(response: Response) {
  const payload = (await response.json()) as JsonResult<T>
  if (!response.ok || !payload.success) {
    if (payload.errorCode === 3000 || payload.errorCode === 3002) {
      expireSession()
    }
    const error = new Error(payload.message || "请求失败")
    ;(error as Error & { errorCode?: number }).errorCode = payload.errorCode
    throw error
  }
  return payload.data
}

export async function request<T>(
  path: string,
  options: RequestOptions = {}
): Promise<T> {
  const { headers, skipAuth, baseUrl, onResponse, ...rest } = options
  delete (rest as RequestOptions).baseUrl
  delete (rest as RequestOptions).onResponse
  const session = readSession()
  const authHeaders = new Headers(headers)

  if (!skipAuth && session?.accessToken) {
    authHeaders.set("Authorization", `Bearer ${session.accessToken}`)
  }
  if (
    !authHeaders.has("Content-Type") &&
    rest.body &&
    !(typeof FormData !== "undefined" && rest.body instanceof FormData)
  ) {
    authHeaders.set("Content-Type", "application/json")
  }

  const requestBaseUrl = baseUrl !== undefined ? baseUrl : API_BASE_URL
  const response = await fetch(`${requestBaseUrl}${path}`, {
    ...rest,
    headers: authHeaders,
    cache: "no-store",
  })
  onResponse?.(response)

  return parseResult<T>(response)
}
