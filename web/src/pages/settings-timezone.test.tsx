import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { AppProviders } from "@/providers/app-providers";
import { resetAuthStore, useAuthStore } from "@/store/auth";

import { SettingsPage } from "./settings-page";

function asRequest(input: RequestInfo | URL, init?: RequestInit): Request {
  if (input instanceof Request) return input;
  return new Request(input, init);
}

describe("SettingsPage timezone", () => {
  beforeEach(() => {
    localStorage.clear();
    resetAuthStore();
    useAuthStore.getState().setToken("token-123");
  });

  it("saves global timezone with system settings", async () => {
    let savedPayload: { subscriptionBaseUrl: string; timezone: string } | null = null;

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
          { status: 200, headers: { "Content-Type": "application/json" } },
        );
      }

      if (
        req.method === "POST" &&
        url.pathname === "/rpc/sboard.panel.v1.SystemService/GetSystemSettings"
      ) {
        return new Response(
          JSON.stringify({
            data: {
              subscriptionBaseUrl: "https://203.0.113.10:8443",
              timezone: "UTC",
            },
          }),
          { status: 200, headers: { "Content-Type": "application/json" } },
        );
      }

      if (
        req.method === "POST" &&
        url.pathname === "/rpc/sboard.panel.v1.AuthService/GetAdminProfile"
      ) {
        return new Response(JSON.stringify({ data: { username: "admin" } }), {
          status: 200,
          headers: { "Content-Type": "application/json" },
        });
      }

      if (
        req.method === "POST" &&
        url.pathname === "/rpc/sboard.panel.v1.SystemService/UpdateSystemSettings"
      ) {
        const body = (await req.json()) as { subscriptionBaseUrl: string; timezone: string };
        savedPayload = body;
        return new Response(JSON.stringify({ data: body }), {
          status: 200,
          headers: { "Content-Type": "application/json" },
        });
      }

      if (
        req.method === "POST" &&
        url.pathname === "/rpc/sboard.panel.v1.AuthService/UpdateAdminProfile"
      ) {
        const body = (await req.json()) as { newUsername?: string };
        return new Response(JSON.stringify({ data: { username: body.newUsername ?? "admin" } }), {
          status: 200,
          headers: { "Content-Type": "application/json" },
        });
      }

      return new Response(JSON.stringify({ error: "not found" }), {
        status: 404,
        headers: { "Content-Type": "application/json" },
      });
    });

    render(
      <AppProviders>
        <SettingsPage />
      </AppProviders>,
    );

    const timezoneInput = await screen.findByLabelText("全局时区");
    await userEvent.clear(timezoneInput);
    await userEvent.type(timezoneInput, "Asia/Shanghai");

    const saveButtons = screen.getAllByRole("button", { name: "保存" });
    await userEvent.click(saveButtons[0]);

    await waitFor(() => {
      expect(savedPayload).toEqual({
        subscriptionBaseUrl: "https://203.0.113.10:8443",
        timezone: "Asia/Shanghai",
      });
    });
  });
});
