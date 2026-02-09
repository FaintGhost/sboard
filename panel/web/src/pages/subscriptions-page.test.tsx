import { render, screen } from "@testing-library/react"
import { beforeEach, describe, expect, it, vi } from "vitest"

import { AppProviders } from "@/providers/app-providers"
import { TooltipProvider } from "@/components/ui/tooltip"
import { resetAuthStore, useAuthStore } from "@/store/auth"

import { SubscriptionsPage } from "./subscriptions-page"

describe("SubscriptionsPage", () => {
  beforeEach(() => {
    localStorage.clear()
    resetAuthStore()
    useAuthStore.getState().setToken("token-123")
  })

  it("uses configured subscription base URL", async () => {
    vi.spyOn(globalThis, "fetch").mockImplementation(async (input) => {
      const req = input as Request
      const url = new URL(req.url)

      if (req.method === "GET" && url.pathname === "/api/users") {
        return new Response(
          JSON.stringify({
            data: [
              {
                id: 1,
                uuid: "67e59f3f-412e-4f46-92cd-495aed76ee35",
                username: "alice",
                group_ids: [],
                traffic_limit: 0,
                traffic_used: 0,
                traffic_reset_day: 0,
                expire_at: null,
                status: "active",
              },
            ],
          }),
          { status: 200, headers: { "Content-Type": "application/json" } },
        )
      }

      if (req.method === "GET" && url.pathname === "/api/system/settings") {
        return new Response(
          JSON.stringify({
            data: {
              subscription_base_url: "https://sub.example.com",
            },
          }),
          { status: 200, headers: { "Content-Type": "application/json" } },
        )
      }

      return new Response(JSON.stringify({ error: "not found" }), {
        status: 404,
        headers: { "Content-Type": "application/json" },
      })
    })

    render(
      <AppProviders>
        <TooltipProvider>
          <SubscriptionsPage />
        </TooltipProvider>
      </AppProviders>,
    )

    expect(await screen.findByText("alice")).toBeInTheDocument()
    expect(
      screen.getByText("https://sub.example.com/api/sub/67e59f3f-412e-4f46-92cd-495aed76ee35"),
    ).toBeInTheDocument()
  })
})
