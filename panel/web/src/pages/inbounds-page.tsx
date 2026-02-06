import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useMemo, useState } from "react"

import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
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
import { Textarea } from "@/components/ui/textarea"
import { ApiError } from "@/lib/api/client"
import { createInbound, deleteInbound, listInbounds, updateInbound } from "@/lib/api/inbounds"
import { listNodes } from "@/lib/api/nodes"
import type { Inbound, Node } from "@/lib/api/types"

type EditState = {
  mode: "create" | "edit"
  inbound: Inbound
  nodeID: number
  tag: string
  protocol: string
  listenPort: number
  publicPort: number
  settingsText: string
  tlsText: string
  transportText: string
}

const defaultNewInbound: Inbound = {
  id: 0,
  uuid: "",
  tag: "",
  node_id: 0,
  protocol: "vless",
  listen_port: 443,
  public_port: 0,
  settings: {},
  tls_settings: null,
  transport_settings: null,
}

const protocolOptions = ["vless", "vmess", "trojan", "shadowsocks"] as const

const shadowsocksMethods = [
  { value: "2022-blake3-aes-128-gcm", keyLength: 16 },
  { value: "2022-blake3-aes-256-gcm", keyLength: 32 },
  // 2022-blake3-chacha20-poly1305 is excluded: sing-box does not support it in multi-user mode.
  { value: "none", keyLength: null },
  { value: "aes-128-gcm", keyLength: null },
  { value: "aes-192-gcm", keyLength: null },
  { value: "aes-256-gcm", keyLength: null },
  { value: "chacha20-ietf-poly1305", keyLength: null },
  { value: "xchacha20-ietf-poly1305", keyLength: null },
] as const

function nodeName(nodes: Node[] | undefined, id: number): string {
  if (!nodes) return String(id)
  const n = nodes.find((x) => x.id === id)
  return n ? n.name : String(id)
}

function normalizeJSON(input: string): { ok: true; value: unknown } | { ok: false; message: string } {
  const raw = input.trim()
  if (!raw) return { ok: true, value: {} }
  try {
    return { ok: true, value: JSON.parse(raw) }
  } catch {
    return { ok: false, message: "JSON 解析失败" }
  }
}

function jsonHint(input: string): string | null {
  const out = normalizeJSON(input)
  return out.ok ? null : out.message
}

function getObjectFromJSONText(input: string): { ok: true; value: Record<string, unknown> } | { ok: false } {
  const out = normalizeJSON(input)
  if (!out.ok) return { ok: false }
  if (!out.value || typeof out.value !== "object" || Array.isArray(out.value)) return { ok: true, value: {} }
  return { ok: true, value: out.value as Record<string, unknown> }
}

function getShadowsocksMethod(settingsText: string): string | null {
  const obj = getObjectFromJSONText(settingsText)
  if (!obj.ok) return null
  const method = obj.value.method
  return typeof method === "string" ? method : null
}

function setShadowsocksMethod(settingsText: string, method: string): string {
  const obj = getObjectFromJSONText(settingsText)
  const next = obj.ok ? { ...obj.value, method } : { method }
  return JSON.stringify(next, null, 2)
}

