import { test, expect } from "@playwright/test";

const BASE_URL = process.env.BASE_URL || "http://localhost:8080";
const NODE_API_URL = process.env.NODE_API_URL || "http://node:3000";

test.describe("系统健康检查", () => {
  test("Panel 健康检查", async ({ request }) => {
    const resp = await request.get(`${BASE_URL}/api/health`);
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(body.status).toBe("ok");
  });

  test("Node 健康检查", async ({ request }) => {
    const resp = await request.get(`${NODE_API_URL}/api/health`);
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    expect(body.status).toBe("ok");
  });
});
