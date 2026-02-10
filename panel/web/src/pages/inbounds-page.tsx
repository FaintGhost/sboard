import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { MoreHorizontal, Pencil, Trash2 } from "lucide-react"
import { lazy, Suspense, useMemo, useState } from "react"
import { useTranslation } from "react-i18next"

import { AsyncButton } from "@/components/ui/async-button"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { PageHeader } from "@/components/page-header"
import { TableEmptyState } from "@/components/table-empty-state"
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
import { Label } from "@/components/ui/label"
import { FieldHint } from "@/components/ui/field-hint"
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
import { ApiError } from "@/lib/api/client"
import { createInbound, deleteInbound, listInbounds, updateInbound } from "@/lib/api/inbounds"
import { listNodes } from "@/lib/api/nodes"
import { checkSingBoxConfig, formatSingBoxConfig, generateSingBoxValue } from "@/lib/api/singbox-tools"
import type { Inbound, Node, SingBoxGenerateCommand } from "@/lib/api/types"
import {
  buildInboundTemplateText,
  buildPresetInboundTemplateText,
  inboundTemplatePresetProtocols,
  parseInboundTemplateToPayload,
  readTemplateProtocol,
  type InboundTemplatePresetProtocol,
} from "@/lib/inbound-template"
import { tableColumnSpacing } from "@/lib/table-spacing"
import { tableTransitionClass } from "@/lib/table-motion"
import { useTableQueryTransition } from "@/lib/table-query-transition"
import { tableToolbarClass } from "@/lib/table-toolbar"

const MonacoEditor = lazy(() => import("@monaco-editor/react"))

type TemplatePreset = InboundTemplatePresetProtocol | "custom"

