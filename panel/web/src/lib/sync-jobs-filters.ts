import type { SyncJobStatus } from "@/lib/api/types"

export const syncJobStatusValues: SyncJobStatus[] = [
  "queued",
  "running",
  "success",
  "failed",
]

export const syncJobSourceValues = [
  "manual_node_sync",
  "manual_retry",
  "auto_inbound_change",
  "auto_user_change",
  "auto_group_change",
] as const

export type SyncJobSource = (typeof syncJobSourceValues)[number]
export type SyncJobsTimeRange = "24h" | "7d" | "30d"

export type SyncJobsPageFilters = {
  nodeFilter: number | "all"
  statusFilter: SyncJobStatus | "all"
  sourceFilter: SyncJobSource | "all"
  timeRange: SyncJobsTimeRange
  page: number
}

const allowedRange = new Set<SyncJobsTimeRange>(["24h", "7d", "30d"])
const allowedStatus = new Set<SyncJobStatus>(syncJobStatusValues)
const allowedSource = new Set<SyncJobSource>(syncJobSourceValues)

export function parseSyncJobsSearchParams(params: URLSearchParams): SyncJobsPageFilters {
  const nodeIDRaw = params.get("node_id")?.trim() ?? ""
  const nodeIDNum = Number(nodeIDRaw)
  const nodeFilter = Number.isInteger(nodeIDNum) && nodeIDNum > 0 ? nodeIDNum : "all"

  const statusRaw = (params.get("status")?.trim() ?? "") as SyncJobStatus
  const statusFilter = allowedStatus.has(statusRaw) ? statusRaw : "all"

  const sourceRaw = (params.get("trigger_source")?.trim() ?? "") as SyncJobSource
  const sourceFilter = allowedSource.has(sourceRaw) ? sourceRaw : "all"

  const rangeRaw = (params.get("range")?.trim() ?? "") as SyncJobsTimeRange
  const timeRange = allowedRange.has(rangeRaw) ? rangeRaw : "24h"

  const pageRaw = Number(params.get("page") ?? "1")
  const page = Number.isInteger(pageRaw) && pageRaw > 0 ? pageRaw : 1

  return {
    nodeFilter,
    statusFilter,
    sourceFilter,
    timeRange,
    page,
  }
}

export function buildSyncJobsSearchParams(filters: SyncJobsPageFilters): URLSearchParams {
  const params = new URLSearchParams()
  if (typeof filters.nodeFilter === "number") params.set("node_id", String(filters.nodeFilter))
  if (filters.statusFilter !== "all") params.set("status", filters.statusFilter)
  if (filters.sourceFilter !== "all") params.set("trigger_source", filters.sourceFilter)
  if (filters.timeRange !== "24h") params.set("range", filters.timeRange)
  if (filters.page > 1) params.set("page", String(filters.page))
  return params
}
