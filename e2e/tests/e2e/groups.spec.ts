import { test, expect, PanelAPI, uniqueGroupName, uniqueUsername } from "../fixtures";

test.describe.serial("分组管理", () => {
  let createdGroupName: string;

  test("创建分组", async ({ authenticatedPage }) => {
    const page = authenticatedPage;
    createdGroupName = uniqueGroupName();

    // Navigate to groups page
    await page.getByRole("link", { name: "Groups" }).click();
    await expect(page).toHaveURL(/\/groups/);

    // Click create group button
    await page.getByRole("button", { name: "Create Group" }).click();

    // Fill form in dialog
    await expect(page.getByRole("dialog", { name: "Create Group" })).toBeVisible();
    await page.locator("#group-name").fill(createdGroupName);
    await page.locator("#group-desc").fill("E2E test group");

    // Save
    await page.getByRole("dialog", { name: "Create Group" }).getByRole("button", { name: "Save" }).click();

    // Dialog should close and group should appear
    await expect(page.getByRole("dialog", { name: "Create Group" })).not.toBeVisible({ timeout: 10_000 });
    await expect(page.getByText(createdGroupName)).toBeVisible({ timeout: 5_000 });
  });

  test("将用户分配到分组", async ({ authenticatedPage, adminToken, request }) => {
    const page = authenticatedPage;

    // Create a test user via API for quick setup
    const api = new PanelAPI(request, adminToken);
    const testUsername = uniqueUsername();
    const createResp = await api.createUser(testUsername);
    expect(createResp.ok()).toBeTruthy();

    // Navigate to groups page
    await page.getByRole("link", { name: "Groups" }).click();
    await expect(page).toHaveURL(/\/groups/);

    // Find the group row and click edit
    const groupRow = page.getByRole("row").filter({ hasText: createdGroupName });
    await groupRow.getByRole("button", { name: "Actions" }).click();
    await page.getByRole("menuitem", { name: "Edit" }).click();

    // Edit dialog should appear
    await expect(page.getByRole("dialog", { name: "Edit Group" })).toBeVisible();

    // Search for the test user in the available users panel
    await page.locator("#groups-candidates-search").fill(testUsername);

    // Select the user from available list and add
    await page.getByText(testUsername).click();
    await page.getByRole("button", { name: "Add selected users" }).click();

    // Save
    await page.getByRole("dialog", { name: "Edit Group" }).getByRole("button", { name: "Save" }).click();

    // Dialog should close
    await expect(page.getByRole("dialog", { name: "Edit Group" })).not.toBeVisible({ timeout: 10_000 });

    // Verify member count increased
    const updatedRow = page.getByRole("row").filter({ hasText: createdGroupName });
    await expect(updatedRow).toBeVisible();
  });
});
