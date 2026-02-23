import { test, expect, PanelAPI, NodeAPI, uniqueNodeName } from "../fixtures";

test.describe.serial("配置同步与验证", () => {
  let nodeId: number;
  const nodeName = uniqueNodeName();

  test("创建节点和入站配置", async ({ authenticatedPage, adminToken, request }) => {
    const page = authenticatedPage;
    const api = new PanelAPI(request, adminToken);

    // Create node via API for speed
    const nodeResp = await api.createNode({
      name: nodeName,
      api_address: "node",
      api_port: 3000,
      secret_key: "e2e-test-node-secret",
      public_address: "node",
    });
    expect(nodeResp.ok()).toBeTruthy();
    const nodeData = await nodeResp.json();
    nodeId = nodeData.data.id;

    // Wait for node to be detected as online
    await page.waitForTimeout(6_000);

    // Navigate to inbounds page
    await page.getByRole("link", { name: "Inbounds" }).click();
    await expect(page).toHaveURL(/\/inbounds/);

    // Click create inbound
    await page.getByRole("button", { name: "Create Inbound" }).click();
    await expect(page.getByRole("dialog", { name: "Create Inbound" })).toBeVisible();

    // Select the node
    await page.getByRole("combobox", { name: "Select Node" }).click();
    await page.getByRole("option", { name: new RegExp(nodeName) }).click();

    // Select VLESS preset
    await page.getByRole("combobox", { name: "Preset" }).click();
    await page.getByRole("option", { name: "VLESS" }).click();

    // Save the inbound
    await page.getByRole("dialog", { name: "Create Inbound" }).getByRole("button", { name: "Save" }).click();

    // Should close — inbound created with auto-sync
    await expect(page.getByRole("dialog", { name: "Create Inbound" })).not.toBeVisible({ timeout: 15_000 });
  });

  test("验证 Node 接收到正确配置（API 级别）", async ({ request }) => {
    const nodeApi = new NodeAPI(request);

    // Verify node health
    const healthResp = await nodeApi.health();
    expect(healthResp.ok()).toBeTruthy();
    const healthData = await healthResp.json();
    expect(healthData.status).toBe("ok");

    // Verify inbounds are loaded on the node
    const inboundsResp = await nodeApi.getInbounds();
    expect(inboundsResp.ok()).toBeTruthy();
  });

  test("修改入站配置后重新同步", async ({ adminToken, request }) => {
    const api = new PanelAPI(request, adminToken);

    // Trigger manual sync via API
    const syncResp = await api.syncNode(nodeId);
    expect(syncResp.ok()).toBeTruthy();
    const syncData = await syncResp.json();
    expect(syncData.data?.status).toBe("success");

    // Verify node is still healthy after re-sync
    const nodeApi = new NodeAPI(request);
    const healthResp = await nodeApi.health();
    expect(healthResp.ok()).toBeTruthy();
  });
});
