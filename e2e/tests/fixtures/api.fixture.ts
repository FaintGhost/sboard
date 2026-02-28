import type { APIRequestContext } from "@playwright/test";

const BASE_URL = process.env.BASE_URL || "http://localhost:8080";
const NODE_API_URL = process.env.NODE_API_URL || "http://node:3000";
const NODE_SECRET_KEY = process.env.NODE_SECRET_KEY || "e2e-test-node-secret";

function i64(v: number): string {
  return String(Math.trunc(v));
}

export class PanelAPI {
  request: APIRequestContext;
  token: string;

  constructor(request: APIRequestContext, token: string) {
    this.request = request;
    this.token = token;
  }

  get headers() {
    return { Authorization: `Bearer ${this.token}` };
  }

  get rpcHeaders() {
    return {
      ...this.headers,
      "Content-Type": "application/json",
    };
  }

  async health() {
    return this.request.post(`${BASE_URL}/rpc/sboard.panel.v1.HealthService/GetHealth`, {
      headers: { "Content-Type": "application/json" },
      data: {},
    });
  }

  async createUser(username: string) {
    return this.request.post(`${BASE_URL}/rpc/sboard.panel.v1.UserService/CreateUser`, {
      headers: this.rpcHeaders,
      data: { username },
    });
  }

  async updateUser(id: number, data: Record<string, unknown>) {
    return this.request.post(`${BASE_URL}/rpc/sboard.panel.v1.UserService/UpdateUser`, {
      headers: this.rpcHeaders,
      data: {
        id: i64(id),
        username: data.username,
        status: data.status,
        expireAt: data.expire_at,
        trafficLimit:
          typeof data.traffic_limit === "number" ? i64(data.traffic_limit) : data.traffic_limit,
        trafficResetDay: data.traffic_reset_day,
      },
    });
  }

  async deleteUser(id: number) {
    return this.request.post(`${BASE_URL}/rpc/sboard.panel.v1.UserService/DeleteUser`, {
      headers: this.rpcHeaders,
      data: { id: i64(id) },
    });
  }

  async listUsers() {
    return this.request.post(`${BASE_URL}/rpc/sboard.panel.v1.UserService/ListUsers`, {
      headers: this.rpcHeaders,
      data: {},
    });
  }

  async createGroup(name: string, description?: string) {
    return this.request.post(`${BASE_URL}/rpc/sboard.panel.v1.GroupService/CreateGroup`, {
      headers: this.rpcHeaders,
      data: { name, description: description || "" },
    });
  }

  async updateGroup(id: number, data: Record<string, unknown>) {
    return this.request.post(`${BASE_URL}/rpc/sboard.panel.v1.GroupService/UpdateGroup`, {
      headers: this.rpcHeaders,
      data: {
        id: i64(id),
        name: data.name,
        description: data.description,
      },
    });
  }

  async deleteGroup(id: number) {
    return this.request.post(`${BASE_URL}/rpc/sboard.panel.v1.GroupService/DeleteGroup`, {
      headers: this.rpcHeaders,
      data: { id: i64(id) },
    });
  }

  async listGroups() {
    return this.request.post(`${BASE_URL}/rpc/sboard.panel.v1.GroupService/ListGroups`, {
      headers: this.rpcHeaders,
      data: {},
    });
  }

  async replaceGroupUsers(groupId: number, userIds: number[]) {
    return this.request.post(`${BASE_URL}/rpc/sboard.panel.v1.GroupService/ReplaceGroupUsers`, {
      headers: this.rpcHeaders,
      data: {
        id: i64(groupId),
        userIds: userIds.map((id) => i64(id)),
      },
    });
  }

  async createNode(data: {
    name: string;
    api_address: string;
    api_port: number;
    secret_key: string;
    public_address: string;
    group_id?: number;
  }) {
    return this.request.post(`${BASE_URL}/rpc/sboard.panel.v1.NodeService/CreateNode`, {
      headers: this.rpcHeaders,
      data: {
        name: data.name,
        apiAddress: data.api_address,
        apiPort: data.api_port,
        secretKey: data.secret_key,
        publicAddress: data.public_address,
        groupId: data.group_id === undefined ? undefined : i64(data.group_id),
      },
    });
  }

  async updateNode(id: number, data: Record<string, unknown>) {
    return this.request.post(`${BASE_URL}/rpc/sboard.panel.v1.NodeService/UpdateNode`, {
      headers: this.rpcHeaders,
      data: {
        id: i64(id),
        name: data.name,
        apiAddress: data.api_address,
        apiPort: data.api_port,
        secretKey: data.secret_key,
        publicAddress: data.public_address,
        groupId: typeof data.group_id === "number" ? i64(data.group_id) : undefined,
        clearGroupId: data.group_id === null,
      },
    });
  }

