const KB = 1024
const MB = 1024 * 1024
const GB = 1024 * 1024 * 1024
const TB = 1024 * 1024 * 1024 * 1024

type ByteUnit = "B" | "KB" | "MB" | "GB" | "TB"

export function bytesToGBString(bytes: number): string {
  if (!Number.isFinite(bytes) || bytes <= 0) return "0"
  const gb = bytes / GB
  const s = gb.toFixed(2)
  return s.replace(/\.00$/, "").replace(/(\.\d)0$/, "$1")
}

export function pickByteUnit(maxBytes: number): ByteUnit {
  if (!Number.isFinite(maxBytes) || maxBytes <= 0) return "B"
  if (maxBytes >= TB) return "TB"
  if (maxBytes >= GB) return "GB"
  if (maxBytes >= MB) return "MB"
  if (maxBytes >= KB) return "KB"
  return "B"
}

function divisorByUnit(unit: ByteUnit): number {
  if (unit === "TB") return TB
  if (unit === "GB") return GB
  if (unit === "MB") return MB
  if (unit === "KB") return KB
  return 1
}

export function formatBytesWithUnit(bytes: number, unit: ByteUnit, fractionDigits = 2): string {
  if (!Number.isFinite(bytes) || bytes <= 0) return "0"
  const value = bytes / divisorByUnit(unit)
  const s = value.toFixed(fractionDigits)
  return s.replace(/\.00$/, "").replace(/(\.\d)0$/, "$1")
}

export function gbStringToBytes(value: string): number | null {
  const v = value.trim()
  if (!v) return null
  const gb = Number(v)
  if (!Number.isFinite(gb) || gb < 0) return null
  return Math.round(gb * GB)
}

export function rfc3339FromDateOnlyUTC(d: Date): string {
  const year = d.getFullYear()
  const month = d.getMonth()
  const day = d.getDate()
  const utc = new Date(Date.UTC(year, month, day, 0, 0, 0))
  return utc.toISOString()
}
