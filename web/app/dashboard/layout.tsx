"use client"

import { Loader2Icon } from "lucide-react"
import { usePathname, useRouter } from "next/navigation"
import type { CSSProperties, ReactNode } from "react"
import { useEffect } from "react"

import { AppSidebar } from "@/components/app-sidebar"
import { useAuth } from "@/components/auth-provider"
import { NotificationProvider } from "@/components/notification-provider"
import { SiteHeader } from "@/components/site-header"
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar"

export default function DashboardLayout({
  children,
}: {
  children: ReactNode
}) {
  const { ready, session } = useAuth()
  const pathname = usePathname()
  const router = useRouter()
  const isLoginRoute = pathname?.startsWith("/dashboard/login") ?? false

  useEffect(() => {
    if (ready && !session && !isLoginRoute) {
      router.replace("/dashboard/login")
    }
  }, [isLoginRoute, ready, router, session])

  if (isLoginRoute) {
    return <>{children}</>
  }

  if (!ready || !session) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-[linear-gradient(160deg,#f3f4f6_0%,#fff7ed_45%,#ecfeff_100%)] p-6">
        <div className="flex items-center gap-3">
          <div className="flex size-10 items-center justify-center rounded-full bg-primary/10 text-primary">
            <Loader2Icon className="size-5 animate-spin" />
          </div>
          <div className="space-y-1">
            <p className="text-base font-medium">正在校验登录态</p>
            <p className="text-sm text-muted-foreground">
              将自动同步当前管理员信息与权限数据
            </p>
          </div>
        </div>
      </div>
    )
  }

  return (
    <SidebarProvider
      className="h-svh min-h-0 overflow-hidden"
      style={
        {
          "--sidebar-width": "calc(var(--spacing) * 54)",
          "--header-height": "calc(var(--spacing) * 12)",
        } as CSSProperties
      }
    >
      <NotificationProvider>
        <AppSidebar variant="inset" />
        <SidebarInset className="min-h-0 overflow-hidden">
          <SiteHeader />
          <div className="flex min-h-0 flex-1 flex-col overflow-hidden">
            <div className="@container/main flex min-h-0 flex-1 flex-col gap-2 overflow-hidden">
              {children}
            </div>
          </div>
        </SidebarInset>
      </NotificationProvider>
    </SidebarProvider>
  )
}
