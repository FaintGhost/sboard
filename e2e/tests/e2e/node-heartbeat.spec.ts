import { test, expect, PanelAPI } from "../fixtures";

test.describe.serial("心跳注册与审批流程", () => {
  const heartbeatUUID = `e2e-heartbeat-${Date.now()}`;
  const heartbeatSecret = "e2e-heartbeat-secret";
  let pendingNodeId: number;

  test("Node 发送心跳创建 pending 记录", async ({ request, adminToken }) => {
    const api = new PanelAPI(request, adminToken);

    // Send heartbeat with an unknown UUID
    const resp = await api.heartbeat({
      uuid: heartbeatUUID,
      secretKey: heartbeatSecret,
      apiAddr: "10.0.0.1:3000",
    });
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(body.status).toBe("NODE_HEARTBEAT_STATUS_PENDING");
  });

  test("pending 节点出现在节点列表中", async ({ request, adminToken }) => {
    const api = new PanelAPI(request, adminToken);

    const resp = await api.listNodes();
    expect(resp.ok()).toBeTruthy();
    const body = await resp.json();
    const pending = body.data?.find(
      (n: { uuid: string; status: string }) => n.uuid === heartbeatUUID,
    );
    expect(pending).toBeDefined();
    expect(pending.status).toBe("pending");
    pendingNodeId = Number(pending.id);
  });

  test("审批 pending 节点成功", async ({ request, adminToken }) => {
    const api = new PanelAPI(request, adminToken);

    const resp = await api.approveNode(pendingNodeId, {
      name: "E2E-Heartbeat-Node",
    });
    expect(resp.ok()).toBeTruthy();
    const body = await resp.json();
    expect(body.data.status).toBe("offline");
    expect(body.data.name).toBe("E2E-Heartbeat-Node");
  });

  test("审批后心跳返回 RECOGNIZED", async ({ request, adminToken }) => {
    const api = new PanelAPI(request, adminToken);

    const resp = await api.heartbeat({
      uuid: heartbeatUUID,
      secretKey: heartbeatSecret,
    });
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(body.status).toBe("NODE_HEARTBEAT_STATUS_RECOGNIZED");
  });

  test("错误 secret key 返回 REJECTED", async ({ request, adminToken }) => {
    const api = new PanelAPI(request, adminToken);

    const resp = await api.heartbeat({
      uuid: heartbeatUUID,
      secretKey: "wrong-key",
    });
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(body.status).toBe("NODE_HEARTBEAT_STATUS_REJECTED");
  });

  test("清理: 删除测试节点", async ({ request, adminToken }) => {
    const api = new PanelAPI(request, adminToken);
    await api.deleteNode(pendingNodeId, true);
  });
});

test.describe.serial("拒绝 pending 节点", () => {
  const rejectUUID = `e2e-reject-${Date.now()}`;
  let rejectNodeId: number;

  test("创建 pending 节点", async ({ request, adminToken }) => {
    const api = new PanelAPI(request, adminToken);
    const resp = await api.heartbeat({
      uuid: rejectUUID,
      secretKey: "reject-key",
      apiAddr: "10.0.0.2:3000",
    });
    expect(resp.ok()).toBeTruthy();
  });

  test("获取 pending 节点 ID", async ({ request, adminToken }) => {
    const api = new PanelAPI(request, adminToken);
    const resp = await api.listNodes();
    const body = await resp.json();
    const pending = body.data?.find(
      (n: { uuid: string }) => n.uuid === rejectUUID,
    );
    expect(pending).toBeDefined();
    rejectNodeId = Number(pending.id);
  });

  test("拒绝 pending 节点成功", async ({ request, adminToken }) => {
    const api = new PanelAPI(request, adminToken);
    const resp = await api.rejectNode(rejectNodeId);
    expect(resp.ok()).toBeTruthy();
  });

  test("拒绝后节点从列表消失", async ({ request, adminToken }) => {
    const api = new PanelAPI(request, adminToken);
    const resp = await api.listNodes();
    const body = await resp.json();
    const found = body.data?.find(
      (n: { uuid: string }) => n.uuid === rejectUUID,
    );
    expect(found).toBeUndefined();
  });
});
