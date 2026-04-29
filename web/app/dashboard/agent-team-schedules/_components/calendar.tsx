"use client"

import { CalendarPlusIcon, GripVerticalIcon } from "lucide-react"

import type {
  AdminAgentTeam,
  AdminAgentTeamSchedule,
  CreateAdminAgentTeamSchedulePayload,
  UpdateAdminAgentTeamSchedulePayload,
} from "@/lib/api/admin"
import { cn, formatDateTime } from "@/lib/utils"

const weekDayNames = ["一", "二", "三", "四", "五", "六", "日"]
const dayMs = 24 * 60 * 60 * 1000
const minuteMs = 60 * 1000
const minDurationMs = 15 * minuteMs

type ScheduleCalendarProps = {
  monthStart: Date
  calendarStart: Date
  calendarEnd: Date
  teams: AdminAgentTeam[]
  schedules: AdminAgentTeamSchedule[]
  loading: boolean
  savingId: number | null
  onCreate: (defaults: Partial<CreateAdminAgentTeamSchedulePayload>) => void
  onEdit: (item: AdminAgentTeamSchedule) => void
  onMove: (payload: UpdateAdminAgentTeamSchedulePayload) => Promise<void>
  onResize: (payload: UpdateAdminAgentTeamSchedulePayload) => Promise<void>
}

type DragState =
  | {
      type: "move"
      item: AdminAgentTeamSchedule
      startX: number
      startY: number
      moved: boolean
    }
  | {
      type: "resize"
      edge: "start" | "end"
      item: AdminAgentTeamSchedule
      moved: boolean
    }

function addDays(date: Date, days: number) {
  const ret = new Date(date)
  ret.setDate(ret.getDate() + days)
  return ret
}

function startOfDay(date: Date) {
  const ret = new Date(date)
  ret.setHours(0, 0, 0, 0)
  return ret
}

function parseLocalDateTime(value: string) {
  const match = value.match(/^(\d{4})-(\d{2})-(\d{2})[ T](\d{2}):(\d{2})(?::(\d{2}))?/)
  if (!match) {
    return new Date(value)
  }
  return new Date(
    Number(match[1]),
    Number(match[2]) - 1,
    Number(match[3]),
    Number(match[4]),
    Number(match[5]),
    Number(match[6] ?? 0)
  )
}

function formatDate(date: Date) {
  const month = String(date.getMonth() + 1).padStart(2, "0")
  const day = String(date.getDate()).padStart(2, "0")
  return `${date.getFullYear()}-${month}-${day}`
}

function formatDateTimeValue(date: Date) {
  const month = String(date.getMonth() + 1).padStart(2, "0")
  const day = String(date.getDate()).padStart(2, "0")
  const hour = String(date.getHours()).padStart(2, "0")
  const minute = String(date.getMinutes()).padStart(2, "0")
  const second = String(date.getSeconds()).padStart(2, "0")
  return `${date.getFullYear()}-${month}-${day} ${hour}:${minute}:${second}`
}

function clamp(value: number, min: number, max: number) {
  return Math.min(Math.max(value, min), max)
}

function roundToQuarterHour(date: Date) {
  const ret = new Date(date)
  ret.setSeconds(0, 0)
  const minutes = ret.getHours() * 60 + ret.getMinutes()
  const rounded = Math.round(minutes / 15) * 15
  ret.setHours(Math.floor(rounded / 60), rounded % 60, 0, 0)
  return ret
}

function getPointerDateInCell(event: PointerEvent, cell: Element) {
  const rect = cell.getBoundingClientRect()
  const day = startOfDay(parseLocalDateTime(`${cell.getAttribute("data-date")} 00:00:00`))
  const ratio = clamp((event.clientX - rect.left) / rect.width, 0, 1)
  return roundToQuarterHour(new Date(day.getTime() + ratio * dayMs))
}

function getDropCell(event: PointerEvent) {
  const element = document.elementFromPoint(event.clientX, event.clientY)
  return element?.closest("[data-schedule-cell]")
}

