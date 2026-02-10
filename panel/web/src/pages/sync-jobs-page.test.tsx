import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { AppProviders } from "@/providers/app-providers";
import { resetAuthStore, useAuthStore } from "@/store/auth";

import { SyncJobsPage } from "./sync-jobs-page";

function job(id: number, status: string, source: string) {
  return {
    id,
    node_id: 1,
    parent_job_id: null,
    trigger_source: source,
    status,
    inbound_count: 1,
    active_user_count: 1,
    payload_hash: `hash-${id}`,
    attempt_count: 1,
    duration_ms: 50,
    error_summary: "",
    created_at: "2026-02-09T00:00:00Z",
    started_at: "2026-02-09T00:00:00Z",
    finished_at: "2026-02-09T00:00:00Z",
  };
}

function renderPage() {
  render(
    <MemoryRouter>
      <AppProviders>
        <SyncJobsPage />
      </AppProviders>
    </MemoryRouter>,
  );
}

describe("SyncJobsPage", () => {
  beforeEach(() => {
    localStorage.clear();
    resetAuthStore();
    useAuthStore.getState().setToken("token-123");
  });

  it("does not flash stale rows when switching status filter", async () => {
    let failedCount = 0;
    let resolveSecondFailed = () => {};

    vi.spyOn(globalThis, "fetch").mockImplementation(async (input) => {
      const req = input as Request;
      const url = new URL(req.url);

      if (req.method === "GET" && url.pathname === "/api/nodes") {
        return new Response(
          JSON.stringify({
            data: [
              {
                id: 1,
                uuid: "n-1",
                name: "node-1",
                api_address: "127.0.0.1",
                api_port: 8080,
                secret_key: "sec",
                public_address: "1.1.1.1",
                group_id: 1,
                status: "online",
              },
            ],
          }),
          { status: 200, headers: { "Content-Type": "application/json" } },
        );
      }

      if (req.method === "GET" && url.pathname === "/api/sync-jobs") {
        const status = url.searchParams.get("status");
        if (status === "failed") {
          failedCount += 1;
          if (failedCount === 1) {
            return new Response(JSON.stringify({ data: [job(2001, "failed", "src-failed-old")] }), {
              status: 200,
              headers: { "Content-Type": "application/json" },
            });
          }

          await new Promise<void>((resolve) => {
            resolveSecondFailed = resolve;
          });

          return new Response(JSON.stringify({ data: [] }), {
            status: 200,
            headers: { "Content-Type": "application/json" },
          });
        }

        return new Response(JSON.stringify({ data: [job(1001, "success", "src-all")] }), {
          status: 200,
          headers: { "Content-Type": "application/json" },
        });
      }

      return new Response(JSON.stringify({ error: "not found" }), {
        status: 404,
        headers: { "Content-Type": "application/json" },
      });
    });

    renderPage();

    expect(await screen.findByText("src-all")).toBeInTheDocument();

    await userEvent.click(screen.getByRole("combobox", { name: "状态筛选" }));
    await userEvent.click(await screen.findByRole("option", { name: "失败" }));

    expect(await screen.findByText("src-failed-old")).toBeInTheDocument();

    await userEvent.click(screen.getByRole("combobox", { name: "状态筛选" }));
    await userEvent.click(await screen.findByRole("option", { name: "全部状态" }));
    expect(await screen.findByText("src-all")).toBeInTheDocument();

    await userEvent.click(screen.getByRole("combobox", { name: "状态筛选" }));
    await userEvent.click(await screen.findByRole("option", { name: "失败" }));

    const table = screen.getByRole("table");
    expect(within(table).queryByText("src-failed-old")).not.toBeInTheDocument();

    resolveSecondFailed();

    expect(await screen.findByText("暂无数据")).toBeInTheDocument();
  });

  it("does not flash all-source rows when switching between empty source filters", async () => {
    let resolveRetrySource = () => {};

    vi.spyOn(globalThis, "fetch").mockImplementation(async (input) => {
      const req = input as Request;
      const url = new URL(req.url);

      if (req.method === "GET" && url.pathname === "/api/nodes") {
        return new Response(
          JSON.stringify({
            data: [
              {
                id: 1,
                uuid: "n-1",
                name: "node-1",
                api_address: "127.0.0.1",
                api_port: 8080,
                secret_key: "sec",
                public_address: "1.1.1.1",
                group_id: 1,
                status: "online",
              },
            ],
          }),
          { status: 200, headers: { "Content-Type": "application/json" } },
        );
      }

      if (req.method === "GET" && url.pathname === "/api/sync-jobs") {
        const source = url.searchParams.get("trigger_source");

        if (source === "manual_node_sync") {
          return new Response(JSON.stringify({ data: [] }), {
            status: 200,
            headers: { "Content-Type": "application/json" },
          });
        }

        if (source === "manual_retry") {
          await new Promise<void>((resolve) => {
            resolveRetrySource = resolve;
          });

          return new Response(JSON.stringify({ data: [] }), {
            status: 200,
            headers: { "Content-Type": "application/json" },
          });
        }

        return new Response(
          JSON.stringify({
            data: [
              job(1001, "success", "src-all-1"),
              job(1002, "success", "src-all-2"),
              job(1003, "failed", "src-all-3"),
            ],
          }),
          { status: 200, headers: { "Content-Type": "application/json" } },
        );
      }

      return new Response(JSON.stringify({ error: "not found" }), {
        status: 404,
        headers: { "Content-Type": "application/json" },
      });
    });

    renderPage();

    expect(await screen.findByText("src-all-1")).toBeInTheDocument();

    await userEvent.click(screen.getByRole("combobox", { name: "触发来源筛选" }));
    await userEvent.click(await screen.findByRole("option", { name: "手动节点同步" }));
    expect(await screen.findByText("暂无数据")).toBeInTheDocument();

    await userEvent.click(screen.getByRole("combobox", { name: "触发来源筛选" }));
    await userEvent.click(await screen.findByRole("option", { name: "手动重试" }));

    await new Promise((resolve) => setTimeout(resolve, 0));

    const table = screen.getByRole("table");
    expect(within(table).queryByText("src-all-1")).not.toBeInTheDocument();
    expect(within(table).queryByText("src-all-2")).not.toBeInTheDocument();
    expect(within(table).queryByText("src-all-3")).not.toBeInTheDocument();
    expect(screen.getByText("暂无数据")).toBeInTheDocument();
    expect(screen.queryByText("加载中...")).not.toBeInTheDocument();

    resolveRetrySource();
    expect(await screen.findByText("暂无数据")).toBeInTheDocument();
  });
});
