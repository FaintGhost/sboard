import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useMemo, useState } from "react"

import { Button } from "@/components/ui/button"
import { ApiError } from "@/lib/api/client"
import { createUser, disableUser, listUsers, updateUser } from "@/lib/api/users"
import type { UserStatus } from "@/lib/api/types"

type StatusFilter = UserStatus | "all"

const statusOptions: Array<{ value: StatusFilter; label: string }> = [
  { value: "all", label: "全部" },
  { value: "active", label: "active" },
  { value: "disabled", label: "disabled" },
  { value: "expired", label: "expired" },
  { value: "traffic_exceeded", label: "traffic_exceeded" },
]

export function UsersPage() {
  const qc = useQueryClient()
  const [status, setStatus] = useState<StatusFilter>("all")
  const [newUsername, setNewUsername] = useState("")

  const queryParams = useMemo(
    () => ({
      limit: 50,
      offset: 0,
      status: status === "all" ? undefined : status,
    }),
    [status],
  )

  const usersQuery = useQuery({
    queryKey: ["users", queryParams],
    queryFn: () => listUsers(queryParams),
  })

  const createMutation = useMutation({
    mutationFn: createUser,
    onSuccess: async () => {
      setNewUsername("")
      await qc.invalidateQueries({ queryKey: ["users"] })
    },
  })

  const disableMutation = useMutation({
    mutationFn: disableUser,
    onSuccess: async () => {
      await qc.invalidateQueries({ queryKey: ["users"] })
    },
  })

  const enableMutation = useMutation({
    mutationFn: (id: number) => updateUser(id, { status: "active" }),
    onSuccess: async () => {
      await qc.invalidateQueries({ queryKey: ["users"] })
    },
  })

  return (
    <section className="space-y-6">
      <header className="space-y-1">
        <h1 className="text-2xl font-semibold tracking-tight text-slate-900">
          用户管理
        </h1>
        <p className="text-sm text-slate-500">
          对接后端 `GET/POST/DELETE /api/users*`。
        </p>
      </header>

      <div className="flex flex-col gap-3 md:flex-row md:items-end md:justify-between">
        <div className="space-y-1">
          <label className="text-sm text-slate-700" htmlFor="status">
            状态筛选
          </label>
          <select
            id="status"
            className="h-10 w-56 rounded-md border border-slate-300 bg-white px-3 text-sm outline-none focus:border-slate-500"
            value={status}
            onChange={(e) => setStatus(e.target.value as StatusFilter)}
          >
            {statusOptions.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>
        </div>

        <div className="flex flex-col gap-2 md:flex-row md:items-end">
          <div className="space-y-1">
            <label className="text-sm text-slate-700" htmlFor="username">
              新建用户（username）
            </label>
            <input
              id="username"
              className="h-10 w-64 rounded-md border border-slate-300 px-3 text-sm outline-none focus:border-slate-500"
              value={newUsername}
              onChange={(e) => setNewUsername(e.target.value)}
              placeholder="例如 alice"
            />
          </div>
          <Button
            onClick={() =>
              createMutation.mutate({ username: newUsername.trim() })
            }
            disabled={!newUsername.trim() || createMutation.isPending}
          >
            {createMutation.isPending ? "创建中..." : "创建用户"}
          </Button>
        </div>
      </div>

      {createMutation.isError ? (
        <p className="text-sm text-red-600">
          {createMutation.error instanceof ApiError
            ? createMutation.error.message
            : "创建失败"}
        </p>
      ) : null}

      <div className="overflow-hidden rounded-xl border border-slate-200">
        <div className="border-b border-slate-200 bg-slate-50 px-4 py-2 text-sm text-slate-600">
          {usersQuery.isLoading ? "加载中..." : null}
          {usersQuery.isError ? "加载失败" : null}
          {usersQuery.data ? `共 ${usersQuery.data.length} 个用户` : null}
        </div>

        <table className="w-full text-left text-sm">
          <thead className="bg-white">
            <tr className="border-b border-slate-200">
              <th className="px-4 py-3 font-medium text-slate-700">username</th>
              <th className="px-4 py-3 font-medium text-slate-700">uuid</th>
              <th className="px-4 py-3 font-medium text-slate-700">status</th>
              <th className="px-4 py-3 font-medium text-slate-700">action</th>
            </tr>
          </thead>
          <tbody className="bg-white">
            {usersQuery.data?.map((u) => (
              <tr key={u.id} className="border-b border-slate-100">
                <td className="px-4 py-3 text-slate-900">{u.username}</td>
                <td className="px-4 py-3 font-mono text-xs text-slate-600">
                  {u.uuid}
                </td>
                <td className="px-4 py-3 text-slate-700">{u.status}</td>
                <td className="px-4 py-3">
                  {u.status === "disabled" ? (
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={() => enableMutation.mutate(u.id)}
                      disabled={enableMutation.isPending}
                    >
                      启用
                    </Button>
                  ) : (
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={() => {
                        if (!confirm(`禁用用户 ${u.username} ?`)) return
                        disableMutation.mutate(u.id)
                      }}
                      disabled={disableMutation.isPending}
                    >
                      禁用
                    </Button>
                  )}
                </td>
              </tr>
            ))}
            {usersQuery.data && usersQuery.data.length === 0 ? (
              <tr>
                <td className="px-4 py-6 text-slate-500" colSpan={4}>
                  暂无数据
                </td>
              </tr>
            ) : null}
          </tbody>
        </table>
      </div>
    </section>
  )
}
