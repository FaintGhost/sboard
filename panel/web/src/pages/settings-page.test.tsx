import { render, screen, waitFor } from "@testing-library/react"
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

  it("renders system info and parses subscription access parts", async () => {
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
              subscription_base_url: "https://203.0.113.10:8443",
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
    expect(await screen.findByDisplayValue("203.0.113.10:8443")).toBeInTheDocument()
  })

  it("saves subscription protocol and ip:port", async () => {
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

    const protocolSelect = await screen.findByRole("combobox", { name: "协议" })
    await userEvent.click(protocolSelect)
    await userEvent.click(await screen.findByRole("option", { name: "HTTPS" }))

    const input = screen.getByLabelText("IP + 端口")
    await userEvent.clear(input)
    await userEvent.type(input, "203.0.113.10:8443")
    await userEvent.click(screen.getByRole("button", { name: "保存" }))

    expect(savedValue).toBe("https://203.0.113.10:8443")
    expect(await screen.findByText("设置已保存")).toBeInTheDocument()
  })

  it("shows validation error for invalid ip:port", async () => {
    let putCalled = false

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
        putCalled = true
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

    const input = await screen.findByLabelText("IP + 端口")
    await waitFor(() => {
      expect(screen.getByRole("button", { name: "保存" })).not.toBeDisabled()
    })

    await userEvent.clear(input)
    await userEvent.type(input, "not-valid")
    await userEvent.click(screen.getByRole("button", { name: "保存" }))

    await waitFor(() => {
      expect(putCalled).toBe(false)
    })
  })
})
