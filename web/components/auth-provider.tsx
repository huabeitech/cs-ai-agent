"use client"

import {
  createContext,
  startTransition,
  useContext,
  useCallback,
  useEffect,
  useState,
  type ReactNode,
} from "react"
import { usePathname, useRouter } from "next/navigation"

import { fetchProfile, logout } from "@/lib/api/auth"
import {
  AUTH_SESSION_EXPIRED_EVENT,
  clearSession,
  readSession,
  writeSession,
  type AuthSession,
} from "@/lib/auth"

type AuthContextValue = {
  session: AuthSession | null
  ready: boolean
  refreshProfile: () => Promise<void>
  signOut: () => Promise<void>
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const pathname = usePathname()
  const router = useRouter()
  const [session, setSession] = useState<AuthSession | null>(null)
  const [ready, setReady] = useState(false)
  const isDashboardLoginRoute = pathname?.startsWith("/dashboard/login") ?? false
  const requiresAuth = (pathname?.startsWith("/dashboard") ?? false) && !isDashboardLoginRoute

  const refreshProfile = useCallback(async () => {
    const stored = readSession()
    if (!stored) {
      setSession(null)
      setReady(true)
      return
    }

    try {
      const profile = await fetchProfile()
      const nextSession: AuthSession = {
        ...stored,
        user: profile.user,
        permissions: profile.permissions,
        roles: profile.roles,
      }
      writeSession(nextSession)
      setSession(nextSession)
    } catch {
      clearSession()
      setSession(null)
      if (requiresAuth) {
        startTransition(() => {
          router.replace("/dashboard/login")
        })
      }
    } finally {
      setReady(true)
    }
  }, [requiresAuth, router])

  async function signOut() {
    try {
      await logout()
    } finally {
      setSession(null)
      startTransition(() => {
        router.replace("/dashboard/login")
      })
    }
  }

  useEffect(() => {
    function handleAuthExpired() {
      setSession(null)
      if (requiresAuth) {
        startTransition(() => {
          router.replace("/dashboard/login")
        })
      }
    }

    window.addEventListener(AUTH_SESSION_EXPIRED_EVENT, handleAuthExpired)
    return () => {
      window.removeEventListener(AUTH_SESSION_EXPIRED_EVENT, handleAuthExpired)
    }
  }, [requiresAuth, router])

  useEffect(() => {
    const stored = readSession()
    setSession(stored)
    if (stored) {
      void refreshProfile()
      return
    }

    setReady(true)
    if (requiresAuth) {
      startTransition(() => {
        router.replace("/dashboard/login")
      })
    }
  }, [requiresAuth, refreshProfile, router])

  return (
    <AuthContext.Provider value={{ session, ready, refreshProfile, signOut }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) {
    throw new Error("useAuth must be used within AuthProvider")
  }
  return ctx
}
