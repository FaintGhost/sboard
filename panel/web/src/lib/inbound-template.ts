import type { Inbound } from "@/lib/rpc/types";

import {
  buildSingBoxConfigTextFromTemplate as convertTemplateToSingBoxConfig,
  buildSingBoxTemplateTextFromPayload,
  extraConfigKey,
  parseInboundTemplateToPayload as parseTemplateToPayload,
  readTemplateProtocolFromText,
  type BuildSingBoxConfigResult,
  type InboundTemplatePayload,
  type ParseResult,
} from "./inbound-template-converter";

export type InboundTemplatePresetProtocol =
  | "vless"
  | "vmess"
  | "trojan"
  | "shadowsocks"
  | "socks"
  | "http"
  | "mixed";

const inboundTemplatePresets: Record<InboundTemplatePresetProtocol, Record<string, unknown>> = {
  vless: {
    tag: "VLESS_TCP_REALITY",
    port: 443,
    protocol: "vless",
    settings: {
      clients: [{ flow: "xtls-rprx-vision" }],
      decryption: "none",
    },
    sniffing: {
      enabled: true,
      destOverride: ["http", "tls", "quic"],
    },
    streamSettings: {
      network: "raw",
      security: "reality",
      realitySettings: {
        show: false,
        xver: 0,
        target: "aws.amazon.com:443",
        shortIds: ["f02ea9614ec73e47"],
        privateKey: "replace-with-your-private-key",
        serverNames: ["aws.amazon.com"],
      },
    },
  },
  vmess: {
    type: "vmess",
    tag: "vmess-in",
    listen_port: 443,
    users: [],
  },
  trojan: {
    type: "trojan",
    tag: "trojan-in",
    listen_port: 443,
    users: [],
  },
  shadowsocks: {
    type: "shadowsocks",
    tag: "ss-in",
    listen_port: 8388,
    method: "2022-blake3-aes-256-gcm",
    password: "8JCsPssfgS8tiRwiMlhARg==",
    users: [],
  },
  socks: {
    type: "socks",
    tag: "socks-in",
    listen_port: 1080,
    users: [],
  },
  http: {
    type: "http",
    tag: "http-in",
    listen_port: 8080,
    users: [],
  },
  mixed: {
    type: "mixed",
    tag: "mixed-in",
    listen_port: 2080,
    users: [],
  },
};

export const inboundTemplatePresetProtocols: InboundTemplatePresetProtocol[] = [
  "vless",
  "vmess",
  "trojan",
  "shadowsocks",
  "socks",
  "http",
  "mixed",
];

function asRecord(value: unknown): Record<string, unknown> | null {
  if (!value || typeof value !== "object" || Array.isArray(value)) return null;
  return value as Record<string, unknown>;
}

export function parseInboundTemplateToPayload(input: string): ParseResult {
  return parseTemplateToPayload(input);
}

export function readTemplateProtocol(input: string): InboundTemplatePresetProtocol | null {
  const protocol = readTemplateProtocolFromText(input);
  if (!protocol) return null;
  return inboundTemplatePresetProtocols.find((item) => item === protocol) ?? null;
}

export function buildPresetInboundTemplateText(protocol: InboundTemplatePresetProtocol): string {
  const full: Record<string, unknown> = {
    inbounds: [inboundTemplatePresets[protocol]],
  };
  return JSON.stringify(full, null, 2);
}

export function buildSingBoxConfigTextFromTemplate(input: string): BuildSingBoxConfigResult {
  return convertTemplateToSingBoxConfig(input);
}

export function buildTemplateTextFromPayload(payload: InboundTemplatePayload): string {
  return buildSingBoxTemplateTextFromPayload(payload);
}

export function buildInboundTemplateText(inbound: Inbound): string {
  const settings = asRecord(inbound.settings) ?? {};
  const tlsSettings = asRecord(inbound.tls_settings);
  const transportSettings = asRecord(inbound.transport_settings);
  const payload: InboundTemplatePayload = {
    tag: inbound.tag,
    protocol: inbound.protocol,
    listen_port: inbound.listen_port,
    public_port: inbound.public_port,
    settings,
    ...(tlsSettings ? { tls_settings: tlsSettings } : {}),
    ...(transportSettings ? { transport_settings: transportSettings } : {}),
  };
  return buildSingBoxTemplateTextFromPayload(payload);
}

export { extraConfigKey };
