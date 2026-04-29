export function startOfDay(date: Date) {
  const ret = new Date(date)
  ret.setHours(0, 0, 0, 0)
  return ret
}

export function startOfWeek(date: Date) {
  const ret = startOfDay(date)
  const day = ret.getDay()
  const offset = day === 0 ? -6 : 1 - day
  ret.setDate(ret.getDate() + offset)
  return ret
}

export function isSameLocalDay(a: Date, b: Date) {
  return a.getFullYear() === b.getFullYear() && a.getMonth() === b.getMonth() && a.getDate() === b.getDate()
}

export function startOfMonth(date: Date) {
  const ret = startOfDay(date)
  ret.setDate(1)
  return ret
}

export function startOfMonthCalendar(date: Date) {
  return startOfWeek(startOfMonth(date))
}

export function endOfMonthCalendar(date: Date) {
  const monthEnd = startOfMonth(date)
  monthEnd.setMonth(monthEnd.getMonth() + 1)
  const ret = startOfWeek(monthEnd)
  if (ret.getTime() < monthEnd.getTime()) {
    ret.setDate(ret.getDate() + 7)
  }
  return ret
}

export function addDays(date: Date, days: number) {
  const ret = new Date(date)
  ret.setDate(ret.getDate() + days)
  return ret
}

export function addMonths(date: Date, months: number) {
  const ret = startOfMonth(date)
  ret.setMonth(ret.getMonth() + months)
  return ret
}

export function formatDateTimeValue(date: Date) {
  const month = String(date.getMonth() + 1).padStart(2, "0")
  const day = String(date.getDate()).padStart(2, "0")
  const hour = String(date.getHours()).padStart(2, "0")
  const minute = String(date.getMinutes()).padStart(2, "0")
  const second = String(date.getSeconds()).padStart(2, "0")
  return `${date.getFullYear()}-${month}-${day} ${hour}:${minute}:${second}`
}

export function formatMonthTitle(monthStart: Date) {
  return `${monthStart.getFullYear()}年${String(monthStart.getMonth() + 1).padStart(2, "0")}月`
}

function formatDate(date: Date) {
  const month = String(date.getMonth() + 1).padStart(2, "0")
  const day = String(date.getDate()).padStart(2, "0")
  return `${date.getFullYear()}-${month}-${day}`
}

export function formatWeekTitle(weekStart: Date) {
  return `${formatDate(weekStart)} - ${formatDate(addDays(weekStart, 6))}`
}
