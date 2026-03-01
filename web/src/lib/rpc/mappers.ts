import type {
  AdminProfile,
  BootstrapResponse,
  BootstrapStatus,
  Group,
  Node,
  NodeTrafficSample,
  SystemInfo,
  SystemSettings,
  TrafficNodeSummary,
  TrafficTimeseriesPoint,
  TrafficTotalSummary,
  User,
} from "@/lib/rpc/types";
import type {
  AdminProfile as RpcAdminProfile,
  BootstrapResponse as RpcBootstrapResponse,
  BootstrapStatus as RpcBootstrapStatus,
  Group as RpcGroup,
  GroupUsersListItem,
  LoginResponseEnvelope,
  Node as RpcNode,
  NodeTrafficSample as RpcNodeTrafficSample,
  SystemInfo as RpcSystemInfo,
  SystemSettings as RpcSystemSettings,
  TrafficNodeSummary as RpcTrafficNodeSummary,
  TrafficTimeseriesPoint as RpcTrafficTimeseriesPoint,
  TrafficTotalSummary as RpcTrafficTotalSummary,
  User as RpcUser,
} from "@/lib/rpc/gen/sboard/panel/v1/panel_pb";

function n64(v: bigint | number | null | undefined): number {
  if (v === null || v === undefined) return 0;
  return typeof v === "bigint" ? Number(v) : v;
}

export function toUser(v: RpcUser): User {
  return {
    id: n64(v.id),
    uuid: v.uuid,
    username: v.username,
    group_ids: v.groupIds.map((x) => n64(x)),
    traffic_limit: n64(v.trafficLimit),
    traffic_used: n64(v.trafficUsed),
    traffic_reset_day: v.trafficResetDay,
    expire_at: v.expireAt ?? null,
    status: v.status,
  } as User;
}

export function toGroup(v: RpcGroup): Group {
  return {
    id: n64(v.id),
    name: v.name,
    description: v.description,
    member_count: n64(v.memberCount),
  } as Group;
}

export function toNode(v: RpcNode): Node {
  return {
    id: n64(v.id),
    uuid: v.uuid,
    name: v.name,
    api_address: v.apiAddress,
    api_port: v.apiPort,
    secret_key: v.secretKey,
    public_address: v.publicAddress,
    group_id: v.groupId === undefined ? null : n64(v.groupId),
    status: v.status,
    last_seen_at: v.lastSeenAt ?? null,
  } as Node;
}

export function toNodeTrafficSample(v: RpcNodeTrafficSample): NodeTrafficSample {
  return {
    id: n64(v.id),
    inbound_tag: v.inboundTag,
    upload: n64(v.upload),
    download: n64(v.download),
    recorded_at: v.recordedAt,
  } as NodeTrafficSample;
}

export function toSystemInfo(v: RpcSystemInfo): SystemInfo {
  return {
    panel_version: v.panelVersion,
    panel_commit_id: v.panelCommitId,
    sing_box_version: v.singBoxVersion,
  } as SystemInfo;
}

export function toSystemSettings(v: RpcSystemSettings): SystemSettings {
  return {
    subscription_base_url: v.subscriptionBaseUrl,
    timezone: v.timezone,
  } as SystemSettings;
}

export function toTrafficNodeSummary(v: RpcTrafficNodeSummary): TrafficNodeSummary {
  return {
    node_id: n64(v.nodeId),
    upload: n64(v.upload),
    download: n64(v.download),
    last_recorded_at: v.lastRecordedAt,
    samples: n64(v.samples),
    inbounds: n64(v.inbounds),
  } as TrafficNodeSummary;
}

export function toTrafficTotalSummary(v: RpcTrafficTotalSummary): TrafficTotalSummary {
  return {
    upload: n64(v.upload),
    download: n64(v.download),
    last_recorded_at: v.lastRecordedAt,
    samples: n64(v.samples),
    nodes: n64(v.nodes),
    inbounds: n64(v.inbounds),
  } as TrafficTotalSummary;
}

export function toTrafficTimeseriesPoint(v: RpcTrafficTimeseriesPoint): TrafficTimeseriesPoint {
  return {
    bucket_start: v.bucketStart,
    upload: n64(v.upload),
    download: n64(v.download),
  } as TrafficTimeseriesPoint;
}

export function toAdminProfile(v: RpcAdminProfile): AdminProfile {
  return { username: v.username } as AdminProfile;
}

export function toBootstrapStatus(v: RpcBootstrapStatus): BootstrapStatus {
  return { needs_setup: v.needsSetup } as BootstrapStatus;
}

export function toBootstrapResponse(v: RpcBootstrapResponse): BootstrapResponse {
  return { ok: v.data?.ok ?? false } as BootstrapResponse;
}

export function toLoginResponse(v: LoginResponseEnvelope): { token: string; expires_at: string } {
  return {
    token: v.data?.token ?? "",
    expires_at: v.data?.expiresAt ?? "",
  };
}

export function toGroupUsersListItem(v: GroupUsersListItem): {
  id: number;
  uuid: string;
  username: string;
  traffic_limit: number;
  traffic_used: number;
  status: string;
} {
  return {
    id: n64(v.id),
    uuid: v.uuid,
    username: v.username,
    traffic_limit: n64(v.trafficLimit),
    traffic_used: n64(v.trafficUsed),
    status: v.status,
  };
}
