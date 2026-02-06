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
import { Skeleton } from "@/components/ui/skeleton"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { ApiError } from "@/lib/api/client"
import { createGroup, deleteGroup, listGroups, updateGroup } from "@/lib/api/groups"
import type { Group } from "@/lib/api/types"

type EditState = {
  mode: "create" | "edit"
  group: Group
  name: string
  description: string
}

const defaultNewGroup: Group = {
  id: 0,
  name: "",
  description: "",
}

export function GroupsPage() {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const [upserting, setUpserting] = useState<EditState | null>(null)

  const queryParams = useMemo(() => ({ limit: 50, offset: 0 }), [])

  const groupsQuery = useQuery({
    queryKey: ["groups", queryParams],
    queryFn: () => listGroups(queryParams),
  })

  const createMutation = useMutation({
    mutationFn: createGroup,
    onSuccess: async () => {
      setUpserting(null)
      await qc.invalidateQueries({ queryKey: ["groups"] })
    },
  })

  const updateMutation = useMutation({
    mutationFn: (input: { id: number; payload: { name?: string; description?: string } }) =>
      updateGroup(input.id, input.payload),
    onSuccess: async () => {
      setUpserting(null)
      await qc.invalidateQueries({ queryKey: ["groups"] })
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: number) => deleteGroup(id),
    onSuccess: async () => {
      await qc.invalidateQueries({ queryKey: ["groups"] })
    },
  })

  return (
    <div className="px-4 lg:px-6">
      <section className="space-y-6">
        <header className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
          <div>
            <h1 className="text-2xl font-semibold tracking-tight">{t("groups.title")}</h1>
            <p className="text-sm text-muted-foreground">{t("groups.subtitle")}</p>
          </div>
          <Button
            onClick={() => {
              createMutation.reset()
              updateMutation.reset()
              setUpserting({
                mode: "create",
                group: defaultNewGroup,
                name: "",
                description: "",
              })
            }}
          >
            {t("groups.createGroup")}
          </Button>
        </header>

        <Card>
          <CardHeader className="pb-3">
            <div className="flex flex-col gap-1.5">
              <CardTitle className="text-base">{t("groups.list")}</CardTitle>
              <CardDescription>
                {groupsQuery.isLoading ? t("common.loading") : null}
                {groupsQuery.isError ? t("common.loadFailed") : null}
                {groupsQuery.data ? t("groups.count", { count: groupsQuery.data.length }) : null}
              </CardDescription>
            </div>
          </CardHeader>
          <CardContent className="p-0">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="pl-6">{t("common.name")}</TableHead>
                  <TableHead>{t("common.description")}</TableHead>
                  <TableHead className="w-12 pr-6">
                    <span className="sr-only">{t("common.actions")}</span>
                  </TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {groupsQuery.isLoading ? (
                  <>
                    {Array.from({ length: 5 }).map((_, i) => (
                      <TableRow key={i}>
                        <TableCell className="pl-6">
                          <Skeleton className="h-4 w-28" />
                        </TableCell>
                        <TableCell>
                          <Skeleton className="h-4 w-56" />
                        </TableCell>
                        <TableCell className="pr-6">
                          <Skeleton className="h-8 w-8" />
                        </TableCell>
                      </TableRow>
                    ))}
                  </>
                ) : null}
                {groupsQuery.data?.map((g) => (
                  <TableRow key={g.id}>
                    <TableCell className="pl-6 font-medium">{g.name}</TableCell>
                    <TableCell className="text-muted-foreground">{g.description}</TableCell>
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
                              createMutation.reset()
                              updateMutation.reset()
                              setUpserting({
                                mode: "edit",
                                group: g,
                                name: g.name,
                                description: g.description ?? "",
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
                            onClick={() => deleteMutation.mutate(g.id)}
                          >
                            <Trash2 className="mr-2 size-4" />
                            {t("common.delete")}
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </TableCell>
                  </TableRow>
                ))}
                {!groupsQuery.isLoading && groupsQuery.data && groupsQuery.data.length === 0 ? (
                  <TableRow>
                    <TableCell
                      className="pl-6 py-8 text-center text-muted-foreground"
                      colSpan={3}
                    >
                      {t("common.noData")}
                    </TableCell>
                  </TableRow>
                ) : null}
              </TableBody>
            </Table>
          </CardContent>
        </Card>

        <Dialog
          open={!!upserting}
          onOpenChange={(open) => (!open ? setUpserting(null) : null)}
        >
          <DialogContent
            aria-label={
              upserting?.mode === "create"
                ? t("groups.createGroup")
                : t("groups.editGroup")
            }
          >
            <DialogHeader>
              <DialogTitle>
                {upserting?.mode === "create"
                  ? t("groups.createGroup")
                  : t("groups.editGroup")}
              </DialogTitle>
              {upserting?.mode === "edit" ? (
                <DialogDescription>{upserting.group.name}</DialogDescription>
              ) : null}
            </DialogHeader>

            {upserting ? (
              <div className="space-y-4">
                <div className="space-y-1">
                  <Label htmlFor="group-name" className="text-sm text-slate-700">
                    {t("groups.name")}
                  </Label>
                  <Input
                    id="group-name"
                    value={upserting.name}
                    onChange={(e) =>
                      setUpserting((prev) => (prev ? { ...prev, name: e.target.value } : prev))
                    }
                    placeholder={t("groups.namePlaceholder")}
                    autoFocus={upserting.mode === "create"}
                  />
                </div>
                <div className="space-y-1">
                  <Label htmlFor="group-desc" className="text-sm text-slate-700">
                    {t("groups.description")}
                  </Label>
                  <Input
                    id="group-desc"
                    value={upserting.description}
                    onChange={(e) =>
                      setUpserting((prev) =>
                        prev ? { ...prev, description: e.target.value } : prev,
                      )
                    }
                    placeholder={t("groups.descriptionPlaceholder")}
                  />
                </div>

                <div className="text-sm text-amber-700">
                  {createMutation.isError || updateMutation.isError ? (
                    (createMutation.error instanceof ApiError
                      ? createMutation.error.message
                      : updateMutation.error instanceof ApiError
                        ? updateMutation.error.message
                        : t("groups.saveFailed"))
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
                  const description = upserting.description.trim()
                  if (!name) return
                  if (upserting.mode === "create") {
                    createMutation.mutate({ name, description })
                  } else {
                    updateMutation.mutate({ id: upserting.group.id, payload: { name, description } })
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
