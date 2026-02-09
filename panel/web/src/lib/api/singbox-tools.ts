import { apiRequest } from "./client"
import type {
  SingBoxCheckResponse,
  SingBoxFormatResponse,
  SingBoxGenerateCommand,
  SingBoxGenerateResponse,
  SingBoxToolMode,
} from "./types"

export function formatSingBoxConfig(payload: { config: string; mode?: SingBoxToolMode }) {
  return apiRequest<SingBoxFormatResponse>("/api/sing-box/format", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  })
}

export function checkSingBoxConfig(payload: { config: string; mode?: SingBoxToolMode }) {
  return apiRequest<SingBoxCheckResponse>("/api/sing-box/check", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  })
}

export function generateSingBoxValue(command: SingBoxGenerateCommand) {
  return apiRequest<SingBoxGenerateResponse>("/api/sing-box/generate", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ command }),
  })
}
