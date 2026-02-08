import { useQuery } from "@tanstack/react-query"
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
import { getSystemInfo } from "@/lib/api/system"

const languages = [
  { code: "zh", nameKey: "settings.langZh" },
  { code: "en", nameKey: "settings.langEn" },
]

export function SettingsPage() {
  const { t, i18n } = useTranslation()
  const apiBaseUrl = window.location.origin
  const systemInfoQuery = useQuery({
    queryKey: ["system-info"],
    queryFn: getSystemInfo,
  })

  const handleLanguageChange = (lang: string) => {
    i18n.changeLanguage(lang)
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
