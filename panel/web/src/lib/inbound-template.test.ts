import { describe, expect, it } from "vitest";

import type { Inbound } from "@/lib/api/types";

import {
  buildInboundTemplateText,
  buildPresetInboundTemplateText,
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
      "method": "2022-blake3-aes-128-gcm",
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
      method: "2022-blake3-aes-128-gcm",
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
        method: "2022-blake3-aes-128-gcm",
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
    expect(inbounds[0].method).toBe("2022-blake3-aes-128-gcm");
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

  it("returns error when required fields missing", () => {
    const out = parseInboundTemplateToPayload(`{"inbounds":[{"tag":"x"}]}`);
    expect(out.ok).toBe(false);
    if (out.ok) return;
    expect(out.error).toContain("type");
  });
});
