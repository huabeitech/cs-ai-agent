import assert from "node:assert/strict"
import test from "node:test"

import { addDays, formatWeekTitle, isSameLocalDay, startOfWeek } from "./calendar-date-range.ts"

test("builds Monday-based week range and title", () => {
  const start = startOfWeek(new Date(2026, 3, 29, 14, 0, 0))

  assert.equal(start.getFullYear(), 2026)
  assert.equal(start.getMonth(), 3)
  assert.equal(start.getDate(), 27)
  assert.equal(formatWeekTitle(start), "2026-04-27 - 2026-05-03")
  assert.equal(addDays(start, 7).getDate(), 4)
})

test("compares local calendar dates", () => {
  assert.equal(isSameLocalDay(new Date(2026, 3, 29, 0, 0), new Date(2026, 3, 29, 23, 59)), true)
  assert.equal(isSameLocalDay(new Date(2026, 3, 29, 23, 59), new Date(2026, 3, 30, 0, 0)), false)
})
