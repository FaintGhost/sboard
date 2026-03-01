import { toApiError } from "@/lib/api/client";
import { i64, rpcCall } from "@/lib/rpc/client";
import {
  createNode as createNodeRPC,
  deleteNode as deleteNodeRPC,
  getNodeHealth as getNodeHealthRPC,
  listNodeTraffic as listNodeTrafficRPC,
  listNodes as listNodesRPC,
  syncNode as syncNodeRPC,
  updateNode as updateNodeRPC,
} from "@/lib/rpc/gen/sboard/panel/v1/panel-NodeService_connectquery";
import { toNode, toNodeTrafficSample } from "@/lib/rpc/mappers";
import { listAllByPage } from "./pagination";
import type { ListNodesParams } from "./types";
import type { Node, NodeTrafficSample } from "./types";

export function listNodes(params: ListNodesParams = {}): Promise<Node[]> {
  return rpcCall(listNodesRPC, {
    limit: params.limit,
    offset: params.offset,
  })
    .then((r) => (r.data ?? []).map(toNode))
    .catch((e) => {
      throw toApiError(e);
    });
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
  return rpcCall(createNodeRPC, {
    name: payload.name,
    apiAddress: payload.api_address,
    apiPort: payload.api_port,
    secretKey: payload.secret_key,
    publicAddress: payload.public_address,
    groupId: i64(payload.group_id),
  })
    .then((r) => toNode(r.data!))
    .catch((e) => {
      throw toApiError(e);
    });
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
  return rpcCall(updateNodeRPC, {
    id: BigInt(id),
    name: payload.name,
    apiAddress: payload.api_address,
    apiPort: payload.api_port,
    secretKey: payload.secret_key,
    publicAddress: payload.public_address,
    groupId: i64(payload.group_id),
    clearGroupId: payload.group_id === null,
  })
    .then((r) => toNode(r.data!))
    .catch((e) => {
      throw toApiError(e);
    });
}

export function deleteNode(
  id: number,
  options: { force?: boolean } = {},
): Promise<{ status: string; force?: boolean; deleted_inbounds?: number }> {
  return rpcCall(deleteNodeRPC, { id: BigInt(id), force: !!options.force })
    .then((r) => ({
      status: r.status,
      force: r.force,
      deleted_inbounds: r.deletedInbounds,
    }))
    .catch((e) => {
      throw toApiError(e);
    });
}

export function nodeHealth(id: number): Promise<{ status: string }> {
  return rpcCall(getNodeHealthRPC, { id: BigInt(id) })
    .then((r) => ({ status: r.status }))
    .catch((e) => {
      throw toApiError(e);
    });
}

export function nodeSync(id: number): Promise<{ status: string }> {
  return rpcCall(syncNodeRPC, { id: BigInt(id) })
    .then((r) => ({ status: r.status }))
    .catch((e) => {
      throw toApiError(e);
    });
}

export function listNodeTraffic(
  id: number,
  params: { limit?: number; offset?: number } = {},
): Promise<NodeTrafficSample[]> {
  return rpcCall(listNodeTrafficRPC, {
    id: BigInt(id),
    limit: params.limit,
    offset: params.offset,
  })
    .then((r) => (r.data ?? []).map(toNodeTrafficSample))
    .catch((e) => {
      throw toApiError(e);
    });
}
