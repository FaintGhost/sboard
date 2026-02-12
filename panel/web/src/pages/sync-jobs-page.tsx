import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { useSearchParams } from "react-router-dom";

import { AsyncButton } from "@/components/ui/async-button";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { PageHeader } from "@/components/page-header";
import { TableEmptyState } from "@/components/table-empty-state";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Badge } from "@/components/ui/badge";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { ApiError } from "@/lib/api/client";
import { listNodes } from "@/lib/api/nodes";
import { getSyncJob, listSyncJobs, retrySyncJob } from "@/lib/api/sync-jobs";
import {
  buildSyncJobsSearchParams,
  parseSyncJobsSearchParams,
  syncJobSourceValues,
  syncJobStatusValues,
  type SyncJobsPageFilters,
  type SyncJobsTimeRange,
} from "@/lib/sync-jobs-filters";
import { tableColumnLayout, tableColumnSpacing } from "@/lib/table-spacing";
import { tableTransitionClass } from "@/lib/table-motion";
import { useTableQueryTransition } from "@/lib/table-query-transition";
import { tableToolbarClass } from "@/lib/table-toolbar";
import { formatDateTimeByTimezone } from "@/lib/datetime";
import { useSystemStore } from "@/store/system";

const pageSize = 20;

function formatDateTime(value: string | undefined, locale: string, timezone: string): string {
  return formatDateTimeByTimezone(value, locale, timezone);
}

function formatDuration(ms: number): string {
  if (!Number.isFinite(ms) || ms <= 0) return "-";
  if (ms < 1000) return `${Math.round(ms)} ms`;

  const seconds = ms / 1000;
  if (seconds < 60) {
    const fixed = seconds >= 10 ? 1 : 2;
    return `${seconds.toFixed(fixed)} s`;
  }

  const mins = Math.floor(seconds / 60);
  const secs = Math.round(seconds % 60);
  return `${mins}m ${secs}s`;
}

function shortHash(value: string): string {
  if (!value) return "-";
  if (value.length <= 20) return value;
  return `${value.slice(0, 10)}...${value.slice(-6)}`;
}

function getStatusVariant(status: string): "default" | "secondary" | "destructive" | "outline" {
  switch (status) {
    case "success":
      return "default";
    case "failed":
      return "destructive";
    case "running":
      return "secondary";
    default:
      return "outline";
  }
}

function fromISOByRange(range: SyncJobsTimeRange): string {
  const now = Date.now();
  const delta =
    range === "24h"
      ? 24 * 60 * 60 * 1000
      : range === "7d"
        ? 7 * 24 * 60 * 60 * 1000
        : 30 * 24 * 60 * 60 * 1000;
  return new Date(now - delta).toISOString();
}

function sourceLabel(t: (key: string) => string, value: string): string {
  const key = `syncJobs.source.${value}`;
  const text = t(key);
  return text === key ? value : text;
}

function statusLabel(t: (key: string) => string, value: string): string {
  const key = `syncJobs.status.${value}`;
  const text = t(key);
  return text === key ? value : text;
}

