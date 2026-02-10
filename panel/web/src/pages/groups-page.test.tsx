import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { AppProviders } from "@/providers/app-providers";
import { resetAuthStore, useAuthStore } from "@/store/auth";

import { GroupsPage } from "./groups-page";

describe("GroupsPage", () => {
  beforeEach(() => {
    localStorage.clear();
    resetAuthStore();
    useAuthStore.getState().setToken("token-123");
  });

  it("does not call replace-group-users when only group name/description changed", async () => {
    let updateCalls = 0;
    let replaceCalls = 0;

    vi.spyOn(globalThis, "fetch").mockImplementation(async (input) => {
      const req = input as Request;
      const url = new URL(req.url);

      if (req.method === "GET" && url.pathname === "/api/groups") {
        return new Response(
          JSON.stringify({
            data: [
              {
                id: 1,
                name: "miot",
                description: "desc",
                member_count: 1,
              },
            ],
          }),
          { status: 200, headers: { "Content-Type": "application/json" } },
        );
      }

      if (req.method === "GET" && url.pathname === "/api/users") {
        return new Response(
          JSON.stringify({
            data: [
              {
                id: 10,
                uuid: "u-10",
                username: "admin",
                group_ids: [1],
                traffic_limit: 0,
                traffic_used: 0,
                traffic_reset_day: 1,
                expire_at: null,
                status: "active",
              },
            ],
          }),
          { status: 200, headers: { "Content-Type": "application/json" } },
        );
      }

      if (req.method === "GET" && url.pathname === "/api/groups/1/users") {
        return new Response(
          JSON.stringify({
            data: [
              {
                id: 10,
                uuid: "u-10",
                username: "admin",
                traffic_limit: 0,
                traffic_used: 0,
                status: "active",
              },
            ],
          }),
          { status: 200, headers: { "Content-Type": "application/json" } },
        );
      }

      if (req.method === "PUT" && url.pathname === "/api/groups/1") {
        updateCalls += 1;
        const body = (await req.json()) as { name?: string; description?: string };
        return new Response(
          JSON.stringify({
            data: {
              id: 1,
              name: body.name ?? "miot",
              description: body.description ?? "desc",
              member_count: 1,
            },
          }),
          { status: 200, headers: { "Content-Type": "application/json" } },
        );
      }

      if (req.method === "PUT" && url.pathname === "/api/groups/1/users") {
        replaceCalls += 1;
        return new Promise<Response>(() => {
          // keep pending to expose bug: unnecessary membership replace causes endless saving
        });
      }

      return new Response(JSON.stringify({ error: "not found" }), {
        status: 404,
        headers: { "Content-Type": "application/json" },
      });
    });

    render(
      <AppProviders>
        <GroupsPage />
      </AppProviders>,
    );

    await screen.findByText("miot");

    const actionsButton = screen.getByRole("button", { name: /操作|actions/i });
    await userEvent.click(actionsButton);
    await userEvent.click(await screen.findByRole("menuitem", { name: /编辑|edit/i }));

    const nameInput = await screen.findByLabelText(/分组名称|group name/i);
    await userEvent.clear(nameInput);
    await userEvent.type(nameInput, "miot-new");

    const saveButton = screen.getByRole("button", { name: /保存|save/i });
    await waitFor(() => {
      expect(saveButton).not.toBeDisabled();
    });
    await userEvent.click(saveButton);

    await waitFor(() => {
      expect(updateCalls).toBe(1);
    });

    await waitFor(
      () => {
        expect(screen.queryByLabelText(/分组名称|group name/i)).not.toBeInTheDocument();
      },
      { timeout: 500 },
    );

    expect(replaceCalls).toBe(0);
  });
});
