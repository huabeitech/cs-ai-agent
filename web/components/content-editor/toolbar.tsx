"use client"

import "./toolbar.css"

import { useRef, type PointerEvent as ReactPointerEvent } from "react"

import { cn } from "@/lib/utils"

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
  const containerRef = useRef<HTMLDivElement>(null)
  const dragStateRef = useRef<{
    pointerId: number
    startX: number
    startScrollLeft: number
    moved: boolean
  } | null>(null)

  const handlePointerDown = (event: ReactPointerEvent<HTMLDivElement>) => {
    if (event.pointerType === "mouse" && event.button !== 0) {
      return
    }
    const container = containerRef.current
    if (!container || container.scrollWidth <= container.clientWidth) {
      return
    }
    dragStateRef.current = {
      pointerId: event.pointerId,
      startX: event.clientX,
      startScrollLeft: container.scrollLeft,
      moved: false,
    }
    container.setPointerCapture(event.pointerId)
  }

  const handlePointerMove = (event: ReactPointerEvent<HTMLDivElement>) => {
    const container = containerRef.current
    const dragState = dragStateRef.current
    if (!container || !dragState || dragState.pointerId !== event.pointerId) {
      return
    }
    const deltaX = event.clientX - dragState.startX
    if (Math.abs(deltaX) > 3) {
      dragState.moved = true
    }
    container.scrollLeft = dragState.startScrollLeft - deltaX
  }

  const handlePointerEnd = (event: ReactPointerEvent<HTMLDivElement>) => {
    const container = containerRef.current
    const dragState = dragStateRef.current
    if (!container || !dragState || dragState.pointerId !== event.pointerId) {
      return
    }
    if (container.hasPointerCapture(event.pointerId)) {
      container.releasePointerCapture(event.pointerId)
    }
    window.setTimeout(() => {
      dragStateRef.current = null
    }, 0)
  }

  return (
    <div
      ref={containerRef}
      className="content-editor-toolbar overflow-x-auto overflow-y-hidden border-b border-border px-1 py-1"
      onPointerDown={handlePointerDown}
      onPointerMove={handlePointerMove}
      onPointerUp={handlePointerEnd}
      onPointerCancel={handlePointerEnd}
    >
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
                onClickCapture={(event) => {
                  if (dragStateRef.current?.moved) {
                    event.preventDefault()
                    event.stopPropagation()
                  }
                }}
                data-pressed={action.pressed ? "true" : "false"}
                className={cn(
                  "mx-[2px] cursor-pointer list-none rounded-[3px] border-none bg-transparent text-[#3f4a54] transition-all duration-300 select-none hover:bg-[#f2f2f2] disabled:cursor-not-allowed disabled:opacity-60 data-[pressed=true]:bg-[#f2f2f2] dark:text-[#999] dark:hover:bg-[#333] dark:data-[pressed=true]:bg-[#333]",
                  action.icon && !action.content
                    ? "flex flex-col items-center px-[2px] py-0"
                    : "flex flex-col items-center px-[6px] py-0"
                )}
              >
                {Icon ? <Icon className="box-content size-4 p-1" /> : null}
                {action.content ? (
                  <span className="text-[12px] leading-none whitespace-nowrap">
                    {action.content}
                  </span>
                ) : null}
              </button>
            )
          })}
        </div>
      </div>
    </div>
  )
}