  async deleteNode(id: number, force?: boolean) {
    return this.request.post(`${BASE_URL}/rpc/sboard.panel.v1.NodeService/DeleteNode`, {
      headers: this.rpcHeaders,
      data: { id: i64(id), force: !!force },
    });
  }

  async listNodes() {
    return this.request.post(`${BASE_URL}/rpc/sboard.panel.v1.NodeService/ListNodes`, {
      headers: this.rpcHeaders,
      data: {},
    });
  }

  async getNodeHealth(id: number) {
    return this.request.post(`${BASE_URL}/rpc/sboard.panel.v1.NodeService/GetNodeHealth`, {
      headers: this.rpcHeaders,
      data: { id: i64(id) },
    });
  }

  async syncNode(id: number) {
    return this.request.post(`${BASE_URL}/rpc/sboard.panel.v1.NodeService/SyncNode`, {
      headers: this.rpcHeaders,
      data: { id: i64(id) },
    });
  }

  async createInbound(data: Record<string, unknown>) {
    return this.request.post(`${BASE_URL}/rpc/sboard.panel.v1.InboundService/CreateInbound`, {
      headers: this.rpcHeaders,
      data: {
        nodeId: typeof data.node_id === "number" ? i64(data.node_id) : data.node_id,
        tag: data.tag,
        protocol: data.protocol,
        listenPort: data.listen_port,
        publicPort: data.public_port,
        settingsJson: JSON.stringify(data.settings ?? {}),
        tlsSettingsJson: JSON.stringify(data.tls_settings ?? {}),
        transportSettingsJson: JSON.stringify(data.transport_settings ?? {}),
      },
    });
  }

  async updateInbound(id: number, data: Record<string, unknown>) {
    return this.request.post(`${BASE_URL}/rpc/sboard.panel.v1.InboundService/UpdateInbound`, {
      headers: this.rpcHeaders,
      data: {
        id: i64(id),
        tag: data.tag,
        protocol: data.protocol,
        listenPort: data.listen_port,
        publicPort: data.public_port,
        settingsJson: data.settings === undefined ? undefined : JSON.stringify(data.settings),
        tlsSettingsJson:
          data.tls_settings === undefined ? undefined : JSON.stringify(data.tls_settings),
        transportSettingsJson:
          data.transport_settings === undefined
            ? undefined
            : JSON.stringify(data.transport_settings),
      },
    });
  }

  async deleteInbound(id: number) {
    return this.request.post(`${BASE_URL}/rpc/sboard.panel.v1.InboundService/DeleteInbound`, {
      headers: this.rpcHeaders,
      data: { id: i64(id) },
    });
  }

  async listInbounds() {
    return this.request.post(`${BASE_URL}/rpc/sboard.panel.v1.InboundService/ListInbounds`, {
      headers: this.rpcHeaders,
      data: {},
    });
  }

  async listSyncJobs() {
    return this.request.post(`${BASE_URL}/rpc/sboard.panel.v1.SyncJobService/ListSyncJobs`, {
      headers: this.rpcHeaders,
      data: {},
    });
  }

  async getSystemSettings() {
    return this.request.post(`${BASE_URL}/rpc/sboard.panel.v1.SystemService/GetSystemSettings`, {
      headers: this.rpcHeaders,
      data: {},
    });
  }

  async updateSystemSettings(data: Record<string, unknown>) {
    return this.request.post(`${BASE_URL}/rpc/sboard.panel.v1.SystemService/UpdateSystemSettings`, {
      headers: this.rpcHeaders,
      data: {
        subscriptionBaseUrl: data.subscription_base_url,
        timezone: data.timezone,
      },
    });
  }
}

export class NodeAPI {
  request: APIRequestContext;

  constructor(request: APIRequestContext) {
    this.request = request;
  }

  get headers() {
    return { Authorization: `Bearer ${NODE_SECRET_KEY}` };
  }

  async health() {
    return this.request.get(`${NODE_API_URL}/api/health`);
  }

  async getTraffic() {
    return this.request.get(`${NODE_API_URL}/api/stats/traffic`, {
      headers: this.headers,
    });
  }

  async getInbounds(options?: { reset?: boolean }) {
    const params = new URLSearchParams();
    if (options?.reset) {
      params.set("reset", "true");
    }

    const query = params.toString();
    const url = query
      ? `${NODE_API_URL}/api/stats/inbounds?${query}`
      : `${NODE_API_URL}/api/stats/inbounds`;

    return this.request.get(url, {
      headers: this.headers,
    });
  }
}

export { BASE_URL, NODE_API_URL, NODE_SECRET_KEY };
