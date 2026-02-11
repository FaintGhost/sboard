export type InboundSupportLevel = "supported" | "convertible" | "unsupported";

export type InboundProtocolCapability = {
  protocol: string;
  level: InboundSupportLevel;
  reason?: string;
};

const capabilityTable: Record<string, InboundProtocolCapability> = {
  vless: { protocol: "vless", level: "supported" },
  vmess: { protocol: "vmess", level: "supported" },
  trojan: { protocol: "trojan", level: "supported" },
  shadowsocks: { protocol: "shadowsocks", level: "supported" },
  socks: { protocol: "socks", level: "supported" },
  http: { protocol: "http", level: "supported" },
  mixed: { protocol: "mixed", level: "supported" },
};

export function getInboundProtocolCapability(protocol: string): InboundProtocolCapability {
  const normalized = protocol.trim().toLowerCase();
  if (!normalized) {
    return { protocol: normalized, level: "unsupported", reason: "missing protocol" };
  }
  return (
    capabilityTable[normalized] ?? {
      protocol: normalized,
      level: "unsupported",
      reason: "not in current sing-box inbound support matrix",
    }
  );
}
