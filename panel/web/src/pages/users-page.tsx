import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useMemo, useState } from "react"

import { Button } from "@/components/ui/button"
import { ApiError } from "@/lib/api/client"
import { createUser, disableUser, listUsers, updateUser } from "@/lib/api/users"
import type { User, UserStatus } from "@/lib/api/types"

type StatusFilter = UserStatus | "all"

const statusOptions: Array<{ value: StatusFilter; label: string }> = [
  { value: "all", label: "全部" },
  { value: "active", label: "active" },
  { value: "disabled", label: "disabled" },
  { value: "expired", label: "expired" },
  { value: "traffic_exceeded", label: "traffic_exceeded" },
]

const editableStatusOptions: Array<{ value: UserStatus; label: string }> = [
  { value: "active", label: "active" },
  { value: "disabled", label: "disabled" },
  { value: "expired", label: "expired" },
  { value: "traffic_exceeded", label: "traffic_exceeded" },
]

type EditState = {
  user: User
  status: UserStatus
  trafficLimit: string
  trafficResetDay: string
  expireAt: string
  clearExpireAt: boolean
}

export function UsersPage() {
  const qc = useQueryClient()
  const [status, setStatus] = useState<StatusFilter>("all")
  const [newUsername, setNewUsername] = useState("")
  const [editing, setEditing] = useState<EditState | null>(null)

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

  const updateMutation = useMutation({
    mutationFn: (input: { id: number; payload: Record<string, unknown> }) =>
      updateUser(input.id, input.payload),
    onSuccess: async () => {
      setEditing(null)
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
                  <div className="flex items-center gap-2">
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={() => {
                        setEditing({
                          user: u,
                          status: u.status,
                          trafficLimit: String(u.traffic_limit ?? 0),
                          trafficResetDay: String(u.traffic_reset_day ?? 0),
                          expireAt: u.expire_at ?? "",
                          clearExpireAt: false,
                        })
                      }}
                    >
                      编辑
                    </Button>
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
                  </div>
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

      {editing ? (
        <div
          className="fixed inset-0 z-50 grid place-items-center bg-black/40 px-6"
          role="dialog"
          aria-modal="true"
          aria-label="编辑用户"
        >
          <div className="w-full max-w-xl rounded-2xl border border-slate-200 bg-white p-6 shadow-xl">
            <div className="flex items-start justify-between gap-4">
              <div className="space-y-1">
                <h2 className="text-lg font-semibold text-slate-900">编辑用户</h2>
                <p className="text-sm text-slate-500">
                  {editing.user.username} ({editing.user.uuid})
                </p>
              </div>
              <Button
                variant="outline"
                size="sm"
                onClick={() => setEditing(null)}
              >
                关闭
              </Button>
            </div>

            <div className="mt-5 grid grid-cols-1 gap-4 md:grid-cols-2">
              <div className="space-y-1">
                <label className="text-sm text-slate-700" htmlFor="edit-status">
                  状态
                </label>
                <select
                  id="edit-status"
                  className="h-10 w-full rounded-md border border-slate-300 bg-white px-3 text-sm outline-none focus:border-slate-500"
                  value={editing.status}
                  onChange={(e) =>
                    setEditing((prev) =>
                      prev
                        ? { ...prev, status: e.target.value as UserStatus }
                        : prev,
                    )
                  }
                >
                  {editableStatusOptions.map((opt) => (
                    <option key={opt.value} value={opt.value}>
                      {opt.label}
                    </option>
                  ))}
                </select>
              </div>

              <div className="space-y-1">
                <label
                  className="text-sm text-slate-700"
                  htmlFor="edit-traffic-limit"
                >
                  流量上限
                </label>
                <input
                  id="edit-traffic-limit"
                  inputMode="numeric"
                  className="h-10 w-full rounded-md border border-slate-300 px-3 text-sm outline-none focus:border-slate-500"
                  value={editing.trafficLimit}
                  onChange={(e) =>
                    setEditing((prev) =>
                      prev ? { ...prev, trafficLimit: e.target.value } : prev,
                    )
                  }
                />
              </div>

              <div className="space-y-1">
                <label
                  className="text-sm text-slate-700"
                  htmlFor="edit-traffic-reset-day"
                >
                  重置日
                </label>
                <input
                  id="edit-traffic-reset-day"
                  inputMode="numeric"
                  className="h-10 w-full rounded-md border border-slate-300 px-3 text-sm outline-none focus:border-slate-500"
                  value={editing.trafficResetDay}
                  onChange={(e) =>
                    setEditing((prev) =>
                      prev
                        ? { ...prev, trafficResetDay: e.target.value }
                        : prev,
                    )
                  }
                />
              </div>

              <div className="space-y-1 md:col-span-2">
                <label className="text-sm text-slate-700" htmlFor="edit-expire">
                  到期时间（RFC3339）
                </label>
                <input
                  id="edit-expire"
                  className="h-10 w-full rounded-md border border-slate-300 px-3 text-sm outline-none focus:border-slate-500"
                  value={editing.expireAt}
                  onChange={(e) =>
                    setEditing((prev) =>
                      prev ? { ...prev, expireAt: e.target.value } : prev,
                    )
                  }
                  placeholder="例如 2026-02-06T12:00:00Z"
                />
                <label className="mt-2 flex items-center gap-2 text-sm text-slate-600">
                  <input
                    type="checkbox"
                    checked={editing.clearExpireAt}
                    onChange={(e) =>
                      setEditing((prev) =>
                        prev
                          ? { ...prev, clearExpireAt: e.target.checked }
                          : prev,
                      )
                    }
                  />
                  清空到期时间
                </label>
              </div>
            </div>

            {updateMutation.isError ? (
              <p className="mt-4 text-sm text-red-600">
                {updateMutation.error instanceof ApiError
                  ? updateMutation.error.message
                  : "保存失败"}
              </p>
            ) : null}

            <div className="mt-6 flex items-center justify-end gap-2">
              <Button variant="outline" onClick={() => setEditing(null)}>
                取消
              </Button>
              <Button
                onClick={() => {
                  const limit = Number(editing.trafficLimit)
                  const resetDay = Number(editing.trafficResetDay)
                  const payload: Record<string, unknown> = {
                    status: editing.status,
                  }
                  if (!Number.isNaN(limit)) payload.traffic_limit = limit
                  if (!Number.isNaN(resetDay)) payload.traffic_reset_day = resetDay
                  if (editing.clearExpireAt) {
                    payload.expire_at = ""
                  } else if (editing.expireAt.trim() !== "") {
                    payload.expire_at = editing.expireAt.trim()
                  }

                  updateMutation.mutate({ id: editing.user.id, payload })
                }}
                disabled={updateMutation.isPending}
              >
                {updateMutation.isPending ? "保存中..." : "保存"}
              </Button>
            </div>
          </div>
        </div>
      ) : null}
    </section>
  )
}
