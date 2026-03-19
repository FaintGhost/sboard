# SBoard

[中文](./README.zh.md)

SBoard is a sing-box based subscription and node management platform.

## Architecture

- **Panel**: management plane (Go/Gin + SQLite + React web UI)
- **Node**: data plane service (embedded sing-box), receives configs from Panel

## Features

- Full admin UI: auth, dashboard, users, groups, nodes, inbounds, subscriptions, settings
- Group-driven delivery: users get node inbounds through group assignment
- Protocols: vless, vmess, trojan, shadowsocks (including 2022), socks, http, mixed
- Subscription formats: sing-box JSON and v2ray (Base64), with User-Agent auto selection
- Inbound template editing with TLS/Reality/Transport support
- Auto node sync after inbound create/update/delete
- RPC-first API (Connect + Protobuf) for Panel frontend/backend integration

## Prerequisites

- Go 1.25+
- Bun (frontend build)
- Linux server (recommended)

## Quick Start (Bare Metal)

1) Start Panel

```bash
cd panel
PANEL_HTTP_ADDR=:8080 \
PANEL_DB_PATH=panel.db \
PANEL_CORS_ALLOW_ORIGINS=http://localhost:5173 \
PANEL_NODE_RPC_SCHEME=http \
PANEL_JWT_SECRET=change-me-in-prod \
GOFLAGS='-tags=with_utls' \
go run ./cmd/panel
```

2) Start Node

```bash
cd node
NODE_HTTP_ADDR=:3000 \
NODE_SECRET_KEY=secret \
NODE_LOG_LEVEL=info \
GOFLAGS='-tags=with_utls' \
go run ./cmd/node
```

3) Start frontend (dev)

```bash
cd web
VITE_API_BASE_URL=http://127.0.0.1:8080 bun run dev
```

Notes:

- Admin APIs use `/rpc/*` by default.
- `/api/*` is reserved for subscription compatibility (`/api/sub/:user_uuid`) and Node-side REST.
- Panel connects to Node over HTTPS by default; set `PANEL_NODE_RPC_SCHEME=http` explicitly for local/dev plain HTTP nodes.

If you still prefer Vite proxy:

```bash
VITE_PROXY_TARGET=http://127.0.0.1:8080 bun run dev
```

4) Open `http://localhost:5173` and complete onboarding (create initial admin).

## Docker Deployment (Recommended)

### Panel

```bash
cd panel
export PANEL_JWT_SECRET='replace-with-strong-random'
docker compose up -d
```

Notes:

- Default image: `faintghost/sboard-panel:latest` (override with `SBOARD_PANEL_IMAGE`)
- Panel serves the built web UI (`PANEL_SERVE_WEB=true`)
- SQLite path is `/data/panel.db` (mapped to `panel/data/`)

### Node

- Do not maintain Node compose manually.
- Recommended flow: add node in Panel first, then copy the generated node `docker-compose.yml` from the Panel UI and run `docker compose up -d` on the node host.

## `PANEL_JWT_SECRET` Explained

`PANEL_JWT_SECRET` is used for admin JWT signing and verification:

- Panel signs JWT on successful admin login
- Protected HTTP/RPC endpoints verify JWT with the same secret
- Changing this secret invalidates all existing tokens
- Empty value fails Panel config validation

### Quick ways to generate a strong secret

Use any of these:

```bash
# Recommended (openssl)
openssl rand -base64 48

# Generic Linux
head -c 48 /dev/urandom | base64

# If pwgen is available
pwgen -s 64 1
```

Recommendations:

- At least 32 bytes of randomness (48-64 chars recommended)
- Never use placeholder values in production

## Build Your Own Images

```bash
# Build and push both images
./scripts/docker-build-push.sh --namespace faintghost --tag latest

# Or build separately
cd node
SBOARD_NODE_IMAGE="faintghost/sboard-node:latest" \
  docker compose -f docker-compose.yml -f docker-compose.build.yml build

cd ../panel
SBOARD_PANEL_IMAGE="faintghost/sboard-panel:latest" \
  docker compose -f docker-compose.yml -f docker-compose.build.yml build
```

