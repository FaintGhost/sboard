import { apiRequest } from "./client";
import type { Group, ListGroupsParams } from "./types";

export function listGroups(params: ListGroupsParams = {}) {
  const query = new URLSearchParams();
  if (typeof params.limit === "number") query.set("limit", String(params.limit));
  if (typeof params.offset === "number") query.set("offset", String(params.offset));
  const suffix = query.toString() ? `?${query.toString()}` : "";
  return apiRequest<Group[]>(`/api/groups${suffix}`);
}

export function createGroup(payload: { name: string; description: string }) {
  return apiRequest<Group>("/api/groups", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
}

export function updateGroup(id: number, payload: Partial<{ name: string; description: string }>) {
  return apiRequest<Group>(`/api/groups/${id}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
}

export function deleteGroup(id: number) {
  return apiRequest<{ status: string }>(`/api/groups/${id}`, { method: "DELETE" });
}