export function SyncJobsPage() {
  const { t, i18n } = useTranslation();
  const timezone = useSystemStore((state) => state.timezone);
  const qc = useQueryClient();
  const spacing = tableColumnSpacing.seven;
  const layout = tableColumnLayout.sevenActionButton;
  const [searchParams, setSearchParams] = useSearchParams();

  const filters = useMemo(() => parseSyncJobsSearchParams(searchParams), [searchParams]);
  const currentOffset = (filters.page - 1) * pageSize;
  const filtersKey = useMemo(() => JSON.stringify(filters), [filters]);

  const [selectedJobID, setSelectedJobID] = useState<number | null>(null);

  function updateFilters(patch: Partial<SyncJobsPageFilters>, resetPage = false) {
    const next: SyncJobsPageFilters = {
      ...filters,
      ...patch,
    };
    if (resetPage) {
      next.page = 1;
    }
    setSearchParams(buildSyncJobsSearchParams(next), { replace: true });
  }

  const nodesQuery = useQuery({
    queryKey: ["nodes", "sync-jobs", { limit: 100, offset: 0 }],
    queryFn: () => listNodes({ limit: 100, offset: 0 }),
  });

  const jobsQuery = useQuery({
    queryKey: ["sync-jobs", filters],
    queryFn: () =>
      listSyncJobs({
        limit: pageSize,
        offset: currentOffset,
        node_id: filters.nodeFilter === "all" ? undefined : filters.nodeFilter,
        status: filters.statusFilter === "all" ? undefined : filters.statusFilter,
        trigger_source: filters.sourceFilter === "all" ? undefined : filters.sourceFilter,
        from: fromISOByRange(filters.timeRange),
        to: new Date().toISOString(),
      }),
    refetchInterval: 15_000,
  });

  const jobsTable = useTableQueryTransition({
    filterKey: filtersKey,
    rows: jobsQuery.data,
    isLoading: jobsQuery.isLoading,
    isFetching: jobsQuery.isFetching,
    isError: jobsQuery.isError,
  });

  const detailQuery = useQuery({
    queryKey: ["sync-job-detail", selectedJobID],
    queryFn: () => getSyncJob(selectedJobID ?? 0),
    enabled: selectedJobID != null,
  });

  const nodeNameByID = useMemo(() => {
    const map = new Map<number, string>();
    for (const item of nodesQuery.data ?? []) {
      map.set(item.id, item.name);
    }
    return map;
  }, [nodesQuery.data]);

  const selectedFromList = useMemo(
    () => (jobsQuery.data ?? []).find((item) => item.id === selectedJobID) ?? null,
    [jobsQuery.data, selectedJobID],
  );

  const retryMutation = useMutation({
    mutationFn: retrySyncJob,
    onSuccess: async (created) => {
      setSelectedJobID(created.id);
      await qc.invalidateQueries({ queryKey: ["sync-jobs"] });
      await qc.invalidateQueries({ queryKey: ["sync-job-detail"] });
    },
  });

  function nodeName(nodeID: number): string {
    return nodeNameByID.get(nodeID) ?? `#${nodeID}`;
  }

  const detailJob = detailQuery.data?.job ?? selectedFromList;
  const detailAttempts = detailQuery.data?.attempts ?? [];
  const visibleJobs = jobsTable.visibleRows;

  return (
    <div className="px-4 lg:px-6">
      <section className="space-y-6">
        <PageHeader title={t("syncJobs.title")} description={t("syncJobs.subtitle")} />

        <Card>
          <CardHeader className="pb-3">
            <div className={tableToolbarClass.container}>
              <div className="flex flex-col gap-1.5">
                <CardTitle className="text-base">{t("syncJobs.list")}</CardTitle>
                <CardDescription>
                  {jobsTable.showLoadingHint ? t("common.loading") : null}
                  {jobsQuery.isError ? t("common.loadFailed") : null}
                  {!jobsTable.showLoadingHint && jobsQuery.data
                    ? t("syncJobs.count", { count: visibleJobs.length })
                    : null}
                </CardDescription>
              </div>
              <div className={tableToolbarClass.filters}>
                <Select
                  value={filters.timeRange}
                  onValueChange={(v) => updateFilters({ timeRange: v as SyncJobsTimeRange }, true)}
                >
                  <SelectTrigger className="w-full sm:w-44" aria-label={t("syncJobs.timeRange")}>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="24h">{t("syncJobs.range24h")}</SelectItem>
                    <SelectItem value="7d">{t("syncJobs.range7d")}</SelectItem>
                    <SelectItem value="30d">{t("syncJobs.range30d")}</SelectItem>
                  </SelectContent>
                </Select>

                <Select
                  value={filters.nodeFilter === "all" ? "all" : String(filters.nodeFilter)}
                  onValueChange={(v) =>
                    updateFilters({ nodeFilter: v === "all" ? "all" : Number(v) }, true)
                  }
                >
                  <SelectTrigger className="w-full sm:w-52" aria-label={t("syncJobs.nodeFilter")}>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">{t("syncJobs.allNodes")}</SelectItem>
                    {nodesQuery.data?.map((node) => (
                      <SelectItem key={node.id} value={String(node.id)}>
                        {node.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>

                <Select
                  value={filters.statusFilter}
                  onValueChange={(v) =>
                    updateFilters({ statusFilter: v as SyncJobsPageFilters["statusFilter"] }, true)
                  }
                >
                  <SelectTrigger className="w-full sm:w-44" aria-label={t("syncJobs.statusFilter")}>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">{t("syncJobs.allStatus")}</SelectItem>
                    {syncJobStatusValues.map((status) => (
                      <SelectItem key={status} value={status}>
                        {statusLabel(t, status)}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>

                <Select
                  value={filters.sourceFilter}
                  onValueChange={(v) =>
                    updateFilters({ sourceFilter: v as SyncJobsPageFilters["sourceFilter"] }, true)
                  }
                >
                  <SelectTrigger className="w-full sm:w-52" aria-label={t("syncJobs.sourceFilter")}>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">{t("syncJobs.allSources")}</SelectItem>
                    {syncJobSourceValues.map((source) => (
                      <SelectItem key={source} value={source}>
                        {sourceLabel(t, source)}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>
          </CardHeader>

          <CardContent className="p-0">
            <Table className={layout.tableClass}>
              <TableHeader>
                <TableRow>
                  <TableHead className={`${spacing.headFirst} ${layout.headFirst}`}>
                    {t("syncJobs.colTime")}
                  </TableHead>
                  <TableHead className={`${spacing.headMiddle} ${layout.headMiddle}`}>
                    {t("syncJobs.colNode")}
                  </TableHead>
                  <TableHead className={`${spacing.headMiddle} ${layout.headMiddle}`}>
                    {t("syncJobs.colSource")}
                  </TableHead>
                  <TableHead className={`${spacing.headMiddle} ${layout.headMiddle}`}>
                    {t("syncJobs.colStatus")}
                  </TableHead>
                  <TableHead className={`${spacing.headMiddle} ${layout.headMiddle}`}>
                    {t("syncJobs.colDuration")}
                  </TableHead>
                  <TableHead className={`${spacing.headMiddle} ${layout.headMiddle}`}>
                    {t("syncJobs.colRetries")}
                  </TableHead>
                  <TableHead className={`${spacing.headLast} ${layout.headLast}`}>
                    {t("common.actions")}
                  </TableHead>
                </TableRow>
              </TableHeader>
              <TableBody className={tableTransitionClass(jobsTable.isTransitioning)}>
                {visibleJobs.map((job) => (
                  <TableRow key={job.id}>
                    <TableCell className={spacing.cellFirst}>
                      {formatDateTime(job.created_at, i18n.language, timezone)}
                    </TableCell>
                    <TableCell className={`${spacing.cellMiddle} font-medium`}>
                      {nodeName(job.node_id)}
                    </TableCell>
                    <TableCell className={`${spacing.cellMiddle} text-muted-foreground`}>
                      {sourceLabel(t, job.trigger_source)}
                    </TableCell>
                    <TableCell className={spacing.cellMiddle}>
                      <Badge variant={getStatusVariant(job.status)}>
                        {statusLabel(t, job.status)}
                      </Badge>
                    </TableCell>
                    <TableCell className={`${spacing.cellMiddle} tabular-nums`}>
                      {formatDuration(job.duration_ms)}
                    </TableCell>
                    <TableCell className={`${spacing.cellMiddle} tabular-nums`}>
                      {job.attempt_count}
                    </TableCell>
                    <TableCell className={spacing.cellLast}>
                      <Button
                        type="button"
                        size="sm"
                        variant="ghost"
                        onClick={() => setSelectedJobID(job.id)}
                        aria-label={t("syncJobs.detailTitle", { id: job.id })}
                      >
                        {t("syncJobs.viewDetail")}
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}

                {jobsTable.showNoData ? (
                  <TableEmptyState
                    colSpan={7}
                    className={`${spacing.cellFirst} py-10 text-center`}
                    message={t("common.noData")}
                  />
                ) : null}
              </TableBody>
            </Table>

            <div className="flex items-center justify-between border-t px-6 py-3 text-sm">
              <span className="text-muted-foreground">
                {t("syncJobs.pageLabel", { page: filters.page })}
              </span>
              <div className="flex gap-2">
                <Button
                  size="sm"
                  variant="outline"
                  disabled={filters.page <= 1 || jobsQuery.isFetching}
                  onClick={() => updateFilters({ page: filters.page - 1 })}
                >
                  {t("syncJobs.prevPage")}
                </Button>
                <Button
                  size="sm"
                  variant="outline"
                  disabled={jobsQuery.isFetching || (jobsQuery.data?.length ?? 0) < pageSize}
                  onClick={() => updateFilters({ page: filters.page + 1 })}
                >
                  {t("syncJobs.nextPage")}
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>

        <Dialog
          open={selectedJobID != null}
          onOpenChange={(open) => !open && setSelectedJobID(null)}
        >
          <DialogContent className="max-w-4xl">
            <DialogHeader>
              <DialogTitle>
                {t("syncJobs.detailTitle", { id: detailJob?.id ?? selectedJobID ?? "-" })}
              </DialogTitle>
              <DialogDescription>
                {detailJob
                  ? t("syncJobs.detailSubtitle", {
                      node: nodeName(detailJob.node_id),
                      time: formatDateTime(detailJob.created_at, i18n.language, timezone),
                    })
                  : t("common.loading")}
              </DialogDescription>
            </DialogHeader>

            {detailQuery.isLoading ? (
              <div className="py-8 text-center text-muted-foreground">{t("common.loading")}</div>
            ) : null}

            {detailQuery.error instanceof ApiError ? (
              <p className="text-sm text-destructive">{detailQuery.error.message}</p>
            ) : null}

            {detailJob ? (
              <div className="space-y-4">
                <div className="space-y-2">
                  <h3 className="text-sm font-semibold">{t("syncJobs.jobInfo")}</h3>
                  <div className="grid gap-2 text-sm sm:grid-cols-2">
                    <p>
                      <span className="text-muted-foreground">{t("syncJobs.jobId")}: </span>
                      {detailJob.id}
                    </p>
                    <p>
                      <span className="text-muted-foreground">{t("syncJobs.parentJobId")}: </span>
                      {detailJob.parent_job_id ?? "-"}
                    </p>
                    <p>
                      <span className="text-muted-foreground">{t("syncJobs.triggerSource")}: </span>
                      {sourceLabel(t, detailJob.trigger_source)}
                    </p>
                    <p>
                      <span className="text-muted-foreground">{t("syncJobs.colStatus")}: </span>
                      <Badge className="ml-1" variant={getStatusVariant(detailJob.status)}>
                        {statusLabel(t, detailJob.status)}
                      </Badge>
                    </p>
                    <p>
                      <span className="text-muted-foreground">{t("syncJobs.createdAt")}: </span>
                      {formatDateTime(detailJob.created_at, i18n.language, timezone)}
                    </p>
                    <p>
                      <span className="text-muted-foreground">{t("syncJobs.startedAt")}: </span>
                      {formatDateTime(detailJob.started_at, i18n.language, timezone)}
                    </p>
                    <p>
                      <span className="text-muted-foreground">{t("syncJobs.finishedAt")}: </span>
                      {formatDateTime(detailJob.finished_at, i18n.language, timezone)}
                    </p>
                    <p>
                      <span className="text-muted-foreground">{t("syncJobs.totalDuration")}: </span>
                      {formatDuration(detailJob.duration_ms)}
                    </p>
                    <p>
                      <span className="text-muted-foreground">{t("syncJobs.retryCount")}: </span>
                      {detailJob.attempt_count}
                    </p>
                    <p className="sm:col-span-2">
                      <span className="text-muted-foreground">{t("syncJobs.errorSummary")}: </span>
                      {detailJob.error_summary || "-"}
                    </p>
                  </div>
                </div>

                <div className="space-y-2">
                  <h3 className="text-sm font-semibold">{t("syncJobs.payloadSummary")}</h3>
                  <div className="grid gap-2 rounded-md border p-3 text-sm sm:grid-cols-3">
                    <p>
                      <span className="text-muted-foreground">{t("syncJobs.inboundCount")}: </span>
                      <span className="tabular-nums">{detailJob.inbound_count ?? 0}</span>
                    </p>
                    <p>
                      <span className="text-muted-foreground">
                        {t("syncJobs.activeUserCount")}:{" "}
                      </span>
                      <span className="tabular-nums">{detailJob.active_user_count ?? 0}</span>
                    </p>
                    <p>
                      <span className="text-muted-foreground">{t("syncJobs.payloadHash")}: </span>
                      <Tooltip>
                        <TooltipTrigger asChild>
                          <span className="font-mono text-xs">
                            {shortHash(detailJob.payload_hash)}
                          </span>
                        </TooltipTrigger>
                        <TooltipContent>{detailJob.payload_hash || "-"}</TooltipContent>
                      </Tooltip>
                    </p>
                  </div>
                </div>

                <div className="space-y-2">
                  <h3 className="text-sm font-semibold">{t("syncJobs.attemptTimeline")}</h3>
                  {detailAttempts.length === 0 ? (
                    <div className="rounded-md border p-3 text-sm text-muted-foreground">
                      {t("syncJobs.noAttempts")}
                    </div>
                  ) : (
                    <div className="space-y-2">
                      {detailAttempts.map((attempt) => (
                        <div key={attempt.id} className="rounded-md border p-3 text-sm">
                          <div className="flex flex-wrap items-center gap-x-3 gap-y-1">
                            <span className="font-medium">
                              {t("syncJobs.attemptNo", { no: attempt.attempt_no })}
                            </span>
                            <Badge variant={getStatusVariant(attempt.status)}>
                              {statusLabel(t, attempt.status)}
                            </Badge>
                            <span className="text-muted-foreground">
                              {t("syncJobs.httpStatus")}: {attempt.http_status || "-"}
                            </span>
                            <span className="text-muted-foreground">
                              {t("syncJobs.duration")}: {formatDuration(attempt.duration_ms)}
                            </span>
                            <span className="text-muted-foreground">
                              {t("syncJobs.backoff")}: {formatDuration(attempt.backoff_ms)}
                            </span>
                          </div>
                          <p className="mt-1 text-xs text-muted-foreground">
                            {formatDateTime(attempt.started_at, i18n.language, timezone)}
                            {attempt.finished_at
                              ? ` → ${formatDateTime(attempt.finished_at, i18n.language, timezone)}`
                              : ""}
                          </p>
                          <p className="mt-1 text-xs text-muted-foreground">
                            {t("syncJobs.error")}: {attempt.error_summary || "-"}
                          </p>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              </div>
            ) : null}

            <DialogFooter className="flex-col items-stretch gap-2 sm:flex-row sm:justify-between">
              {retryMutation.error instanceof ApiError ? (
                <span className="text-xs text-destructive">{retryMutation.error.message}</span>
              ) : (
                <span className="text-xs text-muted-foreground" />
              )}
              <div className="flex gap-2">
                {detailJob?.status === "failed" ? (
                  <AsyncButton
                    variant="outline"
                    disabled={retryMutation.isPending}
                    onClick={() => retryMutation.mutate(detailJob.id)}
                    pending={retryMutation.isPending}
                    pendingText={t("syncJobs.retrying")}
                  >
                    {t("syncJobs.retry")}
                  </AsyncButton>
                ) : null}
                <Button variant="ghost" onClick={() => setSelectedJobID(null)}>
                  {t("common.cancel")}
                </Button>
              </div>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </section>
    </div>
  );
}
