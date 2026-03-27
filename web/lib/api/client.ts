import { clearSession, readSession, writeSession, type AuthSession } from "@/lib/auth"

const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL?.trim() || "http://127.0.0.1:8083"

type JsonResult<T> = {
  errorCode: number
  message: string
  data: T
  success: boolean
}

type RequestOptions = RequestInit & {
  skipAuth?: boolean
  retryOnAuthError?: boolean
}

async function parseResult<T>(response: Response) {
  const payload = (await response.json()) as JsonResult<T>
  if (!response.ok || !payload.success) {
    const error = new Error(payload.message || "请求失败")
    ;(error as Error & { errorCode?: number }).errorCode = payload.errorCode
    throw error
  }
  return payload.data
}

async function refreshAccessToken() {
  const session = readSession()
  if (!session?.refreshToken) {
    clearSession()
    return null
  }

  const data = await request<AuthSession>(
    "/api/auth/refresh_token",
    {
      method: "POST",
      body: JSON.stringify({ refreshToken: session.refreshToken }),
      skipAuth: true,
      headers: {
        "Content-Type": "application/json",
      },
    },
    false
  )
  const merged = {
    ...data,
    refreshToken: data.refreshToken || session.refreshToken,
  }
  writeSession(merged)
  return merged
}

export async function request<T>(
  path: string,
  options: RequestOptions = {},
  retryOnAuthError = true
): Promise<T> {
  const { headers, skipAuth, ...rest } = options
  delete (rest as RequestOptions).retryOnAuthError
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

  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...rest,
    headers: authHeaders,
    cache: "no-store",
  })

  try {
    return await parseResult<T>(response)
  } catch (error) {
    const errorCode = (error as Error & { errorCode?: number }).errorCode
    if (
      !skipAuth &&
      retryOnAuthError &&
      (errorCode === 3000 || errorCode === 3002) &&
      session?.refreshToken
    ) {
      const refreshed = await refreshAccessToken()
      if (!refreshed) {
        throw error
      }
      return request<T>(path, options, false)
    }

    if (errorCode === 3000 || errorCode === 3002) {
      clearSession()
    }
    throw error
  }
}
