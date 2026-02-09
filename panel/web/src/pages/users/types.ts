import type { User, UserStatus } from "@/lib/api/types"

export type StatusFilter = UserStatus | "all"

export type EditState = {
  mode: "create" | "edit"
  user: User
  username: string
  status: UserStatus
  trafficLimit: string
  trafficResetDay: number
  expireDate: Date | null
  clearExpireAt: boolean
  groupIDs: number[]
  groupsLoadedFromServer: boolean
}

export const defaultNewUser: User = {
  id: 0,
  uuid: "",
  username: "",
  group_ids: [],
  traffic_limit: 0,
  traffic_used: 0,
  traffic_reset_day: 0,
  expire_at: null,
  status: "active",
}
