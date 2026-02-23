import { test as base, type Page, type APIRequestContext } from "@playwright/test";

const BASE_URL = process.env.BASE_URL || "http://localhost:8080";
const SETUP_TOKEN = process.env.SETUP_TOKEN || "e2e-test-setup-token";

const ADMIN_USERNAME = "admin";
const ADMIN_PASSWORD = "admin12345678";

type AuthFixtures = {
  authenticatedPage: Page;
  adminToken: string;
};

async function ensureBootstrapAndLogin(request: APIRequestContext): Promise<string> {
  // Check bootstrap status
  const statusResp = await request.get(`${BASE_URL}/api/admin/bootstrap`);
  const statusData = await statusResp.json();

  if (statusData.data?.needs_setup) {
    // Perform bootstrap
    const bootstrapResp = await request.post(`${BASE_URL}/api/admin/bootstrap`, {
      headers: { "X-Setup-Token": SETUP_TOKEN },
      data: {
        username: ADMIN_USERNAME,
        password: ADMIN_PASSWORD,
        confirm_password: ADMIN_PASSWORD,
      },
    });
    if (!bootstrapResp.ok()) {
      throw new Error(`Bootstrap failed: ${bootstrapResp.status()} ${await bootstrapResp.text()}`);
    }
  }

  // Login
  const loginResp = await request.post(`${BASE_URL}/api/admin/login`, {
    data: { username: ADMIN_USERNAME, password: ADMIN_PASSWORD },
  });
  if (!loginResp.ok()) {
    throw new Error(`Login failed: ${loginResp.status()} ${await loginResp.text()}`);
  }
  const loginData = await loginResp.json();
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
