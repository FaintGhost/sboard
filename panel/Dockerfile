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
COPY go.mod ./

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
COPY . .

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
