"use client"

import { useEffect, useRef } from "react"
import { usePathname } from "next/navigation"

import { getPageTitle } from "@/lib/navigation"
import { ThemeToggle } from "@/components/theme-toggle"
import { WorkspaceToggle } from "@/components/workspace-toggle"
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbList,
  BreadcrumbPage,
} from "@/components/ui/breadcrumb"
import { Separator } from "@/components/ui/separator"
import { SidebarTrigger, useSidebar } from "@/components/ui/sidebar"

const SIDEBAR_STORAGE_KEY = "dashboard_sidebar_open"

export function SiteHeader() {
  const pathname = usePathname()
  const { open, setOpen, isMobile } = useSidebar()
  const pageTitle = getPageTitle(pathname)
  const hasRestoredRef = useRef(false)

  useEffect(() => {
    if (hasRestoredRef.current || isMobile) {
      return
    }

    hasRestoredRef.current = true
    const storedValue = window.localStorage.getItem(SIDEBAR_STORAGE_KEY)
    if (storedValue === null) {
      return
    }

    setOpen(storedValue === "true")
  }, [isMobile, setOpen])

  useEffect(() => {
    if (!hasRestoredRef.current || isMobile) {
      return
    }

    window.localStorage.setItem(SIDEBAR_STORAGE_KEY, String(open))
  }, [isMobile, open])

  return (
    <header className="flex h-(--header-height) shrink-0 items-center gap-2 border-b transition-[width,height] ease-linear group-has-data-[collapsible=icon]/sidebar-wrapper:h-(--header-height)">
      <div className="flex w-full items-center justify-between gap-3 px-4 lg:px-6">
        <div className="flex min-w-0 items-center gap-2">
          <SidebarTrigger className="-ml-1" />
          <Separator
            orientation="vertical"
            className="mx-2 h-4 data-vertical:self-auto"
          />
          <div className="min-w-0">
            <Breadcrumb>
              <BreadcrumbList>
                <BreadcrumbItem>
                  <BreadcrumbPage>{pageTitle}</BreadcrumbPage>
                </BreadcrumbItem>
              </BreadcrumbList>
            </Breadcrumb>
          </div>
        </div>
        <div className="flex items-center justify-end gap-3">
          <WorkspaceToggle />
          <ThemeToggle />
        </div>
      </div>
    </header>
  )
}