function buildMovePayload(item: AdminAgentTeamSchedule, date: string): UpdateAdminAgentTeamSchedulePayload {
  const originalStart = parseLocalDateTime(item.startAt)
  const originalEnd = parseLocalDateTime(item.endAt)
  const duration = originalEnd.getTime() - originalStart.getTime()
  const nextDay = startOfDay(parseLocalDateTime(`${date} 00:00:00`))
  const nextStart = new Date(nextDay)
  nextStart.setHours(originalStart.getHours(), originalStart.getMinutes(), originalStart.getSeconds(), 0)
  const nextEnd = new Date(nextStart.getTime() + duration)

  return {
    id: item.id,
    teamId: item.teamId,
    startAt: formatDateTimeValue(nextStart),
    endAt: formatDateTimeValue(nextEnd),
    sourceType: item.sourceType,
    remark: item.remark,
  }
}

function buildResizePayload(
  item: AdminAgentTeamSchedule,
  edge: "start" | "end",
  nextTime: Date
): UpdateAdminAgentTeamSchedulePayload | null {
  const startAt = parseLocalDateTime(item.startAt)
  const endAt = parseLocalDateTime(item.endAt)
  if (edge === "start") {
    if (endAt.getTime() - nextTime.getTime() < minDurationMs) {
      return null
    }
    startAt.setTime(nextTime.getTime())
  } else {
    if (nextTime.getTime() - startAt.getTime() < minDurationMs) {
      return null
    }
    endAt.setTime(nextTime.getTime())
  }
  return {
    id: item.id,
    teamId: item.teamId,
    startAt: formatDateTimeValue(startAt),
    endAt: formatDateTimeValue(endAt),
    sourceType: item.sourceType,
    remark: item.remark,
  }
}

function intersectsDay(item: AdminAgentTeamSchedule, day: Date) {
  const dayStart = startOfDay(day)
  const dayEnd = addDays(dayStart, 1)
  const scheduleStart = parseLocalDateTime(item.startAt)
  const scheduleEnd = parseLocalDateTime(item.endAt)
  return scheduleStart < dayEnd && scheduleEnd > dayStart
}

function buildCalendarDays(calendarStart: Date, calendarEnd: Date) {
  const days: Date[] = []
  for (let current = startOfDay(calendarStart); current < calendarEnd; current = addDays(current, 1)) {
    days.push(current)
  }
  return days
}

