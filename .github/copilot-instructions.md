# SBoard Copilot Instructions

Trust this file first. Only search the repo when these instructions are incomplete or you find they are wrong.

## What this repo is

SBoard is a sing-box based subscription and node management platform.

- `panel/`: Go management plane (`gin` + SQLite) plus RPC backend for the web UI
- `web/`: React 19 + Vite 8 + TypeScript frontend
- `node/`: Go data-plane service that receives config from Panel and manages sing-box
- `e2e/`: Playwright + Docker end-to-end suite
- Workspace/tooling: Go workspace + Moon monorepo, JS package manager is Bun

Repo shape: medium monorepo with Go + TypeScript. Version pins live in `/.prototools`:

- `moon = 2.0.3`
- `go = 1.26.0` (modules currently declare `go 1.25.0`)
- `node = 22.22.0`
- `bun = 1.3.9`

## Where to change code

- Panel entrypoint: `/panel/cmd/panel/main.go`
- Node entrypoint: `/node/cmd/node/main.go`
- Panel REST/RPC boundary: `/panel/internal/api`, `/panel/internal/rpc`
- Node REST/RPC boundary: `/node/internal/api`, `/node/internal/rpc`
- RPC spec source of truth: `/panel/proto/sboard/panel/v1/panel.proto`
- Generated RPC code:
  - Go: `/panel/internal/rpc/gen`
  - TS: `/web/src/lib/rpc/gen`
  - Node RPC gen: `/node/internal/rpc/gen`
- Frontend pages: `/web/src/pages`
- Frontend shared code: `/web/src/components`, `/web/src/lib`, `/web/src/store`
- Moon tasks: `/panel/moon.yml`, `/web/moon.yml`, `/node/moon.yml`, `/e2e/moon.yml`
- CI workflows: `/.github/workflows/ci.yml`, `/.github/workflows/docker-publish.yml`

## CI gates to match locally

`ci.yml` runs 5 jobs:

1. `panel:check-generate`
2. `web:lint web:format-check web:typecheck web:test`
3. `panel:test`
4. `node:test`
5. `e2e:test`

If you touch frontend code, also expect format/typecheck/tests to matter. If you touch `panel.proto`, regenerate and re-check generated files before anything else.

## Bootstrap: use this order

Validated on a fresh clone:

1. Install Bun **before** any Moon or frontend command.
  - In this environment `bun` was missing.
  - `npm install -g bun` worked.
2. Install JS dependencies:
  - `cd web && bun install --frozen-lockfile`
  - `cd e2e && bun install --frozen-lockfile`
3. Use direct workspace commands for local validation if Moon fails.

Important local pitfalls that were actually observed:

- `./scripts/moon-cli.sh --version` failed here because Bun's `bunx` fallback could not resolve `@moonrepo/cli@2.0.3`.
- `npm exec --yes --package=@moonrepo/cli@2.0.3 moon -- --version` worked.
- `moon run ...` still failed in this restricted environment because Moon tried to download a WASM plugin (`plugin::offline`). In CI this is handled by `moonrepo/setup-toolchain@v0`.
- `buf` was **not** installed in this environment, so regeneration was not runnable locally without extra setup.

## Fastest validated local validation path

Run these from the repo root or the listed workspace:

- Generated files freshness:
  - `cd panel && bash check-generate.sh`
  - Result: works and prints `Generated files are up to date.`
- Frontend:
  - `cd web && bun run lint`
  - `cd web && bun run format:check`
  - `cd web && bunx tsc -b`  
    No output means success.
  - `cd web && bun run test`
  - `cd web && bun run build`
    Works; Vite warns about large chunks after minification, but build still succeeds.
- Go backends:
  - `cd panel && go test ./... -count=1`
  - `cd node && go test ./... -count=1`

## Build/run details that avoid common mistakes

- For local `go run` of Panel/Node, use `GOFLAGS='-tags=with_utls'`.
- Do **not** assume `/api/health` exists in default runtime mode.
  - `main.go` starts RPC mode, and `router.go` disables most legacy REST routes when RPC handlers are mounted.
  - Validated health probes are:
    - Panel: `POST /rpc/sboard.panel.v1.HealthService/GetHealth` with JSON body `{}` → `{"status":"ok"}`
    - Node: `POST /rpc/sboard.node.v1.NodeControlService/Health` with JSON body `{}` → `{"status":"ok"}`
- Panel local run needs at least:
  - `PANEL_HTTP_ADDR`
  - `PANEL_DB_PATH`
  - `PANEL_JWT_SECRET`
  - `PANEL_NODE_RPC_SCHEME=http` for local plain HTTP node testing
- Node local run needs at least:
  - `NODE_HTTP_ADDR`
  - `NODE_SECRET_KEY`
- If you start services manually in the shell, always clean them up with explicit `kill <PID>`; stale processes caused port conflicts during validation.
- Avoid `go build ./cmd/...` inside `panel/` or `node/` unless you pass `-o /tmp/...`; otherwise local binaries are created in the workspace directories (`panel/panel`, `node/node`).

## Proto / generation workflow

When changing `/panel/proto/sboard/panel/v1/panel.proto`:

1. Ensure `buf` is installed and `web/node_modules` exists.
2. Run generation (`moon run panel:generate` in a normal environment).
3. Re-run freshness check:
  - `cd panel && bash check-generate.sh`

`/panel/internal/rpc/generate.go` shows generation also formats generated TS files with `bunx oxfmt --write`.

## E2E / Docker

- E2E entrypoints:
  - `cd e2e && bash smoke.sh`
  - `cd e2e && bash run.sh`
- Docker itself was available here, but `smoke.sh` failed during Playwright image build because Bun inside Docker could not fetch `@playwright/test` (`HTTPError parsing package manifest`).
- So E2E needs **both** Docker and working outbound package resolution during image build; do not mark E2E as validated if package fetching is restricted.

## Repo-specific conventions worth knowing

- Frontend uses Oxc, not ESLint.
- Prefer Bun for frontend commands.
- Panel admin API is RPC-first; compatibility REST is mainly `GET /api/sub/:user_uuid`.
- `panel.proto` is the single source of truth for panel management APIs.
- Go workspace root is `/go.work`; both `panel` and `node` are active modules.
