import { describe, expect, it } from "vitest"

import {
  buildSyncJobsSearchParams,
  parseSyncJobsSearchParams,
  type SyncJobsPageFilters,
} from "./sync-jobs-filters"

describe("sync-jobs filters", () => {
  it("parses valid query params", () => {
    const filters = parseSyncJobsSearchParams(
      new URLSearchParams(
        "node_id=12&status=failed&trigger_source=manual_retry&range=7d&page=3",
      ),
    )

    expect(filters).toEqual<SyncJobsPageFilters>({
      nodeFilter: 12,
      statusFilter: "failed",
      sourceFilter: "manual_retry",
      timeRange: "7d",
      page: 3,
    })
  })

  it("falls back to defaults for invalid query params", () => {
    const filters = parseSyncJobsSearchParams(
      new URLSearchParams(
        "node_id=oops&status=unknown&trigger_source=&range=1y&page=0",
      ),
    )

    expect(filters).toEqual<SyncJobsPageFilters>({
      nodeFilter: "all",
      statusFilter: "all",
      sourceFilter: "all",
      timeRange: "24h",
      page: 1,
    })
  })

  it("builds compact query params", () => {
    const params = buildSyncJobsSearchParams({
      nodeFilter: 9,
      statusFilter: "failed",
      sourceFilter: "manual_retry",
      timeRange: "7d",
      page: 2,
    })

    expect(params.toString()).toBe(
      "node_id=9&status=failed&trigger_source=manual_retry&range=7d&page=2",
    )
  })

  it("omits default values when building query params", () => {
    const params = buildSyncJobsSearchParams({
      nodeFilter: "all",
      statusFilter: "all",
      sourceFilter: "all",
      timeRange: "24h",
      page: 1,
    })

    expect(params.toString()).toBe("")
  })
})
