import { useQuery } from "@tanstack/react-query";
import { useState, useMemo } from "react";
import { useTranslation } from "react-i18next";
import { useSearchParams } from "react-router-dom";
import { toast } from "sonner";
import { Copy, Check, ExternalLink, Info } from "lucide-react";

import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/page-header";
import { TableEmptyState } from "@/components/table-empty-state";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
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
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { Badge } from "@/components/ui/badge";
import { listUsers } from "@/lib/api/users";
import { getSystemSettings } from "@/lib/api/system";
import type { User, UserStatus } from "@/lib/api/types";
import { tableColumnSpacing } from "@/lib/table-spacing";
import { tableTransitionClass } from "@/lib/table-motion";
import { useTableQueryTransition } from "@/lib/table-query-transition";
import { buildUserListSearchParams, parseUserListSearchParams } from "@/lib/user-list-filters";
import { tableToolbarClass } from "@/lib/table-toolbar";

type StatusFilter = UserStatus | "all";

function resolveSubscriptionBaseURL(configured?: string): string {
  const value = configured?.trim() ?? "";
  if (value) return value;
  return window.location.origin;
}

function getSubscriptionUrl(userUuid: string, format?: string, configuredBaseURL?: string): string {
  const baseURL = resolveSubscriptionBaseURL(configuredBaseURL);
  const normalizedBase = baseURL.endsWith("/") ? baseURL : `${baseURL}/`;
  const url = new URL(`api/sub/${userUuid}`, normalizedBase);
  if (format) {
    url.searchParams.set("format", format);
  }
  return url.toString();
}

function CopyButton({ text, label }: { text: string; label?: string }) {
  const { t } = useTranslation();
  const [copied, setCopied] = useState(false);
  const accessibleLabel = label ? `${label} ${t("common.copy")}` : t("common.copy");

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(text);
      setCopied(true);
      toast.success(label ? `${label} ${t("common.copied")}` : t("common.copiedToClipboard"));
      setTimeout(() => setCopied(false), 2000);
    } catch {
      toast.error(t("common.copyFailed"));
    }
  };

  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <Button
          type="button"
          variant="ghost"
          size="icon"
          className="h-7 w-7 shrink-0"
          onClick={handleCopy}
          aria-label={accessibleLabel}
        >
          {copied ? <Check className="h-4 w-4 text-green-600" /> : <Copy className="h-4 w-4" />}
        </Button>
      </TooltipTrigger>
      <TooltipContent>{accessibleLabel}</TooltipContent>
    </Tooltip>
  );
}

function StatusBadge({ status }: { status: UserStatus }) {
  const { t } = useTranslation();
  const variants: Record<UserStatus, "default" | "destructive" | "secondary" | "outline"> = {
    active: "default",
    disabled: "secondary",
    expired: "destructive",
    traffic_exceeded: "outline",
  };

  const label =
    status === "traffic_exceeded" ? t("users.status.trafficExceeded") : t(`users.status.${status}`);

  return <Badge variant={variants[status]}>{label}</Badge>;
}

