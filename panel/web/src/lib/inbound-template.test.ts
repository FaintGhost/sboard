import { describe, expect, it } from "vitest";

import type { Inbound } from "@/lib/api/types";

import {
  buildInboundTemplateText,
  buildPresetInboundTemplateText,
  buildSingBoxConfigTextFromTemplate,
  buildTemplateTextFromPayload,
  parseInboundTemplateToPayload,
  readTemplateProtocol,
} from "./inbound-template";

describe("inbound template", () => {
  it("parses full config template and strips users", () => {
    const input = `{
  "$schema": "https://sing-box.sagernet.org/schema/config.json",
  "log": {},
  "dns": {},
  "inbounds": [
    {
      "type": "shadowsocks",
      "tag": "ss-in",
      "listen_port": 8388,
      "method": "2022-blake3-aes-256-gcm",
      "password": "8JCsPssfgS8tiRwiMlhARg==",
      "users": []
    }
  ],
  "outbounds": [],
  "route": {}
}`;

    const out = parseInboundTemplateToPayload(input);
    expect(out.ok).toBe(true);
    if (!out.ok) return;

    expect(out.payload.protocol).toBe("shadowsocks");
    expect(out.payload.tag).toBe("ss-in");
    expect(out.payload.listen_port).toBe(8388);
    expect(out.payload.public_port).toBe(0);
    expect(out.payload.settings).toMatchObject({
      method: "2022-blake3-aes-256-gcm",
      password: "8JCsPssfgS8tiRwiMlhARg==",
      __config: {
        $schema: "https://sing-box.sagernet.org/schema/config.json",
        log: {},
        dns: {},
        outbounds: [],
        route: {},
      },
    });
  });

  it("supports legacy single inbound object template", () => {
    const input = `{
  "type": "vless",
  "tag": "vless-in",
  "listen_port": 443,
  "flow": "xtls-rprx-vision",
  "tls": { "enabled": true },
  "transport": { "type": "ws" },
  "users": []
}`;

    const out = parseInboundTemplateToPayload(input);
    expect(out.ok).toBe(true);
    if (!out.ok) return;

    expect(out.payload.protocol).toBe("vless");
    expect(out.payload.tag).toBe("vless-in");
    expect(out.payload.listen_port).toBe(443);
    expect(out.payload.settings).toEqual({ flow: "xtls-rprx-vision" });
    expect(out.payload.tls_settings).toEqual({ enabled: true });
    expect(out.payload.transport_settings).toEqual({ type: "ws" });
  });

  it("reads vless flow from sing-box users field", () => {
    const input = `{
  "inbounds": [
    {
      "type": "vless",
      "tag": "vless-in",
      "listen_port": 443,
      "users": [
        {
          "name": "u1",
          "uuid": "bf000d23-0752-40b4-affe-68f7707a9661",
          "flow": "xtls-rprx-vision"
        }
      ]
    }
  ]
}`;

    const out = parseInboundTemplateToPayload(input);
    expect(out.ok).toBe(true);
    if (!out.ok) return;

    expect(out.payload.settings).toMatchObject({
      flow: "xtls-rprx-vision",
    });
  });

  it("parses xray-style vless tcp reality template", () => {
    const input = `{
  "inbounds": [
    {
      "tag": "VLESS_TCP_REALITY",
      "port": 443,
      "protocol": "vless",
      "settings": {
        "clients": [
          { "id": "6fd47678-1f45-48f1-8051-fdbcdc2a3ccb", "flow": "xtls-rprx-vision" }
        ],
        "decryption": "none"
      },
      "sniffing": {
        "enabled": true,
        "destOverride": ["http", "tls", "quic"]
      },
      "streamSettings": {
        "network": "raw",
        "security": "reality",
        "realitySettings": {
          "show": false,
          "xver": 0,
          "target": "aws.amazon.com:443",
          "shortIds": ["f02ea9614ec73e47"],
          "privateKey": "0IFd0ciWOFIU0bDhZTF01PPi7OvY9PkuJMGg1ZgY45s",
          "serverNames": ["aws.amazon.com"]
        }
      }
    }
  ]
}`;

    const out = parseInboundTemplateToPayload(input);
    expect(out.ok).toBe(true);
    if (!out.ok) return;

    expect(out.payload.protocol).toBe("vless");
    expect(out.payload.tag).toBe("VLESS_TCP_REALITY");
    expect(out.payload.listen_port).toBe(443);
    expect(out.payload.settings).toMatchObject({
      flow: "xtls-rprx-vision",
    });
    expect(out.payload.settings).not.toHaveProperty("sniff");
    expect(out.payload.settings).not.toHaveProperty("sniff_override_destination");
    expect(out.payload.tls_settings).toMatchObject({
      enabled: true,
      server_name: "aws.amazon.com",
      reality: {
        enabled: true,
        private_key: "0IFd0ciWOFIU0bDhZTF01PPi7OvY9PkuJMGg1ZgY45s",
        short_id: ["f02ea9614ec73e47"],
        handshake: {
          server: "aws.amazon.com",
          server_port: 443,
        },
      },
    });

    expect((out.warnings ?? []).join("\n")).toContain("sniffing is ignored");

    const normalized = buildSingBoxConfigTextFromTemplate(input);
    expect(normalized.ok).toBe(true);
    if (!normalized.ok) return;

    const normalizedObj = JSON.parse(normalized.text) as Record<string, unknown>;
    const inbounds = normalizedObj.inbounds as Array<Record<string, unknown>>;
    expect(inbounds[0].type).toBe("vless");
    expect(inbounds[0].listen_port).toBe(443);
    expect(inbounds[0].protocol).toBeUndefined();
    expect(inbounds[0].port).toBeUndefined();
    expect(inbounds[0].flow).toBeUndefined();
  });

  it("returns error when multiple inbounds are provided", () => {
    const input = `{
  "inbounds": [
    {"type":"vless","tag":"a","listen_port":443},
    {"type":"vmess","tag":"b","listen_port":8443}
  ]
}`;

    const out = parseInboundTemplateToPayload(input);
    expect(out.ok).toBe(false);
    if (out.ok) return;
    expect(out.error).toContain("only one inbound");
  });

  it("builds editable full config template from inbound", () => {
    const inbound: Inbound = {
      id: 1,
      uuid: "u-1",
      tag: "ss-in",
      node_id: 2,
      protocol: "shadowsocks",
      listen_port: 8388,
      public_port: 8388,
      settings: {
        method: "2022-blake3-aes-256-gcm",
        password: "abc",
        __config: {
          $schema: "https://sing-box.sagernet.org/schema/config.json",
          outbounds: [{ type: "direct", tag: "direct" }],
          route: { final: "direct" },
        },
      },
      tls_settings: null,
      transport_settings: null,
    };

    const text = buildInboundTemplateText(inbound);
    const parsed = JSON.parse(text) as Record<string, unknown>;

    expect(Array.isArray(parsed.inbounds)).toBe(true);
    const inbounds = parsed.inbounds as Array<Record<string, unknown>>;
    expect(inbounds).toHaveLength(1);
    expect(inbounds[0].type).toBe("shadowsocks");
    expect(inbounds[0].tag).toBe("ss-in");
    expect(inbounds[0].listen_port).toBe(8388);
    expect(inbounds[0].users).toEqual([]);
    expect(inbounds[0].method).toBe("2022-blake3-aes-256-gcm");
    expect(parsed.$schema).toBe("https://sing-box.sagernet.org/schema/config.json");
    expect(parsed.route).toEqual({ final: "direct" });
  });

  it("reads protocol from preset template", () => {
    const text = buildPresetInboundTemplateText("trojan");
    const parsed = JSON.parse(text) as Record<string, unknown>;

    expect(Object.keys(parsed)).toEqual(["inbounds"]);
    expect(Array.isArray(parsed.inbounds)).toBe(true);
    expect(readTemplateProtocol(text)).toBe("trojan");
  });

  it("supports socks/http/mixed preset templates", () => {
    const socks = buildPresetInboundTemplateText("socks");
    const http = buildPresetInboundTemplateText("http");
    const mixed = buildPresetInboundTemplateText("mixed");

    expect(readTemplateProtocol(socks)).toBe("socks");
    expect(readTemplateProtocol(http)).toBe("http");
    expect(readTemplateProtocol(mixed)).toBe("mixed");

    const parsed = parseInboundTemplateToPayload(socks);
    expect(parsed.ok).toBe(true);
    if (!parsed.ok) return;
    expect(parsed.payload.protocol).toBe("socks");
    expect(parsed.payload.listen_port).toBe(1080);
  });

  it("uses vless tcp reality preset and keeps vmess preset basic", () => {
    const vless = buildPresetInboundTemplateText("vless");
    const vmess = buildPresetInboundTemplateText("vmess");

    const parsedVless = parseInboundTemplateToPayload(vless);
    expect(parsedVless.ok).toBe(true);
    if (parsedVless.ok) {
      expect(parsedVless.payload.protocol).toBe("vless");
      expect(parsedVless.payload.tag).toBe("VLESS_TCP_REALITY");
      expect(parsedVless.payload.settings).toMatchObject({
        flow: "xtls-rprx-vision",
      });
      expect(parsedVless.payload.tls_settings).toMatchObject({
        enabled: true,
        reality: { enabled: true },
      });
    }

    const parsedVmess = parseInboundTemplateToPayload(vmess);
    expect(parsedVmess.ok).toBe(true);
    if (parsedVmess.ok) {
      expect(parsedVmess.payload.protocol).toBe("vmess");
    }
  });

  it("keeps vless flow in editable template text", () => {
    const parsed = parseInboundTemplateToPayload(`{
  "inbounds": [
    {
      "type": "vless",
      "tag": "vless-in",
      "listen_port": 443,
      "settings": {
        "flow": "xtls-rprx-vision"
      }
    }
  ]
}`);
    expect(parsed.ok).toBe(true);
    if (!parsed.ok) return;

    const text = buildTemplateTextFromPayload(parsed.payload);
    const root = JSON.parse(text) as Record<string, unknown>;
    const inbounds = root.inbounds as Array<Record<string, unknown>>;
    const settings = inbounds[0].settings as Record<string, unknown>;
    expect(settings.flow).toBe("xtls-rprx-vision");
  });

  it("returns compatibility warnings for unsupported protocol and ignored fields", () => {
    const out = parseInboundTemplateToPayload(`{
  "inbounds": [
    {
      "type": "dokodemo-door",
      "tag": "x",
      "listen_port": 1234,
      "settings": {
        "clients": [],
        "decryption": "none"
      }
    }
  ]
}`);
    expect(out.ok).toBe(true);
    if (!out.ok) return;
    expect(out.warnings?.length ?? 0).toBeGreaterThan(0);
    expect((out.warnings ?? []).join("\n")).toContain("compatibility warning");
    expect((out.warnings ?? []).join("\n")).toContain("settings.clients");
  });

  it("returns error when required fields missing", () => {
    const out = parseInboundTemplateToPayload(`{"inbounds":[{"tag":"x"}]}`);
    expect(out.ok).toBe(false);
    if (out.ok) return;
    expect(out.error).toContain("type");
  });
});
