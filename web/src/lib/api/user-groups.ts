import { toApiError } from "@/lib/api/client";
import { rpcCall } from "@/lib/rpc/client";
import {
  getUserGroups as getUserGroupsRPC,
  replaceUserGroups as replaceUserGroupsRPC,
} from "@/lib/rpc/gen/sboard/panel/v1/panel-UserService_connectquery";
import type { UserGroups } from "./types";

export function getUserGroups(userId: number): Promise<UserGroups> {
  return rpcCall(getUserGroupsRPC, { id: BigInt(userId) })
    .then((r) => ({ group_ids: (r.groupIds ?? []).map((id) => Number(id)) }))
    .catch((e) => {
      throw toApiError(e);
    });
}

export function putUserGroups(
  userId: number,
  payload: { group_ids: number[] },
): Promise<UserGroups> {
  return rpcCall(replaceUserGroupsRPC, {
    id: BigInt(userId),
    groupIds: payload.group_ids.map((id) => BigInt(id)),
  })
    .then((r) => ({ group_ids: (r.groupIds ?? []).map((id) => Number(id)) }))
    .catch((e) => {
      throw toApiError(e);
    });
}
