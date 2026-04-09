"use client"

import { cn } from "@/lib/utils"

import type { ContentMode } from "./types"

type EditorModeSwitchProps = {
  value: ContentMode
  disabled?: boolean
  onChange: (nextMode: ContentMode) => void
}

const MODE_OPTIONS: Array<{ value: ContentMode; label: string }> = [
  { value: "markdown", label: "Markdown" },
  { value: "html", label: "HTML" },
]

export function EditorModeSwitch({
  value,
  disabled = false,
  onChange,
}: EditorModeSwitchProps) {
  return (
    <div className="mx-0.5 rounded-[3px] border border-border/80 bg-transparent p-0">
      <div className="flex items-center">
        {MODE_OPTIONS.map((option) => {
          const active = option.value === value
          return (
            <button
              key={option.value}
              type="button"
              disabled={disabled}
              onClick={() => onChange(option.value)}
              className={cn(
                "px-1.5 py-0 text-[12px] leading-6 whitespace-nowrap transition-colors",
                "hover:bg-[#f2f2f2] disabled:cursor-not-allowed disabled:opacity-60 dark:hover:bg-[#333]",
                active
                  ? "bg-[#f2f2f2] text-[#3f4a54] dark:bg-[#333] dark:text-[#999]"
                  : "bg-transparent text-[#3f4a54] dark:text-[#999]"
              )}
            >
              {option.label}
            </button>
          )
        })}
      </div>
    </div>
  )
}
