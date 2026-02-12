import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { MoreHorizontal, Pencil, ShieldAlert, Trash2 } from "lucide-react";
import { AnimatePresence, motion, useReducedMotion } from "framer-motion";

import { AsyncButton } from "@/components/ui/async-button";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { PageHeader } from "@/components/page-header";
import { TableEmptyState } from "@/components/table-empty-state";
import { Separator } from "@/components/ui/separator";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import { Checkbox } from "@/components/ui/checkbox";
import { Label } from "@/components/ui/label";
import { FieldHint } from "@/components/ui/field-hint";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Skeleton } from "@/components/ui/skeleton";
import { StatusDot } from "@/components/status-dot";
import { FlashValue } from "@/components/flash-value";
import { ApiError } from "@/lib/api/client";
import { listGroups } from "@/lib/api/groups";
import { createNode, deleteNode, listNodeTraffic, listNodes, updateNode } from "@/lib/api/nodes";
import type { Group, Node, NodeTrafficSample } from "@/lib/api/types";
import { listTrafficNodesSummary, type TrafficNodeSummary } from "@/lib/api/traffic";
import { buildNodeDockerCompose, generateNodeSecretKey } from "@/lib/node-compose";
import { tableColumnLayout, tableColumnSpacing } from "@/lib/table-spacing";
import { formatDateTimeByTimezone } from "@/lib/datetime";
import { bytesToGBString } from "@/lib/units";
import { tableToolbarClass } from "@/lib/table-toolbar";
import { useSystemStore } from "@/store/system";

type EditState = {
  mode: "create" | "edit";
  node: Node;
  name: string;
  apiAddress: string;
  apiPort: number;
  secretKey: string;
  publicAddress: string;
  groupID: number | null;
  linkAddress: boolean;
};

type DeleteNodeState = {
  node: Node;
  force: boolean;
};

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
};

const motionEase: [number, number, number, number] = [0.22, 1, 0.36, 1];
const interactiveTableRowClass =
  "hover:bg-muted/50 data-[state=selected]:bg-muted border-b transition-colors";

function groupName(groups: Group[] | undefined, id: number | null): string {
  if (!groups || id == null) return "-";
  const g = groups.find((x) => x.id === id);
  return g ? g.name : String(id);
}

function formatDateTime(
  value: string | null | undefined,
  locale: string,
  timezone: string,
): string {
  return formatDateTimeByTimezone(value, locale, timezone);
}

