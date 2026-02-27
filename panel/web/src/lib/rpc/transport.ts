import { createConnectTransport } from "@connectrpc/connect-web";
import type { Interceptor } from "@connectrpc/connect";

import { useAuthStore } from "@/store/auth";

const envBase = import.meta.env.VITE_API_BASE_URL?.trim();
const runtimeOrigin =
  typeof globalThis !== "undefined" &&
  "location" in globalThis &&
  typeof globalThis.location?.origin === "string" &&
  globalThis.location.origin
    ? globalThis.location.origin
    : "http://localhost";
const base = envBase ? envBase.replace(/\/+$/, "") : runtimeOrigin.replace(/\/+$/, "");

const authInterceptor: Interceptor = (next) => async (req) => {
  const token = useAuthStore.getState().token;
  if (token) {
    req.header.set("Authorization", `Bearer ${token}`);
  }
  try {
    return await next(req);
  } catch (err) {
    const code =
      typeof err === "object" && err && "code" in err
        ? String((err as { code: unknown }).code)
        : "";
    if (code === "unauthenticated") {
      useAuthStore.getState().clearToken();
    }
    throw err;
  }
};

export const rpcTransport = createConnectTransport({
  baseUrl: `${base}/rpc`,
  interceptors: [authInterceptor],
});
