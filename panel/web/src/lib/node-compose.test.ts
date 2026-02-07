import { describe, expect, it } from "vitest"

import { buildNodeDockerCompose } from "./node-compose"

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
      NODE_STATE_PATH: "/data/last_sync.json"
`)
  })
})
