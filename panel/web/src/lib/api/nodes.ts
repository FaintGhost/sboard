import { apiRequest } from "./client"
import type { ListNodesParams, Node, NodeTrafficSample } from "./types"

export function listNodes(params: ListNodesParams = {}) {
  const query = new URLSearchParams()
  if (typeof params.limit === "number") query.set("limit", String(params.limit))
  if (typeof params.offset === "number") query.set("offset", String(params.offset))
  const suffix = query.toString() ? `?${query.toString()}` : ""
  return apiRequest<Node[]>(`/api/nodes${suffix}`)
}

export function createNode(payload: {
  name: string
  api_address: string
  api_port: number
  secret_key: string
  public_address: string
  group_id: number | null
}) {
  return apiRequest<Node>("/api/nodes", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  })
}

export function updateNode(
  id: number,
  payload: Partial<{
    name: string
    api_address: string
    api_port: number
    secret_key: string
    public_address: string
    group_id: number | null
  }>,
) {
  return apiRequest<Node>(`/api/nodes/${id}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  })
}

export function deleteNode(id: number) {
  return apiRequest<{ status: string }>(`/api/nodes/${id}`, { method: "DELETE" })
}

export function nodeHealth(id: number) {
  return apiRequest<{ status: string }>(`/api/nodes/${id}/health`)
}

export function nodeSync(id: number) {
  return apiRequest<{ status: string }>(`/api/nodes/${id}/sync`, { method: "POST" })
}

export function listNodeTraffic(
  id: number,
  params: { limit?: number; offset?: number } = {},
) {
  const query = new URLSearchParams()
  if (typeof params.limit === "number") query.set("limit", String(params.limit))
  if (typeof params.offset === "number") query.set("offset", String(params.offset))
  const suffix = query.toString() ? `?${query.toString()}` : ""
  // apiRequest() already unwraps the SuccessEnvelope { data: T }.
  // The backend returns { data: NodeTrafficSample[] }, so here we request T = NodeTrafficSample[].
  return apiRequest<NodeTrafficSample[]>(`/api/nodes/${id}/traffic${suffix}`)
}
