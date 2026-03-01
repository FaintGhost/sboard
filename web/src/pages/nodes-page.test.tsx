import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { AppProviders } from "@/providers/app-providers";
import { resetAuthStore, useAuthStore } from "@/store/auth";

import { NodesPage } from "./nodes-page";

function asRequest(input: RequestInfo | URL, init?: RequestInit): Request {
  if (input instanceof Request) return input;
  return new Request(input, init);
}

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

    vi.spyOn(globalThis, "fetch").mockImplementation(async (input, init) => {
      const req = asRequest(input, init);
      const url = new URL(req.url, "http://localhost");

      if (req.method === "POST" && url.pathname === "/rpc/sboard.panel.v1.NodeService/ListNodes") {
        return new Response(JSON.stringify({ data: nodes }), {
          status: 200,
          headers: { "Content-Type": "application/json" },
        });
      }

      if (
        req.method === "POST" &&
        url.pathname === "/rpc/sboard.panel.v1.GroupService/ListGroups"
      ) {
        return new Response(JSON.stringify({ data: [] }), {
          status: 200,
          headers: { "Content-Type": "application/json" },
        });
      }

      if (
        req.method === "POST" &&
        url.pathname === "/rpc/sboard.panel.v1.TrafficService/GetTrafficNodesSummary"
      ) {
        return new Response(JSON.stringify({ data: [] }), {
          status: 200,
          headers: { "Content-Type": "application/json" },
        });
      }

      if (req.method === "POST" && url.pathname === "/rpc/sboard.panel.v1.NodeService/DeleteNode") {
        const body = (await req.json()) as { id: string; force?: boolean };
        expect(body.id).toBe("1");
        deletedIDs.push(1);
        nodes = [];
        return new Response(JSON.stringify({ status: "ok" }), {
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
    await userEvent.click(await screen.findByRole("button", { name: /确认删除|Delete/i }));

    await waitFor(() => {
      expect(deletedIDs).toEqual([1]);
      expect(screen.queryAllByText("node-a")).toHaveLength(0);
    });
  });
});
