import { useEffect, useMemo, useRef, useState } from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useTranslation } from "react-i18next"
import { Check, Copy } from "lucide-react"
import { getTimeZones } from "@vvo/tzdb"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { FieldHint } from "@/components/ui/field-hint"
import { Button } from "@/components/ui/button"
import { AsyncButton } from "@/components/ui/async-button"
import { ApiError } from "@/lib/api/client"
import {
  getAdminProfile,
  getSystemInfo,
  getSystemSettings,
  updateAdminProfile,
  updateSystemSettings,
} from "@/lib/api/system"
import { useSystemStore } from "@/store/system"

type SubscriptionScheme = "http" | "https"
type HostPortValidationCode = "format" | "ip" | "port" | null
type UpdateSystemSettingsPayload = Parameters<typeof updateSystemSettings>[0]
type UpdateAdminProfilePayload = Parameters<typeof updateAdminProfile>[0]
type TimezoneOption = {
  name: string
  label: string
  searchText: string
}

const languages = [
  { code: "zh", nameKey: "settings.langZh" },
  { code: "en", nameKey: "settings.langEn" },
]

function isValidIPv4(host: string): boolean {
  const parts = host.split(".")
  if (parts.length !== 4) return false
  return parts.every((part) => {
    if (!/^\d+$/.test(part)) return false
    const num = Number(part)
    return Number.isInteger(num) && num >= 0 && num <= 255
  })
}

function isValidIPv6(host: string): boolean {
  if (!host.includes(":")) return false
  if (!/^[0-9a-fA-F:]+$/.test(host)) return false
  if (host.includes(":::")) return false
  return true
}

function splitHostPort(value: string): { host: string; port: string } | null {
  const raw = value.trim()
  if (!raw) return null

  if (raw.startsWith("[")) {
    const end = raw.indexOf("]")
    if (end <= 0) return null
    const host = raw.slice(1, end).trim()
    const tail = raw.slice(end + 1)
    if (!tail.startsWith(":")) return null
    const port = tail.slice(1).trim()
    if (!host || !port) return null
    return { host, port }
  }

  const idx = raw.lastIndexOf(":")
  if (idx <= 0) return null
  const host = raw.slice(0, idx).trim()
  const port = raw.slice(idx + 1).trim()
  if (!host || !port) return null
  if (host.includes(":")) return null
  return { host, port }
}

function normalizeHostPort(host: string, port: number): string {
  if (host.includes(":")) {
    return `[${host}]:${port}`
  }
  return `${host}:${port}`
}

function validateHostPort(value: string): HostPortValidationCode {
  const raw = value.trim()
  if (!raw) return null

  const parts = splitHostPort(raw)
  if (!parts) return "format"

  if (!isValidIPv4(parts.host) && !isValidIPv6(parts.host)) {
    return "ip"
  }

  const port = Number(parts.port)
  if (!Number.isInteger(port) || port < 1 || port > 65535) {
    return "port"
  }

  return null
}

function defaultSubscriptionParts(apiBaseURL: string): { scheme: SubscriptionScheme; hostPort: string } {
  try {
    const parsed = new URL(apiBaseURL)
    const scheme = parsed.protocol === "https:" ? "https" : "http"
    return {
      scheme,
      hostPort: parsed.host,
    }
  } catch {
    return {
      scheme: "http",
      hostPort: "",
    }
  }
}

function parseConfiguredSubscriptionBaseURL(configured: string, apiBaseURL: string): {
  scheme: SubscriptionScheme
  hostPort: string
} {
  const fallback = defaultSubscriptionParts(apiBaseURL)
  const value = configured.trim()
  if (!value) return fallback

  try {
    const parsed = new URL(value)
    const scheme = parsed.protocol === "https:" ? "https" : "http"
    const host = parsed.hostname.trim()
    const port = parsed.port.trim()
    if (!host || !port) return fallback
    return {
      scheme,
      hostPort: normalizeHostPort(host, Number(port)),
    }
  } catch {
    return fallback
  }
}

