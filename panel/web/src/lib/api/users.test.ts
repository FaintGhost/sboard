import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { resetAuthStore } from "@/store/auth";

import { listAllUsers } from "./users";

function buildUser(id: number) {
  return {
    id,
    uuid: `u-${id}`,
    username: `user-${id}`,
    group_ids: [],
    traffic_limit: 0,
    traffic_used: 0,
    traffic_reset_day: 1,
    expire_at: null,
    status: "active" as const,
  };
}

describe("users api pagination", () => {
  beforeEach(() => {
    localStorage.clear();
    resetAuthStore();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("listAllUsers paginates with backend max limit", async () => {
    const calls: Array<{ limit: string | null; offset: string | null }> = [];

    vi.spyOn(globalThis, "fetch").mockImplementation(async (input) => {
      const req = input as Request;
      const url = new URL(req.url);
      if (req.method !== "GET" || url.pathname !== "/api/users") {
        return new Response(JSON.stringify({ error: "not found" }), {
          status: 404,
          headers: { "Content-Type": "application/json" },
        });
      }

      const limit = url.searchParams.get("limit");
      const offset = url.searchParams.get("offset");
      calls.push({ limit, offset });

      let data = [] as ReturnType<typeof buildUser>[];
      if (offset === "0") {
        data = Array.from({ length: 500 }, (_, i) => buildUser(i + 1));
      } else if (offset === "500") {
        data = [buildUser(501), buildUser(502)];
      }

      return new Response(JSON.stringify({ data }), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      });
    });

    const users = await listAllUsers();
    expect(users).toHaveLength(502);
    expect(calls).toEqual([
      { limit: "500", offset: "0" },
      { limit: "500", offset: "500" },
    ]);
  });
});
