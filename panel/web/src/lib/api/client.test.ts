import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { ApiError } from "./client";
import { client } from "./gen/client.gen";
import { resetAuthStore, useAuthStore } from "@/store/auth";

describe("API client configuration", () => {
  beforeEach(() => {
    localStorage.clear();
    resetAuthStore();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("provides auth token via client config", () => {
    useAuthStore.getState().setToken("token-123");
    const config = client.getConfig();
    expect(typeof config.auth).toBe("function");
    const token = (config.auth as () => string | undefined)();
    expect(token).toBe("token-123");
  });

  it("returns undefined when no token", () => {
    const config = client.getConfig();
    const token = (config.auth as () => string | undefined)();
    expect(token).toBeUndefined();
  });

  it("error interceptor clears token on 401", async () => {
    useAuthStore.getState().setToken("token-123");

    const interceptorFns = client.interceptors.error.fns;
    expect(interceptorFns.length).toBeGreaterThan(0);

    const fn = interceptorFns[0]!;
    const mockResponse = new Response(null, { status: 401 });

    try {
      await fn({ error: "unauthorized" }, mockResponse, new Request("http://localhost"), {
        url: "/test",
      } as any);
    } catch (e) {
      expect(e).toBeInstanceOf(ApiError);
      expect((e as ApiError).status).toBe(401);
      expect((e as ApiError).message).toBe("unauthorized");
    }

    expect(useAuthStore.getState().token).toBeNull();
  });

  it("error interceptor throws ApiError on non-401 errors", async () => {
    const interceptorFns = client.interceptors.error.fns;
    const fn = interceptorFns[0]!;
    const mockResponse = new Response(null, { status: 500 });

    try {
      await fn({ error: "internal error" }, mockResponse, new Request("http://localhost"), {
        url: "/test",
      } as any);
    } catch (e) {
      expect(e).toBeInstanceOf(ApiError);
      expect((e as ApiError).status).toBe(500);
      expect((e as ApiError).message).toBe("internal error");
    }
  });
});
