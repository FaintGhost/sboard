### Task 4: Docker Build Adaptation

**目标：** 更新 panel Dockerfile 和所有 compose 文件的构建上下文，适配 web 目录提升。

**Files:**
- Modify: `panel/Dockerfile`
- Modify: `panel/docker-compose.build.yml`
- Modify: `e2e/docker-compose.e2e.yml`
- Modify: `.github/workflows/docker-publish.yml`

**Step 1: 更新 panel/Dockerfile**

构建上下文从 `panel/` 改为工作区根。关键变更行：

1. web 阶段不变（`COPY web/...` 现在指向顶层 `web/`，因为上下文变了）
2. builder 阶段 `COPY go.mod` → `COPY panel/go.mod`
3. builder 阶段 `COPY . .` → `COPY panel/ .`

完整 Dockerfile：

```dockerfile
# syntax=docker/dockerfile:1.7

FROM oven/bun:1.3.9-alpine AS web

WORKDIR /web
COPY web/package.json web/bun.lock ./
RUN --mount=type=cache,target=/root/.bun/install/cache \
  bun install --frozen-lockfile
COPY web ./
RUN --mount=type=cache,target=/root/.bun/install/cache \
  bun run build

FROM golang:1.25 AS builder

ARG PANEL_VERSION=unknown
ARG PANEL_COMMIT_ID=unknown
ARG SING_BOX_VERSION=unknown
ARG GO_BUILD_TAGS=with_utls

WORKDIR /src
COPY panel/go.mod ./

# Panel uses github.com/mattn/go-sqlite3 which requires CGO.
RUN apt-get update && apt-get install -y --no-install-recommends \
  gcc \
  libc6-dev \
  libsqlite3-dev \
  && rm -rf /var/lib/apt/lists/*

# The repo does not necessarily ship with go.sum (fresh clone / early stage),
# so we must allow the image build to generate a complete go.sum on the fly.
# Some environments also set GOFLAGS=-mod=readonly; use -mod=mod explicitly.
RUN --mount=type=cache,target=/go/pkg/mod \
  go mod download
COPY panel/ .

# `go mod tidy` does not accept `-mod=...`. Also avoid inheriting GOFLAGS=-mod=readonly.
RUN --mount=type=cache,target=/go/pkg/mod \
  --mount=type=cache,target=/root/.cache/go-build \
  GOFLAGS= go mod tidy
RUN --mount=type=cache,target=/go/pkg/mod \
  --mount=type=cache,target=/root/.cache/go-build \
  CGO_ENABLED=1 go build -mod=mod -tags "${GO_BUILD_TAGS}" -trimpath -ldflags "-s -w \
  -X sboard/panel/internal/buildinfo.PanelVersion=${PANEL_VERSION} \
  -X sboard/panel/internal/buildinfo.PanelCommitID=${PANEL_COMMIT_ID} \
  -X sboard/panel/internal/buildinfo.SingBoxVersion=${SING_BOX_VERSION}" -o /out/sboard-panel ./cmd/panel

FROM debian:stable-slim

WORKDIR /app
RUN apt-get update && apt-get install -y --no-install-recommends \
  ca-certificates \
  libsqlite3-0 \
  && rm -rf /var/lib/apt/lists/*
COPY --from=builder /out/sboard-panel /app/sboard-panel
# The binary expects migrations at `internal/db/migrations`.
COPY --from=builder /src/internal/db/migrations /app/internal/db/migrations
COPY --from=web /web/dist /app/web/dist

ENV PANEL_HTTP_ADDR=:8080
ENV PANEL_DB_PATH=/data/panel.db
ENV PANEL_SERVE_WEB=true
ENV PANEL_WEB_DIR=/app/web/dist
EXPOSE 8080
ENTRYPOINT ["/app/sboard-panel"]
```

**Step 2: 更新 panel/docker-compose.build.yml**

构建上下文从 `.`（panel/）改为 `..`（工作区根）：

```yaml
services:
  sboard-panel:
    build:
      context: ..
      dockerfile: panel/Dockerfile
    image: "${SBOARD_PANEL_IMAGE:-faintghost/sboard-panel:local}"
```

**Step 3: 更新 e2e/docker-compose.e2e.yml**

panel 服务的构建上下文从 `../panel` 改为 `..`：

```yaml
  panel:
    build:
      context: ..
      dockerfile: panel/Dockerfile
```

其余服务（node、probe、sb-client、playwright）不变。

**Step 4: 更新 .github/workflows/docker-publish.yml**

panel 构建的 context 从 `./panel` 改为 `.`：

```yaml
      - name: Build and push panel
        uses: docker/build-push-action@v6
        with:
          context: .
          file: ./panel/Dockerfile
```

node 构建不变（`context: ./node`）。

**Step 5: 验证 Docker 构建（panel）**

```bash
cd panel && docker compose -f docker-compose.yml -f docker-compose.build.yml build && cd ..
```

Expected: 构建成功，镜像包含 Go 二进制和 web/dist。

**Step 6: 提交**

```bash
git add panel/Dockerfile panel/docker-compose.build.yml e2e/docker-compose.e2e.yml .github/workflows/docker-publish.yml
git commit -m "refactor(docker): adapt build context for top-level web directory"
```
