import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useEffect, useMemo, useState } from "react"
import { useTranslation } from "react-i18next"
import { ArrowLeft, ArrowRight, MoreHorizontal, Pencil, Search, Trash2 } from "lucide-react"

import { AsyncButton } from "@/components/ui/async-button"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
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
import { listGroupUsers, replaceGroupUsers } from "@/lib/api/group-users"
import { createGroup, deleteGroup, listGroups, updateGroup } from "@/lib/api/groups"
import { tableColumnSpacing } from "@/lib/table-spacing"
import { listUsers } from "@/lib/api/users"
import type { Group, UserStatus } from "@/lib/api/types"

type EditState = {
  mode: "create" | "edit"
  group: Group
  name: string
  description: string
  initialUserIDs: number[]
  memberIDs: number[]
  membersLoadedFromServer: boolean
}

const defaultNewGroup: Group = {
  id: 0,
  name: "",
  description: "",
  member_count: 0,
}

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

function sortedUniqueIDs(ids: number[]): number[] {
  const uniq = Array.from(new Set(ids.filter((id) => id > 0)))
  uniq.sort((a, b) => a - b)
  return uniq
}

function toggleSelectedIDs(ids: number[], id: number, checked: boolean): number[] {
  if (checked) {
    return sortedUniqueIDs([...ids, id])
  }
  return ids.filter((item) => item !== id)
}

function sameIDs(a: number[], b: number[]): boolean {
  const sa = sortedUniqueIDs(a)
  const sb = sortedUniqueIDs(b)
  return sa.join(",") === sb.join(",")
}

