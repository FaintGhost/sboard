import { toApiError } from "@/lib/api/client";
import { rpcCall } from "@/lib/rpc/client";
import {
  checkSingBox as checkSingBoxRPC,
  formatSingBox as formatSingBoxRPC,
  generateSingBox as generateSingBoxRPC,
} from "@/lib/rpc/gen/sboard/panel/v1/panel-SingBoxToolService_connectquery";
import type {
  SingBoxToolMode,
  SingBoxFormatResponse,
  SingBoxCheckResponse,
  SingBoxGenerateCommand,
  SingBoxGenerateResponse,
} from "./types";

export function formatSingBoxConfig(payload: {
  config: string;
  mode?: SingBoxToolMode;
}): Promise<SingBoxFormatResponse> {
  return rpcCall(formatSingBoxRPC, {
    data: {
      config: payload.config,
      mode: payload.mode,
    },
  })
    .then((r) => ({ formatted: r.data?.formatted ?? "" }))
    .catch((e) => {
      throw toApiError(e);
    });
}

export function checkSingBoxConfig(payload: {
  config: string;
  mode?: SingBoxToolMode;
}): Promise<SingBoxCheckResponse> {
  return rpcCall(checkSingBoxRPC, {
    data: {
      config: payload.config,
      mode: payload.mode,
    },
  })
    .then((r) => ({ ok: !!r.data?.ok, output: r.data?.output ?? "" }))
    .catch((e) => {
      throw toApiError(e);
    });
}

export function generateSingBoxValue(
  command: SingBoxGenerateCommand,
): Promise<SingBoxGenerateResponse> {
  return rpcCall(generateSingBoxRPC, { command })
    .then((r) => ({ output: r.data?.output ?? "" }))
    .catch((e) => {
      throw toApiError(e);
    });
}