export function NodesPage() {
  const { t, i18n } = useTranslation();
  const timezone = useSystemStore((state) => state.timezone);
  const prefersReducedMotion = useReducedMotion();
  const shouldAnimate = !prefersReducedMotion;
  const qc = useQueryClient();
  const spacing = tableColumnSpacing.seven;
  const layout = tableColumnLayout.sevenActionIcon;
  const trafficSpacing = tableColumnSpacing.three;
  const trafficLayout = tableColumnLayout.threeEven;
  const [upserting, setUpserting] = useState<EditState | null>(null);
  const [trafficNode, setTrafficNode] = useState<Node | null>(null);
  const [actionMessage, setActionMessage] = useState<string | null>(null);
  const [deleting, setDeleting] = useState<DeleteNodeState | null>(null);

  const queryParams = useMemo(() => ({ limit: 50, offset: 0 }), []);
  const nodesQuery = useQuery({
    queryKey: ["nodes", queryParams],
    queryFn: () => listNodes(queryParams),
    refetchInterval: 5_000,
  });

  const groupsQuery = useQuery({
    queryKey: ["groups", queryParams],
    queryFn: () => listGroups(queryParams),
  });

  const trafficSummary24hQuery = useQuery({
    queryKey: ["traffic", "nodes", "summary", "24h"],
    queryFn: () => listTrafficNodesSummary({ window: "24h" }),
    refetchInterval: 30_000,
  });

  const trafficSummary1hQuery = useQuery({
    queryKey: ["traffic", "nodes", "summary", "1h"],
    queryFn: () => listTrafficNodesSummary({ window: "1h" }),
    refetchInterval: 30_000,
  });

  const trafficSummaryByNodeID = useMemo(() => {
    const map24 = new Map<number, TrafficNodeSummary>();
    const map1 = new Map<number, TrafficNodeSummary>();
    for (const it of trafficSummary24hQuery.data ?? []) map24.set(it.node_id, it);
    for (const it of trafficSummary1hQuery.data ?? []) map1.set(it.node_id, it);
    return { map24, map1 };
  }, [trafficSummary24hQuery.data, trafficSummary1hQuery.data]);

  const createMutation = useMutation({
    mutationFn: createNode,
    onSuccess: async () => {
      setUpserting(null);
      await qc.invalidateQueries({ queryKey: ["nodes"] });
    },
  });

  const updateMutation = useMutation({
    mutationFn: (input: { id: number; payload: Record<string, unknown> }) =>
      updateNode(input.id, input.payload),
    onSuccess: async () => {
      setUpserting(null);
      await qc.invalidateQueries({ queryKey: ["nodes"] });
    },
  });

  const resolveDeleteNodeErrorMessage = (error: unknown): string => {
    if (error instanceof ApiError) {
      if (error.status === 409 || error.message.toLowerCase().includes("node is in use")) {
        return t("nodes.deleteInUse");
      }
      return error.message;
    }
    return t("nodes.deleteFailed");
  };

  const deleteMutation = useMutation({
    mutationFn: (input: { id: number; force: boolean }) =>
      deleteNode(input.id, { force: input.force }),
    onSuccess: async (_data, variables) => {
      setDeleting(null);
      setActionMessage(variables.force ? t("nodes.deleteForcedSuccess") : t("nodes.deleteSuccess"));
      await qc.invalidateQueries({ queryKey: ["nodes"] });
    },
    onError: (e) => {
      setActionMessage(resolveDeleteNodeErrorMessage(e));
    },
  });

  const deleteErrorMessage = resolveDeleteNodeErrorMessage(deleteMutation.error);

  const trafficQuery = useQuery({
    queryKey: ["nodes", "traffic", trafficNode?.id ?? 0],
    queryFn: async () => {
      if (!trafficNode) return [] as NodeTrafficSample[];
      return listNodeTraffic(trafficNode.id, { limit: 300, offset: 0 });
    },
    enabled: !!trafficNode,
    refetchInterval: trafficNode ? 10_000 : false,
  });

  const trafficByInbound = useMemo(() => {
    const rows = trafficQuery.data ?? [];
    const map = new Map<
      string,
      { inbound: string; upload: number; download: number; last: string }
    >();
    for (const r of rows) {
      const tag = r.inbound_tag ?? "(node)";
      const prev = map.get(tag);
      const last = prev ? (prev.last > r.recorded_at ? prev.last : r.recorded_at) : r.recorded_at;
      map.set(tag, {
        inbound: tag,
        upload: (prev?.upload ?? 0) + r.upload,
        download: (prev?.download ?? 0) + r.download,
        last,
      });
    }
    return Array.from(map.values()).sort((a, b) => (a.inbound < b.inbound ? -1 : 1));
  }, [trafficQuery.data]);

  return (
    <div className="px-4 lg:px-6">
      <motion.section
        className="space-y-6"
        initial={shouldAnimate ? { opacity: 0, y: 12 } : false}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.28, ease: motionEase }}
      >
        <PageHeader
          title={t("nodes.title")}
          description={t("nodes.subtitle")}
          action={
            <Button
              onClick={() => {
                setActionMessage(null);
                createMutation.reset();
                updateMutation.reset();
                setUpserting({
                  mode: "create",
                  node: defaultNewNode,
                  name: "",
                  apiAddress: "127.0.0.1",
                  apiPort: 3000,
                  secretKey: "",
                  publicAddress: "127.0.0.1",
                  groupID: null,
                  linkAddress: true,
                });
              }}
            >
              {t("nodes.createNode")}
            </Button>
          }
        />

        <motion.div
          className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3"
          initial={shouldAnimate ? { opacity: 0 } : false}
          animate={{ opacity: 1 }}
          transition={{ duration: 0.25, ease: motionEase }}
        >
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

          {nodesQuery.data?.map((n, index) => {
            const s24 = trafficSummaryByNodeID.map24.get(n.id);
            const s1 = trafficSummaryByNodeID.map1.get(n.id);
            const last = s24?.last_recorded_at || s1?.last_recorded_at || n.last_seen_at || "";
            const lastFormatted = formatDateTime(last, i18n.language, timezone);
            const up24 = s24?.upload ?? 0;
            const down24 = s24?.download ?? 0;
            const up1 = s1?.upload ?? 0;
            const down1 = s1?.download ?? 0;

            return (
              <motion.div
                key={n.id}
                layout
                initial={shouldAnimate ? { opacity: 0, y: 10, scale: 0.99 } : false}
                animate={{ opacity: 1, y: 0, scale: 1 }}
                transition={{
                  duration: 0.24,
                  delay: Math.min(index * 0.04, 0.24),
                  ease: motionEase,
                }}
              >
                <Card className="border-border/75 bg-card shadow-[0_1px_0_0_rgba(255,255,255,0.25)_inset,0_14px_30px_-30px_rgba(0,0,0,0.55)] dark:shadow-[0_1px_0_0_rgba(255,255,255,0.06)_inset,0_18px_34px_-28px_rgba(0,0,0,0.9)]">
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
                      <FlashValue
                        value={`↑ ${bytesToGBString(up1)} GB  ↓ ${bytesToGBString(down1)} GB`}
                      />
                    </div>
                    <div className="flex items-center justify-between gap-3">
                      <span className="text-muted-foreground">{t("traffic.window24h")}</span>
                      <FlashValue
                        value={`↑ ${bytesToGBString(up24)} GB  ↓ ${bytesToGBString(down24)} GB`}
                      />
                    </div>
                    <div className="flex items-center justify-between gap-3">
                      <span className="text-muted-foreground">{t("nodes.lastUpdatedAt")}</span>
                      <span className="truncate">
                        <FlashValue value={lastFormatted} className="max-w-full" />
                      </span>
                    </div>
                    <div className="flex justify-end gap-2 pt-2">
                      <Button
                        type="button"
                        size="sm"
                        onClick={() => {
                          setActionMessage(null);
                          setTrafficNode(n);
                        }}
                      >
                        {t("nodes.traffic")}
                      </Button>
                    </div>
                  </CardContent>
                </Card>
              </motion.div>
            );
          })}

          {!nodesQuery.isLoading && nodesQuery.data && nodesQuery.data.length === 0 ? (
            <Card className="shadow-xs md:col-span-2 xl:col-span-3">
              <CardContent className="py-10 text-center text-sm text-muted-foreground">
                {t("common.noData")}
              </CardContent>
            </Card>
          ) : null}
        </motion.div>

        <Card>
          <CardHeader className="pb-3">
            <div className={tableToolbarClass.container}>
              <div className="flex flex-col gap-1.5">
                <CardTitle className="text-base">{t("nodes.list")}</CardTitle>
                <CardDescription>
                  {nodesQuery.isLoading ? t("common.loading") : null}
                  {nodesQuery.isError ? t("common.loadFailed") : null}
                  {nodesQuery.data ? t("nodes.count", { count: nodesQuery.data.length }) : null}
                  <AnimatePresence mode="wait">
                    {actionMessage ? (
                      <motion.span
                        key={actionMessage}
                        className="ml-3 inline-block"
                        initial={shouldAnimate ? { opacity: 0, y: -4 } : false}
                        animate={{ opacity: 1, y: 0 }}
                        exit={shouldAnimate ? { opacity: 0, y: -4 } : undefined}
                        transition={{ duration: 0.18, ease: motionEase }}
                      >
                        {actionMessage}
                      </motion.span>
                    ) : null}
                  </AnimatePresence>
                </CardDescription>
              </div>
            </div>
          </CardHeader>
          <CardContent className="p-0">
            <Table className={layout.tableClass}>
              <TableHeader>
                <TableRow>
                  <TableHead className={`${spacing.headFirst} ${layout.headFirst}`}>
                    {t("nodes.name")}
                  </TableHead>
                  <TableHead className={`${spacing.headMiddle} ${layout.headMiddle}`}>
                    {t("nodes.group")}
                  </TableHead>
                  <TableHead className={`${spacing.headMiddle} ${layout.headMiddle}`}>
                    {t("nodes.apiAddress")}
                  </TableHead>
                  <TableHead className={`${spacing.headMiddle} ${layout.headMiddle}`}>
                    {t("nodes.publicAddress")}
                  </TableHead>
                  <TableHead className={`${spacing.headMiddle} ${layout.headMiddle}`}>
                    {t("nodes.status")}
                  </TableHead>
                  <TableHead className={`${spacing.headMiddle} ${layout.headMiddle}`}>
                    {t("nodes.lastSeen")}
                  </TableHead>
                  <TableHead className={`${spacing.headLast} ${layout.headLast}`}>
                    <span className="sr-only">{t("common.actions")}</span>
                  </TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {nodesQuery.isLoading ? (
                  <>
                    {Array.from({ length: 5 }).map((_, i) => (
                      <TableRow key={i}>
                        <TableCell className={spacing.cellFirst}>
                          <Skeleton className="h-4 w-28" />
                        </TableCell>
                        <TableCell className={spacing.cellMiddle}>
                          <Skeleton className="h-4 w-24" />
                        </TableCell>
                        <TableCell className={spacing.cellMiddle}>
                          <Skeleton className="h-4 w-40" />
                        </TableCell>
                        <TableCell className={spacing.cellMiddle}>
                          <Skeleton className="h-4 w-40" />
                        </TableCell>
                        <TableCell className={spacing.cellMiddle}>
                          <Skeleton className="h-4 w-16" />
                        </TableCell>
                        <TableCell className={spacing.cellMiddle}>
                          <Skeleton className="h-4 w-36" />
                        </TableCell>
                        <TableCell className={spacing.cellLast}>
                          <Skeleton className="h-8 w-8" />
                        </TableCell>
                      </TableRow>
                    ))}
                  </>
                ) : null}
                <AnimatePresence initial={false}>
                  {nodesQuery.data?.map((n, index) => (
                    <motion.tr
                      key={n.id}
                      layout
                      className={interactiveTableRowClass}
                      initial={shouldAnimate ? { opacity: 0, y: 6 } : false}
                      animate={{ opacity: 1, y: 0 }}
                      exit={shouldAnimate ? { opacity: 0, y: -4 } : undefined}
                      transition={{
                        duration: 0.18,
                        delay: Math.min(index * 0.02, 0.12),
                        ease: motionEase,
                      }}
                    >
                      <TableCell className={`${spacing.cellFirst} font-medium`}>{n.name}</TableCell>
                      <TableCell className={`${spacing.cellMiddle} text-muted-foreground`}>
                        {groupName(groupsQuery.data, n.group_id)}
                      </TableCell>
                      <TableCell className={`${spacing.cellMiddle} text-muted-foreground`}>
                        {n.api_address}:{n.api_port}
                      </TableCell>
                      <TableCell className={`${spacing.cellMiddle} text-muted-foreground`}>
                        {n.public_address}
                      </TableCell>
                      <TableCell className={spacing.cellMiddle}>
                        <StatusDot
                          status={n.status}
                          labelOnline={t("nodes.statusOnline")}
                          labelOffline={t("nodes.statusOffline")}
                          labelUnknown={t("nodes.statusUnknown")}
                        />
                      </TableCell>
                      <TableCell className={`${spacing.cellMiddle} text-muted-foreground`}>
                        {formatDateTime(n.last_seen_at, i18n.language, timezone)}
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
                                setActionMessage(null);
                                createMutation.reset();
                                updateMutation.reset();
                                setUpserting({
                                  mode: "edit",
                                  node: n,
                                  name: n.name,
                                  apiAddress: n.api_address,
                                  apiPort: n.api_port,
                                  secretKey: n.secret_key,
                                  publicAddress: n.public_address,
                                  groupID: n.group_id,
                                  linkAddress: n.public_address.trim() === n.api_address.trim(),
                                });
                              }}
                            >
                              <Pencil className="mr-2 size-4" />
                              {t("common.edit")}
                            </DropdownMenuItem>
                            <DropdownMenuSeparator />
                            <DropdownMenuItem
                              variant="destructive"
                              disabled={deleteMutation.isPending}
                              onClick={() => {
                                setActionMessage(null);
                                deleteMutation.reset();
                                setDeleting({ node: n, force: false });
                              }}
                            >
                              <Trash2 className="mr-2 size-4" />
                              {t("nodes.deleteNode")}
                            </DropdownMenuItem>
                          </DropdownMenuContent>
                        </DropdownMenu>
                      </TableCell>
                    </motion.tr>
                  ))}
                </AnimatePresence>
                {!nodesQuery.isLoading && nodesQuery.data && nodesQuery.data.length === 0 ? (
                  <TableEmptyState
                    colSpan={7}
                    className={`${spacing.cellFirst} py-10 text-center`}
                    message={t("common.noData")}
                    actionLabel={t("nodes.createNode")}
                    actionTo="/nodes"
                  />
                ) : null}
              </TableBody>
            </Table>
          </CardContent>
        </Card>

        <Dialog open={!!upserting} onOpenChange={(open) => (!open ? setUpserting(null) : null)}>
          <DialogContent
            aria-label={upserting?.mode === "create" ? t("nodes.createNode") : t("nodes.editNode")}
            className="sm:max-w-2xl max-h-[86dvh] overflow-hidden p-0"
          >
            <div className="flex h-full max-h-[86dvh] flex-col">
              <DialogHeader className="border-b px-6 pt-6 pb-4">
                <DialogTitle>
                  {upserting?.mode === "create" ? t("nodes.createNode") : t("nodes.editNode")}
                </DialogTitle>
                <DialogDescription>
                  {upserting?.mode === "edit" ? upserting.node.name : t("nodes.createNode")}
                </DialogDescription>
              </DialogHeader>

              {upserting ? (
                <div className="overflow-y-auto px-6 py-4">
                  <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                    <div className="space-y-1 md:col-span-2">
                      <Label className="text-sm text-foreground" htmlFor="node-name">
                        {t("nodes.name")}
                      </Label>
                      <Input
                        id="node-name"
                        value={upserting.name}
                        onChange={(e) =>
                          setUpserting((p) => (p ? { ...p, name: e.target.value } : p))
                        }
                        placeholder={t("nodes.namePlaceholder")}
                        autoFocus={upserting.mode === "create"}
                      />
                    </div>

                    <div className="space-y-1">
                      <Label className="text-sm text-foreground" htmlFor="node-api-addr">
                        {t("nodes.apiAddress")}
                      </Label>
                      <Input
                        id="node-api-addr"
                        value={upserting.apiAddress}
                        onChange={(e) =>
                          setUpserting((p) => {
                            if (!p) return p;
                            const apiAddress = e.target.value;
                            if (p.linkAddress) {
                              return { ...p, apiAddress, publicAddress: apiAddress };
                            }
                            return { ...p, apiAddress };
                          })
                        }
                        placeholder={t("nodes.apiHostPlaceholder")}
                      />
                    </div>

                    <div className="space-y-1">
                      <Label className="text-sm text-foreground" htmlFor="node-api-port">
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
                      <Label className="text-sm text-foreground" htmlFor="node-secret">
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
                            const key = generateNodeSecretKey(32);
                            setUpserting((p) => (p ? { ...p, secretKey: key } : p));
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
                          <div className="text-xs text-muted-foreground">
                            {t("nodes.deploySubtitle")}
                          </div>
                        </div>

                        <pre className="bg-muted max-h-52 overflow-auto rounded-md p-3 text-xs leading-relaxed">
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
                              });
                              await navigator.clipboard.writeText(yaml);
                              setActionMessage(t("nodes.composeCopied"));
                            }}
                          >
                            {t("nodes.copyCompose")}
                          </Button>
                        </div>
                      </div>
                    </div>

                    <div className="space-y-1 md:col-span-2">
                      <div className="flex items-center justify-between gap-2">
                        <Label className="text-sm text-foreground" htmlFor="node-public">
                          {t("nodes.publicAddress")}
                        </Label>
                        <div className="flex items-center gap-2 text-xs text-muted-foreground">
                          <Checkbox
                            id="node-link-address"
                            checked={upserting.linkAddress}
                            onCheckedChange={(checked) =>
                              setUpserting((p) => {
                                if (!p) return p;
                                const linkAddress = checked === true;
                                if (linkAddress) {
                                  return { ...p, linkAddress, publicAddress: p.apiAddress };
                                }
                                return { ...p, linkAddress };
                              })
                            }
                          />
                          <Label
                            htmlFor="node-link-address"
                            className="cursor-pointer text-xs text-muted-foreground"
                          >
                            {t("nodes.sameAsApiAddress")}
                          </Label>
                        </div>
                      </div>
                      <Input
                        id="node-public"
                        value={upserting.publicAddress}
                        onChange={(e) =>
                          setUpserting((p) => (p ? { ...p, publicAddress: e.target.value } : p))
                        }
                        placeholder={t("nodes.publicAddressPlaceholder")}
                        disabled={upserting.linkAddress}
                      />
                    </div>

                    <div className="space-y-1 md:col-span-2">
                      <div className="flex items-center gap-1">
                        <Label className="text-sm text-foreground">{t("nodes.group")}</Label>
                        <FieldHint label={t("nodes.group")}>
                          {t("nodes.groupRequiredHint")}
                        </FieldHint>
                      </div>
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
                    </div>

                    <div className="text-sm text-destructive md:col-span-2">
                      {createMutation.isError || updateMutation.isError
                        ? createMutation.error instanceof ApiError
                          ? createMutation.error.message
                          : updateMutation.error instanceof ApiError
                            ? updateMutation.error.message
                            : t("nodes.saveFailed")
                        : null}
                    </div>
                  </div>
                </div>
              ) : null}

              <DialogFooter className="border-t bg-background px-6 py-4">
                <Button
                  variant="outline"
                  onClick={() => setUpserting(null)}
                  disabled={createMutation.isPending || updateMutation.isPending}
                >
                  {t("common.cancel")}
                </Button>
                <AsyncButton
                  onClick={() => {
                    if (!upserting) return;
                    const name = upserting.name.trim();
                    const api_address = upserting.apiAddress.trim();
                    const secret_key = upserting.secretKey.trim();
                    const public_address = upserting.publicAddress.trim();
                    const api_port = upserting.apiPort;
                    if (!name || !api_address || !secret_key || !public_address || api_port <= 0)
                      return;

                    const payload = {
                      name,
                      api_address,
                      api_port,
                      secret_key,
                      public_address,
                      group_id: upserting.groupID,
                    };

                    if (upserting.mode === "create") {
                      createMutation.mutate(payload);
                    } else {
                      updateMutation.mutate({ id: upserting.node.id, payload });
                    }
                  }}
                  disabled={createMutation.isPending || updateMutation.isPending}
                  pending={createMutation.isPending || updateMutation.isPending}
                  pendingText={
                    upserting?.mode === "create" ? t("common.creating") : t("common.saving")
                  }
                >
                  {t("common.save")}
                </AsyncButton>
              </DialogFooter>
            </div>
          </DialogContent>
        </Dialog>

        <Dialog
          open={!!deleting}
          onOpenChange={(open) => {
            if (!open) {
              setDeleting(null);
              deleteMutation.reset();
            }
          }}
        >
          <DialogContent
            aria-label={t("nodes.deleteNode")}
            className="overflow-hidden p-0 sm:max-w-xl"
          >
            <DialogHeader className="border-b px-6 pt-6 pb-4">
              <div className="flex items-start gap-3">
                <span className="flex size-9 items-center justify-center rounded-full bg-muted text-destructive">
                  <ShieldAlert className="size-5" />
                </span>
                <div className="space-y-1">
                  <DialogTitle>{t("nodes.deleteDialogTitle")}</DialogTitle>
                  <DialogDescription>{t("nodes.deleteDialogDescription")}</DialogDescription>
                </div>
              </div>
            </DialogHeader>

            <div className="space-y-4 px-6 py-5">
              <div className="rounded-md border bg-muted/30 p-3">
                <p className="text-xs text-muted-foreground">{t("nodes.deleteDialogTarget")}</p>
                <p className="truncate text-sm font-semibold">{deleting?.node.name}</p>
              </div>

              <div className="rounded-md border bg-muted/20 p-3">
                <p className="text-xs font-medium text-foreground">
                  {t("nodes.deleteDialogImpact")}
                </p>
                <ul className="mt-2 space-y-1 text-xs text-muted-foreground">
                  <li>• {t("nodes.deleteImpactNode")}</li>
                  <li>• {t("nodes.deleteImpactInbounds")}</li>
                </ul>
              </div>

              <div className="rounded-lg border border-border/70 bg-background/70 p-3">
                <div className="flex items-start gap-3">
                  <Checkbox
                    id="node-force-delete"
                    checked={deleting?.force ?? false}
                    onCheckedChange={(checked) =>
                      setDeleting((prev) => (prev ? { ...prev, force: checked === true } : prev))
                    }
                    disabled={deleteMutation.isPending}
                  />
                  <div className="space-y-1">
                    <Label
                      htmlFor="node-force-delete"
                      className="cursor-pointer text-sm font-medium"
                    >
                      {t("nodes.forceDeleteLabel")}
                    </Label>
                    <p className="text-xs text-muted-foreground">{t("nodes.forceDeleteHint")}</p>
                  </div>
                </div>
              </div>

              {deleteMutation.isError ? (
                <p className="text-sm text-destructive">{deleteErrorMessage}</p>
              ) : null}
            </div>

            <DialogFooter className="border-t bg-background px-6 py-4">
              <Button
                variant="outline"
                onClick={() => {
                  setDeleting(null);
                  deleteMutation.reset();
                }}
                disabled={deleteMutation.isPending}
              >
                {t("common.cancel")}
              </Button>
              <AsyncButton
                variant="destructive"
                onClick={() => {
                  if (!deleting) return;
                  deleteMutation.mutate({ id: deleting.node.id, force: deleting.force });
                }}
                disabled={deleteMutation.isPending}
                pending={deleteMutation.isPending}
                pendingText={t("common.deleting")}
              >
                {deleting?.force ? t("nodes.forceDeleteAction") : t("nodes.deleteConfirmAction")}
              </AsyncButton>
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
                {trafficQuery.isError
                  ? trafficQuery.error instanceof ApiError
                    ? trafficQuery.error.message
                    : t("common.loadFailed")
                  : null}
              </div>

              <Table className={trafficLayout.tableClass}>
                <TableHeader>
                  <TableRow>
                    <TableHead className={`${trafficSpacing.headFirst} ${trafficLayout.headFirst}`}>
                      {t("inbounds.tag")}
                    </TableHead>
                    <TableHead
                      className={`${trafficSpacing.headMiddle} ${trafficLayout.headMiddle}`}
                    >
                      {t("users.traffic")}
                    </TableHead>
                    <TableHead className={`${trafficSpacing.headLast} ${trafficLayout.headLast}`}>
                      {t("nodes.lastUpdatedAt")}
                    </TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {trafficByInbound.map((r, index) => (
                    <motion.tr
                      key={r.inbound}
                      layout
                      className={interactiveTableRowClass}
                      initial={shouldAnimate ? { opacity: 0, y: 6 } : false}
                      animate={{ opacity: 1, y: 0 }}
                      transition={{
                        duration: 0.16,
                        delay: Math.min(index * 0.015, 0.12),
                        ease: motionEase,
                      }}
                    >
                      <TableCell className={`${trafficSpacing.cellFirst} font-medium`}>
                        {r.inbound}
                      </TableCell>
                      <TableCell className={`${trafficSpacing.cellMiddle} text-muted-foreground`}>
                        ↑ {(r.upload / 1024 ** 3).toFixed(3)} GB
                        {"  "}↓ {(r.download / 1024 ** 3).toFixed(3)} GB
                      </TableCell>
                      <TableCell className={`${trafficSpacing.cellLast} text-muted-foreground`}>
                        {formatDateTime(r.last, i18n.language, timezone)}
                      </TableCell>
                    </motion.tr>
                  ))}
                  {!trafficQuery.isLoading && trafficByInbound.length === 0 ? (
                    <TableRow>
                      <TableCell
                        colSpan={3}
                        className={`${trafficSpacing.cellFirst} py-8 text-center text-muted-foreground`}
                      >
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
      </motion.section>
    </div>
  );
}
