import { apiRequest } from "./client";
import type { User } from "./types";

export function listGroupUsers(groupId: number) {
  return apiRequest<User[]>(`/api/groups/${groupId}/users`);
}

export function replaceGroupUsers(groupId: number, payload: { user_ids: number[] }) {
  return apiRequest<{ user_ids: number[] }>(`/api/groups/${groupId}/users`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
}
