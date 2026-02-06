import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"

import { apiRequest } from "./client"
import { resetAuthStore, useAuthStore } from "@/store/auth"

describe("apiRequest", () => {
  beforeEach(() => {
    localStorage.clear()
    resetAuthStore()
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it("injects bearer token", async () => {
    useAuthStore.getState().setToken("token-123")

    const fetchMock = vi
      .spyOn(globalThis, "fetch")
      .mockResolvedValue(
        new Response(JSON.stringify({ data: { ok: true } }), {
          status: 200,
          headers: { "Content-Type": "application/json" },
        }),
      )

    const result = await apiRequest<{ ok: boolean }>("/api/health")

    expect(result.ok).toBe(true)
    expect(fetchMock).toHaveBeenCalledOnce()

    const request = fetchMock.mock.calls[0]?.[0]
    expect(request).toBeInstanceOf(Request)
    expect((request as Request).headers.get("Authorization")).toBe(
      "Bearer token-123",
    )
  })

  it("clears token on 401", async () => {
    useAuthStore.getState().setToken("token-123")
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ error: "unauthorized" }), {
        status: 401,
        headers: { "Content-Type": "application/json" },
      }),
    )

    await expect(apiRequest("/api/users")).rejects.toThrow("unauthorized")
    expect(useAuthStore.getState().token).toBeNull()
  })
})
