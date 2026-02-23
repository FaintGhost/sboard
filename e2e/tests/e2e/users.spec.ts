import { test, expect, PanelAPI, uniqueUsername } from "../fixtures";

test.describe.serial("用户管理", () => {
  let createdUsername: string;

  test("创建新用户", async ({ authenticatedPage, adminToken, request }) => {
    const page = authenticatedPage;
    createdUsername = uniqueUsername();

    // Navigate to users page
    await page.getByRole("link", { name: "Users" }).click();
    await expect(page).toHaveURL(/\/users/);

    // Click create user button
    await page.getByRole("button", { name: "Create User" }).click();

    // Fill username in dialog
    await expect(page.getByRole("dialog", { name: "Create User" })).toBeVisible();
    await page.locator("#edit-username").fill(createdUsername);

    // Submit
    await page.getByRole("dialog", { name: "Create User" }).getByRole("button", { name: "Create" }).click();

    // Dialog should close and user should appear in the list
    await expect(page.getByRole("dialog", { name: "Create User" })).not.toBeVisible({ timeout: 10_000 });
    await expect(page.getByText(createdUsername)).toBeVisible({ timeout: 5_000 });
  });

  test("编辑用户信息", async ({ authenticatedPage }) => {
    const page = authenticatedPage;

    // Navigate to users page
    await page.getByRole("link", { name: "Users" }).click();
    await expect(page).toHaveURL(/\/users/);

    // Find the user row and click actions
    const userRow = page.getByRole("row").filter({ hasText: createdUsername });
    await userRow.getByRole("button", { name: "Actions" }).click();
    await page.getByRole("menuitem", { name: "Edit" }).click();

    // Edit dialog should appear
    await expect(page.getByRole("dialog", { name: "Edit User" })).toBeVisible();

    // Modify traffic limit
    await page.locator("#edit-traffic-limit").fill("100");

    // Save
    await page.getByRole("dialog", { name: "Edit User" }).getByRole("button", { name: "Save" }).click();

    // Dialog should close
    await expect(page.getByRole("dialog", { name: "Edit User" })).not.toBeVisible({ timeout: 10_000 });
  });

  test("删除用户", async ({ authenticatedPage }) => {
    const page = authenticatedPage;

    // Navigate to users page
    await page.getByRole("link", { name: "Users" }).click();
    await expect(page).toHaveURL(/\/users/);

    // Find the user row and click actions
    const userRow = page.getByRole("row").filter({ hasText: createdUsername });
    await userRow.getByRole("button", { name: "Actions" }).click();
    await page.getByRole("menuitem", { name: "Delete" }).click();

    // Confirm deletion in dialog
    await expect(page.getByRole("dialog", { name: "Delete User" })).toBeVisible();
    await page.getByRole("dialog", { name: "Delete User" }).getByRole("button", { name: "Confirm" }).click();

    // Dialog should close and user should disappear
    await expect(page.getByRole("dialog", { name: "Delete User" })).not.toBeVisible({ timeout: 10_000 });
    await expect(page.getByText(createdUsername)).not.toBeVisible({ timeout: 5_000 });
  });
});
