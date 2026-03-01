import { toApiError } from "@/lib/api/client";
import { rpcCall } from "@/lib/rpc/client";
import {
  getAdminProfile as getAdminProfileRPC,
  updateAdminProfile as updateAdminProfileRPC,
} from "@/lib/rpc/gen/sboard/panel/v1/panel-AuthService_connectquery";
import {
  getSystemInfo as getSystemInfoSystemRPC,
  getSystemSettings as getSystemSettingsSystemRPC,
  updateSystemSettings as updateSystemSettingsSystemRPC,
} from "@/lib/rpc/gen/sboard/panel/v1/panel-SystemService_connectquery";
import { toAdminProfile, toSystemInfo, toSystemSettings } from "@/lib/rpc/mappers";
import type { AdminProfile, SystemInfo, SystemSettings } from "./types";
import type { UpdateAdminProfilePayload } from "./types";

export function getSystemInfo(): Promise<SystemInfo> {
  return rpcCall(getSystemInfoSystemRPC, {})
    .then((r) => toSystemInfo(r.data!))
    .catch((e) => {
      throw toApiError(e);
    });
}

export function getSystemSettings(): Promise<SystemSettings> {
  return rpcCall(getSystemSettingsSystemRPC, {})
    .then((r) => toSystemSettings(r.data!))
    .catch((e) => {
      throw toApiError(e);
    });
}

export function updateSystemSettings(payload: SystemSettings): Promise<SystemSettings> {
  return rpcCall(updateSystemSettingsSystemRPC, {
    subscriptionBaseUrl: payload.subscription_base_url,
    timezone: payload.timezone,
  })
    .then((r) => toSystemSettings(r.data!))
    .catch((e) => {
      throw toApiError(e);
    });
}

export function getAdminProfile(): Promise<AdminProfile> {
  return rpcCall(getAdminProfileRPC, {})
    .then((r) => toAdminProfile(r.data!))
    .catch((e) => {
      throw toApiError(e);
    });
}

export function updateAdminProfile(payload: UpdateAdminProfilePayload): Promise<AdminProfile> {
  return rpcCall(updateAdminProfileRPC, {
    oldPassword: payload.old_password,
    newUsername: payload.new_username,
    newPassword: payload.new_password,
    confirmPassword: payload.confirm_password,
  })
    .then((r) => toAdminProfile(r.data!))
    .catch((e) => {
      throw toApiError(e);
    });
}
