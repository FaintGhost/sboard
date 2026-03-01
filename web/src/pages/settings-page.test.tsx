import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { AppProviders } from "@/providers/app-providers";
import { resetAuthStore, useAuthStore } from "@/store/auth";

import { SettingsPage } from "./settings-page";

function asRequest(input: RequestInfo | URL, init?: RequestInit): Request {
  if (input instanceof Request) return input;
  return new Request(input, init);
}

type AdminProfilePayload = {
  new_username: string;
  old_password: string;
  new_password: string;
  confirm_password: string;
};

type SetupMockOptions = {
  initialSubscriptionBaseURL?: string;
  initialTimezone?: string;
  onPutSystemSettings?: (payload: { subscriptionBaseUrl: string; timezone: string }) => void;
  onPutAdminProfile?: (payload: AdminProfilePayload) => { username: string };
};

function setupSettingsFetchMock(options: SetupMockOptions = {}) {
  const {
    initialSubscriptionBaseURL = "",
    initialTimezone = "UTC",
    onPutSystemSettings,
    onPutAdminProfile,
  } = options;

  vi.spyOn(globalThis, "fetch").mockImplementation(async (input, init) => {
    const req = asRequest(input, init);
    const url = new URL(req.url, "http://localhost");

    if (
      req.method === "POST" &&
      url.pathname === "/rpc/sboard.panel.v1.SystemService/GetSystemInfo"
    ) {
      return new Response(
        JSON.stringify({
          data: {
            panelVersion: "v0.2.0",
            panelCommitId: "abc1234",
            singBoxVersion: "1.12.19",
          },
        }),
        {
          status: 200,
          headers: { "Content-Type": "application/json" },
        },
      );
    }

    if (
      req.method === "POST" &&
      url.pathname === "/rpc/sboard.panel.v1.SystemService/GetSystemSettings"
    ) {
      return new Response(
        JSON.stringify({
          data: {
            subscriptionBaseUrl: initialSubscriptionBaseURL,
            timezone: initialTimezone,
          },
        }),
        {
          status: 200,
          headers: { "Content-Type": "application/json" },
        },
      );
    }

    if (
      req.method === "POST" &&
      url.pathname === "/rpc/sboard.panel.v1.AuthService/GetAdminProfile"
    ) {
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
      );
    }

    if (
      req.method === "POST" &&
      url.pathname === "/rpc/sboard.panel.v1.SystemService/UpdateSystemSettings"
    ) {
      const body = (await req.json()) as { subscriptionBaseUrl: string; timezone: string };
      onPutSystemSettings?.(body);
      return new Response(
        JSON.stringify({
          data: {
            subscriptionBaseUrl: body.subscriptionBaseUrl,
            timezone: body.timezone,
          },
        }),
        {
          status: 200,
          headers: { "Content-Type": "application/json" },
        },
      );
    }

    if (
      req.method === "POST" &&
      url.pathname === "/rpc/sboard.panel.v1.AuthService/UpdateAdminProfile"
    ) {
      const body = (await req.json()) as {
        oldPassword: string;
        newUsername?: string;
        newPassword?: string;
        confirmPassword?: string;
      };
      const updated = onPutAdminProfile?.({
        old_password: body.oldPassword,
        new_username: body.newUsername || "",
        new_password: body.newPassword || "",
        confirm_password: body.confirmPassword || "",
      }) ?? { username: body.newUsername ?? "admin" };
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
      );
    }

    return new Response(JSON.stringify({ error: "not found" }), {
      status: 404,
      headers: { "Content-Type": "application/json" },
    });
  });
}

