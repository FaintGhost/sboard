import { describe, expect, it } from "vitest"

import type { TrafficTimeseriesPoint } from "@/lib/api/traffic"
import { resolveTrafficChartRows } from "@/lib/traffic-chart-data"

function point(at: string, upload: number, download: number): TrafficTimeseriesPoint {
  return {
    bucket_start: at,
    upload,
    download,
  }
}

describe("resolveTrafficChartRows", () => {
  it("falls back to previous rows while query rows are undefined", () => {
    const fallbackRows = [point("2026-02-10T00:00:00Z", 1, 2)]
    expect(resolveTrafficChartRows(undefined, fallbackRows)).toEqual(fallbackRows)
  })

  it("returns empty rows when query explicitly returns empty list", () => {
    const fallbackRows = [point("2026-02-10T00:00:00Z", 1, 2)]
    expect(resolveTrafficChartRows([], fallbackRows)).toEqual([])
  })

  it("uses fresh query rows when available", () => {
    const fallbackRows = [point("2026-02-10T00:00:00Z", 1, 2)]
    const queryRows = [point("2026-02-10T01:00:00Z", 3, 4)]
    expect(resolveTrafficChartRows(queryRows, fallbackRows)).toEqual(queryRows)
  })
})
