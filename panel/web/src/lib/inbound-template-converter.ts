import { getInboundProtocolCapability } from "./inbound-template-capability";

export type InboundTemplatePayload = {
  tag: string;
  protocol: string;
  listen_port: number;
  public_port: number;
  settings: Record<string, unknown>;
  tls_settings?: Record<string, unknown>;
  transport_settings?: Record<string, unknown>;
};

export type ParseResult =
  | { ok: true; payload: InboundTemplatePayload; warnings?: string[] }
  | { ok: false; error: string };

export type BuildSingBoxConfigResult =
  | { ok: true; text: string; warnings?: string[] }
  | { ok: false; error: string };

export const extraConfigKey = "__config";

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

const reservedKeys = new Set([
  "type",
  "protocol",
  "tag",
  "listen",
  "port",
  "listen_port",
  "public_port",
  "users",
  "settings",
  "sniffing",
  "streamSettings",
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

function normalizeProtocol(value: string): string {
  return value.trim().toLowerCase();
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

  const rawProtocol =
    readRequiredString(extracted.inbound, "type") ?? readRequiredString(extracted.inbound, "protocol");
  if (!rawProtocol) {
    return { ok: false, error: "type/protocol required" };
  }
  const protocol = normalizeProtocol(rawProtocol);

  const tag = readRequiredString(extracted.inbound, "tag");
  if (!tag) {
    return { ok: false, error: "tag required" };
  }

  const listenPort = readPort(extracted.inbound.listen_port) ?? readPort(extracted.inbound.port);
  if (listenPort == null) {
    return { ok: false, error: "listen_port/port required and must be > 0" };
  }

  const warnings: string[] = [];
  const capability = getInboundProtocolCapability(protocol);
  if (capability.level !== "supported") {
    warnings.push(
      `protocol \"${protocol}\" compatibility warning: ${capability.reason ?? "unknown reason"}`,
    );
  }

  const settings: Record<string, unknown> = {};
  for (const [key, value] of Object.entries(extracted.inbound)) {
    if (reservedKeys.has(key)) continue;
    settings[key] = value;
  }

  const usersField = asArray(extracted.inbound.users);
  if (protocol === "vless" && usersField && usersField.length > 0) {
    const firstUser = asRecord(usersField[0]);
    const flowFromUser = firstUser ? readRequiredString(firstUser, "flow") : null;
    if (flowFromUser) {
      settings.flow = flowFromUser;
      warnings.push("users is partially mapped: users[0].flow -> flow");
    }
  }

  const nestedSettings = asRecord(extracted.inbound.settings);
  if (nestedSettings) {
    for (const [key, value] of Object.entries(nestedSettings)) {
      if (key === "clients") {
        if (protocol === "vless") {
          const clients = asArray(value) ?? [];
          const firstClient = asRecord(clients[0]);
          const flow = firstClient ? readRequiredString(firstClient, "flow") : null;
          if (flow) {
            settings.flow = flow;
            warnings.push("settings.clients is partially mapped: clients[0].flow -> flow");
          } else {
            warnings.push("template field settings.clients is ignored by converter");
          }
        } else {
          warnings.push("template field settings.clients is ignored by converter");
        }
        continue;
      }
      if (key === "decryption") {
        warnings.push("template field settings.decryption is ignored by converter");
        continue;
      }
      settings[key] = value;
    }
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

  if (protocol === "vless") {
    const sniffing = asRecord(extracted.inbound.sniffing);
    if (sniffing) {
      warnings.push("template field sniffing is ignored by converter");
    }

    const streamSettings = asRecord(extracted.inbound.streamSettings);
    if (streamSettings) {
      const security = readRequiredString(streamSettings, "security");
      if (security === "reality") {
        const realitySettings = asRecord(streamSettings.realitySettings);
        if (realitySettings) {
          const serverNames = (asArray(realitySettings.serverNames) ?? [])
            .map((item) => (typeof item === "string" ? item.trim() : ""))
            .filter(Boolean);
          const shortIds = (asArray(realitySettings.shortIds) ?? [])
            .map((item) => (typeof item === "string" ? item.trim() : ""))
            .filter(Boolean);

          let handshake: Record<string, unknown> | undefined;
          const target = readRequiredString(realitySettings, "target");
          if (target) {
            const idx = target.lastIndexOf(":");
            if (idx > 0 && idx < target.length - 1) {
              const host = target.slice(0, idx).trim();
              const portText = target.slice(idx + 1).trim();
              const port = Number(portText);
              if (host && Number.isInteger(port) && port > 0 && port <= 65535) {
                handshake = { server: host, server_port: port };
              } else {
                warnings.push("streamSettings.realitySettings.target is invalid; handshake mapping skipped");
              }
            } else {
              warnings.push("streamSettings.realitySettings.target should be host:port");
            }
          }

          const reality: Record<string, unknown> = { enabled: true };
          if (handshake) {
            reality.handshake = handshake;
          }
          const privateKey = readRequiredString(realitySettings, "privateKey");
          if (privateKey) {
            reality.private_key = privateKey;
          }
          if (shortIds.length > 0) {
            reality.short_id = shortIds;
          }

          const tlsFromReality = asRecord(tlsSettings) ?? {};
          tlsFromReality.enabled = true;
          if (serverNames.length > 0) {
            tlsFromReality.server_name = serverNames[0];
          }
          tlsFromReality.reality = reality;
          tlsSettings = tlsFromReality;
        }
      }
    }
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
    ...(warnings.length > 0 ? { warnings } : {}),
  };
}

function shouldSkipSettingKeyInInbound(
  protocol: string,
  key: string,
  mode: "strict" | "template",
): boolean {
  if (key === extraConfigKey) return true;
  if (protocol === "vless" && key === "flow") {
    // `flow` should not appear at inbound top-level in strict sing-box config.
    return mode === "strict";
  }
  return reservedKeys.has(key);
}

function toSingBoxInboundFromPayload(
  payload: InboundTemplatePayload,
  mode: "strict" | "template",
): Record<string, unknown> {
  const inboundTemplate: Record<string, unknown> = {
    type: payload.protocol,
    tag: payload.tag,
    listen_port: payload.listen_port,
    users: [],
  };

  const flow = typeof payload.settings?.flow === "string" ? payload.settings.flow.trim() : "";
  if (mode === "template" && payload.protocol === "vless" && flow) {
    inboundTemplate.settings = { flow };
  }

  for (const [key, value] of Object.entries(payload.settings ?? {})) {
    if (shouldSkipSettingKeyInInbound(payload.protocol, key, mode)) continue;
    inboundTemplate[key] = value;
  }

  const tls = asRecord(payload.tls_settings);
  if (tls) {
    inboundTemplate.tls = tls;
  }

  const transport = asRecord(payload.transport_settings);
  if (transport) {
    inboundTemplate.transport = transport;
  }

  return inboundTemplate;
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

export function buildSingBoxConfigTextFromTemplate(input: string): BuildSingBoxConfigResult {
  const parsed = parseInboundTemplateToPayload(input);
  if (!parsed.ok) {
    return parsed;
  }

  const extraConfig = asRecord(parsed.payload.settings?.[extraConfigKey]);
  const inboundTemplate = toSingBoxInboundFromPayload(parsed.payload, "strict");
  const full = buildFullConfigTemplate(inboundTemplate, extraConfig ?? undefined);
  return {
    ok: true,
    text: JSON.stringify(full, null, 2),
    ...(parsed.warnings ? { warnings: parsed.warnings } : {}),
  };
}

export function buildSingBoxTemplateTextFromPayload(payload: InboundTemplatePayload): string {
  const extraConfig = asRecord(payload.settings?.[extraConfigKey]);
  const inboundTemplate = toSingBoxInboundFromPayload(payload, "template");
  const full = buildFullConfigTemplate(inboundTemplate, extraConfig ?? undefined);
  return JSON.stringify(full, null, 2);
}

export function readTemplateProtocolFromText(input: string): string | null {
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

  const protocol = readRequiredString(inbound, "type") ?? readRequiredString(inbound, "protocol");
  if (!protocol) return null;
  return normalizeProtocol(protocol);
}
