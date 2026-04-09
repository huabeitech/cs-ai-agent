"use client"

import type { ComponentProps } from "react"
import Link from "next/link"
import { useMemo } from "react"

import {
  filterDashboardNavForSession,
  filterDashboardSecondaryNavForSession,
} from "@/lib/navigation"
import { useAuth } from "@/components/auth-provider"
import { NavMain } from "@/components/nav-main"
import { NavSecondary } from "@/components/nav-secondary"
import { NavUser } from "@/components/nav-user"
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar"
import { BotMessageSquareIcon } from "lucide-react"

export function AppSidebar({ ...props }: ComponentProps<typeof Sidebar>) {
  const { session } = useAuth()
  const navSections = useMemo(
    () => filterDashboardNavForSession(session?.permissions, session?.roles),
    [session?.permissions, session?.roles]
  )
  const secondaryNavItems = useMemo(
    () => filterDashboardSecondaryNavForSession(session?.permissions, session?.roles),
    [session?.permissions, session?.roles]
  )
  const user = {
    name: session?.user.nickname || session?.user.username || "未登录",
    email: session?.user.username || "guest",
    avatar: session?.user.avatar || "",
  }

  return (
    <Sidebar collapsible="icon" {...props}>
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton
              className="data-[slot=sidebar-menu-button]:p-1.5!"
              render={<Link href="/" />}
            >
              <BotMessageSquareIcon className="size-5!" />
              <span className="text-base font-semibold">贝壳AGENT</span>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>
      <SidebarContent>
        {navSections.map((section) => (
          <NavMain key={section.title} title={section.title} items={section.items} />
        ))}
        {secondaryNavItems.length > 0 ? (
          <NavSecondary items={secondaryNavItems} className="mt-auto" />
        ) : null}
      </SidebarContent>
      <SidebarFooter>
        <NavUser user={user} />
      </SidebarFooter>
    </Sidebar>
  )
}
