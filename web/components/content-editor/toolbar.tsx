"use client"

import { Separator } from "@/components/ui/separator"
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group"

import type { EditorToolbarAction } from "./types"

type EditorToolbarProps = {
  actions: ReadonlyArray<EditorToolbarAction>
}

function isSeparatorAction(
  action: EditorToolbarAction
): action is Extract<EditorToolbarAction, { type: "separator" }> {
  return "type" in action && action.type === "separator"
}

export function EditorToolbar({ actions }: EditorToolbarProps) {
  return (
    <div className="border-b bg-[#f7f7f7] px-2 py-1 dark:bg-muted/25">
      <ToggleGroup className="flex flex-wrap items-center justify-start gap-1">
        {actions.map((action) => {
          if (isSeparatorAction(action)) {
            return (
              <Separator
                key={action.key}
                orientation="vertical"
                className="mx-1 h-5 bg-border/80"
              />
            )
          }

          const Icon = action.icon
          return (
            <ToggleGroupItem
              key={action.key}
              value={action.key}
              aria-label={action.label}
              title={action.label}
              disabled={action.disabled}
              pressed={action.pressed}
              onClick={action.onClick}
              className="size-7 rounded-sm border border-transparent bg-transparent text-[#555] shadow-none hover:border-[#d9d9d9] hover:bg-white hover:text-foreground data-[state=on]:border-[#d9d9d9] data-[state=on]:bg-white data-[state=on]:text-foreground dark:text-muted-foreground dark:hover:border-input dark:hover:bg-background dark:data-[state=on]:border-input dark:data-[state=on]:bg-background"
            >
              <Icon className="size-4" />
            </ToggleGroupItem>
          )
        })}
      </ToggleGroup>
    </div>
  )
}
