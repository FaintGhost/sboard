// Re-export model types from generated OpenAPI types.
// Utility types not directly generated are defined here.

export type {
  User,
  Group,
  Node,
  NodeTrafficSample,
  Inbound,
  SyncJobListItem as SyncJob,
  SyncAttempt,
  SyncJobDetail,
  LoginRequest,
  LoginResponse,
  BootstrapStatus,
  BootstrapRequest,
  BootstrapResponse,
  SystemInfo,
  SystemSettings,
  AdminProfile,
  UpdateAdminProfileRequest as UpdateAdminProfilePayload,
} from "./gen";

export type UserStatus = "active" | "disabled" | "expired" | "traffic_exceeded";

export type SyncJobStatus = "queued" | "running" | "success" | "failed";

export type SingBoxToolMode = "auto" | "inbound" | "config";

export type SingBoxGenerateCommand =
  | "uuid"
  | "reality-keypair"
  | "wg-keypair"
  | "vapid-keypair"
  | "rand-base64-16"
  | "rand-base64-32";

export type SingBoxFormatResponse = {
  formatted: string;
};

export type SingBoxCheckResponse = {
  ok: boolean;
  output: string;
};

export type SingBoxGenerateResponse = {
  output: string;
};

export type ListUsersParams = {
  limit?: number;
  offset?: number;
  status?: UserStatus;
};

export type ListGroupsParams = {
  limit?: number;
  offset?: number;
};

export type ListNodesParams = {
  limit?: number;
  offset?: number;
};

export type ListInboundsParams = {
  limit?: number;
  offset?: number;
  node_id?: number;
};

export type ListSyncJobsParams = {
  limit?: number;
  offset?: number;
  node_id?: number;
  status?: SyncJobStatus;
  trigger_source?: string;
  from?: string;
  to?: string;
};

export type UserGroups = {
  group_ids: number[];
};
