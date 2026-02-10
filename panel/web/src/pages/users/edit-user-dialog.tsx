import { useEffect } from "react"
import { useTranslation } from "react-i18next"
import { useQuery } from "@tanstack/react-query"
import { format } from "date-fns"

import { AsyncButton } from "@/components/ui/async-button"
import { Button } from "@/components/ui/button"
import { Calendar } from "@/components/ui/calendar"
import { Checkbox } from "@/components/ui/checkbox"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { FieldHint } from "@/components/ui/field-hint"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { ApiError } from "@/lib/api/client"
import { listGroups } from "@/lib/api/groups"
import { getUserGroups } from "@/lib/api/user-groups"
import type { UserStatus } from "@/lib/api/types"
import { gbStringToBytes, rfc3339FromDateOnlyUTC } from "@/lib/units"
import type { EditState } from "./types"

type EditUserDialogProps = {
  editState: EditState | null
  setEditState: React.Dispatch<React.SetStateAction<EditState | null>>
  onSave: (state: EditState) => Promise<void>
  createMutation: {
    isPending: boolean
    isError: boolean
    error: Error | null
    reset: () => void
  }
  updateMutation: {
    isPending: boolean
    isError: boolean
    error: Error | null
    reset: () => void
  }
  saveGroupsMutation: {
    isPending: boolean
    isError: boolean
    error: Error | null
    reset: () => void
  }
}

