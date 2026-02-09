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

export type SystemInfo = {
  panel_version: string
  panel_commit_id: string
  sing_box_version: string
}

export type UserStatus = "active" | "disabled" | "expired" | "traffic_exceeded"

export type SystemSettings = {
  subscription_base_url: string
}

export type User = {
  id: number
  uuid: string
  username: string
  group_ids: number[]
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
  member_count: number
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
  last_seen_at?: string | null
}

export type ListNodesParams = {
  limit?: number
  offset?: number
}

export type NodeTrafficSample = {
  id: number
  inbound_tag: string | null
  upload: number
  download: number
  recorded_at: string
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

export type SyncJobStatus = "queued" | "running" | "success" | "failed"

export type SyncJob = {
  id: number
  node_id: number
  parent_job_id?: number | null
  trigger_source: string
  status: SyncJobStatus | string
  inbound_count: number
  active_user_count: number
  payload_hash: string
  attempt_count: number
  duration_ms: number
  error_summary: string
  created_at: string
  started_at?: string
  finished_at?: string
}

export type SyncAttempt = {
  id: number
  attempt_no: number
  status: SyncJobStatus | string
  http_status: number
  duration_ms: number
  error_summary: string
  backoff_ms: number
  started_at: string
  finished_at?: string
}

export type SyncJobDetail = {
  job: SyncJob
  attempts: SyncAttempt[]
}

export type ListSyncJobsParams = {
  limit?: number
  offset?: number
  node_id?: number
  status?: SyncJobStatus
  trigger_source?: string
  from?: string
  to?: string
}

export type SingBoxToolMode = "auto" | "inbound" | "config"

export type SingBoxFormatResponse = {
  formatted: string
}

export type SingBoxCheckResponse = {
  ok: boolean
  output: string
}

export type SingBoxGenerateCommand =
  | "uuid"
  | "reality-keypair"
  | "wg-keypair"
  | "vapid-keypair"
  | "rand-base64-16"
  | "rand-base64-32"

export type SingBoxGenerateResponse = {
  output: string
}
