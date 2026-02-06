import { apiRequest } from "./client"
import type { UserGroups } from "./types"

export function getUserGroups(userId: number) {
  return apiRequest<UserGroups>(`/api/users/${userId}/groups`)
}

export function putUserGroups(userId: number, payload: { group_ids: number[] }) {
  return apiRequest<UserGroups>(`/api/users/${userId}/groups`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  })
}

