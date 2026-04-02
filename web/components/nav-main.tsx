"use client"

import Link from "next/link"
import { usePathname } from "next/navigation"

import {
  SidebarGroupLabel,
  SidebarGroup,
  SidebarGroupContent,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar"

export function NavMain({
  title,
  items,
}: {
  title: string
  items: ReadonlyArray<{
    title: string
    url: string
    icon?: React.ReactNode
  }>
}) {
  const pathname = usePathname()

  const isActive = (itemUrl: string) => {
    if (itemUrl === "/") {
      return pathname === itemUrl
    }
    return pathname === itemUrl || pathname.startsWith(itemUrl + "/")
  }

  return (
    <SidebarGroup>
      <SidebarGroupLabel>{title}</SidebarGroupLabel>
      <SidebarGroupContent>
        <SidebarMenu>
          {items.map((item) => (
            <SidebarMenuItem key={item.title}>
              <SidebarMenuButton
                tooltip={item.title}
                render={<Link href={item.url} />}
                isActive={isActive(item.url)}
              >
                {item.icon}
                <span>{item.title}</span>
              </SidebarMenuButton>
            </SidebarMenuItem>
          ))}
        </SidebarMenu>
      </SidebarGroupContent>
    </SidebarGroup>
  )
}
