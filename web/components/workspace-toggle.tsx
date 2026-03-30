"use client"

import { LayoutDashboardIcon, MessageSquareIcon } from "lucide-react"
import { usePathname } from "next/navigation"

import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"

type Workspace = "console" | "agent"

const workspaceOptions: Array<{
  value: Workspace
  label: string
  url: string
  icon: typeof LayoutDashboardIcon
}> = [
  { value: "console", label: "管理后台", url: "/dashboard", icon: LayoutDashboardIcon },
  { value: "agent", label: "客服工作台", url: "/workspace", icon: MessageSquareIcon },
]

export function WorkspaceToggle() {
  const pathname = usePathname()

  const isAgentWorkspace = pathname.startsWith("/workspace")
  const currentWorkspace = isAgentWorkspace ? "agent" : "console"
  const CurrentIcon = workspaceOptions.find((option) => option.value === currentWorkspace)?.icon ?? LayoutDashboardIcon

  const handleWorkspaceChange = (value: string) => {
    const workspace = value as Workspace
    const target = workspaceOptions.find((option) => option.value === workspace)
    if (target) {
      window.open(target.url, "_blank", "noopener,noreferrer")
    }
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger render={<Button variant="outline" size="sm" />} aria-label="切换工作台">
        <CurrentIcon />
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-40 min-w-40">
        <DropdownMenuRadioGroup
          value={currentWorkspace}
          onValueChange={handleWorkspaceChange}
        >
          {workspaceOptions.map((option) => {
            const Icon = option.icon
            return (
              <DropdownMenuRadioItem key={option.value} value={option.value}>
                <Icon />
                {option.label}
              </DropdownMenuRadioItem>
            )
          })}
        </DropdownMenuRadioGroup>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
