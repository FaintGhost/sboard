import { describe, expect, it } from "vitest"

import { formatDateTimeByTimezone, formatDateYMDByTimezone } from "./datetime"

describe("datetime timezone formatter", () => {
  it("formats datetime with configured timezone", () => {
    const got = formatDateTimeByTimezone("2026-02-08T08:00:00Z", "en-GB", "Asia/Shanghai")
    expect(got).toContain("16:00:00")
  })

  it("formats date-only string with configured timezone", () => {
    const got = formatDateYMDByTimezone("2026-02-08T23:30:00Z", "Asia/Shanghai")
    expect(got).toBe("2026-02-09")
  })

  it("falls back for invalid datetime", () => {
    const got = formatDateTimeByTimezone("invalid", "en-US", "UTC", "N/A")
    expect(got).toBe("N/A")
  })
})
