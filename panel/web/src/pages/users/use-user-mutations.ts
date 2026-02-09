import { useMutation, useQueryClient } from "@tanstack/react-query"

import { createUser, deleteUser, disableUser, updateUser } from "@/lib/api/users"
import { putUserGroups } from "@/lib/api/user-groups"

export function useUserMutations() {
  const qc = useQueryClient()

  const createMutation = useMutation({
    mutationFn: async (input: { username: string; groupIDs: number[] }) => {
      const created = await createUser({ username: input.username })
      if (input.groupIDs.length > 0) {
        await putUserGroups(created.id, { group_ids: input.groupIDs })
      }
      return created
    },
    onSuccess: async () => {
      await qc.invalidateQueries({ queryKey: ["users"] })
    },
  })

  const updateMutation = useMutation({
    mutationFn: (input: { id: number; payload: Record<string, unknown> }) =>
      updateUser(input.id, input.payload),
    onSuccess: async () => {
      await qc.invalidateQueries({ queryKey: ["users"] })
    },
  })

  const saveGroupsMutation = useMutation({
    mutationFn: (input: { userId: number; groupIDs: number[] }) =>
      putUserGroups(input.userId, { group_ids: input.groupIDs }),
    onSuccess: async () => {
      await qc.invalidateQueries({ queryKey: ["user-groups"] })
      await qc.invalidateQueries({ queryKey: ["users"] })
    },
  })

  const disableMutation = useMutation({
    mutationFn: disableUser,
    onSuccess: async () => {
      await qc.invalidateQueries({ queryKey: ["users"] })
    },
  })

  const deleteMutation = useMutation({
    mutationFn: deleteUser,
    onSuccess: async () => {
      await qc.invalidateQueries({ queryKey: ["users"] })
    },
  })

  return {
    createMutation,
    updateMutation,
    saveGroupsMutation,
    disableMutation,
    deleteMutation,
  }
}
