import { toApiError } from "@/lib/api/client";
import { i64, rpcCall } from "@/lib/rpc/client";
import {
  getTrafficNodesSummary as getTrafficNodesSummaryRPC,
  getTrafficTotalSummary as getTrafficTotalSummaryRPC,
  getTrafficTimeseries as getTrafficTimeseriesRPC,
} from "@/lib/rpc/gen/sboard/panel/v1/panel-TrafficService_connectquery";
import {
  toTrafficNodeSummary,
  toTrafficTimeseriesPoint,
  toTrafficTotalSummary,
} from "@/lib/rpc/mappers";
import type { TrafficNodeSummary, TrafficTotalSummary, TrafficTimeseriesPoint } from "./types";

export type { TrafficNodeSummary, TrafficTotalSummary, TrafficTimeseriesPoint };

export function listTrafficNodesSummary(
  params: { window?: string } = {},
): Promise<TrafficNodeSummary[]> {
  return rpcCall(getTrafficNodesSummaryRPC, { window: params.window })
    .then((r) => (r.data ?? []).map(toTrafficNodeSummary))
    .catch((e) => {
      throw toApiError(e);
    });
}

export function getTrafficTotalSummary(
  params: { window?: string } = {},
): Promise<TrafficTotalSummary> {
  return rpcCall(getTrafficTotalSummaryRPC, { window: params.window })
    .then((r) => toTrafficTotalSummary(r.data!))
    .catch((e) => {
      throw toApiError(e);
    });
}

export function listTrafficTimeseries(
  params: {
    window?: string;
    bucket?: "minute" | "hour" | "day";
    node_id?: number;
  } = {},
): Promise<TrafficTimeseriesPoint[]> {
  return rpcCall(getTrafficTimeseriesRPC, {
    window: params.window,
    bucket: params.bucket,
    nodeId: i64(params.node_id),
  })
    .then((r) => (r.data ?? []).map(toTrafficTimeseriesPoint))
    .catch((e) => {
      throw toApiError(e);
    });
}
