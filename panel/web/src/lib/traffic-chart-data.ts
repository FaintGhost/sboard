import type { TrafficTimeseriesPoint } from "@/lib/api/traffic";

export function resolveTrafficChartRows(
  queryRows: TrafficTimeseriesPoint[] | undefined,
  fallbackRows: TrafficTimeseriesPoint[],
): TrafficTimeseriesPoint[] {
  if (queryRows !== undefined) {
    return queryRows;
  }
  return fallbackRows;
}
