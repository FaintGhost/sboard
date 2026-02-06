const GB = 1024 * 1024 * 1024

export function bytesToGBString(bytes: number): string {
  if (!Number.isFinite(bytes) || bytes <= 0) return "0"
  const gb = bytes / GB
  // Keep it readable without trailing zeros.
  const s = gb.toFixed(2)
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

