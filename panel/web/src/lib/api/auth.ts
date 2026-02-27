import { rpcCall } from "@/lib/rpc/client";
import { toApiError } from "@/lib/api/client";
import {
  login as loginRPC,
  getBootstrapStatus as getBootstrapStatusRPC,
  bootstrap as bootstrapRPC,
} from "@/lib/rpc/gen/sboard/panel/v1/panel-AuthService_connectquery";
import type { BootstrapResponse, BootstrapStatus, LoginResponse } from "./types";
import { toBootstrapResponse, toBootstrapStatus, toLoginResponse } from "@/lib/rpc/mappers";

export function loginAdmin(payload: {
  username: string;
  password: string;
}): Promise<LoginResponse> {
  return rpcCall(loginRPC, payload)
    .then((r) => toLoginResponse(r))
    .catch((e) => {
      throw toApiError(e);
    });
}

export function getBootstrapStatus(): Promise<BootstrapStatus> {
  return rpcCall(getBootstrapStatusRPC, {})
    .then((r) => toBootstrapStatus(r.data!))
    .catch((e) => {
      throw toApiError(e);
    });
}

export function bootstrapAdmin(payload: {
  setup_token?: string;
  username: string;
  password: string;
  confirm_password: string;
}): Promise<BootstrapResponse> {
  const req = {
    setupToken: payload.setup_token,
    xSetupToken: payload.setup_token,
    username: payload.username,
    password: payload.password,
    confirmPassword: payload.confirm_password,
  };
  return rpcCall(bootstrapRPC, req)
    .then((r) => toBootstrapResponse(r))
    .catch((e) => {
      throw toApiError(e);
    });
}
