import { mkdir, writeFile } from "node:fs/promises";
import { join } from "node:path";

import { test, expect, PanelAPI, NodeAPI, uniqueGroupName, uniqueInboundTag, uniqueNodeName, uniqueUsername } from "../fixtures";

const BASE_URL = process.env.BASE_URL || "http://localhost:8080";
const PROBE_URL = process.env.PROBE_URL || "http://probe:8080";
const SB_CLIENT_SOCKS5 = process.env.SB_CLIENT_SOCKS5 || "sb-client:10801";
const SB_CLIENT_CONFIG_DIR = process.env.SB_CLIENT_CONFIG_DIR || "/shared";
const SB_CLIENT_CONFIG_PATH = process.env.SB_CLIENT_CONFIG_PATH || join(SB_CLIENT_CONFIG_DIR, "config.json");

type SubscriptionResponse = {
  outbounds?: Array<Record<string, unknown>>;
};

type InboundTrafficRow = {
  tag?: string;
  user?: string;
  uplink?: number;
  downlink?: number;
};

function sanitizeSocksOutbound(
  outbound: Record<string, unknown>,
  fallbackTag: string,
  fallbackUsername: string,
  fallbackPassword: string
) {
  const cleaned: Record<string, unknown> = {
    type: "socks",
    tag: String(outbound.tag || fallbackTag),
    server: String(outbound.server || ""),
    server_port: Number(outbound.server_port || 0),
  };
  cleaned.username = String(outbound.username || fallbackUsername);
  cleaned.password = String(outbound.password || fallbackPassword);
  return cleaned;
}

