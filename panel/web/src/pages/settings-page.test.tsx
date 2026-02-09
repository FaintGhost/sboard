import { render, screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
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

  it("renders system info and subscription base URL from API", async () => {
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

      if (req.method === "GET" && url.pathname === "/api/system/settings") {
        return new Response(
          JSON.stringify({
            data: {
              subscription_base_url: "https://sub.example.com",
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
    expect(await screen.findByDisplayValue("https://sub.example.com")).toBeInTheDocument()
  })

  it("saves subscription base URL", async () => {
    let savedValue = ""

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

      if (req.method === "GET" && url.pathname === "/api/system/settings") {
        return new Response(
          JSON.stringify({
            data: {
              subscription_base_url: "",
            },
          }),
          {
            status: 200,
            headers: { "Content-Type": "application/json" },
          },
        )
      }

      if (req.method === "PUT" && url.pathname === "/api/system/settings") {
        const body = (await req.json()) as { subscription_base_url: string }
        savedValue = body.subscription_base_url
        return new Response(
          JSON.stringify({
            data: {
              subscription_base_url: body.subscription_base_url,
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

    const input = await screen.findByLabelText("订阅基础地址（域名或公网 IP）")
    await userEvent.type(input, "https://sub.example.com")
    await userEvent.click(screen.getByRole("button", { name: "保存" }))

    expect(savedValue).toBe("https://sub.example.com")
    expect(await screen.findByText("设置已保存")).toBeInTheDocument()
  })
})
