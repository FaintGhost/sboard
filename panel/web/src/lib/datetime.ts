import { defaultSystemTimezone, normalizeTimezone } from "@/store/system"

type DateValue = string | number | Date | null | undefined

function parseDate(value: DateValue): Date | null {
  if (value == null) return null
  const date = value instanceof Date ? value : new Date(value)
  if (Number.isNaN(date.getTime())) return null
  return date
}

export function formatDateTimeByTimezone(
  value: DateValue,
  locale: string,
  timezone?: string | null,
  fallback = "-",
): string {
  const date = parseDate(value)
  if (!date) return fallback

  const tz = normalizeTimezone(timezone ?? defaultSystemTimezone)
  try {
    return new Intl.DateTimeFormat(locale, {
      timeZone: tz,
      year: "numeric",
      month: "2-digit",
      day: "2-digit",
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
      hour12: false,
    }).format(date)
  } catch {
    return fallback
  }
}

export function formatDateYMDByTimezone(
  value: DateValue,
  timezone?: string | null,
  fallback = "-",
): string {
  const date = parseDate(value)
  if (!date) return fallback

  const tz = normalizeTimezone(timezone ?? defaultSystemTimezone)
  try {
    const parts = new Intl.DateTimeFormat("en-US", {
      timeZone: tz,
      year: "numeric",
      month: "2-digit",
      day: "2-digit",
    }).formatToParts(date)

    const year = parts.find((part) => part.type === "year")?.value
    const month = parts.find((part) => part.type === "month")?.value
    const day = parts.find((part) => part.type === "day")?.value
    if (!year || !month || !day) return fallback
    return `${year}-${month}-${day}`
  } catch {
    return fallback
  }
}
