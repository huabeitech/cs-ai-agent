"use client"

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
    <div className="overflow-x-auto overflow-y-hidden border-b border-border px-1 py-1 scrollbar-none">
      <div className="flex min-w-max items-center justify-between">
        <div className="flex items-center">
        {actions.map((action) => {
          if (isSeparatorAction(action)) {
            return (
              <span
                key={action.key}
                aria-hidden="true"
                className="relative mx-2 inline-block h-[0.9em] w-px self-center bg-border"
              />
            )
          }

          const Icon = action.icon
          return (
            <button
              key={action.key}
              type="button"
              aria-label={action.label}
              title={action.label}
              disabled={action.disabled}
              onClick={action.onClick}
              data-pressed={action.pressed ? "true" : "false"}
              className="mx-[2px] flex cursor-pointer list-none flex-col items-center rounded-[3px] border-none bg-transparent px-[2px] py-0 text-[#3f4a54] transition-all duration-300 select-none hover:bg-[#f2f2f2] disabled:cursor-not-allowed disabled:opacity-60 data-[pressed=true]:bg-[#f2f2f2] dark:text-[#999] dark:hover:bg-[#333] dark:data-[pressed=true]:bg-[#333]"
            >
              <Icon className="box-content size-4 p-1" />
            </button>
          )
        })}
        </div>
      </div>
    </div>
  )
}
