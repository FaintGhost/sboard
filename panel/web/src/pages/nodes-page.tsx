import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useMemo, useState } from "react"
import { useTranslation } from "react-i18next"
import { MoreHorizontal, Pencil, RefreshCw, Stethoscope, Trash2 } from "lucide-react"

import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Separator } from "@/components/ui/separator"
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
import { Skeleton } from "@/components/ui/skeleton"
import { StatusDot } from "@/components/status-dot"
import { ApiError } from "@/lib/api/client"
import { listGroups } from "@/lib/api/groups"
import { createNode, deleteNode, listNodeTraffic, listNodes, nodeHealth, nodeSync, updateNode } from "@/lib/api/nodes"
import type { Group, Node, NodeTrafficSample } from "@/lib/api/types"
import { listTrafficNodesSummary, type TrafficNodeSummary } from "@/lib/api/traffic"
import { buildNodeDockerCompose, generateNodeSecretKey } from "@/lib/node-compose"
import { bytesToGBString } from "@/lib/units"

type EditState = {
  mode: "create" | "edit"
  node: Node
  name: string
  apiAddress: string
  apiPort: number
  secretKey: string
  publicAddress: string
  groupID: number | null
}

const defaultNewNode: Node = {
  id: 0,
  uuid: "",
  name: "",
  api_address: "",
  api_port: 3000,
  secret_key: "",
  public_address: "",
  group_id: null,
  status: "offline",
}

function groupName(groups: Group[] | undefined, id: number | null): string {
  if (!groups || id == null) return "-"
  const g = groups.find((x) => x.id === id)
  return g ? g.name : String(id)
}

