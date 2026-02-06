import { render, screen, within } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
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

  it("can edit user via PUT and shows updated status", async () => {
    let currentStatus: "active" | "expired" = "active"
    let currentTrafficLimit = 0
    let currentTrafficResetDay = 0

    const fetchMock = vi
      .spyOn(globalThis, "fetch")
      .mockImplementation(async (input) => {
        const req = input as Request
        const url = new URL(req.url)
        const pathname = url.pathname

        if (req.method === "GET" && pathname === "/api/users") {
          return new Response(
            JSON.stringify({
              data: [
                {
                  id: 1,
                  uuid: "u-1",
                  username: "alice",
                  traffic_limit: currentTrafficLimit,
                  traffic_used: 0,
                  traffic_reset_day: currentTrafficResetDay,
                  expire_at: null,
                  status: currentStatus,
                },
              ],
            }),
            { status: 200, headers: { "Content-Type": "application/json" } },
          )
        }

        if (req.method === "PUT" && pathname === "/api/users/1") {
          const body = (await req.json()) as Record<string, unknown>
          expect(body.status).toBe("expired")
          expect(body.traffic_limit).toBe(1024)
          expect(body.traffic_reset_day).toBe(1)

          currentStatus = "expired"
          currentTrafficLimit = 1024
          currentTrafficResetDay = 1

          return new Response(
            JSON.stringify({
              data: {
                id: 1,
                uuid: "u-1",
                username: "alice",
                traffic_limit: 1024,
                traffic_used: 0,
                traffic_reset_day: 1,
                expire_at: null,
                status: "expired",
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
        <UsersPage />
      </AppProviders>,
    )

    expect(await screen.findByText("alice")).toBeInTheDocument()

    await userEvent.click(screen.getByRole("button", { name: "编辑" }))
    await userEvent.selectOptions(screen.getByLabelText("状态"), "expired")
    await userEvent.clear(screen.getByLabelText("流量上限"))
    await userEvent.type(screen.getByLabelText("流量上限"), "1024")
    await userEvent.clear(screen.getByLabelText("重置日"))
    await userEvent.type(screen.getByLabelText("重置日"), "1")
    await userEvent.click(screen.getByRole("button", { name: "保存" }))

    expect(fetchMock).toHaveBeenCalled()
    const table = screen.getByRole("table")
    expect(await within(table).findByText("expired")).toBeInTheDocument()
  })
})
