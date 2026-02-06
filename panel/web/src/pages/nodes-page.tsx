import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useMemo, useState } from "react"
import { useTranslation } from "react-i18next"
import { MoreHorizontal, Pencil, RefreshCw, Stethoscope, Trash2 } from "lucide-react"

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
import { Skeleton } from "@/components/ui/skeleton"
import { ApiError } from "@/lib/api/client"
import { listGroups } from "@/lib/api/groups"
import { createNode, deleteNode, listNodes, nodeHealth, nodeSync, updateNode } from "@/lib/api/nodes"
import type { Group, Node } from "@/lib/api/types"

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
  const [actionMessage, setActionMessage] = useState<string | null>(null)

  const queryParams = useMemo(() => ({ limit: 50, offset: 0 }), [])
  const nodesQuery = useQuery({
    queryKey: ["nodes", queryParams],
    queryFn: () => listNodes(queryParams),
  })

  const groupsQuery = useQuery({
    queryKey: ["groups", queryParams],
    queryFn: () => listGroups(queryParams),
  })

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
                    <TableCell className="pl-6 py-8 text-center text-muted-foreground" colSpan={5}>
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
                  <Input
                    id="node-secret"
                    value={upserting.secretKey}
                    onChange={(e) =>
                      setUpserting((p) => (p ? { ...p, secretKey: e.target.value } : p))
                    }
                    placeholder={t("nodes.secretKeyPlaceholder")}
                  />
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
      </section>
    </div>
  )
}
