import { render, screen, waitFor } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { beforeEach, describe, expect, it, vi } from "vitest"

import { AppProviders } from "@/providers/app-providers"
import { resetAuthStore, useAuthStore } from "@/store/auth"

import { SettingsPage } from "./settings-page"

type AdminProfilePayload = {
  new_username: string
  old_password: string
  new_password: string
  confirm_password: string
}

type SetupMockOptions = {
  initialSubscriptionBaseURL?: string
  onPutSystemSettings?: (value: string) => void
  onPutAdminProfile?: (payload: AdminProfilePayload) => { username: string }
}

function setupSettingsFetchMock(options: SetupMockOptions = {}) {
  const {
    initialSubscriptionBaseURL = "",
    onPutSystemSettings,
    onPutAdminProfile,
  } = options

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
            subscription_base_url: initialSubscriptionBaseURL,
          },
        }),
        {
          status: 200,
          headers: { "Content-Type": "application/json" },
        },
      )
    }

    if (req.method === "GET" && url.pathname === "/api/admin/profile") {
      return new Response(
        JSON.stringify({
          data: {
            username: "admin",
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
      onPutSystemSettings?.(body.subscription_base_url)
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

    if (req.method === "PUT" && url.pathname === "/api/admin/profile") {
      const body = (await req.json()) as AdminProfilePayload
      const updated = onPutAdminProfile?.(body) ?? { username: body.new_username }
      return new Response(
        JSON.stringify({
          data: {
            username: updated.username,
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
}

describe("SettingsPage", () => {
  beforeEach(() => {
    localStorage.clear()
    resetAuthStore()
    useAuthStore.getState().setToken("token-123")
  })

  it("renders system info and parses subscription access parts", async () => {
    setupSettingsFetchMock({ initialSubscriptionBaseURL: "https://203.0.113.10:8443" })

    render(
      <AppProviders>
        <SettingsPage />
      </AppProviders>,
    )

    expect(await screen.findByText("v0.2.0")).toBeInTheDocument()
    expect(screen.getByText("abc1234")).toBeInTheDocument()
    expect(screen.getByText("1.12.19")).toBeInTheDocument()
    expect(await screen.findByDisplayValue("203.0.113.10:8443")).toBeInTheDocument()
    expect(await screen.findByDisplayValue("admin")).toBeInTheDocument()
  })

  it("saves subscription protocol and ip:port", async () => {
    let savedValue = ""
    setupSettingsFetchMock({
      onPutSystemSettings: (value) => {
        savedValue = value
      },
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

    const saveButtons = screen.getAllByRole("button", { name: "保存" })
    await userEvent.click(saveButtons[0])

    expect(savedValue).toBe("https://203.0.113.10:8443")
    expect(await screen.findByText("设置已保存")).toBeInTheDocument()
  })

  it("shows validation error for invalid ip:port", async () => {
    let putCalled = false
    setupSettingsFetchMock({
      onPutSystemSettings: () => {
        putCalled = true
      },
    })

    render(
      <AppProviders>
        <SettingsPage />
      </AppProviders>,
    )

    const input = await screen.findByLabelText("IP + 端口")
    await waitFor(() => {
      const saveButtons = screen.getAllByRole("button", { name: "保存" })
      expect(saveButtons[0]).not.toBeDisabled()
    })

    await userEvent.clear(input)
    await userEvent.type(input, "not-valid")

    const saveButtons = screen.getAllByRole("button", { name: "保存" })
    await userEvent.click(saveButtons[0])

    await waitFor(() => {
      expect(putCalled).toBe(false)
    })
  })

  it("updates admin username and password with old password", async () => {
    let payload: AdminProfilePayload | null = null

    setupSettingsFetchMock({
      onPutAdminProfile: (body) => {
        payload = body
        return { username: body.new_username }
      },
    })

    render(
      <AppProviders>
        <SettingsPage />
      </AppProviders>,
    )

    await screen.findByDisplayValue("admin")
    const usernameInput = screen.getByLabelText("管理员账号")
    await userEvent.clear(usernameInput)
    await userEvent.type(usernameInput, "root")

    await userEvent.type(screen.getByLabelText("旧密码"), "pass12345")
    await userEvent.type(screen.getByLabelText("新密码"), "newpass12345")
    await userEvent.type(screen.getByLabelText("确认新密码"), "newpass12345")

    const saveButtons = screen.getAllByRole("button", { name: "保存" })
    await userEvent.click(saveButtons[1])

    await waitFor(() => {
      expect(payload).toEqual({
        new_username: "root",
        old_password: "pass12345",
        new_password: "newpass12345",
        confirm_password: "newpass12345",
      })
    })

    expect(await screen.findByText("管理员凭证已更新。")).toBeInTheDocument()
  })
})
