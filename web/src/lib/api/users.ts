import { toApiError } from "@/lib/api/client";
import { i64, rpcCall } from "@/lib/rpc/client";
import {
  createUser as createUserRPC,
  deleteUser as deleteUserRPC,
  disableUser as disableUserRPC,
  listUsers as listUsersRPC,
  updateUser as updateUserRPC,
} from "@/lib/rpc/gen/sboard/panel/v1/panel-UserService_connectquery";
import { toUser } from "@/lib/rpc/mappers";
import { listAllByPage } from "./pagination";
import type { ListUsersParams } from "./types";
import type { User } from "./types";

export function listUsers(params: ListUsersParams = {}): Promise<User[]> {
  return rpcCall(listUsersRPC, {
    limit: params.limit,
    offset: params.offset,
    status: params.status,
  })
    .then((r) => (r.data ?? []).map(toUser))
    .catch((e) => {
      throw toApiError(e);
    });
}

export function createUser(payload: { username: string }): Promise<User> {
  return rpcCall(createUserRPC, payload)
    .then((r) => toUser(r.data!))
    .catch((e) => {
      throw toApiError(e);
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
): Promise<User> {
  return rpcCall(updateUserRPC, {
    id: BigInt(id),
    username: payload.username,
    status: payload.status,
    expireAt: payload.expire_at,
    trafficLimit: i64(payload.traffic_limit),
    trafficResetDay: payload.traffic_reset_day,
  })
    .then((r) => toUser(r.data!))
    .catch((e) => {
      throw toApiError(e);
    });
}

export function disableUser(id: number): Promise<User> {
  return rpcCall(disableUserRPC, { id: BigInt(id) })
    .then((r) => toUser(r.data!))
    .catch((e) => {
      throw toApiError(e);
    });
}

export function deleteUser(id: number): Promise<{ message: string }> {
  return rpcCall(deleteUserRPC, { id: BigInt(id) })
    .then((r) => ({ message: r.message ?? "" }))
    .catch((e) => {
      throw toApiError(e);
    });
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
