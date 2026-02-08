import { apiRequest } from "./client"
import type { SystemInfo } from "./types"

export function getSystemInfo() {
  return apiRequest<SystemInfo>("/api/system/info")
}

