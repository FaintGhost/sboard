import { test, expect, PanelAPI, NodeAPI, uniqueNodeName, uniqueUsername } from "../fixtures";

const BASE_URL = process.env.BASE_URL || "http://localhost:8080";

test.describe.serial("订阅管理", () => {
  let userId: number;
  let userUuid: string;
  let nodeId: number;
  let testUsername: string;

  test("前置数据准备 + 验证订阅链接可见", async ({ authenticatedPage, adminToken, request }) => {
    const page = authenticatedPage;
    const api = new PanelAPI(request, adminToken);

    // Create test user
    testUsername = uniqueUsername();
    const userResp = await api.createUser(testUsername);
    expect(userResp.ok()).toBeTruthy();
    const userData = await userResp.json();
    userId = userData.data.id;
    userUuid = userData.data.uuid;

    // Create node
    const nodeName = uniqueNodeName();
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

    // Wait for node to be online
    await page.waitForTimeout(6_000);

    // Navigate to subscriptions page
    await page.getByRole("link", { name: "Subscriptions" }).click();
    await expect(page).toHaveURL(/\/subscriptions/);

    // Search for the test user
    await page.locator("#subscriptions-search").fill(testUsername);

    // The subscription URL should be visible for this user
    await expect(page.getByText(userUuid).first()).toBeVisible({ timeout: 10_000 });
  });

  test("验证订阅内容", async ({ request }) => {
    // Access subscription URL directly via API
    const subResp = await request.get(`${BASE_URL}/api/sub/${userUuid}?format=singbox`);
    expect(subResp.ok()).toBeTruthy();

    const contentType = subResp.headers()["content-type"];
    const body = await subResp.text();

    // Should be valid JSON (sing-box format)
    let parsed: Record<string, unknown>;
    try {
      parsed = JSON.parse(body);
    } catch {
      throw new Error(`Subscription content is not valid JSON: ${body.slice(0, 200)}`);
    }

    // sing-box config should have outbounds
    expect(parsed).toHaveProperty("outbounds");
    expect(Array.isArray(parsed.outbounds)).toBeTruthy();
  });
});
