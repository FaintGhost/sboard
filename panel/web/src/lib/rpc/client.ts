import { callUnaryMethod } from "@connectrpc/connect-query";
import type {
  DescMessage,
  DescMethodUnary,
  MessageInitShape,
  MessageShape,
} from "@bufbuild/protobuf";

import { rpcTransport } from "@/lib/rpc/transport";

export function rpcCall<I extends DescMessage, O extends DescMessage>(
  method: DescMethodUnary<I, O>,
  input: MessageInitShape<I> | undefined,
): Promise<MessageShape<O>> {
  return callUnaryMethod(rpcTransport, method, input);
}

export function i64(n: number | null | undefined): bigint | undefined {
  if (n === null || n === undefined) return undefined;
  return BigInt(Math.trunc(n));
}

export function n64(v: bigint | number | null | undefined): number {
  if (v === null || v === undefined) return 0;
  return typeof v === "bigint" ? Number(v) : v;
}
