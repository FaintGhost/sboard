import "./client";
import { getUserGroups as _getUserGroups, replaceUserGroups as _replaceUserGroups } from "./gen";
import type { UserGroups } from "./types";

export function getUserGroups(userId: number): Promise<UserGroups> {
  return _getUserGroups({ path: { id: userId } }).then((r) => r.data!.data);
}

export function putUserGroups(
  userId: number,
  payload: { group_ids: number[] },
): Promise<UserGroups> {
  return _replaceUserGroups({ path: { id: userId }, body: payload }).then((r) => r.data!.data);
}
