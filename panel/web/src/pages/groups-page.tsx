import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useMemo, useState } from "react"
import { useTranslation } from "react-i18next"

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
        <header className="space-y-1">
          <h1 className="text-2xl font-semibold tracking-tight text-slate-900">
            {t("groups.title")}
          </h1>
          <p className="text-sm text-slate-500">
            {t("groups.hint")}
          </p>
        </header>

        <div className="flex items-center justify-between gap-3">
          <div className="text-sm text-slate-600">
            {groupsQuery.isLoading ? t("common.loading") : null}
            {groupsQuery.isError ? t("common.loadFailed") : null}
            {groupsQuery.data ? t("groups.count", { count: groupsQuery.data.length }) : null}
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
        </div>

        <div className="overflow-hidden rounded-xl border border-slate-200">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="px-4">{t("common.name")}</TableHead>
                <TableHead className="px-4">{t("common.description")}</TableHead>
                <TableHead className="px-4">{t("common.actions")}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {groupsQuery.data?.map((g) => (
                <TableRow key={g.id}>
                  <TableCell className="px-4 font-medium text-slate-900">
                    {g.name}
                  </TableCell>
                  <TableCell className="px-4 text-slate-700">
                    {g.description}
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
                            group: g,
                            name: g.name,
                            description: g.description ?? "",
                          })
                        }}
                      >
                        {t("common.edit")}
                      </Button>
                      <Button
                        size="sm"
                        variant="destructive"
                        disabled={deleteMutation.isPending}
                        onClick={() => deleteMutation.mutate(g.id)}
                      >
                        {t("common.delete")}
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
              {groupsQuery.data && groupsQuery.data.length === 0 ? (
                <TableRow>
                  <TableCell className="px-4 py-6 text-slate-500" colSpan={3}>
                    {t("common.noData")}
                  </TableCell>
                </TableRow>
              ) : null}
            </TableBody>
          </Table>
        </div>

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