export function GroupsPage() {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const spacing = tableColumnSpacing.four

  const [upserting, setUpserting] = useState<EditState | null>(null)
  const [memberSearch, setMemberSearch] = useState("")
  const [candidateSearch, setCandidateSearch] = useState("")
  const [selectedMemberIDs, setSelectedMemberIDs] = useState<number[]>([])
  const [selectedCandidateIDs, setSelectedCandidateIDs] = useState<number[]>([])

  const queryParams = useMemo(() => ({ limit: 50, offset: 0 }), [])

  const groupsQuery = useQuery({
    queryKey: ["groups", queryParams],
    queryFn: () => listGroups(queryParams),
  })

  const usersQuery = useQuery({
    queryKey: ["users", "groups-page", { limit: 1000, offset: 0 }],
    queryFn: () => listUsers({ limit: 1000, offset: 0 }),
    enabled: !!upserting,
  })

  const groupUsersQuery = useQuery({
    queryKey: ["group-users", upserting?.group.id ?? 0],
    queryFn: () => listGroupUsers(upserting?.group.id ?? 0),
    enabled: !!upserting && upserting.mode === "edit" && upserting.group.id > 0,
  })

  useEffect(() => {
    if (!upserting || upserting.mode !== "edit") return
    if (!groupUsersQuery.data) return
    if (upserting.membersLoadedFromServer) return

    const ids = sortedUniqueIDs(groupUsersQuery.data.map((u) => u.id))
    setUpserting((prev) => {
      if (!prev || prev.mode !== "edit") return prev
      return {
        ...prev,
        initialUserIDs: ids,
        memberIDs: ids,
        membersLoadedFromServer: true,
      }
    })
    setSelectedMemberIDs([])
    setSelectedCandidateIDs([])
  }, [upserting, groupUsersQuery.data])

  const saveMutation = useMutation({
    mutationFn: async (input: {
      mode: "create" | "edit"
      groupID: number
      name: string
      description: string
      userIDs: number[]
    }) => {
      const userIDs = sortedUniqueIDs(input.userIDs)
      if (input.mode === "create") {
        const created = await createGroup({ name: input.name, description: input.description })
        await replaceGroupUsers(created.id, { user_ids: userIDs })
        return
      }

      await updateGroup(input.groupID, { name: input.name, description: input.description })
      await replaceGroupUsers(input.groupID, { user_ids: userIDs })
    },
    onSuccess: async () => {
      setUpserting(null)
      await qc.invalidateQueries({ queryKey: ["groups"] })
      await qc.invalidateQueries({ queryKey: ["users"] })
      await qc.invalidateQueries({ queryKey: ["group-users"] })
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: number) => deleteGroup(id),
    onSuccess: async () => {
      await qc.invalidateQueries({ queryKey: ["groups"] })
    },
  })

  const allUsers = usersQuery.data ?? []

  const memberSet = useMemo(
    () => new Set(upserting?.memberIDs ?? []),
    [upserting?.memberIDs],
  )

  const memberUsers = useMemo(
    () => allUsers.filter((u) => memberSet.has(u.id)),
    [allUsers, memberSet],
  )

  const candidateUsers = useMemo(
    () => allUsers.filter((u) => !memberSet.has(u.id)),
    [allUsers, memberSet],
  )

  const filteredMemberUsers = useMemo(() => {
    const keyword = memberSearch.trim().toLowerCase()
    if (!keyword) return memberUsers
    return memberUsers.filter((u) => u.username.toLowerCase().includes(keyword))
  }, [memberUsers, memberSearch])

  const filteredCandidateUsers = useMemo(() => {
    const keyword = candidateSearch.trim().toLowerCase()
    if (!keyword) return candidateUsers
    return candidateUsers.filter((u) => u.username.toLowerCase().includes(keyword))
  }, [candidateUsers, candidateSearch])

  const membersReady = !upserting || upserting.mode === "create" || upserting.membersLoadedFromServer
  const hasMemberChanges = !upserting
    ? false
    : !sameIDs(upserting.initialUserIDs, upserting.memberIDs)

  const moveCandidatesToMembers = () => {
    if (!upserting || selectedCandidateIDs.length === 0) return
    setUpserting((prev) => {
      if (!prev) return prev
      return {
        ...prev,
        memberIDs: sortedUniqueIDs([...prev.memberIDs, ...selectedCandidateIDs]),
      }
    })
    setSelectedCandidateIDs([])
  }

  const moveMembersToCandidates = () => {
    if (!upserting || selectedMemberIDs.length === 0) return
    const removeSet = new Set(selectedMemberIDs)
    setUpserting((prev) => {
      if (!prev) return prev
      return {
        ...prev,
        memberIDs: prev.memberIDs.filter((id) => !removeSet.has(id)),
      }
    })
    setSelectedMemberIDs([])
  }

  const openCreateDialog = () => {
    saveMutation.reset()
    setMemberSearch("")
    setCandidateSearch("")
    setSelectedMemberIDs([])
    setSelectedCandidateIDs([])
    setUpserting({
      mode: "create",
      group: defaultNewGroup,
      name: "",
      description: "",
      initialUserIDs: [],
      memberIDs: [],
      membersLoadedFromServer: true,
    })
  }

  const openEditDialog = (group: Group) => {
    saveMutation.reset()
    setMemberSearch("")
    setCandidateSearch("")
    setSelectedMemberIDs([])
    setSelectedCandidateIDs([])
    setUpserting({
      mode: "edit",
      group,
      name: group.name,
      description: group.description,
      initialUserIDs: [],
      memberIDs: [],
      membersLoadedFromServer: false,
    })
  }

  return (
    <div className="px-4 lg:px-6">
      <section className="space-y-6">
        <header className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
          <div>
            <h1 className="text-2xl font-semibold tracking-tight">{t("groups.title")}</h1>
            <p className="text-sm text-muted-foreground">{t("groups.subtitle")}</p>
          </div>
          <Button onClick={openCreateDialog}>{t("groups.createGroup")}</Button>
        </header>

        <Card>
          <CardHeader className="pb-3">
            <div className="flex flex-col gap-1.5">
              <CardTitle className="text-base">{t("groups.list")}</CardTitle>
              <CardDescription>
                {groupsQuery.isLoading ? t("common.loading") : null}
                {groupsQuery.isError ? t("common.loadFailed") : null}
                {groupsQuery.data ? t("groups.count", { count: groupsQuery.data.length }) : null}
              </CardDescription>
            </div>
          </CardHeader>
          <CardContent className="p-0">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className={spacing.headFirst}>{t("common.name")}</TableHead>
                  <TableHead className={spacing.headMiddle}>{t("groups.membersCount")}</TableHead>
                  <TableHead className={spacing.headMiddle}>{t("common.description")}</TableHead>
                  <TableHead className={`w-12 ${spacing.headLast}`}>
                    <span className="sr-only">{t("common.actions")}</span>
                  </TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {groupsQuery.isLoading ? (
                  <>
                    {Array.from({ length: 5 }).map((_, i) => (
                      <TableRow key={i}>
                        <TableCell className={spacing.cellFirst}>
                          <Skeleton className="h-4 w-28" />
                        </TableCell>
                        <TableCell className={spacing.cellMiddle}>
                          <Skeleton className="h-4 w-10" />
                        </TableCell>
                        <TableCell className={spacing.cellMiddle}>
                          <Skeleton className="h-4 w-56" />
                        </TableCell>
                        <TableCell className={spacing.cellLast}>
                          <Skeleton className="h-8 w-8" />
                        </TableCell>
                      </TableRow>
                    ))}
                  </>
                ) : null}
                {groupsQuery.data?.map((group) => (
                  <TableRow key={group.id}>
                    <TableCell className={`${spacing.cellFirst} font-medium`}>{group.name}</TableCell>
                    <TableCell className={`${spacing.cellMiddle} text-muted-foreground tabular-nums`}>{group.member_count ?? 0}</TableCell>
                    <TableCell className={`${spacing.cellMiddle} text-muted-foreground`}>{group.description}</TableCell>
                    <TableCell className={spacing.cellLast}>
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" size="icon" className="size-8">
                            <MoreHorizontal className="size-4" />
                            <span className="sr-only">{t("common.actions")}</span>
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem onClick={() => openEditDialog(group)}>
                            <Pencil className="mr-2 size-4" />
                            {t("common.edit")}
                          </DropdownMenuItem>
                          <DropdownMenuSeparator />
                          <DropdownMenuItem
                            variant="destructive"
                            disabled={deleteMutation.isPending}
                            onClick={() => deleteMutation.mutate(group.id)}
                          >
                            <Trash2 className="mr-2 size-4" />
                            {t("common.delete")}
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </TableCell>
                  </TableRow>
                ))}
                {!groupsQuery.isLoading && groupsQuery.data && groupsQuery.data.length === 0 ? (
                  <TableRow>
                    <TableCell className={`${spacing.cellFirst} py-8 text-center text-muted-foreground`} colSpan={4}>
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
          onOpenChange={(open) => {
            if (!open) {
              setUpserting(null)
              setSelectedMemberIDs([])
              setSelectedCandidateIDs([])
              setMemberSearch("")
              setCandidateSearch("")
              saveMutation.reset()
            }
          }}
        >
          <DialogContent className="sm:max-w-5xl">
            <DialogHeader>
              <DialogTitle>
                {upserting?.mode === "create" ? t("groups.createGroup") : t("groups.editGroup")}
              </DialogTitle>
              <DialogDescription>
                {upserting?.mode === "edit" ? upserting.group.name : t("groups.createGroup")}
              </DialogDescription>
            </DialogHeader>

            {upserting ? (
              <div className="space-y-4">
                <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                  <div className="space-y-1">
                    <Label htmlFor="group-name" className="text-sm text-slate-700">
                      {t("groups.name")}
                    </Label>
                    <Input
                      id="group-name"
                      value={upserting.name}
                      onChange={(e) =>
                        setUpserting((prev) => (prev ? { ...prev, name: e.target.value } : prev))
                      }
                      placeholder={t("groups.namePlaceholder")}
                      autoFocus={upserting.mode === "create"}
                    />
                  </div>
                  <div className="space-y-1">
                    <Label htmlFor="group-desc" className="text-sm text-slate-700">
                      {t("groups.description")}
                    </Label>
                    <Input
                      id="group-desc"
                      value={upserting.description}
                      onChange={(e) =>
                        setUpserting((prev) => (prev ? { ...prev, description: e.target.value } : prev))
                      }
                      placeholder={t("groups.descriptionPlaceholder")}
                    />
                  </div>
                </div>

                <div className="space-y-2">
                  <Label className="text-sm text-slate-700">{t("groups.manageMembers")}</Label>
                  <div className="grid grid-cols-1 gap-4 rounded-md border p-3 lg:grid-cols-[1fr_auto_1fr]">
                    <div className="flex min-h-0 flex-col rounded-md border">
                      <div className="border-b p-3">
                        <div className="text-sm font-medium">{t("groups.currentMembers")}</div>
                        <div className="mt-2 relative">
                          <Search className="pointer-events-none absolute left-2 top-2.5 size-4 text-muted-foreground" />
                          <Input
                            id="groups-members-search"
                            value={memberSearch}
                            onChange={(e) => setMemberSearch(e.target.value)}
                            className="pl-8"
                            placeholder={t("users.searchPlaceholder")}
                            aria-label={t("groups.currentMembers")}
                          />
                        </div>
                      </div>
                      <div className="min-h-0 max-h-64 flex-1 overflow-y-auto p-2">
                        {usersQuery.isLoading || (!membersReady && upserting.mode === "edit") ? (
                          Array.from({ length: 5 }).map((_, i) => (
                            <div key={i} className="p-2">
                              <Skeleton className="h-8 w-full" />
                            </div>
                          ))
                        ) : filteredMemberUsers.length === 0 ? (
                          <div className="px-2 py-6 text-sm text-muted-foreground">{t("common.noData")}</div>
                        ) : (
                          filteredMemberUsers.map((user) => (
                            <label key={user.id} className="flex cursor-pointer items-center justify-between gap-2 rounded-md px-2 py-2 hover:bg-muted/50">
                              <div className="flex items-center gap-2">
                                <Checkbox
                                  checked={selectedMemberIDs.includes(user.id)}
                                  onCheckedChange={(checked) => {
                                    setSelectedMemberIDs((prev) => toggleSelectedIDs(prev, user.id, checked === true))
                                  }}
                                />
                                <span className="text-sm">{user.username}</span>
                              </div>
                              <StatusBadge status={user.status} />
                            </label>
                          ))
                        )}
                      </div>
                    </div>

                    <div className="flex flex-row items-center justify-center gap-2 lg:flex-col">
                      <Button
                        type="button"
                        size="icon"
                        variant="outline"
                        onClick={moveCandidatesToMembers}
                        disabled={selectedCandidateIDs.length === 0}
                        title={t("groups.addSelected")}
                        aria-label={t("groups.addSelected")}
                      >
                        <ArrowLeft className="size-4" />
                      </Button>
                      <Button
                        type="button"
                        size="icon"
                        variant="outline"
                        onClick={moveMembersToCandidates}
                        disabled={selectedMemberIDs.length === 0}
                        title={t("groups.removeSelected")}
                        aria-label={t("groups.removeSelected")}
                      >
                        <ArrowRight className="size-4" />
                      </Button>
                    </div>

                    <div className="flex min-h-0 flex-col rounded-md border">
                      <div className="border-b p-3">
                        <div className="text-sm font-medium">{t("groups.availableUsers")}</div>
                        <div className="mt-2 relative">
                          <Search className="pointer-events-none absolute left-2 top-2.5 size-4 text-muted-foreground" />
                          <Input
                            id="groups-candidates-search"
                            value={candidateSearch}
                            onChange={(e) => setCandidateSearch(e.target.value)}
                            className="pl-8"
                            placeholder={t("users.searchPlaceholder")}
                            aria-label={t("groups.availableUsers")}
                          />
                        </div>
                      </div>
                      <div className="min-h-0 max-h-64 flex-1 overflow-y-auto p-2">
                        {usersQuery.isLoading || (!membersReady && upserting.mode === "edit") ? (
                          Array.from({ length: 5 }).map((_, i) => (
                            <div key={i} className="p-2">
                              <Skeleton className="h-8 w-full" />
                            </div>
                          ))
                        ) : filteredCandidateUsers.length === 0 ? (
                          <div className="px-2 py-6 text-sm text-muted-foreground">{t("common.noData")}</div>
                        ) : (
                          filteredCandidateUsers.map((user) => (
                            <label key={user.id} className="flex cursor-pointer items-center justify-between gap-2 rounded-md px-2 py-2 hover:bg-muted/50">
                              <div className="flex items-center gap-2">
                                <Checkbox
                                  checked={selectedCandidateIDs.includes(user.id)}
                                  onCheckedChange={(checked) => {
                                    setSelectedCandidateIDs((prev) => toggleSelectedIDs(prev, user.id, checked === true))
                                  }}
                                />
                                <span className="text-sm">{user.username}</span>
                              </div>
                              <StatusBadge status={user.status} />
                            </label>
                          ))
                        )}
                      </div>
                    </div>
                  </div>
                </div>

                <div className="text-sm text-amber-700">
                  {saveMutation.isError
                    ? (saveMutation.error instanceof ApiError
                        ? saveMutation.error.message
                        : t("groups.saveFailed"))
                    : null}
                </div>
              </div>
            ) : null}

            <DialogFooter>
              <Button
                variant="outline"
                onClick={() => setUpserting(null)}
                disabled={saveMutation.isPending}
              >
                {t("common.cancel")}
              </Button>
              <AsyncButton
                onClick={() => {
                  if (!upserting) return
                  const name = upserting.name.trim()
                  const description = upserting.description.trim()
                  if (!name || !membersReady) return
                  saveMutation.mutate({
                    mode: upserting.mode,
                    groupID: upserting.group.id,
                    name,
                    description,
                    userIDs: upserting.memberIDs,
                  })
                }}
                disabled={
                  saveMutation.isPending
                  || usersQuery.isLoading
                  || !membersReady
                  || !upserting?.name.trim()
                  || (upserting?.mode === "edit" && !hasMemberChanges && upserting.name.trim() === upserting.group.name && upserting.description.trim() === upserting.group.description)
                }
                pending={saveMutation.isPending}
                pendingText={t("common.saving")}
              >
                {t("common.save")}
              </AsyncButton>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </section>
    </div>
  )
}