export function NodesPage() {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const [upserting, setUpserting] = useState<EditState | null>(null)
  const [trafficNode, setTrafficNode] = useState<Node | null>(null)
  const [actionMessage, setActionMessage] = useState<string | null>(null)

  const queryParams = useMemo(() => ({ limit: 50, offset: 0 }), [])
  const nodesQuery = useQuery({
    queryKey: ["nodes", queryParams],
    queryFn: () => listNodes(queryParams),
    refetchInterval: 5_000,
  })

  const groupsQuery = useQuery({
    queryKey: ["groups", queryParams],
    queryFn: () => listGroups(queryParams),
  })

  const trafficSummary24hQuery = useQuery({
    queryKey: ["traffic", "nodes", "summary", "24h"],
    queryFn: () => listTrafficNodesSummary({ window: "24h" }),
    refetchInterval: 30_000,
  })

  const trafficSummary1hQuery = useQuery({
    queryKey: ["traffic", "nodes", "summary", "1h"],
    queryFn: () => listTrafficNodesSummary({ window: "1h" }),
    refetchInterval: 30_000,
  })

  const trafficSummaryByNodeID = useMemo(() => {
    const map24 = new Map<number, TrafficNodeSummary>()
    const map1 = new Map<number, TrafficNodeSummary>()
    for (const it of trafficSummary24hQuery.data ?? []) map24.set(it.node_id, it)
    for (const it of trafficSummary1hQuery.data ?? []) map1.set(it.node_id, it)
    return { map24, map1 }
  }, [trafficSummary24hQuery.data, trafficSummary1hQuery.data])

  const createMutation = useMutation({
    mutationFn: createNode,
    onSuccess: async () => {
      setUpserting(null)
      await qc.invalidateQueries({ queryKey: ["nodes"] })
    },
  })

  const updateMutation = useMutation({
    mutationFn: (input: { id: number; payload: Record<string, unknown> }) =>
      updateNode(input.id, input.payload),
    onSuccess: async () => {
      setUpserting(null)
      await qc.invalidateQueries({ queryKey: ["nodes"] })
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: number) => deleteNode(id),
    onSuccess: async () => {
      await qc.invalidateQueries({ queryKey: ["nodes"] })
    },
  })

  const healthMutation = useMutation({
    mutationFn: (id: number) => nodeHealth(id),
    onSuccess: () => setActionMessage(t("nodes.healthOk")),
    onError: (e) => setActionMessage(e instanceof ApiError ? e.message : t("nodes.healthFailed")),
  })

  const syncMutation = useMutation({
    mutationFn: (id: number) => nodeSync(id),
    onSuccess: () => setActionMessage(t("nodes.syncOk")),
    onError: (e) => setActionMessage(e instanceof ApiError ? e.message : t("nodes.syncFailed")),
  })

  const trafficQuery = useQuery({
    queryKey: ["nodes", "traffic", trafficNode?.id ?? 0],
    queryFn: async () => {
      if (!trafficNode) return [] as NodeTrafficSample[]
      return listNodeTraffic(trafficNode.id, { limit: 300, offset: 0 })
    },
    enabled: !!trafficNode,
    refetchInterval: trafficNode ? 10_000 : false,
  })

  const trafficByInbound = useMemo(() => {
    const rows = trafficQuery.data ?? []
    const map = new Map<string, { inbound: string; upload: number; download: number; last: string }>()
    for (const r of rows) {
      const tag = r.inbound_tag ?? "(node)"
      const prev = map.get(tag)
      const last = prev ? (prev.last > r.recorded_at ? prev.last : r.recorded_at) : r.recorded_at
      map.set(tag, {
        inbound: tag,
        upload: (prev?.upload ?? 0) + r.upload,
        download: (prev?.download ?? 0) + r.download,
        last,
      })
    }
    return Array.from(map.values()).sort((a, b) => (a.inbound < b.inbound ? -1 : 1))
  }, [trafficQuery.data])

  return (
    <div className="px-4 lg:px-6">
      <section className="space-y-6">
        <header className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
          <div>
            <h1 className="text-2xl font-semibold tracking-tight">{t("nodes.title")}</h1>
            <p className="text-sm text-muted-foreground">{t("nodes.subtitle")}</p>
          </div>
          <Button
            onClick={() => {
              setActionMessage(null)
              createMutation.reset()
              updateMutation.reset()
              setUpserting({
                mode: "create",
                node: defaultNewNode,
                name: "",
                apiAddress: "127.0.0.1",
                apiPort: 3000,
                secretKey: "",
                publicAddress: "",
                groupID: null,
              })
            }}
          >
            {t("nodes.createNode")}
          </Button>
        </header>

        <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
          {nodesQuery.isLoading ? (
            <>
              {Array.from({ length: 6 }).map((_, i) => (
                <Card key={i} className="shadow-xs">
                  <CardHeader className="space-y-2">
                    <Skeleton className="h-5 w-40" />
                    <Skeleton className="h-4 w-56" />
                  </CardHeader>
                  <CardContent className="space-y-2">
                    <Skeleton className="h-4 w-52" />
                    <Skeleton className="h-4 w-52" />
                    <Skeleton className="h-4 w-44" />
                  </CardContent>
                </Card>
              ))}
            </>
          ) : null}

          {nodesQuery.data?.map((n) => {
            const s24 = trafficSummaryByNodeID.map24.get(n.id)
            const s1 = trafficSummaryByNodeID.map1.get(n.id)
            const last = s24?.last_recorded_at || s1?.last_recorded_at || n.last_seen_at || ""
            const up24 = s24?.upload ?? 0
            const down24 = s24?.download ?? 0
            const up1 = s1?.upload ?? 0
            const down1 = s1?.download ?? 0

            return (
              <Card
                key={n.id}
                className="bg-gradient-to-t from-primary/5 to-card shadow-xs"
              >
                <CardHeader className="pb-3">
                  <div className="flex items-start justify-between gap-3">
                    <div className="min-w-0">
                      <CardTitle className="truncate text-base">{n.name}</CardTitle>
                      <CardDescription className="truncate">
                        {groupName(groupsQuery.data, n.group_id)} · {n.api_address}:{n.api_port}
                      </CardDescription>
                    </div>
                    <StatusDot
                      status={n.status}
                      labelOnline={t("nodes.statusOnline")}
                      labelOffline={t("nodes.statusOffline")}
                      labelUnknown={t("nodes.statusUnknown")}
                      className="shrink-0"
                    />
                  </div>
                </CardHeader>
                <CardContent className="space-y-2 text-sm">
                  <div className="flex items-center justify-between gap-3">
                    <span className="text-muted-foreground">{t("traffic.window1h")}</span>
                    <span className="tabular-nums">
                      ↑ {bytesToGBString(up1)} GB{"  "}↓ {bytesToGBString(down1)} GB
                    </span>
                  </div>
                  <div className="flex items-center justify-between gap-3">
                    <span className="text-muted-foreground">{t("traffic.window24h")}</span>
                    <span className="tabular-nums">
                      ↑ {bytesToGBString(up24)} GB{"  "}↓ {bytesToGBString(down24)} GB
                    </span>
                  </div>
                  <div className="flex items-center justify-between gap-3">
                    <span className="text-muted-foreground">{t("nodes.lastSampleAt")}</span>
                    <span className="truncate tabular-nums">{last || "-"}</span>
                  </div>
                  <div className="flex justify-end gap-2 pt-2">
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      disabled={healthMutation.isPending}
                      onClick={() => healthMutation.mutate(n.id)}
                    >
                      {t("nodes.health")}
                    </Button>
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      disabled={syncMutation.isPending}
                      onClick={() => syncMutation.mutate(n.id)}
                    >
                      {t("nodes.sync")}
                    </Button>
                    <Button
                      type="button"
                      size="sm"
                      onClick={() => {
                        setActionMessage(null)
                        setTrafficNode(n)
                      }}
                    >
                      {t("nodes.traffic")}
                    </Button>
                  </div>
                </CardContent>
              </Card>
            )
          })}

          {!nodesQuery.isLoading && nodesQuery.data && nodesQuery.data.length === 0 ? (
            <Card className="shadow-xs md:col-span-2 xl:col-span-3">
              <CardContent className="py-10 text-center text-sm text-muted-foreground">
                {t("common.noData")}
              </CardContent>
            </Card>
          ) : null}
        </div>

        <Card>
          <CardHeader className="pb-3">
            <div className="flex flex-col gap-1.5">
              <CardTitle className="text-base">{t("nodes.list")}</CardTitle>
              <CardDescription>
                {nodesQuery.isLoading ? t("common.loading") : null}
                {nodesQuery.isError ? t("common.loadFailed") : null}
                {nodesQuery.data ? t("nodes.count", { count: nodesQuery.data.length }) : null}
                {actionMessage ? <span className="ml-3">{actionMessage}</span> : null}
              </CardDescription>
            </div>
          </CardHeader>
          <CardContent className="p-0">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="pl-6">{t("nodes.name")}</TableHead>
                  <TableHead>{t("nodes.group")}</TableHead>
                  <TableHead>{t("nodes.apiAddress")}</TableHead>
                  <TableHead>{t("nodes.publicAddress")}</TableHead>
                  <TableHead>{t("nodes.status")}</TableHead>
                  <TableHead className="w-12 pr-6">
                    <span className="sr-only">{t("common.actions")}</span>
                  </TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {nodesQuery.isLoading ? (
                  <>
                    {Array.from({ length: 5 }).map((_, i) => (
                      <TableRow key={i}>
                        <TableCell className="pl-6">
                          <Skeleton className="h-4 w-28" />
                        </TableCell>
                        <TableCell>
                          <Skeleton className="h-4 w-24" />
                        </TableCell>
                        <TableCell>
                          <Skeleton className="h-4 w-40" />
                        </TableCell>
                        <TableCell>
                          <Skeleton className="h-4 w-40" />
                        </TableCell>
                        <TableCell className="pr-6">
                          <Skeleton className="h-8 w-8" />
                        </TableCell>
                      </TableRow>
                    ))}
                  </>
                ) : null}
                {nodesQuery.data?.map((n) => (
                  <TableRow key={n.id}>
                    <TableCell className="pl-6 font-medium">{n.name}</TableCell>
                    <TableCell className="text-muted-foreground">
                      {groupName(groupsQuery.data, n.group_id)}
                    </TableCell>
                    <TableCell className="text-muted-foreground">
                      {n.api_address}:{n.api_port}
                    </TableCell>
                    <TableCell className="text-muted-foreground">{n.public_address}</TableCell>
                    <TableCell>
                      <StatusDot
                        status={n.status}
                        labelOnline={t("nodes.statusOnline")}
                        labelOffline={t("nodes.statusOffline")}
                        labelUnknown={t("nodes.statusUnknown")}
                      />
                    </TableCell>
                    <TableCell className="pr-6">
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
                              setActionMessage(null)
                              createMutation.reset()
                              updateMutation.reset()
                              setUpserting({
                                mode: "edit",
                                node: n,
                                name: n.name,
                                apiAddress: n.api_address,
                                apiPort: n.api_port,
                                secretKey: n.secret_key,
                                publicAddress: n.public_address,
                                groupID: n.group_id,
                              })
                            }}
                          >
                            <Pencil className="mr-2 size-4" />
                            {t("common.edit")}
                          </DropdownMenuItem>
                          <DropdownMenuItem
                            disabled={healthMutation.isPending}
                            onClick={() => healthMutation.mutate(n.id)}
                          >
                            <Stethoscope className="mr-2 size-4" />
                            {t("nodes.health")}
                          </DropdownMenuItem>
                          <DropdownMenuItem
                            disabled={syncMutation.isPending}
                            onClick={() => syncMutation.mutate(n.id)}
                          >
                            <RefreshCw className="mr-2 size-4" />
                            {t("nodes.sync")}
                          </DropdownMenuItem>
                          <DropdownMenuItem
                            onClick={() => {
                              setActionMessage(null)
                              setTrafficNode(n)
                            }}
                          >
                            {t("nodes.traffic")}
                          </DropdownMenuItem>
                          <DropdownMenuSeparator />
                          <DropdownMenuItem
                            variant="destructive"
                            disabled={deleteMutation.isPending}
                            onClick={() => deleteMutation.mutate(n.id)}
                          >
                            <Trash2 className="mr-2 size-4" />
                            {t("common.delete")}
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </TableCell>
                  </TableRow>
                ))}
                {!nodesQuery.isLoading && nodesQuery.data && nodesQuery.data.length === 0 ? (
                  <TableRow>
                    <TableCell className="pl-6 py-8 text-center text-muted-foreground" colSpan={6}>
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
            aria-label={upserting?.mode === "create" ? t("nodes.createNode") : t("nodes.editNode")}
          >
            <DialogHeader>
              <DialogTitle>
                {upserting?.mode === "create" ? t("nodes.createNode") : t("nodes.editNode")}
              </DialogTitle>
              {upserting?.mode === "edit" ? (
                <DialogDescription>{upserting.node.name}</DialogDescription>
              ) : null}
            </DialogHeader>

            {upserting ? (
              <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                <div className="space-y-1 md:col-span-2">
                  <Label className="text-sm text-slate-700" htmlFor="node-name">
                    {t("nodes.name")}
                  </Label>
                  <Input
                    id="node-name"
                    value={upserting.name}
                    onChange={(e) => setUpserting((p) => (p ? { ...p, name: e.target.value } : p))}
                    placeholder={t("nodes.namePlaceholder")}
                    autoFocus={upserting.mode === "create"}
                  />
                </div>

                <div className="space-y-1">
                  <Label className="text-sm text-slate-700" htmlFor="node-api-addr">
                    {t("nodes.apiAddress")}
                  </Label>
                  <Input
                    id="node-api-addr"
                    value={upserting.apiAddress}
                    onChange={(e) =>
                      setUpserting((p) => (p ? { ...p, apiAddress: e.target.value } : p))
                    }
                    placeholder={t("nodes.apiHostPlaceholder")}
                  />
                </div>

                <div className="space-y-1">
                  <Label className="text-sm text-slate-700" htmlFor="node-api-port">
                    {t("nodes.apiPort")}
                  </Label>
                  <Input
                    id="node-api-port"
                    type="number"
                    value={upserting.apiPort}
                    onChange={(e) =>
                      setUpserting((p) =>
                        p ? { ...p, apiPort: Number(e.target.value || 0) } : p,
                      )
                    }
                    min={1}
                  />
                </div>

                <div className="space-y-1 md:col-span-2">
                  <Label className="text-sm text-slate-700" htmlFor="node-secret">
                    {t("nodes.secretKey")}
                  </Label>
                  <div className="flex gap-2">
                    <Input
                      id="node-secret"
                      value={upserting.secretKey}
                      onChange={(e) =>
                        setUpserting((p) => (p ? { ...p, secretKey: e.target.value } : p))
                      }
                      placeholder={t("nodes.secretKeyPlaceholder")}
                    />
                    <Button
                      type="button"
                      variant="outline"
                      onClick={() => {
                        const key = generateNodeSecretKey(32)
                        setUpserting((p) => (p ? { ...p, secretKey: key } : p))
                      }}
                    >
                      {t("nodes.generateSecret")}
                    </Button>
                  </div>
                </div>

                <div className="md:col-span-2">
                  <Separator className="my-1" />
                  <div className="space-y-2">
                    <div>
                      <div className="text-sm font-medium">{t("nodes.deployTitle")}</div>
                      <div className="text-xs text-muted-foreground">{t("nodes.deploySubtitle")}</div>
                    </div>

                    <pre className="bg-muted overflow-x-auto rounded-md p-3 text-xs leading-relaxed">
                      <code>
                        {buildNodeDockerCompose({
                          port: upserting.apiPort,
                          secretKey: upserting.secretKey.trim() || "change-me",
                          logLevel: "info",
                        })}
                      </code>
                    </pre>

                    <div className="flex justify-end">
                      <Button
                        type="button"
                        variant="outline"
                        onClick={async () => {
                          const yaml = buildNodeDockerCompose({
                            port: upserting.apiPort,
                            secretKey: upserting.secretKey.trim() || "change-me",
                            logLevel: "info",
                          })
                          await navigator.clipboard.writeText(yaml)
                          setActionMessage(t("nodes.composeCopied"))
                        }}
                      >
                        {t("nodes.copyCompose")}
                      </Button>
                    </div>
                  </div>
                </div>

                <div className="space-y-1 md:col-span-2">
                  <Label className="text-sm text-slate-700" htmlFor="node-public">
                    {t("nodes.publicAddress")}
                  </Label>
                  <Input
                    id="node-public"
                    value={upserting.publicAddress}
                    onChange={(e) =>
                      setUpserting((p) => (p ? { ...p, publicAddress: e.target.value } : p))
                    }
                    placeholder={t("nodes.publicAddressPlaceholder")}
                  />
                </div>

                <div className="space-y-1 md:col-span-2">
                  <Label className="text-sm text-slate-700">{t("nodes.group")}</Label>
                  <Select
                    value={upserting.groupID == null ? "none" : String(upserting.groupID)}
                    onValueChange={(v) =>
                      setUpserting((p) =>
                        p ? { ...p, groupID: v === "none" ? null : Number(v) } : p,
                      )
                    }
                  >
                    <SelectTrigger aria-label={t("nodes.selectGroup")}>
                      <SelectValue placeholder={t("nodes.selectGroup")} />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="none">{t("nodes.noGroup")}</SelectItem>
                      {groupsQuery.data?.map((g) => (
                        <SelectItem key={g.id} value={String(g.id)}>
                          {g.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  <p className="text-xs text-slate-500">
                    {t("nodes.groupRequiredHint")}
                  </p>
                </div>

                <div className="text-sm text-amber-700 md:col-span-2">
                  {createMutation.isError || updateMutation.isError ? (
                    (createMutation.error instanceof ApiError
                      ? createMutation.error.message
                      : updateMutation.error instanceof ApiError
                        ? updateMutation.error.message
                        : t("nodes.saveFailed"))
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
                  const name = upserting.name.trim()
                  const api_address = upserting.apiAddress.trim()
                  const secret_key = upserting.secretKey.trim()
                  const public_address = upserting.publicAddress.trim()
                  const api_port = upserting.apiPort
                  if (!name || !api_address || !secret_key || !public_address || api_port <= 0) return

                  const payload = {
                    name,
                    api_address,
                    api_port,
                    secret_key,
                    public_address,
                    group_id: upserting.groupID,
                  }

                  if (upserting.mode === "create") {
                    createMutation.mutate(payload)
                  } else {
                    updateMutation.mutate({ id: upserting.node.id, payload })
                  }
                }}
                disabled={createMutation.isPending || updateMutation.isPending}
              >
                {t("common.save")}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        <Dialog open={!!trafficNode} onOpenChange={(open) => (!open ? setTrafficNode(null) : null)}>
          <DialogContent aria-label={t("nodes.traffic")} className="max-w-3xl">
            <DialogHeader>
              <DialogTitle>{t("nodes.traffic")}</DialogTitle>
              {trafficNode ? <DialogDescription>{trafficNode.name}</DialogDescription> : null}
            </DialogHeader>

            <div className="space-y-3">
              <div className="text-xs text-muted-foreground">
                {trafficQuery.isLoading ? t("common.loading") : null}
                {trafficQuery.isError ? (
                  trafficQuery.error instanceof ApiError ? (
                    trafficQuery.error.message
                  ) : (
                    t("common.loadFailed")
                  )
                ) : null}
              </div>

              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>{t("inbounds.tag")}</TableHead>
                    <TableHead>{t("users.traffic")}</TableHead>
                    <TableHead>{t("nodes.lastSampleAt")}</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {trafficByInbound.map((r) => (
                    <TableRow key={r.inbound}>
                      <TableCell className="font-medium">{r.inbound}</TableCell>
                      <TableCell className="text-muted-foreground">
                        ↑ {(r.upload / (1024 ** 3)).toFixed(3)} GB
                        {"  "}
                        ↓ {(r.download / (1024 ** 3)).toFixed(3)} GB
                      </TableCell>
                      <TableCell className="text-muted-foreground">{r.last || "-"}</TableCell>
                    </TableRow>
                  ))}
                  {!trafficQuery.isLoading && trafficByInbound.length === 0 ? (
                    <TableRow>
                      <TableCell colSpan={3} className="py-8 text-center text-muted-foreground">
                        {t("common.noData")}
                      </TableCell>
                    </TableRow>
                  ) : null}
                </TableBody>
              </Table>
            </div>

            <DialogFooter>
              <Button variant="outline" onClick={() => setTrafficNode(null)}>
                {t("common.cancel")}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </section>
    </div>
  )
}
