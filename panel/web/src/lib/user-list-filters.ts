import type { UserStatus } from "@/lib/api/types";

export type UserListStatusFilter = UserStatus | "all";

export type UserListFilters = {
  statusFilter: UserListStatusFilter;
  search: string;
};

const allowedStatusFilters: UserListStatusFilter[] = [
  "all",
  "active",
  "disabled",
  "expired",
  "traffic_exceeded",
];

function parseStatusFilter(
  raw: string | null,
  fallback: UserListStatusFilter,
): UserListStatusFilter {
  if (!raw) return fallback;
  if (allowedStatusFilters.includes(raw as UserListStatusFilter)) {
    return raw as UserListStatusFilter;
  }
  return fallback;
}

function parseSearch(raw: string | null): string {
  if (!raw) return "";
  return raw.trim();
}

export function parseUserListSearchParams(
  params: URLSearchParams,
  defaultStatusFilter: UserListStatusFilter,
): UserListFilters {
  return {
    statusFilter: parseStatusFilter(params.get("status"), defaultStatusFilter),
    search: parseSearch(params.get("search")),
  };
}

export function buildUserListSearchParams(
  filters: UserListFilters,
  defaultStatusFilter: UserListStatusFilter,
): URLSearchParams {
  const params = new URLSearchParams();

  if (filters.statusFilter !== defaultStatusFilter) {
    params.set("status", filters.statusFilter);
  }

  const keyword = filters.search.trim();
  if (keyword) {
    params.set("search", keyword);
  }

  return params;
}
