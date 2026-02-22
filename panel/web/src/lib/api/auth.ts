import "./client";
import {
  login as _login,
  getBootstrapStatus as _getBootstrapStatus,
  bootstrap as _bootstrap,
} from "./gen";
import type { LoginResponse, BootstrapStatus, BootstrapResponse } from "./gen";

export function loginAdmin(payload: {
  username: string;
  password: string;
}): Promise<LoginResponse> {
  return _login({ body: payload }).then((r) => r.data!.data);
}

export function getBootstrapStatus(): Promise<BootstrapStatus> {
  return _getBootstrapStatus().then((r) => r.data!.data);
}

export function bootstrapAdmin(payload: {
  setup_token?: string;
  username: string;
  password: string;
  confirm_password: string;
}): Promise<BootstrapResponse> {
  const { setup_token, ...body } = payload;
  return _bootstrap({
    body,
    headers: setup_token ? { "X-Setup-Token": setup_token } : undefined,
  }).then((r) => r.data!.data);
}
