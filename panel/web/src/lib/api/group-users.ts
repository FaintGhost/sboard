import "./client";
import { listGroupUsers as _listGroupUsers, replaceGroupUsers as _replaceGroupUsers } from "./gen";
import type { GroupUsersListItem } from "./gen";

export function listGroupUsers(groupId: number): Promise<GroupUsersListItem[]> {
  return _listGroupUsers({ path: { id: groupId } }).then((r) => r.data!.data);
}

export function replaceGroupUsers(
  groupId: number,
  payload: { user_ids: number[] },
): Promise<{ user_ids: number[] }> {
  return _replaceGroupUsers({ path: { id: groupId }, body: payload }).then((r) => r.data!.data);
}
