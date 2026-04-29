import assert from "node:assert/strict"
import test from "node:test"

import { buildDayTimeLayout } from "./calendar-time-layout.ts"

test("scales schedule bars by the visible span of that day", () => {
  const layout = buildDayTimeLayout(
    [
      { id: 1, startAt: "2026-04-29 07:00:00", endAt: "2026-04-29 17:00:00" },
      { id: 2, startAt: "2026-04-29 09:00:00", endAt: "2026-04-29 12:00:00" },
      { id: 3, startAt: "2026-04-29 13:00:00", endAt: "2026-04-29 17:00:00" },
    ],
    new Date(2026, 3, 29)
  )

  assert.equal(layout.rangeLabel, "07:00 - 17:00")
  assert.deepEqual(layout.items.get(1), {
    leftPercent: 0,
    widthPercent: 100,
    startLabel: "07:00",
    endLabel: "17:00",
  })
  assert.deepEqual(layout.items.get(2), {
    leftPercent: 20,
    widthPercent: 30,
    startLabel: "09:00",
    endLabel: "12:00",
  })
  assert.deepEqual(layout.items.get(3), {
    leftPercent: 60,
    widthPercent: 40,
    startLabel: "13:00",
    endLabel: "17:00",
  })
})

test("clips cross-day schedules to the current day before scaling", () => {
  const layout = buildDayTimeLayout(
    [
      { id: 1, startAt: "2026-04-28 22:00:00", endAt: "2026-04-29 08:00:00" },
      { id: 2, startAt: "2026-04-29 07:00:00", endAt: "2026-04-29 17:00:00" },
    ],
    new Date(2026, 3, 29)
  )

  assert.equal(layout.rangeLabel, "00:00 - 17:00")
  assert.deepEqual(layout.items.get(1), {
    leftPercent: 0,
    widthPercent: 47.06,
    startLabel: "00:00",
    endLabel: "08:00",
  })
  assert.deepEqual(layout.items.get(2), {
    leftPercent: 41.18,
    widthPercent: 58.82,
    startLabel: "07:00",
    endLabel: "17:00",
  })
})
