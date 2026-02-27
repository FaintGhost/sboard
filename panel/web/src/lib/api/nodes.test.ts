import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { resetAuthStore } from "@/store/auth";

import { listAllNodes } from "./nodes";

function asRequest(input: RequestInfo | URL, init?: RequestInit): Request {
  if (input instanceof Request) return input;
  return new Request(input, init);
}

function buildNode(id: number) {
  return {
    id,
    uuid: `n-${id}`,
    name: `node-${id}`,
    api_address: "127.0.0.1",
    api_port: 3000,
    secret_key: "secret",
    public_address: "example.com",
    group_id: null,
    status: "online",
    last_seen_at: null,
  };
}

describe("nodes api pagination", () => {
  beforeEach(() => {
    localStorage.clear();
    resetAuthStore();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("listAllNodes paginates with backend max limit", async () => {
    const calls: Array<{ limit: number | undefined; offset: number | undefined }> = [];

    vi.spyOn(globalThis, "fetch").mockImplementation(async (input, init) => {
      const req = asRequest(input, init);
      const url = new URL(req.url, "http://localhost");
      if (
        req.method !== "POST" ||
        !url.pathname.endsWith("/sboard.panel.v1.NodeService/ListNodes")
      ) {
        return new Response(JSON.stringify({ error: "not found" }), {
          status: 404,
          headers: { "Content-Type": "application/json" },
        });
      }

      const body = (await req.json()) as { limit?: number; offset?: number };
      const limit = body.limit;
      const offset = body.offset;
      calls.push({ limit, offset });

      let data = [] as ReturnType<typeof buildNode>[];
      if (offset === 0) {
        data = Array.from({ length: 500 }, (_, i) => buildNode(i + 1));
      } else if (offset === 500) {
        data = [buildNode(501)];
      }

      return new Response(JSON.stringify({ data }), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      });
    });

    const nodes = await listAllNodes();
    expect(nodes).toHaveLength(501);
    expect(calls).toEqual([
      { limit: 500, offset: 0 },
      { limit: 500, offset: 500 },
    ]);
  });
});
