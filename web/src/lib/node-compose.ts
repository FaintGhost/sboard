export type BuildNodeComposeInput = {
  port: number;
  secretKey: string;
  logLevel?: string;
  image?: string;
  containerName?: string;
};

export function buildNodeDockerCompose(input: BuildNodeComposeInput): string {
  const port =
    Number.isFinite(input.port) && input.port >= 1 && input.port <= 65535
      ? Math.trunc(input.port)
      : 3003;
  const secretKey = input.secretKey;
  const logLevel = input.logLevel?.trim() || "info";
  const image = input.image?.trim() || "faintghost/sboard-node:latest";
  const containerName = input.containerName?.trim() || "sboard-node";

  return `services:
  ${containerName}:
    container_name: ${containerName}
    image: ${image}
    restart: unless-stopped
    network_mode: host
    volumes:
      - ./data:/data
    environment:
      NODE_HTTP_ADDR: ":${port}"
      NODE_SECRET_KEY: "${escapeDoubleQuoted(secretKey)}"
      NODE_LOG_LEVEL: "${escapeDoubleQuoted(logLevel)}"
`;
}

function escapeDoubleQuoted(v: string): string {
  // YAML double-quoted string escaping: keep it minimal.
  return v.replaceAll("\\", "\\\\").replaceAll('"', '\\"').replaceAll("\n", "");
}

export function generateNodeSecretKey(bytes: number = 32): string {
  const n = Math.max(16, Math.min(64, Math.trunc(bytes)));
  const buf = new Uint8Array(n);
  crypto.getRandomValues(buf);
  return base64UrlEncode(buf);
}

function base64UrlEncode(buf: Uint8Array): string {
  let s = "";
  for (let i = 0; i < buf.length; i++) s += String.fromCharCode(buf[i]);
  // btoa expects latin1 string.
  const b64 = btoa(s);
  return b64.replaceAll("+", "-").replaceAll("/", "_").replaceAll("=", "");
}