test.describe.serial("强 E2E 全生命周期（配置+流量）", () => {
  test("bootstrap -> 登录 -> 建资源 -> 下发 -> 订阅消费 -> 产流验证", async ({
    adminToken,
    request,
    browser,
  }) => {
    const panelApi = new PanelAPI(request, adminToken);
    const nodeApi = new NodeAPI(request);

    const groupName = uniqueGroupName();
    const username = uniqueUsername();
    const nodeName = uniqueNodeName();
    const inboundTag = uniqueInboundTag();
    const inboundPort = 16080;

    const groupResp = await panelApi.createGroup(groupName, "strong lifecycle e2e");
    expect(groupResp.ok()).toBeTruthy();
    const groupBody = await groupResp.json();
    const groupId = Number(groupBody.data.id);

    const userResp = await panelApi.createUser(username);
    expect(userResp.ok()).toBeTruthy();
    const userBody = await userResp.json();
    const userId = Number(userBody.data.id);
    const userUuid = String(userBody.data.uuid);

    const bindResp = await panelApi.replaceGroupUsers(groupId, [userId]);
    expect(bindResp.ok()).toBeTruthy();

    const nodeResp = await panelApi.createNode({
      name: nodeName,
      api_address: "node",
      api_port: 3000,
      secret_key: process.env.NODE_SECRET_KEY || "e2e-test-node-secret",
      public_address: "node",
      group_id: groupId,
    });
    expect(nodeResp.ok()).toBeTruthy();
    const nodeBody = await nodeResp.json();
    const nodeId = Number(nodeBody.data.id);

    await expect
      .poll(
        async () => {
          const resp = await panelApi.getNodeHealth(nodeId);
          if (!resp.ok()) return "error";
          const body = await resp.json();
          return String(body.status || "unknown");
        },
        { timeout: 60_000, intervals: [1000, 1500, 2000, 2500] }
      )
      .toBe("ok");

    const inboundResp = await panelApi.createInbound({
      node_id: nodeId,
      tag: inboundTag,
      protocol: "socks",
      listen_port: inboundPort,
      public_port: inboundPort,
      settings: { users: [] },
      tls_settings: {},
      transport_settings: {},
    });
    expect(inboundResp.ok()).toBeTruthy();

    const syncResp = await panelApi.syncNode(nodeId);
    expect(syncResp.ok()).toBeTruthy();
    const syncBody = await syncResp.json();
    expect(syncBody.status).toBe("ok");

    await expect
      .poll(
        async () => {
          const resp = await panelApi.listSyncJobs();
          if (!resp.ok()) return "error";
          const body = await resp.json();
          const jobs = Array.isArray(body.data) ? body.data : [];
          const target = jobs.find((job: Record<string, unknown>) => Number(job.nodeId) === nodeId);
          if (!target) return "missing";
          return String(target.status || "unknown");
        },
        { timeout: 30_000, intervals: [1000, 1500, 2000] }
      )
      .toBe("success");

    const subResp = await request.get(`${BASE_URL}/api/sub/${userUuid}?format=singbox`);
    expect(subResp.ok()).toBeTruthy();
    const subBody = (await subResp.json()) as SubscriptionResponse;
    const outbounds = Array.isArray(subBody.outbounds) ? subBody.outbounds : [];
    expect(outbounds.length).toBeGreaterThan(0);

    const targetOutbound = outbounds.find((item) => item.tag === inboundTag) || outbounds[0];
    expect(targetOutbound).toBeTruthy();
    expect(targetOutbound.type).toBe("socks");
    expect(targetOutbound.server).toBe("node");
    expect(Number(targetOutbound.server_port)).toBe(inboundPort);

    const proxyOutbound = sanitizeSocksOutbound(targetOutbound, inboundTag, username, userUuid);

    const config = {
      log: { level: "debug" },
      inbounds: [
        {
          type: "socks",
          tag: "local-socks",
          listen: "0.0.0.0",
          listen_port: 10801,
        },
      ],
      outbounds: [
        proxyOutbound,
        { type: "direct", tag: "direct" },
        { type: "block", tag: "block" },
      ],
      route: {
        rules: [{ inbound: ["local-socks"], outbound: String(proxyOutbound.tag) }],
        final: "direct",
      },
    };

    await mkdir(SB_CLIENT_CONFIG_DIR, { recursive: true });
    await writeFile(SB_CLIENT_CONFIG_PATH, JSON.stringify(config, null, 2), "utf8");

    await expect
      .poll(
        async () => {
          const ctx = await browser.newContext({
            proxy: { server: `socks5://${SB_CLIENT_SOCKS5}` },
          });
          const page = await ctx.newPage();
          try {
            const url = `${PROBE_URL}/?ts=${Date.now()}`;
            const response = await page.goto(url, {
              timeout: 10_000,
              waitUntil: "domcontentloaded",
            });
            if (!response || !response.ok()) return "request-failed";
            const bodyText = (await page.textContent("body")) || "";
            return bodyText.includes("ok") ? "ok" : "unexpected-body";
          } catch (error) {
            const message = error instanceof Error ? error.message : String(error);
            if (message.includes("ECONNREFUSED") || message.includes("ERR_PROXY_CONNECTION_FAILED")) {
              return "proxy-not-ready";
            }
            return "request-failed";
          } finally {
            await ctx.close();
          }
        },
        { timeout: 60_000, intervals: [1000, 1500, 2000, 2500] }
      )
      .toBe("ok");

    const resetResp = await nodeApi.getInbounds({ reset: true });
    expect(resetResp.ok()).toBeTruthy();

    const proxyCtx = await browser.newContext({
      proxy: { server: `socks5://${SB_CLIENT_SOCKS5}` },
    });
    const proxyPage = await proxyCtx.newPage();
    for (let i = 0; i < 3; i++) {
      const resp = await proxyPage.goto(`${PROBE_URL}/?burst=${i}&ts=${Date.now()}`, {
        timeout: 10_000,
        waitUntil: "domcontentloaded",
      });
      expect(resp?.ok()).toBeTruthy();
      await expect(proxyPage.locator("body")).toContainText("ok");
    }
    await proxyCtx.close();

    await expect
      .poll(
        async () => {
          const resp = await nodeApi.getInbounds();
          if (!resp.ok()) return false;
          const body = await resp.json();
          const rows = Array.isArray(body.data) ? (body.data as InboundTrafficRow[]) : [];
          return rows.some((row) => {
            if (row.tag !== inboundTag) return false;
            const userMatched = row.user === username || row.user === userUuid;
            const traffic = Number(row.uplink || 0) + Number(row.downlink || 0);
            return userMatched && traffic > 0;
          });
        },
        { timeout: 40_000, intervals: [1000, 1500, 2000] }
      )
      .toBeTruthy();
  });
});
