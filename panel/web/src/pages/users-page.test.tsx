import { render, screen } from "@testing-library/react"
import { beforeEach, describe, expect, it, vi } from "vitest"

import { AppProviders } from "@/providers/app-providers"
import { resetAuthStore, useAuthStore } from "@/store/auth"

import { UsersPage } from "./users-page"

describe("UsersPage", () => {
  beforeEach(() => {
    localStorage.clear()
    resetAuthStore()
    useAuthStore.getState().setToken("token-123")
  })

  it("renders users returned by API", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(
        JSON.stringify({
          data: [
            {
              id: 1,
              uuid: "u-1",
              username: "alice",
              traffic_limit: 0,
              traffic_used: 0,
              traffic_reset_day: 0,
              expire_at: null,
              status: "active",
            },
          ],
        }),
        { status: 200, headers: { "Content-Type": "application/json" } },
      ),
    )

    render(
      <AppProviders>
        <UsersPage />
      </AppProviders>,
    )

    expect(await screen.findByText("alice")).toBeInTheDocument()
  })
})

