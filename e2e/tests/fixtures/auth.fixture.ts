import { test as base, type Page, type APIRequestContext } from "@playwright/test";

const BASE_URL = process.env.BASE_URL || "http://localhost:8080";
const SETUP_TOKEN = process.env.SETUP_TOKEN || "e2e-test-setup-token";

const ADMIN_USERNAME = "admin";
const ADMIN_PASSWORD = "admin12345678";

type RpcEnvelope<T> = {
  data?: T;
  error?: string;
};

async function rpcPost<T>(
  request: APIRequestContext,
  path: string,
  payload: unknown,
  headers?: Record<string, string>,
): Promise<RpcEnvelope<T>> {
  const resp = await request.post(`${BASE_URL}${path}`, {
    headers: {
      "Content-Type": "application/json",
      ...(headers ?? {}),
    },
    data: payload,
  });

  if (!resp.ok()) {
    throw new Error(`RPC ${path} failed: ${resp.status()} ${await resp.text()}`);
  }

  return (await resp.json()) as RpcEnvelope<T>;
}

type AuthFixtures = {
  authenticatedPage: Page;
  adminToken: string;
};

async function ensureBootstrapAndLogin(request: APIRequestContext): Promise<string> {
  const statusData = await rpcPost<{ needsSetup: boolean }>(
    request,
    "/rpc/sboard.panel.v1.AuthService/GetBootstrapStatus",
    {},
  );

  if (statusData.data?.needsSetup) {
    await rpcPost<{ ok: boolean }>(
      request,
      "/rpc/sboard.panel.v1.AuthService/Bootstrap",
      {
        setupToken: SETUP_TOKEN,
        xSetupToken: SETUP_TOKEN,
        username: ADMIN_USERNAME,
        password: ADMIN_PASSWORD,
        confirmPassword: ADMIN_PASSWORD,
      },
    );
  }

  const loginData = await rpcPost<{ token: string; expiresAt: string }>(
    request,
    "/rpc/sboard.panel.v1.AuthService/Login",
    { username: ADMIN_USERNAME, password: ADMIN_PASSWORD },
  );

  if (!loginData.data?.token) {
    throw new Error("RPC login returned empty token");
  }
  return loginData.data.token;
}

export const test = base.extend<AuthFixtures>({
  adminToken: async ({ request }, use) => {
    const token = await ensureBootstrapAndLogin(request);
    await use(token);
  },

  authenticatedPage: async ({ page, request }, use) => {
    const token = await ensureBootstrapAndLogin(request);

    // Navigate to login page first to set localStorage on the correct origin
    await page.goto(`${BASE_URL}/login`);
    await page.evaluate((t) => {
      localStorage.setItem("sboard_token", t);
    }, token);

    // Navigate to dashboard to confirm auth state
    await page.goto(`${BASE_URL}/`);
    await use(page);
  },
});

export { expect } from "@playwright/test";
export { ADMIN_USERNAME, ADMIN_PASSWORD, BASE_URL, SETUP_TOKEN };
