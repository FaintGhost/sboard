import type { APIRequestContext } from "@playwright/test";

const BASE_URL = process.env.BASE_URL || "http://localhost:8080";
const NODE_API_URL = process.env.NODE_API_URL || "http://node:3000";
const NODE_SECRET_KEY = "e2e-test-node-secret";

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

  async health() {
    return this.request.get(`${BASE_URL}/api/health`);
  }

  async createUser(username: string) {
    return this.request.post(`${BASE_URL}/api/users`, {
      headers: this.headers,
      data: { username },
    });
  }

  async updateUser(id: number, data: Record<string, unknown>) {
    return this.request.put(`${BASE_URL}/api/users/${id}`, {
      headers: this.headers,
      data,
    });
  }

  async deleteUser(id: number) {
    return this.request.delete(`${BASE_URL}/api/users/${id}`, {
      headers: this.headers,
    });
  }

  async listUsers() {
    return this.request.get(`${BASE_URL}/api/users`, {
      headers: this.headers,
    });
  }

  async createGroup(name: string, description?: string) {
    return this.request.post(`${BASE_URL}/api/groups`, {
      headers: this.headers,
      data: { name, description: description || "" },
    });
  }

  async updateGroup(id: number, data: Record<string, unknown>) {
    return this.request.put(`${BASE_URL}/api/groups/${id}`, {
      headers: this.headers,
      data,
    });
  }

  async deleteGroup(id: number) {
    return this.request.delete(`${BASE_URL}/api/groups/${id}`, {
      headers: this.headers,
    });
  }

  async listGroups() {
    return this.request.get(`${BASE_URL}/api/groups`, {
      headers: this.headers,
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
    return this.request.post(`${BASE_URL}/api/nodes`, {
      headers: this.headers,
      data,
    });
  }

  async updateNode(id: number, data: Record<string, unknown>) {
    return this.request.put(`${BASE_URL}/api/nodes/${id}`, {
      headers: this.headers,
      data,
    });
  }

  async deleteNode(id: number, force?: boolean) {
    const url = force
      ? `${BASE_URL}/api/nodes/${id}?force=true`
      : `${BASE_URL}/api/nodes/${id}`;
    return this.request.delete(url, {
      headers: this.headers,
    });
  }

  async listNodes() {
    return this.request.get(`${BASE_URL}/api/nodes`, {
      headers: this.headers,
    });
  }

  async getNodeHealth(id: number) {
    return this.request.get(`${BASE_URL}/api/nodes/${id}/health`, {
      headers: this.headers,
    });
  }

  async syncNode(id: number) {
    return this.request.post(`${BASE_URL}/api/nodes/${id}/sync`, {
      headers: this.headers,
    });
  }

  async createInbound(data: Record<string, unknown>) {
    return this.request.post(`${BASE_URL}/api/inbounds`, {
      headers: this.headers,
      data,
    });
  }

  async updateInbound(id: number, data: Record<string, unknown>) {
    return this.request.put(`${BASE_URL}/api/inbounds/${id}`, {
      headers: this.headers,
      data,
    });
  }

  async deleteInbound(id: number) {
    return this.request.delete(`${BASE_URL}/api/inbounds/${id}`, {
      headers: this.headers,
    });
  }

  async listInbounds() {
    return this.request.get(`${BASE_URL}/api/inbounds`, {
      headers: this.headers,
    });
  }

  async listSyncJobs() {
    return this.request.get(`${BASE_URL}/api/sync-jobs`, {
      headers: this.headers,
    });
  }

  async getSystemSettings() {
    return this.request.get(`${BASE_URL}/api/system/settings`, {
      headers: this.headers,
    });
  }

  async updateSystemSettings(data: Record<string, unknown>) {
    return this.request.put(`${BASE_URL}/api/system/settings`, {
      headers: this.headers,
      data,
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

  async getInbounds() {
    return this.request.get(`${NODE_API_URL}/api/stats/inbounds`, {
      headers: this.headers,
    });
  }
}

export { BASE_URL, NODE_API_URL, NODE_SECRET_KEY };
