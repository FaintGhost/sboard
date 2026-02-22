import "./client";
import {
  listSyncJobs as _listSyncJobs,
  getSyncJob as _getSyncJob,
  retrySyncJob as _retrySyncJob,
} from "./gen";
import type { SyncJobListItem, SyncJobDetail } from "./gen";
import type { ListSyncJobsParams } from "./types";

export function listSyncJobs(params: ListSyncJobsParams = {}): Promise<SyncJobListItem[]> {
  return _listSyncJobs({ query: params as Record<string, unknown> }).then((r) => r.data!.data);
}

export function getSyncJob(id: number): Promise<SyncJobDetail> {
  return _getSyncJob({ path: { id } }).then((r) => r.data!.data);
}

export function retrySyncJob(id: number): Promise<SyncJobListItem> {
  return _retrySyncJob({ path: { id } }).then((r) => r.data!.data!);
}