export function InboundsPage() {
  const qc = useQueryClient()
  const [nodeFilter, setNodeFilter] = useState<number | "all">("all")
  const [upserting, setUpserting] = useState<EditState | null>(null)

  const queryParams = useMemo(() => ({ limit: 50, offset: 0 }), [])

  const nodesQuery = useQuery({
    queryKey: ["nodes", queryParams],
    queryFn: () => listNodes(queryParams),
  })

  const inboundsQuery = useQuery({
    queryKey: ["inbounds", nodeFilter],
    queryFn: () =>
      listInbounds({
        limit: 50,
        offset: 0,
        node_id: nodeFilter === "all" ? undefined : nodeFilter,
      }),
  })

  const createMutation = useMutation({
    mutationFn: createInbound,
    onSuccess: async () => {
      setUpserting(null)
      await qc.invalidateQueries({ queryKey: ["inbounds"] })
    },
  })

  const updateMutation = useMutation({
    mutationFn: (input: { id: number; payload: Record<string, unknown> }) =>
      updateInbound(input.id, input.payload),
    onSuccess: async () => {
      setUpserting(null)
      await qc.invalidateQueries({ queryKey: ["inbounds"] })
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: number) => deleteInbound(id),
    onSuccess: async () => {
      await qc.invalidateQueries({ queryKey: ["inbounds"] })
    },
  })

  return (
    <div className="px-4 lg:px-6">
      <section className="space-y-6">
        <header className="space-y-1">
          <h1 className="text-2xl font-semibold tracking-tight text-slate-900">
            入站管理
          </h1>
          <p className="text-sm text-slate-500">
            入站属于节点；同步节点时会把分组内的用户注入到每个入站的 `users` 字段。
          </p>
        </header>

        <div className="flex flex-col gap-3 md:flex-row md:items-end md:justify-between">
          <div className="space-y-1">
            <Label className="text-sm text-slate-700">节点筛选</Label>
            <Select
              value={nodeFilter === "all" ? "all" : String(nodeFilter)}
              onValueChange={(v) => setNodeFilter(v === "all" ? "all" : Number(v))}
            >
              <SelectTrigger className="w-64" aria-label="节点筛选">
                <SelectValue placeholder="选择节点" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">全部节点</SelectItem>
                {nodesQuery.data?.map((n) => (
                  <SelectItem key={n.id} value={String(n.id)}>
                    {n.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <Button
            onClick={() => {
              const firstNode = nodesQuery.data?.[0]
              if (!firstNode) return
              createMutation.reset()
              updateMutation.reset()
              setUpserting({
                mode: "create",
                inbound: defaultNewInbound,
                nodeID: firstNode.id,
                tag: "",
                protocol: "vless",
                listenPort: 443,
                publicPort: 0,
                settingsText: "{}",
                tlsText: "",
                transportText: "",
              })
            }}
            disabled={!nodesQuery.data || nodesQuery.data.length === 0}
          >
            创建入站
          </Button>
        </div>

        <div className="overflow-hidden rounded-xl border border-slate-200">
          <div className="border-b border-slate-200 bg-slate-50 px-4 py-2 text-sm text-slate-600">
            {inboundsQuery.isLoading ? "加载中..." : null}
            {inboundsQuery.isError ? "加载失败" : null}
            {inboundsQuery.data ? `共 ${inboundsQuery.data.length} 个入站` : null}
          </div>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="px-4">node</TableHead>
                <TableHead className="px-4">tag</TableHead>
                <TableHead className="px-4">protocol</TableHead>
                <TableHead className="px-4">port</TableHead>
                <TableHead className="px-4">action</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {inboundsQuery.data?.map((i) => (
                <TableRow key={i.id}>
                  <TableCell className="px-4 text-slate-900">
                    {nodeName(nodesQuery.data, i.node_id)}
                  </TableCell>
                  <TableCell className="px-4 font-medium text-slate-900">{i.tag}</TableCell>
                  <TableCell className="px-4 text-slate-700">{i.protocol}</TableCell>
                  <TableCell className="px-4 text-slate-700">
                    {i.public_port > 0 ? `${i.public_port} (public)` : i.listen_port}
                  </TableCell>
                  <TableCell className="px-4">
                    <div className="flex items-center gap-2">
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => {
                          createMutation.reset()
                          updateMutation.reset()
                          setUpserting({
                            mode: "edit",
                            inbound: i,
                            nodeID: i.node_id,
                            tag: i.tag,
                            protocol: i.protocol,
                            listenPort: i.listen_port,
                            publicPort: i.public_port ?? 0,
                            settingsText: JSON.stringify(i.settings ?? {}, null, 2),
                            tlsText: i.tls_settings ? JSON.stringify(i.tls_settings, null, 2) : "",
                            transportText: i.transport_settings
                              ? JSON.stringify(i.transport_settings, null, 2)
                              : "",
                          })
                        }}
                      >
                        编辑
                      </Button>
                      <Button
                        size="sm"
                        variant="destructive"
                        disabled={deleteMutation.isPending}
                        onClick={() => deleteMutation.mutate(i.id)}
                      >
                        删除
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
              {inboundsQuery.data && inboundsQuery.data.length === 0 ? (
                <TableRow>
                  <TableCell className="px-4 py-6 text-slate-500" colSpan={5}>
                    暂无数据
                  </TableCell>
                </TableRow>
              ) : null}
            </TableBody>
          </Table>
        </div>

        <Dialog open={!!upserting} onOpenChange={(open) => (!open ? setUpserting(null) : null)}>
          <DialogContent aria-label={upserting?.mode === "create" ? "创建入站" : "编辑入站"}>
            <DialogHeader>
              <DialogTitle>{upserting?.mode === "create" ? "创建入站" : "编辑入站"}</DialogTitle>
              {upserting?.mode === "edit" ? (
                <DialogDescription>{upserting.inbound.tag}</DialogDescription>
              ) : null}
            </DialogHeader>

            {upserting ? (
              <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                <div className="space-y-1 md:col-span-2">
                  <Label className="text-sm text-slate-700">节点（node_id）</Label>
                  <Select
                    value={String(upserting.nodeID)}
                    onValueChange={(v) =>
                      setUpserting((p) => (p ? { ...p, nodeID: Number(v) } : p))
                    }
                  >
                    <SelectTrigger aria-label="选择节点">
                      <SelectValue placeholder="选择节点" />
                    </SelectTrigger>
                    <SelectContent>
                      {nodesQuery.data?.map((n) => (
                        <SelectItem key={n.id} value={String(n.id)}>
                          {n.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>

                <div className="space-y-1">
                  <Label className="text-sm text-slate-700" htmlFor="inb-tag">
                    Tag（tag）
                  </Label>
                  <Input
                    id="inb-tag"
                    value={upserting.tag}
                    onChange={(e) => setUpserting((p) => (p ? { ...p, tag: e.target.value } : p))}
                    placeholder="例如 vless-in"
                    autoFocus={upserting.mode === "create"}
                  />
                </div>

                <div className="space-y-1">
                  <Label className="text-sm text-slate-700">协议（protocol）</Label>
                  <Select
                    value={upserting.protocol}
                    onValueChange={(v) =>
                      setUpserting((p) => (p ? { ...p, protocol: v } : p))
                    }
                  >
                    <SelectTrigger aria-label="选择协议">
                      <SelectValue placeholder="选择协议" />
                    </SelectTrigger>
                    <SelectContent>
                      {protocolOptions.map((p) => (
                        <SelectItem key={p} value={p}>
                          {p}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>

                {upserting.protocol === "shadowsocks" ? (
                  <div className="space-y-1 md:col-span-2">
                    <Label className="text-sm text-slate-700">加密方法（method）</Label>
                    <Select
                      value={getShadowsocksMethod(upserting.settingsText) ?? ""}
                      onValueChange={(v) =>
                        setUpserting((p) =>
                          p ? { ...p, settingsText: setShadowsocksMethod(p.settingsText, v) } : p,
                        )
                      }
                    >
                      <SelectTrigger aria-label="选择 shadowsocks method">
                        <SelectValue placeholder="请选择 method（必填）" />
                      </SelectTrigger>
                      <SelectContent>
                        {shadowsocksMethods.map((m) => (
                          <SelectItem key={m.value} value={m.value}>
                            {m.keyLength ? `${m.value} (${m.keyLength})` : m.value}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                    <p className="text-xs text-slate-500">
                      说明：括号内是 key length；没有标注则按 sing-box 默认规则处理。
                    </p>
                    {!getShadowsocksMethod(upserting.settingsText)?.trim() ? (
                      <p className="text-xs text-amber-700">method 必填，否则节点同步会失败。</p>
                    ) : null}
                  </div>
                ) : null}

                <div className="space-y-1">
                  <Label className="text-sm text-slate-700" htmlFor="inb-listen">
                    监听端口（listen_port）
                  </Label>
                  <Input
                    id="inb-listen"
                    type="number"
                    min={1}
                    value={upserting.listenPort}
                    onChange={(e) =>
                      setUpserting((p) =>
                        p ? { ...p, listenPort: Number(e.target.value || 0) } : p,
                      )
                    }
                  />
                </div>

                <div className="space-y-1">
                  <Label className="text-sm text-slate-700" htmlFor="inb-public">
                    对外端口（public_port）
                  </Label>
                  <Input
                    id="inb-public"
                    type="number"
                    min={0}
                    value={upserting.publicPort}
                    onChange={(e) =>
                      setUpserting((p) =>
                        p ? { ...p, publicPort: Number(e.target.value || 0) } : p,
                      )
                    }
                  />
                  <p className="text-xs text-slate-500">
                    0 表示不设置，对外订阅端口会回退到 listen_port。
                  </p>
                </div>

                <div className="space-y-1 md:col-span-2">
                  <Label className="text-sm text-slate-700" htmlFor="inb-settings">
                    settings（JSON）
                  </Label>
                  <Textarea
                    id="inb-settings"
                    value={upserting.settingsText}
                    onChange={(e) =>
                      setUpserting((p) => (p ? { ...p, settingsText: e.target.value } : p))
                    }
                    rows={8}
                    className="font-mono text-xs"
                  />
                  {jsonHint(upserting.settingsText) ? (
                    <p className="text-xs text-amber-700">{jsonHint(upserting.settingsText)}</p>
                  ) : null}
                </div>

                <div className="space-y-1 md:col-span-2">
                  <Label className="text-sm text-slate-700" htmlFor="inb-tls">
                    tls_settings（JSON，可选）
                  </Label>
                  <Textarea
                    id="inb-tls"
                    value={upserting.tlsText}
                    onChange={(e) =>
                      setUpserting((p) => (p ? { ...p, tlsText: e.target.value } : p))
                    }
                    rows={5}
                    className="font-mono text-xs"
                  />
                  {upserting.tlsText.trim() && jsonHint(upserting.tlsText) ? (
                    <p className="text-xs text-amber-700">{jsonHint(upserting.tlsText)}</p>
                  ) : null}
                </div>

                <div className="space-y-1 md:col-span-2">
                  <Label className="text-sm text-slate-700" htmlFor="inb-transport">
                    transport_settings（JSON，可选）
                  </Label>
                  <Textarea
                    id="inb-transport"
                    value={upserting.transportText}
                    onChange={(e) =>
                      setUpserting((p) => (p ? { ...p, transportText: e.target.value } : p))
                    }
                    rows={5}
                    className="font-mono text-xs"
                  />
                  {upserting.transportText.trim() && jsonHint(upserting.transportText) ? (
                    <p className="text-xs text-amber-700">{jsonHint(upserting.transportText)}</p>
                  ) : null}
                </div>

                <div className="text-sm text-amber-700 md:col-span-2">
                  {createMutation.isError || updateMutation.isError ? (
                    (createMutation.error instanceof ApiError
                      ? createMutation.error.message
                      : updateMutation.error instanceof ApiError
                        ? updateMutation.error.message
                        : "保存失败")
                  ) : null}
                </div>
              </div>
            ) : null}

            <DialogFooter>
              <Button
                variant="outline"
                onClick={() => setUpserting(null)}
                disabled={createMutation.isPending || updateMutation.isPending}
              >
                取消
              </Button>
              <Button
                onClick={() => {
                  if (!upserting) return
                  const tag = upserting.tag.trim()
                  const protocol = upserting.protocol.trim()
                  if (!tag || !protocol || upserting.nodeID <= 0 || upserting.listenPort <= 0) return

                  const settings = normalizeJSON(upserting.settingsText)
                  if (!settings.ok) return
                  if (protocol === "shadowsocks") {
                    const method = getShadowsocksMethod(upserting.settingsText)
                    if (!method || !method.trim()) return
                  }
                  const tls = normalizeJSON(upserting.tlsText)
                  if (!tls.ok) return
                  const transport = normalizeJSON(upserting.transportText)
                  if (!transport.ok) return

                  const payload = {
                    node_id: upserting.nodeID,
                    tag,
                    protocol,
                    listen_port: upserting.listenPort,
                    public_port: upserting.publicPort,
                    settings: settings.value,
                    tls_settings: upserting.tlsText.trim() ? tls.value : undefined,
                    transport_settings: upserting.transportText.trim() ? transport.value : undefined,
                  }

                  if (upserting.mode === "create") {
                    createMutation.mutate(payload)
                  } else {
                    updateMutation.mutate({ id: upserting.inbound.id, payload })
                  }
                }}
                disabled={createMutation.isPending || updateMutation.isPending}
              >
                保存
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </section>
    </div>
  )
}
