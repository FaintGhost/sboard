import { beforeEach, describe, expect, it, vi } from "vitest";
import { Code, ConnectError } from "@connectrpc/connect";

import { resetAuthStore, useAuthStore } from "@/store/auth";
import { authInterceptor } from "./transport";

describe("authInterceptor", () => {
  beforeEach(() => {
    localStorage.clear();
    resetAuthStore();
  });

  it("attaches Bearer token to requests", async () => {
    useAuthStore.getState().setToken("test-token");

    const req = { header: new Headers() };
    const next = vi.fn().mockResolvedValue({ status: 200 });

    await authInterceptor(next)(req as any);

    expect(req.header.get("Authorization")).toBe("Bearer test-token");
    expect(next).toHaveBeenCalledWith(req);
  });

  it("clears token when RPC returns unauthenticated error", async () => {
    useAuthStore.getState().setToken("will-be-cleared");
    expect(useAuthStore.getState().token).toBe("will-be-cleared");

    const req = { header: new Headers() };
    const next = vi
      .fn()
      .mockRejectedValue(new ConnectError("unauthenticated", Code.Unauthenticated));

    await expect(authInterceptor(next)(req as any)).rejects.toThrow(ConnectError);

    expect(useAuthStore.getState().token).toBeNull();
  });

  it("does not clear token on other errors", async () => {
    useAuthStore.getState().setToken("should-remain");

    const req = { header: new Headers() };
    const next = vi.fn().mockRejectedValue(new ConnectError("internal", Code.Internal));

    await expect(authInterceptor(next)(req as any)).rejects.toThrow(ConnectError);

    expect(useAuthStore.getState().token).toBe("should-remain");
  });

  it("clears expired token before sending request", async () => {
    useAuthStore.getState().setToken("expired-token", "2000-01-01T00:00:00.000Z");

    const req = { header: new Headers() };
    const next = vi.fn().mockResolvedValue({ status: 200 });

    await authInterceptor(next)(req as any);

    expect(req.header.get("Authorization")).toBeNull();
    expect(useAuthStore.getState().token).toBeNull();
    expect(next).toHaveBeenCalledWith(req);
  });
});
