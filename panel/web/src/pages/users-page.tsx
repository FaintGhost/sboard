import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useEffect, useMemo, useState } from "react"
import { useTranslation } from "react-i18next"
import { format } from "date-fns"
import { MoreHorizontal, Pencil, Ban, Search, Trash2 } from "lucide-react"

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
import { createUser, deleteUser, disableUser, listUsers, updateUser } from "@/lib/api/users"
import { listGroups } from "@/lib/api/groups"
import { getUserGroups, putUserGroups } from "@/lib/api/user-groups"
import type { User, UserStatus } from "@/lib/api/types"
import { tableColumnSpacing } from "@/lib/table-spacing"
import { bytesToGBString, gbStringToBytes, rfc3339FromDateOnlyUTC } from "@/lib/units"

type StatusFilter = UserStatus | "all"

function StatusBadge({ status }: { status: UserStatus }) {
  const { t } = useTranslation()
  const variant = status === "active"
    ? "default"
    : status === "disabled"
      ? "secondary"
      : "destructive"
  const label = status === "traffic_exceeded"
    ? t("users.status.trafficExceeded")
    : t(`users.status.${status}`)
  return <Badge variant={variant}>{label}</Badge>
}

function formatTraffic(used: number, limit: number, t: (key: string, options?: Record<string, unknown>) => string): string {
  const usedGB = bytesToGBString(used)
  if (limit === 0) return t("users.trafficUnlimited", { used: usedGB })
  const limitGB = bytesToGBString(limit)
  return t("users.trafficFormat", { used: usedGB, limit: limitGB })
}

function formatExpireDate(expireAt: string | null, t: (key: string) => string): string {
  if (!expireAt) return t("common.permanent")
  const date = new Date(expireAt)
  if (Number.isNaN(date.getTime())) return t("common.permanent")
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
  group_ids: [],
  traffic_limit: 0,
  traffic_used: 0,
  traffic_reset_day: 0,
  expire_at: null,
  status: "active",
}

