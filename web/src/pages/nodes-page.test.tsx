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

function mockFetchForNodes(getNodes: () => unknown[]) {
  return vi.spyOn(globalThis, "fetch").mockImplementation(async (input, init) => {
    const req = asRequest(input, init);
    const url = new URL(req.url, "http://localhost");

    if (req.method === "POST" && url.pathname === "/rpc/sboard.panel.v1.NodeService/ListNodes") {
      return new Response(JSON.stringify({ data: getNodes() }), {
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

    return new Response(JSON.stringify({ data: [] }), {
      status: 200,
      headers: { "Content-Type": "application/json" },
    });
  });
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

  it("shows pending nodes banner when pending nodes exist", async () => {
    const nodes = [
      {
        id: 1,
        uuid: "abcd1234-5678-9abc-def0-1234567890ab",
        name: "",
        api_address: "10.0.0.5",
        api_port: 3000,
        secret_key: "",
        public_address: "",
        group_id: null,
        status: "pending",
        last_seen_at: "2026-03-01T12:00:00Z",
      },
      {
        id: 2,
        uuid: "efgh5678-1234-5678-abcd-ef0123456789",
        name: "node-existing",
        api_address: "10.0.0.1",
        api_port: 3000,
        secret_key: "key",
        public_address: "1.1.1.1",
        group_id: null,
        status: "online",
        last_seen_at: null,
      },
    ];

    mockFetchForNodes(() => nodes);

    render(
      <MemoryRouter>
        <AppProviders>
          <NodesPage />
        </AppProviders>
      </MemoryRouter>,
    );

    // Pending banner should appear with count
    expect(await screen.findByText("1 个节点等待认领")).toBeInTheDocument();

    // UUID prefix should be shown
    expect(screen.getByText("abcd1234")).toBeInTheDocument();

    // API address should be shown
    expect(screen.getByText("10.0.0.5:3000")).toBeInTheDocument();

    // Approve and reject buttons should be present
    expect(screen.getByRole("button", { name: "审批" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "忽略" })).toBeInTheDocument();

    // Active node should still appear in the node list
    expect((await screen.findAllByText("node-existing")).length).toBeGreaterThan(0);
  });

  it("does not show pending banner when there are no pending nodes", async () => {
    const nodes = [
      {
        id: 1,
        uuid: "n-1",
        name: "node-online",
        api_address: "127.0.0.1",
        api_port: 3000,
        secret_key: "secret",
        public_address: "1.1.1.1",
        group_id: null,
        status: "online",
        last_seen_at: null,
      },
    ];

    mockFetchForNodes(() => nodes);

    render(
      <MemoryRouter>
        <AppProviders>
          <NodesPage />
        </AppProviders>
      </MemoryRouter>,
    );

    // Wait for node to appear
    expect((await screen.findAllByText("node-online")).length).toBeGreaterThan(0);

    // Pending banner should NOT appear
    expect(screen.queryByText(/个节点等待认领/)).not.toBeInTheDocument();
  });

  it("opens approve dialog when clicking approve button", async () => {
    const nodes = [
      {
        id: 10,
        uuid: "abcd1234-5678-9abc-def0-1234567890ab",
        name: "",
        api_address: "10.0.0.5",
        api_port: 3000,
        secret_key: "",
        public_address: "",
        group_id: null,
        status: "pending",
        last_seen_at: null,
      },
    ];

    mockFetchForNodes(() => nodes);

    render(
      <MemoryRouter>
        <AppProviders>
          <NodesPage />
        </AppProviders>
      </MemoryRouter>,
    );

    // Wait for pending banner to appear, then click approve
    await screen.findByText("1 个节点等待认领");
    await userEvent.click(screen.getByRole("button", { name: "审批" }));

    // Approve dialog should open with title
    expect(await screen.findByText("审批节点")).toBeInTheDocument();

    // Node info should be shown in dialog
    expect(screen.getAllByText("abcd1234").length).toBeGreaterThanOrEqual(2);
    expect(screen.getAllByText("10.0.0.5:3000").length).toBeGreaterThanOrEqual(2);

    // Name input should be present and required
    expect(screen.getByLabelText(/节点名称/)).toBeInTheDocument();
  });

  it("calls ApproveNode RPC when form is submitted", async () => {
    let nodes = [
      {
        id: 10,
        uuid: "abcd1234-5678-9abc-def0-1234567890ab",
        name: "",
        api_address: "10.0.0.5",
        api_port: 3000,
        secret_key: "",
        public_address: "",
        group_id: null,
        status: "pending",
        last_seen_at: null,
      },
    ];
    let approvedPayload: Record<string, unknown> | null = null;

    vi.spyOn(globalThis, "fetch").mockImplementation(async (input, init) => {
      const req = asRequest(input, init);
      const url = new URL(req.url, "http://localhost");

      if (req.method === "POST" && url.pathname === "/rpc/sboard.panel.v1.NodeService/ListNodes") {
        return new Response(JSON.stringify({ data: nodes }), {
          status: 200,
          headers: { "Content-Type": "application/json" },
        });
      }

      if (req.method === "POST" && url.pathname === "/rpc/sboard.panel.v1.NodeService/ApproveNode") {
        approvedPayload = (await req.json()) as Record<string, unknown>;
        nodes = [
          {
            id: 10,
            uuid: "abcd1234-5678-9abc-def0-1234567890ab",
            name: "US-West-1",
            api_address: "10.0.0.5",
            api_port: 3000,
            secret_key: "auto-generated",
            public_address: "10.0.0.5",
            group_id: null,
            status: "offline",
            last_seen_at: null,
          },
        ];
        return new Response(
          JSON.stringify({
            data: nodes[0],
          }),
          { status: 200, headers: { "Content-Type": "application/json" } },
        );
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

    // Wait for pending banner and click approve
    await screen.findByText("1 个节点等待认领");
    await userEvent.click(screen.getByRole("button", { name: "审批" }));

    // Fill in the name
    const nameInput = await screen.findByLabelText(/节点名称/);
    await userEvent.type(nameInput, "US-West-1");

    // Submit
    const approveButtons = screen.getAllByRole("button", { name: "审批" });
    const dialogApproveButton = approveButtons[approveButtons.length - 1];
    await userEvent.click(dialogApproveButton);

    await waitFor(() => {
      expect(approvedPayload).toBeTruthy();
      expect((approvedPayload as Record<string, unknown>).name).toBe("US-West-1");
      expect((approvedPayload as Record<string, unknown>).id).toBe("10");
    });
  });
});