## GitHub Actions: Auto Publish Images

Workflow: `.github/workflows/docker-publish.yml`

- Triggers:
  - push to `main`/`master`
  - tag push (`v*`)
  - manual run (`workflow_dispatch`)
- Targets:
  - Docker Hub: `docker.io/<DOCKERHUB_USERNAME>/sboard-node`, `docker.io/<DOCKERHUB_USERNAME>/sboard-panel`
  - GHCR: `ghcr.io/<repo_owner>/sboard-node`, `ghcr.io/<repo_owner>/sboard-panel`
- Default tags: `branch`, `tag`, `sha`, `latest` (default branch only)

Required repository secrets:

- `DOCKERHUB_USERNAME`
- `DOCKERHUB_TOKEN` (Docker Hub access token recommended)

Note:

- GHCR uses `GITHUB_TOKEN`, no extra token is required.

## Configuration

### Panel

| Env | Description | Default |
|-----|-------------|---------|
| `PANEL_HTTP_ADDR` | listen address | `:8080` |
| `PANEL_DB_PATH` | SQLite path | `panel.db` |
| `PANEL_JWT_SECRET` | JWT signing secret (required) | - |
| `PANEL_SETUP_TOKEN` | bootstrap token (optional) | auto-generated |
| `PANEL_SERVE_WEB` | serve built web assets | `false` |
| `PANEL_WEB_DIR` | web build directory | `web/dist` |
| `PANEL_CORS_ALLOW_ORIGINS` | allowed origins | `http://localhost:5173` |
| `PANEL_NODE_RPC_SCHEME` | default scheme for Panel -> Node RPC (`https` or `http`) | `https` |
| `PANEL_LOG_REQUESTS` | log HTTP requests | `true` |

### Node

| Env | Description | Default |
|-----|-------------|---------|
| `NODE_HTTP_ADDR` | listen address | `:3000` |
| `NODE_SECRET_KEY` | API secret | - |
| `NODE_LOG_LEVEL` | log level | `info` |
| `PANEL_URL` | Panel base URL for heartbeat; `https` is assumed when scheme is omitted | - |

## API

Panel admin APIs are RPC-first (Connect, path `/rpc/*`).

### Panel admin (RPC)

- Entry: `POST /rpc/sboard.panel.v1.<Service>/<Method>`
- Common: `HealthService`, `AuthService`, `SystemService`
- Business: `UserService`, `GroupService`, `NodeService`, `InboundService`
- Ops: `TrafficService`, `SyncJobService`, `SingBoxToolService`

### Compatibility REST

- Subscription: `GET /api/sub/:user_uuid`

Subscription behavior:

- `?format=singbox`: sing-box JSON
- `?format=v2ray`: v2ray-style Base64 links
- without `format`: auto-selected by User-Agent

## Development

Project structure:

```text
panel/               # Panel backend
  cmd/panel/         # entrypoint
  internal/rpc/      # RPC implementation and generated code
  proto/             # Protobuf contracts
  internal/api/      # HTTP compatibility layer (includes subscription endpoint)
web/                 # React frontend (independent moon project)
node/                # Node service
  cmd/node/          # entrypoint
  internal/api/      # Node HTTP API
  internal/sync/     # sing-box config parse/validate
  internal/core/     # sing-box runtime management
docs/                # design and planning docs
.moon/               # Moon workspace config
.prototools          # version pinning (moon/go/node/bun)
```

RPC proto workflow:

```bash
# Generate Go + TS code
moon run panel:generate

# Ensure generated code is in sync with spec
moon run panel:check-generate
```

Frontend quality tools (Oxc):

```bash
cd web
bun run lint
bun run format:check
bunx tsc -b
bun run test
```

Delivery gates:

- `moon run web:lint`
- `moon run web:format-check`
- `moon run web:typecheck`
- `moon run web:test`
- `moon run panel:check-generate`
- `moon run e2e:test`
