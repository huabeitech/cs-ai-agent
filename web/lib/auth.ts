export type AuthUser = {
  id: number
  username: string
  nickname: string
  avatar: string
  status: number
  roles: string[]
}

export type AuthSession = {
  accessToken: string
  expiresAt?: string
  user: AuthUser
  permissions: string[]
  roles: string[]
}

const SESSION_STORAGE_KEY = "cs-ai-agent-session"
export const AUTH_SESSION_EXPIRED_EVENT = "cs-ai-agent-auth-expired"

function hasWindow() {
  return typeof window !== "undefined"
}

export function readSession(): AuthSession | null {
  if (!hasWindow()) {
    return null
  }

  const raw = window.localStorage.getItem(SESSION_STORAGE_KEY)
  if (!raw) {
    return null
  }

  try {
    return JSON.parse(raw) as AuthSession
  } catch {
    window.localStorage.removeItem(SESSION_STORAGE_KEY)
    return null
  }
}

export function writeSession(session: AuthSession) {
  if (!hasWindow()) {
    return
  }
  window.localStorage.setItem(SESSION_STORAGE_KEY, JSON.stringify(session))
}

export function clearSession() {
  if (!hasWindow()) {
    return
  }
  window.localStorage.removeItem(SESSION_STORAGE_KEY)
}

export function expireSession() {
  if (!hasWindow()) {
    return
  }
  clearSession()
  window.dispatchEvent(new Event(AUTH_SESSION_EXPIRED_EVENT))
}
