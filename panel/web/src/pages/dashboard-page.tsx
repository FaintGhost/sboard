import { useMemo } from "react";
import { useQuery } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";
import { Link } from "react-router-dom";
import {
  IconArrowRight,
  IconCloud,
  IconRefresh,
  IconServer2,
  IconUsers,
} from "@tabler/icons-react";

import { listUsers } from "@/lib/api/users";
import { listAllNodes } from "@/lib/api/nodes";
import { ChartAreaInteractive } from "@/components/chart-area-interactive";
import { SectionCards } from "@/components/section-cards";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { listTrafficNodesSummary, getTrafficTotalSummary } from "@/lib/api/traffic";
import { tableColumnSpacing } from "@/lib/table-spacing";
import { bytesToGBString } from "@/lib/units";

export function DashboardPage() {
  const { t } = useTranslation();
  const spacing = tableColumnSpacing.three;
  const usersQuery = useQuery({
    queryKey: ["users", "dashboard-preview"],
    queryFn: () => listUsers({ limit: 10, offset: 0 }),
  });

  const nodesQuery = useQuery({
    queryKey: ["nodes", "dashboard"],
    queryFn: () => listAllNodes(),
    refetchInterval: 10_000,
  });

  const total1hQuery = useQuery({
    queryKey: ["traffic", "total", "summary", "1h"],
    queryFn: () => getTrafficTotalSummary({ window: "1h" }),
    refetchInterval: 30_000,
  });

  const total24hQuery = useQuery({
    queryKey: ["traffic", "total", "summary", "24h"],
    queryFn: () => getTrafficTotalSummary({ window: "24h" }),
    refetchInterval: 30_000,
  });

  const totalAllQuery = useQuery({
    queryKey: ["traffic", "total", "summary", "all"],
    queryFn: () => getTrafficTotalSummary({ window: "all" }),
    refetchInterval: 60_000,
  });

  const nodes24hQuery = useQuery({
    queryKey: ["traffic", "nodes", "summary", "24h", "dashboard"],
    queryFn: () => listTrafficNodesSummary({ window: "24h" }),
    refetchInterval: 30_000,
  });

  const topNodes = (nodes24hQuery.data ?? [])
    .slice()
    .sort((a, b) => b.upload + b.download - (a.upload + a.download))
    .slice(0, 8);

  const nodeNameByID = new Map<number, string>();
  for (const n of nodesQuery.data ?? []) nodeNameByID.set(n.id, n.name);

  const activeUsers = useMemo(
    () => (usersQuery.data ?? []).filter((user) => user.status === "active").length,
    [usersQuery.data],
  );

  const onlineNodes = useMemo(
    () => (nodesQuery.data ?? []).filter((node) => node.status === "online").length,
    [nodesQuery.data],
  );

  const quickLinks = [
    {
      label: t("nav.users"),
      to: "/users",
      icon: IconUsers,
      value: usersQuery.data ? String(usersQuery.data.length) : "-",
      sub: t("dashboard.quickUsersSub", { active: activeUsers }),
    },
    {
      label: t("nav.nodes"),
      to: "/nodes",
      icon: IconServer2,
      value: nodesQuery.data ? String(nodesQuery.data.length) : "-",
      sub: t("dashboard.quickNodesSub", { online: onlineNodes }),
    },
    {
      label: t("nav.syncJobs"),
      to: "/sync-jobs",
      icon: IconRefresh,
      value: topNodes.length > 0 ? String(topNodes.length) : "-",
      sub: t("dashboard.quickSyncSub"),
    },
    {
      label: t("nav.subscriptions"),
      to: "/subscriptions",
      icon: IconCloud,
      value: usersQuery.data ? String(usersQuery.data.length) : "-",
      sub: t("dashboard.quickSubsSub"),
    },
  ];

  return (
    <div className="flex flex-1 flex-col gap-4">
      <SectionCards
        total1h={total1hQuery.data}
        total24h={total24hQuery.data}
        totalAll={totalAllQuery.data}
        isLoading={total1hQuery.isLoading || total24hQuery.isLoading || totalAllQuery.isLoading}
      />

      <div className="grid gap-4 px-4 lg:px-6">
        <Card className="@container/card border-border/80 bg-card shadow-[0_1px_0_0_rgba(255,255,255,0.32)_inset,0_20px_36px_-30px_rgba(0,0,0,0.48)] dark:shadow-[0_1px_0_0_rgba(255,255,255,0.08)_inset,0_24px_44px_-34px_rgba(0,0,0,0.88)]">
          <CardHeader>
            <CardTitle>{t("dashboard.systemOverviewTitle")}</CardTitle>
            <CardDescription>{t("dashboard.systemOverviewSubtitle")}</CardDescription>
          </CardHeader>
          <CardContent className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
            {quickLinks.map((item) => (
              <Link
                key={item.to}
                to={item.to}
                className="group rounded-lg border border-border/75 bg-background/85 p-3 transition-all hover:border-border hover:bg-background"
              >
                <div className="mb-2 flex items-center justify-between gap-3">
                  <div className="flex items-center gap-2 text-sm font-medium text-foreground">
                    <item.icon className="size-4 text-primary" />
                    {item.label}
                  </div>
                  <IconArrowRight className="size-4 text-muted-foreground transition-transform group-hover:translate-x-0.5" />
                </div>
                <div className="text-2xl font-semibold tabular-nums">{item.value}</div>
                <div className="mt-1 text-xs text-muted-foreground">{item.sub}</div>
              </Link>
            ))}

            {usersQuery.isError || nodesQuery.isError ? (
              <div className="sm:col-span-2 xl:col-span-4 rounded-md border border-destructive/30 bg-destructive/5 px-3 py-2 text-xs text-destructive">
                {t("dashboard.requestFailedHint")}
              </div>
            ) : null}
          </CardContent>
        </Card>

        <ChartAreaInteractive />

        <Card className="@container/card">
          <CardHeader>
            <CardTitle>{t("dashboard.topNodesTitle")}</CardTitle>
            <CardDescription>{t("dashboard.topNodesSubtitle")}</CardDescription>
          </CardHeader>
          <CardContent className="p-0">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className={spacing.headFirst}>{t("nodes.name")}</TableHead>
                  <TableHead className={spacing.headMiddle}>{t("dashboard.uplink")}</TableHead>
                  <TableHead className={spacing.headLast}>{t("dashboard.downlink")}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {nodes24hQuery.isLoading ? (
                  <TableRow>
                    <TableCell
                      className={`${spacing.cellFirst} py-8 text-center text-muted-foreground`}
                      colSpan={3}
                    >
                      {t("common.loading")}
                    </TableCell>
                  </TableRow>
                ) : null}
                {!nodes24hQuery.isLoading && topNodes.length === 0 ? (
                  <TableRow>
                    <TableCell
                      className={`${spacing.cellFirst} py-8 text-center text-muted-foreground`}
                      colSpan={3}
                    >
                      {t("common.noData")}
                    </TableCell>
                  </TableRow>
                ) : null}
                {topNodes.map((n) => (
                  <TableRow key={n.node_id}>
                    <TableCell className={`${spacing.cellFirst} font-medium`}>
                      {nodeNameByID.get(n.node_id) ?? `#${n.node_id}`}
                    </TableCell>
                    <TableCell
                      className={`${spacing.cellMiddle} text-muted-foreground tabular-nums`}
                    >
                      {bytesToGBString(n.upload)} GB
                    </TableCell>
                    <TableCell className={`${spacing.cellLast} text-muted-foreground tabular-nums`}>
                      {bytesToGBString(n.download)} GB
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>

            {topNodes.length > 0 ? (
              <div className="flex justify-end border-t px-4 py-3">
                <Button asChild size="sm" variant="outline">
                  <Link to="/nodes">{t("dashboard.viewAllNodes")}</Link>
                </Button>
              </div>
            ) : null}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
