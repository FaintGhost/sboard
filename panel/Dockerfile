# syntax=docker/dockerfile:1.7

FROM node:22-alpine AS web

WORKDIR /web
COPY web/package.json web/package-lock.json ./
RUN --mount=type=cache,target=/root/.npm \
  npm ci
COPY web ./
RUN --mount=type=cache,target=/root/.npm \
  npm run build

FROM golang:1.25 AS builder

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
  CGO_ENABLED=1 go build -mod=mod -trimpath -ldflags "-s -w" -o /out/sboard-panel ./cmd/panel

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

ARG PANEL_VERSION=unknown
ARG PANEL_COMMIT_ID=unknown
ARG SING_BOX_VERSION=unknown

ENV PANEL_HTTP_ADDR=:8080
ENV PANEL_DB_PATH=/data/panel.db
ENV PANEL_SERVE_WEB=true
ENV PANEL_WEB_DIR=/app/web/dist
ENV PANEL_VERSION=${PANEL_VERSION}
ENV PANEL_COMMIT_ID=${PANEL_COMMIT_ID}
ENV SING_BOX_VERSION=${SING_BOX_VERSION}

EXPOSE 8080
ENTRYPOINT ["/app/sboard-panel"]
