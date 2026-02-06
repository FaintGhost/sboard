export type LoginRequest = {
  username: string
  password: string
}

export type LoginResponse = {
  token: string
  expires_at: string
}

export type BootstrapStatus = {
  needs_setup: boolean
}

export type BootstrapRequest = {
  setup_token: string
  username: string
  password: string
  confirm_password: string
}

export type BootstrapResponse = {
  ok: boolean
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

export type Group = {
  id: number
  name: string
  description: string
}

export type ListGroupsParams = {
  limit?: number
  offset?: number
}

export type UserGroups = {
  group_ids: number[]
}

export type Node = {
  id: number
  uuid: string
  name: string
  api_address: string
  api_port: number
  secret_key: string
  public_address: string
  group_id: number | null
  status: string
}

export type ListNodesParams = {
  limit?: number
  offset?: number
}

export type Inbound = {
  id: number
  uuid: string
  tag: string
  node_id: number
  protocol: string
  listen_port: number
  public_port: number
  settings: unknown
  tls_settings: unknown | null
  transport_settings: unknown | null
}

export type ListInboundsParams = {
  limit?: number
  offset?: number
  node_id?: number
}