export function UsersPage() {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const [status, setStatus] = useState<StatusFilter>("all")
  const [search, setSearch] = useState("")
  const [upserting, setUpserting] = useState<EditState | null>(null)
  const [disablingUser, setDisablingUser] = useState<User | null>(null)
  const [deletingUser, setDeletingUser] = useState<User | null>(null)
  const spacing = tableColumnSpacing.five

  const statusOptions: Array<{ value: StatusFilter; label: string }> = [
    { value: "all", label: t("common.all") },
    { value: "active", label: t("users.status.active") },
    { value: "disabled", label: t("users.status.disabled") },
    { value: "expired", label: t("users.status.expired") },
    { value: "traffic_exceeded", label: t("users.status.trafficExceeded") },
  ]

  const editableStatusOptions: Array<{ value: UserStatus; label: string }> = [
    { value: "active", label: t("users.status.active") },
    { value: "disabled", label: t("users.status.disabled") },
    { value: "expired", label: t("users.status.expired") },
    { value: "traffic_exceeded", label: t("users.status.trafficExceeded") },
  ]

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
    mutationFn: async (input: { username: string; groupIDs: number[] }) => {
      const created = await createUser({ username: input.username })
      if (input.groupIDs.length > 0) {
        await putUserGroups(created.id, { group_ids: input.groupIDs })
      }
      return created
    },
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

  const disableMutation = useMutation({
    mutationFn: disableUser,
    onSuccess: async () => {
      setDisablingUser(null)
      await qc.invalidateQueries({ queryKey: ["users"] })
    },
  })

  const deleteMutation = useMutation({
    mutationFn: deleteUser,
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

  const groupNameByID = useMemo(() => {
    const map = new Map<number, string>()
    for (const group of groupsQuery.data ?? []) {
      map.set(group.id, group.name)
    }
    return map
  }, [groupsQuery.data])

  return (
    <div className="px-4 lg:px-6">
      <section className="space-y-6">
        <header className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
          <div>
            <h1 className="text-2xl font-semibold tracking-tight">{t("users.title")}</h1>
            <p className="text-sm text-muted-foreground">
              {t("users.subtitle")}
            </p>
          </div>
          <Button
            onClick={() => {
              createMutation.reset()
              updateMutation.reset()
              saveGroupsMutation.reset()
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
            {t("users.createUser")}
          </Button>
        </header>

        <Card>
          <CardHeader className="pb-3">
            <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
              <div className="flex flex-col gap-1.5">
                <CardTitle className="text-base">{t("users.list")}</CardTitle>
                <CardDescription>
                  {usersQuery.isLoading ? t("common.loading") : null}
                  {usersQuery.isError ? t("common.loadFailed") : null}
                  {usersQuery.data ? t("users.count", { count: filteredUsers.length }) : null}
                </CardDescription>
              </div>
              <div className="flex flex-col gap-2 sm:flex-row sm:items-center">
                <div className="relative">
                  <Search className="absolute left-2.5 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
                  <Input
                    placeholder={t("users.searchPlaceholder")}
                    value={search}
                    onChange={(e) => setSearch(e.target.value)}
                    className="pl-8 w-full sm:w-48"
                  />
                </div>
                <Select
                  value={status}
                  onValueChange={(value) => setStatus(value as StatusFilter)}
                >
                  <SelectTrigger className="w-full sm:w-36" aria-label={t("users.statusFilter")}>
                    <SelectValue placeholder={t("users.statusFilter")} />
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
                  <TableHead className={spacing.headFirst}>{t("users.username")}</TableHead>
                  <TableHead className={spacing.headMiddle}>{t("users.groups")}</TableHead>
                  <TableHead className={spacing.headMiddle}>{t("common.status")}</TableHead>
                  <TableHead className={`${spacing.headMiddle} hidden md:table-cell`}>{t("users.traffic")}</TableHead>
                  <TableHead className={`${spacing.headMiddle} hidden sm:table-cell`}>{t("users.expireDate")}</TableHead>
                  <TableHead className={`${spacing.headLast} w-12`}>
                    <span className="sr-only">{t("common.actions")}</span>
                  </TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {usersQuery.isLoading ? (
                  <>
                    {Array.from({ length: 5 }).map((_, i) => (
                      <TableRow key={i}>
                        <TableCell className={spacing.cellFirst}><Skeleton className="h-4 w-24" /></TableCell>
                        <TableCell className={spacing.cellMiddle}><Skeleton className="h-5 w-28" /></TableCell>
                        <TableCell className={spacing.cellMiddle}><Skeleton className="h-5 w-16" /></TableCell>
                        <TableCell className={`${spacing.cellMiddle} hidden md:table-cell`}><Skeleton className="h-4 w-20" /></TableCell>
                        <TableCell className={`${spacing.cellMiddle} hidden sm:table-cell`}><Skeleton className="h-4 w-20" /></TableCell>
                        <TableCell className={spacing.cellLast}><Skeleton className="h-8 w-8" /></TableCell>
                      </TableRow>
                    ))}
                  </>
                ) : null}
                {filteredUsers.map((u) => {
                  const visibleGroupIDs = u.group_ids.filter((groupID) => groupNameByID.has(groupID))
                  return (
                  <TableRow key={u.id}>
                    <TableCell className={`${spacing.cellFirst} font-medium`}>{u.username}</TableCell>
                    <TableCell className={spacing.cellMiddle}>
                      {visibleGroupIDs.length > 0 ? (
                        <div className="flex flex-wrap gap-1">
                          {visibleGroupIDs.map((groupID) => {
                            const groupName = groupNameByID.get(groupID)
                            return (
                              <Badge key={`${u.id}-${groupID}`} variant="secondary">
                                {groupName}
                              </Badge>
                            )
                          })}
                        </div>
                      ) : (
                        <span className="text-muted-foreground">-</span>
                      )}
                    </TableCell>
                    <TableCell className={spacing.cellMiddle}><StatusBadge status={u.status} /></TableCell>
                    <TableCell className={`${spacing.cellMiddle} hidden md:table-cell text-muted-foreground`}>
                      {formatTraffic(u.traffic_used, u.traffic_limit, t)}
                    </TableCell>
                    <TableCell className={`${spacing.cellMiddle} hidden sm:table-cell text-muted-foreground`}>
                      {formatExpireDate(u.expire_at, t)}
                    </TableCell>
                    <TableCell className={spacing.cellLast}>
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" size="icon" className="size-8">
                            <MoreHorizontal className="size-4" />
                            <span className="sr-only">{t("common.actions")}</span>
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
                              saveGroupsMutation.reset()
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
                            {t("common.edit")}
                          </DropdownMenuItem>
                          <DropdownMenuSeparator />
                          <DropdownMenuItem
                            variant="destructive"
                            disabled={u.status === "disabled"}
                            onClick={() => {
                              disableMutation.reset()
                              setDisablingUser(u)
                            }}
                          >
                            <Ban className="mr-2 size-4" />
                            {t("common.disable")}
                          </DropdownMenuItem>
                          <DropdownMenuItem
                            variant="destructive"
                            onClick={() => {
                              deleteMutation.reset()
                              setDeletingUser(u)
                            }}
                          >
                            <Trash2 className="mr-2 size-4" />
                            {t("common.delete")}
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </TableCell>
                  </TableRow>
                )})}
                {!usersQuery.isLoading && filteredUsers.length === 0 ? (
                  <TableRow>
                    <TableCell className={`${spacing.cellFirst} py-8 text-center text-muted-foreground`} colSpan={6}>
                      {t("common.noData")}
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
          <DialogContent aria-label={upserting?.mode === "create" ? t("users.createUser") : t("users.editUser")}>
            <DialogHeader>
              <DialogTitle>
                {upserting?.mode === "create" ? t("users.createUser") : t("users.editUser")}
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
                    {t("users.username")}
                  </Label>
                  <Input
                    id="edit-username"
                    value={upserting.username}
                    onChange={(e) =>
                      setUpserting((prev) =>
                        prev ? { ...prev, username: e.target.value } : prev,
                      )
                    }
                    placeholder={t("users.usernamePlaceholder")}
                    autoFocus={upserting.mode === "create"}
                  />
                </div>

                <div className="space-y-2 md:col-span-2">
                  <Label className="text-sm text-slate-700">{t("users.groups")}</Label>
                  <p className="text-xs text-slate-500">
                    {t("users.groupsHint")}
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
                          {t("users.noGroups")}
                        </div>
                      ) : null}
                    </div>
                  </div>
                  <div className="text-sm text-amber-700">
                    {saveGroupsMutation.isError ? (
                      saveGroupsMutation.error instanceof ApiError ? saveGroupsMutation.error.message : t("users.saveGroupsFailed")
                    ) : null}
                  </div>
                </div>

                {upserting.mode === "edit" ? (
                  <>
                    <div className="space-y-1">
                      <Label className="text-sm text-slate-700">{t("common.status")}</Label>
                      <Select
                        value={upserting.status}
                        onValueChange={(value) =>
                          setUpserting((prev) =>
                            prev ? { ...prev, status: value as UserStatus } : prev,
                          )
                        }
                      >
                        <SelectTrigger className="w-full" aria-label={t("common.status")}>
                          <SelectValue placeholder={t("users.statusFilter")} />
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
                        {t("users.trafficLimit")}
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
                          aria-label={t("users.trafficLimit")}
                        />
                        <span className="pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 text-xs text-slate-500">
                          GB
                        </span>
                      </div>
                      <p className="text-xs text-slate-500">{t("users.trafficLimitHint")}</p>
                    </div>

                    <div className="space-y-1">
                      <Label className="text-sm text-slate-700" htmlFor="edit-traffic-reset-day">
                        {t("users.trafficResetDay")}
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
                        aria-label={t("users.trafficResetDay")}
                      />
                      <p className="text-xs text-slate-500">
                        {t("users.trafficResetDayHint")}
                      </p>
                    </div>

                    <div className="space-y-1 md:col-span-2">
                      <Label className="text-sm text-slate-700" htmlFor="edit-expire">
                        {t("users.expireDate")}
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
                                <span className="text-slate-500">{t("users.selectDate")}</span>
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
                          {t("users.clearDate")}
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
                  : t("users.createFailed")}
              </p>
            ) : null}

            {upserting?.mode === "edit" && updateMutation.isError ? (
              <p className="text-sm text-red-600">
                {updateMutation.error instanceof ApiError
                  ? updateMutation.error.message
                  : t("users.saveFailed")}
              </p>
            ) : null}

            <DialogFooter>
              <Button variant="outline" onClick={() => setUpserting(null)}>
                {t("common.cancel")}
              </Button>
              <Button
                onClick={async () => {
                  if (!upserting) return

                  if (upserting.mode === "create") {
                    createMutation.mutate({
                      username: upserting.username.trim(),
                      groupIDs: upserting.groupIDs,
                    })
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

                  // Save user info and groups together
                  await Promise.all([
                    updateMutation.mutateAsync({ id: upserting.user.id, payload }),
                    saveGroupsMutation.mutateAsync({ userId: upserting.user.id, groupIDs: upserting.groupIDs }),
                  ])
                }}
                disabled={
                  !upserting ||
                  (upserting.mode === "create"
                    ? createMutation.isPending || !upserting.username.trim()
                    : updateMutation.isPending || saveGroupsMutation.isPending || !upserting.username.trim())
                }
              >
                {upserting?.mode === "create"
                  ? createMutation.isPending
                    ? t("common.creating")
                    : t("common.create")
                  : updateMutation.isPending
                    ? t("common.saving")
                    : t("common.save")}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        <Dialog
          open={!!disablingUser}
          onOpenChange={(open) => (!open ? setDisablingUser(null) : null)}
        >
          <DialogContent aria-label={t("users.disableUser")}>
            <DialogHeader>
              <DialogTitle>{t("users.disableUser")}</DialogTitle>
              <DialogDescription>
                {t("users.disableConfirm", { username: disablingUser?.username })}
              </DialogDescription>
            </DialogHeader>

            {disableMutation.isError ? (
              <p className="text-sm text-red-600">
                {disableMutation.error instanceof ApiError
                  ? disableMutation.error.message
                  : t("users.disableFailed")}
              </p>
            ) : null}

            <DialogFooter>
              <Button variant="outline" onClick={() => setDisablingUser(null)}>
                {t("common.cancel")}
              </Button>
              <Button
                variant="destructive"
                onClick={() => {
                  if (!disablingUser) return
                  disableMutation.mutate(disablingUser.id)
                }}
                disabled={disableMutation.isPending}
              >
                {disableMutation.isPending ? t("common.disabling") : t("common.confirm")}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        <Dialog
          open={!!deletingUser}
          onOpenChange={(open) => (!open ? setDeletingUser(null) : null)}
        >
          <DialogContent aria-label={t("users.deleteUser")}>
            <DialogHeader>
              <DialogTitle>{t("users.deleteUser")}</DialogTitle>
              <DialogDescription>
                {t("users.deleteConfirm", { username: deletingUser?.username })}
              </DialogDescription>
            </DialogHeader>

            {deleteMutation.isError ? (
              <p className="text-sm text-red-600">
                {deleteMutation.error instanceof ApiError
                  ? deleteMutation.error.message
                  : t("users.deleteFailed")}
              </p>
            ) : null}

            <DialogFooter>
              <Button variant="outline" onClick={() => setDeletingUser(null)}>
                {t("common.cancel")}
              </Button>
              <Button
                variant="destructive"
                onClick={() => {
                  if (!deletingUser) return
                  deleteMutation.mutate(deletingUser.id)
                }}
                disabled={deleteMutation.isPending}
              >
                {deleteMutation.isPending ? t("common.deleting") : t("common.confirm")}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </section>
    </div>
  )
}
