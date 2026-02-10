import { apiRequest } from "./client";

export type TrafficNodeSummary = {
  node_id: number;
  upload: number;
  download: number;
  last_recorded_at: string;
  samples: number;
  inbounds: number;
};

export type TrafficTotalSummary = {
  upload: number;
  download: number;
  last_recorded_at: string;
  samples: number;
  nodes: number;
  inbounds: number;
};

export type TrafficTimeseriesPoint = {
  bucket_start: string;
  upload: number;
  download: number;
};

export function listTrafficNodesSummary(params: { window?: string } = {}) {
  const qs = new URLSearchParams();
  if (params.window) qs.set("window", params.window);
  const suffix = qs.toString() ? `?${qs.toString()}` : "";
  return apiRequest<TrafficNodeSummary[]>(`/api/traffic/nodes/summary${suffix}`);
}

export function getTrafficTotalSummary(params: { window?: string } = {}) {
  const qs = new URLSearchParams();
  if (params.window) qs.set("window", params.window);
  const suffix = qs.toString() ? `?${qs.toString()}` : "";
  return apiRequest<TrafficTotalSummary>(`/api/traffic/total/summary${suffix}`);
}

export function listTrafficTimeseries(
  params: {
    window?: string;
    bucket?: "minute" | "hour" | "day";
    node_id?: number;
  } = {},
) {
  const qs = new URLSearchParams();
  if (params.window) qs.set("window", params.window);
  if (params.bucket) qs.set("bucket", params.bucket);
  if (typeof params.node_id === "number") qs.set("node_id", String(params.node_id));
  const suffix = qs.toString() ? `?${qs.toString()}` : "";
  return apiRequest<TrafficTimeseriesPoint[]>(`/api/traffic/timeseries${suffix}`);
}
