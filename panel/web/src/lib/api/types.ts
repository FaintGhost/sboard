export type LoginRequest = {
  username: string
  password: string
}

export type LoginResponse = {
  token: string
  expires_at: string
}

export type UserStatus = "active" | "disabled" | "expired" | "traffic_exceeded"

export type User = {
  id: number
  uuid: string
  username: string
  traffic_limit: number
  traffic_used: number
  traffic_reset_day: number
  expire_at: string | null
  status: UserStatus
}

export type ListUsersParams = {
  limit?: number
  offset?: number
  status?: UserStatus
}
