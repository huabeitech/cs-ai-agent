"use client"

import type { ReactNode } from "react"
import { useRouter } from "next/navigation"
import { useEffect } from "react"

import { ImageLightboxProvider } from "@/components/image-lightbox"
import { WorkspaceHeader } from "@/components/workspace-header"
import { WorkspaceSidebar } from "@/components/workspace-sidebar"
import { useAuth } from "@/components/auth-provider"
import { Card, CardContent } from "@/components/ui/card"

export default function AgentWorkbenchLayout({
  children,
}: {
  children: ReactNode
}) {
  const { ready, session } = useAuth()
  const router = useRouter()

  useEffect(() => {
    if (ready && !session) {
      router.replace("/login")
    }
  }, [ready, router, session])

  if (!ready || !session) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-[linear-gradient(160deg,#f3f4f6_0%,#fff7ed_45%,#ecfeff_100%)] p-6">
        <Card className="w-full max-w-md border-0 bg-white/90 shadow-xl shadow-slate-200/60 backdrop-blur">
          <CardContent className="flex flex-col items-center gap-3 py-12 text-center">
            <div className="size-10 animate-pulse rounded-full bg-primary/10" />
            <div className="space-y-1">
              <p className="text-base font-medium">正在校验客服登录态</p>
              <p className="text-sm text-muted-foreground">
                将自动同步当前客服信息与权限数据
              </p>
            </div>
          </CardContent>
        </Card>
      </div>
    )
  }

  return (
    <ImageLightboxProvider>
      <div className="flex h-svh min-h-0 flex-col">
        <WorkspaceHeader />
        <div className="flex min-h-0 flex-1 overflow-hidden">
          <WorkspaceSidebar />
          <main className="flex min-h-0 min-w-0 flex-1 overflow-hidden">
            {children}
          </main>
        </div>
      </div>
    </ImageLightboxProvider>
  )
}