describe("SettingsPage", () => {
  beforeEach(() => {
    localStorage.clear();
    resetAuthStore();
    useAuthStore.getState().setToken("token-123");
  });

  it("renders system info and parses subscription access parts", async () => {
    setupSettingsFetchMock({ initialSubscriptionBaseURL: "https://203.0.113.10:8443" });

    render(
      <AppProviders>
        <SettingsPage />
      </AppProviders>,
    );

    expect(await screen.findByText("v0.2.0")).toBeInTheDocument();
    expect(screen.getByText("abc1234")).toBeInTheDocument();
    expect(screen.getByText("1.12.19")).toBeInTheDocument();
    expect(await screen.findByDisplayValue("203.0.113.10:8443")).toBeInTheDocument();
    expect(await screen.findByDisplayValue("UTC")).toBeInTheDocument();
    expect(await screen.findByDisplayValue("admin")).toBeInTheDocument();
  });

  it("saves subscription protocol and ip:port", async () => {
    let savedValue = "";
    let savedTimezone = "";
    setupSettingsFetchMock({
      onPutSystemSettings: (payload) => {
        savedValue = payload.subscriptionBaseUrl;
        savedTimezone = payload.timezone;
      },
    });

    render(
      <AppProviders>
        <SettingsPage />
      </AppProviders>,
    );

    const protocolSelect = await screen.findByRole("combobox", { name: "协议" });
    await userEvent.click(protocolSelect);
    await userEvent.click(await screen.findByRole("option", { name: "HTTPS" }));

    const input = screen.getByLabelText("IP + 端口");
    await userEvent.clear(input);
    await userEvent.type(input, "203.0.113.10:8443");

    const saveButtons = screen.getAllByRole("button", { name: "保存" });
    await userEvent.click(saveButtons[0]);

    expect(savedValue).toBe("https://203.0.113.10:8443");
    expect(savedTimezone).toBe("UTC");
    expect(await screen.findByText("设置已保存")).toBeInTheDocument();
  });

  it("saves timezone from general settings", async () => {
    let savedTimezone = "";
    setupSettingsFetchMock({
      initialSubscriptionBaseURL: "https://203.0.113.10:8443",
      onPutSystemSettings: (payload) => {
        savedTimezone = payload.timezone;
      },
    });

    render(
      <AppProviders>
        <SettingsPage />
      </AppProviders>,
    );

    const timezoneInput = await screen.findByLabelText("全局时区");
    await userEvent.clear(timezoneInput);
    await userEvent.type(timezoneInput, "Asia/Shanghai");

    const generalCardTitle = await screen.findByText("通用设置");
    const generalCard = generalCardTitle.closest('[data-slot="card"]');
    expect(generalCard).toBeTruthy();
    const generalSaveButton = within(generalCard as HTMLElement).getByRole("button", {
      name: "保存",
    });
    await userEvent.click(generalSaveButton);

    await waitFor(() => {
      expect(savedTimezone).toBe("Asia/Shanghai");
    });
  });

  it("shows validation error for invalid ip:port", async () => {
    let putCalled = false;
    setupSettingsFetchMock({
      onPutSystemSettings: () => {
        putCalled = true;
      },
    });

    render(
      <AppProviders>
        <SettingsPage />
      </AppProviders>,
    );

    const input = await screen.findByLabelText("IP + 端口");
    await waitFor(() => {
      const saveButtons = screen.getAllByRole("button", { name: "保存" });
      expect(saveButtons[0]).not.toBeDisabled();
    });

    await userEvent.clear(input);
    await userEvent.type(input, "not-valid");

    const saveButtons = screen.getAllByRole("button", { name: "保存" });
    await userEvent.click(saveButtons[0]);

    await waitFor(() => {
      expect(putCalled).toBe(false);
    });
  });

  it("updates admin username and password with old password", async () => {
    let payload: AdminProfilePayload | null = null;

    setupSettingsFetchMock({
      onPutAdminProfile: (body) => {
        payload = body;
        return { username: body.new_username };
      },
    });

    render(
      <AppProviders>
        <SettingsPage />
      </AppProviders>,
    );

    await screen.findByDisplayValue("admin");
    const usernameInput = screen.getByLabelText("管理员账号");
    await userEvent.clear(usernameInput);
    await userEvent.type(usernameInput, "root");

    await userEvent.type(screen.getByLabelText("旧密码"), "pass12345");
    await userEvent.type(screen.getByLabelText("新密码"), "newpass12345");
    await userEvent.type(screen.getByLabelText("确认新密码"), "newpass12345");

    const saveButtons = screen.getAllByRole("button", { name: "保存" });
    await userEvent.click(saveButtons[1]);

    await waitFor(() => {
      expect(payload).toEqual({
        new_username: "root",
        old_password: "pass12345",
        new_password: "newpass12345",
        confirm_password: "newpass12345",
      });
    });

    expect(await screen.findByText("管理员凭证已更新。")).toBeInTheDocument();
  });
});
