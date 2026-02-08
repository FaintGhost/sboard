import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useMemo, useState } from "react"
import { useTranslation } from "react-i18next"
import { MoreHorizontal, Pencil, Trash2 } from "lucide-react"

import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
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
import { Skeleton } from "@/components/ui/skeleton"
import { ApiError } from "@/lib/api/client"
import { createInbound, deleteInbound, listInbounds, updateInbound } from "@/lib/api/inbounds"
import { listNodes } from "@/lib/api/nodes"
import type { Inbound, Node } from "@/lib/api/types"
import { tableColumnSpacing } from "@/lib/table-spacing"

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

function normalizeJSON(input: string): { ok: true; value: unknown } | { ok: false } {
  const raw = input.trim()
  if (!raw) return { ok: true, value: {} }
  try {
    return { ok: true, value: JSON.parse(raw) }
  } catch {
    return { ok: false }
  }
}

function jsonHint(t: (key: string) => string, input: string): string | null {
  const out = normalizeJSON(input)
  return out.ok ? null : t("inbounds.jsonParseFailed")
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
  const { t } = useTranslation()
  const qc = useQueryClient()
  const spacing = tableColumnSpacing.five
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
        <header className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
          <div>
            <h1 className="text-2xl font-semibold tracking-tight">{t("inbounds.title")}</h1>
            <p className="text-sm text-muted-foreground">{t("inbounds.subtitle")}</p>
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
            {t("inbounds.createInbound")}
          </Button>
        </header>

        <Card>
          <CardHeader className="pb-3">
            <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
              <div className="flex flex-col gap-1.5">
                <CardTitle className="text-base">{t("inbounds.list")}</CardTitle>
                <CardDescription>
                  {inboundsQuery.isLoading ? t("common.loading") : null}
                  {inboundsQuery.isError ? t("common.loadFailed") : null}
                  {inboundsQuery.data ? t("inbounds.count", { count: inboundsQuery.data.length }) : null}
                </CardDescription>
              </div>
              <div className="flex flex-col gap-2 sm:flex-row sm:items-center">
                <Select
                  value={nodeFilter === "all" ? "all" : String(nodeFilter)}
                  onValueChange={(v) => setNodeFilter(v === "all" ? "all" : Number(v))}
                >
                  <SelectTrigger className="w-full sm:w-56" aria-label={t("inbounds.nodeFilter")}>
                    <SelectValue placeholder={t("inbounds.selectNode")} />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">{t("inbounds.allNodes")}</SelectItem>
                    {nodesQuery.data?.map((n) => (
                      <SelectItem key={n.id} value={String(n.id)}>
                        {n.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>
          </CardHeader>
          <CardContent className="p-0">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className={spacing.headFirst}>{t("inbounds.node")}</TableHead>
                  <TableHead className={spacing.headMiddle}>{t("inbounds.tag")}</TableHead>
                  <TableHead className={spacing.headMiddle}>{t("inbounds.protocol")}</TableHead>
                  <TableHead className={spacing.headMiddle}>{t("inbounds.port")}</TableHead>
                  <TableHead className={`w-12 ${spacing.headLast}`}>
                    <span className="sr-only">{t("common.actions")}</span>
                  </TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {inboundsQuery.isLoading ? (
                  <>
                    {Array.from({ length: 5 }).map((_, i) => (
                      <TableRow key={i}>
                        <TableCell className={spacing.cellFirst}>
                          <Skeleton className="h-4 w-28" />
                        </TableCell>
                        <TableCell className={spacing.cellMiddle}>
                          <Skeleton className="h-4 w-28" />
                        </TableCell>
                        <TableCell className={spacing.cellMiddle}>
                          <Skeleton className="h-4 w-24" />
                        </TableCell>
                        <TableCell className={spacing.cellMiddle}>
                          <Skeleton className="h-4 w-20" />
                        </TableCell>
                        <TableCell className={spacing.cellLast}>
                          <Skeleton className="h-8 w-8" />
                        </TableCell>
                      </TableRow>
                    ))}
                  </>
                ) : null}
                {inboundsQuery.data?.map((i) => (
                  <TableRow key={i.id}>
                    <TableCell className={`${spacing.cellFirst} font-medium`}>
                      {nodeName(nodesQuery.data, i.node_id)}
                    </TableCell>
                    <TableCell className={`${spacing.cellMiddle} font-medium`}>{i.tag}</TableCell>
                    <TableCell className={`${spacing.cellMiddle} text-muted-foreground`}>{i.protocol}</TableCell>
                    <TableCell className={`${spacing.cellMiddle} text-muted-foreground`}>
                      {i.public_port > 0
                        ? `${i.public_port} (${t("inbounds.publicPortShort")})`
                        : i.listen_port}
                    </TableCell>
                    <TableCell className={spacing.cellLast}>
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" size="icon" className="size-8">
                            <MoreHorizontal className="size-4" />
                            <span className="sr-only">{t("common.actions")}</span>
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem
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
                            <Pencil className="mr-2 size-4" />
                            {t("common.edit")}
                          </DropdownMenuItem>
                          <DropdownMenuSeparator />
                          <DropdownMenuItem
                            variant="destructive"
                            disabled={deleteMutation.isPending}
                            onClick={() => deleteMutation.mutate(i.id)}
                          >
                            <Trash2 className="mr-2 size-4" />
                            {t("common.delete")}
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </TableCell>
                  </TableRow>
                ))}
                {!inboundsQuery.isLoading && inboundsQuery.data && inboundsQuery.data.length === 0 ? (
                  <TableRow>
                    <TableCell className={`${spacing.cellFirst} py-8 text-center text-muted-foreground`} colSpan={5}>
                      {t("common.noData")}
                    </TableCell>
                  </TableRow>
                ) : null}
              </TableBody>
            </Table>
          </CardContent>
        </Card>

        <Dialog open={!!upserting} onOpenChange={(open) => (!open ? setUpserting(null) : null)}>
          <DialogContent
            aria-label={
              upserting?.mode === "create"
                ? t("inbounds.createInbound")
                : t("inbounds.editInbound")
            }
          >
            <DialogHeader>
              <DialogTitle>
                {upserting?.mode === "create"
                  ? t("inbounds.createInbound")
                  : t("inbounds.editInbound")}
              </DialogTitle>
              {upserting?.mode === "edit" ? (
                <DialogDescription>{upserting.inbound.tag}</DialogDescription>
              ) : null}
            </DialogHeader>

            {upserting ? (
              <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                <div className="space-y-1 md:col-span-2">
                  <Label className="text-sm text-slate-700">{t("inbounds.node")}</Label>
                  <Select
                    value={String(upserting.nodeID)}
                    onValueChange={(v) =>
                      setUpserting((p) => (p ? { ...p, nodeID: Number(v) } : p))
                    }
                  >
                    <SelectTrigger aria-label={t("inbounds.selectNode")}>
                      <SelectValue placeholder={t("inbounds.selectNode")} />
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
                    {t("inbounds.tag")}
                  </Label>
                  <Input
                    id="inb-tag"
                    value={upserting.tag}
                    onChange={(e) => setUpserting((p) => (p ? { ...p, tag: e.target.value } : p))}
                    placeholder={t("inbounds.tagPlaceholder")}
                    autoFocus={upserting.mode === "create"}
                  />
                </div>

                <div className="space-y-1">
                  <Label className="text-sm text-slate-700">{t("inbounds.protocol")}</Label>
                  <Select
                    value={upserting.protocol}
                    onValueChange={(v) =>
                      setUpserting((p) => (p ? { ...p, protocol: v } : p))
                    }
                  >
                    <SelectTrigger aria-label={t("inbounds.selectProtocol")}>
                      <SelectValue placeholder={t("inbounds.selectProtocol")} />
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
                    <Label className="text-sm text-slate-700">{t("inbounds.ssMethod")}</Label>
                    <Select
                      value={getShadowsocksMethod(upserting.settingsText) ?? ""}
                      onValueChange={(v) =>
                        setUpserting((p) =>
                          p ? { ...p, settingsText: setShadowsocksMethod(p.settingsText, v) } : p,
                        )
                      }
                    >
                      <SelectTrigger aria-label={t("inbounds.ssMethod")}>
                        <SelectValue placeholder={t("inbounds.ssMethodPlaceholder")} />
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
                      {t("inbounds.ssMethodHelp")}
                    </p>
                    {!getShadowsocksMethod(upserting.settingsText)?.trim() ? (
                      <p className="text-xs text-amber-700">{t("inbounds.ssMethodRequiredHint")}</p>
                    ) : null}
                  </div>
                ) : null}

                <div className="space-y-1">
                  <Label className="text-sm text-slate-700" htmlFor="inb-listen">
                    {t("inbounds.listenPort")}
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
                    {t("inbounds.publicPort")}
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
                    {t("inbounds.publicPortHelp")}
                  </p>
                </div>

                <div className="space-y-1 md:col-span-2">
                  <Label className="text-sm text-slate-700" htmlFor="inb-settings">
                    {t("inbounds.settings")}
                  </Label>
                  <Textarea
                    id="inb-settings"
                    value={upserting.settingsText}
                    onChange={(e) =>
                      setUpserting((p) => (p ? { ...p, settingsText: e.target.value } : p))
                    }
                    rows={8}
                    className="font-mono text-xs"
                    placeholder={t("inbounds.settingsPlaceholder")}
                  />
                  {jsonHint(t, upserting.settingsText) ? (
                    <p className="text-xs text-amber-700">{jsonHint(t, upserting.settingsText)}</p>
                  ) : null}
                </div>

                <div className="space-y-1 md:col-span-2">
                  <Label className="text-sm text-slate-700" htmlFor="inb-tls">
                    {t("inbounds.tlsSettings")}
                  </Label>
                  <Textarea
                    id="inb-tls"
                    value={upserting.tlsText}
                    onChange={(e) =>
                      setUpserting((p) => (p ? { ...p, tlsText: e.target.value } : p))
                    }
                    rows={5}
                    className="font-mono text-xs"
                    placeholder={t("inbounds.tlsSettingsPlaceholder")}
                  />
                  {upserting.tlsText.trim() && jsonHint(t, upserting.tlsText) ? (
                    <p className="text-xs text-amber-700">{jsonHint(t, upserting.tlsText)}</p>
                  ) : null}
                </div>

                <div className="space-y-1 md:col-span-2">
                  <Label className="text-sm text-slate-700" htmlFor="inb-transport">
                    {t("inbounds.transportSettings")}
                  </Label>
                  <Textarea
                    id="inb-transport"
                    value={upserting.transportText}
                    onChange={(e) =>
                      setUpserting((p) => (p ? { ...p, transportText: e.target.value } : p))
                    }
                    rows={5}
                    className="font-mono text-xs"
                    placeholder={t("inbounds.transportSettingsPlaceholder")}
                  />
                  {upserting.transportText.trim() && jsonHint(t, upserting.transportText) ? (
                    <p className="text-xs text-amber-700">{jsonHint(t, upserting.transportText)}</p>
                  ) : null}
                </div>

                <div className="text-sm text-amber-700 md:col-span-2">
                  {createMutation.isError || updateMutation.isError ? (
                    (createMutation.error instanceof ApiError
                      ? createMutation.error.message
                      : updateMutation.error instanceof ApiError
                        ? updateMutation.error.message
                        : t("inbounds.saveFailed"))
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
                {t("common.cancel")}
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
                {t("common.save")}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </section>
    </div>
  )
}
