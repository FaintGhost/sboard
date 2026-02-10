import { create } from "zustand"

export const defaultSystemTimezone = "UTC"

export function normalizeTimezone(raw?: string | null): string {
  const value = (raw ?? "").trim()
  if (!value) return defaultSystemTimezone

  try {
    new Intl.DateTimeFormat(undefined, { timeZone: value })
    return value
  } catch {
    return defaultSystemTimezone
  }
}

type SystemState = {
  timezone: string
  setTimezone: (timezone: string) => void
}

export const useSystemStore = create<SystemState>((set) => ({
  timezone: defaultSystemTimezone,
  setTimezone: (timezone) => set({ timezone: normalizeTimezone(timezone) }),
}))
