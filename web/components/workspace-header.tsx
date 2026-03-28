"use client"

import Link from "next/link"

import { WorkspaceToggle } from "@/components/workspace-toggle"
import {
  Menubar,
  MenubarContent,
  MenubarItem,
  MenubarMenu,
  MenubarSeparator,
  MenubarShortcut,
  MenubarTrigger,
} from "@/components/ui/menubar"
import { MessageSquareIcon, SettingsIcon, LogOutIcon } from "lucide-react"
import { useAuth } from "@/components/auth-provider"

export function WorkspaceHeader() {
  const { signOut } = useAuth()

  return (
    <header className="flex h-12 shrink-0 items-center border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="flex w-full items-center justify-between px-4">
        <div className="flex items-center gap-1">
          <Link href="/workspace" className="flex items-center gap-2 px-2">
            <MessageSquareIcon className="size-4" />
            <span className="text-sm font-medium">客服工作台</span>
          </Link>
          <Menubar className="border-none bg-transparent shadow-none">
            <MenubarMenu>
              <MenubarTrigger className="text-sm">会话</MenubarTrigger>
            </MenubarMenu>
            <MenubarMenu>
              <MenubarTrigger className="text-sm">设置</MenubarTrigger>
              <MenubarContent>
                <MenubarItem>
                  <SettingsIcon className="mr-2 size-4" />
                  个人设置
                </MenubarItem>
                <MenubarSeparator />
                <MenubarItem
                  onClick={() => {
                    void signOut()
                  }}
                >
                  <LogOutIcon className="mr-2 size-4" />
                  退出登录
                </MenubarItem>
              </MenubarContent>
            </MenubarMenu>
          </Menubar>
        </div>
        <div className="flex items-center gap-2">
          <WorkspaceToggle />
        </div>
      </div>
    </header>
  )
}
