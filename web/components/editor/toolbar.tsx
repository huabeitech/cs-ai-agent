"use client"

import type { LucideIcon } from "lucide-react"

import { Separator } from "@/components/ui/separator"
import {
  ToggleGroup,
  ToggleGroupItem,
} from "@/components/ui/toggle-group"

type EditorToolbarButtonAction = {
  key: string
  label: string
  icon: LucideIcon
  onClick: () => void
  disabled?: boolean
  pressed?: boolean
}

type EditorToolbarSeparatorAction = {
  key: string
  type: "separator"
}

export type EditorToolbarAction = EditorToolbarButtonAction | EditorToolbarSeparatorAction

type EditorToolbarProps = {
  actions: EditorToolbarAction[]
}

export function EditorToolbar({ actions }: EditorToolbarProps) {
  return (
    <div className="flex items-center gap-1 border-b p-2">
      <ToggleGroup className="flex-wrap gap-1">
        {actions.map((action) => {
          if (action.type === "separator") {
            return <Separator key={action.key} orientation="vertical" className="mx-1 h-6" />
          }
          const Icon = action.icon
          return (
            <ToggleGroupItem
              key={action.key}
              value={action.key}
              aria-label={action.label}
              disabled={action.disabled}
              pressed={action.pressed}
              onClick={action.onClick}
            >
              <Icon className="size-4" />
            </ToggleGroupItem>
          )
        })}
      </ToggleGroup>
    </div>
  )
}