type EditState = {
  mode: "create" | "edit"
  inbound: Inbound
  nodeID: number
  templateText: string
  preset: TemplatePreset
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

const defaultPreset: InboundTemplatePresetProtocol = "vless"

const generateCommandOptions: Array<{ value: SingBoxGenerateCommand; labelKey: string }> = [
  { value: "uuid", labelKey: "inbounds.generateUuid" },
  { value: "reality-keypair", labelKey: "inbounds.generateReality" },
  { value: "wg-keypair", labelKey: "inbounds.generateWg" },
  { value: "vapid-keypair", labelKey: "inbounds.generateVapid" },
  { value: "rand-base64-16", labelKey: "inbounds.generateRand16" },
  { value: "rand-base64-32", labelKey: "inbounds.generateRand32" },
]

function nodeName(nodes: Node[] | undefined, id: number): string {
  if (!nodes) return String(id)
  const node = nodes.find((item) => item.id === id)
  return node ? node.name : String(id)
}

function templateHint(t: (key: string) => string, input: string): string | null {
  const parsed = parseInboundTemplateToPayload(input)
  return parsed.ok ? null : `${t("inbounds.templateParseFailed")}: ${parsed.error}`
}

function presetTextKey(preset: TemplatePreset): string {
  switch (preset) {
    case "vless":
      return "inbounds.presetVless"
    case "vmess":
      return "inbounds.presetVmess"
    case "trojan":
      return "inbounds.presetTrojan"
    case "shadowsocks":
      return "inbounds.presetShadowsocks"
    default:
      return "inbounds.presetCustom"
  }
}

function toApiErrorMessage(error: unknown, fallback: string): string {
  if (error instanceof ApiError) return error.message
  if (error instanceof Error) return error.message
  return fallback
}

function applyGeneratedValueToTemplate(
  templateText: string,
  command: SingBoxGenerateCommand,
  output: string,
): string {
  if (command !== "rand-base64-16" && command !== "rand-base64-32") {
    return templateText
  }

  try {
    const parsed = JSON.parse(templateText) as Record<string, unknown>
    if (!parsed || typeof parsed !== "object" || Array.isArray(parsed)) {
      return templateText
    }

    const inbounds = parsed.inbounds
    if (Array.isArray(inbounds) && inbounds.length > 0) {
      const first = inbounds[0]
      if (first && typeof first === "object" && !Array.isArray(first)) {
        ;(first as Record<string, unknown>).password = output.trim()
        return JSON.stringify(parsed, null, 2)
      }
    }

    parsed.password = output.trim()
    return JSON.stringify(parsed, null, 2)
  } catch {
    return templateText
  }
}

export function InboundsPage() {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const spacing = tableColumnSpacing.five
  const [nodeFilter, setNodeFilter] = useState<number | "all">("all")
  const [upserting, setUpserting] = useState<EditState | null>(null)
  const [toolMessage, setToolMessage] = useState<{ tone: "ok" | "error"; text: string } | null>(null)
  const [generateCommand, setGenerateCommand] = useState<SingBoxGenerateCommand>("uuid")
  const [generateOutput, setGenerateOutput] = useState<string>("")

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

  const inboundsTable = useTableQueryTransition({
    filterKey: String(nodeFilter),
    rows: inboundsQuery.data,
    isLoading: inboundsQuery.isLoading,
    isFetching: inboundsQuery.isFetching,
    isError: inboundsQuery.isError,
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

  const formatMutation = useMutation({
    mutationFn: (config: string) => formatSingBoxConfig({ config, mode: "inbound" }),
  })

  const checkMutation = useMutation({
    mutationFn: (config: string) => checkSingBoxConfig({ config, mode: "inbound" }),
  })

  const generateMutation = useMutation({
    mutationFn: (command: SingBoxGenerateCommand) => generateSingBoxValue(command),
  })

  const openCreateDialog = () => {
    const firstNode = nodesQuery.data?.[0]
    if (!firstNode) return

    createMutation.reset()
    updateMutation.reset()
    formatMutation.reset()
    checkMutation.reset()
    generateMutation.reset()
    setToolMessage(null)
    setGenerateOutput("")
    setGenerateCommand("uuid")

    setUpserting({
      mode: "create",
      inbound: defaultNewInbound,
      nodeID: firstNode.id,
      templateText: buildPresetInboundTemplateText(defaultPreset),
      preset: defaultPreset,
    })
  }

  const openEditDialog = (inbound: Inbound) => {
    createMutation.reset()
    updateMutation.reset()
    formatMutation.reset()
    checkMutation.reset()
    generateMutation.reset()
    setToolMessage(null)
    setGenerateOutput("")
    setGenerateCommand("uuid")

    const templateText = buildInboundTemplateText(inbound)
    setUpserting({
      mode: "edit",
      inbound,
      nodeID: inbound.node_id,
      templateText,
      preset: readTemplateProtocol(templateText) ?? "custom",
    })
  }

  const currentTemplateHint = upserting ? templateHint(t, upserting.templateText) : null

  const mutationErrorText =
    createMutation.isError || updateMutation.isError
      ? toApiErrorMessage(
          createMutation.error ?? updateMutation.error,
          t("inbounds.saveFailed"),
        )
      : null

  return (
    <div className="px-4 lg:px-6">
      <section className="space-y-6">
        <PageHeader
          title={t("inbounds.title")}
          description={t("inbounds.subtitle")}
          action={(
            <Button
              onClick={openCreateDialog}
              disabled={!nodesQuery.data || nodesQuery.data.length === 0}
            >
              {t("inbounds.createInbound")}
            </Button>
          )}
        />

        <Card>
          <CardHeader className="pb-3">
            <div className={tableToolbarClass.container}>
              <div className="flex flex-col gap-1.5">
                <CardTitle className="text-base">{t("inbounds.list")}</CardTitle>
                <CardDescription>
                  {inboundsTable.showLoadingHint ? t("common.loading") : null}
                  {inboundsQuery.isError ? t("common.loadFailed") : null}
                  {!inboundsTable.showLoadingHint && inboundsQuery.data
                    ? t("inbounds.count", { count: inboundsTable.visibleRows.length })
                    : null}
                </CardDescription>
              </div>
              <div className={tableToolbarClass.filters}>
                <Select
                  value={nodeFilter === "all" ? "all" : String(nodeFilter)}
                  onValueChange={(value) => setNodeFilter(value === "all" ? "all" : Number(value))}
                >
                  <SelectTrigger className="w-full sm:w-56" aria-label={t("inbounds.nodeFilter")}>
                    <SelectValue placeholder={t("inbounds.selectNode")} />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">{t("inbounds.allNodes")}</SelectItem>
                    {nodesQuery.data?.map((node) => (
                      <SelectItem key={node.id} value={String(node.id)}>
                        {node.name}
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
              <TableBody className={tableTransitionClass(inboundsTable.isTransitioning)}>
                {inboundsTable.visibleRows.map((inbound) => (
                  <TableRow key={inbound.id}>
                    <TableCell className={`${spacing.cellFirst} font-medium`}>
                      {nodeName(nodesQuery.data, inbound.node_id)}
                    </TableCell>
                    <TableCell className={`${spacing.cellMiddle} font-medium`}>{inbound.tag}</TableCell>
                    <TableCell className={`${spacing.cellMiddle} text-muted-foreground`}>
                      {inbound.protocol}
                    </TableCell>
                    <TableCell className={`${spacing.cellMiddle} text-muted-foreground`}>
                      {inbound.public_port > 0
                        ? `${inbound.public_port} (${t("inbounds.publicPortShort")})`
                        : inbound.listen_port}
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
                          <DropdownMenuItem onClick={() => openEditDialog(inbound)}>
                            <Pencil className="mr-2 size-4" />
                            {t("common.edit")}
                          </DropdownMenuItem>
                          <DropdownMenuSeparator />
                          <DropdownMenuItem
                            variant="destructive"
                            disabled={deleteMutation.isPending}
                            onClick={() => deleteMutation.mutate(inbound.id)}
                          >
                            <Trash2 className="mr-2 size-4" />
                            {t("common.delete")}
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </TableCell>
                  </TableRow>
                ))}

                {inboundsTable.showNoData ? (
                  <TableEmptyState
                    colSpan={5}
                    className={`${spacing.cellFirst} py-10 text-center`}
                    message={t("common.noData")}
                    actionLabel={t("inbounds.createInbound")}
                    actionTo="/inbounds"
                  />
                ) : null}
              </TableBody>
            </Table>
          </CardContent>
        </Card>

        <Dialog open={!!upserting} onOpenChange={(open) => (!open ? setUpserting(null) : null)}>
          <DialogContent
            className="sm:max-w-4xl"
            aria-label={upserting?.mode === "create" ? t("inbounds.createInbound") : t("inbounds.editInbound")}
          >
            <DialogHeader>
              <DialogTitle>
                {upserting?.mode === "create" ? t("inbounds.createInbound") : t("inbounds.editInbound")}
              </DialogTitle>
              <DialogDescription>
                {upserting?.mode === "edit" ? upserting.inbound.tag : t("inbounds.createInbound")}
              </DialogDescription>
            </DialogHeader>

            {upserting ? (
              <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                <div className="space-y-1 md:col-span-2">
                  <Label className="text-sm text-slate-700">{t("inbounds.node")}</Label>
                  <Select
                    value={String(upserting.nodeID)}
                    onValueChange={(value) =>
                      setUpserting((prev) => (prev ? { ...prev, nodeID: Number(value) } : prev))
                    }
                  >
                    <SelectTrigger aria-label={t("inbounds.selectNode")}>
                      <SelectValue placeholder={t("inbounds.selectNode")} />
                    </SelectTrigger>
                    <SelectContent>
                      {nodesQuery.data?.map((node) => (
                        <SelectItem key={node.id} value={String(node.id)}>
                          {node.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>

                <div className="space-y-2 md:col-span-2">
                  <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
                    <div className="flex items-center gap-1">
                      <Label className="text-sm text-slate-700">{t("inbounds.template")}</Label>
                      <FieldHint label={t("inbounds.template")}>
                        <p>{t("inbounds.templateHelp")}</p>
                        <p className="mt-1">{t("inbounds.usersManagedHint")}</p>
                      </FieldHint>
                    </div>
                    <div className="flex flex-col gap-2 sm:flex-row sm:items-center">
                      <Label className="text-xs text-slate-500">{t("inbounds.templatePreset")}</Label>
                      <Select
                        value={upserting.preset}
                        onValueChange={(value) => {
                          const preset = value as TemplatePreset
                          if (preset === "custom") {
                            setUpserting((prev) => (prev ? { ...prev, preset } : prev))
                            return
                          }
                          setUpserting((prev) =>
                            prev
                              ? {
                                  ...prev,
                                  preset,
                                  templateText: buildPresetInboundTemplateText(preset),
                                }
                              : prev,
                          )
                          setToolMessage(null)
                        }}
                      >
                        <SelectTrigger className="w-full sm:w-56" aria-label={t("inbounds.templatePreset")}>
                          <SelectValue placeholder={t("inbounds.templatePreset")} />
                        </SelectTrigger>
                        <SelectContent>
                          {inboundTemplatePresetProtocols.map((protocol) => (
                            <SelectItem key={protocol} value={protocol}>
                              {t(presetTextKey(protocol))}
                            </SelectItem>
                          ))}
                          <SelectItem value="custom">{t("inbounds.presetCustom")}</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>
                  </div>

                  <div className="overflow-hidden rounded-md border">
                    <Suspense fallback={<div className="h-[360px] px-3 py-2 text-sm text-muted-foreground">加载中...</div>}>
                      <MonacoEditor
                        height="360px"
                        defaultLanguage="json"
                        language="json"
                        theme="vs-dark"
                        value={upserting.templateText}
                        onChange={(value) => {
                          setUpserting((prev) =>
                            prev
                              ? {
                                  ...prev,
                                  templateText: value ?? "",
                                  preset: readTemplateProtocol(value ?? "") ?? "custom",
                                }
                              : prev,
                          )
                          setToolMessage(null)
                        }}
                        options={{
                          minimap: { enabled: false },
                          fontSize: 13,
                          lineNumbersMinChars: 3,
                          scrollBeyondLastLine: false,
                          wordWrap: "on",
                          tabSize: 2,
                          automaticLayout: true,
                        }}
                      />
                    </Suspense>
                  </div>

                  <div className="flex flex-col gap-2 sm:flex-row sm:flex-wrap sm:items-center">
                    <AsyncButton
                      type="button"
                      variant="outline"
                      size="sm"
                      disabled={formatMutation.isPending || !upserting.templateText.trim()}
                      pending={formatMutation.isPending}
                      pendingText={t("common.loading")}
                      onClick={() => {
                        formatMutation.mutate(upserting.templateText, {
                          onSuccess: (result) => {
                            setUpserting((prev) =>
                              prev
                                ? {
                                    ...prev,
                                    templateText: result.formatted,
                                    preset: readTemplateProtocol(result.formatted) ?? "custom",
                                  }
                                : prev,
                            )
                            setToolMessage({ tone: "ok", text: t("inbounds.formatSuccess") })
                          },
                          onError: (error) => {
                            setToolMessage({
                              tone: "error",
                              text: toApiErrorMessage(error, t("inbounds.formatFailed")),
                            })
                          },
                        })
                      }}
                    >
                      {t("inbounds.formatTemplate")}
                    </AsyncButton>

                    <AsyncButton
                      type="button"
                      variant="outline"
                      size="sm"
                      disabled={checkMutation.isPending || !upserting.templateText.trim()}
                      pending={checkMutation.isPending}
                      pendingText={t("common.loading")}
                      onClick={() => {
                        checkMutation.mutate(upserting.templateText, {
                          onSuccess: (result) => {
                            if (result.ok) {
                              setToolMessage({ tone: "ok", text: t("inbounds.checkSuccess") })
                            } else {
                              setToolMessage({
                                tone: "error",
                                text: `${t("inbounds.checkFailed")}: ${result.output}`,
                              })
                            }
                          },
                          onError: (error) => {
                            setToolMessage({
                              tone: "error",
                              text: toApiErrorMessage(error, t("inbounds.checkFailed")),
                            })
                          },
                        })
                      }}
                    >
                      {t("inbounds.checkTemplate")}
                    </AsyncButton>

                    <Select
                      value={generateCommand}
                      onValueChange={(value) => setGenerateCommand(value as SingBoxGenerateCommand)}
                    >
                      <SelectTrigger className="w-full sm:w-56" aria-label={t("inbounds.generateCommand")}>
                        <SelectValue placeholder={t("inbounds.generateCommand")} />
                      </SelectTrigger>
                      <SelectContent>
                        {generateCommandOptions.map((option) => (
                          <SelectItem key={option.value} value={option.value}>
                            {t(option.labelKey)}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>

                    <AsyncButton
                      type="button"
                      variant="outline"
                      size="sm"
                      disabled={generateMutation.isPending}
                      pending={generateMutation.isPending}
                      pendingText={t("common.loading")}
                      onClick={() => {
                        generateMutation.mutate(generateCommand, {
                          onSuccess: (result) => {
                            setGenerateOutput(result.output)
                            setUpserting((prev) =>
                              prev
                                ? {
                                    ...prev,
                                    templateText: applyGeneratedValueToTemplate(
                                      prev.templateText,
                                      generateCommand,
                                      result.output,
                                    ),
                                  }
                                : prev,
                            )
                            setToolMessage({
                              tone: "ok",
                              text:
                                generateCommand === "rand-base64-16" || generateCommand === "rand-base64-32"
                                  ? t("inbounds.generateAndFillPassword")
                                  : t("inbounds.generateSuccess"),
                            })
                          },
                          onError: (error) => {
                            setToolMessage({
                              tone: "error",
                              text: toApiErrorMessage(error, t("inbounds.generateFailed")),
                            })
                          },
                        })
                      }}
                    >
                      {t("inbounds.generateRun")}
                    </AsyncButton>
                  </div>

                  {currentTemplateHint ? (
                    <p className="text-xs text-amber-700">{currentTemplateHint}</p>
                  ) : null}

                  {toolMessage ? (
                    <p className={toolMessage.tone === "ok" ? "text-xs text-emerald-700" : "text-xs text-amber-700"}>
                      {toolMessage.text}
                    </p>
                  ) : null}

                  {generateOutput ? (
                    <div className="rounded-md border bg-muted/30 p-2">
                      <p className="mb-1 text-xs text-slate-500">{t("inbounds.generateOutput")}</p>
                      <pre className="whitespace-pre-wrap break-all text-xs text-slate-700">{generateOutput}</pre>
                    </div>
                  ) : null}
                </div>

                <div className="text-sm text-amber-700 md:col-span-2">
                  {mutationErrorText}
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
              <AsyncButton
                onClick={() => {
                  if (!upserting) return
                  if (upserting.nodeID <= 0) return

                  const parsedTemplate = parseInboundTemplateToPayload(upserting.templateText)
                  if (!parsedTemplate.ok) return

                  const payload = {
                    node_id: upserting.nodeID,
                    ...parsedTemplate.payload,
                  }

                  if (upserting.mode === "create") {
                    createMutation.mutate(payload)
                  } else {
                    updateMutation.mutate({ id: upserting.inbound.id, payload })
                  }
                }}
                disabled={createMutation.isPending || updateMutation.isPending}
                pending={createMutation.isPending || updateMutation.isPending}
                pendingText={
                  upserting?.mode === "create"
                    ? t("common.creating")
                    : t("common.saving")
                }
              >
                {t("common.save")}
              </AsyncButton>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </section>
    </div>
  )
}
