import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { resetAuthStore } from "@/store/auth";

import { listAllNodes } from "./nodes";

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
    const calls: Array<{ limit: string | null; offset: string | null }> = [];

    vi.spyOn(globalThis, "fetch").mockImplementation(async (input) => {
      const req = input as Request;
      const url = new URL(req.url);
      if (req.method !== "GET" || url.pathname !== "/api/nodes") {
        return new Response(JSON.stringify({ error: "not found" }), {
          status: 404,
          headers: { "Content-Type": "application/json" },
        });
      }

      const limit = url.searchParams.get("limit");
      const offset = url.searchParams.get("offset");
      calls.push({ limit, offset });

      let data = [] as ReturnType<typeof buildNode>[];
      if (offset === "0") {
        data = Array.from({ length: 500 }, (_, i) => buildNode(i + 1));
      } else if (offset === "500") {
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
      { limit: "500", offset: "0" },
      { limit: "500", offset: "500" },
    ]);
  });
});
