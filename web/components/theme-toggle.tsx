"use client"

import { useSyncExternalStore } from "react"
import { LaptopIcon, MoonIcon, SunIcon } from "lucide-react"
import { useTheme } from "next-themes"

import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"

type ThemeMode = "light" | "dark" | "system"

const themeOptions: Array<{
  value: ThemeMode
  label: string
  icon: typeof SunIcon
}> = [
  { value: "light", label: "浅色模式", icon: SunIcon },
  { value: "dark", label: "深色模式", icon: MoonIcon },
  { value: "system", label: "跟随系统", icon: LaptopIcon },
]

export function ThemeToggle() {
  const { theme, setTheme } = useTheme()
  const mounted = useSyncExternalStore(
    () => () => {},
    () => true,
    () => false
  )

  const activeTheme = mounted ? ((theme as ThemeMode | undefined) ?? "system") : "system"
  const ActiveIcon =
    themeOptions.find((option) => option.value === activeTheme)?.icon ?? LaptopIcon

  return (
    <DropdownMenu>
      <DropdownMenuTrigger render={<Button variant="outline" size="sm" />} aria-label="切换主题">
        <ActiveIcon />
        {/* <span className="hidden sm:inline">主题</span> */}
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-40 min-w-40">
        <DropdownMenuRadioGroup
          value={activeTheme}
          onValueChange={(value) => setTheme(value as ThemeMode)}
        >
          {themeOptions.map((option) => {
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
