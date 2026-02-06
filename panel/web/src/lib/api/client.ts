import { useAuthStore } from "@/store/auth"

type SuccessEnvelope<T> = {
  data: T
}

type ErrorEnvelope = {
  error?: string
}

export class ApiError extends Error {
  status: number

  constructor(status: number, message: string) {
    super(message)
    this.name = "ApiError"
    this.status = status
  }
}

function buildURL(path: string): string {
  if (/^https?:\/\//.test(path)) {
    return path
  }

  const base = import.meta.env.VITE_API_BASE_URL?.trim() ?? ""
  if (base) {
    return new URL(path, base).toString()
  }

  if (typeof window !== "undefined") {
    return new URL(path, window.location.origin).toString()
  }

  return new URL(path, "http://localhost").toString()
}

export async function apiRequest<T>(path: string, init?: RequestInit): Promise<T> {
  const headers = new Headers(init?.headers)
  headers.set("Accept", "application/json")

  const token = useAuthStore.getState().token
  if (token) {
    headers.set("Authorization", `Bearer ${token}`)
  }

  const response = await fetch(
    new Request(buildURL(path), {
      ...init,
      headers,
    }),
  )

  const contentType = response.headers.get("Content-Type") ?? ""
  const isJSON = contentType.includes("application/json")
  const payload = isJSON ? ((await response.json()) as SuccessEnvelope<T> & ErrorEnvelope) : null

  if (!response.ok) {
    if (response.status === 401) {
      useAuthStore.getState().clearToken()
    }
    throw new ApiError(
      response.status,
      payload?.error ?? `request failed with status ${response.status}`,
    )
  }

  return payload?.data as T
}
