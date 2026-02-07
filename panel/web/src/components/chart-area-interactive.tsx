"use client"

import * as React from "react"
import { useQuery } from "@tanstack/react-query"
import { Area, AreaChart, CartesianGrid, XAxis, YAxis } from "recharts"
import { useTranslation } from "react-i18next"

import { useIsMobile } from "@/hooks/use-mobile"
import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import {
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
  type ChartConfig,
} from "@/components/ui/chart"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group"
import { listTrafficTimeseries } from "@/lib/api/traffic"
import { bytesToGBString } from "@/lib/units"

type RangeKey = "24h" | "7d" | "30d"

function rangeToQueryParams(range: RangeKey): { window: string; bucket: "hour" | "day" } {
  if (range === "24h") return { window: "24h", bucket: "hour" }
  if (range === "7d") return { window: "168h", bucket: "hour" }
  return { window: "30d", bucket: "day" }
}

const chartConfig = {
  upload: {
    label: "Upload",
    color: "var(--primary)",
  },
  download: {
    label: "Download",
    color: "hsl(var(--muted-foreground))",
  },
} satisfies ChartConfig

export function ChartAreaInteractive() {
  const { t, i18n } = useTranslation()
  const isMobile = useIsMobile()
  const [timeRange, setTimeRange] = React.useState<RangeKey>("24h")

  React.useEffect(() => {
    if (isMobile) setTimeRange("24h")
  }, [isMobile])

  const q = rangeToQueryParams(timeRange)
  const tsQuery = useQuery({
    queryKey: ["traffic", "timeseries", "global", q.window, q.bucket],
    queryFn: () => listTrafficTimeseries({ window: q.window, bucket: q.bucket }),
    refetchInterval: 30_000,
  })

  const chartData = React.useMemo(() => {
    const rows = tsQuery.data ?? []
    return rows.map((r) => ({
      at: r.bucket_start,
      upload: r.upload,
      download: r.download,
    }))
  }, [tsQuery.data])

  return (
    <Card className="@container/card">
      <CardHeader>
        <CardTitle>{t("dashboard.trafficTrendTitle")}</CardTitle>
        <CardDescription>{t("dashboard.trafficTrendSubtitle")}</CardDescription>
        <CardAction>
          <ToggleGroup
            type="single"
            value={timeRange}
            onValueChange={(v) => setTimeRange((v as RangeKey) || "24h")}
            variant="outline"
            className="hidden *:data-[slot=toggle-group-item]:!px-4 @[767px]/card:flex"
          >
            <ToggleGroupItem value="24h">{t("dashboard.range24h")}</ToggleGroupItem>
            <ToggleGroupItem value="7d">{t("dashboard.range7d")}</ToggleGroupItem>
            <ToggleGroupItem value="30d">{t("dashboard.range30d")}</ToggleGroupItem>
          </ToggleGroup>

          <Select value={timeRange} onValueChange={(v) => setTimeRange((v as RangeKey) || "24h")}>
            <SelectTrigger
              className="flex w-40 **:data-[slot=select-value]:block **:data-[slot=select-value]:truncate @[767px]/card:hidden"
              size="sm"
              aria-label={t("dashboard.selectRange")}
            >
              <SelectValue placeholder={t("dashboard.range24h")} />
            </SelectTrigger>
            <SelectContent className="rounded-xl">
              <SelectItem value="24h" className="rounded-lg">
                {t("dashboard.range24h")}
              </SelectItem>
              <SelectItem value="7d" className="rounded-lg">
                {t("dashboard.range7d")}
              </SelectItem>
              <SelectItem value="30d" className="rounded-lg">
                {t("dashboard.range30d")}
              </SelectItem>
            </SelectContent>
          </Select>
        </CardAction>
      </CardHeader>

      <CardContent className="px-2 pt-4 sm:px-6 sm:pt-6">
        <ChartContainer config={chartConfig} className="aspect-auto h-[260px] w-full">
          <AreaChart data={chartData}>
            <defs>
              <linearGradient id="fillUpload" x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor="var(--color-upload)" stopOpacity={0.9} />
                <stop offset="95%" stopColor="var(--color-upload)" stopOpacity={0.08} />
              </linearGradient>
              <linearGradient id="fillDownload" x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor="var(--color-download)" stopOpacity={0.55} />
                <stop offset="95%" stopColor="var(--color-download)" stopOpacity={0.08} />
              </linearGradient>
            </defs>

            <CartesianGrid vertical={false} />
            <YAxis
              tickLine={false}
              axisLine={false}
              width={64}
              tickFormatter={(v) => `${bytesToGBString(Number(v))}G`}
            />
            <XAxis
              dataKey="at"
              tickLine={false}
              axisLine={false}
              tickMargin={8}
              minTickGap={32}
              tickFormatter={(value) => {
                const d = new Date(value)
                if (timeRange === "24h") {
                  return d.toLocaleTimeString(i18n.language, { hour: "2-digit", minute: "2-digit" })
                }
                return d.toLocaleDateString(i18n.language, { month: "short", day: "numeric" })
              }}
            />

            <ChartTooltip
              cursor={false}
              content={
                <ChartTooltipContent
                  labelFormatter={(label) => {
                    const d = new Date(label)
                    return d.toLocaleString(i18n.language)
                  }}
                  formatter={(value, name) => {
                    const n = typeof value === "number" ? value : Number(value)
                    const label = name === "upload" ? t("dashboard.uplink") : t("dashboard.downlink")
                    return [`${bytesToGBString(n)} GB`, label]
                  }}
                />
              }
            />

            <Area
              dataKey="upload"
              type="monotone"
              stroke="var(--color-upload)"
              fill="url(#fillUpload)"
              strokeWidth={2}
              dot={false}
            />
            <Area
              dataKey="download"
              type="monotone"
              stroke="var(--color-download)"
              fill="url(#fillDownload)"
              strokeWidth={2}
              dot={false}
            />
          </AreaChart>
        </ChartContainer>

        {tsQuery.isLoading ? (
          <div className="pt-3 text-xs text-muted-foreground">{t("common.loading")}</div>
        ) : null}
        {tsQuery.isError ? (
          <div className="pt-3 text-xs text-destructive">{t("common.loadFailed")}</div>
        ) : null}
      </CardContent>
    </Card>
  )
}

