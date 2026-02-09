import { useEffect, useMemo, useRef, useState } from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useTranslation } from "react-i18next"
import { Check, Copy } from "lucide-react"
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
import { Button } from "@/components/ui/button"
import { AsyncButton } from "@/components/ui/async-button"
import { ApiError } from "@/lib/api/client"
import { getSystemInfo, getSystemSettings, updateSystemSettings } from "@/lib/api/system"

type SubscriptionScheme = "http" | "https"
type HostPortValidationCode = "format" | "ip" | "port" | null
type UpdateSystemSettingsPayload = Parameters<typeof updateSystemSettings>[0]

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

  const [subscriptionScheme, setSubscriptionScheme] = useState<SubscriptionScheme>("http")
  const [subscriptionHostPort, setSubscriptionHostPort] = useState("")
  const [settingsMessage, setSettingsMessage] = useState<string | null>(null)
  const [validationError, setValidationError] = useState<string | null>(null)
  const [copied, setCopied] = useState(false)

  const systemInfoQuery = useQuery({
    queryKey: ["system-info"],
    queryFn: getSystemInfo,
  })

  const systemSettingsQuery = useQuery({
    queryKey: ["system-settings"],
    queryFn: getSystemSettings,
  })

  useEffect(() => {
    if (!systemSettingsQuery.data) return
    const parsed = parseConfiguredSubscriptionBaseURL(
      systemSettingsQuery.data.subscription_base_url ?? "",
      apiBaseUrl,
    )
    setSubscriptionScheme(parsed.scheme)
    setSubscriptionHostPort(parsed.hostPort)
  }, [systemSettingsQuery.data, apiBaseUrl])

  const updateSettingsMutation = useMutation({
    mutationFn: (payload: UpdateSystemSettingsPayload) => updateSystemSettings(payload),
    onSuccess: async (data) => {
      const parsed = parseConfiguredSubscriptionBaseURL(
        data.subscription_base_url ?? "",
        apiBaseUrl,
      )
      setSubscriptionScheme(parsed.scheme)
      setSubscriptionHostPort(parsed.hostPort)
      setValidationError(null)
      setSettingsMessage(t("settings.saveSuccess"))
      await qc.invalidateQueries({ queryKey: ["system-settings"] })
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
      setSettingsMessage(t("common.copiedToClipboard"))
    } catch {
      setSettingsMessage(t("common.copyFailed"))
    }
  }

  function handleSaveSubscriptionAccess() {
    setSettingsMessage(null)
    setValidationError(null)
    setCopied(false)

    const hostPort = subscriptionHostPort.trim()
    if (!hostPort) {
      updateSettingsMutation.mutate({ subscription_base_url: "" })
      return
    }

    const validationMessage = resolveValidationError(hostPort)
    if (validationMessage) {
      setValidationError(validationMessage)
      hostPortInputRef.current?.focus()
      return
    }

    updateSettingsMutation.mutate({
      subscription_base_url: `${subscriptionScheme}://${hostPort}`,
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
        <Card>
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
                      setSettingsMessage(null)
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
                  <Label htmlFor="subscription-host-port">{t("settings.subscriptionAddress")}</Label>
                  <Input
                    id="subscription-host-port"
                    ref={hostPortInputRef}
                    value={subscriptionHostPort}
                    onChange={(e) => {
                      setSubscriptionHostPort(e.target.value)
                      setValidationError(null)
                      setSettingsMessage(null)
                    }}
                    onBlur={handleHostPortBlur}
                    placeholder={t("settings.subscriptionAddressPlaceholder")}
                    aria-invalid={!!validationError}
                    aria-describedby="subscription-host-port-help"
                    autoComplete="off"
                  />
                </div>
              </div>
              <p id="subscription-host-port-help" className="text-xs text-muted-foreground">
                {t("settings.subscriptionAddressHelp")}
              </p>
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
              <code className="block break-all text-xs bg-slate-100 px-3 py-2 rounded font-mono">
                {resolvedSubscriptionBaseURL}
              </code>
            </div>

            <div className="min-h-5 space-y-1" aria-live="polite">
              {settingsMessage ? <p className="text-sm text-emerald-700">{settingsMessage}</p> : null}
              {validationError ? <p role="alert" className="text-sm text-destructive">{validationError}</p> : null}
              {updateSettingsMutation.error instanceof ApiError ? (
                <p role="alert" className="text-sm text-destructive">{updateSettingsMutation.error.message}</p>
              ) : null}
            </div>

            <div className="flex justify-end">
              <AsyncButton
                type="button"
                onClick={handleSaveSubscriptionAccess}
                disabled={updateSettingsMutation.isPending || systemSettingsQuery.isLoading}
                pending={updateSettingsMutation.isPending}
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
            <CardTitle>{t("settings.language")}</CardTitle>
            <CardDescription>{t("settings.selectLanguage")}</CardDescription>
          </CardHeader>
          <CardContent>
            <Select value={i18n.language} onValueChange={handleLanguageChange}>
              <SelectTrigger className="w-full">
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
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>{t("settings.subscriptionFormats")}</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid gap-4 sm:grid-cols-2">
              <div className="space-y-2 p-4 bg-slate-50 rounded-lg">
                <div className="font-medium">{t("subscriptions.formatSingbox")}</div>
                <p className="text-sm text-slate-600">
                  {t("settings.singboxFormatHint")}
                </p>
              </div>
              <div className="space-y-2 p-4 bg-slate-50 rounded-lg">
                <div className="font-medium">{t("subscriptions.formatV2ray")} (Base64)</div>
                <p className="text-sm text-slate-600">
                  {t("settings.v2rayFormatHint")}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
