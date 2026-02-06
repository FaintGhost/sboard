import { useQuery } from "@tanstack/react-query"

import { listUsers } from "@/lib/api/users"

export function DashboardPage() {
  const usersQuery = useQuery({
    queryKey: ["users", "dashboard-preview"],
    queryFn: () => listUsers({ limit: 10, offset: 0 }),
  })

  return (
    <section className="space-y-4">
      <header>
        <h1 className="text-2xl font-semibold tracking-tight text-slate-900">
          仪表盘
        </h1>
        <p className="mt-1 text-sm text-slate-500">当前已接通认证与用户 API。</p>
      </header>

      <div className="rounded-lg border border-slate-200 bg-slate-50 p-4">
        <p className="text-sm text-slate-600">用户预览（`GET /api/users`）</p>
        {usersQuery.isLoading ? <p className="mt-2 text-sm">加载中...</p> : null}
        {usersQuery.isError ? (
          <p className="mt-2 text-sm text-red-600">请求失败，请检查 token 或后端服务。</p>
        ) : null}
        {usersQuery.data ? (
          <p className="mt-2 text-sm text-slate-800">
            已返回 {usersQuery.data.length} 条用户记录
          </p>
        ) : null}
      </div>
    </section>
  )
}
