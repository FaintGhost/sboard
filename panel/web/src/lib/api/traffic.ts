import "./client";
import {
  getTrafficNodesSummary as _getTrafficNodesSummary,
  getTrafficTotalSummary as _getTrafficTotalSummary,
  getTrafficTimeseries as _getTrafficTimeseries,
} from "./gen";
import type { TrafficNodeSummary, TrafficTotalSummary, TrafficTimeseriesPoint } from "./gen";

export type { TrafficNodeSummary, TrafficTotalSummary, TrafficTimeseriesPoint };

export function listTrafficNodesSummary(
  params: { window?: string } = {},
): Promise<TrafficNodeSummary[]> {
  return _getTrafficNodesSummary({ query: params }).then((r) => r.data!.data);
}

export function getTrafficTotalSummary(
  params: { window?: string } = {},
): Promise<TrafficTotalSummary> {
  return _getTrafficTotalSummary({ query: params }).then((r) => r.data!.data);
}

export function listTrafficTimeseries(
  params: {
    window?: string;
    bucket?: "minute" | "hour" | "day";
    node_id?: number;
  } = {},
): Promise<TrafficTimeseriesPoint[]> {
  return _getTrafficTimeseries({ query: params }).then((r) => r.data!.data);
}
