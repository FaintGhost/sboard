import "./client";
import {
  listGroups as _listGroups,
  createGroup as _createGroup,
  updateGroup as _updateGroup,
  deleteGroup as _deleteGroup,
} from "./gen";
import type { Group } from "./gen";
import type { ListGroupsParams } from "./types";

export function listGroups(params: ListGroupsParams = {}): Promise<Group[]> {
  return _listGroups({ query: params }).then((r) => r.data!.data);
}

export function createGroup(payload: { name: string; description: string }): Promise<Group> {
  return _createGroup({ body: payload }).then((r) => r.data!.data);
}

export function updateGroup(
  id: number,
  payload: Partial<{ name: string; description: string }>,
): Promise<Group> {
  return _updateGroup({ path: { id }, body: payload }).then((r) => r.data!.data);
}

export function deleteGroup(id: number): Promise<{ status: string }> {
  return _deleteGroup({ path: { id } }).then((r) => r.data!);
}
