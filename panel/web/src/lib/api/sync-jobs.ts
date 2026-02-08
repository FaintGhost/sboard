import { apiRequest } from "./client"
import type { ListSyncJobsParams, SyncJob, SyncJobDetail } from "./types"

export function listSyncJobs(params: ListSyncJobsParams = {}) {
  const query = new URLSearchParams()

  if (typeof params.limit === "number") query.set("limit", String(params.limit))
  if (typeof params.offset === "number") query.set("offset", String(params.offset))
  if (typeof params.node_id === "number") query.set("node_id", String(params.node_id))
  if (params.status) query.set("status", params.status)
  if (params.trigger_source) query.set("trigger_source", params.trigger_source)
  if (params.from) query.set("from", params.from)
  if (params.to) query.set("to", params.to)

  const suffix = query.toString() ? `?${query.toString()}` : ""
  return apiRequest<SyncJob[]>(`/api/sync-jobs${suffix}`)
}

export function getSyncJob(id: number) {
  return apiRequest<SyncJobDetail>(`/api/sync-jobs/${id}`)
}

export function retrySyncJob(id: number) {
  return apiRequest<SyncJob>(`/api/sync-jobs/${id}/retry`, { method: "POST" })
}
