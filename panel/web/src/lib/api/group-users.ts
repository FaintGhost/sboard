import { toApiError } from "@/lib/api/client";
import { rpcCall } from "@/lib/rpc/client";
import {
  listGroupUsers as listGroupUsersRPC,
  replaceGroupUsers as replaceGroupUsersRPC,
} from "@/lib/rpc/gen/sboard/panel/v1/panel-GroupService_connectquery";
import { toGroupUsersListItem } from "@/lib/rpc/mappers";

type GroupUsersListItem = ReturnType<typeof toGroupUsersListItem>;

export function listGroupUsers(groupId: number): Promise<GroupUsersListItem[]> {
  return rpcCall(listGroupUsersRPC, { id: BigInt(groupId) })
    .then((r) => (r.data ?? []).map(toGroupUsersListItem))
    .catch((e) => {
      throw toApiError(e);
    });
}

export function replaceGroupUsers(
  groupId: number,
  payload: { user_ids: number[] },
): Promise<{ user_ids: number[] }> {
  return rpcCall(replaceGroupUsersRPC, {
    id: BigInt(groupId),
    userIds: payload.user_ids.map((id) => BigInt(id)),
  })
    .then((r) => ({ user_ids: (r.userIds ?? []).map((id) => Number(id)) }))
    .catch((e) => {
      throw toApiError(e);
    });
}
