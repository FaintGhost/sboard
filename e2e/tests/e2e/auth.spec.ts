import { test, expect, ADMIN_USERNAME, ADMIN_PASSWORD, BASE_URL, SETUP_TOKEN } from "../fixtures";

async function ensureBootstrap(request: import("@playwright/test").APIRequestContext) {
  const statusResp = await request.get(`${BASE_URL}/api/admin/bootstrap`);
  const statusData = await statusResp.json();
  if (statusData.data?.needs_setup) {
    await request.post(`${BASE_URL}/api/admin/bootstrap`, {
      headers: { "X-Setup-Token": SETUP_TOKEN },
      data: {
        username: ADMIN_USERNAME,
        password: ADMIN_PASSWORD,
        confirm_password: ADMIN_PASSWORD,
      },
    });
  }
}

test.describe("认证管理", () => {
  test("登录成功", async ({ page, request }) => {
    await ensureBootstrap(request);
    await page.goto(`${BASE_URL}/login`);
    await expect(page.getByText("Admin Login")).toBeVisible({ timeout: 10_000 });

    await page.locator("#username").fill(ADMIN_USERNAME);
    await page.locator("#password").fill(ADMIN_PASSWORD);
    await page.getByRole("button", { name: "Login" }).click();

    // Should redirect to dashboard
    await expect(page).toHaveURL(/^\/$|\/$/, { timeout: 10_000 });
    // Sidebar should be visible
    await expect(page.getByRole("link", { name: "Dashboard" })).toBeVisible();
  });

  test("登录失败 - 错误密码", async ({ page, request }) => {
    await ensureBootstrap(request);
    await page.goto(`${BASE_URL}/login`);
    await expect(page.getByText("Admin Login")).toBeVisible({ timeout: 10_000 });

    await page.locator("#username").fill(ADMIN_USERNAME);
    await page.locator("#password").fill("wrong-password");
    await page.getByRole("button", { name: "Login" }).click();

    // Should show error and stay on login page
    await expect(page.getByText(/failed|error|invalid/i)).toBeVisible({ timeout: 5_000 });
    await expect(page).toHaveURL(/\/login/);
  });

  test("未认证访问受保护页面", async ({ page }) => {
    // Clear any existing auth state
    await page.goto(`${BASE_URL}/login`);
    await page.evaluate(() => localStorage.clear());

    // Try to access protected page
    await page.goto(`${BASE_URL}/users`);

    // Should redirect to login
    await expect(page).toHaveURL(/\/login/, { timeout: 5_000 });
  });
});
