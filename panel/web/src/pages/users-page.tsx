import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useEffect, useMemo, useState } from "react"
import { format } from "date-fns"
import { MoreHorizontal, Pencil, Ban, Search } from "lucide-react"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Calendar } from "@/components/ui/calendar"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Checkbox } from "@/components/ui/checkbox"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Skeleton } from "@/components/ui/skeleton"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { ApiError } from "@/lib/api/client"
import { createUser, disableUser, listUsers, updateUser } from "@/lib/api/users"
import { listGroups } from "@/lib/api/groups"
import { getUserGroups, putUserGroups } from "@/lib/api/user-groups"
import type { User, UserStatus } from "@/lib/api/types"
import { bytesToGBString, gbStringToBytes, rfc3339FromDateOnlyUTC } from "@/lib/units"

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

function StatusBadge({ status }: { status: UserStatus }) {
  const variant = status === "active"
    ? "default"
    : status === "disabled"
      ? "secondary"
      : "destructive"
  const label = status === "traffic_exceeded" ? "流量超限" : status
  return <Badge variant={variant}>{label}</Badge>
}

function formatTraffic(used: number, limit: number): string {
  const usedGB = bytesToGBString(used)
  if (limit === 0) return `${usedGB} GB / 无限`
  const limitGB = bytesToGBString(limit)
  return `${usedGB} / ${limitGB} GB`
}

function formatExpireDate(expireAt: string | null): string {
  if (!expireAt) return "永久"
  const date = new Date(expireAt)
  if (Number.isNaN(date.getTime())) return "永久"
  return format(date, "yyyy-MM-dd")
}

type EditState = {
  mode: "create" | "edit"
  user: User
  username: string
  status: UserStatus
  trafficLimit: string
  trafficResetDay: number
  expireDate: Date | null
  clearExpireAt: boolean
  groupIDs: number[]
  groupsLoadedFromServer: boolean
}

const defaultNewUser: User = {
  id: 0,
  uuid: "",
  username: "",
  traffic_limit: 0,
  traffic_used: 0,
  traffic_reset_day: 0,
  expire_at: null,
  status: "active",
}

