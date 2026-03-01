import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { AppProviders } from "@/providers/app-providers";
import { resetAuthStore, useAuthStore } from "@/store/auth";
import { MemoryRouter } from "react-router-dom";

import { UsersPage } from "./users-page";

function asRequest(input: RequestInfo | URL, init?: RequestInit): Request {
  if (input instanceof Request) return input;
  return new Request(input, init);
}

describe("UsersPage", () => {
  beforeEach(() => {
    localStorage.clear();
    resetAuthStore();
    useAuthStore.getState().setToken("token-123");
  });

  it("renders users returned by API", async () => {
    vi.spyOn(globalThis, "fetch").mockImplementation(async (input, init) => {
      const req = asRequest(input, init);
      const url = new URL(req.url, "http://localhost");
      if (req.method === "POST" && url.pathname === "/rpc/sboard.panel.v1.UserService/ListUsers") {
        return new Response(
          JSON.stringify({
            data: [
              {
                id: "1",
                uuid: "u-1",
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
        url.pathname === "/rpc/sboard.panel.v1.GroupService/ListGroups"
      ) {
        return new Response(JSON.stringify({ data: [] }), {
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
      <MemoryRouter>
        <AppProviders>
          <UsersPage />
        </AppProviders>
      </MemoryRouter>,
    );

    expect(await screen.findByText("alice")).toBeInTheDocument();
    expect(
      screen.queryByText("已过期/流量超限通常由系统自动判定，无需手动切换。"),
    ).not.toBeInTheDocument();
    expect(screen.getByRole("table")).toHaveClass("md:table-fixed");
  });

  it("can edit user via PUT and shows updated status", async () => {
    let currentStatus: "active" | "expired" = "active";
    let currentTrafficLimit = 0;
    let currentTrafficResetDay = 0;
    const unexpectedRequests: Array<{ method: string; pathname: string }> = [];

    const fetchMock = vi.spyOn(globalThis, "fetch").mockImplementation(async (input, init) => {
      const req = asRequest(input, init);
      const url = new URL(req.url, "http://localhost");
      const pathname = url.pathname;

      if (req.method === "POST" && pathname === "/rpc/sboard.panel.v1.UserService/ListUsers") {
        return new Response(
          JSON.stringify({
            data: [
              {
                id: "1",
                uuid: "u-1",
                username: "alice",
                groupIds: [],
                trafficLimit: String(currentTrafficLimit),
                trafficUsed: "0",
                trafficResetDay: currentTrafficResetDay,
                expireAt: null,
                status: currentStatus,
              },
            ],
          }),
          { status: 200, headers: { "Content-Type": "application/json" } },
        );
      }

      if (req.method === "POST" && pathname === "/rpc/sboard.panel.v1.GroupService/ListGroups") {
        return new Response(JSON.stringify({ data: [] }), {
          status: 200,
          headers: { "Content-Type": "application/json" },
        });
      }

      if (req.method === "POST" && pathname === "/rpc/sboard.panel.v1.UserService/GetUserGroups") {
        return new Response(JSON.stringify({ groupIds: [] }), {
          status: 200,
          headers: { "Content-Type": "application/json" },
        });
      }

      if (
        req.method === "POST" &&
        pathname === "/rpc/sboard.panel.v1.UserService/ReplaceUserGroups"
      ) {
        const body = (await req.json()) as Record<string, unknown>;
        const groupIds = Array.isArray(body.groupIds) ? body.groupIds : [];
        expect(groupIds).toEqual([]);
        return new Response(JSON.stringify({ groupIds: [] }), {
          status: 200,
          headers: { "Content-Type": "application/json" },
        });
      }

      if (req.method === "POST" && pathname === "/rpc/sboard.panel.v1.UserService/UpdateUser") {
        const body = (await req.json()) as Record<string, unknown>;
        expect(body.status).toBe("expired");
        expect(Number(body.trafficLimit)).toBe(1024 * 1024 * 1024);
        expect(body.trafficResetDay).toBe(1);

        currentStatus = "expired";
        currentTrafficLimit = 1024 * 1024 * 1024;
        currentTrafficResetDay = 1;

        return new Response(
          JSON.stringify({
            data: {
              id: "1",
              uuid: "u-1",
              username: "alice",
              groupIds: [],
              trafficLimit: String(1024 * 1024 * 1024),
              trafficUsed: "0",
              trafficResetDay: 1,
              expireAt: null,
              status: "expired",
            },
          }),
          { status: 200, headers: { "Content-Type": "application/json" } },
        );
      }

      unexpectedRequests.push({ method: req.method, pathname });
      return new Response(JSON.stringify({ data: null }), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      });
    });

    render(
      <MemoryRouter>
        <AppProviders>
          <UsersPage />
        </AppProviders>
      </MemoryRouter>,
    );

    expect(await screen.findByText("alice")).toBeInTheDocument();

    const row = screen.getByRole("row", { name: /alice/i });
    await userEvent.click(within(row).getByRole("button", { name: "操作" }));
    await userEvent.click(await screen.findByRole("menuitem", { name: "编辑" }));
    await userEvent.click(screen.getByLabelText("状态"));
    await userEvent.click(await screen.findByText("已过期"));
    await userEvent.clear(screen.getByLabelText("流量上限（GB）"));
    await userEvent.type(screen.getByLabelText("流量上限（GB）"), "1");
    await userEvent.clear(screen.getByLabelText("重置日"));
    await userEvent.type(screen.getByLabelText("重置日"), "1");
    await userEvent.click(screen.getByRole("button", { name: "保存" }));

    expect(fetchMock).toHaveBeenCalled();
    const table = screen.getByRole("table");
    expect(await within(table).findByText("已过期")).toBeInTheDocument();

    // Flush any background refetches so unexpected endpoints don't surface as unhandled rejections.
    await new Promise((r) => setTimeout(r, 0));
    expect(unexpectedRequests).toEqual([]);
  });

  it("creates user and binds selected groups", async () => {
    let users: Array<Record<string, unknown>> = [];
    const unexpectedRequests: Array<{ method: string; pathname: string }> = [];

    vi.spyOn(globalThis, "fetch").mockImplementation(async (input, init) => {
      const req = asRequest(input, init);
      const url = new URL(req.url, "http://localhost");
      const pathname = url.pathname;

      if (req.method === "POST" && pathname === "/rpc/sboard.panel.v1.UserService/ListUsers") {
        return new Response(JSON.stringify({ data: users }), {
          status: 200,
          headers: { "Content-Type": "application/json" },
        });
      }

      if (req.method === "POST" && pathname === "/rpc/sboard.panel.v1.GroupService/ListGroups") {
        return new Response(
          JSON.stringify({
            data: [
              {
                id: "11",
                name: "VIP",
                description: "vip users",
                memberCount: "0",
              },
            ],
          }),
          {
            status: 200,
            headers: { "Content-Type": "application/json" },
          },
        );
      }

      if (req.method === "POST" && pathname === "/rpc/sboard.panel.v1.UserService/CreateUser") {
        const body = (await req.json()) as Record<string, unknown>;
        expect(body.username).toBe("bob");

        users = [
          {
            id: "99",
            uuid: "u-99",
            username: "bob",
            groupIds: [],
            trafficLimit: "0",
            trafficUsed: "0",
            trafficResetDay: 0,
            expireAt: null,
            status: "active",
          },
        ];

        return new Response(
          JSON.stringify({
            data: {
              id: "99",
              uuid: "u-99",
              username: "bob",
              groupIds: [],
              trafficLimit: "0",
              trafficUsed: "0",
              trafficResetDay: 0,
              expireAt: null,
              status: "active",
            },
          }),
          {
            status: 200,
            headers: { "Content-Type": "application/json" },
          },
        );
      }

      if (req.method === "POST" && pathname === "/rpc/sboard.panel.v1.UserService/UpdateUser") {
        const body = (await req.json()) as Record<string, unknown>;
        expect(body.status).toBe("active");
        expect(Number(body.trafficLimit)).toBe(2 * 1024 * 1024 * 1024);
        expect(body.trafficResetDay).toBe(5);
        expect(typeof body.expireAt).toBe("string");
        expect(String(body.expireAt)).toMatch(/-15T00:00:00(?:\.000)?Z$/);
        return new Response(
          JSON.stringify({
            data: {
              id: "99",
              uuid: "u-99",
              username: "bob",
              groupIds: [],
              trafficLimit: String(2 * 1024 * 1024 * 1024),
              trafficUsed: "0",
              trafficResetDay: 5,
              expireAt: "2030-01-15T00:00:00Z",
              status: "active",
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
        pathname === "/rpc/sboard.panel.v1.UserService/ReplaceUserGroups"
      ) {
        const body = (await req.json()) as Record<string, unknown>;
        const groupIds = Array.isArray(body.groupIds)
          ? (body.groupIds as Array<string | number>)
          : [];
        expect(groupIds.map((v) => Number(v))).toEqual([11]);
        return new Response(JSON.stringify({ groupIds: ["11"] }), {
          status: 200,
          headers: { "Content-Type": "application/json" },
        });
      }

      if (req.method === "POST" && pathname === "/rpc/sboard.panel.v1.UserService/GetUserGroups") {
        return new Response(JSON.stringify({ groupIds: ["11"] }), {
          status: 200,
          headers: { "Content-Type": "application/json" },
        });
      }

      unexpectedRequests.push({ method: req.method, pathname });
      return new Response(JSON.stringify({ data: null }), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      });
    });

    render(
      <MemoryRouter>
        <AppProviders>
          <UsersPage />
        </AppProviders>
      </MemoryRouter>,
    );

    await userEvent.click(await screen.findByRole("button", { name: "创建用户" }));
    await userEvent.type(screen.getByLabelText("用户名"), "bob");
    await userEvent.clear(screen.getByLabelText("流量上限（GB）"));
    await userEvent.type(screen.getByLabelText("流量上限（GB）"), "2");
    await userEvent.clear(screen.getByLabelText("重置日"));
    await userEvent.type(screen.getByLabelText("重置日"), "5");
    await userEvent.click(screen.getByLabelText("到期日期"));
    await userEvent.click(await screen.findByRole("button", { name: /15/ }));
    await userEvent.click(screen.getByText("VIP"));
    await userEvent.click(screen.getByRole("button", { name: "创建" }));

    expect(await screen.findByText("bob")).toBeInTheDocument();

    await new Promise((r) => setTimeout(r, 0));
    expect(unexpectedRequests).toEqual([]);
  });
});
