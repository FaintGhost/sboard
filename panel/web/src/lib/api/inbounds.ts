import { apiRequest } from "./client";
import type { Inbound, ListInboundsParams } from "./types";

export function listInbounds(params: ListInboundsParams = {}) {
  const query = new URLSearchParams();
  if (typeof params.limit === "number") query.set("limit", String(params.limit));
  if (typeof params.offset === "number") query.set("offset", String(params.offset));
  if (typeof params.node_id === "number") query.set("node_id", String(params.node_id));
  const suffix = query.toString() ? `?${query.toString()}` : "";
  return apiRequest<Inbound[]>(`/api/inbounds${suffix}`);
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
}) {
  return apiRequest<Inbound>("/api/inbounds", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
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
) {
  return apiRequest<Inbound>(`/api/inbounds/${id}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
}

export function deleteInbound(id: number) {
  return apiRequest<{ status: string }>(`/api/inbounds/${id}`, { method: "DELETE" });
}
