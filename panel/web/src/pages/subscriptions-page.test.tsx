import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { AppProviders } from "@/providers/app-providers";
import { TooltipProvider } from "@/components/ui/tooltip";
import { resetAuthStore, useAuthStore } from "@/store/auth";
import { MemoryRouter } from "react-router-dom";

import { SubscriptionsPage } from "./subscriptions-page";

function asRequest(input: RequestInfo | URL, init?: RequestInit): Request {
  if (input instanceof Request) return input;
  return new Request(input, init);
}

describe("SubscriptionsPage", () => {
  beforeEach(() => {
    localStorage.clear();
    resetAuthStore();
    useAuthStore.getState().setToken("token-123");
  });

  it("uses configured subscription base URL", async () => {
    vi.spyOn(globalThis, "fetch").mockImplementation(async (input, init) => {
      const req = asRequest(input, init);
      const url = new URL(req.url, "http://localhost");

      if (req.method === "POST" && url.pathname === "/rpc/sboard.panel.v1.UserService/ListUsers") {
        return new Response(
          JSON.stringify({
            data: [
              {
                id: "1",
                uuid: "67e59f3f-412e-4f46-92cd-495aed76ee35",
                username: "alice",
                groupIds: [],
                trafficLimit: "0",
                trafficUsed: "0",
                trafficResetDay: 0,
                expireAt: null,
                status: "active",
              },
            ],
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
              subscriptionBaseUrl: "https://sub.example.com",
              timezone: "UTC",
            },
          }),
          { status: 200, headers: { "Content-Type": "application/json" } },
        );
      }

      return new Response(JSON.stringify({ error: "not found" }), {
        status: 404,
        headers: { "Content-Type": "application/json" },
      });
    });

    render(
      <MemoryRouter>
        <AppProviders>
          <TooltipProvider>
            <SubscriptionsPage />
          </TooltipProvider>
        </AppProviders>
      </MemoryRouter>,
    );

    expect(await screen.findByText("alice")).toBeInTheDocument();
    expect(
      screen.getByText("https://sub.example.com/api/sub/67e59f3f-412e-4f46-92cd-495aed76ee35"),
    ).toBeInTheDocument();
  });
});
