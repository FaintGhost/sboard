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
    vi.spyOn(globalThis, "fetch").mockImplementation(async (input) => {
      const req = input as Request
      const url = new URL(req.url)
      if (req.method === "GET" && url.pathname === "/api/users") {
        return new Response(
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
        )
      }
      if (req.method === "GET" && url.pathname === "/api/groups") {
        return new Response(JSON.stringify({ data: [] }), {
          status: 200,
          headers: { "Content-Type": "application/json" },
        })
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
    expect(
      screen.queryByText("已过期/流量超限通常由系统自动判定，无需手动切换。"),
    ).not.toBeInTheDocument()
    expect(screen.getByRole("table")).toHaveClass("table-fixed")
  })

  it("can edit user via PUT and shows updated status", async () => {
    let currentStatus: "active" | "expired" = "active"
    let currentTrafficLimit = 0
    let currentTrafficResetDay = 0
    const unexpectedRequests: Array<{ method: string; pathname: string }> = []

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

        if (req.method === "GET" && pathname === "/api/groups") {
          return new Response(JSON.stringify({ data: [] }), {
            status: 200,
            headers: { "Content-Type": "application/json" },
          })
        }

        if (req.method === "GET" && pathname === "/api/users/1/groups") {
          return new Response(JSON.stringify({ data: { group_ids: [] } }), {
            status: 200,
            headers: { "Content-Type": "application/json" },
          })
        }

        if (req.method === "PUT" && pathname === "/api/users/1/groups") {
          const body = (await req.json()) as Record<string, unknown>
          expect(body.group_ids).toEqual([])
          return new Response(JSON.stringify({ data: { group_ids: [] } }), {
            status: 200,
            headers: { "Content-Type": "application/json" },
          })
        }

        if (req.method === "PUT" && pathname === "/api/users/1") {
          const body = (await req.json()) as Record<string, unknown>
          expect(body.status).toBe("expired")
          expect(body.traffic_limit).toBe(1024 * 1024 * 1024)
          expect(body.traffic_reset_day).toBe(1)

          currentStatus = "expired"
          currentTrafficLimit = 1024 * 1024 * 1024
          currentTrafficResetDay = 1

          return new Response(
            JSON.stringify({
              data: {
                id: 1,
                uuid: "u-1",
                username: "alice",
                traffic_limit: 1024 * 1024 * 1024,
                traffic_used: 0,
                traffic_reset_day: 1,
                expire_at: null,
                status: "expired",
              },
            }),
            { status: 200, headers: { "Content-Type": "application/json" } },
          )
        }

        unexpectedRequests.push({ method: req.method, pathname })
        return new Response(JSON.stringify({ data: null }), {
          status: 200,
          headers: { "Content-Type": "application/json" },
        })
      })

    render(
      <AppProviders>
        <UsersPage />
      </AppProviders>,
    )

    expect(await screen.findByText("alice")).toBeInTheDocument()

    const row = screen.getByRole("row", { name: /alice/i })
    await userEvent.click(within(row).getByRole("button", { name: "操作" }))
    await userEvent.click(await screen.findByRole("menuitem", { name: "编辑" }))
    await userEvent.click(screen.getByLabelText("状态"))
    await userEvent.click(await screen.findByText("已过期"))
    await userEvent.clear(screen.getByLabelText("流量上限（GB）"))
    await userEvent.type(screen.getByLabelText("流量上限（GB）"), "1")
    await userEvent.clear(screen.getByLabelText("重置日"))
    await userEvent.type(screen.getByLabelText("重置日"), "1")
    await userEvent.click(screen.getByRole("button", { name: "保存" }))

    expect(fetchMock).toHaveBeenCalled()
    const table = screen.getByRole("table")
    expect(await within(table).findByText("已过期")).toBeInTheDocument()

    // Flush any background refetches so unexpected endpoints don't surface as unhandled rejections.
    await new Promise((r) => setTimeout(r, 0))
    expect(unexpectedRequests).toEqual([])
  })
})
