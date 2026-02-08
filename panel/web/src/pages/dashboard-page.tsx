import { useQuery } from "@tanstack/react-query"
import { useTranslation } from "react-i18next"

import { listUsers } from "@/lib/api/users"
import { listNodes } from "@/lib/api/nodes"
import { ChartAreaInteractive } from "@/components/chart-area-interactive"
import { SectionCards } from "@/components/section-cards"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { listTrafficNodesSummary, getTrafficTotalSummary } from "@/lib/api/traffic"
import { bytesToGBString } from "@/lib/units"

export function DashboardPage() {
  const { t } = useTranslation()
  const usersQuery = useQuery({
    queryKey: ["users", "dashboard-preview"],
    queryFn: () => listUsers({ limit: 10, offset: 0 }),
  })

  const nodesQuery = useQuery({
    queryKey: ["nodes", "dashboard"],
    queryFn: () => listNodes({ limit: 1000, offset: 0 }),
    refetchInterval: 10_000,
  })

  const total1hQuery = useQuery({
    queryKey: ["traffic", "total", "summary", "1h"],
    queryFn: () => getTrafficTotalSummary({ window: "1h" }),
    refetchInterval: 30_000,
  })

  const total24hQuery = useQuery({
    queryKey: ["traffic", "total", "summary", "24h"],
    queryFn: () => getTrafficTotalSummary({ window: "24h" }),
    refetchInterval: 30_000,
  })

  const totalAllQuery = useQuery({
    queryKey: ["traffic", "total", "summary", "all"],
    queryFn: () => getTrafficTotalSummary({ window: "all" }),
    refetchInterval: 60_000,
  })

  const nodes24hQuery = useQuery({
    queryKey: ["traffic", "nodes", "summary", "24h", "dashboard"],
    queryFn: () => listTrafficNodesSummary({ window: "24h" }),
    refetchInterval: 30_000,
  })

  const topNodes = (nodes24hQuery.data ?? [])
    .slice()
    .sort((a, b) => (b.upload + b.download) - (a.upload + a.download))
    .slice(0, 8)

  const nodeNameByID = new Map<number, string>()
  for (const n of nodesQuery.data ?? []) nodeNameByID.set(n.id, n.name)

  return (
    <div className="flex flex-1 flex-col gap-4">
      <SectionCards
        total1h={total1hQuery.data}
        total24h={total24hQuery.data}
        totalAll={totalAllQuery.data}
        isLoading={total1hQuery.isLoading || total24hQuery.isLoading || totalAllQuery.isLoading}
      />

      <div className="grid gap-4 px-4 lg:px-6">
        <Card className="@container/card">
          <CardHeader>
            <CardTitle>{t("dashboard.backendConnectivityTitle")}</CardTitle>
            <CardDescription>{t("dashboard.usersPreviewSubtitle")}</CardDescription>
          </CardHeader>
          <CardContent className="text-sm">
            {usersQuery.isLoading ? <p>{t("common.loading")}</p> : null}
            {usersQuery.isError ? (
              <p className="text-destructive">
                {t("dashboard.requestFailedHint")}
              </p>
            ) : null}
            {usersQuery.data ? (
              <p>{t("dashboard.usersReturned", { count: usersQuery.data.length })}</p>
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
                  <TableHead className="pl-6">{t("nodes.name")}</TableHead>
                  <TableHead>{t("dashboard.uplink")}</TableHead>
                  <TableHead className="pr-6">{t("dashboard.downlink")}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {nodes24hQuery.isLoading ? (
                  <TableRow>
                    <TableCell className="pl-6 py-8 text-center text-muted-foreground" colSpan={3}>
                      {t("common.loading")}
                    </TableCell>
                  </TableRow>
                ) : null}
                {!nodes24hQuery.isLoading && topNodes.length === 0 ? (
                  <TableRow>
                    <TableCell className="pl-6 py-8 text-center text-muted-foreground" colSpan={3}>
                      {t("common.noData")}
                    </TableCell>
                  </TableRow>
                ) : null}
                {topNodes.map((n) => (
                  <TableRow key={n.node_id}>
                    <TableCell className="pl-6 font-medium">
                      {nodeNameByID.get(n.node_id) ?? `#${n.node_id}`}
                    </TableCell>
                    <TableCell className="text-muted-foreground tabular-nums">
                      {bytesToGBString(n.upload)} GB
                    </TableCell>
                    <TableCell className="pr-6 text-muted-foreground tabular-nums">
                      {bytesToGBString(n.download)} GB
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
