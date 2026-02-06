import { apiRequest } from "./client"
import type { LoginRequest, LoginResponse } from "./types"

export function loginAdmin(payload: LoginRequest) {
  return apiRequest<LoginResponse>("/api/admin/login", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(payload),
  })
}
