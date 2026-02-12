import { apiRequest } from "./client";
import { listAllByPage } from "./pagination";
import type { ListUsersParams, User } from "./types";

export function createUser(payload: { username: string }) {
  return apiRequest<User>("/api/users", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(payload),
  });
}

export function updateUser(
  id: number,
  payload: Partial<{
    username: string;
    status: string;
    expire_at: string;
    traffic_limit: number;
    traffic_reset_day: number;
  }>,
) {
  return apiRequest<User>(`/api/users/${id}`, {
    method: "PUT",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(payload),
  });
}

export function disableUser(id: number) {
  return apiRequest<User>(`/api/users/${id}`, {
    method: "DELETE",
  });
}

export function deleteUser(id: number) {
  return apiRequest<{ message: string }>(`/api/users/${id}?hard=true`, {
    method: "DELETE",
  });
}

export function listUsers(params: ListUsersParams = {}) {
  const query = new URLSearchParams();
  if (typeof params.limit === "number") {
    query.set("limit", String(params.limit));
  }
  if (typeof params.offset === "number") {
    query.set("offset", String(params.offset));
  }
  if (params.status) {
    query.set("status", params.status);
  }

  const suffix = query.toString() ? `?${query.toString()}` : "";
  return apiRequest<User[]>(`/api/users${suffix}`);
}

export function listAllUsers(params: Omit<ListUsersParams, "limit" | "offset"> = {}) {
  return listAllByPage<User>((page) =>
    listUsers({
      ...params,
      ...page,
    }),
  );
}
