import "./client";
import {
  getSystemInfo as _getSystemInfo,
  getSystemSettings as _getSystemSettings,
  updateSystemSettings as _updateSystemSettings,
  getAdminProfile as _getAdminProfile,
  updateAdminProfile as _updateAdminProfile,
} from "./gen";
import type { SystemInfo, SystemSettings, AdminProfile } from "./gen";
import type { UpdateAdminProfilePayload } from "./types";

export function getSystemInfo(): Promise<SystemInfo> {
  return _getSystemInfo().then((r) => r.data!.data);
}

export function getSystemSettings(): Promise<SystemSettings> {
  return _getSystemSettings().then((r) => r.data!.data);
}

export function updateSystemSettings(payload: SystemSettings): Promise<SystemSettings> {
  return _updateSystemSettings({ body: payload }).then((r) => r.data!.data);
}

export function getAdminProfile(): Promise<AdminProfile> {
  return _getAdminProfile().then((r) => r.data!.data);
}

export function updateAdminProfile(payload: UpdateAdminProfilePayload): Promise<AdminProfile> {
  return _updateAdminProfile({
    body: {
      old_password: payload.old_password,
      new_username: payload.new_username,
      new_password: payload.new_password,
      confirm_password: payload.confirm_password,
    },
  }).then((r) => r.data!.data);
}
