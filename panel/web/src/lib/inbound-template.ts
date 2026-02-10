import type { Inbound } from "@/lib/api/types";

export type InboundTemplatePresetProtocol = "vless" | "vmess" | "trojan" | "shadowsocks";

type InboundTemplatePayload = {
  tag: string;
  protocol: string;
  listen_port: number;
  public_port: number;
  settings: Record<string, unknown>;
  tls_settings?: Record<string, unknown>;
  transport_settings?: Record<string, unknown>;
};

type ParseResult = { ok: true; payload: InboundTemplatePayload } | { ok: false; error: string };

const extraConfigKey = "__config";

const fullConfigKeys = [
  "$schema",
  "log",
  "dns",
  "ntp",
  "certificate",
  "endpoints",
  "outbounds",
  "route",
  "services",
  "experimental",
] as const;

export const inboundTemplatePresetProtocols: InboundTemplatePresetProtocol[] = [
  "vless",
  "vmess",
  "trojan",
  "shadowsocks",
];

const inboundTemplatePresets: Record<InboundTemplatePresetProtocol, Record<string, unknown>> = {
  vless: {
    type: "vless",
    tag: "vless-in",
    listen_port: 443,
    users: [],
    flow: "xtls-rprx-vision",
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
    method: "2022-blake3-aes-128-gcm",
    password: "8JCsPssfgS8tiRwiMlhARg==",
    users: [],
  },
};

const reservedKeys = new Set([
  "type",
  "tag",
  "listen",
  "listen_port",
  "public_port",
  "users",
  "tls",
  "transport",
]);

function asRecord(value: unknown): Record<string, unknown> | null {
  if (!value || typeof value !== "object" || Array.isArray(value)) return null;
  return value as Record<string, unknown>;
}

function asArray(value: unknown): unknown[] | null {
  if (!Array.isArray(value)) return null;
  return value;
}

function shouldKeepExtraConfigValue(value: unknown): boolean {
  if (value == null) return false;
  if (Array.isArray(value)) return value.length > 0;
  if (typeof value === "object") return Object.keys(value as Record<string, unknown>).length > 0;
  return true;
}

function readRequiredString(obj: Record<string, unknown>, key: string): string | null {
  const value = obj[key];
  if (typeof value !== "string") return null;
  const trimmed = value.trim();
  return trimmed ? trimmed : null;
}

function readPort(value: unknown): number | null {
  if (typeof value !== "number" || !Number.isFinite(value)) return null;
  if (!Number.isInteger(value)) return null;
  return value > 0 ? value : null;
}

function extractInboundAndExtraConfig(
  template: Record<string, unknown>,
): { inbound: Record<string, unknown>; extraConfig?: Record<string, unknown> } | { error: string } {
  const maybeInbounds = template.inbounds;
  if (maybeInbounds == null) {
    return { inbound: template };
  }

  const inbounds = asArray(maybeInbounds);
  if (!inbounds) {
    return { error: "inbounds must be an array" };
  }
  if (inbounds.length === 0) {
    return { error: "inbounds requires at least one item" };
  }
  if (inbounds.length > 1) {
    return { error: "only one inbound is supported in template" };
  }

  const inbound = asRecord(inbounds[0]);
  if (!inbound) {
    return { error: "inbounds[0] must be an object" };
  }

  const extraConfig: Record<string, unknown> = {};
  for (const key of fullConfigKeys) {
    if (template[key] !== undefined) {
      extraConfig[key] = template[key];
    }
  }

  return { inbound, extraConfig };
}

function buildFullConfigTemplate(
  inbound: Record<string, unknown>,
  extraConfig?: Record<string, unknown>,
): Record<string, unknown> {
  const full: Record<string, unknown> = {
    inbounds: [inbound],
  };

  if (extraConfig) {
    for (const key of fullConfigKeys) {
      const value = extraConfig[key];
      if (shouldKeepExtraConfigValue(value)) {
        full[key] = extraConfig[key];
      }
    }
  }

  return full;
}

