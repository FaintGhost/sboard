import { client } from "./gen/client.gen";
import { useAuthStore } from "@/store/auth";

// Configure base URL from environment if set.
const envBase = import.meta.env.VITE_API_BASE_URL?.trim();
if (envBase) {
  const base = envBase.replace(/\/+$/, "");
  client.setConfig({ baseUrl: `${base}/api` });
}

// Configure auth: provide JWT token for authenticated endpoints.
client.setConfig({
  auth: () => useAuthStore.getState().token ?? undefined,
});

// Error interceptor: always throws ApiError on HTTP errors.
// This preserves the same error-handling pattern as the old apiRequest().
client.interceptors.error.use((error, response) => {
  if (response?.status === 401) {
    useAuthStore.getState().clearToken();
  }
  let msg: string;
  if (typeof error === "object" && error !== null && "error" in error) {
    msg = (error as { error: string }).error;
  } else if (typeof error === "string") {
    msg = error;
  } else {
    msg = `request failed with status ${response?.status ?? 0}`;
  }
  throw new ApiError(response?.status ?? 0, msg);
});

export class ApiError extends Error {
  status: number;

  constructor(status: number, message: string) {
    super(message);
    this.name = "ApiError";
    this.status = status;
  }
}
