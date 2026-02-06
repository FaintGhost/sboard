import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"

export function SettingsPage() {
  const apiBaseUrl = window.location.origin

  return (
    <div className="px-4 lg:px-6 space-y-6">
      <div className="space-y-2">
        <h1 className="text-xl font-semibold text-slate-900">系统设置</h1>
        <p className="text-sm text-slate-500">
          查看系统信息和配置（更多设置项将在后续版本中添加）。
        </p>
      </div>

      <div className="grid gap-6 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>API 信息</CardTitle>
            <CardDescription>当前系统的 API 端点信息</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <div className="text-sm font-medium text-slate-700">API 基础地址</div>
              <code className="block text-xs bg-slate-100 px-3 py-2 rounded font-mono">
                {apiBaseUrl}
              </code>
            </div>
            <div className="space-y-2">
              <div className="text-sm font-medium text-slate-700">订阅端点</div>
              <code className="block text-xs bg-slate-100 px-3 py-2 rounded font-mono">
                {apiBaseUrl}/api/sub/{"<user_uuid>"}
              </code>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>系统状态</CardTitle>
            <CardDescription>当前系统运行状态</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <span className="text-sm text-slate-700">前端版本</span>
              <Badge variant="outline">v0.1.0</Badge>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm text-slate-700">环境</span>
              <Badge variant="secondary">
                {import.meta.env.DEV ? "开发" : "生产"}
              </Badge>
            </div>
          </CardContent>
        </Card>

        <Card className="md:col-span-2">
          <CardHeader>
            <CardTitle>订阅格式说明</CardTitle>
            <CardDescription>支持的订阅格式及其用途</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid gap-4 sm:grid-cols-2">
              <div className="space-y-2 p-4 bg-slate-50 rounded-lg">
                <div className="font-medium">sing-box</div>
                <p className="text-sm text-slate-600">
                  原生 JSON 配置格式，适用于 sing-box、SFA (Android)、SFI (iOS) 客户端
                </p>
              </div>
              <div className="space-y-2 p-4 bg-slate-50 rounded-lg">
                <div className="font-medium">v2ray (Base64)</div>
                <p className="text-sm text-slate-600">
                  Base64 编码的订阅链接，兼容 V2RayN、Shadowrocket 等客户端
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
