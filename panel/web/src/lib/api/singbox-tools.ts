import "./client";
import {
  formatSingBox as _formatSingBox,
  checkSingBox as _checkSingBox,
  generateSingBox as _generateSingBox,
} from "./gen";
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
  return _formatSingBox({ body: payload }).then((r) => r.data!.data);
}

export function checkSingBoxConfig(payload: {
  config: string;
  mode?: SingBoxToolMode;
}): Promise<SingBoxCheckResponse> {
  return _checkSingBox({ body: payload }).then((r) => r.data!.data);
}

export function generateSingBoxValue(
  command: SingBoxGenerateCommand,
): Promise<SingBoxGenerateResponse> {
  return _generateSingBox({ body: { command } }).then((r) => r.data!.data);
}
