import { apiRequest } from "./client"
import type { SystemInfo, SystemSettings } from "./types"

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
