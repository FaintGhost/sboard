import { apiRequest } from "./client"
import type {
  BootstrapRequest,
  BootstrapResponse,
  BootstrapStatus,
  LoginRequest,
  LoginResponse,
} from "./types"

export function loginAdmin(payload: LoginRequest) {
  return apiRequest<LoginResponse>("/api/admin/login", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(payload),
  })
}

export function getBootstrapStatus() {
  return apiRequest<BootstrapStatus>("/api/admin/bootstrap")
}

export function bootstrapAdmin(payload: BootstrapRequest) {
  return apiRequest<BootstrapResponse>("/api/admin/bootstrap", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(payload),
  })
}
