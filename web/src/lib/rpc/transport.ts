import { Code, ConnectError } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import type { Interceptor } from "@connectrpc/connect";

import { getValidAuthSnapshot, useAuthStore } from "@/store/auth";

const envBase = import.meta.env.VITE_API_BASE_URL?.trim();
const runtimeOrigin =
  typeof globalThis !== "undefined" &&
  "location" in globalThis &&
  typeof globalThis.location?.origin === "string" &&
  globalThis.location.origin
    ? globalThis.location.origin
    : "http://localhost";
const base = envBase ? envBase.replace(/\/+$/, "") : runtimeOrigin.replace(/\/+$/, "");

export const authInterceptor: Interceptor = (next) => async (req) => {
  const { token } = getValidAuthSnapshot();
  if (token) {
    req.header.set("Authorization", `Bearer ${token}`);
  }
  try {
    return await next(req);
  } catch (err) {
    if (err instanceof ConnectError && err.code === Code.Unauthenticated) {
      useAuthStore.getState().clearToken();
    }
    throw err;
  }
};

export const rpcTransport = createConnectTransport({
  baseUrl: `${base}/rpc`,
  interceptors: [authInterceptor],
});
