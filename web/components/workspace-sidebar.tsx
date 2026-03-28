"use client"

import Link from "next/link"
import { usePathname } from "next/navigation"

import {
  BotMessageSquareIcon,
  // UsersIcon,
  // BookOpenTextIcon,
  // SettingsIcon,
} from "lucide-react"

const agentNavItems = [
  {
    title: "会话",
    url: "/workspace/conversations",
    icon: <BotMessageSquareIcon className="size-4" />,
  },
  // {
  //   title: "客户",
  //   url: "/workspace/customers",
  //   icon: <UsersIcon className="size-4" />,
  // },
  // {
  //   title: "知识库",
  //   url: "/workspace/knowledge",
  //   icon: <BookOpenTextIcon className="size-4" />,
  // },
  // {
  //   title: "设置",
  //   url: "/workspace/settings",
  //   icon: <SettingsIcon className="size-4" />,
  // },
]

export function WorkspaceSidebar() {
  const pathname = usePathname()

  return (
    <nav className="hidden w-14 shrink-0 flex-col items-center border-r bg-background py-2 lg:flex">
      <div className="flex flex-col gap-1">
        {agentNavItems.map((item) => {
          const isActive = pathname === item.url || pathname.startsWith(item.url + "/")
          return (
            <Link
              key={item.url}
              href={item.url}
              className={`flex size-10 items-center justify-center rounded-md transition-colors ${
                isActive
                  ? "bg-primary/10 text-primary"
                  : "text-muted-foreground hover:bg-muted hover:text-foreground"
              }`}
              title={item.title}
            >
              {item.icon}
            </Link>
          )
        })}
      </div>
    </nav>
  )
}
