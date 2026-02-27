import { test, expect, uniqueNodeName } from "../fixtures";

test.describe.serial("节点管理", () => {
  let createdNodeName: string;

  test("创建节点", async ({ authenticatedPage }) => {
    const page = authenticatedPage;
    createdNodeName = uniqueNodeName();

    // Navigate to nodes page
    await page.getByRole("link", { name: "Nodes" }).click();
    await expect(page).toHaveURL(/\/nodes/);

    // Click create node button
    await page.getByRole("button", { name: "Create Node" }).click();

    // Fill form in dialog
    await expect(page.getByRole("dialog", { name: "Create Node" })).toBeVisible();
    await page.locator("#node-name").fill(createdNodeName);
    await page.locator("#node-api-addr").fill("node");
    await page.locator("#node-api-port").fill("3000");
    await page.locator("#node-secret").fill("e2e-test-node-secret");

    // Uncheck "Same as API address" to set custom public address
    const linkCheckbox = page.locator("#node-link-address");
    if (await linkCheckbox.isChecked()) {
      await linkCheckbox.uncheck();
    }
    await page.locator("#node-public").fill("node");

    // Save
    await page.getByRole("dialog", { name: "Create Node" }).getByRole("button", { name: "Save" }).click();

    // Dialog should close and node should appear
    await expect(page.getByRole("dialog", { name: "Create Node" })).not.toBeVisible({ timeout: 10_000 });
    await expect(page.getByRole("cell", { name: createdNodeName }).first()).toBeVisible({
      timeout: 5_000,
    });
  });

  test("查看节点健康状态", async ({ authenticatedPage }) => {
    const page = authenticatedPage;

    // Navigate to nodes page
    await page.getByRole("link", { name: "Nodes" }).click();
    await expect(page).toHaveURL(/\/nodes/);

    // Wait for node monitor to poll (interval is 5s in test env)
    // The node card should show Online status
    await expect(page.getByText("Online").first()).toBeVisible({ timeout: 30_000 });
  });

  test("删除节点", async ({ authenticatedPage }) => {
    const page = authenticatedPage;

    // Navigate to nodes page
    await page.getByRole("link", { name: "Nodes" }).click();
    await expect(page).toHaveURL(/\/nodes/);

    // Find the node in the table and delete it
    // Use the table view's action dropdown
    const nodeRow = page.getByRole("row").filter({ hasText: createdNodeName });
    await nodeRow.getByRole("button", { name: "Actions" }).click();
    await page.getByRole("menuitem", { name: "Delete Node" }).click();

    // Confirm deletion in dialog
    await expect(page.getByRole("dialog", { name: "Delete Node" })).toBeVisible();
    await page.getByRole("dialog", { name: "Delete Node" }).getByRole("button", { name: "Delete" }).click();

    // Dialog should close and node should disappear
    await expect(page.getByRole("dialog", { name: "Delete Node" })).not.toBeVisible({ timeout: 10_000 });
    await expect(page.getByText(createdNodeName)).not.toBeVisible({ timeout: 5_000 });
  });
});
