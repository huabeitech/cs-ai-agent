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
      if (pathname?.startsWith("/dashboard")) {
        startTransition(() => {
          router.replace("/login")
        })
      }
    } finally {
      setReady(true)
    }
  }, [pathname, router])

  async function signOut() {
    const current = readSession()
    await logout(current?.refreshToken)
    setSession(null)
    startTransition(() => {
      router.replace("/login")
    })
  }

  useEffect(() => {
    const stored = readSession()
      setSession(stored)
    if (stored) {
      void refreshProfile()
      return
    }

    setReady(true)
    if (pathname?.startsWith("/dashboard")) {
      startTransition(() => {
        router.replace("/login")
      })
    }
  }, [pathname, refreshProfile, router])

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
