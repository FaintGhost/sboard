import { toApiError } from "@/lib/api/client";
import { rpcCall } from "@/lib/rpc/client";
import {
  createGroup as createGroupRPC,
  deleteGroup as deleteGroupRPC,
  listGroups as listGroupsRPC,
  updateGroup as updateGroupRPC,
} from "@/lib/rpc/gen/sboard/panel/v1/panel-GroupService_connectquery";
import { toGroup } from "@/lib/rpc/mappers";
import type { ListGroupsParams } from "./types";
import type { Group } from "./types";

export function listGroups(params: ListGroupsParams = {}): Promise<Group[]> {
  return rpcCall(listGroupsRPC, {
    limit: params.limit,
    offset: params.offset,
  })
    .then((r) => (r.data ?? []).map(toGroup))
    .catch((e) => {
      throw toApiError(e);
    });
}

export function createGroup(payload: { name: string; description: string }): Promise<Group> {
  return rpcCall(createGroupRPC, {
    name: payload.name,
    description: payload.description,
  })
    .then((r) => toGroup(r.data!))
    .catch((e) => {
      throw toApiError(e);
    });
}

export function updateGroup(
  id: number,
  payload: Partial<{ name: string; description: string }>,
): Promise<Group> {
  return rpcCall(updateGroupRPC, {
    id: BigInt(id),
    name: payload.name,
    description: payload.description,
  })
    .then((r) => toGroup(r.data!))
    .catch((e) => {
      throw toApiError(e);
    });
}

export function deleteGroup(id: number): Promise<{ status: string }> {
  return rpcCall(deleteGroupRPC, { id: BigInt(id) })
    .then((r) => ({ status: r.status }))
    .catch((e) => {
      throw toApiError(e);
    });
}
