import { ConnectError } from "@connectrpc/connect";

export function getErrorMessage(error: unknown, fallback: string): string {
  if (error instanceof ConnectError) {
    return error.rawMessage || error.message || fallback;
  }
  if (error instanceof Error && error.message.trim()) {
    return error.message;
  }
  return fallback;
}
