import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { AppProviders } from "@/providers/app-providers";
import { resetAuthStore, useAuthStore } from "@/store/auth";

import { SettingsPage } from "./settings-page";

describe("SettingsPage timezone", () => {
  beforeEach(() => {
    localStorage.clear();
    resetAuthStore();
    useAuthStore.getState().setToken("token-123");
  });

  it("saves global timezone with system settings", async () => {
    let savedPayload: { subscription_base_url: string; timezone: string } | null = null;

    vi.spyOn(globalThis, "fetch").mockImplementation(async (input) => {
      const req = input as Request;
      const url = new URL(req.url);

      if (req.method === "GET" && url.pathname === "/api/system/info") {
        return new Response(
          JSON.stringify({
            data: {
              panel_version: "v0.2.0",
              panel_commit_id: "abc1234",
              sing_box_version: "1.12.19",
            },
          }),
          { status: 200, headers: { "Content-Type": "application/json" } },
        );
      }

      if (req.method === "GET" && url.pathname === "/api/system/settings") {
        return new Response(
          JSON.stringify({
            data: {
              subscription_base_url: "https://203.0.113.10:8443",
              timezone: "UTC",
            },
          }),
          { status: 200, headers: { "Content-Type": "application/json" } },
        );
      }

      if (req.method === "GET" && url.pathname === "/api/admin/profile") {
        return new Response(JSON.stringify({ data: { username: "admin" } }), {
          status: 200,
          headers: { "Content-Type": "application/json" },
        });
      }

      if (req.method === "PUT" && url.pathname === "/api/system/settings") {
        const body = (await req.json()) as { subscription_base_url: string; timezone: string };
        savedPayload = body;
        return new Response(JSON.stringify({ data: body }), {
          status: 200,
          headers: { "Content-Type": "application/json" },
        });
      }

      if (req.method === "PUT" && url.pathname === "/api/admin/profile") {
        const body = await req.json();
        return new Response(JSON.stringify({ data: { username: body.new_username ?? "admin" } }), {
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
        subscription_base_url: "https://203.0.113.10:8443",
        timezone: "Asia/Shanghai",
      });
    });
  });
});