export function parseInboundTemplateToPayload(input: string): ParseResult {
  let raw: unknown;
  try {
    raw = JSON.parse(input);
  } catch {
    return { ok: false, error: "invalid JSON" };
  }

  const template = asRecord(raw);
  if (!template) {
    return { ok: false, error: "template must be a JSON object" };
  }

  const extracted = extractInboundAndExtraConfig(template);
  if ("error" in extracted) {
    return { ok: false, error: extracted.error };
  }

  const protocol = readRequiredString(extracted.inbound, "type");
  if (!protocol) {
    return { ok: false, error: "type required" };
  }

  const tag = readRequiredString(extracted.inbound, "tag");
  if (!tag) {
    return { ok: false, error: "tag required" };
  }

  const listenPort = readPort(extracted.inbound.listen_port);
  if (listenPort == null) {
    return { ok: false, error: "listen_port required and must be > 0" };
  }

  const settings: Record<string, unknown> = {};
  for (const [key, value] of Object.entries(extracted.inbound)) {
    if (reservedKeys.has(key)) continue;
    settings[key] = value;
  }
  if (extracted.extraConfig) {
    settings[extraConfigKey] = extracted.extraConfig;
  }

  let tlsSettings: Record<string, unknown> | undefined;
  if (extracted.inbound.tls != null) {
    const tls = asRecord(extracted.inbound.tls);
    if (!tls) {
      return { ok: false, error: "tls must be an object" };
    }
    tlsSettings = tls;
  }

  let transportSettings: Record<string, unknown> | undefined;
  if (extracted.inbound.transport != null) {
    const transport = asRecord(extracted.inbound.transport);
    if (!transport) {
      return { ok: false, error: "transport must be an object" };
    }
    transportSettings = transport;
  }

  return {
    ok: true,
    payload: {
      tag,
      protocol,
      listen_port: listenPort,
      public_port: 0,
      settings,
      tls_settings: tlsSettings,
      transport_settings: transportSettings,
    },
  };
}

export function readTemplateProtocol(input: string): InboundTemplatePresetProtocol | null {
  let raw: unknown;
  try {
    raw = JSON.parse(input);
  } catch {
    return null;
  }
  const root = asRecord(raw);
  if (!root) return null;

  let inbound = root;
  const inbounds = asArray(root.inbounds);
  if (inbounds && inbounds.length > 0) {
    const first = asRecord(inbounds[0]);
    if (first) {
      inbound = first;
    }
  }

  const protocol = readRequiredString(inbound, "type");
  if (!protocol) return null;
  return inboundTemplatePresetProtocols.find((item) => item === protocol) ?? null;
}

export function buildPresetInboundTemplateText(protocol: InboundTemplatePresetProtocol): string {
  const full = buildFullConfigTemplate(inboundTemplatePresets[protocol]);
  return JSON.stringify(full, null, 2);
}

export function buildInboundTemplateText(inbound: Inbound): string {
  const settings = asRecord(inbound.settings) ?? {};
  const extraConfig = asRecord(settings[extraConfigKey]);

  const inboundTemplate: Record<string, unknown> = {
    type: inbound.protocol,
    tag: inbound.tag,
    listen_port: inbound.listen_port,
    users: [],
  };

  for (const [key, value] of Object.entries(settings)) {
    if (key === extraConfigKey) continue;
    if (reservedKeys.has(key)) continue;
    inboundTemplate[key] = value;
  }

  const tls = asRecord(inbound.tls_settings);
  if (tls) {
    inboundTemplate.tls = tls;
  }

  const transport = asRecord(inbound.transport_settings);
  if (transport) {
    inboundTemplate.transport = transport;
  }

  const full = buildFullConfigTemplate(inboundTemplate, extraConfig ?? undefined);
  return JSON.stringify(full, null, 2);
}
