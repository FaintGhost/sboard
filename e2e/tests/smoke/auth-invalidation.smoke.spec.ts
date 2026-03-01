import { test, expect } from "@playwright/test";

const BASE_URL = process.env.BASE_URL || "http://localhost:8080";

test.describe("Token 失效验证", () => {
  test("expired/fake token on protected RPC returns 401", async ({ request }) => {
    const resp = await request.post(
      `${BASE_URL}/rpc/sboard.panel.v1.UserService/ListUsers`,
      {
        headers: {
          "Content-Type": "application/json",
          Authorization: "Bearer fake-expired-token-abc123",
        },
        data: {},
      },
    );
    expect(resp.status()).toBe(401);
  });

  test("no token on protected RPC returns 401", async ({ request }) => {
    const resp = await request.post(
      `${BASE_URL}/rpc/sboard.panel.v1.UserService/ListUsers`,
      {
        headers: {
          "Content-Type": "application/json",
        },
        data: {},
      },
    );
    expect(resp.status()).toBe(401);
  });

  test("public endpoints work without token", async ({ request }) => {
    const resp = await request.post(
      `${BASE_URL}/rpc/sboard.panel.v1.AuthService/GetBootstrapStatus`,
      {
        headers: {
          "Content-Type": "application/json",
        },
        data: {},
      },
    );
    expect(resp.status()).toBe(200);
  });
});
