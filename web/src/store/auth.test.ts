import { beforeEach, describe, expect, it } from "vitest";

import { resetAuthStore, useAuthStore } from "./auth";

describe("auth store", () => {
  beforeEach(() => {
    localStorage.clear();
    resetAuthStore();
  });

  it("can set and clear token with expiry", () => {
    const expiresAt = "2099-01-01T00:00:00.000Z";
    useAuthStore.getState().setToken("abc", expiresAt);
    expect(useAuthStore.getState().token).toBe("abc");
    expect(useAuthStore.getState().expiresAt).toBe(expiresAt);
    expect(localStorage.getItem("sboard_token")).toBe("abc");
    expect(localStorage.getItem("sboard_token_expires_at")).toBe(expiresAt);

    useAuthStore.getState().clearToken();
    expect(useAuthStore.getState().token).toBeNull();
    expect(useAuthStore.getState().expiresAt).toBeNull();
    expect(localStorage.getItem("sboard_token")).toBeNull();
    expect(localStorage.getItem("sboard_token_expires_at")).toBeNull();
  });

  it("clears expired token when reloading from storage", () => {
    localStorage.setItem("sboard_token", "expired");
    localStorage.setItem("sboard_token_expires_at", "2000-01-01T00:00:00.000Z");

    resetAuthStore();

    expect(useAuthStore.getState().token).toBeNull();
    expect(useAuthStore.getState().expiresAt).toBeNull();
    expect(localStorage.getItem("sboard_token")).toBeNull();
    expect(localStorage.getItem("sboard_token_expires_at")).toBeNull();
  });
});
