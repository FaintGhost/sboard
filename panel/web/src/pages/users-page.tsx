import { useQuery } from "@tanstack/react-query";
import { useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { MoreHorizontal, Pencil, Ban, Search, Trash2 } from "lucide-react";
import { useSearchParams } from "react-router-dom";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { PageHeader } from "@/components/page-header";
import { TableEmptyState } from "@/components/table-empty-state";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { listUsers } from "@/lib/api/users";
import { listGroups } from "@/lib/api/groups";
import type { User, UserStatus } from "@/lib/api/types";
import { tableColumnLayout, tableColumnSpacing } from "@/lib/table-spacing";
import { tableTransitionClass } from "@/lib/table-motion";
import { useTableQueryTransition } from "@/lib/table-query-transition";
import { bytesToGBString } from "@/lib/units";
import { buildUserListSearchParams, parseUserListSearchParams } from "@/lib/user-list-filters";
import { tableToolbarClass } from "@/lib/table-toolbar";
import { formatDateYMDByTimezone } from "@/lib/datetime";
import { useSystemStore } from "@/store/system";

import {
  DisableUserDialog,
  DeleteUserDialog,
  EditUserDialog,
  buildUpdatePayload,
  useUserMutations,
  defaultNewUser,
  type EditState,
  type StatusFilter,
} from "./users";

function StatusBadge({ status }: { status: string }) {
  const { t } = useTranslation();
  const variant =
    status === "active" ? "default" : status === "disabled" ? "secondary" : "destructive";
  const label =
    status === "traffic_exceeded" ? t("users.status.trafficExceeded") : t(`users.status.${status}`);
  return <Badge variant={variant}>{label}</Badge>;
}

function formatTraffic(
  used: number,
  limit: number,
  t: (key: string, options?: Record<string, unknown>) => string,
): string {
  const usedGB = bytesToGBString(used);
  if (limit === 0) return t("users.trafficUnlimited", { used: usedGB });
  const limitGB = bytesToGBString(limit);
  return t("users.trafficFormat", { used: usedGB, limit: limitGB });
}

function formatExpireDate(
  expireAt: string | null,
  t: (key: string) => string,
  timezone: string,
): string {
  if (!expireAt) return t("common.permanent");
  const formatted = formatDateYMDByTimezone(expireAt, timezone, "");
  if (!formatted) return t("common.permanent");
  return formatted;
}

export function UsersPage() {
  const { t } = useTranslation();
  const timezone = useSystemStore((state) => state.timezone);
  const [searchParams, setSearchParams] = useSearchParams();
  const filters = useMemo(() => parseUserListSearchParams(searchParams, "all"), [searchParams]);
  const status = filters.statusFilter;
  const search = filters.search;
  const [upserting, setUpserting] = useState<EditState | null>(null);
  const [disablingUser, setDisablingUser] = useState<User | null>(null);
  const [deletingUser, setDeletingUser] = useState<User | null>(null);
  const spacing = tableColumnSpacing.five;
  const layout = tableColumnLayout.sixActionIcon;

  const updateFilters = (patch: Partial<{ statusFilter: StatusFilter; search: string }>) => {
    const next = {
      statusFilter: (patch.statusFilter ?? status) as StatusFilter,
      search: patch.search ?? search,
    };
    setSearchParams(buildUserListSearchParams(next, "all"), { replace: true });
  };

  const { createMutation, updateMutation, saveGroupsMutation, disableMutation, deleteMutation } =
    useUserMutations();

  const statusOptions: Array<{ value: StatusFilter; label: string }> = [
    { value: "all", label: t("common.all") },
    { value: "active", label: t("users.status.active") },
    { value: "disabled", label: t("users.status.disabled") },
    { value: "expired", label: t("users.status.expired") },
    { value: "traffic_exceeded", label: t("users.status.trafficExceeded") },
  ];

  const queryParams = useMemo(
    () => ({
      limit: 50,
      offset: 0,
      status: status === "all" ? undefined : status,
    }),
    [status],
  );

  const usersQuery = useQuery({
    queryKey: ["users", queryParams],
    queryFn: () => listUsers(queryParams),
  });

  const usersTable = useTableQueryTransition({
    filterKey: status,
    rows: usersQuery.data,
    isLoading: usersQuery.isLoading,
    isFetching: usersQuery.isFetching,
    isError: usersQuery.isError,
  });

  const groupsQuery = useQuery({
    queryKey: ["groups", { limit: 200, offset: 0 }],
    queryFn: () => listGroups({ limit: 200, offset: 0 }),
  });

  // Filter users by search keyword
  const filteredUsers = useMemo(() => {
    const users = usersTable.visibleRows;
    if (!search.trim()) return users;
    const keyword = search.trim().toLowerCase();
    return users.filter((u) => u.username.toLowerCase().includes(keyword));
  }, [usersTable.visibleRows, search]);

  const groupNameByID = useMemo(() => {
    const map = new Map<number, string>();
    for (const group of groupsQuery.data ?? []) {
      map.set(group.id, group.name);
    }
    return map;
  }, [groupsQuery.data]);

  const openCreateDialog = () => {
    createMutation.reset();
    updateMutation.reset();
    saveGroupsMutation.reset();
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
    });
  };

  const openEditDialog = (u: User) => {
    const parsedExpire =
      u.expire_at && !Number.isNaN(Date.parse(u.expire_at)) ? new Date(u.expire_at) : null;
    createMutation.reset();
    updateMutation.reset();
    saveGroupsMutation.reset();
    setUpserting({
      mode: "edit",
      user: u,
      username: u.username,
      status: u.status as UserStatus,
      trafficLimit: bytesToGBString(u.traffic_limit ?? 0),
      trafficResetDay: u.traffic_reset_day ?? 0,
      expireDate: parsedExpire,
      clearExpireAt: false,
      groupIDs: [],
      groupsLoadedFromServer: false,
    });
  };

  const handleSave = async (state: EditState) => {
    if (state.mode === "create") {
      const created = await createMutation.mutateAsync({
        username: state.username.trim(),
      });
      const payload = buildUpdatePayload({
        ...state,
        mode: "edit",
        user: created,
      });
      await Promise.all([
        updateMutation.mutateAsync({ id: created.id, payload }),
        saveGroupsMutation.mutateAsync({ userId: created.id, groupIDs: state.groupIDs }),
      ]);
      setUpserting(null);
      return;
    }

    const payload = buildUpdatePayload(state);
    await Promise.all([
      updateMutation.mutateAsync({ id: state.user.id, payload }),
      saveGroupsMutation.mutateAsync({ userId: state.user.id, groupIDs: state.groupIDs }),
    ]);
    setUpserting(null);
  };

  const handleDisable = (userId: number) => {
    disableMutation.mutate(userId, {
      onSuccess: () => setDisablingUser(null),
    });
  };

  const handleDelete = (userId: number) => {
    deleteMutation.mutate(userId, {
      onSuccess: () => setDeletingUser(null),
    });
  };

  return (
    <div className="px-4 lg:px-6">
      <section className="space-y-6">
        <PageHeader
          title={t("users.title")}
          description={t("users.subtitle")}
          action={<Button onClick={openCreateDialog}>{t("users.createUser")}</Button>}
        />

        <Card>
          <CardHeader className="pb-3">
            <div className={tableToolbarClass.container}>
              <div className="flex flex-col gap-1.5">
                <CardTitle className="text-base">{t("users.list")}</CardTitle>
                <CardDescription>
                  {usersTable.showLoadingHint ? t("common.loading") : null}
                  {usersQuery.isError ? t("common.loadFailed") : null}
                  {!usersTable.showLoadingHint && usersQuery.data
                    ? t("users.count", { count: filteredUsers.length })
                    : null}
                </CardDescription>
              </div>
              <div className={tableToolbarClass.filters}>
                <div className={tableToolbarClass.searchWrap}>
                  <Label htmlFor="users-search" className="sr-only">
                    {t("users.searchPlaceholder")}
                  </Label>
                  <Search className="absolute left-2.5 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
                  <Input
                    id="users-search"
                    placeholder={t("users.searchPlaceholder")}
                    value={search}
                    onChange={(e) => updateFilters({ search: e.target.value })}
                    className="pl-8 w-full sm:w-48"
                  />
                </div>
                <Select
                  value={status}
                  onValueChange={(value) => updateFilters({ statusFilter: value as StatusFilter })}
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
            <Table className={layout.tableClass}>
              <TableHeader>
                <TableRow>
                  <TableHead className={`${spacing.headFirst} ${layout.headFirst}`}>
                    {t("users.username")}
                  </TableHead>
                  <TableHead className={`${spacing.headMiddle} ${layout.headMiddle}`}>
                    {t("users.groups")}
                  </TableHead>
                  <TableHead className={`${spacing.headMiddle} ${layout.headMiddle}`}>
                    {t("common.status")}
                  </TableHead>
                  <TableHead
                    className={`${spacing.headMiddle} ${layout.headMiddle} hidden md:table-cell`}
                  >
                    {t("users.traffic")}
                  </TableHead>
                  <TableHead
                    className={`${spacing.headMiddle} ${layout.headMiddle} hidden sm:table-cell`}
                  >
                    {t("users.expireDate")}
                  </TableHead>
                  <TableHead className={`${spacing.headLast} ${layout.headLast}`}>
                    <span className="sr-only">{t("common.actions")}</span>
                  </TableHead>
                </TableRow>
              </TableHeader>
              <TableBody className={tableTransitionClass(usersTable.isTransitioning)}>
                {filteredUsers.map((u) => {
                  const visibleGroupIDs = (u.group_ids ?? []).filter((groupID) =>
                    groupNameByID.has(groupID),
                  );
                  return (
                    <TableRow key={u.id}>
                      <TableCell className={`${spacing.cellFirst} font-medium`}>
                        {u.username}
                      </TableCell>
                      <TableCell className={spacing.cellMiddle}>
                        {visibleGroupIDs.length > 0 ? (
                          <div className="flex flex-wrap gap-1">
                            {visibleGroupIDs.map((groupID) => {
                              const groupName = groupNameByID.get(groupID);
                              return (
                                <Badge key={`${u.id}-${groupID}`} variant="secondary">
                                  {groupName}
                                </Badge>
                              );
                            })}
                          </div>
                        ) : (
                          <span className="text-muted-foreground">-</span>
                        )}
                      </TableCell>
                      <TableCell className={spacing.cellMiddle}>
                        <StatusBadge status={u.status} />
                      </TableCell>
                      <TableCell
                        className={`${spacing.cellMiddle} hidden md:table-cell text-muted-foreground`}
                      >
                        {formatTraffic(u.traffic_used, u.traffic_limit, t)}
                      </TableCell>
                      <TableCell
                        className={`${spacing.cellMiddle} hidden sm:table-cell text-muted-foreground`}
                      >
                        {formatExpireDate(u.expire_at, t, timezone)}
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
                            <DropdownMenuItem onClick={() => openEditDialog(u)}>
                              <Pencil className="mr-2 size-4" />
                              {t("common.edit")}
                            </DropdownMenuItem>
                            <DropdownMenuSeparator />
                            <DropdownMenuItem
                              variant="destructive"
                              disabled={u.status === "disabled"}
                              onClick={() => {
                                disableMutation.reset();
                                setDisablingUser(u);
                              }}
                            >
                              <Ban className="mr-2 size-4" />
                              {t("common.disable")}
                            </DropdownMenuItem>
                            <DropdownMenuItem
                              variant="destructive"
                              onClick={() => {
                                deleteMutation.reset();
                                setDeletingUser(u);
                              }}
                            >
                              <Trash2 className="mr-2 size-4" />
                              {t("common.delete")}
                            </DropdownMenuItem>
                          </DropdownMenuContent>
                        </DropdownMenu>
                      </TableCell>
                    </TableRow>
                  );
                })}
                {usersTable.showNoData || filteredUsers.length === 0 ? (
                  <TableEmptyState
                    colSpan={6}
                    message={t("common.noData")}
                    actionLabel={t("users.createUser")}
                    actionTo="/users"
                  />
                ) : null}
              </TableBody>
            </Table>
          </CardContent>
        </Card>

        <EditUserDialog
          editState={upserting}
          setEditState={setUpserting}
          onSave={handleSave}
          createMutation={createMutation}
          updateMutation={updateMutation}
          saveGroupsMutation={saveGroupsMutation}
        />

        <DisableUserDialog
          user={disablingUser}
          onClose={() => setDisablingUser(null)}
          onConfirm={handleDisable}
          isPending={disableMutation.isPending}
          isError={disableMutation.isError}
          error={disableMutation.error}
        />

        <DeleteUserDialog
          user={deletingUser}
          onClose={() => setDeletingUser(null)}
          onConfirm={handleDelete}
          isPending={deleteMutation.isPending}
          isError={deleteMutation.isError}
          error={deleteMutation.error}
        />
      </section>
    </div>
  );
}
