import { beforeEach, describe, expect, it } from "vitest"

import { resetAuthStore, useAuthStore } from "./auth"

describe("auth store", () => {
  beforeEach(() => {
    localStorage.clear()
    resetAuthStore()
  })

  it("can set and clear token", () => {
    useAuthStore.getState().setToken("abc")
    expect(useAuthStore.getState().token).toBe("abc")
    expect(localStorage.getItem("sboard_token")).toBe("abc")

    useAuthStore.getState().clearToken()
    expect(useAuthStore.getState().token).toBeNull()
    expect(localStorage.getItem("sboard_token")).toBeNull()
  })
})
