import { test, expect } from "@playwright/test";

const BASE_URL = process.env.BASE_URL || "http://localhost:8080";

test.describe("心跳注册 API", () => {
  test("Heartbeat endpoint is public (no auth required)", async ({ request }) => {
    const resp = await request.post(
      `${BASE_URL}/rpc/sboard.panel.v1.NodeRegistrationService/Heartbeat`,
      {
        headers: { "Content-Type": "application/json" },
        data: {
          uuid: "smoke-test-heartbeat-uuid",
          secretKey: "smoke-test-key",
          apiAddr: "192.168.1.1:3000",
          version: "smoke-test",
        },
      },
    );
    expect(resp.status()).toBe(200);
    const body = await resp.json();
    // Unknown node → PENDING
    expect(body.status).toBe("NODE_HEARTBEAT_STATUS_PENDING");
  });

  test("duplicate heartbeat returns PENDING without error", async ({ request }) => {
    const uuid = "smoke-test-duplicate-uuid";

    // First heartbeat
    const resp1 = await request.post(
      `${BASE_URL}/rpc/sboard.panel.v1.NodeRegistrationService/Heartbeat`,
      {
        headers: { "Content-Type": "application/json" },
        data: { uuid, secretKey: "key-1", apiAddr: ":3000" },
      },
    );
    expect(resp1.status()).toBe(200);

    // Second heartbeat — same UUID
    const resp2 = await request.post(
      `${BASE_URL}/rpc/sboard.panel.v1.NodeRegistrationService/Heartbeat`,
      {
        headers: { "Content-Type": "application/json" },
        data: { uuid, secretKey: "key-1", apiAddr: ":3000" },
      },
    );
    expect(resp2.status()).toBe(200);
    const body = await resp2.json();
    // Already pending → should be RECOGNIZED (key matches the pending record)
    // or PENDING depending on implementation
    expect(["NODE_HEARTBEAT_STATUS_RECOGNIZED", "NODE_HEARTBEAT_STATUS_PENDING"]).toContain(
      body.status,
    );
  });
});
