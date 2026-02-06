import { useQuery } from "@tanstack/react-query"

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
            <CardTitle>后端连通性</CardTitle>
            <CardDescription>用户预览（GET /api/users）</CardDescription>
          </CardHeader>
          <CardContent className="text-sm">
            {usersQuery.isLoading ? <p>加载中...</p> : null}
            {usersQuery.isError ? (
              <p className="text-destructive">
                请求失败，请检查 token 或后端服务。
              </p>
            ) : null}
            {usersQuery.data ? (
              <p>已返回 {usersQuery.data.length} 条用户记录</p>
            ) : null}
          </CardContent>
        </Card>

        <ChartAreaInteractive />
      </div>

      <DataTable data={dashboardData as any} />
    </div>
  )
}
