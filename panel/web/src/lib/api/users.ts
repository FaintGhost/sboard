import "./client";
import {
  listUsers as _listUsers,
  createUser as _createUser,
  updateUser as _updateUser,
  deleteUser as _deleteUser,
} from "./gen";
import type { User } from "./gen";
import { listAllByPage } from "./pagination";
import type { ListUsersParams } from "./types";

export function listUsers(params: ListUsersParams = {}): Promise<User[]> {
  return _listUsers({ query: params as Record<string, unknown> }).then((r) => r.data!.data);
}

export function createUser(payload: { username: string }): Promise<User> {
  return _createUser({ body: payload }).then((r) => r.data!.data);
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
): Promise<User> {
  return _updateUser({ path: { id }, body: payload }).then((r) => r.data!.data);
}

export function disableUser(id: number): Promise<User> {
  return _deleteUser({ path: { id } }).then((r) => r.data!.data!);
}

export function deleteUser(id: number): Promise<{ message: string }> {
  return _deleteUser({ path: { id }, query: { hard: "true" } }).then(
    (r) => r.data! as unknown as { message: string },
  );
}

export function listAllUsers(
  params: Omit<ListUsersParams, "limit" | "offset"> = {},
): Promise<User[]> {
  return listAllByPage<User>((page) =>
    listUsers({
      ...params,
      ...page,
    }),
  );
}
