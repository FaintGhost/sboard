import { apiRequest } from "./client"
import type {
  AdminProfile,
  SystemInfo,
  SystemSettings,
  UpdateAdminProfilePayload,
} from "./types"

export function getSystemInfo() {
  return apiRequest<SystemInfo>("/api/system/info")
}

export function getSystemSettings() {
  return apiRequest<SystemSettings>("/api/system/settings")
}

export function updateSystemSettings(payload: SystemSettings) {
  return apiRequest<SystemSettings>("/api/system/settings", {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  })
}


export function getAdminProfile() {
  return apiRequest<AdminProfile>("/api/admin/profile")
}

export function updateAdminProfile(payload: UpdateAdminProfilePayload) {
  return apiRequest<AdminProfile>("/api/admin/profile", {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  })
}
