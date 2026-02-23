import { test, expect } from "@playwright/test";

const BASE_URL = process.env.BASE_URL || "http://localhost:8080";
const SETUP_TOKEN = process.env.SETUP_TOKEN || "e2e-test-setup-token";

const ADMIN_USERNAME = "admin";
const ADMIN_PASSWORD = "admin12345678";

test.describe.serial("Bootstrap 初始化", () => {
  test("首次访问重定向到登录页", async ({ page }) => {
    await page.goto(`${BASE_URL}/`);
    // Unauthenticated should redirect to login
    await expect(page).toHaveURL(/\/login/);
  });

  test("显示 Bootstrap 表单并成功完成初始化", async ({ page }) => {
    await page.goto(`${BASE_URL}/login`);

    // Wait for bootstrap status check to complete
    // Should show Bootstrap form since DB is fresh
    await expect(page.getByText("Bootstrap admin")).toBeVisible({ timeout: 15_000 });

    // Fill bootstrap form
    await page.locator("#setup_token").fill(SETUP_TOKEN);
    await page.locator("#username").fill(ADMIN_USERNAME);
    await page.locator("#password").fill(ADMIN_PASSWORD);
    await page.locator("#confirm_password").fill(ADMIN_PASSWORD);

    // Submit
    await page.getByRole("button", { name: "Create admin" }).click();

    // Should auto-login and redirect to dashboard
    await expect(page).toHaveURL(/^\/$|\/$/,  { timeout: 15_000 });
  });

  test("登录已初始化的系统", async ({ page }) => {
    await page.goto(`${BASE_URL}/login`);

    // Should show login form (not bootstrap) since admin already exists
    await expect(page.getByText("Admin Login")).toBeVisible({ timeout: 10_000 });

    // Fill login form
    await page.locator("#username").fill(ADMIN_USERNAME);
    await page.locator("#password").fill(ADMIN_PASSWORD);

    // Submit
    await page.getByRole("button", { name: "Login" }).click();

    // Should redirect to dashboard
    await expect(page).toHaveURL(/^\/$|\/$/,  { timeout: 10_000 });
    // Sidebar should be visible
    await expect(page.getByRole("link", { name: "Users" })).toBeVisible();
  });
});
