import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { AppProviders } from "@/providers/app-providers";
import { resetAuthStore, useAuthStore } from "@/store/auth";

import { NodesPage } from "./nodes-page";

describe("NodesPage", () => {
  beforeEach(() => {
    localStorage.clear();
    resetAuthStore();
    useAuthStore.getState().setToken("token-123");
  });

  it("supports deleting node from row actions", async () => {
    let nodes = [
      {
        id: 1,
        uuid: "n-1",
        name: "node-a",
        api_address: "127.0.0.1",
        api_port: 3000,
        secret_key: "secret",
        public_address: "1.1.1.1",
        group_id: null,
        status: "offline",
        last_seen_at: null,
      },
    ];
    const deletedIDs: number[] = [];

    vi.spyOn(globalThis, "fetch").mockImplementation(async (input) => {
      const req = input as Request;
      const url = new URL(req.url);

      if (req.method === "GET" && url.pathname === "/api/nodes") {
        return new Response(JSON.stringify({ data: nodes }), {
          status: 200,
          headers: { "Content-Type": "application/json" },
        });
      }

      if (req.method === "GET" && url.pathname === "/api/groups") {
        return new Response(JSON.stringify({ data: [] }), {
          status: 200,
          headers: { "Content-Type": "application/json" },
        });
      }

      if (req.method === "GET" && url.pathname === "/api/traffic/nodes/summary") {
        return new Response(JSON.stringify({ data: [] }), {
          status: 200,
          headers: { "Content-Type": "application/json" },
        });
      }

      if (req.method === "DELETE" && url.pathname === "/api/nodes/1") {
        deletedIDs.push(1);
        nodes = [];
        return new Response(JSON.stringify({ data: { status: "ok" } }), {
          status: 200,
          headers: { "Content-Type": "application/json" },
        });
      }

      return new Response(JSON.stringify({ data: [] }), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      });
    });

    render(
      <MemoryRouter>
        <AppProviders>
          <NodesPage />
        </AppProviders>
      </MemoryRouter>,
    );

    expect((await screen.findAllByText("node-a")).length).toBeGreaterThan(0);

    const actionButtons = await screen.findAllByRole("button", { name: "操作" });
    await userEvent.click(actionButtons[0]);
    await userEvent.click(await screen.findByRole("menuitem", { name: "删除节点" }));

    await waitFor(() => {
      expect(screen.queryByText("node-a")).not.toBeInTheDocument();
    });

    expect(deletedIDs).toEqual([1]);
  });
});
