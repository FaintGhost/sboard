import { test, expect } from "@playwright/test";

const BASE_URL = process.env.BASE_URL || "http://localhost:8080";
const NODE_API_URL = process.env.NODE_API_URL || "http://node:3000";

test.describe("系统健康检查", () => {
  test("Panel 健康检查", async ({ request }) => {
    const resp = await request.post(`${BASE_URL}/rpc/sboard.panel.v1.HealthService/GetHealth`, {
      data: {},
    });
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(body.status).toBe("ok");
  });

  test("Node 健康检查", async ({ request }) => {
    const resp = await request.post(`${NODE_API_URL}/rpc/sboard.node.v1.NodeControlService/Health`, {
      data: {},
    });
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(body.status).toBe("ok");
  });
});
