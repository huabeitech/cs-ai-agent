"use client"

import type { CSSProperties, ReactNode } from "react"
import { useRouter } from "next/navigation"
import { useEffect } from "react"

import { AppSidebar } from "@/components/app-sidebar"
import { useAuth } from "@/components/auth-provider"
import { SiteHeader } from "@/components/site-header"
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar"
import { Card, CardContent } from "@/components/ui/card"

export default function DashboardLayout({
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
              <p className="text-base font-medium">正在校验后台登录态</p>
              <p className="text-sm text-muted-foreground">
                将自动同步当前管理员信息与权限数据
              </p>
            </div>
          </CardContent>
        </Card>
      </div>
    )
  }

  return (
    <SidebarProvider
      style={
        {
          "--sidebar-width": "calc(var(--spacing) * 54)",
          "--header-height": "calc(var(--spacing) * 12)",
        } as CSSProperties
      }
    >
      <AppSidebar variant="inset" />
      <SidebarInset>
        <SiteHeader />
        <div className="flex flex-1 flex-col">
          <div className="@container/main flex flex-1 flex-col gap-2">
            {children}
          </div>
        </div>
      </SidebarInset>
    </SidebarProvider>
  )
}
