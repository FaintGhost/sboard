import { useQuery } from "@tanstack/react-query"
import { useState, useMemo } from "react"
import { toast } from "sonner"
import { Copy, Check, ExternalLink, Info } from "lucide-react"

import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip"
import { Badge } from "@/components/ui/badge"
import { listUsers } from "@/lib/api/users"
import type { User, UserStatus } from "@/lib/api/types"

type StatusFilter = UserStatus | "all"

const statusOptions: Array<{ value: StatusFilter; label: string }> = [
  { value: "all", label: "全部" },
  { value: "active", label: "active" },
  { value: "disabled", label: "disabled" },
  { value: "expired", label: "expired" },
  { value: "traffic_exceeded", label: "traffic_exceeded" },
]

function getSubscriptionUrl(userUuid: string, format?: string): string {
  const base = `${window.location.origin}/api/sub/${userUuid}`
  return format ? `${base}?format=${format}` : base
}

function CopyButton({ text, label }: { text: string; label?: string }) {
  const [copied, setCopied] = useState(false)

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(text)
      setCopied(true)
      toast.success(label ? `${label} 已复制` : "已复制到剪贴板")
      setTimeout(() => setCopied(false), 2000)
    } catch {
      toast.error("复制失败")
    }
  }

  return (
    <Button
      variant="ghost"
      size="icon"
      className="h-7 w-7 shrink-0"
      onClick={handleCopy}
    >
      {copied ? (
        <Check className="h-4 w-4 text-green-600" />
      ) : (
        <Copy className="h-4 w-4" />
      )}
    </Button>
  )
}

function StatusBadge({ status }: { status: UserStatus }) {
  const variants: Record<UserStatus, "default" | "destructive" | "secondary" | "outline"> = {
    active: "default",
    disabled: "secondary",
    expired: "destructive",
    traffic_exceeded: "outline",
  }

  return <Badge variant={variants[status]}>{status}</Badge>
}

export function SubscriptionsPage() {
  const [statusFilter, setStatusFilter] = useState<StatusFilter>("active")
  const [search, setSearch] = useState("")

  const queryParams = useMemo(
    () => ({
      limit: 100,
      offset: 0,
      status: statusFilter === "all" ? undefined : statusFilter,
    }),
    [statusFilter],
  )

  const usersQuery = useQuery({
    queryKey: ["users", queryParams],
    queryFn: () => listUsers(queryParams),
  })

  const filteredUsers = useMemo(() => {
    if (!usersQuery.data) return []
    if (!search.trim()) return usersQuery.data
    const lowerSearch = search.toLowerCase()
    return usersQuery.data.filter(
      (u) =>
        u.username.toLowerCase().includes(lowerSearch) ||
        u.uuid.toLowerCase().includes(lowerSearch),
    )
  }, [usersQuery.data, search])

  return (
    <div className="px-4 lg:px-6 space-y-6">
      <div className="space-y-2">
        <h1 className="text-xl font-semibold text-slate-900">订阅管理</h1>
        <p className="text-sm text-slate-500">
          查看和复制用户订阅链接，分享给客户端使用。
        </p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            订阅行为说明
            <Tooltip>
              <TooltipTrigger>
                <Info className="h-4 w-4 text-slate-400" />
              </TooltipTrigger>
              <TooltipContent className="max-w-xs">
                <p>订阅链接会根据客户端 User-Agent 自动返回对应格式</p>
              </TooltipContent>
            </Tooltip>
          </CardTitle>
          <CardDescription>
            订阅由后端生成，根据请求参数和客户端 UA 返回不同格式
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-3 text-sm">
          <div className="grid gap-2">
            <div className="flex items-start gap-2">
              <Badge variant="outline" className="shrink-0 mt-0.5">
                ?format=singbox
              </Badge>
              <span className="text-slate-600">
                显式指定格式，返回 sing-box JSON 配置
              </span>
            </div>
            <div className="flex items-start gap-2">
              <Badge variant="outline" className="shrink-0 mt-0.5">
                ?format=v2ray
              </Badge>
              <span className="text-slate-600">
                返回 Base64 编码的订阅链接（兼容 V2rayN/Shadowrocket）
              </span>
            </div>
            <div className="flex items-start gap-2">
              <Badge variant="secondary" className="shrink-0 mt-0.5">
                UA: sing-box/SFA/SFI
              </Badge>
              <span className="text-slate-600">
                客户端 User-Agent 匹配时，自动返回 JSON 格式
              </span>
            </div>
            <div className="flex items-start gap-2">
              <Badge variant="secondary" className="shrink-0 mt-0.5">
                其他 UA
              </Badge>
              <span className="text-slate-600">
                默认返回 Base64 编码的 JSON（兼容性最好）
              </span>
            </div>
          </div>
        </CardContent>
      </Card>

      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex items-center gap-2">
          <Input
            placeholder="搜索用户名或 UUID..."
            className="w-64"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
          <Select
            value={statusFilter}
            onValueChange={(v) => setStatusFilter(v as StatusFilter)}
          >
            <SelectTrigger className="w-40">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {statusOptions.map((opt) => (
                <SelectItem key={opt.value} value={opt.value}>
                  {opt.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </div>

      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-[180px]">用户名</TableHead>
              <TableHead className="w-[100px]">状态</TableHead>
              <TableHead>订阅链接</TableHead>
              <TableHead className="w-[140px] text-right">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {usersQuery.isLoading ? (
              <TableRow>
                <TableCell colSpan={4} className="text-center text-slate-500">
                  加载中...
                </TableCell>
              </TableRow>
            ) : filteredUsers.length === 0 ? (
              <TableRow>
                <TableCell colSpan={4} className="text-center text-slate-500">
                  暂无用户
                </TableCell>
              </TableRow>
            ) : (
              filteredUsers.map((user) => (
                <UserSubscriptionRow key={user.id} user={user} />
              ))
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  )
}

function UserSubscriptionRow({ user }: { user: User }) {
  const subUrl = getSubscriptionUrl(user.uuid)
  const singboxUrl = getSubscriptionUrl(user.uuid, "singbox")

  return (
    <TableRow>
      <TableCell className="font-medium">{user.username}</TableCell>
      <TableCell>
        <StatusBadge status={user.status} />
      </TableCell>
      <TableCell>
        <div className="flex items-center gap-2 max-w-md">
          <code className="flex-1 truncate text-xs bg-slate-100 px-2 py-1 rounded font-mono">
            {subUrl}
          </code>
          <CopyButton text={subUrl} label="订阅链接" />
        </div>
      </TableCell>
      <TableCell className="text-right">
        <div className="flex items-center justify-end gap-1">
          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                variant="ghost"
                size="icon"
                className="h-7 w-7"
                onClick={() => window.open(singboxUrl, "_blank")}
              >
                <ExternalLink className="h-4 w-4" />
              </Button>
            </TooltipTrigger>
            <TooltipContent>在新窗口预览 (sing-box 格式)</TooltipContent>
          </Tooltip>
          <CopyButton text={singboxUrl} label="sing-box 链接" />
        </div>
      </TableCell>
    </TableRow>
  )
}
