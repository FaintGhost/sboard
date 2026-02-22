import "./client";
import {
  listInbounds as _listInbounds,
  createInbound as _createInbound,
  updateInbound as _updateInbound,
  deleteInbound as _deleteInbound,
} from "./gen";
import type { Inbound } from "./gen";
import type { ListInboundsParams } from "./types";

export function listInbounds(params: ListInboundsParams = {}): Promise<Inbound[]> {
  return _listInbounds({ query: params }).then((r) => r.data!.data);
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
  return _createInbound({
    body: payload as Parameters<typeof _createInbound>[0]["body"],
  }).then((r) => r.data!.data);
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
  return _updateInbound({
    path: { id },
    body: payload as Parameters<typeof _updateInbound>[0]["body"],
  }).then((r) => r.data!.data);
}

export function deleteInbound(id: number): Promise<{ status: string }> {
  return _deleteInbound({ path: { id } }).then((r) => r.data!);
}