export function SubscriptionsPage() {
  const { t } = useTranslation();
  const spacing = tableColumnSpacing.four;
  const [searchParams, setSearchParams] = useSearchParams();
  const filters = useMemo(() => parseUserListSearchParams(searchParams, "active"), [searchParams]);
  const statusFilter = filters.statusFilter as StatusFilter;
  const search = filters.search;

  const updateFilters = (patch: Partial<{ statusFilter: StatusFilter; search: string }>) => {
    const next = {
      statusFilter: (patch.statusFilter ?? statusFilter) as StatusFilter,
      search: patch.search ?? search,
    };
    setSearchParams(buildUserListSearchParams(next, "active"), { replace: true });
  };

  const statusOptions: Array<{ value: StatusFilter; label: string }> = [
    { value: "all", label: t("common.all") },
    { value: "active", label: t("users.status.active") },
    { value: "disabled", label: t("users.status.disabled") },
    { value: "expired", label: t("users.status.expired") },
    { value: "traffic_exceeded", label: t("users.status.trafficExceeded") },
  ];

  const queryParams = useMemo(
    () => ({
      limit: 100,
      offset: 0,
      status: statusFilter === "all" ? undefined : statusFilter,
    }),
    [statusFilter],
  );

  const usersQuery = useQuery({
    queryKey: ["users", queryParams],
    queryFn: () => listUsers(queryParams),
  });

  const systemSettingsQuery = useQuery({
    queryKey: ["system-settings"],
    queryFn: getSystemSettings,
  });

  const usersTable = useTableQueryTransition({
    filterKey: statusFilter,
    rows: usersQuery.data,
    isLoading: usersQuery.isLoading,
    isFetching: usersQuery.isFetching,
    isError: usersQuery.isError,
  });

  const filteredUsers = useMemo(() => {
    const users = usersTable.visibleRows;
    if (!search.trim()) return users;
    const lowerSearch = search.toLowerCase();
    return users.filter(
      (u) =>
        u.username.toLowerCase().includes(lowerSearch) ||
        u.uuid.toLowerCase().includes(lowerSearch),
    );
  }, [usersTable.visibleRows, search]);

  return (
    <div className="px-4 lg:px-6 space-y-6">
      <PageHeader title={t("subscriptions.title")} description={t("subscriptions.subtitle")} />

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            {t("subscriptions.behaviorTitle")}
            <Tooltip>
              <TooltipTrigger asChild>
                <button
                  type="button"
                  className="inline-flex items-center"
                  aria-label={t("subscriptions.behaviorTitle")}
                >
                  <Info className="h-4 w-4 text-slate-400" />
                </button>
              </TooltipTrigger>
              <TooltipContent className="max-w-xs">
                <p>{t("subscriptions.behaviorTooltip")}</p>
              </TooltipContent>
            </Tooltip>
          </CardTitle>
          <CardDescription>{t("subscriptions.behaviorDescription")}</CardDescription>
        </CardHeader>
        <CardContent className="space-y-3 text-sm">
          <div className="grid gap-2">
            <div className="flex items-start gap-2">
              <Badge variant="outline" className="shrink-0 mt-0.5">
                ?format=singbox
              </Badge>
              <span className="text-slate-600">{t("subscriptions.ruleSingbox")}</span>
            </div>
            <div className="flex items-start gap-2">
              <Badge variant="outline" className="shrink-0 mt-0.5">
                ?format=v2ray
              </Badge>
              <span className="text-slate-600">{t("subscriptions.ruleV2ray")}</span>
            </div>
            <div className="flex items-start gap-2">
              <Badge variant="secondary" className="shrink-0 mt-0.5">
                {t("subscriptions.uaMatchLabel")}
              </Badge>
              <span className="text-slate-600">{t("subscriptions.ruleUaMatch")}</span>
            </div>
            <div className="flex items-start gap-2">
              <Badge variant="secondary" className="shrink-0 mt-0.5">
                {t("subscriptions.uaOtherLabel")}
              </Badge>
              <span className="text-slate-600">{t("subscriptions.ruleUaOther")}</span>
            </div>
          </div>
        </CardContent>
      </Card>

      <div className={tableToolbarClass.container}>
        <div className={tableToolbarClass.filters}>
          <Label htmlFor="subscriptions-search" className="sr-only">
            {t("subscriptions.searchPlaceholder")}
          </Label>
          <Input
            id="subscriptions-search"
            placeholder={t("subscriptions.searchPlaceholder")}
            className="w-full sm:w-64"
            value={search}
            onChange={(e) => updateFilters({ search: e.target.value })}
          />
          <Select
            value={statusFilter}
            onValueChange={(v) => updateFilters({ statusFilter: v as StatusFilter })}
          >
            <SelectTrigger className="w-40" aria-label={t("common.status")}>
              <SelectValue />
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

      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className={spacing.headFirst}>{t("users.username")}</TableHead>
              <TableHead className={spacing.headMiddle}>{t("common.status")}</TableHead>
              <TableHead className={spacing.headMiddle}>
                {t("subscriptions.subscriptionUrl")}
              </TableHead>
              <TableHead className={`${spacing.headLast} w-[140px] text-right`}>
                {t("common.actions")}
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody className={tableTransitionClass(usersTable.isTransitioning)}>
            {usersTable.showNoData || filteredUsers.length === 0 ? (
              <TableEmptyState
                colSpan={4}
                className={`${spacing.cellFirst} py-10 text-center`}
                message={t("common.noData")}
                actionLabel={t("users.createUser")}
                actionTo="/users"
              />
            ) : (
              filteredUsers.map((user) => (
                <UserSubscriptionRow
                  key={user.id}
                  user={user}
                  subscriptionBaseURL={systemSettingsQuery.data?.subscription_base_url}
                />
              ))
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}

function UserSubscriptionRow({
  user,
  subscriptionBaseURL,
}: {
  user: User;
  subscriptionBaseURL?: string;
}) {
  const { t } = useTranslation();
  const spacing = tableColumnSpacing.four;
  const subUrl = getSubscriptionUrl(user.uuid, undefined, subscriptionBaseURL);
  const singboxUrl = getSubscriptionUrl(user.uuid, "singbox", subscriptionBaseURL);

  return (
    <TableRow>
      <TableCell className={`${spacing.cellFirst} font-medium`}>{user.username}</TableCell>
      <TableCell className={spacing.cellMiddle}>
        <StatusBadge status={user.status} />
      </TableCell>
      <TableCell className={spacing.cellMiddle}>
        <div className="flex items-center gap-2 max-w-md">
          <code className="flex-1 truncate text-xs bg-slate-100 px-2 py-1 rounded font-mono">
            {subUrl}
          </code>
          <CopyButton text={subUrl} label={t("subscriptions.subscriptionUrl")} />
        </div>
      </TableCell>
      <TableCell className={`${spacing.cellLast} text-right`}>
        <div className="flex items-center justify-end gap-1">
          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                type="button"
                variant="ghost"
                size="icon"
                className="h-7 w-7"
                onClick={() => window.open(singboxUrl, "_blank")}
                aria-label={t("subscriptions.previewSingbox")}
              >
                <ExternalLink className="h-4 w-4" />
              </Button>
            </TooltipTrigger>
            <TooltipContent>{t("subscriptions.previewSingbox")}</TooltipContent>
          </Tooltip>
          <CopyButton text={singboxUrl} label={t("subscriptions.copySingboxUrl")} />
        </div>
      </TableCell>
    </TableRow>
  );
}