export function SettingsPage() {
  const { t, i18n } = useTranslation()
  const qc = useQueryClient()
  const apiBaseUrl = window.location.origin
  const hostPortInputRef = useRef<HTMLInputElement>(null)
  const timezoneInputRef = useRef<HTMLInputElement>(null)
  const setGlobalTimezone = useSystemStore((state) => state.setTimezone)

  const [subscriptionScheme, setSubscriptionScheme] = useState<SubscriptionScheme>("http")
  const [subscriptionHostPort, setSubscriptionHostPort] = useState("")
  const [subscriptionMessage, setSubscriptionMessage] = useState<string | null>(null)
  const [generalMessage, setGeneralMessage] = useState<string | null>(null)
  const [validationError, setValidationError] = useState<string | null>(null)
  const [timezoneValidationError, setTimezoneValidationError] = useState<string | null>(null)
  const [copied, setCopied] = useState(false)
  const [timezone, setTimezone] = useState("UTC")
  const [timezoneTouched, setTimezoneTouched] = useState(false)

  const [adminUsername, setAdminUsername] = useState("")
  const [adminOldPassword, setAdminOldPassword] = useState("")
  const [adminNewPassword, setAdminNewPassword] = useState("")
  const [adminConfirmPassword, setAdminConfirmPassword] = useState("")
  const [adminMessage, setAdminMessage] = useState<string | null>(null)
  const [adminValidationError, setAdminValidationError] = useState<string | null>(null)

  const timezoneOptions = useMemo<TimezoneOption[]>(() => {
    return getTimeZones({ includeUtc: true }).map((item) => {
      const offset = item.currentTimeFormat.startsWith("GMT")
        ? item.currentTimeFormat.replace("GMT", "UTC")
        : item.currentTimeFormat
      const searchableMainCities = item.mainCities.join(" ")

      return {
        name: item.name,
        label: `(${offset}) ${item.name}`,
        searchText: `${item.name} ${item.alternativeName} ${item.countryName} ${searchableMainCities}`.toLowerCase(),
      }
    })
  }, [])

  const timezoneNameMap = useMemo(() => {
    const map = new Map<string, string>()
    timezoneOptions.forEach((item) => {
      map.set(item.name.toLowerCase(), item.name)
    })
    return map
  }, [timezoneOptions])

  const filteredTimezoneOptions = useMemo(() => {
    const query = timezone.trim().toLowerCase()
    if (!query) return timezoneOptions
    return timezoneOptions.filter((item) => item.searchText.includes(query))
  }, [timezone, timezoneOptions])

  const systemInfoQuery = useQuery({
    queryKey: ["system-info"],
    queryFn: getSystemInfo,
  })

  const systemSettingsQuery = useQuery({
    queryKey: ["system-settings"],
    queryFn: getSystemSettings,
  })

  const adminProfileQuery = useQuery({
    queryKey: ["admin-profile"],
    queryFn: getAdminProfile,
  })

  useEffect(() => {
    if (!systemSettingsQuery.data) return
    const parsed = parseConfiguredSubscriptionBaseURL(
      systemSettingsQuery.data.subscription_base_url ?? "",
      apiBaseUrl,
    )
    setSubscriptionScheme(parsed.scheme)
    setSubscriptionHostPort(parsed.hostPort)
    if (!timezoneTouched) {
      setTimezone(systemSettingsQuery.data.timezone ?? "UTC")
    }
    setGlobalTimezone(systemSettingsQuery.data.timezone ?? "UTC")
  }, [systemSettingsQuery.data, apiBaseUrl, setGlobalTimezone, timezoneTouched])

  useEffect(() => {
    if (!adminProfileQuery.data) return
    setAdminUsername((prev) => (prev.trim() ? prev : adminProfileQuery.data.username))
  }, [adminProfileQuery.data])

  async function onSystemSettingsSaved(data: { subscription_base_url: string; timezone: string }) {
    const parsed = parseConfiguredSubscriptionBaseURL(
      data.subscription_base_url ?? "",
      apiBaseUrl,
    )
    setSubscriptionScheme(parsed.scheme)
    setSubscriptionHostPort(parsed.hostPort)
    setTimezone(data.timezone ?? "UTC")
    setTimezoneTouched(false)
    setGlobalTimezone(data.timezone ?? "UTC")
    setValidationError(null)
    setTimezoneValidationError(null)
    await qc.invalidateQueries({ queryKey: ["system-settings"] })
  }

  const updateSubscriptionMutation = useMutation({
    mutationFn: (payload: UpdateSystemSettingsPayload) => updateSystemSettings(payload),
    onSuccess: async (data) => {
      await onSystemSettingsSaved(data)
      setSubscriptionMessage(t("settings.saveSuccess"))
      setGeneralMessage(null)
    },
  })

  const updateGeneralMutation = useMutation({
    mutationFn: (payload: UpdateSystemSettingsPayload) => updateSystemSettings(payload),
    onSuccess: async (data) => {
      await onSystemSettingsSaved(data)
      setGeneralMessage(t("settings.saveSuccess"))
      setSubscriptionMessage(null)
    },
  })

  const updateAdminProfileMutation = useMutation({
    mutationFn: (payload: UpdateAdminProfilePayload) => updateAdminProfile(payload),
    onSuccess: async (data) => {
      setAdminUsername(data.username)
      setAdminOldPassword("")
      setAdminNewPassword("")
      setAdminConfirmPassword("")
      setAdminValidationError(null)
      setAdminMessage(t("settings.adminUpdateSuccess"))
      await qc.invalidateQueries({ queryKey: ["admin-profile"] })
    },
  })

  useEffect(() => {
    if (!copied) return
    const timer = window.setTimeout(() => {
      setCopied(false)
    }, 1600)
    return () => {
      window.clearTimeout(timer)
    }
  }, [copied])

  const handleLanguageChange = (lang: string) => {
    i18n.changeLanguage(lang)
    setGeneralMessage(null)
  }

  const resolveTimezoneForSave = (): string | null => {
    const raw = timezone.trim()
    if (!raw) {
      setTimezone("UTC")
      setTimezoneValidationError(null)
      return "UTC"
    }

    const canonical = timezoneNameMap.get(raw.toLowerCase()) ?? raw
    try {
      new Intl.DateTimeFormat(undefined, { timeZone: canonical })
      setTimezone(canonical)
      setTimezoneValidationError(null)
      return canonical
    } catch {
      setTimezoneValidationError(t("settings.timezoneInvalid"))
      timezoneInputRef.current?.focus()
      return null
    }
  }

  const resolvedSubscriptionBaseURL = useMemo(() => {
    const hostPort = subscriptionHostPort.trim()
    if (!hostPort) return apiBaseUrl
    return `${subscriptionScheme}://${hostPort}`
  }, [apiBaseUrl, subscriptionHostPort, subscriptionScheme])

  const copySupported = typeof navigator !== "undefined" && !!navigator.clipboard?.writeText

  const resolveValidationError = (hostPort: string): string | null => {
    const validateCode = validateHostPort(hostPort)
    if (validateCode === "format") return t("settings.subscriptionAddressInvalidFormat")
    if (validateCode === "ip") return t("settings.subscriptionAddressInvalidIP")
    if (validateCode === "port") return t("settings.subscriptionAddressInvalidPort")
    return null
  }

  const handleHostPortBlur = () => {
    const hostPort = subscriptionHostPort.trim()
    if (!hostPort) {
      setValidationError(null)
      return
    }
    setValidationError(resolveValidationError(hostPort))
  }

  const handleCopyBaseURL = async () => {
    if (!copySupported) return
    try {
      await navigator.clipboard.writeText(resolvedSubscriptionBaseURL)
      setCopied(true)
      setSubscriptionMessage(t("common.copiedToClipboard"))
    } catch {
      setSubscriptionMessage(t("common.copyFailed"))
    }
  }

  function handleSaveSubscriptionAccess() {
    setSubscriptionMessage(null)
    setGeneralMessage(null)
    setValidationError(null)
    setTimezoneValidationError(null)
    setCopied(false)

    const timezoneValue = resolveTimezoneForSave()
    if (!timezoneValue) return

    const hostPort = subscriptionHostPort.trim()
    if (!hostPort) {
      updateSubscriptionMutation.mutate({
        subscription_base_url: "",
        timezone: timezoneValue,
      })
      return
    }

    const validationMessage = resolveValidationError(hostPort)
    if (validationMessage) {
      setValidationError(validationMessage)
      hostPortInputRef.current?.focus()
      return
    }

    updateSubscriptionMutation.mutate({
      subscription_base_url: `${subscriptionScheme}://${hostPort}`,
      timezone: timezoneValue,
    })
  }

  function handleSaveGeneralSettings() {
    setGeneralMessage(null)
    setSubscriptionMessage(null)
    setTimezoneValidationError(null)

    const timezoneValue = resolveTimezoneForSave()
    if (!timezoneValue) return

    const hostPort = subscriptionHostPort.trim()
    if (hostPort) {
      const subscriptionValidationMessage = resolveValidationError(hostPort)
      if (subscriptionValidationMessage) {
        setValidationError(subscriptionValidationMessage)
        hostPortInputRef.current?.focus()
        return
      }
    }

    updateGeneralMutation.mutate({
      subscription_base_url: hostPort ? `${subscriptionScheme}://${hostPort}` : "",
      timezone: timezoneValue,
    })
  }

  function handleSaveAdminProfile() {
    setAdminMessage(null)
    setAdminValidationError(null)

    const newUsername = adminUsername.trim()
    if (!newUsername) {
      setAdminValidationError(t("settings.adminUsernameRequired"))
      return
    }

    if (!adminOldPassword.trim()) {
      setAdminValidationError(t("settings.adminOldPasswordRequired"))
      return
    }

    const wantsPasswordChange = adminNewPassword.trim() !== "" || adminConfirmPassword.trim() !== ""

    if (wantsPasswordChange) {
      if (adminNewPassword !== adminConfirmPassword) {
        setAdminValidationError(t("settings.adminPasswordsNotMatch"))
        return
      }
      if (adminNewPassword.length < 8) {
        setAdminValidationError(t("settings.adminPasswordTooShort"))
        return
      }
    }

    if (!wantsPasswordChange && adminProfileQuery.data && newUsername === adminProfileQuery.data.username) {
      setAdminValidationError(t("settings.adminNoChanges"))
      return
    }

    updateAdminProfileMutation.mutate({
      new_username: newUsername,
      old_password: adminOldPassword,
      new_password: adminNewPassword,
      confirm_password: adminConfirmPassword,
    })
  }

  return (
    <div className="px-4 lg:px-6 space-y-6">
      <div className="space-y-2">
        <h1 className="text-xl font-semibold text-slate-900">{t("settings.title")}</h1>
        <p className="text-sm text-slate-500">
          {t("settings.subtitle")}
        </p>
      </div>

      <div className="grid gap-6 md:grid-cols-2">
        <Card className="border-border/80 bg-card shadow-[0_1px_0_0_rgba(255,255,255,0.34)_inset,0_16px_34px_-30px_rgba(0,0,0,0.5)] dark:shadow-[0_1px_0_0_rgba(255,255,255,0.08)_inset,0_20px_40px_-34px_rgba(0,0,0,0.9)]">
          <CardHeader>
            <CardTitle>{t("settings.subscriptionAccess")}</CardTitle>
            <CardDescription>{t("settings.subscriptionAccessHint")}</CardDescription>
          </CardHeader>
          <CardContent className="space-y-5">
            <div className="space-y-2">
              <div className="grid gap-3 md:grid-cols-[120px_minmax(0,1fr)] md:items-end">
                <div className="space-y-2">
                  <Label htmlFor="subscription-scheme">{t("settings.subscriptionProtocol")}</Label>
                  <Select
                    value={subscriptionScheme}
                    onValueChange={(value) => {
                      setSubscriptionScheme(value as SubscriptionScheme)
                      setValidationError(null)
                      setSubscriptionMessage(null)
                    }}
                  >
                    <SelectTrigger
                      id="subscription-scheme"
                      className="w-full"
                      aria-label={t("settings.subscriptionProtocol")}
                    >
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="http">{t("settings.subscriptionSchemeHttp")}</SelectItem>
                      <SelectItem value="https">{t("settings.subscriptionSchemeHttps")}</SelectItem>
                    </SelectContent>
                  </Select>
                </div>

                <div className="min-w-0 flex-1 space-y-2">
                  <div className="flex items-center gap-1">
                    <Label htmlFor="subscription-host-port">{t("settings.subscriptionAddress")}</Label>
                    <FieldHint label={t("settings.subscriptionAddress")}>{t("settings.subscriptionAddressHelp")}</FieldHint>
                  </div>
                  <Input
                    id="subscription-host-port"
                    ref={hostPortInputRef}
                    value={subscriptionHostPort}
                    onChange={(e) => {
                      setSubscriptionHostPort(e.target.value)
                      setValidationError(null)
                      setSubscriptionMessage(null)
                    }}
                    onBlur={handleHostPortBlur}
                    placeholder={t("settings.subscriptionAddressPlaceholder")}
                    aria-invalid={!!validationError}
                    autoComplete="off"
                  />
                </div>
              </div>
            </div>

            <div className="space-y-2">
              <div className="flex items-center justify-between gap-3">
                <div className="text-sm font-medium text-slate-700">{t("settings.subscriptionBaseUrlPreview")}</div>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  className="h-8 shrink-0"
                  onClick={handleCopyBaseURL}
                  disabled={!copySupported}
                >
                  {copied ? <Check className="mr-1 h-3.5 w-3.5" /> : <Copy className="mr-1 h-3.5 w-3.5" />}
                  {copied ? t("common.copied") : t("common.copy")}
                </Button>
              </div>
              <code className="block break-all rounded-md border border-border/75 bg-background/65 px-3 py-2 text-xs font-mono">
                {resolvedSubscriptionBaseURL}
              </code>
            </div>

            <div className="min-h-5 space-y-1" aria-live="polite">
              {subscriptionMessage ? <p className="text-sm text-emerald-700">{subscriptionMessage}</p> : null}
              {validationError ? <p role="alert" className="text-sm text-destructive">{validationError}</p> : null}
              {updateSubscriptionMutation.error instanceof ApiError ? (
                <p role="alert" className="text-sm text-destructive">{updateSubscriptionMutation.error.message}</p>
              ) : null}
            </div>

            <div className="flex justify-end">
              <AsyncButton
                type="button"
                onClick={handleSaveSubscriptionAccess}
                disabled={updateSubscriptionMutation.isPending || systemSettingsQuery.isLoading}
                pending={updateSubscriptionMutation.isPending}
                pendingText={t("common.saving")}
              >
                {t("common.save")}
              </AsyncButton>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>{t("settings.systemInfo")}</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <span className="text-sm text-slate-700">{t("settings.version")}</span>
              <Badge variant="outline">{systemInfoQuery.data?.panel_version ?? "N/A"}</Badge>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm text-slate-700">{t("settings.panelCommitId")}</span>
              <Badge variant="outline">{systemInfoQuery.data?.panel_commit_id ?? "N/A"}</Badge>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm text-slate-700">{t("settings.singboxCoreVersion")}</span>
              <Badge variant="outline">{systemInfoQuery.data?.sing_box_version ?? "N/A"}</Badge>
            </div>
            <div className="flex items-center justify-between gap-3">
              <span className="text-sm text-slate-700">{t("settings.subscriptionBaseUrlLabel")}</span>
              <code className="rounded bg-slate-100 px-2 py-1 text-xs font-mono">
                {resolvedSubscriptionBaseURL}
              </code>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>{t("settings.adminCredentials")}</CardTitle>
            <CardDescription>{t("settings.adminCredentialsHint")}</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="admin-username">{t("settings.adminUsername")}</Label>
              <Input
                id="admin-username"
                autoComplete="username"
                value={adminUsername}
                onChange={(e) => {
                  setAdminUsername(e.target.value)
                  setAdminMessage(null)
                  setAdminValidationError(null)
                }}
                placeholder={t("settings.adminUsernamePlaceholder")}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="admin-old-password">{t("settings.adminOldPassword")}</Label>
              <Input
                id="admin-old-password"
                type="password"
                autoComplete="current-password"
                value={adminOldPassword}
                onChange={(e) => {
                  setAdminOldPassword(e.target.value)
                  setAdminMessage(null)
                  setAdminValidationError(null)
                }}
                placeholder={t("settings.adminOldPasswordPlaceholder")}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="admin-new-password">{t("settings.adminNewPassword")}</Label>
              <Input
                id="admin-new-password"
                type="password"
                autoComplete="new-password"
                value={adminNewPassword}
                onChange={(e) => {
                  setAdminNewPassword(e.target.value)
                  setAdminMessage(null)
                  setAdminValidationError(null)
                }}
                placeholder={t("settings.adminNewPasswordPlaceholder")}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="admin-confirm-password">{t("settings.adminConfirmPassword")}</Label>
              <Input
                id="admin-confirm-password"
                type="password"
                autoComplete="new-password"
                value={adminConfirmPassword}
                onChange={(e) => {
                  setAdminConfirmPassword(e.target.value)
                  setAdminMessage(null)
                  setAdminValidationError(null)
                }}
                placeholder={t("settings.adminConfirmPasswordPlaceholder")}
              />
            </div>

            <div className="min-h-5 space-y-1" aria-live="polite">
              {adminMessage ? <p className="text-sm text-emerald-700">{adminMessage}</p> : null}
              {adminValidationError ? <p role="alert" className="text-sm text-destructive">{adminValidationError}</p> : null}
              {updateAdminProfileMutation.error instanceof ApiError ? (
                <p role="alert" className="text-sm text-destructive">{updateAdminProfileMutation.error.message}</p>
              ) : null}
            </div>

            <div className="flex justify-end">
              <AsyncButton
                type="button"
                onClick={handleSaveAdminProfile}
                disabled={updateAdminProfileMutation.isPending || adminProfileQuery.isLoading}
                pending={updateAdminProfileMutation.isPending}
                pendingText={t("common.saving")}
              >
                {t("common.save")}
              </AsyncButton>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>{t("settings.generalSettings")}</CardTitle>
            <CardDescription>{t("settings.generalSettingsHint")}</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="settings-language">{t("settings.language")}</Label>
              <Select value={i18n.language} onValueChange={handleLanguageChange}>
                <SelectTrigger id="settings-language" className="w-full">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {languages.map((lang) => (
                    <SelectItem key={lang.code} value={lang.code}>
                      {t(lang.nameKey)}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-2">
              <div className="flex items-center gap-1">
                <Label htmlFor="system-timezone">{t("settings.timezone")}</Label>
                <FieldHint label={t("settings.timezone")}>{t("settings.timezoneHelp")}</FieldHint>
              </div>
              <Input
                id="system-timezone"
                ref={timezoneInputRef}
                list="system-timezone-options"
                value={timezone}
                onChange={(e) => {
                  setTimezone(e.target.value)
                  setTimezoneTouched(true)
                  setGeneralMessage(null)
                  setTimezoneValidationError(null)
                }}
                onBlur={() => {
                  if (!timezone.trim()) {
                    setTimezone("UTC")
                    setTimezoneValidationError(null)
                    return
                  }
                  const matched = timezoneNameMap.get(timezone.trim().toLowerCase())
                  if (matched) {
                    setTimezone(matched)
                  }
                }}
                placeholder={t("settings.timezonePlaceholder")}
                autoComplete="off"
                aria-invalid={!!timezoneValidationError}
              />
              <datalist id="system-timezone-options">
                {filteredTimezoneOptions.map((item) => (
                  <option key={item.name} value={item.name} label={item.label} />
                ))}
              </datalist>
            </div>

            <div className="min-h-5 space-y-1" aria-live="polite">
              {generalMessage ? <p className="text-sm text-emerald-700">{generalMessage}</p> : null}
              {timezoneValidationError ? <p role="alert" className="text-sm text-destructive">{timezoneValidationError}</p> : null}
              {updateGeneralMutation.error instanceof ApiError ? (
                <p role="alert" className="text-sm text-destructive">{updateGeneralMutation.error.message}</p>
              ) : null}
            </div>

            <div className="flex justify-end">
              <AsyncButton
                type="button"
                onClick={handleSaveGeneralSettings}
                disabled={updateGeneralMutation.isPending || systemSettingsQuery.isLoading}
                pending={updateGeneralMutation.isPending}
                pendingText={t("common.saving")}
              >
                {t("common.save")}
              </AsyncButton>
            </div>
          </CardContent>
        </Card>

      </div>
    </div>
  )
}
