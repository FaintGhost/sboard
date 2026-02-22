import "./client";
import {
  listNodes as _listNodes,
  createNode as _createNode,
  updateNode as _updateNode,
  deleteNode as _deleteNode,
  getNodeHealth as _getNodeHealth,
  syncNode as _syncNode,
  listNodeTraffic as _listNodeTraffic,
} from "./gen";
import type { Node, NodeTrafficSample } from "./gen";
import { listAllByPage } from "./pagination";
import type { ListNodesParams } from "./types";

export function listNodes(params: ListNodesParams = {}): Promise<Node[]> {
  return _listNodes({ query: params }).then((r) => r.data!.data);
}

export function listAllNodes(): Promise<Node[]> {
  return listAllByPage<Node>((page) => listNodes(page));
}

export function createNode(payload: {
  name: string;
  api_address: string;
  api_port: number;
  secret_key: string;
  public_address: string;
  group_id: number | null;
}): Promise<Node> {
  return _createNode({ body: payload }).then((r) => r.data!.data);
}

export function updateNode(
  id: number,
  payload: Partial<{
    name: string;
    api_address: string;
    api_port: number;
    secret_key: string;
    public_address: string;
    group_id: number | null;
  }>,
): Promise<Node> {
  return _updateNode({ path: { id }, body: payload }).then((r) => r.data!.data);
}

export function deleteNode(
  id: number,
  options: { force?: boolean } = {},
): Promise<{ status: string; force?: boolean; deleted_inbounds?: number }> {
  return _deleteNode({
    path: { id },
    query: options.force ? { force: "true" } : undefined,
  }).then((r) => r.data!);
}

export function nodeHealth(id: number): Promise<{ status: string }> {
  return _getNodeHealth({ path: { id } }).then((r) => r.data!);
}

export function nodeSync(id: number): Promise<{ status: string }> {
  return _syncNode({ path: { id } }).then((r) => r.data!);
}

export function listNodeTraffic(
  id: number,
  params: { limit?: number; offset?: number } = {},
): Promise<NodeTrafficSample[]> {
  return _listNodeTraffic({ path: { id }, query: params }).then((r) => r.data!.data);
}
