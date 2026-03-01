import { describe, expect, it } from "vitest";

import { buildNodeDockerCompose } from "./node-compose";

describe("buildNodeDockerCompose", () => {
  it("renders docker-compose.yml for node", () => {
    expect(
      buildNodeDockerCompose({
        port: 3003,
        secretKey: "69186918",
        logLevel: "info",
      }),
    ).toBe(`services:
  sboard-node:
    container_name: sboard-node
    image: faintghost/sboard-node:latest
    restart: unless-stopped
    network_mode: host
    volumes:
      - ./data:/data
    environment:
      NODE_HTTP_ADDR: ":3003"
      NODE_SECRET_KEY: "69186918"
      NODE_LOG_LEVEL: "info"
`);
  });

  it("includes PANEL_URL and NODE_UUID when provided", () => {
    expect(
      buildNodeDockerCompose({
        port: 3003,
        secretKey: "69186918",
        logLevel: "info",
        panelUrl: "https://panel.example.com",
        nodeUuid: "550e8400-e29b-41d4-a716-446655440000",
      }),
    ).toBe(`services:
  sboard-node:
    container_name: sboard-node
    image: faintghost/sboard-node:latest
    restart: unless-stopped
    network_mode: host
    volumes:
      - ./data:/data
    environment:
      NODE_HTTP_ADDR: ":3003"
      NODE_SECRET_KEY: "69186918"
      NODE_LOG_LEVEL: "info"
      PANEL_URL: "https://panel.example.com"
      NODE_UUID: "550e8400-e29b-41d4-a716-446655440000"
`);
  });

  it("omits PANEL_URL when panelUrl is empty", () => {
    expect(
      buildNodeDockerCompose({
        port: 3003,
        secretKey: "69186918",
        panelUrl: "",
        nodeUuid: "550e8400-e29b-41d4-a716-446655440000",
      }),
    ).toBe(`services:
  sboard-node:
    container_name: sboard-node
    image: faintghost/sboard-node:latest
    restart: unless-stopped
    network_mode: host
    volumes:
      - ./data:/data
    environment:
      NODE_HTTP_ADDR: ":3003"
      NODE_SECRET_KEY: "69186918"
      NODE_LOG_LEVEL: "info"
      NODE_UUID: "550e8400-e29b-41d4-a716-446655440000"
`);
  });

  it("omits PANEL_URL and NODE_UUID when both are undefined", () => {
    expect(
      buildNodeDockerCompose({
        port: 3003,
        secretKey: "69186918",
      }),
    ).toBe(`services:
  sboard-node:
    container_name: sboard-node
    image: faintghost/sboard-node:latest
    restart: unless-stopped
    network_mode: host
    volumes:
      - ./data:/data
    environment:
      NODE_HTTP_ADDR: ":3003"
      NODE_SECRET_KEY: "69186918"
      NODE_LOG_LEVEL: "info"
`);
  });

  it("escapes special characters in panelUrl and nodeUuid", () => {
    expect(
      buildNodeDockerCompose({
        port: 3003,
        secretKey: "key",
        panelUrl: 'https://example.com/path?"q=1',
        nodeUuid: 'uuid-with-"quote',
      }),
    ).toBe(`services:
  sboard-node:
    container_name: sboard-node
    image: faintghost/sboard-node:latest
    restart: unless-stopped
    network_mode: host
    volumes:
      - ./data:/data
    environment:
      NODE_HTTP_ADDR: ":3003"
      NODE_SECRET_KEY: "key"
      NODE_LOG_LEVEL: "info"
      PANEL_URL: "https://example.com/path?\\"q=1"
      NODE_UUID: "uuid-with-\\"quote"
`);
  });
});
