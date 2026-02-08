import { render, screen } from "@testing-library/react"
import { beforeEach, describe, expect, it, vi } from "vitest"

import { AppProviders } from "@/providers/app-providers"
import { resetAuthStore, useAuthStore } from "@/store/auth"

import { SettingsPage } from "./settings-page"

describe("SettingsPage", () => {
  beforeEach(() => {
    localStorage.clear()
    resetAuthStore()
    useAuthStore.getState().setToken("token-123")
  })

  it("renders system info from API", async () => {
    vi.spyOn(globalThis, "fetch").mockImplementation(async (input) => {
      const req = input as Request
      const url = new URL(req.url)

      if (req.method === "GET" && url.pathname === "/api/system/info") {
        return new Response(
          JSON.stringify({
            data: {
              panel_version: "v0.2.0",
              panel_commit_id: "abc1234",
              sing_box_version: "1.12.19",
            },
          }),
          {
            status: 200,
            headers: { "Content-Type": "application/json" },
          },
        )
      }

      return new Response(JSON.stringify({ error: "not found" }), {
        status: 404,
        headers: { "Content-Type": "application/json" },
      })
    })

    render(
      <AppProviders>
        <SettingsPage />
      </AppProviders>,
    )

    expect(await screen.findByText("v0.2.0")).toBeInTheDocument()
    expect(screen.getByText("abc1234")).toBeInTheDocument()
    expect(screen.getByText("1.12.19")).toBeInTheDocument()
  })
})

