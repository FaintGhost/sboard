import { useEffect, useState } from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useTranslation } from "react-i18next"
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
import { ApiError } from "@/lib/api/client"
import { getSystemInfo, getSystemSettings, updateSystemSettings } from "@/lib/api/system"

const languages = [
  { code: "zh", nameKey: "settings.langZh" },
  { code: "en", nameKey: "settings.langEn" },
]

export function SettingsPage() {
  const { t, i18n } = useTranslation()
  const qc = useQueryClient()
  const apiBaseUrl = window.location.origin
  const [subscriptionBaseURL, setSubscriptionBaseURL] = useState("")
  const [settingsMessage, setSettingsMessage] = useState<string | null>(null)

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
    setSubscriptionBaseURL(systemSettingsQuery.data.subscription_base_url ?? "")
  }, [systemSettingsQuery.data])

  const updateSettingsMutation = useMutation({
    mutationFn: updateSystemSettings,
    onSuccess: async (data) => {
      setSubscriptionBaseURL(data.subscription_base_url ?? "")
      setSettingsMessage(t("settings.saveSuccess"))
      await qc.invalidateQueries({ queryKey: ["system-settings"] })
    },
  })

  const handleLanguageChange = (lang: string) => {
    i18n.changeLanguage(lang)
  }

  const resolvedSubscriptionBaseURL = subscriptionBaseURL.trim() || apiBaseUrl

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
            <CardTitle>{t("settings.systemInfo")}</CardTitle>
            <CardDescription>{t("settings.apiEndpoint")}</CardDescription>
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
            <div className="space-y-2">
              <div className="text-sm font-medium text-slate-700">{t("settings.apiEndpoint")}</div>
              <code className="block text-xs bg-slate-100 px-3 py-2 rounded font-mono">
                {apiBaseUrl}
              </code>
            </div>
          </CardContent>
        </Card>

        <Card className="md:col-span-2">
          <CardHeader>
            <CardTitle>{t("settings.subscriptionAccess")}</CardTitle>
            <CardDescription>{t("settings.subscriptionAccessHint")}</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="subscription-base-url">{t("settings.subscriptionBaseUrl")}</Label>
              <Input
                id="subscription-base-url"
                value={subscriptionBaseURL}
                onChange={(e) => {
                  setSubscriptionBaseURL(e.target.value)
                  setSettingsMessage(null)
                }}
                placeholder={t("settings.subscriptionBaseUrlPlaceholder")}
              />
              <p className="text-xs text-muted-foreground">{t("settings.subscriptionBaseUrlHelp")}</p>
            </div>

            <div className="space-y-2">
              <div className="text-sm font-medium text-slate-700">{t("settings.subscriptionBaseUrlPreview")}</div>
              <code className="block text-xs bg-slate-100 px-3 py-2 rounded font-mono">
                {resolvedSubscriptionBaseURL}
              </code>
            </div>

            {settingsMessage ? <p className="text-sm text-emerald-700">{settingsMessage}</p> : null}
            {updateSettingsMutation.error instanceof ApiError ? (
              <p className="text-sm text-destructive">{updateSettingsMutation.error.message}</p>
            ) : null}

            <div className="flex justify-end">
              <Button
                onClick={() =>
                  updateSettingsMutation.mutate({
                    subscription_base_url: subscriptionBaseURL,
                  })
                }
                disabled={updateSettingsMutation.isPending || systemSettingsQuery.isLoading}
              >
                {updateSettingsMutation.isPending ? t("common.saving") : t("common.save")}
              </Button>
            </div>
          </CardContent>
        </Card>

        <Card className="md:col-span-2">
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