export function UsersPage() {
  const qc = useQueryClient()
  const [status, setStatus] = useState<StatusFilter>("all")
  const [search, setSearch] = useState("")
  const [upserting, setUpserting] = useState<EditState | null>(null)
  const [deletingUser, setDeletingUser] = useState<User | null>(null)

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

  const groupsQuery = useQuery({
    queryKey: ["groups", { limit: 200, offset: 0 }],
    queryFn: () => listGroups({ limit: 200, offset: 0 }),
  })

  const userGroupsQuery = useQuery({
    queryKey: ["user-groups", upserting?.user.id ?? 0],
    queryFn: () => getUserGroups(upserting?.user.id ?? 0),
    enabled: !!upserting && upserting.mode === "edit" && upserting.user.id > 0,
  })

  useEffect(() => {
    if (!upserting || upserting.mode !== "edit") return
    if (!userGroupsQuery.data) return
    if (upserting.groupsLoadedFromServer) return

    setUpserting((prev) => {
      if (!prev || prev.mode !== "edit") return prev
      return {
        ...prev,
        groupIDs: (userGroupsQuery.data?.group_ids ?? []).slice().sort((a, b) => a - b),
        groupsLoadedFromServer: true,
      }
    })
  }, [upserting, userGroupsQuery.data])

  const createMutation = useMutation({
    mutationFn: createUser,
    onSuccess: async () => {
      setUpserting(null)
      await qc.invalidateQueries({ queryKey: ["users"] })
    },
  })

  const updateMutation = useMutation({
    mutationFn: (input: { id: number; payload: Record<string, unknown> }) =>
      updateUser(input.id, input.payload),
    onSuccess: async () => {
      setUpserting(null)
      await qc.invalidateQueries({ queryKey: ["users"] })
    },
  })

  const saveGroupsMutation = useMutation({
    mutationFn: (input: { userId: number; groupIDs: number[] }) =>
      putUserGroups(input.userId, { group_ids: input.groupIDs }),
    onSuccess: async () => {
      await qc.invalidateQueries({ queryKey: ["user-groups"] })
      await qc.invalidateQueries({ queryKey: ["users"] })
    },
  })

  const deleteMutation = useMutation({
    mutationFn: disableUser,
    onSuccess: async () => {
      setDeletingUser(null)
      await qc.invalidateQueries({ queryKey: ["users"] })
    },
  })

  // Filter users by search keyword
  const filteredUsers = useMemo(() => {
    const users = usersQuery.data ?? []
    if (!search.trim()) return users
    const keyword = search.trim().toLowerCase()
    return users.filter((u) => u.username.toLowerCase().includes(keyword))
  }, [usersQuery.data, search])

  return (
    <div className="px-4 lg:px-6">
      <section className="space-y-6">
        <header className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
          <div>
            <h1 className="text-2xl font-semibold tracking-tight">用户管理</h1>
            <p className="text-sm text-muted-foreground">
              管理系统中的所有用户账号
            </p>
          </div>
          <Button
            onClick={() => {
              createMutation.reset()
              updateMutation.reset()
              setUpserting({
                mode: "create",
                user: defaultNewUser,
                username: "",
                status: "active",
                trafficLimit: "0",
                trafficResetDay: 0,
                expireDate: null,
                clearExpireAt: false,
                groupIDs: [],
                groupsLoadedFromServer: true,
              })
            }}
          >
            创建用户
          </Button>
        </header>

        <Card>
          <CardHeader className="pb-3">
            <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
              <div className="flex flex-col gap-1.5">
                <CardTitle className="text-base">用户列表</CardTitle>
                <CardDescription>
                  {usersQuery.isLoading ? "加载中..." : null}
                  {usersQuery.isError ? "加载失败" : null}
                  {usersQuery.data ? `共 ${filteredUsers.length} 个用户` : null}
                </CardDescription>
              </div>
              <div className="flex flex-col gap-2 sm:flex-row sm:items-center">
                <div className="relative">
                  <Search className="absolute left-2.5 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
                  <Input
                    placeholder="搜索用户名..."
                    value={search}
                    onChange={(e) => setSearch(e.target.value)}
                    className="pl-8 w-full sm:w-48"
                  />
                </div>
                <Select
                  value={status}
                  onValueChange={(value) => setStatus(value as StatusFilter)}
                >
                  <SelectTrigger className="w-full sm:w-36" aria-label="状态筛选">
                    <SelectValue placeholder="选择状态" />
                  </SelectTrigger>
                  <SelectContent>
                    {statusOptions.map((opt) => (
                      <SelectItem key={opt.value} value={opt.value}>
                        {opt.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>
          </CardHeader>
          <CardContent className="p-0">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="pl-6">用户名</TableHead>
                  <TableHead>状态</TableHead>
                  <TableHead className="hidden md:table-cell">流量</TableHead>
                  <TableHead className="hidden sm:table-cell">到期日期</TableHead>
                  <TableHead className="w-12 pr-6">
                    <span className="sr-only">操作</span>
                  </TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {usersQuery.isLoading ? (
                  <>
                    {Array.from({ length: 5 }).map((_, i) => (
                      <TableRow key={i}>
                        <TableCell className="pl-6"><Skeleton className="h-4 w-24" /></TableCell>
                        <TableCell><Skeleton className="h-5 w-16" /></TableCell>
                        <TableCell className="hidden md:table-cell"><Skeleton className="h-4 w-20" /></TableCell>
                        <TableCell className="hidden sm:table-cell"><Skeleton className="h-4 w-20" /></TableCell>
                        <TableCell className="pr-6"><Skeleton className="h-8 w-8" /></TableCell>
                      </TableRow>
                    ))}
                  </>
                ) : null}
                {filteredUsers.map((u) => (
                  <TableRow key={u.id}>
                    <TableCell className="pl-6 font-medium">{u.username}</TableCell>
                    <TableCell><StatusBadge status={u.status} /></TableCell>
                    <TableCell className="hidden md:table-cell text-muted-foreground">
                      {formatTraffic(u.traffic_used, u.traffic_limit)}
                    </TableCell>
                    <TableCell className="hidden sm:table-cell text-muted-foreground">
                      {formatExpireDate(u.expire_at)}
                    </TableCell>
                    <TableCell className="pr-6">
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" size="icon" className="size-8">
                            <MoreHorizontal className="size-4" />
                            <span className="sr-only">操作菜单</span>
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem
                            onClick={() => {
                              const parsedExpire =
                                u.expire_at && !Number.isNaN(Date.parse(u.expire_at))
                                  ? new Date(u.expire_at)
                                  : null
                              createMutation.reset()
                              updateMutation.reset()
                              setUpserting({
                                mode: "edit",
                                user: u,
                                username: u.username,
                                status: u.status,
                                trafficLimit: bytesToGBString(u.traffic_limit ?? 0),
                                trafficResetDay: u.traffic_reset_day ?? 0,
                                expireDate: parsedExpire,
                                clearExpireAt: false,
                                groupIDs: [],
                                groupsLoadedFromServer: false,
                              })
                            }}
                          >
                            <Pencil className="mr-2 size-4" />
                            编辑
                          </DropdownMenuItem>
                          <DropdownMenuSeparator />
                          <DropdownMenuItem
                            variant="destructive"
                            disabled={u.status === "disabled"}
                            onClick={() => {
                              deleteMutation.reset()
                              setDeletingUser(u)
                            }}
                          >
                            <Ban className="mr-2 size-4" />
                            禁用
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </TableCell>
                  </TableRow>
                ))}
                {!usersQuery.isLoading && filteredUsers.length === 0 ? (
                  <TableRow>
                    <TableCell className="pl-6 py-8 text-center text-muted-foreground" colSpan={5}>
                      暂无数据
                    </TableCell>
                  </TableRow>
                ) : null}
              </TableBody>
            </Table>
          </CardContent>
        </Card>

        <Dialog
          open={!!upserting}
          onOpenChange={(open) => (!open ? setUpserting(null) : null)}
        >
          <DialogContent aria-label={upserting?.mode === "create" ? "创建用户" : "编辑用户"}>
            <DialogHeader>
              <DialogTitle>
                {upserting?.mode === "create" ? "创建用户" : "编辑用户"}
              </DialogTitle>
              {upserting?.mode === "edit" ? (
                <DialogDescription>
                  {upserting.user.username}
                </DialogDescription>
              ) : null}
            </DialogHeader>

            {upserting ? (
              <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                <div className="space-y-1 md:col-span-2">
                  <Label className="text-sm text-slate-700" htmlFor="edit-username">
                    用户名（username）
                  </Label>
                  <Input
                    id="edit-username"
                    value={upserting.username}
                    onChange={(e) =>
                      setUpserting((prev) =>
                        prev ? { ...prev, username: e.target.value } : prev,
                      )
                    }
                    placeholder="例如 alice"
                    autoFocus={upserting.mode === "create"}
                  />
                </div>

                <div className="space-y-2 md:col-span-2">
                  <div className="flex items-center justify-between">
                    <Label className="text-sm text-slate-700">分组（group_ids）</Label>
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      disabled={
                        upserting.mode !== "edit" ||
                        userGroupsQuery.isLoading ||
                        saveGroupsMutation.isPending
                      }
                      onClick={() => {
                        if (upserting.mode !== "edit") return
                        saveGroupsMutation.mutate({ userId: upserting.user.id, groupIDs: upserting.groupIDs })
                      }}
                    >
                      保存分组
                    </Button>
                  </div>
                  <p className="text-xs text-slate-500">
                    用户可以属于多个分组。订阅下发按分组生效。
                  </p>
                  <div className="rounded-lg border border-slate-200 p-3">
                    <div className="mt-3 grid grid-cols-1 gap-2 md:grid-cols-2">
                      {groupsQuery.data?.map((g) => {
                        const checked = upserting.groupIDs.includes(g.id)
                        return (
                          <label key={g.id} className="flex cursor-pointer items-center gap-2 text-sm text-slate-800">
                            <Checkbox
                              checked={checked}
                              onCheckedChange={(v) => {
                                const next = Boolean(v)
                                setUpserting((p) => {
                                  if (!p) return p
                                  const set = new Set(p.groupIDs)
                                  if (next) set.add(g.id)
                                  else set.delete(g.id)
                                  return { ...p, groupIDs: Array.from(set.values()).sort((a, b) => a - b) }
                                })
                              }}
                            />
                            <span className="font-medium">{g.name}</span>
                            <span className="text-xs text-slate-500">{g.description}</span>
                          </label>
                        )
                      })}
                      {groupsQuery.data && groupsQuery.data.length === 0 ? (
                        <div className="text-sm text-slate-500 md:col-span-2">
                          还没有分组，请先去"分组管理"创建。
                        </div>
                      ) : null}
                    </div>
                  </div>
                  <div className="text-sm text-amber-700">
                    {saveGroupsMutation.isError ? (
                      saveGroupsMutation.error instanceof ApiError ? saveGroupsMutation.error.message : "保存分组失败"
                    ) : null}
                  </div>
                </div>

                {upserting.mode === "edit" ? (
                  <>
                    <div className="space-y-1">
                      <Label className="text-sm text-slate-700">状态</Label>
                      <Select
                        value={upserting.status}
                        onValueChange={(value) =>
                          setUpserting((prev) =>
                            prev ? { ...prev, status: value as UserStatus } : prev,
                          )
                        }
                      >
                        <SelectTrigger className="w-full" aria-label="状态">
                          <SelectValue placeholder="选择状态" />
                        </SelectTrigger>
                        <SelectContent>
                          {editableStatusOptions.map((opt) => (
                            <SelectItem key={opt.value} value={opt.value}>
                              {opt.label}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>

                    <div className="space-y-1">
                      <Label className="text-sm text-slate-700" htmlFor="edit-traffic-limit">
                        流量上限（GB）
                      </Label>
                      <div className="relative">
                        <Input
                          id="edit-traffic-limit"
                          inputMode="decimal"
                          value={upserting.trafficLimit}
                          onChange={(e) =>
                            setUpserting((prev) =>
                              prev ? { ...prev, trafficLimit: e.target.value } : prev,
                            )
                          }
                          className="pr-12"
                          aria-label="流量上限"
                        />
                        <span className="pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 text-xs text-slate-500">
                          GB
                        </span>
                      </div>
                      <p className="text-xs text-slate-500">0 表示不限流量。</p>
                    </div>

                    <div className="space-y-1">
                      <Label className="text-sm text-slate-700" htmlFor="edit-traffic-reset-day">
                        重置日
                      </Label>
                      <Input
                        id="edit-traffic-reset-day"
                        type="number"
                        min={0}
                        max={31}
                        step={1}
                        inputMode="numeric"
                        value={String(upserting.trafficResetDay)}
                        onChange={(e) => {
                          const v = Number(e.target.value)
                          setUpserting((prev) =>
                            prev
                              ? {
                                  ...prev,
                                  trafficResetDay: Number.isFinite(v) ? v : 0,
                                }
                              : prev,
                          )
                        }}
                        onBlur={() =>
                          setUpserting((prev) => {
                            if (!prev) return prev
                            const v = Math.trunc(prev.trafficResetDay)
                            const clamped = Math.min(31, Math.max(0, v))
                            return { ...prev, trafficResetDay: clamped }
                          })
                        }
                        aria-label="重置日"
                      />
                      <p className="text-xs text-slate-500">
                        取值范围 0-31。0 表示不自动重置。若填写 29/30/31 且当月无该日期，则按当月最后一天计算。
                      </p>
                    </div>

                    <div className="space-y-1 md:col-span-2">
                      <Label className="text-sm text-slate-700" htmlFor="edit-expire">
                        到期日期
                      </Label>
                      <div className="flex flex-col gap-2 md:flex-row md:items-center">
                        <Popover>
                          <PopoverTrigger asChild>
                            <Button
                              id="edit-expire"
                              variant="outline"
                              className="w-full justify-start font-normal md:flex-1"
                            >
                              {upserting.expireDate ? (
                                format(upserting.expireDate, "yyyy-MM-dd")
                              ) : (
                                <span className="text-slate-500">选择日期</span>
                              )}
                            </Button>
                          </PopoverTrigger>
                          <PopoverContent className="w-auto p-0" align="start">
                            <Calendar
                              mode="single"
                              selected={upserting.expireDate ?? undefined}
                              onSelect={(date) =>
                                setUpserting((prev) =>
                                  prev
                                    ? {
                                        ...prev,
                                        expireDate: date ?? null,
                                        clearExpireAt: false,
                                      }
                                    : prev,
                                )
                              }
                              initialFocus
                            />
                          </PopoverContent>
                        </Popover>

                        <Button
                          type="button"
                          variant="outline"
                          className="md:w-24"
                          onClick={() =>
                            setUpserting((prev) =>
                              prev
                                ? { ...prev, expireDate: null, clearExpireAt: true }
                                : prev,
                            )
                          }
                          disabled={upserting.clearExpireAt}
                        >
                          清空
                        </Button>
                      </div>
                    </div>
                  </>
                ) : null}
              </div>
            ) : null}

            {upserting?.mode === "create" && createMutation.isError ? (
              <p className="text-sm text-red-600">
                {createMutation.error instanceof ApiError
                  ? createMutation.error.message
                  : "创建失败"}
              </p>
            ) : null}

            {upserting?.mode === "edit" && updateMutation.isError ? (
              <p className="text-sm text-red-600">
                {updateMutation.error instanceof ApiError
                  ? updateMutation.error.message
                  : "保存失败"}
              </p>
            ) : null}

            <DialogFooter>
              <Button variant="outline" onClick={() => setUpserting(null)}>
                取消
              </Button>
              <Button
                onClick={() => {
                  if (!upserting) return

                  if (upserting.mode === "create") {
                    createMutation.mutate({ username: upserting.username.trim() })
                    return
                  }

                  const payload: Record<string, unknown> = {}
                  const username = upserting.username.trim()
                  if (username && username !== upserting.user.username) {
                    payload.username = username
                  }
                  payload.status = upserting.status

                  const bytes = gbStringToBytes(upserting.trafficLimit)
                  if (bytes !== null) payload.traffic_limit = bytes

                  if (
                    Number.isInteger(upserting.trafficResetDay) &&
                    upserting.trafficResetDay >= 0 &&
                    upserting.trafficResetDay <= 31
                  ) {
                    payload.traffic_reset_day = upserting.trafficResetDay
                  }

                  if (upserting.clearExpireAt) {
                    payload.expire_at = ""
                  } else if (upserting.expireDate) {
                    payload.expire_at = rfc3339FromDateOnlyUTC(upserting.expireDate)
                  }

                  updateMutation.mutate({ id: upserting.user.id, payload })
                }}
                disabled={
                  !upserting ||
                  (upserting.mode === "create"
                    ? createMutation.isPending || !upserting.username.trim()
                    : updateMutation.isPending || !upserting.username.trim())
                }
              >
                {upserting?.mode === "create"
                  ? createMutation.isPending
                    ? "创建中..."
                    : "创建"
                  : updateMutation.isPending
                    ? "保存中..."
                    : "保存"}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        <Dialog
          open={!!deletingUser}
          onOpenChange={(open) => (!open ? setDeletingUser(null) : null)}
        >
          <DialogContent aria-label="禁用用户">
            <DialogHeader>
              <DialogTitle>禁用用户</DialogTitle>
              <DialogDescription>
                确定要禁用用户 <strong>{deletingUser?.username}</strong> 吗？禁用后该用户将无法使用订阅。
              </DialogDescription>
            </DialogHeader>

            {deleteMutation.isError ? (
              <p className="text-sm text-red-600">
                {deleteMutation.error instanceof ApiError
                  ? deleteMutation.error.message
                  : "禁用失败"}
              </p>
            ) : null}

            <DialogFooter>
              <Button variant="outline" onClick={() => setDeletingUser(null)}>
                取消
              </Button>
              <Button
                variant="destructive"
                onClick={() => {
                  if (!deletingUser) return
                  deleteMutation.mutate(deletingUser.id)
                }}
                disabled={deleteMutation.isPending}
              >
                {deleteMutation.isPending ? "禁用中..." : "确认禁用"}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </section>
    </div>
  )
}
