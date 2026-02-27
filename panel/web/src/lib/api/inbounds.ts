import { toApiError } from "@/lib/api/client";
import { i64, rpcCall } from "@/lib/rpc/client";
import {
  createInbound as createInboundRPC,
  deleteInbound as deleteInboundRPC,
  listInbounds as listInboundsRPC,
  updateInbound as updateInboundRPC,
} from "@/lib/rpc/gen/sboard/panel/v1/panel-InboundService_connectquery";
import type { ListInboundsParams } from "./types";
import type { Inbound } from "./types";

function mapInbound(v: {
  id: bigint;
  uuid: string;
  nodeId: bigint;
  tag: string;
  protocol: string;
  listenPort: number;
  publicPort: number;
  settingsJson: string;
  tlsSettingsJson: string;
  transportSettingsJson: string;
}): Inbound {
  return {
    id: Number(v.id),
    uuid: v.uuid,
    node_id: Number(v.nodeId),
    tag: v.tag,
    protocol: v.protocol,
    listen_port: v.listenPort,
    public_port: v.publicPort,
    settings: v.settingsJson ? JSON.parse(v.settingsJson) : {},
    tls_settings: v.tlsSettingsJson ? JSON.parse(v.tlsSettingsJson) : {},
    transport_settings: v.transportSettingsJson ? JSON.parse(v.transportSettingsJson) : {},
  };
}

export function listInbounds(params: ListInboundsParams = {}): Promise<Inbound[]> {
  return rpcCall(listInboundsRPC, {
    limit: params.limit,
    offset: params.offset,
    nodeId: i64(params.node_id),
  })
    .then((r) => (r.data ?? []).map(mapInbound))
    .catch((e) => {
      throw toApiError(e);
    });
}

export function createInbound(payload: {
  node_id: number;
  tag: string;
  protocol: string;
  listen_port: number;
  public_port: number;
  settings: unknown;
  tls_settings?: unknown;
  transport_settings?: unknown;
}): Promise<Inbound> {
  return rpcCall(createInboundRPC, {
    nodeId: BigInt(payload.node_id),
    tag: payload.tag,
    protocol: payload.protocol,
    listenPort: payload.listen_port,
    publicPort: payload.public_port,
    settingsJson: JSON.stringify(payload.settings ?? {}),
    tlsSettingsJson: JSON.stringify(payload.tls_settings ?? {}),
    transportSettingsJson: JSON.stringify(payload.transport_settings ?? {}),
  })
    .then((r) => mapInbound(r.data!))
    .catch((e) => {
      throw toApiError(e);
    });
}

export function updateInbound(
  id: number,
  payload: Partial<{
    node_id: number;
    tag: string;
    protocol: string;
    listen_port: number;
    public_port: number;
    settings: unknown;
    tls_settings: unknown;
    transport_settings: unknown;
  }>,
): Promise<Inbound> {
  return rpcCall(updateInboundRPC, {
    id: BigInt(id),
    tag: payload.tag,
    protocol: payload.protocol,
    listenPort: payload.listen_port,
    publicPort: payload.public_port,
    settingsJson: payload.settings === undefined ? undefined : JSON.stringify(payload.settings),
    tlsSettingsJson:
      payload.tls_settings === undefined ? undefined : JSON.stringify(payload.tls_settings),
    transportSettingsJson:
      payload.transport_settings === undefined
        ? undefined
        : JSON.stringify(payload.transport_settings),
  })
    .then((r) => mapInbound(r.data!))
    .catch((e) => {
      throw toApiError(e);
    });
}

export function deleteInbound(id: number): Promise<{ status: string }> {
  return rpcCall(deleteInboundRPC, { id: BigInt(id) })
    .then((r) => ({ status: r.status }))
    .catch((e) => {
      throw toApiError(e);
    });
}
