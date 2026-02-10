import { describe, expect, it } from "vitest"

import { formatBytesWithUnit, pickByteUnit } from "@/lib/units"

describe("pickByteUnit", () => {
  it("returns B for zero and invalid values", () => {
    expect(pickByteUnit(0)).toBe("B")
    expect(pickByteUnit(Number.NaN)).toBe("B")
  })

  it("returns MB for megabyte-level values", () => {
    expect(pickByteUnit(5 * 1024 * 1024)).toBe("MB")
  })

  it("returns GB for gigabyte-level values", () => {
    expect(pickByteUnit(2 * 1024 * 1024 * 1024)).toBe("GB")
  })
})

describe("formatBytesWithUnit", () => {
  it("formats bytes in selected unit", () => {
    expect(formatBytesWithUnit(5 * 1024 * 1024, "MB")).toBe("5")
    expect(formatBytesWithUnit(1536, "KB")).toBe("1.5")
  })

  it("returns 0 for non-positive values", () => {
    expect(formatBytesWithUnit(0, "MB")).toBe("0")
    expect(formatBytesWithUnit(-100, "MB")).toBe("0")
  })
})
