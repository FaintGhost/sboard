import { toApiError } from "@/lib/api/client";
import { i64, rpcCall } from "@/lib/rpc/client";
import {
  getSyncJob as getSyncJobRPC,
  listSyncJobs as listSyncJobsRPC,
  retrySyncJob as retrySyncJobRPC,
} from "@/lib/rpc/gen/sboard/panel/v1/panel-SyncJobService_connectquery";
import type { ListSyncJobsParams, SyncJob, SyncJobDetail, SyncAttempt } from "./types";

function n64(v: bigint | number | null | undefined): number {
  if (v === null || v === undefined) return 0;
  return typeof v === "bigint" ? Number(v) : v;
}

function mapSyncJob(v: {
  id: bigint;
  nodeId: bigint;
  parentJobId?: bigint;
  triggerSource: string;
  status: string;
  inboundCount: number;
  activeUserCount: number;
  payloadHash: string;
  attemptCount: number;
  durationMs: bigint;
  errorSummary: string;
  createdAt: string;
  startedAt?: string;
  finishedAt?: string;
}): SyncJob {
  return {
    id: n64(v.id),
    node_id: n64(v.nodeId),
    parent_job_id: v.parentJobId === undefined ? null : n64(v.parentJobId),
    trigger_source: v.triggerSource,
    status: v.status,
    inbound_count: v.inboundCount,
    active_user_count: v.activeUserCount,
    payload_hash: v.payloadHash,
    attempt_count: v.attemptCount,
    duration_ms: n64(v.durationMs),
    error_summary: v.errorSummary,
    created_at: v.createdAt,
    started_at: v.startedAt ?? null,
    finished_at: v.finishedAt ?? null,
  };
}

function mapSyncAttempt(v: {
  id: bigint;
  attemptNo: number;
  status: string;
  httpStatus: number;
  durationMs: bigint;
  errorSummary: string;
  backoffMs: bigint;
  startedAt: string;
  finishedAt?: string;
}): SyncAttempt {
  return {
    id: n64(v.id),
    attempt_no: v.attemptNo,
    status: v.status,
    http_status: v.httpStatus,
    duration_ms: n64(v.durationMs),
    error_summary: v.errorSummary,
    backoff_ms: n64(v.backoffMs),
    started_at: v.startedAt,
    finished_at: v.finishedAt ?? null,
  };
}

export function listSyncJobs(params: ListSyncJobsParams = {}): Promise<SyncJob[]> {
  return rpcCall(listSyncJobsRPC, {
    limit: params.limit,
    offset: params.offset,
    nodeId: i64(params.node_id),
    status: params.status,
    triggerSource: params.trigger_source,
    from: params.from,
    to: params.to,
  })
    .then((r) => (r.data ?? []).map(mapSyncJob))
    .catch((e) => {
      throw toApiError(e);
    });
}

export function getSyncJob(id: number): Promise<SyncJobDetail> {
  return rpcCall(getSyncJobRPC, { id: BigInt(id) })
    .then((r) => {
      if (!r.data || !r.data.job) {
        throw new Error("sync job detail is empty");
      }
      return {
        job: mapSyncJob(r.data.job),
        attempts: (r.data.attempts ?? []).map(mapSyncAttempt),
      };
    })
    .catch((e) => {
      throw toApiError(e);
    });
}

export function retrySyncJob(id: number): Promise<SyncJob> {
  return rpcCall(retrySyncJobRPC, { id: BigInt(id) })
    .then((r) => {
      if (!r.data) {
        throw new Error(r.status || "retry succeeded without job payload");
      }
      return mapSyncJob(r.data);
    })
    .catch((e) => {
      throw toApiError(e);
    });
}