export function ScheduleCalendar({
  monthStart,
  calendarStart,
  calendarEnd,
  teams,
  schedules,
  loading,
  savingId,
  onCreate,
  onEdit,
  onMove,
  onResize,
}: ScheduleCalendarProps) {
  const days = buildCalendarDays(calendarStart, calendarEnd)
  const defaultTeamID = teams[0]?.id ?? 0

  function handleBlankCellClick(day: Date) {
    const startAt = new Date(day)
    startAt.setHours(9, 0, 0, 0)
    const endAt = new Date(day)
    endAt.setHours(18, 0, 0, 0)
    onCreate({
      teamId: defaultTeamID || undefined,
      startAt: formatDateTimeValue(startAt),
      endAt: formatDateTimeValue(endAt),
      sourceType: "manual",
      remark: "",
    })
  }

  function handlePointerDown(event: React.PointerEvent, item: AdminAgentTeamSchedule, type: DragState["type"], edge?: "start" | "end") {
    event.preventDefault()
    event.stopPropagation()
    const target = event.currentTarget as HTMLElement
    target.setPointerCapture(event.pointerId)
    const state: DragState =
      type === "resize"
        ? { type: "resize", edge: edge ?? "end", item, moved: false }
        : { type: "move", item, startX: event.clientX, startY: event.clientY, moved: false }

    function handlePointerMove(moveEvent: PointerEvent) {
      if (state.type === "move") {
        if (Math.abs(moveEvent.clientX - state.startX) > 4 || Math.abs(moveEvent.clientY - state.startY) > 4) {
          state.moved = true
        }
      } else {
        state.moved = true
      }
    }

    async function handlePointerUp(upEvent: PointerEvent) {
      target.releasePointerCapture(event.pointerId)
      window.removeEventListener("pointermove", handlePointerMove)
      window.removeEventListener("pointerup", handlePointerUp)
      if (!state.moved) {
        onEdit(item)
        return
      }
      const cell = getDropCell(upEvent)
      if (!cell) {
        return
      }
      if (state.type === "move") {
        const date = cell.getAttribute("data-date")
        if (!date) {
          return
        }
        await onMove(buildMovePayload(item, date))
        return
      }
      const payload = buildResizePayload(item, state.edge, getPointerDateInCell(upEvent, cell))
      if (payload) {
        await onResize(payload)
      }
    }

    window.addEventListener("pointermove", handlePointerMove)
    window.addEventListener("pointerup", handlePointerUp)
  }

  if (teams.length === 0 && !loading) {
    return (
      <div className="flex min-h-64 items-center justify-center rounded-lg border bg-background text-sm text-muted-foreground">
        暂无客服组，无法展示排班日历
      </div>
    )
  }

  return (
    <div className="min-w-[960px] overflow-hidden rounded-lg border bg-background">
      <div className="grid grid-cols-7 border-b bg-muted/40">
        {weekDayNames.map((name) => (
          <div key={name} className="flex h-10 items-center justify-center border-l first:border-l-0 text-sm font-medium">
            周{name}
          </div>
        ))}
      </div>
      <div className={cn("grid grid-cols-7", loading && "opacity-60")}>
        {days.map((day, dayIndex) => {
          const date = formatDate(day)
          const inMonth = day.getMonth() === monthStart.getMonth()
          const daySchedules = schedules.filter((item) => intersectsDay(item, day))
          return (
            <div
              key={date}
              data-schedule-cell
              data-date={date}
              role="button"
              tabIndex={0}
              className={cn(
                "min-h-36 border-l border-t bg-background p-2 text-left outline-none transition-colors first:border-l-0 hover:bg-muted/20 focus-visible:ring-2 focus-visible:ring-ring",
                dayIndex % 7 === 0 && "border-l-0",
                !inMonth && "bg-muted/20 text-muted-foreground"
              )}
              onClick={(event) => {
                if ((event.target as HTMLElement).closest("[data-schedule-block]")) {
                  return
                }
                handleBlankCellClick(day)
              }}
              onKeyDown={(event) => {
                if (event.key === "Enter" || event.key === " ") {
                  event.preventDefault()
                  handleBlankCellClick(day)
                }
              }}
            >
              <div className="mb-2 flex items-center justify-between gap-2">
                <div className={cn("text-sm font-medium", !inMonth && "text-muted-foreground")}>{day.getDate()}</div>
                <CalendarPlusIcon className="size-3.5 text-muted-foreground" />
              </div>
              <div className="space-y-1">
                {daySchedules.slice(0, 5).map((item) => {
                  const teamName = item.teamName || teams.find((team) => team.id === item.teamId)?.name || `客服组#${item.teamId}`
                  const busy = savingId === item.id
                  return (
                    <div
                      key={`${item.id}-${date}`}
                      data-schedule-block
                      role="button"
                      tabIndex={0}
                      className={cn(
                        "relative cursor-grab rounded-md border border-primary/20 bg-primary/10 px-2 py-1.5 pl-4 pr-4 text-primary shadow-sm outline-none active:cursor-grabbing",
                        busy && "pointer-events-none opacity-60"
                      )}
                      onPointerDown={(event) => handlePointerDown(event, item, "move")}
                      onKeyDown={(event) => {
                        if (event.key === "Enter" || event.key === " ") {
                          event.preventDefault()
                          onEdit(item)
                        }
                      }}
                    >
                      <div
                        className="absolute left-0 top-0 flex h-full w-3 cursor-ew-resize items-center justify-center bg-primary/15"
                        onPointerDown={(event) => handlePointerDown(event, item, "resize", "start")}
                      >
                        <GripVerticalIcon className="size-3" />
                      </div>
                      <div
                        className="absolute right-0 top-0 flex h-full w-3 cursor-ew-resize items-center justify-center bg-primary/15"
                        onPointerDown={(event) => handlePointerDown(event, item, "resize", "end")}
                      >
                        <GripVerticalIcon className="size-3" />
                      </div>
                      <div className="truncate text-xs font-medium">{teamName}</div>
                      <div className="truncate text-xs">
                        {formatDateTime(item.startAt).slice(11, 16)} - {formatDateTime(item.endAt).slice(11, 16)}
                      </div>
                      {item.remark ? <div className="truncate text-[11px] text-primary/80">{item.remark}</div> : null}
                    </div>
                  )
                })}
                {daySchedules.length > 5 ? (
                  <div className="text-xs text-muted-foreground">还有 {daySchedules.length - 5} 条</div>
                ) : null}
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}
