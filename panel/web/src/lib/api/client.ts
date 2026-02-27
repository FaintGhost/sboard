import { Code, ConnectError } from "@connectrpc/connect";

export class ApiError extends Error {
  status: number;

  constructor(status: number, message: string) {
    super(message);
    this.name = "ApiError";
    this.status = status;
  }
}

function statusFromCode(code: Code): number {
  switch (code) {
    case Code.Unauthenticated:
      return 401;
    case Code.NotFound:
      return 404;
    case Code.InvalidArgument:
      return 400;
    case Code.FailedPrecondition:
      return 409;
    case Code.Unavailable:
      return 502;
    default:
      return 500;
  }
}

export function toApiError(error: unknown): ApiError {
  if (error instanceof ApiError) return error;
  if (error instanceof ConnectError) {
    return new ApiError(statusFromCode(error.code), error.rawMessage || error.message);
  }
  if (error instanceof Error) {
    return new ApiError(500, error.message);
  }
  return new ApiError(500, "request failed");
}
