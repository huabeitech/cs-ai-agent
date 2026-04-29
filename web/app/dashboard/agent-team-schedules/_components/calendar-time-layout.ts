const dayMs = 24 * 60 * 60 * 1000

export type TimeLayoutSchedule = {
  id: number
  startAt: string
  endAt: string
}

export type TimeLayoutItem = {
  leftPercent: number
  widthPercent: number
  startLabel: string
  endLabel: string
}

export type DayTimeLayout = {
  rangeLabel: string
  items: Map<number, TimeLayoutItem>
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

function startOfDay(date: Date) {
  const ret = new Date(date)
  ret.setHours(0, 0, 0, 0)
  return ret
}

function formatTime(date: Date) {
  const hour = String(date.getHours()).padStart(2, "0")
  const minute = String(date.getMinutes()).padStart(2, "0")
  return `${hour}:${minute}`
}

function roundPercent(value: number) {
  return Math.round(value * 100) / 100
}

export function buildDayTimeLayout(schedules: TimeLayoutSchedule[], day: Date): DayTimeLayout {
  const dayStart = startOfDay(day)
  const dayEnd = new Date(dayStart.getTime() + dayMs)
  const visibleItems = schedules
    .map((item) => {
      const scheduleStart = parseLocalDateTime(item.startAt)
      const scheduleEnd = parseLocalDateTime(item.endAt)
      const visibleStart = new Date(Math.max(scheduleStart.getTime(), dayStart.getTime()))
      const visibleEnd = new Date(Math.min(scheduleEnd.getTime(), dayEnd.getTime()))
      return { item, visibleStart, visibleEnd }
    })
    .filter(({ visibleStart, visibleEnd }) => visibleEnd > visibleStart)

  if (visibleItems.length === 0) {
    return { rangeLabel: "", items: new Map() }
  }

  const rangeStart = new Date(Math.min(...visibleItems.map(({ visibleStart }) => visibleStart.getTime())))
  const rangeEnd = new Date(Math.max(...visibleItems.map(({ visibleEnd }) => visibleEnd.getTime())))
  const rangeMs = Math.max(rangeEnd.getTime() - rangeStart.getTime(), 1)
  const items = new Map<number, TimeLayoutItem>()

  visibleItems.forEach(({ item, visibleStart, visibleEnd }) => {
    items.set(item.id, {
      leftPercent: roundPercent(((visibleStart.getTime() - rangeStart.getTime()) / rangeMs) * 100),
      widthPercent: roundPercent(((visibleEnd.getTime() - visibleStart.getTime()) / rangeMs) * 100),
      startLabel: formatTime(visibleStart),
      endLabel: formatTime(visibleEnd),
    })
  })

  return {
    rangeLabel: `${formatTime(rangeStart)} - ${formatTime(rangeEnd)}`,
    items,
  }
}
