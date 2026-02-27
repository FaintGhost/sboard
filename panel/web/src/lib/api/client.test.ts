import { describe, expect, it } from "vitest";
import { Code, ConnectError } from "@connectrpc/connect";

import { ApiError, toApiError } from "./client";

describe("API error adapter", () => {
  it("maps Connect unauthenticated to 401", () => {
    const e = new ConnectError("unauthorized", Code.Unauthenticated);
    const mapped = toApiError(e);
    expect(mapped).toBeInstanceOf(ApiError);
    expect(mapped.status).toBe(401);
    expect(mapped.message).toContain("unauthorized");
  });

  it("maps generic errors to 500", () => {
    const mapped = toApiError(new Error("boom"));
    expect(mapped.status).toBe(500);
    expect(mapped.message).toBe("boom");
  });
});
