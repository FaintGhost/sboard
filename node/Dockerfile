# syntax=docker/dockerfile:1.7

FROM golang:1.25 AS builder

ARG GO_BUILD_TAGS=with_utls

WORKDIR /src
COPY go.mod ./

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
  CGO_ENABLED=0 go build -mod=mod -tags "${GO_BUILD_TAGS}" -trimpath -ldflags "-s -w" -o /out/sboard-node ./cmd/node

FROM debian:stable-slim
WORKDIR /app
COPY --from=builder /out/sboard-node /app/sboard-node

ENV NODE_HTTP_ADDR=:3000
ENV NODE_LOG_LEVEL=info

EXPOSE 3000
ENTRYPOINT ["/app/sboard-node"]
