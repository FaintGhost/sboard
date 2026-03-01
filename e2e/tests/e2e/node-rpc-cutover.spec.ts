import { test, expect, PanelAPI, NodeAPI, uniqueGroupName, uniqueInboundTag, uniqueNodeName, uniqueUsername } from "../fixtures";

const BASE_URL = process.env.BASE_URL || "http://localhost:8080";
const NODE_SECRET_KEY = process.env.NODE_SECRET_KEY || "e2e-test-node-secret";

test.describe.serial("Node RPC 直切边界", () => {
  test("同步成功 + Node RPC 健康 + 旧 REST 不可用 + 订阅 REST 保持可用", async ({ adminToken, request }) => {
    const panelApi = new PanelAPI(request, adminToken);
    const nodeApi = new NodeAPI(request);

    const groupResp = await panelApi.createGroup(uniqueGroupName(), "node rpc cutover group");
    expect(groupResp.ok()).toBeTruthy();
    const groupBody = await groupResp.json();
    const groupId = Number(groupBody.data.id);

    const userResp = await panelApi.createUser(uniqueUsername());
    expect(userResp.ok()).toBeTruthy();
    const userBody = await userResp.json();
    const userId = Number(userBody.data.id);
    const userUuid = String(userBody.data.uuid);

    const bindResp = await panelApi.replaceGroupUsers(groupId, [userId]);
    expect(bindResp.ok()).toBeTruthy();

    const nodeResp = await panelApi.createNode({
      name: uniqueNodeName(),
      api_address: "node",
      api_port: 3000,
      secret_key: NODE_SECRET_KEY,
      public_address: "node",
      group_id: groupId,
    });
    expect(nodeResp.ok()).toBeTruthy();
    const nodeBody = await nodeResp.json();
    const nodeId = Number(nodeBody.data.id);

    const inboundResp = await panelApi.createInbound({
      node_id: nodeId,
      tag: uniqueInboundTag(),
      protocol: "socks",
      listen_port: 16080,
      public_port: 16080,
      settings: { users: [] },
      tls_settings: {},
      transport_settings: {},
    });
    expect(inboundResp.ok()).toBeTruthy();

    const syncResp = await panelApi.syncNode(nodeId);
    expect(syncResp.ok()).toBeTruthy();
    const syncBody = await syncResp.json();
    expect(syncBody.status).toBe("ok");

    const healthResp = await nodeApi.health();
    expect(healthResp.ok()).toBeTruthy();
    const healthBody = await healthResp.json();
    expect(healthBody.status).toBe("ok");

    const legacy = await nodeApi.expectLegacyRESTUnavailable();
    expect(legacy.healthStatus).toBe(404);
    expect(legacy.trafficStatus).toBe(404);
    expect(legacy.inboundsStatus).toBe(404);
    expect(legacy.syncStatus).toBe(404);

    const subResp = await request.get(`${BASE_URL}/api/sub/${userUuid}?format=singbox`);
    expect(subResp.ok()).toBeTruthy();
    const subBody = await subResp.json();
    expect(Array.isArray(subBody.outbounds)).toBeTruthy();
    expect(subBody.outbounds.length).toBeGreaterThan(0);
  });
});
