import { useQuery } from "@tanstack/react-query"
import { useTranslation } from "react-i18next"

import { listUsers } from "@/lib/api/users"
import dashboardData from "@/app/dashboard/data.json"
import { ChartAreaInteractive } from "@/components/chart-area-interactive"
import { DataTable } from "@/components/data-table"
import { SectionCards } from "@/components/section-cards"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"

export function DashboardPage() {
  const { t } = useTranslation()
  const usersQuery = useQuery({
    queryKey: ["users", "dashboard-preview"],
    queryFn: () => listUsers({ limit: 10, offset: 0 }),
  })

  return (
    <div className="flex flex-1 flex-col gap-4">
      <SectionCards />

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
      </div>

      <DataTable data={dashboardData as any} />
    </div>
  )
}