export function EditUserDialog({
  editState,
  setEditState,
  onSave,
  createMutation,
  updateMutation,
  saveGroupsMutation,
}: EditUserDialogProps) {
  const { t } = useTranslation()

  const editableStatusOptions: Array<{ value: UserStatus; label: string }> = [
    { value: "active", label: t("users.status.active") },
    { value: "disabled", label: t("users.status.disabled") },
    { value: "expired", label: t("users.status.expired") },
    { value: "traffic_exceeded", label: t("users.status.trafficExceeded") },
  ]

  const groupsQuery = useQuery({
    queryKey: ["groups", { limit: 200, offset: 0 }],
    queryFn: () => listGroups({ limit: 200, offset: 0 }),
  })

  const userGroupsQuery = useQuery({
    queryKey: ["user-groups", editState?.user.id ?? 0],
    queryFn: () => getUserGroups(editState?.user.id ?? 0),
    enabled: !!editState && editState.mode === "edit" && editState.user.id > 0,
  })

  useEffect(() => {
    if (!editState || editState.mode !== "edit") return
    if (!userGroupsQuery.data) return
    if (editState.groupsLoadedFromServer) return

    setEditState((prev) => {
      if (!prev || prev.mode !== "edit") return prev
      return {
        ...prev,
        groupIDs: (userGroupsQuery.data?.group_ids ?? []).slice().sort((a, b) => a - b),
        groupsLoadedFromServer: true,
      }
    })
  }, [editState, userGroupsQuery.data, setEditState])

  const handleSave = async () => {
    if (!editState) return
    await onSave(editState)
  }

  const isPending = createMutation.isPending || updateMutation.isPending || saveGroupsMutation.isPending

  const hasSaveError = createMutation.isError || updateMutation.isError || saveGroupsMutation.isError
  const saveError = createMutation.error ?? updateMutation.error ?? saveGroupsMutation.error

  return (
    <Dialog
      open={!!editState}
      onOpenChange={(open) => (!open ? setEditState(null) : null)}
    >
      <DialogContent aria-label={editState?.mode === "create" ? t("users.createUser") : t("users.editUser")}>
        <DialogHeader>
          <DialogTitle>
            {editState?.mode === "create" ? t("users.createUser") : t("users.editUser")}
          </DialogTitle>
          <DialogDescription>
            {editState?.mode === "edit" ? editState.user.username : t("users.createUser")}
          </DialogDescription>
        </DialogHeader>

        {editState ? (
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
            <div className="space-y-1 md:col-span-2">
              <Label className="text-sm text-slate-700" htmlFor="edit-username">
                {t("users.username")}
              </Label>
              <Input
                id="edit-username"
                value={editState.username}
                onChange={(e) =>
                  setEditState((prev) =>
                    prev ? { ...prev, username: e.target.value } : prev,
                  )
                }
                placeholder={t("users.usernamePlaceholder")}
                autoFocus={editState.mode === "create"}
              />
            </div>

            <div className="space-y-2 md:col-span-2">
              <div className="flex items-center gap-1">
                <Label className="text-sm text-slate-700">{t("users.groups")}</Label>
                <FieldHint label={t("users.groups")}>{t("users.groupsHint")}</FieldHint>
              </div>
              <div className="rounded-lg border border-slate-200 p-3">
                <div className="grid grid-cols-1 gap-2 md:grid-cols-2">
                  {groupsQuery.data?.map((g) => {
                    const checked = editState.groupIDs.includes(g.id)
                    return (
                      <label
                        key={g.id}
                        className="flex w-full cursor-pointer items-center gap-2 rounded-md px-2 py-2 text-sm text-slate-800 transition-colors hover:bg-slate-50"
                      >
                        <Checkbox
                          checked={checked}
                          onCheckedChange={(v) => {
                            const next = Boolean(v)
                            setEditState((p) => {
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
            </div>

            <>
              <div className="space-y-1">
                <Label className="text-sm text-slate-700">{t("common.status")}</Label>
                <Select
                  value={editState.status}
                  onValueChange={(value) =>
                    setEditState((prev) =>
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
                <div className="flex items-center gap-1">
                  <Label className="text-sm text-slate-700" htmlFor="edit-traffic-limit">
                    {t("users.trafficLimit")}
                  </Label>
                  <FieldHint label={t("users.trafficLimit")}>
                    {t("users.trafficLimitHint")}
                  </FieldHint>
                </div>
                <div className="relative">
                  <Input
                    id="edit-traffic-limit"
                    inputMode="decimal"
                    value={editState.trafficLimit}
                    onChange={(e) =>
                      setEditState((prev) =>
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
              </div>

              <div className="space-y-1">
                <div className="flex items-center gap-1">
                  <Label className="text-sm text-slate-700" htmlFor="edit-traffic-reset-day">
                    {t("users.trafficResetDay")}
                  </Label>
                  <FieldHint label={t("users.trafficResetDay")}>
                    {t("users.trafficResetDayHint")}
                  </FieldHint>
                </div>
                <Input
                  id="edit-traffic-reset-day"
                  type="number"
                  min={0}
                  max={31}
                  step={1}
                  inputMode="numeric"
                  value={String(editState.trafficResetDay)}
                  onChange={(e) => {
                    const v = Number(e.target.value)
                    setEditState((prev) =>
                      prev
                        ? {
                          ...prev,
                          trafficResetDay: Number.isFinite(v) ? v : 0,
                        }
                        : prev,
                    )
                  }}
                  onBlur={() =>
                    setEditState((prev) => {
                      if (!prev) return prev
                      const v = Math.trunc(prev.trafficResetDay)
                      const clamped = Math.min(31, Math.max(0, v))
                      return { ...prev, trafficResetDay: clamped }
                    })
                  }
                  aria-label={t("users.trafficResetDay")}
                />
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
                        {editState.expireDate ? (
                          format(editState.expireDate, "yyyy-MM-dd")
                        ) : (
                          <span className="text-slate-500">{t("users.selectDate")}</span>
                        )}
                      </Button>
                    </PopoverTrigger>
                    <PopoverContent className="w-auto p-0" align="start">
                      <Calendar
                        mode="single"
                        selected={editState.expireDate ?? undefined}
                        onSelect={(date) =>
                          setEditState((prev) =>
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
                      setEditState((prev) =>
                        prev
                          ? { ...prev, expireDate: null, clearExpireAt: true }
                          : prev,
                      )
                    }
                    disabled={editState.clearExpireAt}
                  >
                    {t("users.clearDate")}
                  </Button>
                </div>
              </div>
            </>
          </div>
        ) : null}

        {editState?.mode === "create" && hasSaveError ? (
          <p className="text-sm text-red-600">
            {saveError instanceof ApiError
              ? saveError.message
              : t("users.createFailed")}
          </p>
        ) : null}

        {editState?.mode === "edit" && hasSaveError ? (
          <p className="text-sm text-red-600">
            {saveError instanceof ApiError
              ? saveError.message
              : t("users.saveFailed")}
          </p>
        ) : null}

        <DialogFooter>
          <Button variant="outline" onClick={() => setEditState(null)}>
            {t("common.cancel")}
          </Button>
          <AsyncButton
            onClick={handleSave}
            disabled={
              !editState ||
              isPending ||
              !editState.username.trim()
            }
            pending={isPending}
            pendingText={
              editState?.mode === "create"
                ? t("common.creating")
                : t("common.saving")
            }
          >
            {editState?.mode === "create" ? t("common.create") : t("common.save")}
          </AsyncButton>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

export function buildUpdatePayload(editState: EditState): Record<string, unknown> {
  const payload: Record<string, unknown> = {}
  const username = editState.username.trim()
  if (username && username !== editState.user.username) {
    payload.username = username
  }
  payload.status = editState.status

  const bytes = gbStringToBytes(editState.trafficLimit)
  if (bytes !== null) payload.traffic_limit = bytes

  if (
    Number.isInteger(editState.trafficResetDay) &&
    editState.trafficResetDay >= 0 &&
    editState.trafficResetDay <= 31
  ) {
    payload.traffic_reset_day = editState.trafficResetDay
  }

  if (editState.clearExpireAt) {
    payload.expire_at = ""
  } else if (editState.expireDate) {
    payload.expire_at = rfc3339FromDateOnlyUTC(editState.expireDate)
  }

  return payload
}
